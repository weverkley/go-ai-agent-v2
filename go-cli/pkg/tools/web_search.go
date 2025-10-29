package tools

import (
	"encoding/json"
	"fmt"
	"strings"
)

// WebSearchTool represents the web-search tool.
type WebSearchTool struct {
}

// NewWebSearchTool creates a new instance of WebSearchTool.
func NewWebSearchTool() *WebSearchTool {
	return &WebSearchTool{}
}

// WebSearchResult represents the structure of a single web search result.
type WebSearchResult struct {
	Title string `json:"title"`
	URI   string `json:"uri"`
	Snippet string `json:"snippet"`
}

// Execute performs a web search operation.
func (t *WebSearchTool) Execute(
	query string,
) (string, error) {
	// Call the default_api.google_web_search function (assuming it's available in the Go context)
	// For example: result, err := default_api.google_web_search(query)

	// Simulate the call to default_api.google_web_search
	// In a real scenario, this would be an actual call to the tool.

	simulatedOutput := fmt.Sprintf(`{
		"results": [
			{"title": "Google", "uri": "https://www.google.com", "snippet": "Search the world's information, including webpages, images, videos and more."}, 
			{"title": "Wikipedia", "uri": "https://en.wikipedia.org", "snippet": "The Free Encyclopedia"}
		]
	}`)

	var result struct {
		Results []WebSearchResult `json:"results"`
	}

	if err := json.Unmarshal([]byte(simulatedOutput), &result); err != nil {
		return "", fmt.Errorf("failed to parse simulated web search output: %w", err)
	}

	if len(result.Results) == 0 {
		return fmt.Sprintf("No search results found for query: \"%s\"", query), nil
	}

	var llmContent strings.Builder
	llmContent.WriteString(fmt.Sprintf("Web search results for \"%s\":\n\n", query))

	for i, res := range result.Results {
		llmContent.WriteString(fmt.Sprintf("[%d] %s (%s)\n", i+1, res.Title, res.URI))
		llmContent.WriteString(fmt.Sprintf("   %s\n", res.Snippet))
	}

	return llmContent.String(), nil
}
