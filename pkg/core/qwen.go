package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/prompts"
	"go-ai-agent-v2/go-cli/pkg/telemetry"
	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/pkoukk/tiktoken-go"
	"github.com/sashabaranov/go-openai"
)

// toOpenAIMessages converts a slice of generic *types.Content to []openai.ChatCompletionMessage.
func toOpenAIMessages(contents []*types.Content, logger telemetry.TelemetryLogger) ([]openai.ChatCompletionMessage, error) {
	var messages []openai.ChatCompletionMessage
	for _, content := range contents {
		var chatMessage openai.ChatCompletionMessage
		switch content.Role {
		case "user":
			chatMessage.Role = openai.ChatMessageRoleUser
		case "model":
			chatMessage.Role = openai.ChatMessageRoleAssistant
		case "function", "tool": // Map both 'function' and 'tool' to openai.ChatMessageRoleTool
			chatMessage.Role = openai.ChatMessageRoleTool
		case "system":
			chatMessage.Role = openai.ChatMessageRoleSystem
		default:
			// Fallback for unknown roles, log a warning or return an error if strict validation is needed
			logger.LogWarnf("toOpenAIMessages: Unknown content role '%s', mapping to user role as fallback.", content.Role)
			chatMessage.Role = openai.ChatMessageRoleUser
		}

		var contentParts []string
		var toolCalls []openai.ToolCall

		for _, part := range content.Parts {
			if part.Text != "" {
				contentParts = append(contentParts, part.Text)
			} else if part.FunctionCall != nil {
				argsBytes, err := json.Marshal(part.FunctionCall.Args)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal function call arguments: %w", err)
				}
				toolCalls = append(toolCalls, openai.ToolCall{
					ID:   part.FunctionCall.ID, // Assuming ID is populated
					Type: openai.ToolTypeFunction,
					Function: openai.FunctionCall{
						Name:      part.FunctionCall.Name,
						Arguments: string(argsBytes),
					},
				})
			} else if part.FunctionResponse != nil {
				responseBytes, err := json.Marshal(part.FunctionResponse.Response)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal function response: %w", err)
				}
				contentParts = append(contentParts, fmt.Sprintf("Tool response for %s: %s", part.FunctionResponse.Name, string(responseBytes)))
			}
		}

		if len(contentParts) > 0 {
			chatMessage.Content = strings.Join(contentParts, "\n")
		}
		if len(toolCalls) > 0 {
			chatMessage.ToolCalls = toolCalls
		}
		messages = append(messages, chatMessage)
	}
	return messages, nil
}

// fromOpenAIMessage converts an openai.ChatCompletionMessage to a generic *types.Content.
func fromOpenAIMessage(msg openai.ChatCompletionMessage) (*types.Content, error) {
	content := &types.Content{Role: msg.Role}
	var parts []types.Part

	if msg.Content != "" {
		parts = append(parts, types.Part{Text: msg.Content})
	}

	for _, tc := range msg.ToolCalls {
		var args map[string]interface{}
		if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tool call arguments from response: %w", err)
		}
		parts = append(parts, types.Part{FunctionCall: &types.FunctionCall{
			ID:   tc.ID,
			Name: tc.Function.Name,
			Args: args,
		}})
	}

	content.Parts = parts
	return content, nil
}

// toOpenAITools converts generic []*types.ToolDefinition to []openai.Tool.
func toOpenAITools(toolRegistry types.ToolRegistryInterface, logger telemetry.TelemetryLogger) []openai.Tool {
	if toolRegistry == nil {
		return nil
	}

	allTools := toolRegistry.GetAllTools()
	if allTools == nil {
		return nil
	}

	openaiTools := make([]openai.Tool, 0, len(allTools)) // Pre-allocate capacity
	for _, t := range allTools {
		openaiTools = append(openaiTools, openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        t.Name(),
				Description: t.Description(),
				Parameters:  t.Parameters(),
			},
		})
	}
	return openaiTools
}

// QwenChat represents a Qwen chat client.
type QwenChat struct {
	client               *openai.Client
	modelName            string
	generationConfig     types.GenerateContentConfig // New field
	startHistory         []*types.Content
	toolRegistry         types.ToolRegistryInterface
	ToolConfirmationChan chan types.ToolConfirmationOutcome
	logger               telemetry.TelemetryLogger // New field for telemetry logger
}

// NewQwenChat creates a new QwenChat instance.
func NewQwenChat(cfg types.Config, generationConfig types.GenerateContentConfig, startHistory []*types.Content, logger telemetry.TelemetryLogger) (Executor, error) {
	logger.LogDebugf("NewQwenChat: Initializing...")
	apiKey := os.Getenv("QWEN_API_KEY")
	if apiKey == "" {
		logger.LogErrorf("NewQwenChat: QWEN_API_KEY environment variable not set")
		return nil, fmt.Errorf("QWEN_API_KEY environment variable not set")
	}

	modelVal, ok := cfg.Get("model")
	if !ok {
		return nil, fmt.Errorf("model not found in config")
	}
	modelName, ok := modelVal.(string)
	if !ok {
		return nil, fmt.Errorf("model in config is not a string")
	}

	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://dashscope-intl.aliyuncs.com/compatible-mode/v1"
	client := openai.NewClientWithConfig(config)

	toolRegistryVal, ok := cfg.Get("toolRegistry")
	var qwenChatToolRegistry types.ToolRegistryInterface
	if ok && toolRegistryVal != nil {
		if tr, toolRegistryOk := toolRegistryVal.(types.ToolRegistryInterface); toolRegistryOk {
			qwenChatToolRegistry = tr
		}
	}
	logger.LogDebugf("NewQwenChat: Initialization complete for model '%s'.", modelName)
	return &QwenChat{
		client:               client,
		modelName:            modelName,
		generationConfig:     generationConfig, // Initialize new field
		startHistory:         startHistory,
		toolRegistry:         qwenChatToolRegistry,
		ToolConfirmationChan: make(chan types.ToolConfirmationOutcome, 1),
		logger:               logger,
	}, nil
}

func (qc *QwenChat) StreamContent(ctx context.Context, contents ...*types.Content) (<-chan any, error) {
	eventChan := make(chan any)

	go func() {
		defer close(eventChan)
		qc.logger.LogDebugf("QwenExecutor: StreamContent goroutine started.")

		if len(contents) == 0 {
			qc.logger.LogErrorf("QwenExecutor: StreamContent called with no content.")
			eventChan <- types.ErrorEvent{Err: fmt.Errorf("QwenExecutor: StreamContent called with no content")}
			return
		}

		messages, err := toOpenAIMessages(contents, qc.logger)
		if err != nil {
			qc.logger.LogErrorf("QwenExecutor: toOpenAIMessages failed: %v", err)
			eventChan <- types.ErrorEvent{Err: fmt.Errorf("failed to convert messages: %w", err)}
			return
		}
		qc.logger.LogDebugf("QwenExecutor: Converted %d messages for OpenAI API.", len(messages))

		// Prepend system message if provided
		if qc.generationConfig.SystemInstruction != "" {
			systemMessage := openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleSystem,
				Content: qc.generationConfig.SystemInstruction,
			}
			messages = append([]openai.ChatCompletionMessage{systemMessage}, messages...)
			qc.logger.LogDebugf("QwenExecutor: Prepended system message.")
		}

		var openaiTools []openai.Tool
		if qc.toolRegistry != nil {
			qc.logger.LogDebugf("QwenExecutor: Building OpenAI tools...")
			openaiTools = toOpenAITools(qc.toolRegistry, qc.logger)
			qc.logger.LogDebugf("QwenExecutor: Finished building %d tools.", len(openaiTools))
		}

		req := openai.ChatCompletionRequest{
			Model:    qc.modelName,
			Messages: messages,
			Stream:   true,
			Tools:    openaiTools,
		}

		qc.logger.LogDebugf("QwenExecutor: Calling CreateChatCompletionStream...")
		stream, err := qc.client.CreateChatCompletionStream(ctx, req)
		if err != nil {
			qc.logger.LogErrorf("QwenExecutor: CreateChatCompletionStream failed: %v", err)
			eventChan <- types.ErrorEvent{Err: fmt.Errorf("failed to create Qwen stream: %w", err)}
			return
		}
		defer stream.Close()
		qc.logger.LogDebugf("QwenExecutor: Stream created. Waiting for response...")

		toolCallBuffers := make(map[string]strings.Builder)
		toolCallNames := make(map[string]string)
		var lastToolCallId string

		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				qc.logger.LogDebugf("QwenExecutor: Stream finished (EOF).")
				break
			}
			if err != nil {
				qc.logger.LogErrorf("QwenExecutor: Error receiving from stream: %v", err)
				eventChan <- types.ErrorEvent{Err: fmt.Errorf("error receiving Qwen stream: %w", err)}
				return
			}
			qc.logger.LogDebugf("QwenExecutor: Received a response chunk.")

			if len(response.Choices) > 0 {
				delta := response.Choices[0].Delta
				if delta.Content != "" {
					eventChan <- types.Part{Text: delta.Content}
				}
				if delta.ToolCalls != nil {
					for _, tc := range delta.ToolCalls {
						currentID := tc.ID
						currentName := tc.Function.Name
						if currentID == "" {
							if lastToolCallId != "" {
								currentID = lastToolCallId
							} else {
								qc.logger.LogWarnf("QwenExecutor: Received ToolCall with empty ID and no lastToolCallId, arguments: '%s'", tc.Function.Arguments)
								continue
							}
						}
						builder := toolCallBuffers[currentID]
						builder.WriteString(tc.Function.Arguments)
						toolCallBuffers[currentID] = builder
						if currentName != "" {
							toolCallNames[currentID] = currentName
						}
						if tc.ID != "" {
							lastToolCallId = tc.ID
						}
					}
				}

				if response.Choices[0].FinishReason == openai.FinishReasonToolCalls || response.Choices[0].FinishReason == openai.FinishReasonStop {
					qc.logger.LogDebugf("QwenExecutor: Stream received finish reason: %s. Processing %d buffered tool calls.", response.Choices[0].FinishReason, len(toolCallBuffers))
					for id, builder := range toolCallBuffers {
						jsonArgs := builder.String()
						var args map[string]any
						if jsonArgs != "" {
							if err := json.Unmarshal([]byte(jsonArgs), &args); err != nil {
								qc.logger.LogErrorf("QwenExecutor: Failed to unmarshal tool arguments for ID %s: %v", id, err)
								eventChan <- types.ErrorEvent{Err: fmt.Errorf("failed to unmarshal accumulated tool arguments for ID %s: %w, args: '%s'", id, err, jsonArgs)}
								continue
							}
						}
						name := toolCallNames[id]
						qc.logger.LogDebugf("QwenExecutor: Sending complete tool call '%s' (ID: %s)", name, id)
						eventChan <- types.Part{FunctionCall: &types.FunctionCall{
							ID:   id,
							Name: name,
							Args: args,
						}}
						delete(toolCallBuffers, id)
						delete(toolCallNames, id)
					}
				}
			}
		}
		qc.logger.LogDebugf("QwenExecutor: Finished processing stream.")
	}()

	return eventChan, nil
}

// SetHistory sets the chat history for QwenChat.
func (qc *QwenChat) SetHistory(history []*types.Content) error {
	qc.startHistory = history
	return nil
}

func (qc *QwenChat) GenerateContent(contents ...*types.Content) (*types.GenerateContentResponse, error) {
	historyMessages, err := toOpenAIMessages(qc.startHistory, qc.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to convert history: %w", err)
	}

	requestMessages, err := toOpenAIMessages(contents, qc.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to convert request contents: %w", err)
	}

	messages := append(historyMessages, requestMessages...)

	req := openai.ChatCompletionRequest{
		Model:    qc.modelName,
		Messages: messages,
	}

	if qc.toolRegistry != nil {
		req.Tools = toOpenAITools(qc.toolRegistry, qc.logger)
	}

	ctx := context.Background()
	resp, err := qc.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create Qwen chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from Qwen API")
	}

	genericContent, err := fromOpenAIMessage(resp.Choices[0].Message)
	if err != nil {
		return nil, fmt.Errorf("failed to convert openai message to generic content: %w", err)
	}

	return &types.GenerateContentResponse{
		Candidates: []*types.Candidate{
			{
				Content: genericContent,
			},
		},
	}, nil
}

func (qc *QwenChat) ExecuteTool(ctx context.Context, fc *types.FunctionCall) (types.ToolResult, error) {
	if qc.toolRegistry == nil {
		return types.ToolResult{}, fmt.Errorf("tool registry not initialized")
	}

	tool, err := qc.toolRegistry.GetTool(fc.Name)
	if err != nil {
		return types.ToolResult{}, fmt.Errorf("tool %s not found: %w", fc.Name, err)
	}

	return tool.Execute(ctx, fc.Args)
}

func (qc *QwenChat) SendMessageStream(modelName string, messageParams types.MessageParams, promptId string) (<-chan types.StreamResponse, error) {
	return nil, fmt.Errorf("SendMessageStream not implemented for QwenChat")
}

func (qc *QwenChat) ListModels() ([]string, error) {
	return []string{"qwen-turbo", "qwen-plus", "qwen-max"}, nil
}

func (qc *QwenChat) GetHistory() ([]*types.Content, error) {
	return qc.startHistory, nil
}

func (qc *QwenChat) CompressChat(history []*types.Content, promptId string) (*types.ChatCompressionResult, error) {
	// 1. Get the summarization prompt
	summarizePrompt, ok := prompts.GetPrompt("compression")
	if !ok {
		return nil, fmt.Errorf("chat compression prompt not found")
	}

	// 2. Combine the prompt and the history
	var historyText strings.Builder
	for _, content := range history {
		for _, part := range content.Parts {
			historyText.WriteString(fmt.Sprintf("%s: %s\n", content.Role, part.Text))
		}
	}

	fullPrompt := summarizePrompt + "\n\n--- CONVERSATION HISTORY ---\n" + historyText.String()

	// 3. Count original tokens using tiktoken
	tke, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		return nil, fmt.Errorf("failed to get tiktoken encoding: %w", err)
	}
	originalTokenCount := len(tke.Encode(fullPrompt, nil, nil))

	// 4. Call the model to get the summary
	req := openai.ChatCompletionRequest{
		Model: qc.modelName,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: summarizePrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: historyText.String(),
			},
		},
	}
	resp, err := qc.client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate summary from qwen: %w", err)
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("received an empty summary response from qwen")
	}
	summaryText := resp.Choices[0].Message.Content

	// 5. Count new tokens
	newTokenCount := len(tke.Encode(summaryText, nil, nil))

	return &types.ChatCompressionResult{
		Summary:            summaryText,
		OriginalTokenCount: originalTokenCount,
		NewTokenCount:      newTokenCount,
		CompressionStatus:  "OK",
	}, nil
}

// GenerateContentWithTools is a placeholder implementation for QwenChat.
func (qc *QwenChat) GenerateContentWithTools(ctx context.Context, history []*types.Content, tools []types.Tool) (*types.GenerateContentResponse, error) {
	// Convert []types.Tool to []*types.ToolDefinition
	toolDefinitions := make([]*types.ToolDefinition, len(tools))
	for i, tool := range tools {
		toolDefinitions[i] = &types.ToolDefinition{
			FunctionDeclarations: []*types.FunctionDeclaration{
				{
					Name:        tool.Name(),
					Description: tool.Description(),
					Parameters:  tool.Parameters(),
				},
			},
		}
	}
	// This would then use the toolDefinitions
	// For now, returning not implemented, but the conversion is done.
	return nil, fmt.Errorf("GenerateContentWithTools not implemented for QwenChat")
}

// SetUserConfirmationChannel is a no-op for QwenChat.
func (qc *QwenChat) SetUserConfirmationChannel(ch chan bool) {
	// No-op
}

// SetToolConfirmationChannel sets the channel for tool confirmation.
func (qc *QwenChat) SetToolConfirmationChannel(ch chan types.ToolConfirmationOutcome) {
	qc.ToolConfirmationChan = ch
}

// Name returns the name of the executor (model name for Qwen).
func (qc *QwenChat) Name() string {
	return qc.modelName
}
