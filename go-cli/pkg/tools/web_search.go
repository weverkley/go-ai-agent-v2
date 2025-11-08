package tools

import (
	"context"
	"fmt"
	"os"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/types"

	"google.golang.org/api/customsearch/v1"
	"google.golang.org/api/option"
)

// WebSearchTool represents the web-search tool.
type WebSearchTool struct {
	*types.BaseDeclarativeTool
}

// NewWebSearchTool creates a new instance of WebSearchTool.
func NewWebSearchTool() *WebSearchTool {
	return &WebSearchTool{
		types.NewBaseDeclarativeTool(
			"web_search",
			"web_search",
			"Performs a web search using Google Search (via the Gemini API) and returns the results. This tool is useful for finding information on the internet based on a query.",
			types.KindOther, // Assuming KindOther for now
			types.JsonSchemaObject{
				Type: "object",
				Properties: map[string]types.JsonSchemaProperty{
					"query": {
						Type:        "string",
						Description: "The search query to find information on the web.",
					},
				},
				Required: []string{"query"},
			},
			false, // isOutputMarkdown
			false, // canUpdateOutput
			nil,   // MessageBus
		),
	}
}

// Execute performs a web search operation.
func (t *WebSearchTool) Execute(args map[string]any) (types.ToolResult, error) {
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return types.ToolResult{}, fmt.Errorf("invalid or missing 'query' argument")
	}

	apiKey := os.Getenv("GOOGLE_API_KEY")
	cx := os.Getenv("GOOGLE_CUSTOM_SEARCH_CX")

	if apiKey == "" || cx == "" {
		return types.ToolResult{}, fmt.Errorf("GOOGLE_API_KEY and GOOGLE_CUSTOM_SEARCH_CX environment variables must be set for web search")
	}

	ctx := context.Background()
	svc, err := customsearch.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return types.ToolResult{}, fmt.Errorf("failed to create customsearch service: %w", err)
	}

	resp, err := svc.Cse.List().Q(query).Cx(cx).Do()
	if err != nil {
		return types.ToolResult{}, fmt.Errorf("failed to perform custom search: %w", err)
	}

	var llmContent strings.Builder
	llmContent.WriteString(fmt.Sprintf("Web search results for \"%s\":\n\n", query))

	if len(resp.Items) == 0 {
		return types.ToolResult{
			LLMContent:    fmt.Sprintf("No search results found for query: \"%s\"", query),
			ReturnDisplay: fmt.Sprintf("No search results found for query: \"%s\"", query),
		}, nil
	}

	for i, item := range resp.Items {
		llmContent.WriteString(fmt.Sprintf("[%d] %s (%s)\n", i+1, item.Title, item.Link))
		llmContent.WriteString(fmt.Sprintf("   %s\n", item.Snippet))
	}

	resultMessage := llmContent.String()
	return types.ToolResult{
		LLMContent:    resultMessage,
		ReturnDisplay: resultMessage,
	}, nil
}
