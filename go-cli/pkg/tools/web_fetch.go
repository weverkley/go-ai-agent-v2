package tools

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// WebFetchTool represents the web-fetch tool.
type WebFetchTool struct {
}

// NewWebFetchTool creates a new instance of WebFetchTool.
func NewWebFetchTool() *WebFetchTool {
	return &WebFetchTool{}
}

// Execute performs a web fetch operation.
func (t *WebFetchTool) Execute(
	urls []string,
	summarize bool,
	extractKeyPoints bool,
) (string, error) {
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

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			results.WriteString(fmt.Sprintf("Error reading body from %s: %v\n", url, err))
			continue
		}

		content := string(body)
		results.WriteString(fmt.Sprintf("\n--- Content from %s ---\n", url))
		results.WriteString(content)
		results.WriteString("\n--- End of Content ---\n")

		if summarize {
			results.WriteString(fmt.Sprintf("\n(Summarization not yet implemented for %s)\n", url))
		}
		if extractKeyPoints {
			results.WriteString(fmt.Sprintf("\n(Key point extraction not yet implemented for %s)\n", url))
		}
	}

	return results.String(), nil
}
