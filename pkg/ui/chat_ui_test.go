package ui

import (
	"context"
	"os"
	"testing"
	"time"

	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/core"
	"go-ai-agent-v2/go-cli/pkg/services"
	"go-ai-agent-v2/go-cli/pkg/types"

	tea "github.com/charmbracelet/bubbletea"
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
	sessionsPath, err := os.MkdirTemp("", "sessions_*")
	assert.NoError(t, err)

	store, err := services.NewFileSessionStore(sessionsPath)
	assert.NoError(t, err)

	ss, err := services.NewSessionService(store)
	assert.NoError(t, err)

	cleanup := func() {
		os.RemoveAll(sessionsPath)
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

	contextService := services.NewContextService(".")

	sessionID := "test-session"
	// Pass an empty generation config for the test setup
	generationConfig := types.GenerateContentConfig{}
	chatService, err := services.NewChatService(executor, types.ToolRegistryInterface(types.NewToolRegistry()), sessionService, mockSettingsService, contextService, appConfig, generationConfig, nil)
	assert.NoError(t, err)

	model := NewChatModel(chatService, sessionService, contextService, "mock", appConfig, dummyCommandExecutor, dummyShellService, realGitService, realWorkspaceService, sessionID)
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

func TestUpdate_ToolConfirmationFlow(t *testing.T) {
	// Helper function to run a sub-test for each confirmation outcome
	testConfirmation := func(t *testing.T, key rune, expectedOutcome types.ToolConfirmationOutcome) {
		// Setup
		executor := &core.MockExecutor{}
		model := newTestModel(t, executor)
		model.isStreaming = true        // We must be streaming to receive a confirmation request
		model.streamCh = make(chan any) // Ensure streamCh is not nil

		// 1. Simulate the service sending a ToolConfirmationRequestEvent
		confirmationEvent := types.ToolConfirmationRequestEvent{
			ToolCallID: "confirm-123",
			ToolName:   "test-tool",
			Message:    "Do you want to proceed?",
		}
		newModel, cmd := model.Update(streamEventMsg{event: confirmationEvent})

		// Assert: Model is now awaiting confirmation
		chatModel, ok := newModel.(*ChatModel)
		assert.True(t, ok)
		assert.True(t, chatModel.awaitingToolConfirmation)
		assert.Equal(t, "Awaiting tool confirmation...", chatModel.status)
		assert.NotNil(t, chatModel.toolConfirmationRequest)
		assert.Nil(t, cmd) // Should be nil now, as the stream listener is paused

		// 2. Simulate user pressing the key
		newModel, cmd = chatModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{key}})
		chatModel, ok = newModel.(*ChatModel)
		assert.True(t, ok)

		// Assert: Model is no longer awaiting confirmation and is resuming
		assert.False(t, chatModel.awaitingToolConfirmation)
		assert.Contains(t, chatModel.status, "Resuming...")
		assert.NotNil(t, cmd) // Should return a waitForEvent command

		// Assert: The confirmation was sent to the service's channel
		select {
		case outcome := <-model.chatService.GetToolConfirmationChannel():
			assert.Equal(t, expectedOutcome, outcome)
		case <-time.After(1 * time.Second):
			t.Fatalf("timed out waiting for confirmation outcome '%s'", expectedOutcome)
		}
	}

	t.Run("user confirms with 'y'", func(t *testing.T) {
		testConfirmation(t, 'y', types.ToolConfirmationOutcomeProceedOnce)
	})
	t.Run("user confirms with 'Y'", func(t *testing.T) {
		testConfirmation(t, 'Y', types.ToolConfirmationOutcomeProceedOnce)
	})
	t.Run("user confirms always with 'a'", func(t *testing.T) {
		testConfirmation(t, 'a', types.ToolConfirmationOutcomeProceedAlways)
	})
	t.Run("user confirms always with 'A'", func(t *testing.T) {
		testConfirmation(t, 'A', types.ToolConfirmationOutcomeProceedAlways)
	})
	t.Run("user cancels with 'n'", func(t *testing.T) {
		testConfirmation(t, 'n', types.ToolConfirmationOutcomeCancel)
	})
	t.Run("user cancels with 'N'", func(t *testing.T) {
		testConfirmation(t, 'N', types.ToolConfirmationOutcomeCancel)
	})
	t.Run("user cancels with 'esc'", func(t *testing.T) {
		// Setup for esc is slightly different as it uses KeyType
		executor := &core.MockExecutor{}
		model := newTestModel(t, executor)
		model.isStreaming = true
		model.streamCh = make(chan any)
		confirmationEvent := types.ToolConfirmationRequestEvent{ToolCallID: "confirm-esc", ToolName: "test-tool", Message: "..."}
		newModel, _ := model.Update(streamEventMsg{event: confirmationEvent})
		chatModel, _ := newModel.(*ChatModel)

		newModel, _ = chatModel.Update(tea.KeyMsg{Type: tea.KeyEsc})
		chatModel, _ = newModel.(*ChatModel)

		assert.False(t, chatModel.awaitingToolConfirmation)
		select {
		case outcome := <-model.chatService.GetToolConfirmationChannel():
			assert.Equal(t, types.ToolConfirmationOutcomeCancel, outcome)
		case <-time.After(1 * time.Second):
			t.Fatal("timed out waiting for confirmation outcome 'Cancel'")
		}
	})
	t.Run("user modifies with 'm'", func(t *testing.T) {
		testConfirmation(t, 'm', types.ToolConfirmationOutcomeModifyWithEditor)
	})
	t.Run("user modifies with 'M'", func(t *testing.T) {
		testConfirmation(t, 'M', types.ToolConfirmationOutcomeModifyWithEditor)
	})
}
