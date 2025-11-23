package tools

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go-ai-agent-v2/go-cli/pkg/services"
	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/stretchr/testify/assert"
	"google.golang.org/api/customsearch/v1"
	"google.golang.org/api/option"
)

func TestNewWebSearchTool(t *testing.T) {
	mockSettingsService := &services.MockSettingsService{}
	tool := NewWebSearchTool(mockSettingsService, nil, nil)

	assert.NotNil(t, tool)
	assert.Equal(t, "web_search", tool.Name())
	assert.Equal(t, "Performs a web search using Google Search (via the Gemini API) and returns the results.", tool.Description())
	assert.NotNil(t, tool.settingsService)
	assert.NotNil(t, tool.httpClient)
	assert.NotNil(t, tool.googleCustomSearchServiceFactory)
}

func TestWebSearchTool_searchWithGoogle(t *testing.T) {
	ctx := context.Background()
	query := "test query"
	apiKey := "test-api-key"
	cxId := "test-cx-id"

	t.Run("Successful search", func(t *testing.T) {
		mockGoogleServiceFactory := func(ctx context.Context, key string) (*customsearch.Service, error) {
			assert.Equal(t, apiKey, key)
			mockClient := &http.Client{
				Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					assert.Contains(t, req.URL.String(), "q=test+query")
					assert.Contains(t, req.URL.String(), "cx=test-cx-id")

					respBody := `{
						"items": [
							{
								"title": "Result 1",
								"link": "http://example.com/1",
								"snippet": "Snippet 1"
							}
						]
					}`
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(respBody)),
						Header:     make(http.Header),
					}, nil
				}),
			}
			svc, _ := customsearch.NewService(ctx, option.WithHTTPClient(mockClient), option.WithAPIKey(key))
			return svc, nil
		}
		tool := NewWebSearchTool(&services.MockSettingsService{}, nil, mockGoogleServiceFactory)
		result, err := tool.searchWithGoogle(ctx, query, apiKey, cxId)
		assert.NoError(t, err)
		assert.Contains(t, result, "Web search results for \"test query\" (Google Custom Search):")
		assert.Contains(t, result, "[1] Result 1 (http://example.com/1)")
		assert.Contains(t, result, "Snippet 1")
	})

	t.Run("No results found", func(t *testing.T) {
		mockGoogleServiceFactory := func(ctx context.Context, key string) (*customsearch.Service, error) {
			mockClient := &http.Client{
				Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					respBody := `{"items": []}`
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(respBody)),
						Header:     make(http.Header),
					}, nil
				}),
			}
			svc, _ := customsearch.NewService(ctx, option.WithHTTPClient(mockClient), option.WithAPIKey(key))
			return svc, nil
		}
		tool := NewWebSearchTool(&services.MockSettingsService{}, nil, mockGoogleServiceFactory)
		result, err := tool.searchWithGoogle(ctx, query, apiKey, cxId)
		assert.NoError(t, err)
		assert.Contains(t, result, "No search results found for query: \"test query\" (Google Custom Search)")
	})

	t.Run("Missing API key", func(t *testing.T) {
		tool := NewWebSearchTool(&services.MockSettingsService{}, nil, nil)
		result, err := tool.searchWithGoogle(ctx, query, "", cxId)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API key or CX ID is not set")
		assert.Empty(t, result)
	})

	t.Run("API error", func(t *testing.T) {
		mockGoogleServiceFactory := func(ctx context.Context, key string) (*customsearch.Service, error) {
			mockClient := &http.Client{
				Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       io.NopCloser(strings.NewReader("Internal Server Error")),
						Header:     make(http.Header),
					}, nil
				}),
			}
			svc, _ := customsearch.NewService(ctx, option.WithHTTPClient(mockClient), option.WithAPIKey(key))
			return svc, nil
		}
		tool := NewWebSearchTool(&services.MockSettingsService{}, nil, mockGoogleServiceFactory)
		result, err := tool.searchWithGoogle(ctx, query, apiKey, cxId)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to perform custom search")
		assert.Empty(t, result)
	})
}

func TestWebSearchTool_searchWithTavily(t *testing.T) {
	ctx := context.Background()
	query := "test query"
	apiKey := "test-tavily-key"

	t.Run("Successful search", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			body, _ := io.ReadAll(r.Body)
			assert.Contains(t, string(body), `"q":"test query"`)
			assert.Contains(t, string(body), `"api_key":"test-tavily-key"`)

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"results": [
					{
						"title": "Tavily Result 1",
						"url": "http://tavily.com/1",
						"content": "Tavily Snippet 1"
					}
				],
				"query": "test query"
			}`))
		}))
		defer server.Close()

		mockHttpClient := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			req.URL.Host = strings.TrimPrefix(server.URL, "http://")
			req.URL.Scheme = "http"
			return server.Client().Transport.RoundTrip(req)
		})}
		tool := NewWebSearchTool(&services.MockSettingsService{}, mockHttpClient, nil)
		result, err := tool.searchWithTavily(ctx, query, apiKey)
		assert.NoError(t, err)
		assert.Contains(t, result, "Web search results for \"test query\" (Tavily):")
		assert.Contains(t, result, "[1] Tavily Result 1 (http://tavily.com/1)")
		assert.Contains(t, result, "Tavily Snippet 1")
	})

	t.Run("No results found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"results": [],
				"query": "test query"
			}`))
		}))
		defer server.Close()

		mockHttpClient := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			req.URL.Host = strings.TrimPrefix(server.URL, "http://")
			req.URL.Scheme = "http"
			return server.Client().Transport.RoundTrip(req)
		})}
		tool := NewWebSearchTool(&services.MockSettingsService{}, mockHttpClient, nil)
		result, err := tool.searchWithTavily(ctx, query, apiKey)
		assert.NoError(t, err)
		assert.Contains(t, result, "No search results found for query: \"test query\" (Tavily)")
	})

	t.Run("Missing API key", func(t *testing.T) {
		tool := NewWebSearchTool(&services.MockSettingsService{}, nil, nil)
		result, err := tool.searchWithTavily(ctx, query, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Tavily API key is not set")
		assert.Empty(t, result)
	})

	t.Run("API error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Internal Server Error"))
		}))
		defer server.Close()

		mockHttpClient := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			req.URL.Host = strings.TrimPrefix(server.URL, "http://")
			req.URL.Scheme = "http"
			return server.Client().Transport.RoundTrip(req)
		})}
		tool := NewWebSearchTool(&services.MockSettingsService{}, mockHttpClient, nil)
		result, err := tool.searchWithTavily(ctx, query, apiKey)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Tavily API returned non-200 status")
		assert.Empty(t, result)
	})
}

func TestWebSearchTool_Execute(t *testing.T) {
	ctx := context.Background()
	query := "test query"

	t.Run("Google Custom Search Provider - Success", func(t *testing.T) {
		mockSettingsService := &services.MockSettingsService{}
		mockSettingsService.On("GetWebSearchProvider").Return(types.WebSearchProviderGoogleCustomSearch)
		mockSettingsService.On("GetGoogleCustomSearchSettings").Return(&types.GoogleCustomSearchSettings{
			ApiKey: "test-google-api-key",
			CxId:   "test-google-cx-id",
		})

		mockGoogleServiceFactory := func(ctx context.Context, key string) (*customsearch.Service, error) {
			mockClient := &http.Client{
				Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					respBody := `{
						"items": [
							{
								"title": "Google Result 1",
								"link": "http://google.com/1",
								"snippet": "Google Snippet 1"
							}
						]
					}`
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(respBody)),
						Header:     make(http.Header),
					}, nil
				}),
			}
			svc, _ := customsearch.NewService(ctx, option.WithHTTPClient(mockClient), option.WithAPIKey(key))
			return svc, nil
		}

		tool := NewWebSearchTool(mockSettingsService, nil, mockGoogleServiceFactory)
		result, err := tool.Execute(ctx, map[string]any{"query": query})

		assert.NoError(t, err)
		assert.Contains(t, result.LLMContent, "Google Result 1")
		assert.Contains(t, result.ReturnDisplay, "Google Result 1")
		mockSettingsService.AssertExpectations(t)
	})

	t.Run("Tavily Provider - Success", func(t *testing.T) {
		mockSettingsService := &services.MockSettingsService{}
		mockSettingsService.On("GetWebSearchProvider").Return(types.WebSearchProviderTavily)
		mockSettingsService.On("GetTavilySettings").Return(&types.TavilySettings{
			ApiKey: "test-tavily-api-key",
		})

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"results": [
					{
						"title": "Tavily Result 1",
						"url": "http://tavily.com/1",
						"content": "Tavily Snippet 1"
					}
				],
				"query": "test query"
			}`))
		}))
		defer server.Close()

		mockHttpClient := &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				req.URL.Host = strings.TrimPrefix(server.URL, "http://")
				req.URL.Scheme = "http"
				return server.Client().Transport.RoundTrip(req)
			}),
		}

		tool := NewWebSearchTool(mockSettingsService, mockHttpClient, nil)
		result, err := tool.Execute(ctx, map[string]any{"query": query})

		assert.NoError(t, err)
		assert.Contains(t, result.LLMContent, "Tavily Result 1")
		assert.Contains(t, result.ReturnDisplay, "Tavily Result 1")
		mockSettingsService.AssertExpectations(t)
	})

	t.Run("Unsupported Provider", func(t *testing.T) {
		mockSettingsService := &services.MockSettingsService{}
		mockSettingsService.On("GetWebSearchProvider").Return(types.WebSearchProvider("unsupported"))

		tool := NewWebSearchTool(mockSettingsService, nil, nil)
		_, err := tool.Execute(ctx, map[string]any{"query": query})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported web search provider: unsupported")
		mockSettingsService.AssertExpectations(t)
	})

	t.Run("Missing Query", func(t *testing.T) {
		mockSettingsService := &services.MockSettingsService{}

		tool := NewWebSearchTool(mockSettingsService, nil, nil)
		_, err := tool.Execute(ctx, map[string]any{"query": ""})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid or missing 'query' argument")
	})
}

// roundTripFunc is a helper to create a http.RoundTripper from a function.
type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
