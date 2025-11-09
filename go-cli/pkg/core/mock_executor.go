package core

import (
	"fmt"
	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/google/generative-ai-go/genai"
)

// MockExecutor is a mock implementation of the Executor interface for testing.
type MockExecutor struct {
	GenerateContentFunc   func(contents ...*genai.Content) (*genai.GenerateContentResponse, error)
	ExecuteToolFunc       func(fc *genai.FunctionCall) (types.ToolResult, error)
	SendMessageStreamFunc func(modelName string, messageParams types.MessageParams, promptId string) (<-chan types.StreamResponse, error)
	ListModelsFunc        func() ([]string, error)
	GetHistoryFunc        func() ([]*genai.Content, error)
	SetHistoryFunc        func(history []*genai.Content) error
	CompressChatFunc      func(promptId string, force bool) (*types.ChatCompressionResult, error)
}

// GenerateContent mocks the GenerateContent method.
func (m *MockExecutor) GenerateContent(contents ...*genai.Content) (*genai.GenerateContentResponse, error) {
	if m.GenerateContentFunc != nil {
		return m.GenerateContentFunc(contents...)
	}
	return nil, fmt.Errorf("GenerateContent not implemented in mock")
}

// ExecuteTool mocks the ExecuteTool method.
func (m *MockExecutor) ExecuteTool(fc *genai.FunctionCall) (types.ToolResult, error) {
	if m.ExecuteToolFunc != nil {
		return m.ExecuteToolFunc(fc)
	}
	return types.ToolResult{}, fmt.Errorf("ExecuteTool not implemented in mock")
}

// SendMessageStream mocks the SendMessageStream method.
func (m *MockExecutor) SendMessageStream(modelName string, messageParams types.MessageParams, promptId string) (<-chan types.StreamResponse, error) {
	if m.SendMessageStreamFunc != nil {
		return m.SendMessageStreamFunc(modelName, messageParams, promptId)
	}
	respChan := make(chan types.StreamResponse)
	close(respChan)
	return respChan, fmt.Errorf("SendMessageStream not implemented in mock")
}

// ListModels mocks the ListModels method.
func (m *MockExecutor) ListModels() ([]string, error) {
	if m.ListModelsFunc != nil {
		return m.ListModelsFunc()
	}
	return nil, fmt.Errorf("ListModels not implemented in mock")
}

// GetHistory mocks the GetHistory method.
func (m *MockExecutor) GetHistory() ([]*genai.Content, error) {
	if m.GetHistoryFunc != nil {
		return m.GetHistoryFunc()
	}
	return nil, fmt.Errorf("GetHistory not implemented in mock")
}

// SetHistory mocks the SetHistory method.
func (m *MockExecutor) SetHistory(history []*genai.Content) error {
	if m.SetHistoryFunc != nil {
		return m.SetHistoryFunc(history)
	}
	return fmt.Errorf("SetHistory not implemented in mock")
}

// CompressChat mocks the CompressChat method.
func (m *MockExecutor) CompressChat(promptId string, force bool) (*types.ChatCompressionResult, error) {
	if m.CompressChatFunc != nil {
		return m.CompressChatFunc(promptId, force)
	}
	return nil, fmt.Errorf("CompressChat not implemented in mock")
}