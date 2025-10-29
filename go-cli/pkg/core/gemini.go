package core

import (
	"context"
	"fmt"
	"os"

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
	client *genai.Client
	model  *genai.GenerativeModel
}

// NewGeminiChat creates a new GeminiChat instance.
func NewGeminiChat() (*GeminiChat, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable not set")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	model := client.GenerativeModel("gemini-pro-latest") // Using gemini-pro-latest model

	return &GeminiChat{client: client, model: model}, nil
}

// GenerateContent generates content using the Gemini API.
func (gc *GeminiChat) GenerateContent(prompt string) (string, error) {
	ctx := context.Background()

	resp, err := gc.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content generated")
	}

	var generatedText string
	for _, part := range resp.Candidates[0].Content.Parts {
		if txt, ok := part.(genai.Text);
			ok {
			generatedText += string(txt)
		}
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
