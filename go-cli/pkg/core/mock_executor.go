package core

import (
	"fmt"
	"time"

	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/google/generative-ai-go/genai"
)

// MockExecutor is a mock implementation of the Executor interface for testing.
type MockExecutor struct {
	GenerateContentFunc func(contents ...*genai.Content) (*genai.GenerateContentResponse, error)
	ExecuteToolFunc     func(fc *genai.FunctionCall) (types.ToolResult, error)
	SendMessageStreamFunc func(modelName string, messageParams types.MessageParams, promptId string) (<-chan types.StreamResponse, error)
	ListModelsFunc      func() ([]string, error)
}

// NewMockExecutor creates a new MockExecutor instance.
func NewMockExecutor() *MockExecutor {
	return &MockExecutor{
		GenerateContentFunc: func(contents ...*genai.Content) (*genai.GenerateContentResponse, error) {
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
		},
		ExecuteToolFunc: func(fc *genai.FunctionCall) (types.ToolResult, error) {
			// Default mock implementation: return a dummy tool result
			return types.ToolResult{
				LLMContent:    fmt.Sprintf("Mocked result for tool %s with args %v", fc.Name, fc.Args),
				ReturnDisplay: fmt.Sprintf("Mocked result for tool %s with args %v", fc.Name, fc.Args),
			}, nil
		},
		SendMessageStreamFunc: func(modelName string, messageParams types.MessageParams, promptId string) (<-chan types.StreamResponse, error) {
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
		},
		ListModelsFunc: func() ([]string, error) {
			return []string{"mock-model-1", "mock-model-2"}, nil
		},
	}
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
