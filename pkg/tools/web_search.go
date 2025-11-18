package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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
	settingsService types.SettingsServiceIface
	httpClient      *http.Client // For Tavily API
	googleCustomSearchServiceFactory func(ctx context.Context, apiKey string) (*customsearch.Service, error) // For Google Custom Search
}

// NewWebSearchTool creates a new instance of WebSearchTool.
func NewWebSearchTool(
	settingsService types.SettingsServiceIface,
	httpClient *http.Client,
	googleCustomSearchServiceFactory func(ctx context.Context, apiKey string) (*customsearch.Service, error),
) *WebSearchTool {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	if googleCustomSearchServiceFactory == nil {
		googleCustomSearchServiceFactory = func(ctx context.Context, apiKey string) (*customsearch.Service, error) {
			return customsearch.NewService(ctx, option.WithAPIKey(apiKey))
		}
	}

	return &WebSearchTool{
	BaseDeclarativeTool: types.NewBaseDeclarativeTool(
		types.WEB_SEARCH_TOOL_NAME,
		"Web Search",
		"Performs a web search using Google Search (via the Gemini API) and returns the results.",
		types.KindOther,
		(&types.JsonSchemaObject{
			Type: "object",
		}).SetProperties(map[string]*types.JsonSchemaProperty{
			"query": &types.JsonSchemaProperty{
				Type:        "string",
				Description: "The search query to find information on the web.",
			},
		}).SetRequired([]string{"query"}),
		false, // isOutputMarkdown
		false, // canUpdateOutput
		nil,   // MessageBus
	),
		settingsService: settingsService,
		httpClient:      httpClient,
		googleCustomSearchServiceFactory: googleCustomSearchServiceFactory,
	}
}

func (t *WebSearchTool) searchWithTavily(ctx context.Context, query string, apiKey string) (string, error) {
	if apiKey == "" {
		return "", &types.ToolError{
			Message: "Tavily API key is not set",
			Type:    types.ToolErrorTypeAPIKeyNotSet,
		}
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
		return "", &types.ToolError{
			Message: fmt.Sprintf("Failed to marshal Tavily request body: %v", err),
			Type:    types.ToolErrorTypeExecutionFailed,
		}
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(requestBody)))
	if err != nil {
		return "", &types.ToolError{
			Message: fmt.Sprintf("Failed to create Tavily request: %v", err),
			Type:    types.ToolErrorTypeExecutionFailed,
		}
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return "", &types.ToolError{
			Message: fmt.Sprintf("Failed to make Tavily API request: %v", err),
			Type:    types.ToolErrorTypeExecutionFailed,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", &types.ToolError{
			Message: fmt.Sprintf("Tavily API returned non-200 status: %d %s", resp.StatusCode, resp.Status),
			Type:    types.ToolErrorTypeAPIError,
		}
	}

	var tavilyResp tavilyResponse
	if err := json.NewDecoder(resp.Body).Decode(&tavilyResp); err != nil {
		return "", &types.ToolError{
			Message: fmt.Sprintf("Failed to decode Tavily API response: %v", err),
			Type:    types.ToolErrorTypeAPIError,
		}
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

func (t *WebSearchTool) searchWithGoogle(ctx context.Context, query string, apiKey string, cxId string) (string, error) {
	if apiKey == "" || cxId == "" {
		return "", &types.ToolError{
			Message: "Google Custom Search API key or CX ID is not set",
			Type:    types.ToolErrorTypeAPIKeyNotSet,
		}
	}

	svc, err := t.googleCustomSearchServiceFactory(ctx, apiKey)
	if err != nil {
		return "", &types.ToolError{
			Message: fmt.Sprintf("Failed to create customsearch service: %v", err),
			Type:    types.ToolErrorTypeExecutionFailed,
		}
	}

	resp, err := svc.Cse.List().Q(query).Cx(cxId).Do()
	if err != nil {
		return "", &types.ToolError{
			Message: fmt.Sprintf("Failed to perform custom search: %v", err),
			Type:    types.ToolErrorTypeAPIError,
		}
	}

	var llmContent strings.Builder
	llmContent.WriteString(fmt.Sprintf("Web search results for \"%s\" (Google Custom Search):\n\n", query))

	if len(resp.Items) == 0 {
		return fmt.Sprintf("No search results found for query: \"%s\" (Google Custom Search)", query), nil
	}

	for i, item := range resp.Items {
		llmContent.WriteString(fmt.Sprintf("[%d] %s (%s)\n", i+1, item.Title, item.Link))
		llmContent.WriteString(fmt.Sprintf("   %s\n", item.Snippet))
	}

	return llmContent.String(), nil
}

// Execute performs a web search operation.
func (t *WebSearchTool) Execute(ctx context.Context, args map[string]any) (types.ToolResult, error) {
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: "Invalid or missing 'query' argument",
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("invalid or missing 'query' argument")
	}

	provider := t.settingsService.GetWebSearchProvider()
	var searchResult string
	var err error

	switch provider {
	case types.WebSearchProviderGoogleCustomSearch:
		googleCustomSearchConfig := t.settingsService.GetGoogleCustomSearchSettings()
		apiKey := googleCustomSearchConfig.ApiKey
		cxId := googleCustomSearchConfig.CxId

		searchResult, err = t.searchWithGoogle(ctx, query, apiKey, cxId)
		if err != nil {
			return types.ToolResult{
				Error: &types.ToolError{
					Message: fmt.Sprintf("Google Custom Search failed: %v", err),
					Type:    types.ToolErrorTypeExecutionFailed,
				},
			}, fmt.Errorf("Google Custom Search failed: %w", err)
		}

	case types.WebSearchProviderTavily:
		tavilyConfig := t.settingsService.GetTavilySettings()
		apiKey := tavilyConfig.ApiKey

		searchResult, err = t.searchWithTavily(ctx, query, apiKey)
		if err != nil {
			return types.ToolResult{
				Error: &types.ToolError{
					Message: fmt.Sprintf("Tavily search failed: %v", err),
					Type:    types.ToolErrorTypeExecutionFailed,
				},
			}, fmt.Errorf("Tavily search failed: %w", err)
		}

	default:
		return types.ToolResult{
			Error: &types.ToolError{
				Message: fmt.Sprintf("Unsupported web search provider: %s", provider),
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("unsupported web search provider: %s", provider)
	}

	return types.ToolResult{
		LLMContent:    searchResult,
		ReturnDisplay: searchResult,
	}, nil
}
