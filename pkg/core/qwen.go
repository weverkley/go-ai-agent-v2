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

func (qc *QwenChat) StreamContent(ctx context.Context, contents []*types.Content, tools []types.Tool) (<-chan any, error) {
	messageParts := []types.Part{}
	for _, content := range contents {
		messageParts = append(messageParts, content.Parts...)
	}

	toolDefinitions := make([]*types.ToolDefinition, 0, len(tools))
	for _, tool := range tools {
		toolDefinitions = append(toolDefinitions, &types.ToolDefinition{
			FunctionDeclarations: []*types.FunctionDeclaration{
				{
					Name:        tool.Name(),
					Description: tool.Description(),
					Parameters:  tool.Parameters(),
				},
			},
		})
	}

	messageParams := types.MessageParams{
		Message:     messageParts,
		Tools:       toolDefinitions, // Pass the converted tool definitions
		AbortSignal: ctx,
	}

	streamResponseChan, err := qc.SendMessageStream(qc.modelName, messageParams, "")
	if err != nil {
		return nil, err
	}

	// Convert <-chan types.StreamResponse to <-chan any
	anyStreamChan := make(chan any)
	go func() {
		defer close(anyStreamChan)
		for sr := range streamResponseChan {
			if sr.Type == types.StreamEventTypeChunk {
				if sr.Value != nil {
					if genContent, ok := sr.Value.(*types.GenerateContentResponse); ok && len(genContent.Candidates) > 0 {
						for _, part := range genContent.Candidates[0].Content.Parts {
							if part.Text != "" {
								anyStreamChan <- types.Part{Text: part.Text}
							}
							if part.FunctionCall != nil {
								anyStreamChan <- types.Part{FunctionCall: part.FunctionCall}
							}
							// Handle TokenCountEvent from the "<!-- TokenCount: ... -->" format
							if strings.HasPrefix(part.Text, "<!-- TokenCount:") {
								var inputTokens, outputTokens int
								fmt.Sscanf(part.Text, "<!-- TokenCount: %d input, %d output -->", &inputTokens, &outputTokens)
								anyStreamChan <- types.TokenCountEvent{InputTokens: inputTokens, OutputTokens: outputTokens}
							}
						}
					}
				}
			} else if sr.Type == types.StreamEventTypeError {
				anyStreamChan <- types.ErrorEvent{Err: sr.Error}
			} else if sr.Type == types.StreamEventTypeTokenCount {
				if tokenEvent, ok := sr.Value.(types.TokenCountEvent); ok {
					anyStreamChan <- tokenEvent
				}
			}
		}
	}()

	return anyStreamChan, nil
}

func (qc *QwenChat) SendMessageStream(modelName string, messageParams types.MessageParams, promptId string) (<-chan types.StreamResponse, error) {
	eventChan := make(chan types.StreamResponse)
	go func() {
		defer close(eventChan)
		qc.logger.LogDebugf("QwenExecutor: SendMessageStream goroutine started for promptId: %s.", promptId)

		messages := make([]openai.ChatCompletionMessage, 0)

		// Start with qc.startHistory
		historyMessages, err := toOpenAIMessages(qc.startHistory, qc.logger)
		if err != nil {
			qc.logger.LogErrorf("QwenExecutor: toOpenAIMessages failed for startHistory: %v", err)
			eventChan <- types.StreamResponse{Type: types.StreamEventTypeError, Error: fmt.Errorf("failed to convert start history: %w", err)}
			return
		}
		messages = append(messages, historyMessages...)

		// Add messages from messageParams.Message (which represents current turn's user input or tool responses)
		currentTurnContents := []*types.Content{{Parts: messageParams.Message, Role: "user"}} // Assuming messageParams.Message are user parts
		currentTurnMessages, err := toOpenAIMessages(currentTurnContents, qc.logger)
		if err != nil {
			qc.logger.LogErrorf("QwenExecutor: toOpenAIMessages failed for messageParams.Message: %v", err)
			eventChan <- types.StreamResponse{Type: types.StreamEventTypeError, Error: fmt.Errorf("failed to convert current message parts: %w", err)}
			return
		}
		messages = append(messages, currentTurnMessages...)

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

		// Count input tokens
		var inputText strings.Builder
		for _, msg := range messages {
			inputText.WriteString(msg.Content)
			if msg.ToolCalls != nil {
				for _, tc := range msg.ToolCalls {
					inputText.WriteString(tc.Function.Arguments)
				}
			}
		}
		tke, err := tiktoken.GetEncoding("cl100k_base")
		if err != nil {
			eventChan <- types.StreamResponse{Type: types.StreamEventTypeError, Error: fmt.Errorf("failed to get tiktoken encoding: %w", err)}
			return
		}
		inputTokens := len(tke.Encode(inputText.String(), nil, nil))

		var openaiTools []openai.Tool
		if qc.toolRegistry != nil && messageParams.Tools != nil {
			qc.logger.LogDebugf("QwenExecutor: Building OpenAI tools from messageParams...")
			for _, toolDef := range messageParams.Tools {
				for _, fd := range toolDef.FunctionDeclarations {
					openaiTools = append(openaiTools, openai.Tool{
						Type: openai.ToolTypeFunction,
						Function: &openai.FunctionDefinition{
							Name:        fd.Name,
							Description: fd.Description,
							Parameters:  fd.Parameters,
						},
					})
				}
			}
			qc.logger.LogDebugf("QwenExecutor: Finished building %d tools.", len(openaiTools))
		}

		req := openai.ChatCompletionRequest{
			Model:    modelName,
			Messages: messages,
			Stream:   true,
			Tools:    openaiTools,
		}

		qc.logger.LogDebugf("QwenExecutor: Calling CreateChatCompletionStream...")
		stream, err := qc.client.CreateChatCompletionStream(messageParams.AbortSignal, req)
		if err != nil {
			qc.logger.LogErrorf("QwenExecutor: CreateChatCompletionStream failed: %v", err)
			eventChan <- types.StreamResponse{Type: types.StreamEventTypeError, Error: fmt.Errorf("failed to create Qwen stream: %w", err)}
			return
		}
		defer stream.Close()
		qc.logger.LogDebugf("QwenExecutor: Stream created. Waiting for response...")

		var outputText strings.Builder
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
				eventChan <- types.StreamResponse{Type: types.StreamEventTypeError, Error: fmt.Errorf("error receiving Qwen stream: %w", err)}
				return
			}
			qc.logger.LogDebugf("QwenExecutor: Received a response chunk.")

			if len(response.Choices) > 0 {
				delta := response.Choices[0].Delta
				if delta.Content != "" {
					outputText.WriteString(delta.Content)
					eventChan <- types.StreamResponse{Type: types.StreamEventTypeChunk, Value: &types.GenerateContentResponse{
						Candidates: []*types.Candidate{
							{
								Content: &types.Content{Parts: []types.Part{{Text: delta.Content}}},
							},
						},
					}}
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
								eventChan <- types.StreamResponse{Type: types.StreamEventTypeError, Error: fmt.Errorf("failed to unmarshal accumulated tool arguments for ID %s: %w, args: '%s'", id, err, jsonArgs)}
								continue
							}
						}
						name := toolCallNames[id]
						qc.logger.LogDebugf("QwenExecutor: Sending complete tool call '%s' (ID: %s)", name, id)
						eventChan <- types.StreamResponse{Type: types.StreamEventTypeChunk, Value: &types.GenerateContentResponse{
							Candidates: []*types.Candidate{
								{
									Content: &types.Content{Parts: []types.Part{{FunctionCall: &types.FunctionCall{
										ID:   id,
										Name: name,
										Args: args,
									}}}},
								},
							},
						}}
						delete(toolCallBuffers, id)
						delete(toolCallNames, id)
					}
				}
			}
		} // Closes `for { response, err := stream.Recv() ... }`
		outputTokens := len(tke.Encode(outputText.String(), nil, nil))
		eventChan <- types.StreamResponse{Type: types.StreamEventTypeTokenCount, Value: types.TokenCountEvent{InputTokens: inputTokens, OutputTokens: outputTokens}}
		qc.logger.LogDebugf("QwenExecutor: Finished processing stream.")
	}()

	return eventChan, nil
}

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
	inputTokens := len(tke.Encode(fullPrompt, nil, nil))

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
	outputTokens := len(tke.Encode(summaryText, nil, nil))

	return &types.ChatCompressionResult{
		Summary:            summaryText,
		OriginalTokenCount: inputTokens,
		NewTokenCount:      outputTokens,
		InputTokens:        inputTokens,
		OutputTokens:       outputTokens,
		CompressionStatus:  "OK",
	}, nil
}

// GenerateContentWithTools is a placeholder implementation for QwenChat.
func (qc *QwenChat) GenerateContentWithTools(ctx context.Context, history []*types.Content, tools []types.Tool) (*types.GenerateContentResponse, error) {
	// Convert history to OpenAI messages
	messages, err := toOpenAIMessages(history, qc.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to convert history for GenerateContentWithTools: %w", err)
	}

	// Convert types.Tool to openai.Tool
	openaiTools := make([]openai.Tool, 0, len(tools))
	for _, t := range tools {
		openaiTools = append(openaiTools, openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        t.Name(),
				Description: t.Description(),
				Parameters:  t.Parameters(),
			},
		})
	}

	// Add system instruction if present
	if qc.generationConfig.SystemInstruction != "" {
		systemMessage := openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: qc.generationConfig.SystemInstruction,
		}
		messages = append([]openai.ChatCompletionMessage{systemMessage}, messages...)
	}

	req := openai.ChatCompletionRequest{
		Model:    qc.modelName,
		Messages: messages,
		Tools:    openaiTools,
	}

	resp, err := qc.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create Qwen chat completion with tools: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from Qwen API with tools")
	}

	genericContent, err := fromOpenAIMessage(resp.Choices[0].Message)
	if err != nil {
		return nil, fmt.Errorf("failed to convert openai message to generic content after tool call: %w", err)
	}

	return &types.GenerateContentResponse{
		Candidates: []*types.Candidate{
			{
				Content: genericContent,
			},
		},
	}, nil
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
