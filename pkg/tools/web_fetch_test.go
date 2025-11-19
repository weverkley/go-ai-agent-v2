package tools

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/stretchr/testify/assert"
)

func TestWebFetchTool_Execute(t *testing.T) {
	tool := NewWebFetchTool()

	tests := []struct {
		name          string
		args          map[string]any
		setupServer   func() *httptest.Server
		expectedLLMContent string
		expectedReturnDisplay string
		expectedError string
	}{
		{
			name:          "missing prompt argument",
			args:          map[string]any{},
			setupServer:   func() *httptest.Server { return nil },
			expectedError: "invalid or missing 'prompt' argument",
		},
		{
			name:          "empty prompt argument",
			args:          map[string]any{"prompt": ""},
			setupServer:   func() *httptest.Server { return nil },
			expectedError: "invalid or missing 'prompt' argument",
		},
		{
			name:          "no URLs found in prompt",
			args:          map[string]any{"prompt": "This is a test prompt with no URLs."},
			setupServer:   func() *httptest.Server { return nil },
			expectedLLMContent:    "No URLs found in the prompt.",
			expectedReturnDisplay: "No URLs found in the prompt.",
		},
		{
			name: "successful fetch of a single URL",
			args: map[string]any{"prompt": "Fetch this: http://example.com/test"},
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, "/test", r.URL.Path)
					w.WriteHeader(http.StatusOK)
					fmt.Fprint(w, "Hello from example.com")
				}))
			},
			expectedLLMContent:    "Web Fetch Results:\n\n--- Content from http://example.com/test ---\nHello from example.com\n--- End of Content ---\n",
			expectedReturnDisplay: "Web Fetch Results:\n\n--- Content from http://example.com/test ---\nHello from example.com\n--- End of Content ---\n",
		},
		{
			name: "successful fetch of multiple URLs",
			args: map[string]any{"prompt": "Fetch these: http://example.com/1 and http://example.com/2"},
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path == "/1" {
						w.WriteHeader(http.StatusOK)
						fmt.Fprint(w, "Content from 1")
					} else if r.URL.Path == "/2" {
						w.WriteHeader(http.StatusOK)
						fmt.Fprint(w, "Content from 2")
					} else {
						w.WriteHeader(http.StatusNotFound)
					}
				}))
			},
			expectedLLMContent:    "Web Fetch Results:\n\n--- Content from http://example.com/1 ---\nContent from 1\n--- End of Content ---\n\n--- Content from http://example.com/2 ---\nContent from 2\n--- End of Content ---\n",
			expectedReturnDisplay: "Web Fetch Results:\n\n--- Content from http://example.com/1 ---\nContent from 1\n--- End of Content ---\n\n--- Content from http://example.com/2 ---\nContent from 2\n--- End of Content ---\n",
		},
		{
			name: "fetch error - non-200 status",
			args: map[string]any{"prompt": "Fetch this: http://example.com/error"},
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}))
			},
			expectedLLMContent:    "Web Fetch Results:\nError fetching http://example.com/error: Status 500\n",
			expectedReturnDisplay: "Web Fetch Results:\nError fetching http://example.com/error: Status 500\n",
			expectedError:         "multiple errors occurred during web fetch",
		},
		{
			name: "fetch error - network issue",
			args: map[string]any{"prompt": "Fetch this: http://nonexistent.domain"},
			setupServer: func() *httptest.Server {
				return nil // No server to simulate network error
			},
			expectedLLMContent:    "Web Fetch Results:\nError fetching http://nonexistent.domain: Get \"http://nonexistent.domain\": dial tcp: lookup nonexistent.domain",
			expectedReturnDisplay: "Web Fetch Results:\nError fetching http://nonexistent.domain: Get \"http://nonexistent.domain\": dial tcp: lookup nonexistent.domain",
			expectedError:         "multiple errors occurred during web fetch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server *httptest.Server
			if tt.setupServer != nil {
				server = tt.setupServer()
				if server != nil {
					defer server.Close()
					// Replace example.com with the test server URL for the test case
					if strings.Contains(tt.args["prompt"].(string), "http://example.com") {
						tt.args["prompt"] = strings.ReplaceAll(tt.args["prompt"].(string), "http://example.com", server.URL)
						tt.expectedLLMContent = strings.ReplaceAll(tt.expectedLLMContent, "http://example.com", server.URL)
						tt.expectedReturnDisplay = strings.ReplaceAll(tt.expectedReturnDisplay, "http://example.com", server.URL)
					}
				}
			}

			result, err := tool.Execute(context.Background(), tt.args)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				if result.Error != nil {
					assert.Equal(t, types.ToolErrorTypeExecutionFailed, result.Error.Type)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedLLMContent, result.LLMContent)
				assert.Equal(t, tt.expectedReturnDisplay, result.ReturnDisplay)
			}
		})
	}
}
