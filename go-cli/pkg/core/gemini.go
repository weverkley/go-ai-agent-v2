package core

import (
	"context"
	"fmt"
	"os"

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
	client       *genai.Client
	model        *genai.GenerativeModel
	toolRegistry *tools.ToolRegistry
}

// NewGeminiChat creates a new GeminiChat instance.
func NewGeminiChat(registry *tools.ToolRegistry) (*GeminiChat, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable not set")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	model := client.GenerativeModel("gemini-pro-latest")
	if registry != nil && len(registry.GetTools()) > 0 {
		model.Tools = registry.GetTools()
	}

	return &GeminiChat{client: client, model: model, toolRegistry: registry}, nil
}

// GenerateContent generates content using the Gemini API, handling tool calls.
func (gc *GeminiChat) GenerateContent(prompt string) (string, error) {
	ctx := context.Background()
	chat := gc.model.StartChat()

	// Send initial prompt
	resp, err := chat.SendMessage(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to send initial message: %w", err)
	}

	var generatedText string
	for {
		if resp == nil || len(resp.Candidates) == 0 {
			break // No more responses
		}

		candidate := resp.Candidates[0]
		if candidate.Content == nil {
			break
		}

		// Aggregate text parts
		for _, part := range candidate.Content.Parts {
			if txt, ok := part.(genai.Text); ok {
				generatedText += string(txt)
			}
		}

		// Check for function calls
		if len(candidate.Content.Parts) > 0 {
			if fc, ok := candidate.Content.Parts[0].(genai.FunctionCall); ok {
				tool, err := gc.toolRegistry.GetTool(fc.Name)
				if err != nil {
					return "", fmt.Errorf("unknown tool called: %s", fc.Name)
				}

				toolOutput, err := tool.Execute(fc.Args)
				if err != nil {
					return "", fmt.Errorf("error executing tool %s: %w", fc.Name, err)
				}

				// Send tool output back to the model
				resp, err = chat.SendMessage(ctx, &genai.FunctionResponse{
					Name:     fc.Name,
					Response: map[string]any{"output": toolOutput},
				})
				if err != nil {
					return "", fmt.Errorf("failed to send tool response: %w", err)
				}
				continue // Continue the loop to process the model's response to the tool output
			}
		}

		// If there are no more function calls, we are done.
		break
	}

	return generatedText, nil
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
