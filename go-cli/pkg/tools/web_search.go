package tools

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/customsearch/v1"
	"google.golang.org/api/option"
)

// WebSearchTool represents the web-search tool.
type WebSearchTool struct{}

// NewWebSearchTool creates a new instance of WebSearchTool.
func NewWebSearchTool() *WebSearchTool {
	return &WebSearchTool{}
}

// Name returns the name of the tool.
func (t *WebSearchTool) Name() string {
	return "web_search"
}

// Definition returns the tool's definition for the Gemini API.
func (t *WebSearchTool) Definition() *genai.Tool {
	return &genai.Tool{
		FunctionDeclarations: []*genai.FunctionDeclaration{
			{
				Name:        t.Name(),
				Description: "Performs a web search using Google Search and returns the results.",
				Parameters: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"query": {
							Type:        genai.TypeString,
							Description: "The search query to find information on the web.",
						},
					},
					Required: []string{"query"},
				},
			},
		},
	}
}

// Execute performs a web search operation.
func (t *WebSearchTool) Execute(args map[string]any) (string, error) {
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return "", fmt.Errorf("invalid or missing 'query' argument")
	}

	apiKey := os.Getenv("GOOGLE_API_KEY")
	cx := os.Getenv("GOOGLE_CUSTOM_SEARCH_CX")

	if apiKey == "" || cx == "" {
		return "", fmt.Errorf("GOOGLE_API_KEY and GOOGLE_CUSTOM_SEARCH_CX environment variables must be set for web search")
	}

	ctx := context.Background()
	svc, err := customsearch.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return "", fmt.Errorf("failed to create customsearch service: %w", err)
	}

	resp, err := svc.Cse.List().Q(query).Cx(cx).Do()
	if err != nil {
		return "", fmt.Errorf("failed to perform custom search: %w", err)
	}

	var llmContent strings.Builder
	llmContent.WriteString(fmt.Sprintf("Web search results for \"%s\":\n\n", query))

	if len(resp.Items) == 0 {
		return fmt.Sprintf("No search results found for query: \"%s\"", query), nil
	}

	for i, item := range resp.Items {
		llmContent.WriteString(fmt.Sprintf("[%d] %s (%s)\n", i+1, item.Title, item.Link))
		llmContent.WriteString(fmt.Sprintf("   %s\n", item.Snippet))
	}

	return llmContent.String(), nil
}
