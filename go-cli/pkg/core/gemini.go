package core

import (
	"context"
	"fmt"
	"os"

	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// ContentGenerator interface represents the ability to generate content.
type ContentGenerator interface {
	GenerateContent(prompt string) (string, error)
}

// GeminiChat represents a Gemini chat client.
type GeminiChat struct {
	client           *genai.Client
	model            *genai.GenerativeModel
	Name             string
	generationConfig types.GenerateContentConfig
	startHistory     []*genai.Content
	toolRegistry     *types.ToolRegistry // Add ToolRegistry
}

// NewGeminiChat creates a new GeminiChat instance.
func NewGeminiChat(cfg types.Config, generationConfig types.GenerateContentConfig, startHistory []*genai.Content) (Executor, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable not set")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	model := client.GenerativeModel(cfg.Model())

	// Apply generation config
	model.SetTemperature(generationConfig.Temperature)
	model.SetTopP(generationConfig.TopP)

	// Set tools for the model
	if cfg.GetToolRegistry() != nil {
		model.Tools = cfg.GetToolRegistry().GetTools()
	}

	return &GeminiChat{
		client:           client,
		model:            model,
		Name:             cfg.Model(),
		generationConfig: generationConfig,
		startHistory:     startHistory,
		toolRegistry:     cfg.GetToolRegistry(), // Store the ToolRegistry
	}, nil
}

// NewUserContent creates a new genai.Content with user role and text part.
func NewUserContent(text string) *genai.Content {
	return &genai.Content{
		Parts: []genai.Part{genai.Text(text)},
		Role:  "user",
	}
}

// NewFunctionResponsePart creates a new genai.Part for a function response.
func NewFunctionResponsePart(name string, response interface{}) genai.Part {
	// Ensure response is of type map[string]any
	respMap, ok := response.(map[string]any)
	if !ok {
		// Handle error or convert if necessary. For now, return an empty map.
		respMap = make(map[string]any)
		respMap["error"] = fmt.Sprintf("invalid response type: %T", response)
	}
	return genai.FunctionResponse{
		Name:     name,
		Response: respMap,
	}
}

// NewFunctionCallContent creates a new genai.Content with model role and function call parts.
func NewFunctionCallContent(calls ...*genai.FunctionCall) *genai.Content {
	parts := make([]genai.Part, len(calls))
	for i, call := range calls {
		parts[i] = call
	}
	return &genai.Content{
		Parts: parts,
		Role:  "model",
	}
}

// NewToolContent creates a new genai.Content with tool role and tool response parts.
func NewToolContent(responses ...genai.Part) *genai.Content {
	return &genai.Content{
		Parts: responses,
		Role:  "tool",
	}
}

// GenerateContent generates content using the Gemini API, handling tool calls.
func (gc *GeminiChat) GenerateContent(contents ...*genai.Content) (*genai.GenerateContentResponse, error) {
	ctx := context.Background()

	// Convert []*genai.Content to []genai.Part
	var parts []genai.Part
	for _, content := range contents {
		parts = append(parts, content.Parts...)
	}

	resp, err := gc.model.GenerateContent(ctx, parts...)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	return resp, nil
}

// ExecuteTool executes a tool call.
func (gc *GeminiChat) ExecuteTool(fc *genai.FunctionCall) (types.ToolResult, error) {
	if gc.toolRegistry == nil {
		return types.ToolResult{}, fmt.Errorf("tool registry not initialized")
	}

	tool, err := gc.toolRegistry.GetTool(fc.Name)
	if err != nil {
		return types.ToolResult{}, fmt.Errorf("tool %s not found: %w", fc.Name, err)
	}

	// Convert map[string]interface{} to map[string]any
	args := make(map[string]any)
	for k, v := range fc.Args {
		args[k] = v
	}

	return tool.Execute(args)
}

// SendMessageStream generates content using the Gemini API and streams responses.
func (gc *GeminiChat) SendMessageStream(modelName string, messageParams types.MessageParams, promptId string) (<-chan types.StreamResponse, error) {
	respChan := make(chan types.StreamResponse)

	cs := gc.model.StartChat()
	cs.History = gc.startHistory

	// Prepare tools for the model
	if gc.toolRegistry != nil {
		gc.model.Tools = gc.toolRegistry.GetTools()
	}

	// Convert types.Part to genai.Part
	genaiParts := make([]genai.Part, len(messageParams.Message))
	for i, part := range messageParams.Message {
		if part.Text != "" {
			genaiParts[i] = genai.Text(part.Text)
		} else if part.FunctionResponse != nil {
			genaiParts[i] = genai.FunctionResponse{
				Name:     part.FunctionResponse.Name,
				Response: part.FunctionResponse.Response,
			}
		} else if part.InlineData != nil {
			genaiParts[i] = genai.Blob{
				MIMEType: part.InlineData.MimeType,
				Data:     []byte(part.InlineData.Data),
			}
		} else if part.FileData != nil {
			genaiParts[i] = genai.Text(fmt.Sprintf("File data: %s (%s)", part.FileData.FileURL, part.FileData.MimeType))
		}
	}

	go func() {
		defer close(respChan)

		iter := cs.SendMessageStream(messageParams.AbortSignal, genaiParts...)
		for {
			resp, err := iter.Next()
			if err == iterator.Done {
				return
			}
			if err != nil {
				respChan <- types.StreamResponse{Type: types.StreamEventTypeError, Error: err}
				return
			}
			respChan <- types.StreamResponse{Type: types.StreamEventTypeChunk, Value: resp}
		}
	}()

	return respChan, nil
}

// ListModels lists available Gemini models.
func (gc *GeminiChat) ListModels() ([]string, error) {
	ctx := context.Background()

	var modelNames []string
	it := gc.client.ListModels(ctx)
	for {
		model, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list models: %w", err)
		}
		modelNames = append(modelNames, model.Name)
	}
	return modelNames, nil
}

// GetHistory returns the current chat history.
func (gc *GeminiChat) GetHistory() ([]*genai.Content, error) {
	// For now, return the initial history. A more complete implementation
	// would track the full conversation history.
	return gc.startHistory, nil
}

// SetHistory sets the chat history.
func (gc *GeminiChat) SetHistory(history []*genai.Content) error {
	gc.startHistory = history
	return nil
}

// CompressChat compresses the chat history by replacing it with a summary.
func (gc *GeminiChat) CompressChat(promptId string, force bool) (*types.ChatCompressionResult, error) {
	ctx := context.Background() // Define ctx here

	fmt.Printf("DEBUG: len(gc.startHistory) = %d\n", len(gc.startHistory))

	// Convert []*genai.Content to []genai.Part for token counting
	var originalHistoryParts []genai.Part
	for _, content := range gc.startHistory {
		originalHistoryParts = append(originalHistoryParts, content.Parts...)
	}

	// Implement actual token counting for original history
	originalTokenCountResp, err := gc.model.CountTokens(ctx, originalHistoryParts...)
	if err != nil {
		return nil, fmt.Errorf("failed to count tokens for original history: %w", err)
	}
	originalTokenCount := originalTokenCountResp.TotalTokens

	if len(gc.startHistory) <= 2 { // Only compress if there's a user-initiated conversation
		return nil, fmt.Errorf("no conversation found to compress")
	}

	// Construct a prompt to summarize the chat history
	summaryPrompt := "Summarize the following conversation:\n\n"
	for _, content := range gc.startHistory {
		for _, part := range content.Parts {
			if text, ok := part.(genai.Text); ok {
				summaryPrompt += string(text) + "\n"
			}
		}
	}

	resp, err := gc.model.GenerateContent(
		ctx,
		genai.Text(summaryPrompt),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate summary: %w", err)
	}

	generatedSummary := ""
	for _, part := range resp.Candidates[0].Content.Parts {
		if text, ok := part.(genai.Text); ok {
			generatedSummary += string(text)
		}
	}

	// Replace the history with the generated summary
	gc.startHistory = []*genai.Content{
		{Role: "model", Parts: []genai.Part{genai.Text(generatedSummary)}},
	}

	// Convert []*genai.Content to []genai.Part for token counting
	var newHistoryParts []genai.Part
	for _, content := range gc.startHistory {
		newHistoryParts = append(newHistoryParts, content.Parts...)
	}

	// Implement actual token counting for new history
	newTokenCountResp, err := gc.model.CountTokens(ctx, newHistoryParts...)
	if err != nil {
		return nil, fmt.Errorf("failed to count tokens for new history: %w", err)
	}
	newTokenCount := newTokenCountResp.TotalTokens

	return &types.ChatCompressionResult{
		OriginalTokenCount: int(originalTokenCount), // Cast to int
		NewTokenCount:      int(newTokenCount),      // Cast to int
		CompressionStatus:  "success",
	}, nil
}
