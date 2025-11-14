package core

import (
	"context"
	"fmt"
	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/google/generative-ai-go/genai"
)

// MockExecutor is a mock implementation of the Executor interface for testing.
type MockExecutor struct {
	GenerateContentFunc        func(contents ...*genai.Content) (*genai.GenerateContentResponse, error)
	GenerateContentWithToolsFunc func(ctx context.Context, history []*genai.Content, tools []*genai.Tool) (*genai.GenerateContentResponse, error)
	ExecuteToolFunc            func(fc *genai.FunctionCall) (types.ToolResult, error)
	SendMessageStreamFunc      func(modelName string, messageParams types.MessageParams, promptId string) (<-chan types.StreamResponse, error)
	ListModelsFunc             func() ([]string, error)
	GetHistoryFunc             func() ([]*genai.Content, error)
	SetHistoryFunc             func(history []*genai.Content) error
	CompressChatFunc           func(promptId string, force bool) (*types.ChatCompressionResult, error)
	GenerateStreamFunc         func(contents ...*genai.Content) (<-chan any, error)
}

// NewExpressAPIMockExecutor creates a mock executor that simulates creating an Express API.
func NewExpressAPIMockExecutor() *MockExecutor {
	mock := &MockExecutor{}

	// This mock doesn't need to execute tools, as the GenerateStream mock will
	// simulate the full lifecycle of tool calls.
	mock.ExecuteToolFunc = func(fc *genai.FunctionCall) (types.ToolResult, error) {
		return types.ToolResult{
			LLMContent:    fmt.Sprintf("Successfully executed %s", fc.Name),
			ReturnDisplay: fmt.Sprintf("Tool %s executed.", fc.Name),
		}, nil
	}

	// Mock GenerateStream to follow a script of events.
	mock.GenerateStreamFunc = func(contents ...*genai.Content) (<-chan any, error) {
		eventChan := make(chan any)

		go func() {
			defer close(eventChan)
			eventChan <- types.StreamingStartedEvent{}
			
			// Step 1: mkdir
			eventChan <- types.ThinkingEvent{}
			eventChan <- types.ToolCallStartEvent{ToolCallID: "mock-1", ToolName: "execute_command", Args: map[string]interface{}{"command": "mkdir my-express-app"}}
			eventChan <- types.ToolCallEndEvent{ToolCallID: "mock-1", ToolName: "execute_command", Result: "Created directory my-express-app"}

			// Step 2: npm init
			eventChan <- types.ThinkingEvent{}
			eventChan <- types.ToolCallStartEvent{ToolCallID: "mock-2", ToolName: "execute_command", Args: map[string]interface{}{"command": "npm init -y", "path": "my-express-app"}}
			eventChan <- types.ToolCallEndEvent{ToolCallID: "mock-2", ToolName: "execute_command", Result: "Initialized package.json"}

			// Step 3: npm install
			eventChan <- types.ThinkingEvent{}
			eventChan <- types.ToolCallStartEvent{ToolCallID: "mock-3", ToolName: "execute_command", Args: map[string]interface{}{"command": "npm install express", "path": "my-express-app"}}
			eventChan <- types.ToolCallEndEvent{ToolCallID: "mock-3", ToolName: "execute_command", Result: "Installed express"}

			// Step 4: write_file
			eventChan <- types.ThinkingEvent{}
			serverCode := "const express = require('express');\nconst app = express();\nconst port = 3000;\n\napp.get('/', (req, res) => {\n  res.send('Hello World!');\n});\n\napp.listen(port, () => {\n  console.log(`Example app listening on port ${port}`);\n});"
			eventChan <- types.ToolCallStartEvent{ToolCallID: "mock-4", ToolName: "write_file", Args: map[string]interface{}{"file_path": "my-express-app/index.js", "content": serverCode}}
			eventChan <- types.ToolCallEndEvent{ToolCallID: "mock-4", ToolName: "write_file", Result: "Wrote index.js"}

			// Step 5: Final Response
			eventChan <- types.ThinkingEvent{}
			eventChan <- types.FinalResponseEvent{Content: "I have created the Express API in `my-express-app/index.js`."}
		}()

		return eventChan, nil
	}

	return mock
}


// GenerateContent mocks the GenerateContent method.
func (m *MockExecutor) GenerateContent(contents ...*genai.Content) (*genai.GenerateContentResponse, error) {
	if m.GenerateContentFunc != nil {
		return m.GenerateContentFunc(contents...)
	}
	return nil, fmt.Errorf("GenerateContent not implemented in mock")
}

// GenerateContentWithTools mocks the GenerateContentWithTools method.
func (m *MockExecutor) GenerateContentWithTools(ctx context.Context, history []*genai.Content, tools []*genai.Tool) (*genai.GenerateContentResponse, error) {
	if m.GenerateContentWithToolsFunc != nil {
		return m.GenerateContentWithToolsFunc(ctx, history, tools)
	}
	return nil, fmt.Errorf("GenerateContentWithTools not implemented in mock")
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
	
// GenerateStream mocks the GenerateStream method.
func (m *MockExecutor) GenerateStream(contents ...*genai.Content) (<-chan any, error) {
	if m.GenerateStreamFunc != nil {
		return m.GenerateStreamFunc(contents...)
	}
	respChan := make(chan any)
	close(respChan)
	return respChan, fmt.Errorf("GenerateStream not implemented in mock")
}

