package core

import (
	"context"
	"fmt"
	"os"

	"go-ai-agent-v2/go-cli/pkg/telemetry"
	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// ContentGenerator interface represents the ability to generate content.
type ContentGenerator interface {
	GenerateContent(prompt string) (string, error)
}

// GoaiagentChat represents a Go AI Agent chat client.
type GoaiagentChat struct {
	client                *genai.Client
	model                 *genai.GenerativeModel
	Name                  string
	generationConfig      types.GenerateContentConfig
	startHistory          []*types.Content // Changed to generic type
	toolRegistry          types.ToolRegistryInterface
	toolCallCounter       int
	userConfirmationChan  chan bool
	ToolConfirmationChan  chan types.ToolConfirmationOutcome // New channel for rich confirmation
}

// NewGoaiagentChat creates a new GoaiagentChat instance.
func NewGoaiagentChat(cfg types.Config, generationConfig types.GenerateContentConfig, startHistory []*types.Content) (Executor, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable not set")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Go AI Agent client: %w", err)
	}

	modelVal, found := cfg.Get("model")
	if !found || modelVal == nil {
		return nil, fmt.Errorf("model name not found in config")
	}
	modelName, ok := modelVal.(string)
	if !ok {
		return nil, fmt.Errorf("model name in config is not a string")
	}
	model := client.GenerativeModel(modelName)

	model.SetTemperature(generationConfig.Temperature)
	model.SetTopP(generationConfig.TopP)

	var geminiChatToolRegistry types.ToolRegistryInterface
	if toolRegistryVal, ok := cfg.Get("toolRegistry"); ok && toolRegistryVal != nil {
		if tr, toolRegistryOk := toolRegistryVal.(types.ToolRegistryInterface); toolRegistryOk {
			geminiChatToolRegistry = tr
		} else {
			return nil, fmt.Errorf("tool registry in config is not of expected type types.ToolRegistryInterface")
		}
	}

	return &GoaiagentChat{
		client:                client,
		model:                 model,
		Name:                  modelName,
		generationConfig:      generationConfig,
		startHistory:          startHistory,
		toolRegistry:          geminiChatToolRegistry,
		toolCallCounter:       0,
		userConfirmationChan:  make(chan bool, 1),
		ToolConfirmationChan:  make(chan types.ToolConfirmationOutcome, 1), // Initialize
	}, nil
}

// toGenaiContent converts a generic *types.Content to *genai.Content.
func toGenaiContent(content *types.Content) *genai.Content {
	if content == nil {
		return nil
	}
	genaiParts := make([]genai.Part, len(content.Parts))
	for i, part := range content.Parts {
		var genaiPart genai.Part
		if part.Text != "" {
			genaiPart = genai.Text(part.Text)
		} else if part.FunctionCall != nil {
			genaiPart = &genai.FunctionCall{
				Name: part.FunctionCall.Name,
				Args: part.FunctionCall.Args,
			}
		} else if part.FunctionResponse != nil {
			genaiPart = &genai.FunctionResponse{
				Name:     part.FunctionResponse.Name,
				Response: part.FunctionResponse.Response,
			}
		} else if part.InlineData != nil {
			genaiPart = genai.Blob{
				MIMEType: part.InlineData.MimeType,
				Data:     []byte(part.InlineData.Data),
			}
		}
		genaiParts[i] = genaiPart
	}
	return &genai.Content{Role: content.Role, Parts: genaiParts}
}

// fromGenaiContent converts a *genai.Content to a generic *types.Content.
func fromGenaiContent(content *genai.Content) *types.Content {
	if content == nil {
		return nil
	}
	genericParts := make([]types.Part, len(content.Parts))
	for i, part := range content.Parts {
		var genericPart types.Part
		switch p := part.(type) {
		case genai.Text:
			genericPart.Text = string(p)
		case *genai.FunctionCall:
			genericPart.FunctionCall = &types.FunctionCall{
				Name: p.Name,
				Args: p.Args,
			}
		case *genai.FunctionResponse:
			genericPart.FunctionResponse = &types.FunctionResponse{
				Name:     p.Name,
				Response: p.Response,
			}
		case genai.Blob:
			genericPart.InlineData = &types.InlineData{
				MimeType: p.MIMEType,
				Data:     string(p.Data),
			}
		}
		genericParts[i] = genericPart
	}
	return &types.Content{Role: content.Role, Parts: genericParts}
}

// toGenaiToolsFromToolInterface converts generic []types.Tool to []*genai.Tool.
func toGenaiToolsFromToolInterface(tools []types.Tool) []*genai.Tool {
	if tools == nil {
		return nil
	}
	genaiTools := make([]*genai.Tool, len(tools))
	for i, tool := range tools {
		genaiTools[i] = &genai.Tool{
			FunctionDeclarations: []*genai.FunctionDeclaration{
				{
					Name:        tool.Name(),
					Description: tool.Description(),
					Parameters:  toGenaiSchema(tool.Parameters()),
				},
			},
		}
	}
	return genaiTools
}

// toGenaiTools converts generic []*types.ToolDefinition to []*genai.Tool.
func toGenaiTools(tools []*types.ToolDefinition) []*genai.Tool {
	if tools == nil {
		return nil
	}
	genaiTools := make([]*genai.Tool, len(tools))
	for i, tool := range tools {
		genaiDeclarations := make([]*genai.FunctionDeclaration, len(tool.FunctionDeclarations))
		for j, decl := range tool.FunctionDeclarations {
			genaiDeclarations[j] = &genai.FunctionDeclaration{
				Name:        decl.Name,
				Description: decl.Description,
				Parameters:  toGenaiSchema(decl.Parameters),
			}
		}
		genaiTools[i] = &genai.Tool{FunctionDeclarations: genaiDeclarations}
	}
	return genaiTools
}

// toGenaiSchema converts a generic *types.JsonSchemaObject to *genai.Schema.
func toGenaiSchema(schema *types.JsonSchemaObject) *genai.Schema {
	if schema == nil {
		return nil
	}
	properties := make(map[string]*genai.Schema)
	for k, v := range schema.Properties {
		properties[k] = &genai.Schema{
			Type:        toGenaiType(v.Type),
			Description: v.Description,
			Items:       toGenaiSchema(v.Items),
			Enum:        v.Enum,
		}
	}
	return &genai.Schema{
		Type:       toGenaiType(schema.Type),
		Properties: properties,
		Required:   schema.Required,
	}
}

func toGenaiType(t string) genai.Type {
	switch t {
	case "string":
		return genai.TypeString
	case "number":
		return genai.TypeNumber
	case "integer":
		return genai.TypeInteger
	case "boolean":
		return genai.TypeBoolean
	case "array":
		return genai.TypeArray
	case "object":
		return genai.TypeObject
	default:
		return genai.TypeString
	}
}

// GenerateContent generates content using the Gemini API.
func (gc *GoaiagentChat) GenerateContent(contents ...*types.Content) (*types.GenerateContentResponse, error) {
	ctx := context.Background()

	var parts []genai.Part
	for _, content := range contents {
		genaiContent := toGenaiContent(content)
		parts = append(parts, genaiContent.Parts...)
	}

	resp, err := gc.model.GenerateContent(ctx, parts...)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	// Convert genai response to generic response
	genericResp := &types.GenerateContentResponse{
		Candidates: make([]*types.Candidate, len(resp.Candidates)),
	}
	for i, cand := range resp.Candidates {
		genericResp.Candidates[i] = &types.Candidate{
			Content: fromGenaiContent(cand.Content),
		}
	}

	return genericResp, nil
}

// ExecuteTool executes a tool call.
func (gc *GoaiagentChat) ExecuteTool(ctx context.Context, fc *types.FunctionCall) (types.ToolResult, error) {
	if gc.toolRegistry == nil {
		return types.ToolResult{}, fmt.Errorf("tool registry not initialized")
	}

	tool, err := gc.toolRegistry.GetTool(fc.Name)
	if err != nil {
		return types.ToolResult{}, fmt.Errorf("tool %s not found: %w", fc.Name, err)
	}

	return tool.Execute(ctx, fc.Args)
}

// SendMessageStream generates content using the Gemini API and streams responses.
func (gc *GoaiagentChat) SendMessageStream(modelName string, messageParams types.MessageParams, promptId string) (<-chan types.StreamResponse, error) {
	respChan := make(chan types.StreamResponse)

	// Convert generic history to genai history for the chat session
	genaiHistory := make([]*genai.Content, len(gc.startHistory))
	for i, c := range gc.startHistory {
		genaiHistory[i] = toGenaiContent(c)
	}

	cs := gc.model.StartChat()
	cs.History = genaiHistory

	// Convert generic tools to genai tools
	gc.model.Tools = toGenaiTools(messageParams.Tools)

	// Convert generic parts to genai parts
	genaiParts := make([]genai.Part, len(messageParams.Message))
	for i, part := range messageParams.Message {
		var genaiPart genai.Part
		if part.Text != "" {
			genaiPart = genai.Text(part.Text)
		} else if part.FunctionResponse != nil {
			genaiPart = &genai.FunctionResponse{
				Name:     part.FunctionResponse.Name,
				Response: part.FunctionResponse.Response,
			}
		} else if part.InlineData != nil {
			genaiPart = genai.Blob{
				MIMEType: part.InlineData.MimeType,
				Data:     []byte(part.InlineData.Data),
			}
		}
		genaiParts[i] = genaiPart
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

			// Convert genai response to generic response
			genericResp := &types.GenerateContentResponse{
				Candidates: make([]*types.Candidate, len(resp.Candidates)),
			}
			for i, cand := range resp.Candidates {
				genericResp.Candidates[i] = &types.Candidate{
					Content: fromGenaiContent(cand.Content),
				}
			}
			respChan <- types.StreamResponse{Type: types.StreamEventTypeChunk, Value: genericResp}
		}
	}()

	return respChan, nil
}

// ListModels lists available Gemini models.
func (gc *GoaiagentChat) ListModels() ([]string, error) {
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
func (gc *GoaiagentChat) GetHistory() ([]*types.Content, error) {
	return gc.startHistory, nil
}

// SetHistory sets the chat history.
func (gc *GoaiagentChat) SetHistory(history []*types.Content) error {
	gc.startHistory = history
	return nil
}

// CompressChat compresses the chat history by replacing it with a summary.
func (gc *GoaiagentChat) CompressChat(promptId string, force bool) (*types.ChatCompressionResult, error) {
	ctx := context.Background()

	telemetry.LogDebugf("CompressChat: len(gc.startHistory) = %d", len(gc.startHistory))

	genaiHistory := make([]*genai.Content, len(gc.startHistory))
	for i, c := range gc.startHistory {
		genaiHistory[i] = toGenaiContent(c)
	}

	var originalHistoryParts []genai.Part
	for _, content := range genaiHistory {
		originalHistoryParts = append(originalHistoryParts, content.Parts...)
	}

	originalTokenCountResp, err := gc.model.CountTokens(ctx, originalHistoryParts...)
	if err != nil {
		return nil, fmt.Errorf("failed to count tokens for original history: %w", err)
	}
	originalTokenCount := originalTokenCountResp.TotalTokens

	if len(gc.startHistory) <= 2 {
		return nil, fmt.Errorf("no conversation found to compress")
	}

	summaryPrompt := "Summarize the following conversation:\n\n"
	for _, content := range gc.startHistory {
		for _, part := range content.Parts {
			if part.Text != "" {
				summaryPrompt += part.Text + "\n"
			}
		}
	}

	resp, err := gc.model.GenerateContent(ctx, genai.Text(summaryPrompt))
	if err != nil {
		return nil, fmt.Errorf("failed to generate summary: %w", err)
	}

	genericResp := fromGenaiContent(resp.Candidates[0].Content)
	generatedSummary := ""
	if len(genericResp.Parts) > 0 {
		generatedSummary = genericResp.Parts[0].Text
	}

	gc.startHistory = []*types.Content{
		{Role: "model", Parts: []types.Part{{Text: generatedSummary}}},
	}

	newGenaiHistory := toGenaiContent(gc.startHistory[0])
	newTokenCountResp, err := gc.model.CountTokens(ctx, newGenaiHistory.Parts...)
	if err != nil {
		return nil, fmt.Errorf("failed to count tokens for new history: %w", err)
	}
	newTokenCount := newTokenCountResp.TotalTokens

	return &types.ChatCompressionResult{
		OriginalTokenCount: int(originalTokenCount),
		NewTokenCount:      int(newTokenCount),
		CompressionStatus:  "success",
	}, nil
}

// SetUserConfirmationChannel sets the channel for user confirmation.
func (gc *GoaiagentChat) SetUserConfirmationChannel(ch chan bool) {
	gc.userConfirmationChan = ch
}

// SetToolConfirmationChannel sets the channel for tool confirmation.
func (gc *GoaiagentChat) SetToolConfirmationChannel(ch chan types.ToolConfirmationOutcome) {
	gc.ToolConfirmationChan = ch
}

func toGenaiParts(parts []types.Part) []genai.Part {
	genaiParts := make([]genai.Part, len(parts))
	for i, part := range parts {
		var genaiPart genai.Part
		if part.Text != "" {
			genaiPart = genai.Text(part.Text)
		} else if part.FunctionCall != nil {
			genaiPart = &genai.FunctionCall{
				Name: part.FunctionCall.Name,
				Args: part.FunctionCall.Args,
			}
		} else if part.FunctionResponse != nil {
			genaiPart = &genai.FunctionResponse{
				Name:     part.FunctionResponse.Name,
				Response: part.FunctionResponse.Response,
			}
		} else if part.InlineData != nil {
			genaiPart = genai.Blob{
				MIMEType: part.InlineData.MimeType,
				Data:     []byte(part.InlineData.Data),
			}
		}
		genaiParts[i] = genaiPart
	}
	return genaiParts
}

// toGenaiContents converts a slice of generic *types.Content to []*genai.Content.
func toGenaiContents(contents []*types.Content) []*genai.Content {
	genaiContents := make([]*genai.Content, len(contents))
	for i, c := range contents {
		genaiContents[i] = toGenaiContent(c)
	}
	return genaiContents
}

// fromGenaiFunctionCall converts a *genai.FunctionCall to a generic *types.FunctionCall.
func fromGenaiFunctionCall(fc *genai.FunctionCall) *types.FunctionCall {
	if fc == nil {
		return nil
	}
	return &types.FunctionCall{
		Name: fc.Name,
		Args: fc.Args,
	}
}

// StreamContent sends the chat history to the model and streams back response parts.
func (gc *GoaiagentChat) StreamContent(ctx context.Context, history ...*types.Content) (<-chan any, error) {
	telemetry.LogDebugf("GoaiagentChat.StreamContent called")
	eventChan := make(chan any)

	go func() {
		defer close(eventChan)

		cs := gc.model.StartChat()
		cs.History = toGenaiContents(history[:len(history)-1])
		lastMessage := history[len(history)-1]

		iter := cs.SendMessageStream(ctx, toGenaiParts(lastMessage.Parts)...)

		for {
			resp, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				telemetry.LogErrorf("Error receiving from Gemini stream: %v", err)
				eventChan <- types.ErrorEvent{Err: err}
				return
			}

			if resp != nil && len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
				for _, part := range resp.Candidates[0].Content.Parts {
					eventChan <- fromGenaiPart(part)
				}
			}
		}
	}()

	return eventChan, nil
}

// fromGenaiPart converts a single genai.Part to a types.Part.
func fromGenaiPart(part genai.Part) types.Part {
	var genericPart types.Part
	switch p := part.(type) {
	case genai.Text:
		genericPart.Text = string(p)
	case *genai.FunctionCall:
		genericPart.FunctionCall = fromGenaiFunctionCall(p)
	case *genai.FunctionResponse:
		genericPart.FunctionResponse = &types.FunctionResponse{
			Name:     p.Name,
			Response: p.Response,
		}
	case genai.Blob:
		genericPart.InlineData = &types.InlineData{
			MimeType: p.MIMEType,
			Data:     string(p.Data),
		}
	}
	return genericPart
}
// GenerateContentWithTools generates content using the Gemini API, including tools.
func (gc *GoaiagentChat) GenerateContentWithTools(ctx context.Context, history []*types.Content, tools []types.Tool) (*types.GenerateContentResponse, error) {
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

	gc.model.Tools = toGenaiTools(toolDefinitions)

	cs := gc.model.StartChat()

	if len(history) == 0 {
		return nil, fmt.Errorf("history cannot be empty")
	}

	genaiHistory := make([]*genai.Content, len(history))
	for i, c := range history {
		genaiHistory[i] = toGenaiContent(c)
	}

	if len(genaiHistory) > 1 {
		cs.History = genaiHistory[:len(genaiHistory)-1]
	}

	lastMessage := genaiHistory[len(genaiHistory)-1]
	resp, err := cs.SendMessage(ctx, lastMessage.Parts...)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content with tools: %w", err)
	}
	
	// Convert genai response to generic response
	genericResp := &types.GenerateContentResponse{
		Candidates: make([]*types.Candidate, len(resp.Candidates)),
	}
	for i, cand := range resp.Candidates {
		genericResp.Candidates[i] = &types.Candidate{
			Content: fromGenaiContent(cand.Content),
		}
	}

	return genericResp, nil
}
