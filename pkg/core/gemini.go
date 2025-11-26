package core

import (
	"context"
	"fmt"
	"os"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/prompts"
	"go-ai-agent-v2/go-cli/pkg/telemetry"
	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/google/generative-ai-go/genai"
	"github.com/pkoukk/tiktoken-go"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// GeminiChat represents a Gemini AI client.
type GeminiChat struct {
	client               *genai.Client
	model                *genai.GenerativeModel
	modelName            string
	generationConfig     types.GenerateContentConfig
	toolRegistry         types.ToolRegistryInterface
	userConfirmationChan chan bool
	ToolConfirmationChan chan types.ToolConfirmationOutcome
	logger               telemetry.TelemetryLogger
}

func NewGeminiChat(cfg types.Config, generationConfig types.GenerateContentConfig, startHistory []*types.Content, logger telemetry.TelemetryLogger) (Executor, error) {
	logger.LogDebugf("NewGeminiChat: Initializing...")
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		logger.LogErrorf("NewGeminiChat: GEMINI_API_KEY environment variable not set")
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable not set")
	}
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		logger.LogErrorf("NewGeminiChat: Failed to create genai.NewClient: %v", err)
		return nil, fmt.Errorf("failed to create Go AI Agent client: %w", err)
	}
	modelVal, _ := cfg.Get("model")
	modelName, _ := modelVal.(string)
	model := client.GenerativeModel(modelName)
	model.SetTemperature(generationConfig.Temperature)
	model.SetTopP(generationConfig.TopP)
	if generationConfig.SystemInstruction != "" {
		model.SystemInstruction = &genai.Content{
			Parts: []genai.Part{genai.Text(generationConfig.SystemInstruction)},
		}
	}

	var toolRegistry types.ToolRegistryInterface
	if toolRegistryVal, ok := cfg.Get("toolRegistry"); ok && toolRegistryVal != nil {
		if tr, ok := toolRegistryVal.(types.ToolRegistryInterface); ok {
			toolRegistry = tr
		}
	}
	logger.LogDebugf("NewGeminiChat: Initialization complete for model '%s'.", modelName)
	return &GeminiChat{
		client:               client,
		model:                model,
		modelName:            modelName,
		generationConfig:     generationConfig,
		toolRegistry:         toolRegistry,
		userConfirmationChan: make(chan bool, 1),
		ToolConfirmationChan: make(chan types.ToolConfirmationOutcome, 1),
		logger:               logger,
	}, nil
}

func toGenaiContent(content *types.Content) *genai.Content {
	if content == nil {
		return nil
	}
	genaiParts := make([]genai.Part, 0, len(content.Parts))
	for _, part := range content.Parts {
		var genaiPart genai.Part
		if part.Text != "" {
			genaiPart = genai.Text(part.Text)
		} else if part.FunctionCall != nil {
			genaiPart = &genai.FunctionCall{Name: part.FunctionCall.Name, Args: part.FunctionCall.Args}
		} else if part.FunctionResponse != nil {
			genaiPart = &genai.FunctionResponse{Name: part.FunctionResponse.Name, Response: part.FunctionResponse.Response}
		}
		if genaiPart != nil {
			genaiParts = append(genaiParts, genaiPart)
		}
	}
	return &genai.Content{Role: content.Role, Parts: genaiParts}
}

func toGenaiContents(contents []*types.Content) []*genai.Content {
	var genaiContents []*genai.Content
	for _, c := range contents {
		genaiContents = append(genaiContents, toGenaiContent(c))
	}
	return genaiContents
}

func fromGenaiPart(part genai.Part, logger telemetry.TelemetryLogger) types.Part {
	logger.LogDebugf("GeminiExecutor: Processing a response part of type %T", part)
	var genericPart types.Part
	switch p := part.(type) {
	case genai.Text:
		genericPart.Text = string(p)
	case genai.FunctionCall:
		genericPart.FunctionCall = &types.FunctionCall{Name: p.Name, Args: p.Args}
	case *genai.FunctionResponse:
		genericPart.FunctionResponse = &types.FunctionResponse{Name: p.Name, Response: p.Response}
	case genai.Blob:
		genericPart.InlineData = &types.InlineData{MimeType: p.MIMEType, Data: string(p.Data)}
	default:
		logger.LogWarnf("GeminiExecutor: Unhandled genai.Part type: %T", p)
	}
	return genericPart
}

func buildGeminiTools(toolRegistry types.ToolRegistryInterface, logger telemetry.TelemetryLogger) []*genai.Tool {
	logger.LogDebugf("buildGeminiTools: Building tools from registry.")
	if toolRegistry == nil {
		logger.LogWarnf("buildGeminiTools: Tool registry is nil.")
		return nil
	}
	allTools := toolRegistry.GetAllTools()
	if allTools == nil {
		logger.LogWarnf("buildGeminiTools: No tools found in registry.")
		return nil
	}

	// ... (rest of function is unchanged)
	var convertObject func(obj *types.JsonSchemaObject) *genai.Schema
	var convertProperty func(prop *types.JsonSchemaProperty) *genai.Schema
	var toGenaiType func(t string) genai.Type

	toGenaiType = func(t string) genai.Type {
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

	convertProperty = func(prop *types.JsonSchemaProperty) *genai.Schema {
		if prop == nil {
			return nil
		}
		s := &genai.Schema{
			Type:        toGenaiType(prop.Type),
			Description: prop.Description,
			Enum:        prop.Enum,
		}
		if prop.Type == "object" {
			props := make(map[string]*genai.Schema)
			for k, v := range prop.Properties {
				props[k] = convertProperty(v)
			}
			s.Properties = props
			s.Required = prop.Required
		}
		if prop.Type == "array" {
			s.Items = convertObject(prop.Items)
		}
		return s
	}

	convertObject = func(obj *types.JsonSchemaObject) *genai.Schema {
		if obj == nil {
			return nil
		}
		props := make(map[string]*genai.Schema)
		for k, v := range obj.Properties {
			props[k] = convertProperty(v)
		}
		return &genai.Schema{
			Type:       toGenaiType(obj.Type),
			Properties: props,
			Required:   obj.Required,
		}
	}

	var genaiTools []*genai.Tool
	for _, tool := range allTools {
		params := convertObject(tool.Parameters())
		genaiTools = append(genaiTools, &genai.Tool{
			FunctionDeclarations: []*genai.FunctionDeclaration{
				{
					Name:        tool.Name(),
					Description: tool.Description(),
					Parameters:  params,
				},
			},
		})
	}
	logger.LogDebugf("buildGeminiTools: Finished building %d tools.", len(genaiTools))
	return genaiTools
}

func (gc *GeminiChat) StreamContent(ctx context.Context, history ...*types.Content) (<-chan any, error) {
	eventChan := make(chan any)

			go func() {
			defer close(eventChan)
			gc.logger.LogDebugf("GeminiExecutor: StreamContent goroutine started.")
	
			// Use the model from the struct, which has been configured.
			if gc.toolRegistry != nil {
				gc.model.Tools = buildGeminiTools(gc.toolRegistry, gc.logger)
			}
	
			cs := gc.model.StartChat()
	
			if len(history) > 1 {
				historyToSet := history[:len(history)-1]
				gc.logger.LogDebugf("GeminiExecutor: Setting chat history with %d previous messages.", len(historyToSet))
				cs.History = toGenaiContents(historyToSet)
			} else {
				gc.logger.LogDebugf("GeminiExecutor: No previous history to set.")
			}
	
			// Count input tokens
			var inputText strings.Builder
			for _, content := range history {
				for _, part := range content.Parts {
					inputText.WriteString(part.Text)
				}
			}
			tke, err := tiktoken.GetEncoding("cl100k_base")
			if err != nil {
				eventChan <- types.ErrorEvent{Err: fmt.Errorf("failed to get tiktoken encoding: %w", err)}
				return
			}
			inputTokens := len(tke.Encode(inputText.String(), nil, nil))
	
			var lastParts []genai.Part
			if len(history) > 0 {
				lastContent := history[len(history)-1]
				if convertedContent := toGenaiContent(lastContent); convertedContent != nil {
					lastParts = convertedContent.Parts
				}
			} else {
				gc.logger.LogErrorf("GeminiExecutor: StreamContent called with empty history.")
				eventChan <- types.ErrorEvent{Err: fmt.Errorf("StreamContent called with empty history")}
				return
			}
			gc.logger.LogDebugf("GeminiExecutor: Sending last message with %d parts.", len(lastParts))
	
			iter := cs.SendMessageStream(ctx, lastParts...)
			gc.logger.LogDebugf("GeminiExecutor: SendMessageStream called. Waiting for response...")
	
			var outputText strings.Builder
			for {
				resp, err := iter.Next()
				if err == iterator.Done {
					gc.logger.LogDebugf("GeminiExecutor: Stream iterator finished (Done).")
					break
				}
				if err != nil {
					gc.logger.LogErrorf("GeminiExecutor: Error receiving from Gemini stream: %v", err)
					eventChan <- types.ErrorEvent{Err: err}
					return
				}
				gc.logger.LogDebugf("GeminiExecutor: Received a response chunk from Gemini stream.")
	
				if resp != nil && len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
					for _, part := range resp.Candidates[0].Content.Parts {
						if text, ok := part.(genai.Text); ok {
							outputText.WriteString(string(text))
						}
						eventChan <- fromGenaiPart(part, gc.logger)
					}
				}
			}
			outputTokens := len(tke.Encode(outputText.String(), nil, nil))
			eventChan <- types.TokenCountEvent{InputTokens: inputTokens, OutputTokens: outputTokens}
			gc.logger.LogDebugf("GeminiExecutor: Finished processing Gemini stream.")
		}()
	
		return eventChan, nil
	}
// Methods to satisfy the Executor interface
func (gc *GeminiChat) SetUserConfirmationChannel(ch chan bool) {
	gc.userConfirmationChan = ch
}
func (gc *GeminiChat) SetToolConfirmationChannel(ch chan types.ToolConfirmationOutcome) {
	gc.ToolConfirmationChan = ch
}
func (gc *GeminiChat) Name() string { return gc.modelName }
func (gc *GeminiChat) GenerateContent(contents ...*types.Content) (*types.GenerateContentResponse, error) {
	return nil, fmt.Errorf("not implemented")
}
func (gc *GeminiChat) ExecuteTool(ctx context.Context, fc *types.FunctionCall) (types.ToolResult, error) {
	return types.ToolResult{}, fmt.Errorf("not implemented")
}
func (gc *GeminiChat) SendMessageStream(modelName string, messageParams types.MessageParams, promptId string) (<-chan types.StreamResponse, error) {
	return nil, fmt.Errorf("not implemented")
}
func (gc *GeminiChat) ListModels() ([]string, error) {
	return []string{gc.modelName}, nil
}
func (gc *GeminiChat) GetHistory() ([]*types.Content, error) {
	return nil, fmt.Errorf("history is managed by ChatService")
}
func (gc *GeminiChat) SetHistory(history []*types.Content) error {
	return fmt.Errorf("history is managed by ChatService")
}

func (gc *GeminiChat) GenerateContentWithTools(ctx context.Context, history []*types.Content, tools []types.Tool) (*types.GenerateContentResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

// CompressChat summarizes the chat history.
func (gc *GeminiChat) CompressChat(history []*types.Content, promptID string) (*types.ChatCompressionResult, error) {
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

	// 3. Count original tokens
	originalCount, err := gc.model.CountTokens(context.Background(), genai.Text(fullPrompt))
	if err != nil {
		return nil, fmt.Errorf("failed to count original tokens: %w", err)
	}
	inputTokens := int(originalCount.TotalTokens)

	// 4. Call the model to get the summary
	resp, err := gc.model.GenerateContent(context.Background(), genai.Text(fullPrompt))
	if err != nil {
		return nil, fmt.Errorf("failed to generate summary: %w", err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("received an empty summary response from the model")
	}

	summaryPart := resp.Candidates[0].Content.Parts[0]
	summaryText, ok := summaryPart.(genai.Text)
	if !ok {
		return nil, fmt.Errorf("unexpected response part type for summary: %T", summaryPart)
	}

	// 5. Count new tokens
	newCount, err := gc.model.CountTokens(context.Background(), summaryPart)
	if err != nil {
		return nil, fmt.Errorf("failed to count new tokens: %w", err)
	}
	outputTokens := int(newCount.TotalTokens)

	return &types.ChatCompressionResult{
		Summary:            string(summaryText),
		OriginalTokenCount: inputTokens,
		NewTokenCount:      outputTokens,
		InputTokens:        inputTokens,
		OutputTokens:       outputTokens,
		CompressionStatus:  "OK",
	}, nil
}

