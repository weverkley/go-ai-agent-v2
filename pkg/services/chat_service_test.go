package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"go-ai-agent-v2/go-cli/pkg/core"
	"go-ai-agent-v2/go-cli/pkg/types"
	"go-ai-agent-v2/go-cli/pkg/config" // New import
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockWriteFileTool for testing
type MockWriteFileTool struct {
	types.BaseDeclarativeTool
}

func (t *MockWriteFileTool) Name() string        { return types.WRITE_FILE_TOOL_NAME }
func (t *MockWriteFileTool) Description() string { return "Writes content to a file." }
func (t *MockWriteFileTool) Parameters() *types.JsonSchemaObject {
	return types.NewJsonSchemaObject().SetProperties(map[string]*types.JsonSchemaProperty{
		"file_path": {
			Type:        "string",
			Description: "The path to the file to write.",
		},
		"content": {
			Type:        "string",
			Description: "The content to write to the file.",
		},
	}).SetRequired([]string{"file_path", "content"})
}
func (t *MockWriteFileTool) ServerName() string { return "mock" }
func (t *MockWriteFileTool) Execute(ctx context.Context, args map[string]any) (types.ToolResult, error) {
	filePath, ok := args["file_path"].(string)
	if !ok {
		return types.ToolResult{}, fmt.Errorf("file_path argument missing or invalid")
	}
	// In a real test, you might write to a temp file.
	return types.ToolResult{
		LLMContent:    fmt.Sprintf("Successfully wrote to %s", filePath),
		ReturnDisplay: fmt.Sprintf("Wrote to: %s", filePath),
	}, nil
}

// MockUserConfirmTool for testing
type MockUserConfirmTool struct {
	types.BaseDeclarativeTool
}

func (t *MockUserConfirmTool) Name() string        { return types.USER_CONFIRM_TOOL_NAME }
func (t *MockUserConfirmTool) Description() string { return "Asks the user for confirmation." }
func (t *MockUserConfirmTool) Parameters() *types.JsonSchemaObject {
	return types.NewJsonSchemaObject().SetProperties(map[string]*types.JsonSchemaProperty{
		"message": {
			Type:        "string",
			Description: "The message to display to the user for confirmation.",
		},
	}).SetRequired([]string{"message"})
}
func (t *MockUserConfirmTool) ServerName() string { return "mock" }
func (t *MockUserConfirmTool) Execute(ctx context.Context, args map[string]any) (types.ToolResult, error) {
	// This tool is intercepted by the ChatService, so this Execute method shouldn't be called in the tests.
	return types.ToolResult{}, fmt.Errorf("user_confirm Execute should not be called directly")
}


// setupTestChatService creates a temporary directory and a ChatService instance for testing.
func setupTestChatService(t *testing.T) (*ChatService, *core.MockExecutor, *SessionService, types.ToolRegistryInterface, *MockSettingsService, types.Config, string, func()) {
	projectRoot, err := os.MkdirTemp("", "chat_service_test_*")
	assert.NoError(t, err)

	goaiagentDir := filepath.Join(projectRoot, ".goaiagent")
	err = os.Mkdir(goaiagentDir, 0755)
	assert.NoError(t, err)

	sessionService, err := NewSessionService(goaiagentDir)
	assert.NoError(t, err)

	mockSettingsService := new(MockSettingsService)
	mockSettingsService.On("Get", mock.Anything).Return(nil, false).Maybe()
	mockSettingsService.On("GetWorkspaceDir").Return(projectRoot).Maybe()


	toolRegistry := types.NewToolRegistry()
	toolRegistry.Register(&MockWriteFileTool{})
	toolRegistry.Register(&MockUserConfirmTool{})
	
	mockExecutor := &core.MockExecutor{}

	appConfig := config.NewConfig(&config.ConfigParameters{}) // Create a simple mock config

	chatService, err := NewChatService(mockExecutor, toolRegistry, sessionService, "test_session_id", mockSettingsService, appConfig, nil)
	assert.NoError(t, err)

	cleanup := func() {
		os.RemoveAll(projectRoot)
		mockSettingsService.AssertExpectations(t)
	}

	return chatService, mockExecutor, sessionService, toolRegistry, mockSettingsService, appConfig, projectRoot, cleanup
}

func TestChatService_SendMessage_ToolConfirmation(t *testing.T) {
	t.Run("User confirms tool execution", func(t *testing.T) {
		chatService, mockExecutor, _, _, mockSettingsService, _, projectRoot, cleanup := setupTestChatService(t)
		defer cleanup()

		mockExecutor.StreamContentFunc = func(ctx context.Context, contents ...*types.Content) (<-chan any, error) {
			eventChan := make(chan any)
			go func() {
				defer close(eventChan)
				if len(contents) == 1 {
					eventChan <- types.Part{FunctionCall: &types.FunctionCall{
						Name: types.WRITE_FILE_TOOL_NAME,
						Args: map[string]interface{}{"file_path": filepath.Join(projectRoot, "test.txt"), "content": "Hello"},
					}}
				} else if len(contents) == 3 {
					eventChan <- types.Part{Text: "Mock: Task finished."}
				}
			}()
			return eventChan, nil
		}
		mockSettingsService.On("GetDangerousTools").Return([]string{types.WRITE_FILE_TOOL_NAME}).Once()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		eventChan, err := chatService.SendMessage(ctx, "test")
		assert.NoError(t, err)

		for event := range eventChan {
			if _, ok := event.(types.ToolConfirmationRequestEvent); ok {
				chatService.ToolConfirmationChan <- types.ToolConfirmationOutcomeProceedOnce
			}
		}
	})

	t.Run("User cancels tool execution", func(t *testing.T) {
		chatService, mockExecutor, _, _, mockSettingsService, _, _, cleanup := setupTestChatService(t)
		defer cleanup()
		
		mockExecutor.StreamContentFunc = func(ctx context.Context, contents ...*types.Content) (<-chan any, error) {
			eventChan := make(chan any)
			go func() {
				defer close(eventChan)
				if len(contents) == 1 {
					eventChan <- types.Part{FunctionCall: &types.FunctionCall{
						Name: types.WRITE_FILE_TOOL_NAME,
						Args: map[string]interface{}{"file_path": "cancel.txt", "content": "Cancel this"},
					}}
				} else if len(contents) == 3 {
					eventChan <- types.Part{Text: "Mock: Okay, cancelled."}
				}
			}()
			return eventChan, nil
		}
		mockSettingsService.On("GetDangerousTools").Return([]string{types.WRITE_FILE_TOOL_NAME}).Once()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		eventChan, err := chatService.SendMessage(ctx, "test")
		assert.NoError(t, err)

		var endEvent types.ToolCallEndEvent
		for event := range eventChan {
			if _, ok := event.(types.ToolConfirmationRequestEvent); ok {
				chatService.ToolConfirmationChan <- types.ToolConfirmationOutcomeCancel
			}
			if e, ok := event.(types.ToolCallEndEvent); ok {
				endEvent = e
			}
		}

		assert.Error(t, endEvent.Err)
		assert.Contains(t, endEvent.Result, "Tool execution cancelled by user")
	})
}
func TestChatService_SendMessage_UserConfirmTool(t *testing.T) {
	t.Run("User confirms user_confirm tool", func(t *testing.T) {
		chatService, mockExecutor, _, _, mockSettingsService, _, _, cleanup := setupTestChatService(t)
		defer cleanup()
		mockSettingsService.On("GetDangerousTools").Return([]string{types.USER_CONFIRM_TOOL_NAME}).Once()

		mockExecutor.StreamContentFunc = func(currentCtx context.Context, contents ...*types.Content) (<-chan any, error) {
			eventChan := make(chan any)
			go func() {
				defer close(eventChan)
				if len(contents) == 1 {
					eventChan <- types.Part{FunctionCall: &types.FunctionCall{
						Name: types.USER_CONFIRM_TOOL_NAME,
						Args: map[string]interface{}{"message": "Do you want to proceed?"},
					}}
				} else if len(contents) == 3 {
					eventChan <- types.Part{Text: "Mock: User confirmation handled."}
				}
			}()
			return eventChan, nil
		}

		eventChan, err := chatService.SendMessage(context.Background(), "test")
		assert.NoError(t, err)

		for event := range eventChan {
			if _, ok := event.(types.ToolConfirmationRequestEvent); ok {
				chatService.GetToolConfirmationChannel() <- types.ToolConfirmationOutcomeProceedOnce
			}
		}
		
		history := chatService.GetHistory()
		assert.Len(t, history, 4)
		assert.Equal(t, "continue", history[2].Parts[0].FunctionResponse.Response["result"].(map[string]interface{})["result"])
	})
}

func TestChatService_SendMessage_ProceedAlways(t *testing.T) {
	t.Run("User selects Proceed Always for a tool", func(t *testing.T) {
		chatService, mockExecutor, _, _, mockSettingsService, _, _, cleanup := setupTestChatService(t)
		defer cleanup()
		mockSettingsService.On("GetDangerousTools").Return([]string{types.WRITE_FILE_TOOL_NAME}).Twice()

		// --- First Tool Call: User selects "Proceed Always" ---
		mockExecutor.StreamContentFunc = func(ctx context.Context, contents ...*types.Content) (<-chan any, error) {
			eventChan := make(chan any)
			go func() {
				defer close(eventChan)
				if len(contents) == 1 {
					eventChan <- types.Part{FunctionCall: &types.FunctionCall{
						Name: types.WRITE_FILE_TOOL_NAME,
						Args: map[string]interface{}{"file_path": "test_always.txt", "content": "Always proceed content."},
					}}
				} else if len(contents) == 3 {
					eventChan <- types.Part{Text: "Mock: First call done."}
				}
			}()
			return eventChan, nil
		}

		eventChan1, err := chatService.SendMessage(context.Background(), "First call")
		assert.NoError(t, err)

		for event := range eventChan1 {
			if _, ok := event.(types.ToolConfirmationRequestEvent); ok {
				chatService.ToolConfirmationChan <- types.ToolConfirmationOutcomeProceedAlways
			}
		}

		// --- Second Tool Call: Verify automatic execution ---
		mockExecutor.StreamContentFunc = func(ctx context.Context, contents ...*types.Content) (<-chan any, error) {
			eventChan := make(chan any)
			go func() {
				defer close(eventChan)
				if len(contents) == 5 {
					eventChan <- types.Part{FunctionCall: &types.FunctionCall{
						Name: types.WRITE_FILE_TOOL_NAME,
						Args: map[string]interface{}{"file_path": "test_always.txt", "content": "Second call content."},
					}}
				} else if len(contents) == 7 {
					eventChan <- types.Part{Text: "Mock: Second call done."}
				}
			}()
			return eventChan, nil
		}

		eventChan2, err := chatService.SendMessage(context.Background(), "Second call")
		assert.NoError(t, err)

		var confirmationRequested bool
		for event := range eventChan2 {
			if _, ok := event.(types.ToolConfirmationRequestEvent); ok {
				confirmationRequested = true
			}
		}
		assert.False(t, confirmationRequested, "Tool confirmation should NOT be requested for the second call.")
	})
}
