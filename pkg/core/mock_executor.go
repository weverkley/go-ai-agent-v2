package core

import (
	"context"
	"fmt"
	"os"
	"path"
	"time"

	"go-ai-agent-v2/go-cli/pkg/telemetry"
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
	GenerateStreamFunc         func(ctx context.Context, contents ...*types.Content) (<-chan any, error)
	toolRegistry               types.ToolRegistryInterface
	UserConfirmationChan       chan bool // Expose for testing
}

// NewRealisticMockExecutor creates a mock executor that simulates a realistic workflow.
func NewRealisticMockExecutor(toolRegistry types.ToolRegistryInterface) *MockExecutor {
	mock := &MockExecutor{
		toolRegistry: toolRegistry,
	}

	// Implement the real ExecuteTool method
	mock.ExecuteToolFunc = func(ctx context.Context, fc *types.FunctionCall) (types.ToolResult, error) {
		tool, err := mock.toolRegistry.GetTool(fc.Name)
		if err != nil {
			return types.ToolResult{}, err
		}
		return tool.Execute(ctx, fc.Args)
	}

	// Mock GenerateStream to follow a realistic script of events and execute real tools.
	mock.GenerateStreamFunc = func(ctx context.Context, contents ...*types.Content) (<-chan any, error) {
		eventChan := make(chan any)

		go func() {
			defer close(eventChan)
			eventChan <- types.StreamingStartedEvent{}

			// Get current working directory for absolute paths
			cwd, err := os.Getwd()
			if err != nil {
				eventChan <- types.ErrorEvent{Err: fmt.Errorf("failed to get current working directory: %w", err)}
				return
			}
			todoAPIPath := path.Join(cwd, "todo-api")
			todoAPIIndexPath := path.Join(todoAPIPath, "index.js")

			steps := []types.ToolCallStartEvent{
				// Phase 1: Project Setup and Basic Server
				{ToolCallID: "mock-1", ToolName: "execute_command", Args: map[string]interface{}{"command": "mkdir todo-api"}},
				{ToolCallID: "mock-2", ToolName: "execute_command", Args: map[string]interface{}{"command": "cd todo-api && npm init -y"}},
				{ToolCallID: "mock-3", ToolName: "execute_command", Args: map[string]interface{}{"command": "cd todo-api && npm install express body-parser"}},
				{ToolCallID: "mock-4", ToolName: "write_file", Args: map[string]interface{}{"file_path": todoAPIIndexPath, "content": "const express = require('express');\nconst bodyParser = require('body-parser');\nconst app = express();\nconst port = 3000;\n\napp.use(bodyParser.json());\n\nlet todos = []; // In-memory storage for todos\n\napp.get('/', (req, res) => {\n  res.send('Todo API is running!');\n});\n\napp.listen(port, () => {\n  console.log(`Todo API listening on port ${port}`);\n});"}},
				{ToolCallID: "mock-5", ToolName: "read_file", Args: map[string]interface{}{"file_path": todoAPIIndexPath}},
				{ToolCallID: "mock-6", ToolName: "user_confirm", Args: map[string]interface{}{"message": "Please run 'cd todo-api && node index.js &' in another terminal. Press 'continue' once the server is running."}},

				{ToolCallID: "mock-7", ToolName: "web_fetch", Args: map[string]interface{}{"prompt": "http://localhost:3000"}},

				// Phase 2: Implement Todo Routes
				{ToolCallID: "mock-8", ToolName: "smart_edit", Args: map[string]interface{}{
					"file_path":   todoAPIIndexPath,
					"instruction": "Add GET all todos route",
					"old_string":  "app.get('/', (req, res) => {\n  res.send('Todo API is running!');\n});",
					"new_string":  "app.get('/', (req, res) => {\n  res.send('Todo API is running!');\n});\n\n// GET all todos\napp.get('/todos', (req, res) => {\n  res.json(todos);\n});",
				}},
				{ToolCallID: "mock-9", ToolName: "smart_edit", Args: map[string]interface{}{
					"file_path":   todoAPIIndexPath,
					"instruction": "Add POST new todo route",
					"old_string":  "app.get('/todos', (req, res) => {\n  res.json(todos);\n});",
					"new_string":  "app.get('/todos', (req, res) => {\n  res.json(todos);\n});\n\n// POST a new todo\napp.post('/todos', (req, res) => {\n  const newTodo = { id: todos.length + 1, title: req.body.title, completed: false };\n  todos.push(newTodo);\n  res.status(201).json(newTodo);\n});",
				}},
				{ToolCallID: "mock-10", ToolName: "smart_edit", Args: map[string]interface{}{
					"file_path":   todoAPIIndexPath,
					"instruction": "Add GET todo by ID route",
					"old_string":  "app.post('/todos', (req, res) => {\n  const newTodo = { id: todos.length + 1, title: req.body.title, completed: false };\n  todos.push(newTodo);\n  res.status(201).json(newTodo);\n});",
					"new_string":  "app.post('/todos', (req, res) => {\n  const newTodo = { id: todos.length + 1, title: req.body.title, completed: false };\n  todos.push(newTodo);\n  res.status(201).json(newTodo);\n});\n\n// GET todo by ID\napp.get('/todos/:id', (req, res) => {\n  const todo = todos.find(t => t.id === parseInt(req.params.id));\n  if (!todo) return res.status(404).send('Todo not found');\n  res.json(todo);\n});",
				}},
				{ToolCallID: "mock-11", ToolName: "smart_edit", Args: map[string]interface{}{
				    "file_path":   todoAPIIndexPath,					"instruction": "Add PUT update todo route",
					"old_string":  "app.get('/todos/:id', (req, res) => {\n  const todo = todos.find(t => t.id === parseInt(req.params.id));\n  if (!todo) return res.status(404).send('Todo not found');\n  res.json(todo);\n});",
					"new_string":  "app.get('/todos/:id', (req, res) => {\n  const todo = todos.find(t => t.id === parseInt(req.params.id));\n  if (!todo) return res.status(404).send('Todo not found');\n  res.json(todo);\n});\n\n// PUT update todo\napp.put('/todos/:id', (req, res) => {\n  const todo = todos.find(t => t.id === parseInt(req.params.id));\n  if (!todo) return res.status(404).send('Todo not found');\n\n  todo.title = req.body.title !== undefined ? req.body.title : todo.title;\n  todo.completed = req.body.completed !== undefined ? req.body.completed : todo.completed;\n  res.json(todo);\n});",
				}},
				{ToolCallID: "mock-12", ToolName: "smart_edit", Args: map[string]interface{}{
					"file_path":   todoAPIIndexPath,
					"instruction": "Add DELETE todo route",
					"old_string":  "app.put('/todos/:id', (req, res) => {\n  const todo = todos.find(t => t.id === parseInt(req.params.id));\n  if (!todo) return res.status(404).send('Todo not found');\n\n  todo.title = req.body.title !== undefined ? req.body.title : todo.title;\n  todo.completed = req.body.completed !== undefined ? req.body.completed : todo.completed;\n  res.json(todo);\n});",
					"new_string":  "app.put('/todos/:id', (req, res) => {\n  const todo = todos.find(t => t.id === parseInt(req.params.id));\n  if (!todo) return res.status(404).send('Todo not found');\n\n  todo.title = req.body.title !== undefined ? req.body.title : todo.title;\n  todo.completed = req.body.completed !== undefined ? req.body.completed : todo.completed;\n  res.json(todo);\n});\n\n// DELETE a todo\napp.delete('/todos/:id', (req, res) => {\n  const index = todos.findIndex(t => t.id === parseInt(req.params.id));\n  if (index === -1) return res.status(404).send('Todo not found');\n\n  const deletedTodo = todos.splice(index, 1);\n  res.json(deletedTodo);\n});",
				}},
				{ToolCallID: "mock-13", ToolName: "read_file", Args: map[string]interface{}{"file_path": todoAPIIndexPath}},

				// Phase 3: Testing and Validation (using other tools)
				{ToolCallID: "mock-14", ToolName: "execute_command", Args: map[string]interface{}{"command": "kill $(lsof -t -i:3000) || true"}},
				{ToolCallID: "mock-15", ToolName: "user_confirm", Args: map[string]interface{}{"message": "The todo-api server was killed. Please restart it by running 'cd todo-api && node index.js &' in another terminal. Press 'continue' once the server is running."}},

				{ToolCallID: "mock-16", ToolName: "web_fetch", Args: map[string]interface{}{"prompt": "http://localhost:3000/todos"}}, // GET all - should be empty array
				{ToolCallID: "mock-17", ToolName: "web_fetch", Args: map[string]interface{}{"prompt": "http://localhost:3000/todos", "method": "POST", "body": `{"title": "Learn Go AI Agent"}`}}, // add a todo
				{ToolCallID: "mock-18", ToolName: "web_fetch", Args: map[string]interface{}{"prompt": "http://localhost:3000/todos"}}, // GET all - should have one todo
				{ToolCallID: "mock-19", ToolName: "web_fetch", Args: map[string]interface{}{"prompt": "http://localhost:3000/todos/1"}}, // GET todo by ID
				{ToolCallID: "mock-20", ToolName: "web_fetch", Args: map[string]interface{}{"prompt": "http://localhost:3000/todos/1", "method": "PUT", "body": `{"completed": true}`}}, // update todo
				{ToolCallID: "mock-21", ToolName: "web_fetch", Args: map[string]interface{}{"prompt": "http://localhost:3000/todos/1"}}, // GET todo by ID - verify update
				{ToolCallID: "mock-22", ToolName: "web_fetch", Args: map[string]interface{}{"prompt": "http://localhost:3000/todos/1", "method": "DELETE"}}, // delete todo
				{ToolCallID: "mock-23", ToolName: "web_fetch", Args: map[string]interface{}{"prompt": "http://localhost:3000/todos"}}, // GET all - should be empty again
				{ToolCallID: "mock-24", ToolName: "grep", Args: map[string]interface{}{"pattern": "app.get", "path": "todo-api", "include": "*.js"}},
				{ToolCallID: "mock-25", ToolName: "glob", Args: map[string]interface{}{"pattern": "**/*.js", "path": "todo-api"}},
				{ToolCallID: "mock-26", ToolName: "list_directory", Args: map[string]interface{}{"path": "todo-api"}},
				{ToolCallID: "mock-27", ToolName: "read_many_files", Args: map[string]interface{}{"paths": []string{path.Join(todoAPIPath, "**", "*.js")}}},
				{ToolCallID: "mock-28", ToolName: "save_memory", Args: map[string]interface{}{"fact": "Todo API development started"}},
				{ToolCallID: "mock-29", ToolName: "write_todos", Args: map[string]interface{}{"todos": []any{map[string]any{"description": "Implement user authentication", "status": "pending"}, map[string]any{"description": "Add database integration", "status": "pending"}}}},
				{ToolCallID: "mock-34", ToolName: "find_unused_code", Args: map[string]interface{}{"directory": todoAPIPath}},
				{ToolCallID: "mock-35", ToolName: "extract_function", Args: map[string]interface{}{"filePath": todoAPIIndexPath, "startLine": 10, "endLine": 15, "newFunctionName": "handleRoot"}},
				{ToolCallID: "mock-36", ToolName: "web_search", Args: map[string]interface{}{"query": "express.js best practices for todo api"}},
				{ToolCallID: "mock-37", ToolName: "web_fetch", Args: map[string]interface{}{"prompt": "summarize https://expressjs.com/en/guide/routing.html"}},
				{ToolCallID: "mock-38", ToolName: "ls", Args: map[string]interface{}{"path": todoAPIPath, "long": true}},
				{ToolCallID: "mock-39", ToolName: "execute_command", Args: map[string]interface{}{"command": "kill $(lsof -t -i:3000) || true"}},
			}

						for _, step := range steps {
							telemetry.LogDebugf("MockExecutor: Processing step: %s (%s)", step.ToolCallID, step.ToolName)
						select {
						case <-ctx.Done():
							telemetry.LogDebugf("MockExecutor: Context cancelled, exiting loop.")
							eventChan <- types.ErrorEvent{Err: ctx.Err()}
							return
						default:
							// Continue if context is not cancelled
						}
							
							eventChan <- types.ThinkingEvent{}
							time.Sleep(100 * time.Millisecond) // Simulate thinking time

							eventChan <- step // Emit ToolCallStartEvent
							
							// Add a small delay to allow the UI to render the start event
							time.Sleep(100 * time.Millisecond)

							var result types.ToolResult
							var err error

							if step.ToolName == types.USER_CONFIRM_TOOL_NAME {
								// Simulate user confirming immediately
								result = types.ToolResult{
									LLMContent:    "continue",
									ReturnDisplay: "User confirmed tool execution (mock).",
								}
								err = nil // No error
							} else if step.ToolName == types.WEB_FETCH_TOOL_NAME {
								if prompt, ok := step.Args["prompt"].(string); ok && prompt == "http://localhost:3000" {
									result = types.ToolResult{
										LLMContent:    "Mock server is running.",
										ReturnDisplay: "Mock server is running.",
									}
									err = nil
								} else {
									result, err = mock.ExecuteTool(ctx, &types.FunctionCall{Name: step.ToolName, Args: step.Args})
								}
							} else {
								result, err = mock.ExecuteTool(ctx, &types.FunctionCall{Name: step.ToolName, Args: step.Args})
							}

							time.Sleep(500 * time.Millisecond) // Simulate tool execution time

							eventChan <- types.ToolCallEndEvent{
								ToolCallID: step.ToolCallID,
								ToolName:   step.ToolName,
								Result:     result.ReturnDisplay,
								Err:        err,
							}
							telemetry.LogDebugf("MockExecutor: Finished step: %s (%s) with error: %v", step.ToolCallID, step.ToolName, err)
						}
			// Final Response
			select {
			case <-ctx.Done():
				eventChan <- types.ErrorEvent{Err: ctx.Err()}
				return
			default:
				eventChan <- types.ThinkingEvent{}
				time.Sleep(1 * time.Second)
				eventChan <- types.FinalResponseEvent{Content: "I have created the Express API in `todo-api/index.js` and verified the contents of the files."}
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
	
// GenerateStream mocks the GenerateStream method.
func (m *MockExecutor) GenerateStream(ctx context.Context, contents ...*types.Content) (<-chan any, error) {
	if m.GenerateStreamFunc != nil {
		return m.GenerateStreamFunc(ctx, contents...)
	}
	// Return nil channel and error if not implemented
	return nil, fmt.Errorf("GenerateStream not implemented in mock")
}

// SetUserConfirmationChannel mocks the SetUserConfirmationChannel method.
func (m *MockExecutor) SetUserConfirmationChannel(ch chan bool) {
	m.UserConfirmationChan = ch
}