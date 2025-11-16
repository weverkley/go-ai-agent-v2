package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/services"
	"go-ai-agent-v2/go-cli/pkg/types"

	"google.golang.org/api/customsearch/v1"
	"google.golang.org/api/option"
)

// tavilySearchResult represents a single search result from Tavily.
type tavilySearchResult struct {
	Title    string `json:"title"`
	URL      string `json:"url"`
	Content  string `json:"content"`
	RawContent string `json:"raw_content"`
}

// tavilyResponse represents the overall response structure from Tavily.
type tavilyResponse struct {
	Results []tavilySearchResult `json:"results"`
	Query   string               `json:"query"`
}

// WebSearchTool represents the web-search tool.
type WebSearchTool struct {
	*types.BaseDeclarativeTool
	settingsService *services.SettingsService
}

// NewWebSearchTool creates a new instance of WebSearchTool.
func NewWebSearchTool(settingsService *services.SettingsService) *WebSearchTool {
	return &WebSearchTool{
		BaseDeclarativeTool: types.NewBaseDeclarativeTool(
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
		settingsService: settingsService,
	}
}

func searchWithTavily(ctx context.Context, query string, apiKey string) (string, error) {
	if apiKey == "" {
		return "", fmt.Errorf("Tavily API key is not set")
	}

	// Tavily API endpoint
	url := "https://api.tavily.com/search"

	// Request body
	requestBody, err := json.Marshal(map[string]interface{}{
		"q":       query,
		"api_key": apiKey,
		"search_depth": "basic", // or "advanced"
		"include_answer": false,
		"include_raw_content": false,
		"max_results": 5,
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal Tavily request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(requestBody)))
	if err != nil {
		return "", fmt.Errorf("failed to create Tavily request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make Tavily API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Tavily API returned non-200 status: %d %s", resp.StatusCode, resp.Status)
	}

	var tavilyResp tavilyResponse
	if err := json.NewDecoder(resp.Body).Decode(&tavilyResp); err != nil {
		return "", fmt.Errorf("failed to decode Tavily API response: %w", err)
	}

	var llmContent strings.Builder
	llmContent.WriteString(fmt.Sprintf("Web search results for \"%s\" (Tavily):\n\n", query))

	if len(tavilyResp.Results) == 0 {
		return fmt.Sprintf("No search results found for query: \"%s\" (Tavily)", query), nil
	}

	for i, item := range tavilyResp.Results {
		llmContent.WriteString(fmt.Sprintf("[%d] %s (%s)\n", i+1, item.Title, item.URL))
		llmContent.WriteString(fmt.Sprintf("   %s\n", item.Content))
	}

	return llmContent.String(), nil
}

// Execute performs a web search operation.
func (t *WebSearchTool) Execute(ctx context.Context, args map[string]any) (types.ToolResult, error) {
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return types.ToolResult{}, fmt.Errorf("invalid or missing 'query' argument")
	}

	provider := t.settingsService.GetWebSearchProvider()
	var searchResult string
	var err error

	switch provider {
	case types.WebSearchProviderGoogleCustomSearch:
		googleCustomSearchConfig := t.settingsService.GetGoogleCustomSearchSettings()
		apiKey := googleCustomSearchConfig.ApiKey
		cxId := googleCustomSearchConfig.CxId

		if apiKey == "" || cxId == "" {
			return types.ToolResult{}, fmt.Errorf("googleCustomSearch.apiKey and googleCustomSearch.cxId must be set in settings.json for web search")
		}

		svc, errSvc := customsearch.NewService(ctx, option.WithAPIKey(apiKey))
		if errSvc != nil {
			return types.ToolResult{}, fmt.Errorf("failed to create customsearch service: %w", errSvc)
		}

		resp, errResp := svc.Cse.List().Q(query).Cx(cxId).Do()
		if errResp != nil {
			return types.ToolResult{}, fmt.Errorf("failed to perform custom search: %w", errResp)
		}

		var llmContent strings.Builder
		llmContent.WriteString(fmt.Sprintf("Web search results for \"%s\":\n\n", query))

		if len(resp.Items) == 0 {
			searchResult = fmt.Sprintf("No search results found for query: \"%s\"", query)
		} else {
			for i, item := range resp.Items {
				llmContent.WriteString(fmt.Sprintf("[%d] %s (%s)\n", i+1, item.Title, item.Link))
				llmContent.WriteString(fmt.Sprintf("   %s\n", item.Snippet))
			}
			searchResult = llmContent.String()
		}

	case types.WebSearchProviderTavily:
		tavilyConfig := t.settingsService.GetTavilySettings()
		apiKey := tavilyConfig.ApiKey

		searchResult, err = searchWithTavily(ctx, query, apiKey)
		if err != nil {
			return types.ToolResult{}, fmt.Errorf("Tavily search failed: %w", err)
		}

	default:
		return types.ToolResult{}, fmt.Errorf("unsupported web search provider: %s", provider)
	}

	return types.ToolResult{
		LLMContent:    searchResult,
		ReturnDisplay: searchResult,
	}, nil
}
