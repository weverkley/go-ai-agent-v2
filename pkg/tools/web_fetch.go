package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/types"
)

// WebFetchTool represents the web-fetch tool.
type WebFetchTool struct {
	*types.BaseDeclarativeTool
}

// NewWebFetchTool creates a new instance of WebFetchTool.
func NewWebFetchTool() *WebFetchTool {
	return &WebFetchTool{
		types.NewBaseDeclarativeTool(
			"web_fetch",
			"web_fetch",
			"Processes content from URL(s), including local and private network addresses (e.g., localhost), embedded in a prompt. Include up to 20 URLs and instructions (e.g., summarize, extract specific data) directly in the 'prompt' parameter.",
			types.KindOther, // Assuming KindOther for now
			types.JsonSchemaObject{
				Type: "object",
				Properties: map[string]types.JsonSchemaProperty{
					"prompt": {
						Type:        "string",
						Description: "A comprehensive prompt that includes the URL(s) (up to 20) to fetch and specific instructions on how to process their content (e.g., \"Summarize https://example.com/article and extract key points from https://another.com/data\"). All URLs to be fetched must be valid and complete, starting with \"http://\" or \"https://\", and be fully-formed with a valid hostname (e.g., a domain name like \"example.com\" or an IP address). For example, \"https://example.com\" is valid, but \"example.com\" is not.",
					},
				},
				Required: []string{"prompt"},
			},
			false, // isOutputMarkdown
			false, // canUpdateOutput
			nil,   // MessageBus
		),
	}
}

// extractUrls finds all URLs in a given text.
func extractUrls(text string) []string {
	// A simple regex to find URLs. This can be improved.
	re := regexp.MustCompile(`https?://[^\n]+`)
	return re.FindAllString(text, -1)
}

// Execute performs a web fetch operation.
func (t *WebFetchTool) Execute(ctx context.Context, args map[string]any) (types.ToolResult, error) {
	prompt, ok := args["prompt"].(string)
	if !ok || prompt == "" {
		return types.ToolResult{}, fmt.Errorf("invalid or missing 'prompt' argument")
	}

	urls := extractUrls(prompt)
	if len(urls) == 0 {
		return types.ToolResult{
			LLMContent:    "No URLs found in the prompt.",
			ReturnDisplay: "No URLs found in the prompt.",
		}, nil
	}

	var results strings.Builder
	results.WriteString("Web Fetch Results:\n")

	for _, url := range urls {
		resp, err := http.Get(url)
		if err != nil {
			results.WriteString(fmt.Sprintf("Error fetching %s: %v\n", url, err))
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			results.WriteString(fmt.Sprintf("Error fetching %s: Status %d\n", url, resp.StatusCode))
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			results.WriteString(fmt.Sprintf("Error reading body from %s: %v\n", url, err))
			continue
		}

		content := string(body)
		results.WriteString(fmt.Sprintf("\n--- Content from %s ---\n", url))
		results.WriteString(content)
		results.WriteString("\n--- End of Content ---\n")
	}

	resultMessage := results.String()
	return types.ToolResult{
		LLMContent:    resultMessage,
		ReturnDisplay: resultMessage,
	}, nil
}

