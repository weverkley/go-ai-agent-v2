package core

import (
	"context"
	"fmt"
	"os"
	"path"
	"time"

	"go-ai-agent-v2/go-cli/pkg/types"
)

// MockExecutor is a mock implementation of the Executor interface for testing.
type MockExecutor struct {
	GenerateContentFunc        func(contents ...*types.Content) (*types.GenerateContentResponse, error)
	GenerateContentWithToolsFunc func(ctx context.Context, history []*types.Content, tools []types.Tool) (*types.GenerateContentResponse, error)
	ExecuteToolFunc            func(ctx context.Context, fc *types.FunctionCall) (types.ToolResult, error)
	SendMessageStreamFunc      func(modelName string, messageParams types.MessageParams, promptId string) (<-chan types.StreamResponse, error)
	ListModelsFunc             func() ([]string, error)
	GetHistoryFunc             func() ([]*types.Content, error)
	SetHistoryFunc             func(history []*types.Content) error
	CompressChatFunc           func(promptId string, force bool) (*types.ChatCompressionResult, error)
	StreamContentFunc          func(ctx context.Context, contents ...*types.Content) (<-chan any, error)
	toolRegistry               types.ToolRegistryInterface
	UserConfirmationChan       chan bool // Expose for testing
	mockStep                   int     // State for the realistic mock flow
}

// NewRealisticMockExecutor creates a mock executor that simulates a realistic workflow.
func NewRealisticMockExecutor(toolRegistry types.ToolRegistryInterface) *MockExecutor {
	mock := &MockExecutor{
		toolRegistry: toolRegistry,
		mockStep:     0,
	}

	// Implement the real ExecuteTool method for the mock's tools
	mock.ExecuteToolFunc = func(ctx context.Context, fc *types.FunctionCall) (types.ToolResult, error) {
		tool, err := mock.toolRegistry.GetTool(fc.Name)
		if err != nil {
			return types.ToolResult{}, err
		}
		return tool.Execute(ctx, fc.Args)
	}

	// This is the core mock for StreamContent. It generates a scripted sequence of events.
	mock.StreamContentFunc = func(ctx context.Context, contents ...*types.Content) (<-chan any, error) {
		eventChan := make(chan any)

		// Get current working directory for absolute paths
		cwd, _ := os.Getwd()
		todoAPIPath := path.Join(cwd, "todo-api")
		todoAPIIndexPath := path.Join(todoAPIPath, "index.js")

		// Define the scripted steps for the mock executor
		steps := []types.Part{
			// Phase 1: Project Setup and Basic Server
			{FunctionCall: &types.FunctionCall{Name: "execute_command", Args: map[string]interface{}{"command": "mkdir todo-api"}}},
			{FunctionCall: &types.FunctionCall{Name: "execute_command", Args: map[string]interface{}{"command": "cd todo-api && npm init -y"}}},
			{FunctionCall: &types.FunctionCall{Name: "execute_command", Args: map[string]interface{}{"command": "cd todo-api && npm install express body-parser"}}},
			{FunctionCall: &types.FunctionCall{Name: "write_file", Args: map[string]interface{}{"file_path": todoAPIIndexPath, "content": "const express = require('express');\nconst bodyParser = require('body-parser');\nconst app = express();\nconst port = 3000;\n\napp.use(bodyParser.json());\n\nlet todos = []; // In-memory storage for todos\n\napp.get('/', (req, res) => {\n  res.send('Todo API is running!');\n});\n\napp.listen(port, () => {\n  console.log(`Todo API listening on port ${port}`);\n});"}}},
			{FunctionCall: &types.FunctionCall{Name: "read_file", Args: map[string]interface{}{"absolute_path": todoAPIIndexPath}}},
			{FunctionCall: &types.FunctionCall{Name: "user_confirm", Args: map[string]interface{}{"message": "Please run 'cd todo-api && node index.js &' in another terminal. Press 'continue' once the server is running."}}},
			{FunctionCall: &types.FunctionCall{Name: "web_fetch", Args: map[string]interface{}{"prompt": "http://localhost:3000"}}},
			// Phase 2: Implement Todo Routes
			{FunctionCall: &types.FunctionCall{Name: "smart_edit", Args: map[string]interface{}{
				"file_path":   todoAPIIndexPath,
				"instruction": "Add GET all todos route",
				"old_string":  "app.get('/', (req, res) => {\n  res.send('Todo API is running!');\n});",
				"new_string":  "app.get('/', (req, res) => {\n  res.send('Todo API is running!');\n});\n\n// GET all todos\napp.get('/todos', (req, res) => {\n  res.json(todos);\n});",
			}}},
			{FunctionCall: &types.FunctionCall{Name: "smart_edit", Args: map[string]interface{}{
				"file_path":   todoAPIIndexPath,
				"instruction": "Add POST new todo route",
				"old_string":  "app.get('/todos', (req, res) => {\n  res.json(todos);\n});",
				"new_string":  "app.get('/todos', (req, res) => {\n  res.json(todos);\n});\n\n// POST a new todo\napp.post('/todos', (req, res) => {\n  const newTodo = { id: todos.length + 1, title: req.body.title, completed: false };\n  todos.push(newTodo);\n  res.status(201).json(newTodo);\n});",
			}}},
			{FunctionCall: &types.FunctionCall{Name: "smart_edit", Args: map[string]interface{}{
				"file_path":   todoAPIIndexPath,
				"instruction": "Add GET todo by ID route",
				"old_string":  "app.post('/todos', (req, res) => {\n  const newTodo = { id: todos.length + 1, title: req.body.title, completed: false };\n  todos.push(newTodo);\n  res.status(201).json(newTodo);\n});",
				"new_string":  "app.post('/todos', (req, res) => {\n  const newTodo = { id: todos.length + 1, title: req.body.title, completed: false };\n  todos.push(newTodo);\n  res.status(201).json(newTodo);\n});\n\n// GET todo by ID\napp.get('/todos/:id', (req, res) => {\n  const todo = todos.find(t => t.id === parseInt(req.params.id));\n  if (!todo) return res.status(404).send('Todo not found');\n  res.json(todo);\n});",
			}}},
			// Final text response
			{Text: "I have created the Express API in `todo-api/index.js` and added all the required routes."},
		}

		go func() {
			defer close(eventChan)
			// Check if the last message is a response to our user_confirm
			if len(contents) > 0 {
				lastContent := contents[len(contents)-1]
				if lastContent.Role == "user" && len(lastContent.Parts) > 0 {
					if fr := lastContent.Parts[0].FunctionResponse; fr != nil && fr.Name == "user_confirm" {
						// It is a confirmation response.
						if resp, ok := fr.Response["result"].(string); ok && resp == "continue" {
							// User confirmed. The step *after* user_confirm is at the current mockStep.
							if mock.mockStep < len(steps) {
								eventChan <- steps[mock.mockStep]
								mock.mockStep++
							}
							return
						} else {
							// User cancelled.
							eventChan <- types.Part{Text: "Operation cancelled by user."}
							// We can stop the flow here by just returning.
							return
						}
					}
				}
			}

			// Default behavior: just play the next step
			if mock.mockStep < len(steps) {
				time.Sleep(200 * time.Millisecond) // Simulate thinking
				eventChan <- steps[mock.mockStep]
				mock.mockStep++
			} else {
				// If we are out of steps, just send a generic message.
				eventChan <- types.Part{Text: "Mock flow complete. What's next?"}
			}
		}()

		return eventChan, nil
	}

	return mock
}


// GenerateContent mocks the GenerateContent method.
func (m *MockExecutor) GenerateContent(contents ...*types.Content) (*types.GenerateContentResponse, error) {
	if m.GenerateContentFunc != nil {
		return m.GenerateContentFunc(contents...)
	}
	return nil, fmt.Errorf("GenerateContent not implemented in mock")
}

// GenerateContentWithTools mocks the GenerateContentWithTools method.
func (m *MockExecutor) GenerateContentWithTools(ctx context.Context, history []*types.Content, tools []types.Tool) (*types.GenerateContentResponse, error) {
	if m.GenerateContentWithToolsFunc != nil {
		return m.GenerateContentWithToolsFunc(ctx, history, tools)
	}
	return nil, fmt.Errorf("GenerateContentWithTools not implemented in mock")
}

// ExecuteTool mocks the ExecuteTool method.
func (m *MockExecutor) ExecuteTool(ctx context.Context, fc *types.FunctionCall) (types.ToolResult, error) {
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
func (m *MockExecutor) GetHistory() ([]*types.Content, error) {
	if m.GetHistoryFunc != nil {
		return m.GetHistoryFunc()
	}
	return nil, fmt.Errorf("GetHistory not implemented in mock")
}

// SetHistory mocks the SetHistory method.
func (m *MockExecutor) SetHistory(history []*types.Content) error {
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
	
// StreamContent mocks the StreamContent method.
func (m *MockExecutor) StreamContent(ctx context.Context, contents ...*types.Content) (<-chan any, error) {
	if m.StreamContentFunc != nil {
		return m.StreamContentFunc(ctx, contents...)
	}
	// Return nil channel and error if not implemented
	return nil, fmt.Errorf("StreamContent not implemented in mock")
}

// SetUserConfirmationChannel mocks the SetUserConfirmationChannel method.
func (m *MockExecutor) SetUserConfirmationChannel(ch chan bool) {
	m.UserConfirmationChan = ch
}
