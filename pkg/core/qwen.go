package core

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/google/generative-ai-go/genai"
	"github.com/sashabaranov/go-openai"
)

// QwenChat represents a Qwen chat client.
type QwenChat struct {
	client       *openai.Client
	modelName    string
	startHistory []openai.ChatCompletionMessage
}

// NewQwenChat creates a new QwenChat instance.
func NewQwenChat(cfg types.Config, generationConfig types.GenerateContentConfig, startHistory []*genai.Content) (Executor, error) {
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

	// Convert genai.Content to openai.ChatCompletionMessage
	var history []openai.ChatCompletionMessage
	for _, content := range startHistory {
		for _, part := range content.Parts {
			if text, ok := part.(genai.Text); ok {
				history = append(history, openai.ChatCompletionMessage{
					Role:    content.Role,
					Content: string(text),
				})
			}
		}
	}

	return &QwenChat{
		client:       client,
		modelName:    modelName,
		startHistory: history,
	}, nil
}

// GenerateStream implements the streaming generation for QwenChat.
func (qc *QwenChat) GenerateStream(contents ...*genai.Content) (<-chan any, error) {
	eventChan := make(chan any)

	go func() {
		defer close(eventChan)

		eventChan <- types.StreamingStartedEvent{}
		eventChan <- types.ThinkingEvent{}

		var messages []openai.ChatCompletionMessage
		messages = append(messages, qc.startHistory...)

		for _, content := range contents {
			for _, part := range content.Parts {
				if text, ok := part.(genai.Text); ok {
					messages = append(messages, openai.ChatCompletionMessage{
						Role:    openai.ChatMessageRoleUser,
						Content: string(text),
					})
				}
			}
		}

		req := openai.ChatCompletionRequest{
			Model:    qc.modelName,
			Messages: messages,
			Stream:   true,
		}

		ctx := context.Background()
		stream, err := qc.client.CreateChatCompletionStream(ctx, req)
		if err != nil {
			eventChan <- types.ErrorEvent{Err: fmt.Errorf("failed to create Qwen stream: %w", err)}
			return
		}
		defer stream.Close()

		var accumulatedText string
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
				accumulatedText += response.Choices[0].Delta.Content
			}
		}

		eventChan <- types.FinalResponseEvent{Content: accumulatedText}
	}()

	return eventChan, nil
}

// SetHistory sets the chat history for QwenChat.
func (qc *QwenChat) SetHistory(history []*genai.Content) error {
	var newHistory []openai.ChatCompletionMessage
	for _, content := range history {
		for _, part := range content.Parts {
			if text, ok := part.(genai.Text); ok {
				newHistory = append(newHistory, openai.ChatCompletionMessage{
					Role:    content.Role,
					Content: string(text),
				})
			}
		}
	}
	qc.startHistory = newHistory
	return nil
}

// The following methods are not yet implemented for QwenChat.

func (qc *QwenChat) GenerateContent(contents ...*genai.Content) (*genai.GenerateContentResponse, error) {
	var messages []openai.ChatCompletionMessage
	messages = append(messages, qc.startHistory...)

	for _, content := range contents {
		for _, part := range content.Parts {
			if text, ok := part.(genai.Text); ok {
				messages = append(messages, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleUser,
					Content: string(text),
				})
			}
		}
	}

	req := openai.ChatCompletionRequest{
		Model:    qc.modelName,
		Messages: messages,
	}

	ctx := context.Background()
	resp, err := qc.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create Qwen chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from Qwen API")
	}

	// Convert openai.ChatCompletionResponse to genai.GenerateContentResponse
	var genaiParts []genai.Part
	genaiParts = append(genaiParts, genai.Text(resp.Choices[0].Message.Content))

	return &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: &genai.Content{
					Parts: genaiParts,
					Role:  "model",
				},
			},
		},
	}, nil
}

func (qc *QwenChat) ExecuteTool(fc *genai.FunctionCall) (types.ToolResult, error) {
	return types.ToolResult{}, fmt.Errorf("ExecuteTool not implemented for QwenChat")
}

func (qc *QwenChat) SendMessageStream(modelName string, messageParams types.MessageParams, promptId string) (<-chan types.StreamResponse, error) {
	return nil, fmt.Errorf("SendMessageStream not implemented for QwenChat")
}

func (qc *QwenChat) ListModels() ([]string, error) {
	// This would require a specific Qwen API call to list models.
	// For now, we'll return a mock list.
	return []string{"qwen-turbo", "qwen-plus", "qwen-max"}, nil
}

func (qc *QwenChat) GetHistory() ([]*genai.Content, error) {
	var history []*genai.Content
	for _, msg := range qc.startHistory {
		history = append(history, &genai.Content{
			Role:  msg.Role,
			Parts: []genai.Part{genai.Text(msg.Content)},
		})
	}
	return history, nil
}

func (qc *QwenChat) CompressChat(promptId string, force bool) (*types.ChatCompressionResult, error) {
	if len(qc.startHistory) <= 2 {
		return nil, fmt.Errorf("no conversation found to compress")
	}

	var summaryPrompt string
	for _, msg := range qc.startHistory {
		summaryPrompt += fmt.Sprintf("%s: %s\n", msg.Role, msg.Content)
	}
	summaryPrompt = "Summarize the following conversation:\n\n" + summaryPrompt

	summaryContent := &genai.Content{
		Parts: []genai.Part{genai.Text(summaryPrompt)},
	}

	resp, err := qc.GenerateContent(summaryContent)
	if err != nil {
		return nil, fmt.Errorf("failed to generate summary for Qwen: %w", err)
	}

	if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
		if text, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
			qc.startHistory = []openai.ChatCompletionMessage{
				{Role: "model", Content: string(text)},
			}
		}
	}

	// Token counting is not implemented for Qwen, so we return a nil result.
	return nil, nil
}