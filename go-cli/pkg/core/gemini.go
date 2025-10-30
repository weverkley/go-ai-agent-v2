package core

import (
	"context"
	"fmt"
	"os"

	"go-ai-agent-v2/go-cli/pkg/core/agents"
	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/tools"

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
	generationConfig agents.GenerateContentConfig
	startHistory     []genai.Content
	// toolRegistry *tools.ToolRegistry // Removed, passed directly to SendMessageStream
}

// NewGeminiChat creates a new GeminiChat instance.
func NewGeminiChat(cfg *config.Config, generationConfig agents.GenerateContentConfig, startHistory []genai.Content) (*GeminiChat, error) {
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
	// TODO: Handle ThinkingConfig and SystemInstruction

	return &GeminiChat{
		client:           client,
		model:            model,
		generationConfig: generationConfig,
		startHistory:     startHistory,
	}, nil
}

// SendMessageStream generates content using the Gemini API and streams responses.
func (gc *GeminiChat) SendMessageStream(modelName string, messageParams agents.MessageParams, promptId string) (<-chan agents.StreamResponse, error) {
	respChan := make(chan agents.StreamResponse)

	cs := gc.model.StartChat()

	// Convert []genai.Content to []*genai.Content for history
	historyPtrs := make([]*genai.Content, len(gc.startHistory))
	for i := range gc.startHistory {
		historyPtrs[i] = &gc.startHistory[i]
	}
	cs.History = historyPtrs

	// Prepare tools for the model
	if len(messageParams.Tools) > 0 {
		genaiTools := make([]*genai.Tool, len(messageParams.Tools))
		for i, tool := range messageParams.Tools {
			genaiTools[i] = &genai.Tool{
				FunctionDeclarations: []*genai.FunctionDeclaration{&tool},
			}
		}
		gc.model.Tools = genaiTools
	}

	// Convert agents.Part to genai.Part
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
			// Handle FileData if necessary, currently not directly supported by genai.Part
			// For now, we'll skip or convert to text if possible.
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
				respChan <- agents.StreamResponse{Type: agents.StreamEventTypeError, Error: err}
				return
			}
			respChan <- agents.StreamResponse{Type: agents.StreamEventTypeChunk, Value: resp}
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
