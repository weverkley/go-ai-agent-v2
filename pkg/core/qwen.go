package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/telemetry"
	"go-ai-agent-v2/go-cli/pkg/types"

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
func toOpenAITools(tools []*types.ToolDefinition) []openai.Tool {
	if tools == nil {
		return nil
	}
	openaiTools := make([]openai.Tool, 0)
	for _, tool := range tools {
		for _, decl := range tool.FunctionDeclarations {
			openaiTools = append(openaiTools, openai.Tool{
				Type: openai.ToolTypeFunction,
				Function: &openai.FunctionDefinition{
					Name:        decl.Name,
					Description: decl.Description,
					Parameters:  decl.Parameters,
				},
			})
		}
	}
	return openaiTools
}

// QwenChat represents a Qwen chat client.
type QwenChat struct {
	client               *openai.Client
	modelName            string
	startHistory         []*types.Content
	toolRegistry         types.ToolRegistryInterface
	ToolConfirmationChan chan types.ToolConfirmationOutcome
	logger               telemetry.TelemetryLogger // New field for telemetry logger
}

// NewQwenChat creates a new QwenChat instance.
func NewQwenChat(cfg types.Config, generationConfig types.GenerateContentConfig, startHistory []*types.Content, logger telemetry.TelemetryLogger) (Executor, error) {
	apiKey := os.Getenv("QWEN_API_KEY")
	if apiKey == "" {
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

	return &QwenChat{
		client:               client,
		modelName:            modelName,
		startHistory:         startHistory,
		toolRegistry:         qwenChatToolRegistry,
		ToolConfirmationChan: make(chan types.ToolConfirmationOutcome, 1),
		logger:               logger, // Assign the logger
	}, nil
}

// StreamContent implements the streaming generation for QwenChat.
func (qc *QwenChat) StreamContent(ctx context.Context, contents ...*types.Content) (<-chan any, error) {
	eventChan := make(chan any)

	go func() {
		defer close(eventChan)

		historyMessages, err := toOpenAIMessages(contents[:len(contents)-1], qc.logger)
		if err != nil {
			eventChan <- types.ErrorEvent{Err: fmt.Errorf("failed to convert history: %w", err)}
			return
		}

		lastMessage, err := toOpenAIMessages(contents[len(contents)-1:], qc.logger)
		if err != nil {
			eventChan <- types.ErrorEvent{Err: fmt.Errorf("failed to convert last message: %w", err)}
			return
		}

		messages := append(historyMessages, lastMessage...)

		// Extract and log the prompt before sending
		var promptBuilder strings.Builder
		for _, msg := range messages {
			if msg.Content != "" {
				promptBuilder.WriteString(msg.Content)
			}
		}
		if promptBuilder.Len() > 0 {
			qc.logger.LogPrompt(promptBuilder.String())
		}

		var openaiTools []openai.Tool
		if qc.toolRegistry != nil {
			allTools := qc.toolRegistry.GetAllTools()
			toolDefinitions := make([]*types.ToolDefinition, len(allTools))
			for i, t := range allTools {
				toolDefinitions[i] = &types.ToolDefinition{
					FunctionDeclarations: []*types.FunctionDeclaration{
						{
							Name:        t.Name(),
							Description: t.Description(),
							Parameters:  t.Parameters(),
						},
					},
				}
			}
			openaiTools = toOpenAITools(toolDefinitions)
		}

		req := openai.ChatCompletionRequest{
			Model:    qc.modelName,
			Messages: messages,
			Stream:   true,
			Tools:    openaiTools,
		}

		stream, err := qc.client.CreateChatCompletionStream(ctx, req)
		if err != nil {
			eventChan <- types.ErrorEvent{Err: fmt.Errorf("failed to create Qwen stream: %w", err)}
			return
		}
		defer stream.Close()

		toolCallBuffers := make(map[string]strings.Builder)
		toolCallNames := make(map[string]string)
		var lastToolCallId string // To track the ID of the tool call currently being streamed

		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				eventChan <- types.ErrorEvent{Err: fmt.Errorf("error receiving Qwen stream: %w", err)}
				return
			}

			if len(response.Choices) > 0 {
				delta := response.Choices[0].Delta
				if delta.Content != "" {
					eventChan <- types.Part{Text: delta.Content}
				}
				if delta.ToolCalls != nil {
					// Accumulate tool call arguments across chunks
					for _, tc := range delta.ToolCalls {
						currentID := tc.ID
						currentName := tc.Function.Name

						// If tc.ID is empty, it's a continuation of the previous tool call
						if currentID == "" {
							if lastToolCallId != "" {
								currentID = lastToolCallId // Use the last known ID
							} else {
								// This should ideally not happen if the first chunk always has an ID
								qc.logger.LogWarnf("Qwen StreamContent: Received ToolCall with empty ID and no lastToolCallId, arguments: '%s'", tc.Function.Arguments)
								continue // Skip this chunk if we can't identify the tool call
							}
						}

						// Append arguments to the buffer for the determined ID
						builder := toolCallBuffers[currentID]
						builder.WriteString(tc.Function.Arguments)
						toolCallBuffers[currentID] = builder

						// Store the name if present (usually in the first chunk, or if it's explicitly repeated)
						if currentName != "" {
							toolCallNames[currentID] = currentName
						}

						// Update lastToolCallId if a non-empty ID was encountered
						if tc.ID != "" {
							lastToolCallId = tc.ID
						}
					}
				}

				// Check if the message is finished, then process any accumulated tool calls
				if response.Choices[0].FinishReason == openai.FinishReasonToolCalls || response.Choices[0].FinishReason == openai.FinishReasonStop || errors.Is(err, io.EOF) {
					for id, builder := range toolCallBuffers {
						jsonArgs := builder.String()
						var args map[string]any
						if jsonArgs != "" {
							if err := json.Unmarshal([]byte(jsonArgs), &args); err != nil {
								eventChan <- types.ErrorEvent{Err: fmt.Errorf("failed to unmarshal accumulated tool arguments for ID %s: %w, args: '%s'", id, err, jsonArgs)}
								// Do not return here, try to process other tool calls
								continue
							}
						}
						// Retrieve the name from the stored map
						name := toolCallNames[id]
						eventChan <- types.Part{FunctionCall: &types.FunctionCall{
							ID:   id,
							Name: name,
							Args: args,
						}}
						// Clear the buffer for this tool call after processing
						delete(toolCallBuffers, id)
						delete(toolCallNames, id)
					}
					// If EOF is received, we break the loop.
					if errors.Is(err, io.EOF) {
						break
					}
				}
			}
		}
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
		allTools := qc.toolRegistry.GetAllTools()
		toolDefinitions := make([]*types.ToolDefinition, len(allTools))
		for i, t := range allTools {
			toolDefinitions[i] = &types.ToolDefinition{
				FunctionDeclarations: []*types.FunctionDeclaration{
					{
						Name:        t.Name(),
						Description: t.Description(),
						Parameters:  t.Parameters(),
					},
				},
			}
		}
		req.Tools = toOpenAITools(toolDefinitions)
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

func (qc *QwenChat) CompressChat(promptId string, force bool) (*types.ChatCompressionResult, error) {
	return nil, fmt.Errorf("CompressChat not implemented for QwenChat")
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
