package core

import (
	"context"
	"fmt"
	"time"
	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/google/generative-ai-go/genai"
)

// MockExecutor is a mock implementation of the Executor interface for testing.
type MockExecutor struct {
	GenerateContentFunc        func(contents ...*genai.Content) (*genai.GenerateContentResponse, error)
	GenerateContentWithToolsFunc func(ctx context.Context, history []*genai.Content, tools []*genai.Tool) (*genai.GenerateContentResponse, error)
	ExecuteToolFunc            func(ctx context.Context, fc *genai.FunctionCall) (types.ToolResult, error)
	SendMessageStreamFunc      func(modelName string, messageParams types.MessageParams, promptId string) (<-chan types.StreamResponse, error)
	ListModelsFunc             func() ([]string, error)
	GetHistoryFunc             func() ([]*genai.Content, error)
	SetHistoryFunc             func(history []*genai.Content) error
	CompressChatFunc           func(promptId string, force bool) (*types.ChatCompressionResult, error)
	GenerateStreamFunc         func(ctx context.Context, contents ...*genai.Content) (<-chan any, error)
	toolRegistry               types.ToolRegistryInterface
}

// NewRealisticMockExecutor creates a mock executor that simulates a realistic workflow.
func NewRealisticMockExecutor(toolRegistry types.ToolRegistryInterface) *MockExecutor {
	mock := &MockExecutor{
		toolRegistry: toolRegistry,
	}

	// Implement the real ExecuteTool method
	mock.ExecuteToolFunc = func(ctx context.Context, fc *genai.FunctionCall) (types.ToolResult, error) {
		tool, err := mock.toolRegistry.GetTool(fc.Name)
		if err != nil {
			return types.ToolResult{}, err
		}
		return tool.Execute(ctx, fc.Args)
	}

	// Mock GenerateStream to follow a realistic script of events and execute real tools.
	mock.GenerateStreamFunc = func(ctx context.Context, contents ...*genai.Content) (<-chan any, error) {
		eventChan := make(chan any)

		go func() {
			defer close(eventChan)
			eventChan <- types.StreamingStartedEvent{}

			steps := []types.ToolCallStartEvent{
				{ToolCallID: "mock-1", ToolName: "execute_command", Args: map[string]interface{}{"command": "mkdir my-express-app"}},
				{ToolCallID: "mock-2", ToolName: "execute_command", Args: map[string]interface{}{"command": "npm init -y", "dir": "my-express-app"}},
				{ToolCallID: "mock-3", ToolName: "read_file", Args: map[string]interface{}{"file_path": "my-express-app/package.json"}},
				{ToolCallID: "mock-4", ToolName: "execute_command", Args: map[string]interface{}{"command": "npm install express", "dir": "my-express-app"}},
				{ToolCallID: "mock-5", ToolName: "write_file", Args: map[string]interface{}{"file_path": "my-express-app/index.js", "content": "const express = require('express');\nconst app = express();\nconst port = 3000;\n\napp.get('/', (req, res) => {\n  res.send('Hello World!');\n});\n\napp.listen(port, () => {\n  console.log(`Example app listening on port ${port}`);\n});"}},
				{ToolCallID: "mock-6", ToolName: "read_file", Args: map[string]interface{}{"file_path": "my-express-app/index.js"}},
				{ToolCallID: "mock-7", ToolName: "grep", Args: map[string]interface{}{"pattern": "func", "path": "pkg/tools", "include": "*.go"}},
				{ToolCallID: "mock-8", ToolName: "glob", Args: map[string]interface{}{"pattern": "*.go", "path": "pkg/tools"}},
				{ToolCallID: "mock-9", ToolName: "list_directory", Args: map[string]interface{}{"path": "pkg/tools"}},
				{ToolCallID: "mock-10", ToolName: "smart_edit", Args: map[string]interface{}{"file_path": "testdata/temp.txt", "old_string": "old content", "new_string": "new content"}},
			}

			for _, step := range steps {
				eventChan <- types.ThinkingEvent{}
				time.Sleep(1 * time.Second) // Simulate model thinking time

				eventChan <- step // Emit ToolCallStartEvent

				toolResult, err := mock.ExecuteTool(ctx, &genai.FunctionCall{Name: step.ToolName, Args: step.Args})

				time.Sleep(500 * time.Millisecond) // Simulate tool execution time

				eventChan <- types.ToolCallEndEvent{
					ToolCallID: step.ToolCallID,
					ToolName:   step.ToolName,
					Result:     toolResult.ReturnDisplay,
					Err:        err,
				}
			}

			// Final Response
			eventChan <- types.ThinkingEvent{}
			time.Sleep(1 * time.Second)
			eventChan <- types.FinalResponseEvent{Content: "I have created the Express API in `my-express-app/index.js` and verified the contents of the files."}
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
func (m *MockExecutor) ExecuteTool(ctx context.Context, fc *genai.FunctionCall) (types.ToolResult, error) {
	if m.ExecuteToolFunc != nil {
		return m.ExecuteToolFunc(ctx, fc)
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
func (m *MockExecutor) GenerateStream(ctx context.Context, contents ...*genai.Content) (<-chan any, error) {
	if m.GenerateStreamFunc != nil {
		return m.GenerateStreamFunc(ctx, contents...)
	}
	respChan := make(chan any)
	close(respChan)
	return respChan, fmt.Errorf("GenerateStream not implemented in mock")
}