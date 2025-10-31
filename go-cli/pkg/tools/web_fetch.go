package tools

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/google/generative-ai-go/genai"
)

// WebFetchTool represents the web-fetch tool.
type WebFetchTool struct{}

// NewWebFetchTool creates a new instance of WebFetchTool.
func NewWebFetchTool() *WebFetchTool {
	return &WebFetchTool{}
}

// Name returns the name of the tool.
func (t *WebFetchTool) Name() string {
	return "web_fetch"
}

// Definition returns the tool's definition for the Gemini API.
func (t *WebFetchTool) Definition() *genai.Tool {
	return &genai.Tool{
		FunctionDeclarations: []*genai.FunctionDeclaration{
			{
				Name:        t.Name(),
				Description: "Processes content from URL(s) embedded in a prompt.",
				Parameters: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"prompt": {
							Type:        genai.TypeString,
							Description: "A prompt that includes the URL(s) to fetch and instructions on how to process their content.",
						},
					},
					Required: []string{"prompt"},
				},
			},
		},
	}
}

// extractUrls finds all URLs in a given text.
func extractUrls(text string) []string {
	// A simple regex to find URLs. This can be improved.
	re := regexp.MustCompile(`https?://[^\n]+`)
	return re.FindAllString(text, -1)
}

// Execute performs a web fetch operation.
func (t *WebFetchTool) Execute(args map[string]any) (types.ToolResult, error) {
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

