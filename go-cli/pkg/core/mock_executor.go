package core

import (
	"fmt"
	"time"

	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/google/generative-ai-go/genai"
)

// MockExecutor is a mock implementation of the Executor interface for testing.
type MockExecutor struct {
	GenerateContentFunc         func(contents ...*genai.Content) (*genai.GenerateContentResponse, error)
	ExecuteToolFunc             func(fc *genai.FunctionCall) (types.ToolResult, error)
	SendMessageStreamFunc       func(modelName string, messageParams types.MessageParams, promptId string) (<-chan types.StreamResponse, error)
	ListModelsFunc              func() ([]string, error)
	DefaultGenerateContentResponse *genai.GenerateContentResponse // New field for configurable default response
	DefaultExecuteToolResult    *types.ToolResult              // New field for configurable default tool execution result
	GetHistoryFunc              func() ([]*genai.Content, error) // New field for configurable mock history
	SetHistoryFunc              func(history []*genai.Content) error
	mockHistory                 []*genai.Content // Field to store mock history
}

// NewMockExecutor creates a new MockExecutor instance.
func NewMockExecutor(defaultResponse *genai.GenerateContentResponse, defaultToolResult *types.ToolResult) *MockExecutor {
	me := &MockExecutor{
		DefaultGenerateContentResponse: defaultResponse,
		DefaultExecuteToolResult:    defaultToolResult,
	}
	me.GenerateContentFunc = func(contents ...*genai.Content) (*genai.GenerateContentResponse, error) {
			if me.DefaultGenerateContentResponse != nil {
				return me.DefaultGenerateContentResponse, nil
			}
			// Default mock implementation: return a dummy response
			return &genai.GenerateContentResponse{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []genai.Part{genai.Text("Mocked response from GenerateContent.")},
						},
					},
				},
			}, nil
		}
	me.ExecuteToolFunc = func(fc *genai.FunctionCall) (types.ToolResult, error) {
			if me.DefaultExecuteToolResult != nil {
				return *me.DefaultExecuteToolResult, nil
			}
			// Default mock implementation: return a generic success
			return types.ToolResult{
				LLMContent:    fmt.Sprintf("Mocked result for tool %s with args %v", fc.Name, fc.Args),
				ReturnDisplay: fmt.Sprintf("Mocked result for tool %s with args %v", fc.Name, fc.Args),
			}, nil
		}
			me.SendMessageStreamFunc = func(modelName string, messageParams types.MessageParams, promptId string) (<-chan types.StreamResponse, error) {
					respChan := make(chan types.StreamResponse)
					go func() {
						defer close(respChan)
						// Simulate a streamed response
						respChan <- types.StreamResponse{
							Type: types.StreamEventTypeChunk,
							Value: &genai.GenerateContentResponse{
								Candidates: []*genai.Candidate{
									{
										Content: &genai.Content{
											Parts: []genai.Part{genai.Text("Mocked streamed response chunk 1.")},
										},
									},
								},
							},
						}
						time.Sleep(50 * time.Millisecond)
						respChan <- types.StreamResponse{
							Type: types.StreamEventTypeChunk,
							Value: &genai.GenerateContentResponse{
								Candidates: []*genai.Candidate{
									{
										Content: &genai.Content{
											Parts: []genai.Part{genai.Text("Mocked streamed response chunk 2.")},
										},
									},
								},
							},
						}
					}()
					return respChan, nil
				}
	me.ListModelsFunc = func() ([]string, error) {
			return []string{"mock-model-1", "mock-model-2"}, nil
		}
	me.GetHistoryFunc = func() ([]*genai.Content, error) {
			return me.mockHistory, nil
		}
	me.SetHistoryFunc = func(history []*genai.Content) error {
			me.mockHistory = history
			return nil
		}
	return me
}

// GenerateContent implements the Executor interface.
func (me *MockExecutor) GenerateContent(contents ...*genai.Content) (*genai.GenerateContentResponse, error) {
	return me.GenerateContentFunc(contents...)
}

// ExecuteTool implements the Executor interface.
func (me *MockExecutor) ExecuteTool(fc *genai.FunctionCall) (types.ToolResult, error) {
	return me.ExecuteToolFunc(fc)
}

// SendMessageStream implements the Executor interface.
func (me *MockExecutor) SendMessageStream(modelName string, messageParams types.MessageParams, promptId string) (<-chan types.StreamResponse, error) {
	return me.SendMessageStreamFunc(modelName, messageParams, promptId)
}

// ListModels implements the Executor interface.
func (me *MockExecutor) ListModels() ([]string, error) {
	return me.ListModelsFunc()
}

// GetHistory implements the Executor interface.
func (me *MockExecutor) GetHistory() ([]*genai.Content, error) {
	return me.GetHistoryFunc()
}

// SetHistory implements the Executor interface.
func (me *MockExecutor) SetHistory(history []*genai.Content) error {
	return me.SetHistoryFunc(history)
}
