package core

import (
	"context"
	"fmt"
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
	UserConfirmationChan       chan bool                            // Expose for testing
	ToolConfirmationChan       chan types.ToolConfirmationOutcome // New channel for rich confirmation
	MockStep                   int                                  // State for the realistic mock flow
}

const (
	jsContentTodoApi = `const express = require('express');
const bodyParser = require('body-parser');
const app = express();
const port = 3000;

app.use(bodyParser.json());

let todos = []; // In-memory storage for todos

app.get('/', (req, res) => {
  res.send('Todo API is running!');
});

app.listen(port, () => {
  console.log(` + "`" + `Todo API listening on port ${port}` + "`" + `);
});`

	jsOldStringTodoApiGet = `app.get('/', (req, res) => {
  res.send('Todo API is running!');
});`

	jsNewStringTodoApiGet = `app.get('/', (req, res) => {
  res.send('Todo API is running!');
});

// GET all todos
app.get('/todos', (req, res) => {
  res.json(todos);
});`

	jsOldStringTodoApiPost = `app.get('/todos', (req, res) => {
  res.json(todos);
});`

	jsNewStringTodoApiPost = `app.get('/todos', (req, res) => {
  res.json(todos);
});

// POST a new todo
app.post('/todos', (req, res) => {
  const newTodo = { id: todos.length + 1, title: req.body.title, completed: false };
  todos.push(newTodo);
  res.status(201).json(newTodo);
});`

	jsOldStringTodoApiGetById = `app.post('/todos', (req, res) => {
  const newTodo = { id: todos.length + 1, title: req.body.title, completed: false };
  todos.push(newTodo);
  res.status(201).json(newTodo);
});`

	jsNewStringTodoApiGetById = `app.post('/todos', (req, res) => {
  const newTodo = { id: todos.length + 1, title: req.body.title, completed: false };
  todos.push(newTodo);
  res.status(201).json(newTodo);
});

// GET todo by ID
app.get('/todos/:id', (req, res) => {
  const todo = todos.find(t => t.id === parseInt(req.params.id));
  if (!todo) return res.status(404).send('Todo not found');
  res.json(todo);
});`
)

// NewRealisticMockExecutor creates a mock executor that simulates a realistic workflow
// for creating a simple Node.js Todo API.
func NewRealisticMockExecutor(toolRegistry types.ToolRegistryInterface) *MockExecutor {
	mock := &MockExecutor{
		toolRegistry: toolRegistry,
		MockStep:     0,
	}

	// Implement the real ExecuteTool method for the mock's tools
	mock.ExecuteToolFunc = func(ctx context.Context, fc *types.FunctionCall) (types.ToolResult, error) {
		if fc.Name == types.EXECUTE_COMMAND_TOOL_NAME {
			command, _ := fc.Args["command"].(string)
			return types.ToolResult{
				LLMContent:    fmt.Sprintf("Mock command '%s' executed successfully.", command),
				ReturnDisplay: fmt.Sprintf("Mock command '%s' executed successfully.", command),
			}, nil
		}
		// Delegate to actual tool logic for write_todos for proper parsing
		if fc.Name == types.WRITE_TODOS_TOOL_NAME {
			tool, err := mock.toolRegistry.GetTool(types.WRITE_TODOS_TOOL_NAME)
			if err != nil {
				return types.ToolResult{}, fmt.Errorf("write_todos tool not found in mock registry: %w", err)
			}
			return tool.Execute(ctx, fc.Args)
		}
		// For other tools, just return a success
		return types.ToolResult{LLMContent: "Mock tool executed successfully.", ReturnDisplay: "Mock tool success."}, nil
	}

	mock.StreamContentFunc = func(ctx context.Context, contents ...*types.Content) (<-chan any, error) {
		eventChan := make(chan any)
		go func() {
			defer close(eventChan)

			// Helper to simulate confirmation and tool execution
			simulateToolCall := func(fc *types.FunctionCall) (map[string]any, error) {
				time.Sleep(1 * time.Second) // Simulate processing time

				isDangerous := (fc.Name == types.WRITE_FILE_TOOL_NAME || fc.Name == types.SMART_EDIT_TOOL_NAME || fc.Name == types.EXECUTE_COMMAND_TOOL_NAME || fc.Name == types.USER_CONFIRM_TOOL_NAME)

				if isDangerous {
					// For user_confirm, the message is passed in args. For others, it's generic.
					confirmationMessage := fmt.Sprintf("Confirm execution of tool '%s'?", fc.Name)
					if fc.Name == types.USER_CONFIRM_TOOL_NAME {
						if msg, ok := fc.Args["message"].(string); ok {
							confirmationMessage = msg
						}
					}

					eventChan <- types.ToolConfirmationRequestEvent{
						ToolCallID: "mock-tool-call",
						ToolName:   fc.Name,
						ToolArgs:   fc.Args,
						Type:       "exec",
						Message:    confirmationMessage,
					}

					outcome := <-mock.ToolConfirmationChan
					switch outcome {
					case types.ToolConfirmationOutcomeProceedOnce, types.ToolConfirmationOutcomeProceedAlways:
						if fc.Name == types.USER_CONFIRM_TOOL_NAME {
							return map[string]any{"result": "continue"}, nil
						}
						res, err := mock.ExecuteToolFunc(ctx, fc)
						if err != nil {
							return map[string]any{"error": err.Error()}, err
						}
						return res.LLMContent.(map[string]any), nil
					case types.ToolConfirmationOutcomeCancel:
						if fc.Name == types.USER_CONFIRM_TOOL_NAME {
							return map[string]any{"result": "cancel"}, nil
						}
						return map[string]any{"error": "Tool execution cancelled by user."}, fmt.Errorf("tool execution cancelled by user")
					default:
						return map[string]any{"error": "Unknown confirmation outcome."}, fmt.Errorf("unknown confirmation outcome")
					}
				} else {
					// Not dangerous, execute directly
					res, err := mock.ExecuteToolFunc(ctx, fc)
					if err != nil {
						return map[string]any{"error": err.Error()}, err
					}
					// Ensure res.LLMContent is a map for FunctionResponse
				if m, ok := res.LLMContent.(map[string]any); ok {
						return m, nil
					}
					return map[string]any{"result": res.LLMContent}, nil // Wrap in map if not already
				}
			}

			// Simple state machine based on history length to simulate a conversation
			switch len(contents) {
			case 1: // Initial user prompt
				fc := &types.FunctionCall{
					Name: types.WRITE_TODOS_TOOL_NAME,
					Args: map[string]interface{}{
						"todos": []interface{}{
							map[string]interface{}{"description": "Create api.js with basic Express server.", "status": "in_progress"},
							map[string]interface{}{"description": "Add GET /todos endpoint.", "status": "pending"},
							map[string]interface{}{"description": "Add POST /todos endpoint.", "status": "pending"},
							map[string]interface{}{"description": "Add GET /todos/:id endpoint.", "status": "pending"},
							map[string]interface{}{"description": "Provide final instructions.", "status": "pending"},
						},
					},
				}
				eventChan <- types.Part{FunctionCall: fc}

				res, err := simulateToolCall(fc)
				if err != nil {
					eventChan <- types.ErrorEvent{Err: err}
					return
				}
				eventChan <- types.Part{FunctionResponse: &types.FunctionResponse{Name: fc.Name, Response: res}}

			case 3: // After write_todos response (initial plan)
				fc := &types.FunctionCall{
					Name: types.WRITE_FILE_TOOL_NAME,
					Args: map[string]interface{}{
						"file_path": "api.js",
						"content":   jsContentTodoApi,
					},
				}
				eventChan <- types.Part{FunctionCall: fc}

				res, err := simulateToolCall(fc)
				if err != nil {
					eventChan <- types.ErrorEvent{Err: err}
					return
				}
				eventChan <- types.Part{FunctionResponse: &types.FunctionResponse{Name: fc.Name, Response: res}}

			case 5: // After write_file response
				// Update plan and add the first GET endpoint
				eventChan <- types.Part{FunctionCall: &types.FunctionCall{
					Name: types.WRITE_TODOS_TOOL_NAME,
					Args: map[string]interface{}{
						"todos": []interface{}{
							map[string]interface{}{"description": "Create api.js with basic Express server.", "status": "completed"},
							map[string]interface{}{"description": "Add GET /todos endpoint.", "status": "in_progress"},
							map[string]interface{}{"description": "Add POST /todos endpoint.", "status": "pending"},
							map[string]interface{}{"description": "Add GET /todos/:id endpoint.", "status": "pending"},
							map[string]interface{}{"description": "Provide final instructions.", "status": "pending"},
						},
					},
				}}
				fc := &types.FunctionCall{
					Name: types.SMART_EDIT_TOOL_NAME,
					Args: map[string]interface{}{
						"file_path":   "api.js",
						"instruction": "Add an endpoint to get all todos.",
						"old_string":  jsOldStringTodoApiGet,
						"new_string":  jsNewStringTodoApiGet,
					},
				}
				eventChan <- types.Part{FunctionCall: fc}

				res, err := simulateToolCall(fc)
				if err != nil {
					eventChan <- types.ErrorEvent{Err: err}
					return
				}
				eventChan <- types.Part{FunctionResponse: &types.FunctionResponse{Name: fc.Name, Response: res}}

			case 7: // After adding GET /todos
				// Update plan and add the POST endpoint
				eventChan <- types.Part{FunctionCall: &types.FunctionCall{
					Name: types.WRITE_TODOS_TOOL_NAME,
					Args: map[string]interface{}{
						"todos": []interface{}{
							map[string]interface{}{"description": "Create api.js with basic Express server.", "status": "completed"},
							map[string]interface{}{"description": "Add GET /todos endpoint.", "status": "completed"},
							map[string]interface{}{"description": "Add POST /todos endpoint.", "status": "in_progress"},
							map[string]interface{}{"description": "Add GET /todos/:id endpoint.", "status": "pending"},
							map[string]interface{}{"description": "Provide final instructions.", "status": "pending"},
						},
					},
				}}
				fc := &types.FunctionCall{
					Name: types.SMART_EDIT_TOOL_NAME,
					Args: map[string]interface{}{
						"file_path":   "api.js",
						"instruction": "Add an endpoint to create a new todo.",
						"old_string":  jsOldStringTodoApiPost,
						"new_string":  jsNewStringTodoApiPost,
					},
				}
				eventChan <- types.Part{FunctionCall: fc}

				res, err := simulateToolCall(fc)
				if err != nil {
					eventChan <- types.ErrorEvent{Err: err}
					return
				}
				eventChan <- types.Part{FunctionResponse: &types.FunctionResponse{Name: fc.Name, Response: res}}

			case 9: // After adding POST /todos
				// Update plan and add the GET by ID endpoint
				eventChan <- types.Part{FunctionCall: &types.FunctionCall{
					Name: types.WRITE_TODOS_TOOL_NAME,
					Args: map[string]interface{}{
						"todos": []interface{}{
							map[string]interface{}{"description": "Create api.js with basic Express server.", "status": "completed"},
							map[string]interface{}{"description": "Add GET /todos endpoint.", "status": "completed"},
							map[string]interface{}{"description": "Add POST /todos endpoint.", "status": "completed"},
							map[string]interface{}{"description": "Add GET /todos/:id endpoint.", "status": "in_progress"},
							map[string]interface{}{"description": "Provide final instructions.", "status": "pending"},
						},
					},
				}}
				fc := &types.FunctionCall{
					Name: types.SMART_EDIT_TOOL_NAME,
					Args: map[string]interface{}{
						"file_path":   "api.js",
						"instruction": "Add an endpoint to get a single todo by its ID.",
						"old_string":  jsOldStringTodoApiGetById,
						"new_string":  jsNewStringTodoApiGetById,
					},
				}
				eventChan <- types.Part{FunctionCall: fc}

				res, err := simulateToolCall(fc)
				if err != nil {
					eventChan <- types.ErrorEvent{Err: err}
					return
				}
				eventChan <- types.Part{FunctionResponse: &types.FunctionResponse{Name: fc.Name, Response: res}}

			case 11: // After adding GET by ID
				fc := &types.FunctionCall{
					Name: types.USER_CONFIRM_TOOL_NAME,
					Args: map[string]interface{}{
						"message": "I have implemented all API endpoints. Shall I provide the final instructions on how to run the server?",
					},
				}
				eventChan <- types.Part{FunctionCall: fc}

				res, err := simulateToolCall(fc)
				if err != nil {
					eventChan <- types.ErrorEvent{Err: err}
					return
				}
				eventChan <- types.Part{FunctionResponse: &types.FunctionResponse{Name: fc.Name, Response: res}}

			case 13: // After user_confirm response
				fc := &types.FunctionCall{
					Name: types.WRITE_TODOS_TOOL_NAME,
					Args: map[string]interface{}{
						"todos": []interface{}{
							map[string]interface{}{"description": "Create api.js with basic Express server.", "status": "completed"},
							map[string]interface{}{"description": "Add GET /todos endpoint.", "status": "completed"},
							map[string]interface{}{"description": "Add POST /todos endpoint.", "status": "completed"},
							map[string]interface{}{"description": "Add GET /todos/:id endpoint.", "status": "completed"},
							map[string]interface{}{"description": "Provide final instructions.", "status": "completed"},
						},
					},
				}
				eventChan <- types.Part{FunctionCall: fc}

				res, err := simulateToolCall(fc)
				if err != nil {
					eventChan <- types.ErrorEvent{Err: err}
					return
				}
				eventChan <- types.Part{FunctionResponse: &types.FunctionResponse{Name: fc.Name, Response: res}}

			case 15: // After final plan update
				finalResponse := "The `api.js` file is complete.\n\nTo run the server, first install the dependencies:\n```sh\nnpm init -y\nnpm install express body-parser\n```\n\nThen, start the server:\n```sh\nnode api.js\n```"
				eventChan <- types.Part{Text: finalResponse}

			default:
				eventChan <- types.Part{Text: "Mock: I've lost track of the conversation. Please start over."}
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

// SetToolConfirmationChannel implements the Executor interface for mock.
func (m *MockExecutor) SetToolConfirmationChannel(ch chan types.ToolConfirmationOutcome) {
	m.ToolConfirmationChan = ch
}
