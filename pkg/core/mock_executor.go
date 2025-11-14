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
				// Phase 1: Project Setup and Basic Server
				{ToolCallID: "mock-1", ToolName: "execute_command", Args: map[string]interface{}{"command": "mkdir todo-api"}},
				{ToolCallID: "mock-2", ToolName: "execute_command", Args: map[string]interface{}{"command": "cd todo-api && npm init -y"}},
				{ToolCallID: "mock-3", ToolName: "execute_command", Args: map[string]interface{}{"command": "cd todo-api && npm install express body-parser"}},
				{ToolCallID: "mock-4", ToolName: "write_file", Args: map[string]interface{}{"file_path": "todo-api/index.js", "content": "const express = require('express');\nconst bodyParser = require('body-parser');\nconst app = express();\nconst port = 3000;\n\napp.use(bodyParser.json());\n\nlet todos = []; // In-memory storage for todos\n\napp.get('/', (req, res) => {\n  res.send('Todo API is running!');\n});\n\napp.listen(port, () => {\n  console.log(`Todo API listening on port ${port}`);\n});"}},
				{ToolCallID: "mock-5", ToolName: "read_file", Args: map[string]interface{}{"file_path": "todo-api/index.js"}},
				{ToolCallID: "mock-6", ToolName: "execute_command", Args: map[string]interface{}{"command": "cd todo-api && node index.js &"}},
				{ToolCallID: "mock-7", ToolName: "web_fetch", Args: map[string]interface{}{"prompt": "http://localhost:3000"}},

				// Phase 2: Implement Todo Routes
				{ToolCallID: "mock-8", ToolName: "smart_edit", Args: map[string]interface{}{
					"file_path":   "todo-api/index.js",
					"instruction": "Add GET all todos route",
					"old_string":  "app.get('/', (req, res) => {\n  res.send('Todo API is running!');\n});",
					"new_string":  "app.get('/', (req, res) => {\n  res.send('Todo API is running!');\n});\n\n// GET all todos\napp.get('/todos', (req, res) => {\n  res.json(todos);\n});",
				}},
				{ToolCallID: "mock-9", ToolName: "smart_edit", Args: map[string]interface{}{
					"file_path":   "todo-api/index.js",
					"instruction": "Add POST new todo route",
					"old_string":  "app.get('/todos', (req, res) => {\n  res.json(todos);\n});",
					"new_string":  "app.get('/todos', (req, res) => {\n  res.json(todos);\n});\n\n// POST a new todo\napp.post('/todos', (req, res) => {\n  const newTodo = { id: todos.length + 1, title: req.body.title, completed: false };\n  todos.push(newTodo);\n  res.status(201).json(newTodo);\n});",
				}},
				{ToolCallID: "mock-10", ToolName: "smart_edit", Args: map[string]interface{}{
					"file_path":   "todo-api/index.js",
					"instruction": "Add GET todo by ID route",
					"old_string":  "app.post('/todos', (req, res) => {\n  const newTodo = { id: todos.length + 1, title: req.body.title, completed: false };\n  todos.push(newTodo);\n  res.status(201).json(newTodo);\n});",
					"new_string":  "app.post('/todos', (req, res) => {\n  const newTodo = { id: todos.length + 1, title: req.body.title, completed: false };\n  todos.push(newTodo);\n  res.status(201).json(newTodo);\n});\n\n// GET todo by ID\napp.get('/todos/:id', (req, res) => {\n  const todo = todos.find(t => t.id === parseInt(req.params.id));\n  if (!todo) return res.status(404).send('Todo not found');\n  res.json(todo);\n});",
				}},
				{ToolCallID: "mock-11", ToolName: "smart_edit", Args: map[string]interface{}{
					"file_path":   "todo-api/index.js",
					"instruction": "Add PUT update todo route",
					"old_string":  "app.get('/todos/:id', (req, res) => {\n  const todo = todos.find(t => t.id === parseInt(req.params.id));\n  if (!todo) return res.status(404).send('Todo not found');\n  res.json(todo);\n});",
					"new_string":  "app.get('/todos/:id', (req, res) => {\n  const todo = todos.find(t => t.id === parseInt(req.params.id));\n  if (!todo) return res.status(404).send('Todo not found');\n  res.json(todo);\n});\n\n// PUT update todo\napp.put('/todos/:id', (req, res) => {\n  const todo = todos.find(t => t.id === parseInt(req.params.id));\n  if (!todo) return res.status(404).send('Todo not found');\n\n  todo.title = req.body.title !== undefined ? req.body.title : todo.title;\n  todo.completed = req.body.completed !== undefined ? req.body.completed : todo.completed;\n  res.json(todo);\n});",
				}},
				{ToolCallID: "mock-12", ToolName: "smart_edit", Args: map[string]interface{}{
					"file_path":   "todo-api/index.js",
					"instruction": "Add DELETE todo route",
					"old_string":  "app.put('/todos/:id', (req, res) => {\n  const todo = todos.find(t => t.id === parseInt(req.params.id));\n  if (!todo) return res.status(404).send('Todo not found');\n\n  todo.title = req.body.title !== undefined ? req.body.title : todo.title;\n  todo.completed = req.body.completed !== undefined ? req.body.completed : todo.completed;\n  res.json(todo);\n});",
					"new_string":  "app.put('/todos/:id', (req, res) => {\n  const todo = todos.find(t => t.id === parseInt(req.params.id));\n  if (!todo) return res.status(404).send('Todo not found');\n\n  todo.title = req.body.title !== undefined ? req.body.title : todo.title;\n  todo.completed = req.body.completed !== undefined ? req.body.completed : todo.completed;\n  res.json(todo);\n});\n\n// DELETE a todo\napp.delete('/todos/:id', (req, res) => {\n  const index = todos.findIndex(t => t.id === parseInt(req.params.id));\n  if (index === -1) return res.status(404).send('Todo not found');\n\n  const deletedTodo = todos.splice(index, 1);\n  res.json(deletedTodo);\n});",
				}},
				{ToolCallID: "mock-13", ToolName: "read_file", Args: map[string]interface{}{"file_path": "todo-api/index.js"}},

				// Phase 3: Testing and Validation (using other tools)
				{ToolCallID: "mock-14", ToolName: "execute_command", Args: map[string]interface{}{"command": "kill $(lsof -t -i:3000) || true"}},
				{ToolCallID: "mock-15", ToolName: "execute_command", Args: map[string]interface{}{"command": "cd todo-api && node index.js &"}},
				{ToolCallID: "mock-16", ToolName: "web_fetch", Args: map[string]interface{}{"prompt": "http://localhost:3000/todos"}}, // GET all - should be empty array
				{ToolCallID: "mock-17", ToolName: "web_fetch", Args: map[string]interface{}{"prompt": "http://localhost:3000/todos", "method": "POST", "body": `{"title": "Learn Gemini CLI"}`}}, // add a todo
				{ToolCallID: "mock-18", ToolName: "web_fetch", Args: map[string]interface{}{"prompt": "http://localhost:3000/todos"}}, // GET all - should have one todo
				{ToolCallID: "mock-19", ToolName: "web_fetch", Args: map[string]interface{}{"prompt": "http://localhost:3000/todos/1"}}, // GET todo by ID
				{ToolCallID: "mock-20", ToolName: "web_fetch", Args: map[string]interface{}{"prompt": "http://localhost:3000/todos/1", "method": "PUT", "body": `{"completed": true}`}}, // update todo
				{ToolCallID: "mock-21", ToolName: "web_fetch", Args: map[string]interface{}{"prompt": "http://localhost:3000/todos/1"}}, // GET todo by ID - verify update
				{ToolCallID: "mock-22", ToolName: "web_fetch", Args: map[string]interface{}{"prompt": "http://localhost:3000/todos/1", "method": "DELETE"}}, // delete todo
				{ToolCallID: "mock-23", ToolName: "web_fetch", Args: map[string]interface{}{"prompt": "http://localhost:3000/todos"}}, // GET all - should be empty again
				{ToolCallID: "mock-24", ToolName: "grep", Args: map[string]interface{}{"pattern": "app.get", "path": "todo-api", "include": "*.js"}},
				{ToolCallID: "mock-25", ToolName: "glob", Args: map[string]interface{}{"pattern": "**/*.js", "path": "todo-api"}},
				{ToolCallID: "mock-26", ToolName: "list_directory", Args: map[string]interface{}{"path": "todo-api"}},
				{ToolCallID: "mock-27", ToolName: "read_many_files", Args: map[string]interface{}{"include": []string{"todo-api/**/*.js"}}},
				{ToolCallID: "mock-28", ToolName: "memory_tool", Args: map[string]interface{}{"fact": "Todo API development started"}},
				{ToolCallID: "mock-29", ToolName: "write_todos", Args: map[string]interface{}{"todos": []map[string]interface{}{{"description": "Implement user authentication", "status": "pending"}, {"description": "Add database integration", "status": "pending"}}}},
				{ToolCallID: "mock-30", ToolName: "checkout_branch", Args: map[string]interface{}{"branch_name": "feature/auth"}},
				{ToolCallID: "mock-31", ToolName: "get_current_branch", Args: map[string]interface{}{}},
				{ToolCallID: "mock-32", ToolName: "get_remote_url", Args: map[string]interface{}{}},
				{ToolCallID: "mock-33", ToolName: "pull", Args: map[string]interface{}{}},
				{ToolCallID: "mock-34", ToolName: "find_unused_code", Args: map[string]interface{}{"path": "todo-api"}},
				{ToolCallID: "mock-35", ToolName: "extract_function", Args: map[string]interface{}{"file_path": "todo-api/index.js", "start_line": 10, "end_line": 15, "new_function_name": "handleRoot"}},
				{ToolCallID: "mock-36", ToolName: "web_search", Args: map[string]interface{}{"query": "express.js best practices for todo api"}},
				{ToolCallID: "mock-37", ToolName: "web_fetch", Args: map[string]interface{}{"prompt": "summarize https://expressjs.com/en/guide/routing.html"}},
				{ToolCallID: "mock-38", ToolName: "ls", Args: map[string]interface{}{"path": "todo-api", "long": true}},
				{ToolCallID: "mock-39", ToolName: "execute_command", Args: map[string]interface{}{"command": "kill $(lsof -t -i:3000) || true"}},
			}

			for _, step := range steps {
				select {
				case <-ctx.Done():
					// Context was cancelled, stop streaming
					eventChan <- types.ErrorEvent{Err: ctx.Err()}
					return
				default:
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
			}

			// Final Response
			select {
			case <-ctx.Done():
				eventChan <- types.ErrorEvent{Err: ctx.Err()}
				return
			default:
				eventChan <- types.ThinkingEvent{}
				time.Sleep(1 * time.Second)
				eventChan <- types.FinalResponseEvent{Content: "I have created the Express API in `my-express-app/index.js` and verified the contents of the files."}
			}
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