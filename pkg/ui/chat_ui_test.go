package ui

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/core"
	"go-ai-agent-v2/go-cli/pkg/services"
	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mock Services ---

type MockShellExecutionService struct {
	mock.Mock
}

func (m *MockShellExecutionService) ExecuteCommand(ctx context.Context, command string, dir string) (string, string, error) {
	args := m.Called(ctx, command, dir)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockShellExecutionService) ExecuteCommandInBackground(command string, dir string) (int, error) {
	args := m.Called(command, dir)
	return args.Int(0), args.Error(1)
}

func (m *MockShellExecutionService) KillAllProcesses() {
	m.Called()
}

// MockSettingsService is a mock implementation of types.SettingsServiceIface


// --- Helper for creating a test model ---
func setupTestSessionService(t *testing.T) (*services.SessionService, func()) {
	projectRoot, err := os.MkdirTemp("", "project_root_*")
	assert.NoError(t, err)
	goaiagentDir := filepath.Join(projectRoot, ".goaiagent")
	err = os.Mkdir(goaiagentDir, 0755)
	assert.NoError(t, err)

	ss, err := services.NewSessionService(goaiagentDir)
	assert.NoError(t, err)

	cleanup := func() {
		os.RemoveAll(projectRoot)
	}
	return ss, cleanup
}

func newTestModel(t *testing.T, executor core.Executor) *ChatModel {
	dummyCommandExecutor := func(args []string) (string, error) { return "command executed", nil }
	dummyShellService := new(MockShellExecutionService)
	realGitService := services.NewGitService()
	realWorkspaceService := services.NewWorkspaceService(".")

	appConfig := config.NewConfig(&config.ConfigParameters{})

	// Setup mock settings service
	mockSettingsService := new(services.MockSettingsService)
	mockSettingsService.On("GetDangerousTools").Return([]string{}).Maybe()
	mockSettingsService.On("GetWorkspaceDir").Return("").Maybe()

	sessionService, cleanup := setupTestSessionService(t)
	t.Cleanup(cleanup) // Use t.Cleanup to automatically call the cleanup function when the test finishes.

	sessionID := "test-session"
	chatService, err := services.NewChatService(executor, types.NewToolRegistry(), sessionService, sessionID, mockSettingsService)
	assert.NoError(t, err)

	model := NewChatModel(chatService, sessionService, "mock", appConfig, dummyCommandExecutor, dummyShellService, realGitService, realWorkspaceService, sessionID)
	return model
}

func TestNewChatModel(t *testing.T) {
	executor := &core.MockExecutor{}
	model := newTestModel(t, executor)

	assert.NotNil(t, model)
	assert.Equal(t, "Ready", model.status)
	assert.False(t, model.isStreaming)
	assert.Len(t, model.messages, 2) // initial logo and tips
}

func TestUpdate_UserInput(t *testing.T) {
	// Setup
	executor := &core.MockExecutor{
		StreamContentFunc: func(ctx context.Context, contents ...*types.Content) (<-chan any, error) {
			ch := make(chan any)
			go func() {
				defer close(ch)
				ch <- types.FinalResponseEvent{Content: "response"}
			}()
			return ch, nil
		},
	}
	model := newTestModel(t, executor)
	model.textarea.SetValue("hello")

	// Execute
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd := model.Update(msg)

	// Assert
	assert.NotNil(t, newModel)
	assert.NotNil(t, cmd)

	chatModel, ok := newModel.(*ChatModel)
	assert.True(t, ok)
	assert.True(t, chatModel.isStreaming)
	assert.Equal(t, "Sending...", chatModel.status)
	// 2 initial messages + 1 user message
	assert.Len(t, chatModel.messages, 3)
	userMsg, ok := chatModel.messages[2].(UserMessage)
	assert.True(t, ok)
	assert.Equal(t, "hello", userMsg.Content)
}

func TestUpdate_SlashCommand_Clear(t *testing.T) {
	// Setup
	executor := &core.MockExecutor{}
	model := newTestModel(t, executor)
	model.messages = append(model.messages, UserMessage{Content: "test"}) // Add a message
	assert.Len(t, model.messages, 3)
	model.textarea.SetValue("/clear")

	// Execute
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd := model.Update(msg)

	// Assert
	assert.Nil(t, cmd)
	chatModel, ok := newModel.(*ChatModel)
	assert.True(t, ok)
	assert.Len(t, chatModel.messages, 2) // Should be reset to the 2 initial messages
}

func TestUpdate_SlashCommand_Quit(t *testing.T) {
	// Setup
	executor := &core.MockExecutor{}
	model := newTestModel(t, executor)
	model.textarea.SetValue("/quit")

	// Execute
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd := model.Update(msg)

	// Assert: It should dispatch a command to be executed
	assert.NotNil(t, cmd)
	assert.NotNil(t, newModel)

	// Now, simulate the end of the command execution
	finishMsg := cmd()
	finalModel, finalCmd := newModel.Update(finishMsg)

	// Assert: The final command should be tea.Quit
	assert.NotNil(t, finalModel)
	assert.IsType(t, tea.Quit(), finalCmd())
}

func TestUpdate_StreamingEvents(t *testing.T) {
	// Setup
	model := newTestModel(t, &core.MockExecutor{})
	ch := make(chan any, 5)
	ch <- types.StreamingStartedEvent{}
	ch <- types.ThinkingEvent{}
	ch <- types.ToolCallStartEvent{ToolCallID: "1", ToolName: "test-tool"}
	ch <- types.ToolCallEndEvent{ToolCallID: "1", ToolName: "test-tool", Result: "success"}
	ch <- types.FinalResponseEvent{Content: "final response"}
	close(ch)

	model.streamCh = ch
	var newModel tea.Model
	var cmd tea.Cmd
	chatModel := model

	// --- StreamingStartedEvent ---
	newModel, cmd = chatModel.Update(streamEventMsg{event: <-ch})
	chatModel = newModel.(*ChatModel)
	assert.Equal(t, "Stream started...", chatModel.status)
	assert.NotNil(t, cmd)

	// --- ThinkingEvent ---
	newModel, cmd = chatModel.Update(streamEventMsg{event: <-ch})
	chatModel = newModel.(*ChatModel)
	assert.Equal(t, "Thinking...", chatModel.status)
	assert.NotNil(t, cmd)

	// --- ToolCallStartEvent ---
	newModel, cmd = chatModel.Update(streamEventMsg{event: <-ch})
	chatModel = newModel.(*ChatModel)
	assert.Equal(t, "Executing tool...", chatModel.status)
	assert.Len(t, chatModel.messages, 3) // 2 initial + 1 tool call group
	assert.Equal(t, 1, chatModel.toolCallCount)
	assert.NotNil(t, cmd)

	// --- ToolCallEndEvent ---
	newModel, cmd = chatModel.Update(streamEventMsg{event: <-ch})
	chatModel = newModel.(*ChatModel)
	assert.Equal(t, "Got tool result...", chatModel.status)
	assert.Len(t, chatModel.messages, 3) // Still 3, end event updates the group
	assert.NotNil(t, cmd)

	// --- FinalResponseEvent ---
	newModel, cmd = chatModel.Update(streamEventMsg{event: <-ch})
	chatModel = newModel.(*ChatModel)
	assert.Len(t, chatModel.messages, 4) // 2 initial + 1 tool group + 1 final response
	assert.NotNil(t, cmd)

	// --- streamFinishMsg ---
	newModel, cmd = chatModel.Update(streamFinishMsg{})
	chatModel = newModel.(*ChatModel)
	assert.False(t, chatModel.isStreaming)
	assert.Equal(t, "Ready", chatModel.status)
	assert.Nil(t, cmd)
}


func TestUpdate_UserConfirmationFlow(t *testing.T) {
	// Setup
	executor := &core.MockExecutor{}
	model := newTestModel(t, executor)
	model.isStreaming = true // We must be streaming to receive a confirmation request

	// 1. Simulate the executor sending a UserConfirmationRequestEvent
	confirmationEvent := types.UserConfirmationRequestEvent{
		ToolCallID: "confirm-123",
		Message:    "Do you want to proceed?",
	}
	newModel, cmd := model.Update(streamEventMsg{event: confirmationEvent})

	// Assert: Model is now awaiting confirmation
	chatModel, ok := newModel.(*ChatModel)
	assert.True(t, ok)
	assert.True(t, chatModel.awaitingConfirmation)
	assert.Equal(t, "Awaiting user confirmation...", chatModel.status)
	assert.Nil(t, cmd) // Should be nil now, as the stream listener is paused

	// 2. Simulate user pressing 'c' to confirm
	newModel, cmd = chatModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	chatModel, ok = newModel.(*ChatModel)
	assert.True(t, ok)

	// Assert: Model is no longer awaiting confirmation and is resuming
	assert.False(t, chatModel.awaitingConfirmation)
	assert.Equal(t, "User confirmed. Resuming...", chatModel.status)
	assert.NotNil(t, cmd) // Should return a waitForEvent command

	// Assert: The confirmation was sent to the service's channel
	select {
	case confirmed := <-model.chatService.GetUserConfirmationChannel():
		assert.True(t, confirmed)
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for confirmation")
	}

	// 3. Simulate user pressing 'x' to cancel
	model.isStreaming = true
	model.awaitingConfirmation = false // Reset state
	newModel, _ = model.Update(streamEventMsg{event: confirmationEvent})
	chatModel, _ = newModel.(*ChatModel) // Enter confirmation state again

	newModel, cmd = chatModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	chatModel, ok = newModel.(*ChatModel)
	assert.True(t, ok)

	// Assert: Model is no longer awaiting and is cancelled
	assert.False(t, chatModel.awaitingConfirmation)
	assert.Equal(t, "User cancelled. Resuming...", chatModel.status)
	assert.NotNil(t, cmd)

	// Assert: The cancellation was sent to the service's channel
	select {
	case confirmed := <-model.chatService.GetUserConfirmationChannel():
		assert.False(t, confirmed)
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for cancellation")
	}
}
