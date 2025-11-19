package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/sashabaranov/go-openai"
)

// toOpenAIMessages converts a slice of generic *types.Content to []openai.ChatCompletionMessage.
func toOpenAIMessages(contents []*types.Content) ([]openai.ChatCompletionMessage, error) {
	var messages []openai.ChatCompletionMessage
	for _, content := range contents {
		var chatMessage openai.ChatCompletionMessage
		chatMessage.Role = content.Role

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
	client       *openai.Client
	modelName    string
	startHistory []*types.Content
	toolRegistry types.ToolRegistryInterface
}

// NewQwenChat creates a new QwenChat instance.
func NewQwenChat(cfg types.Config, generationConfig types.GenerateContentConfig, startHistory []*types.Content) (Executor, error) {
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
	config.BaseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	client := openai.NewClientWithConfig(config)

	toolRegistryVal, ok := cfg.Get("toolRegistry")
	var qwenChatToolRegistry types.ToolRegistryInterface
	if ok && toolRegistryVal != nil {
		if tr, toolRegistryOk := toolRegistryVal.(types.ToolRegistryInterface); toolRegistryOk {
			qwenChatToolRegistry = tr
		}
	}

	return &QwenChat{
		client:       client,
		modelName:    modelName,
		startHistory: startHistory,
		toolRegistry: qwenChatToolRegistry,
	}, nil
}

// GenerateStream implements the streaming generation for QwenChat.
func (qc *QwenChat) GenerateStream(ctx context.Context, contents ...*types.Content) (<-chan any, error) {
	eventChan := make(chan any)

	go func() {
		defer close(eventChan)

		eventChan <- types.StreamingStartedEvent{}
		eventChan <- types.ThinkingEvent{}

		historyMessages, err := toOpenAIMessages(qc.startHistory)
		if err != nil {
			eventChan <- types.ErrorEvent{Err: fmt.Errorf("failed to convert history: %w", err)}
			return
		}
		
		requestMessages, err := toOpenAIMessages(contents)
		if err != nil {
			eventChan <- types.ErrorEvent{Err: fmt.Errorf("failed to convert request contents: %w", err)}
			return
		}

		messages := append(historyMessages, requestMessages...)

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

		var accumulatedText string
		var toolCalls []openai.ToolCall
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
				accumulatedText += delta.Content
				if delta.ToolCalls != nil {
					toolCalls = append(toolCalls, delta.ToolCalls...)
				}
			}
		}

		if len(toolCalls) > 0 {
			// Execute tool calls
			var toolResponses []openai.ChatCompletionMessage
			for _, tc := range toolCalls {
				var args map[string]any
				if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
					eventChan <- types.ErrorEvent{Err: fmt.Errorf("failed to unmarshal tool arguments: %w", err)}
					return
				}

				toolCallID := "mock-id-for-qwen" // Generate a mock ID
				eventChan <- types.ToolCallStartEvent{
					ToolCallID: toolCallID,
					ToolName:   tc.Function.Name,
					Args:       args,
				}

				toolResult, err := qc.ExecuteTool(ctx, &types.FunctionCall{Name: tc.Function.Name, Args: args})
				if err != nil {
					eventChan <- types.ErrorEvent{Err: fmt.Errorf("error executing tool %s: %w", tc.Function.Name, err)}
					return
				}

				eventChan <- types.ToolCallEndEvent{
					ToolCallID: toolCallID,
					ToolName:   tc.Function.Name,
					Result:     toolResult.ReturnDisplay,
					Err:        err,
				}

				responseBytes, err := json.Marshal(map[string]interface{}{"result": toolResult.LLMContent})
				if err != nil {
					eventChan <- types.ErrorEvent{Err: fmt.Errorf("failed to marshal tool result: %w", err)}
					return
				}

				toolResponses = append(toolResponses, openai.ChatCompletionMessage{
					Role:       openai.ChatMessageRoleTool,
					Content:    string(responseBytes),
					ToolCallID: tc.ID,
				})
			}

			// Send tool responses back to the model
			messages = append(messages, toolResponses...)
			req = openai.ChatCompletionRequest{
				Model:    qc.modelName,
				Messages: messages,
				Stream:   true,
			}

			stream, err = qc.client.CreateChatCompletionStream(ctx, req)
			if err != nil {
				eventChan <- types.ErrorEvent{Err: fmt.Errorf("failed to create Qwen stream after tool call: %w", err)}
				return
			}
			defer stream.Close()

			accumulatedText = ""
			for {
				response, err := stream.Recv()
				if errors.Is(err, io.EOF) {
					break
				}
				if err != nil {
					eventChan <- types.ErrorEvent{Err: fmt.Errorf("error receiving Qwen stream after tool call: %w", err)}
					return
				}

				if len(response.Choices) > 0 {
					delta := response.Choices[0].Delta
					accumulatedText += delta.Content
				}
			}
		}

		eventChan <- types.FinalResponseEvent{Content: accumulatedText}
	}()

	return eventChan, nil
}

// SetHistory sets the chat history for QwenChat.
func (qc *QwenChat) SetHistory(history []*types.Content) error {
	qc.startHistory = history
	return nil
}

func (qc *QwenChat) GenerateContent(contents ...*types.Content) (*types.GenerateContentResponse, error) {
	historyMessages, err := toOpenAIMessages(qc.startHistory)
	if err != nil {
		return nil, fmt.Errorf("failed to convert history: %w", err)
	}

	requestMessages, err := toOpenAIMessages(contents)
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