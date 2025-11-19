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
			(&types.JsonSchemaObject{
				Type: "object",
			}).SetProperties(map[string]*types.JsonSchemaProperty{
				"prompt": &types.JsonSchemaProperty{
					Type:        "string",
					Description: "A comprehensive prompt that includes the URL(s) (up to 20) to fetch and specific instructions on how to process their content (e.g., \"Summarize https://example.com/article and extract key points from https://another.com/data\"). All URLs to be fetched must be valid and complete, starting with \"http://\" or \"https://\", and be fully-formed with a valid hostname (e.g., a domain name like \"example.com\" or an IP address). For example, \"https://example.com\" is valid, but \"example.com\" is not.",
				},
			}).SetRequired([]string{"prompt"}),
			false, // isOutputMarkdown
			false, // canUpdateOutput
			nil,   // MessageBus
		),
	}
}

// extractUrls finds all URLs in a given text.
func extractUrls(text string) []string {
	// A simple regex to find URLs, stopping at whitespace or newline.
	re := regexp.MustCompile(`https?://[^\s\n]+`)
	return re.FindAllString(text, -1)
}

// Execute performs a web fetch operation.
func (t *WebFetchTool) Execute(ctx context.Context, args map[string]any) (types.ToolResult, error) {
	prompt, ok := args["prompt"].(string)
	if !ok || prompt == "" {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: "invalid or missing 'prompt' argument",
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("invalid or missing 'prompt' argument")
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

	var fetchErrorStrings []string
	var returnErr error

	for _, url := range urls {
		// Clean trailing characters that might be part of surrounding text
		url = strings.TrimRight(url, `.,;:)!?'"`)

		resp, err := http.Get(url)
		if err != nil {
			errorMsg := fmt.Sprintf("Error fetching %s: %v", url, err)
			fetchErrorStrings = append(fetchErrorStrings, errorMsg)
			results.WriteString(errorMsg + "\n")
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			errorMsg := fmt.Sprintf("Error fetching %s: Status %d", url, resp.StatusCode)
			fetchErrorStrings = append(fetchErrorStrings, errorMsg)
			results.WriteString(errorMsg + "\n")
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			errorMsg := fmt.Sprintf("Error reading body from %s: %v", url, err)
			fetchErrorStrings = append(fetchErrorStrings, errorMsg)
			results.WriteString(errorMsg + "\n")
			continue
		}

		content := string(body)
		results.WriteString(fmt.Sprintf("\n--- Content from %s ---\n", url))
		results.WriteString(content)
		results.WriteString("\n--- End of Content ---\n")
	}

	var toolError *types.ToolError
	if len(fetchErrorStrings) > 0 {
		errorSummary := strings.Join(fetchErrorStrings, "; ")
		toolError = &types.ToolError{
			Message: fmt.Sprintf("Multiple errors occurred during web fetch: %s", errorSummary),
			Type:    types.ToolErrorTypeExecutionFailed,
		}
		returnErr = fmt.Errorf("multiple errors occurred during web fetch")
	}

	resultMessage := results.String()
	return types.ToolResult{
		LLMContent:    resultMessage,
		ReturnDisplay: resultMessage,
		Error:         toolError,
	}, returnErr
}

