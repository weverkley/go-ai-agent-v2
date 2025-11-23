package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"go-ai-agent-v2/go-cli/pkg/core"
	"go-ai-agent-v2/go-cli/pkg/types"
	"github.com/stretchr/testify/assert"
)

// MockSettingsService for testing ChatService
type MockSettingsService struct {
	mu             sync.RWMutex
	dangerousTools []string
	settings       map[string]interface{}
}

func NewMockSettingsService() *MockSettingsService {
	return &MockSettingsService{
		dangerousTools: []string{"execute_command", "write_file", "smart_edit"},
		settings:       make(map[string]interface{}),
	}
}

func (m *MockSettingsService) Get(key string) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if val, ok := m.settings[key]; ok {
		return val, true
	}
	// Provide sensible defaults for common settings if not explicitly set
	switch key {
	case "executor":
		return "mock", true
	case "model":
		return "mock-model", true
	case "debugMode":
		return false, true
	case "telemetry":
		return &types.TelemetrySettings{Enabled: false}, true
	case "approvalMode":
		return types.ApprovalModeDefault, true
	}
	return nil, false
}

func (m *MockSettingsService) Set(key string, value interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.settings[key] = value
	return nil
}

func (m *MockSettingsService) GetTelemetrySettings() *types.TelemetrySettings {
	return &types.TelemetrySettings{Enabled: false}
}
func (m *MockSettingsService) GetGoogleCustomSearchSettings() *types.GoogleCustomSearchSettings { return &types.GoogleCustomSearchSettings{} }
func (m *MockSettingsService) GetWebSearchProvider() types.WebSearchProvider                    { return types.WebSearchProviderGoogleCustomSearch }
func (m *MockSettingsService) GetTavilySettings() *types.TavilySettings                         { return &types.TavilySettings{} }
func (m *MockSettingsService) GetDangerousTools() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.dangerousTools
}
func (m *MockSettingsService) AllSettings() map[string]interface{} { return make(map[string]interface{}) }
func (m *MockSettingsService) Reset() error                        { return nil }
func (m *MockSettingsService) Save() error                         { return nil }

// setupTestChatService creates a temporary directory and a ChatService instance for testing.
func setupTestChatService(t *testing.T) (*ChatService, *core.MockExecutor, *SessionService, types.ToolRegistryInterface, *MockSettingsService, string, func()) {
	// Create a temp directory for session files
	projectRoot, err := os.MkdirTemp("", "chat_service_test_*")
	assert.NoError(t, err)

	goaiagentDir := filepath.Join(projectRoot, ".goaiagent")
	err = os.Mkdir(goaiagentDir, 0755)
	assert.NoError(t, err)

	sessionService, err := NewSessionService(goaiagentDir)
	assert.NoError(t, err)

	mockSettingsService := NewMockSettingsService()

	// Setup a minimal tool registry
	toolRegistry := types.NewToolRegistry()
	toolRegistry.Register(&MockWriteFileTool{}) // Register the mock write_file tool
	toolRegistry.Register(&MockUserConfirmTool{}) // Register the mock user_confirm tool
	mockExecutor := core.NewRealisticMockExecutor(toolRegistry) // Pass toolRegistry to mockExecutor

	chatService, err := NewChatService(mockExecutor, toolRegistry, sessionService, "test_session_id", mockSettingsService)
	assert.NoError(t, err)

	cleanup := func() {
		os.RemoveAll(projectRoot)
	}

	return chatService, mockExecutor, sessionService, toolRegistry, mockSettingsService, projectRoot, cleanup
}

func TestChatService_SendMessage_ToolConfirmation(t *testing.T) {
	// The outer chatService, mockExecutor, etc are not directly used here
	// They are passed to each sub-test's setupTestChatService call.
	// So we can ignore them with `_`
	_, _, _, _, _, _, cleanupOuter := setupTestChatService(t)
	defer cleanupOuter()
	// Scenario 1: User confirms tool execution (PROCEED_ONCE)
	t.Run("User confirms tool execution", func(t *testing.T) {
		chatService, _, _, _, mockSettingsService, _, cleanup := setupTestChatService(t) // mockExecutor not used directly in this scope
		defer cleanup()
		// Ensure write_file is considered a dangerous tool
		mockSettingsService.dangerousTools = []string{types.WRITE_FILE_TOOL_NAME}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		eventChan, err := chatService.SendMessage(ctx, "Please write 'Hello' to a file named 'test.txt'.")
		assert.NoError(t, err)

		var receivedEvents []any
		for event := range eventChan {
			receivedEvents = append(receivedEvents, event)

			if _, ok := event.(types.ToolConfirmationRequestEvent); ok {
				assert.Equal(t, types.WRITE_FILE_TOOL_NAME, event.(types.ToolConfirmationRequestEvent).ToolName)
				assert.Contains(t, event.(types.ToolConfirmationRequestEvent).Message, "Apply this change?")
				chatService.ToolConfirmationChan <- types.ToolConfirmationOutcomeProceedOnce
			}
		}

		// Assertions for the sequence of events
		assert.Len(t, receivedEvents, 7)

		assert.IsType(t, types.StreamingStartedEvent{}, receivedEvents[0])
		assert.IsType(t, types.ThinkingEvent{}, receivedEvents[1])
		assert.IsType(t, types.ToolCallStartEvent{}, receivedEvents[2])
		assert.IsType(t, types.ToolConfirmationRequestEvent{}, receivedEvents[3])
		assert.IsType(t, types.ToolCallEndEvent{}, receivedEvents[4])
		assert.IsType(t, types.ThinkingEvent{}, receivedEvents[5]) // Second ThinkingEvent
		assert.IsType(t, types.FinalResponseEvent{}, receivedEvents[6])

		toolCallEndEvent := receivedEvents[4].(types.ToolCallEndEvent) // Corrected index
		assert.NoError(t, toolCallEndEvent.Err)
		assert.Contains(t, toolCallEndEvent.Result, "Successfully wrote to test_file.txt")

		finalResponseEvent := receivedEvents[6].(types.FinalResponseEvent) // Corrected index
		assert.Contains(t, finalResponseEvent.Content, "Mock: Based on your confirmation, I have finished the task.")

		// Verify history
		history := chatService.GetHistory()
		assert.Len(t, history, 4) // User message, Model's FunctionCall, User's FunctionResponse, Model's final response
		assert.Equal(t, "user", history[0].Role)
		assert.Equal(t, "model", history[1].Role)
		assert.Equal(t, "user", history[2].Role) // Tool response
		assert.NotNil(t, history[2].Parts[0].FunctionResponse)
		assert.Contains(t, history[2].Parts[0].FunctionResponse.Response["result"].(string), "Successfully wrote to test_file.txt")
	})

	// Scenario 2: User cancels tool execution
	t.Run("User cancels tool execution", func(t *testing.T) {
		chatService, localMockExecutor, _, _, localMockSettingsService, _, cleanup := setupTestChatService(t)
		defer cleanup()
		// Reset mockExecutor state for new run
		localMockExecutor.MockStep = 0
		localMockSettingsService.dangerousTools = []string{types.WRITE_FILE_TOOL_NAME} // Use localMockSettingsService

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		eventChan, err := chatService.SendMessage(ctx, "Please write 'Cancel this' to a file named 'cancel.txt'.")
		assert.NoError(t, err)

		var receivedEvents []any
		for event := range eventChan {
			receivedEvents = append(receivedEvents, event)

			if _, ok := event.(types.ToolConfirmationRequestEvent); ok {
				chatService.ToolConfirmationChan <- types.ToolConfirmationOutcomeCancel
			}
		}

		// Assertions for the sequence of events
		assert.Len(t, receivedEvents, 7)

		assert.IsType(t, types.StreamingStartedEvent{}, receivedEvents[0])
		assert.IsType(t, types.ThinkingEvent{}, receivedEvents[1])
		assert.IsType(t, types.ToolCallStartEvent{}, receivedEvents[2])
		assert.IsType(t, types.ToolConfirmationRequestEvent{}, receivedEvents[3])
		assert.IsType(t, types.ToolCallEndEvent{}, receivedEvents[4])
		assert.IsType(t, types.ThinkingEvent{}, receivedEvents[5]) // Second ThinkingEvent
		assert.IsType(t, types.FinalResponseEvent{}, receivedEvents[6])

		toolCallEndEvent := receivedEvents[4].(types.ToolCallEndEvent) // Corrected index
		assert.Error(t, toolCallEndEvent.Err)
		assert.Contains(t, toolCallEndEvent.Result, "Tool execution cancelled by user.")
		assert.Contains(t, toolCallEndEvent.Err.Error(), "tool execution cancelled by user")

				finalResponseEvent := receivedEvents[6].(types.FinalResponseEvent) // Corrected index

				assert.Contains(t, finalResponseEvent.Content, "Mock: Based on your confirmation, I have finished the task.")

			}) // Closing brace for t.Run
} // Closing brace for TestChatService_SendMessage_ToolConfirmation
func TestChatService_SendMessage_UserConfirmTool(t *testing.T) {
	_, _, _, _, _, _, cleanupOuter := setupTestChatService(t)
	defer cleanupOuter()



	// Scenario 1: User confirms (true)
	t.Run("User confirms user_confirm tool", func(t *testing.T) {
		chatService, mockExecutor, _, toolRegistry, _, _, cleanup := setupTestChatService(t)
		defer cleanup()
		toolRegistry.Register(&MockUserConfirmTool{})

		localCtx, localCancel := context.WithCancel(context.Background())
		defer localCancel()

		mockExecutor.StreamContentFunc = func(currentCtx context.Context, contents ...*types.Content) (<-chan any, error) {
			eventChan := make(chan any)
			go func() {
				defer close(eventChan)

				currentHistoryLength := len(contents)

				if currentHistoryLength == 1 { // Initial user message, model proposes user_confirm tool
					toolCallID := "mock-user-confirm-1"
					message := "Do you want to proceed with the operation?"

					eventChan <- types.Part{FunctionCall: &types.FunctionCall{
						ID:   toolCallID,
						Name: types.USER_CONFIRM_TOOL_NAME,
						Args: map[string]interface{}{"message": message},
					}}
				} else if currentHistoryLength == 3 { // After user_confirm tool has been handled by ChatService
					// The ChatService will append the tool response to history, now the mock provides final text
					eventChan <- types.Part{Text: "Mock: User confirmation handled."}
				} else {
					eventChan <- types.Part{Text: "Mock: Unexpected stream content request."}
				}
			}()
			return eventChan, nil
		}

		eventChan, err := chatService.SendMessage(localCtx, "Ask for confirmation.")
		assert.NoError(t, err)

		var receivedEvents []any
		for event := range eventChan {
			receivedEvents = append(receivedEvents, event)
			if _, ok := event.(types.UserConfirmationRequestEvent); ok {
				mockExecutor.UserConfirmationChan <- true
			}
		}

		assert.Len(t, receivedEvents, 5)

		assert.IsType(t, types.StreamingStartedEvent{}, receivedEvents[0])
		assert.IsType(t, types.ThinkingEvent{}, receivedEvents[1])
		assert.IsType(t, types.UserConfirmationRequestEvent{}, receivedEvents[2])
		assert.IsType(t, types.ThinkingEvent{}, receivedEvents[3]) // Second ThinkingEvent
		assert.IsType(t, types.FinalResponseEvent{}, receivedEvents[4])

		finalResponseEvent := receivedEvents[4].(types.FinalResponseEvent) // Corrected index
		assert.Contains(t, finalResponseEvent.Content, "Mock: User confirmation handled.")

		history := chatService.GetHistory()
		assert.Len(t, history, 4) // User message, Model's FunctionCall, User's FunctionResponse, Model's final response
		assert.Contains(t, history[2].Parts[0].FunctionResponse.Response["result"].(string), "continue")
	})

	// Scenario 2: User cancels (false)
	t.Run("User cancels user_confirm tool", func(t *testing.T) {
		chatService, mockExecutor, _, toolRegistry, _, _, cleanup := setupTestChatService(t)
		defer cleanup()
		toolRegistry.Register(&MockUserConfirmTool{})

		localCtx, localCancel := context.WithCancel(context.Background())
		defer localCancel()

		mockExecutor.StreamContentFunc = func(currentCtx context.Context, contents ...*types.Content) (<-chan any, error) {
			eventChan := make(chan any)
			go func() {
				defer close(eventChan)

				currentHistoryLength := len(contents)

				if currentHistoryLength == 1 { // Initial user message, model proposes user_confirm tool
					toolCallID := "mock-user-confirm-1"
					message := "Do you want to proceed with the operation?"

					eventChan <- types.Part{FunctionCall: &types.FunctionCall{
						ID:   toolCallID,
						Name: types.USER_CONFIRM_TOOL_NAME,
						Args: map[string]interface{}{"message": message},
					}}
				} else if currentHistoryLength == 3 { // After user_confirm tool has been handled by ChatService
					// The ChatService will append the tool response to history, now the mock provides final text
					eventChan <- types.Part{Text: "Mock: User confirmation handled."}
				} else {
					eventChan <- types.Part{Text: "Mock: Unexpected stream content request."}
				}
			}()
			return eventChan, nil
		}


		eventChan, err := chatService.SendMessage(localCtx, "Ask for confirmation, then cancel.")
		assert.NoError(t, err)

		var receivedEvents []any
		for event := range eventChan {
			receivedEvents = append(receivedEvents, event)
			if _, ok := event.(types.UserConfirmationRequestEvent); ok {
				mockExecutor.UserConfirmationChan <- false
			}
		}

		assert.Len(t, receivedEvents, 5)

		assert.IsType(t, types.StreamingStartedEvent{}, receivedEvents[0])
		assert.IsType(t, types.ThinkingEvent{}, receivedEvents[1])
		assert.IsType(t, types.UserConfirmationRequestEvent{}, receivedEvents[2])
		assert.IsType(t, types.ThinkingEvent{}, receivedEvents[3]) // Second ThinkingEvent
		assert.IsType(t, types.FinalResponseEvent{}, receivedEvents[4])

		finalResponseEvent := receivedEvents[4].(types.FinalResponseEvent)
		assert.Contains(t, finalResponseEvent.Content, "Mock: User confirmation handled.")

		history := chatService.GetHistory()
		assert.Len(t, history, 4)
		assert.Contains(t, history[2].Parts[0].FunctionResponse.Response["result"].(string), "cancel")
	})
}

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
	content, ok := args["content"].(string)
	if !ok {
		return types.ToolResult{}, fmt.Errorf("content argument missing or invalid")
	}
	// Simulate writing to file by creating a dummy file
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return types.ToolResult{}, fmt.Errorf("failed to write mock file: %w", err)
	}
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
	message, ok := args["message"].(string)
	if !ok {
		return types.ToolResult{}, fmt.Errorf("message argument missing or invalid")
	}
	return types.ToolResult{
		LLMContent:    fmt.Sprintf("Confirmation requested: %s", message),
		ReturnDisplay: fmt.Sprintf("User Confirmation: %s", message),
	}, nil
}
