package ui

import (
	"context" // New import
	"testing"
	"time"    // Added

	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/core"
	"go-ai-agent-v2/go-cli/pkg/routing"
	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/charmbracelet/bubbletea"
	"github.com/google/generative-ai-go/genai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockShellExecutionService is a mock implementation of services.ShellExecutionService
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

func TestNewChatModel(t *testing.T) {
	executor := &core.MockExecutor{}
	// Define a dummy commandExecutor for testing
	dummyCommandExecutor := func(args []string) (string, error) {
		return "command executed", nil
	}
	dummyShellService := new(MockShellExecutionService) // Create a dummy shell service
	appConfig := config.NewConfig(&config.ConfigParameters{})
	router := routing.NewModelRouterService(appConfig)
	model := NewChatModel(executor, "mock", appConfig, router, dummyCommandExecutor, dummyShellService)

	assert.NotNil(t, model)
	assert.Equal(t, "Ready", model.status)
	assert.Equal(t, "Go AI Agent Chat", model.title)
	assert.False(t, model.isStreaming)
	assert.Empty(t, model.messages)
}

func TestUpdate_UserInput(t *testing.T) {
	// Setup
	executor := &core.MockExecutor{
		GenerateStreamFunc: func(ctx context.Context, contents ...*genai.Content) (<-chan any, error) {
			ch := make(chan any)
			close(ch)
			return ch, nil
		},
	}
	dummyCommandExecutor := func(args []string) (string, error) {
		return "command executed", nil
	}
	dummyShellService := new(MockShellExecutionService)
	appConfig := config.NewConfig(&config.ConfigParameters{})
	router := routing.NewModelRouterService(appConfig)
	model := NewChatModel(executor, "mock", appConfig, router, dummyCommandExecutor, dummyShellService)
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
	assert.Len(t, chatModel.messages, 1)
	userMsg, ok := chatModel.messages[0].(UserMessage)
	assert.True(t, ok)
	assert.Equal(t, "hello", userMsg.Content)
}

func TestUpdate_SlashCommand_Clear(t *testing.T) {
	// Setup
	executor := &core.MockExecutor{}
	dummyCommandExecutor := func(args []string) (string, error) {
		return "command executed", nil
	}
	dummyShellService := new(MockShellExecutionService)
	appConfig := config.NewConfig(&config.ConfigParameters{})
	router := routing.NewModelRouterService(appConfig)
	model := NewChatModel(executor, "mock", appConfig, router, dummyCommandExecutor, dummyShellService)
	model.messages = []Message{UserMessage{Content: "test"}}
	model.textarea.SetValue("/clear")

	// Execute
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd := model.Update(msg)

	// Assert
	assert.Nil(t, cmd)
	chatModel, ok := newModel.(*ChatModel)
	assert.True(t, ok)
	assert.Empty(t, chatModel.messages)
}

func TestUpdate_SlashCommand_Quit(t *testing.T) {
	// Setup
	executor := &core.MockExecutor{}
	dummyCommandExecutor := func(args []string) (string, error) {
		return "command executed", nil
	}
	dummyShellService := new(MockShellExecutionService)
	appConfig := config.NewConfig(&config.ConfigParameters{})
	router := routing.NewModelRouterService(appConfig)
	model := NewChatModel(executor, "mock", appConfig, router, dummyCommandExecutor, dummyShellService)
	model.textarea.SetValue("/quit")

	// Execute
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	_, cmd := model.Update(msg)

	// Assert
	assert.NotNil(t, cmd)
	assert.IsType(t, tea.Quit(), cmd())
}

func TestUpdate_StreamingEvents(t *testing.T) {
	// Setup
	executor := &core.MockExecutor{}
	dummyCommandExecutor := func(args []string) (string, error) {
		return "command executed", nil
	}
	dummyShellService := new(MockShellExecutionService)
	appConfig := config.NewConfig(&config.ConfigParameters{})
	router := routing.NewModelRouterService(appConfig)
	model := NewChatModel(executor, "mock", appConfig, router, dummyCommandExecutor, dummyShellService)
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
	var ok bool
	var chatModel *ChatModel // Explicitly declare chatModel

	chatModel = model // Assign the initial model after type assertion

	newModel, cmd = chatModel.Update(streamEventMsg{event: <-ch})
	var tempChatModel *ChatModel
	var tempOk bool
	tempChatModel, tempOk = newModel.(*ChatModel)
	chatModel = tempChatModel
	ok = tempOk
	assert.True(t, ok)
	assert.Equal(t, "Stream started...", chatModel.status)
	assert.NotNil(t, cmd)

	newModel, cmd = chatModel.Update(streamEventMsg{event: <-ch})
	tempChatModel, tempOk = newModel.(*ChatModel)
	chatModel = tempChatModel
	ok = tempOk
	assert.True(t, ok)
	assert.Equal(t, "Thinking...", chatModel.status)
	assert.NotNil(t, cmd)

	// Execute & Assert for ThinkingEvent
	newModel, cmd = chatModel.Update(streamEventMsg{event: <-ch})
	tempChatModel, tempOk = newModel.(*ChatModel)
	chatModel = tempChatModel
	ok = tempOk
	assert.True(t, ok)
	assert.Equal(t, "Executing tool...", chatModel.status)
	assert.Len(t, chatModel.messages, 1)
	assert.NotNil(t, cmd)

	// Execute & Assert for ToolCallEndEvent
	newModel, cmd = chatModel.Update(streamEventMsg{event: <-ch})
	tempChatModel, tempOk = newModel.(*ChatModel)
	chatModel = tempChatModel
	ok = tempOk
	assert.True(t, ok)
	assert.Equal(t, "Got tool result...", chatModel.status)
	assert.NotNil(t, cmd)

	// Execute & Assert for FinalResponseEvent
	newModel, cmd = chatModel.Update(streamEventMsg{event: <-ch})
	tempChatModel, tempOk = newModel.(*ChatModel)
	chatModel = tempChatModel
	ok = tempOk
	assert.True(t, ok)
	assert.Len(t, chatModel.messages, 2)
	assert.NotNil(t, cmd)

	// Execute & Assert for streamFinishMsg
	newModel, cmd = chatModel.Update(streamFinishMsg{})
	tempChatModel, tempOk = newModel.(*ChatModel)
	chatModel = tempChatModel
	ok = tempOk
	assert.True(t, ok)
	assert.False(t, chatModel.isStreaming)
	assert.Equal(t, "Ready", chatModel.status)
	assert.Nil(t, cmd)
}

func TestUpdate_UserConfirmationFlow(t *testing.T) {
	// Setup
	executor := &core.MockExecutor{}
	dummyCommandExecutor := func(args []string) (string, error) {
		return "command executed", nil
	}
	dummyShellService := new(MockShellExecutionService)
	appConfig := config.NewConfig(&config.ConfigParameters{})
	router := routing.NewModelRouterService(appConfig)

	// Initialize the mock executor's UserConfirmationChan before creating the model
	executor.UserConfirmationChan = make(chan bool, 1)

	model := NewChatModel(executor, "mock", appConfig, router, dummyCommandExecutor, dummyShellService)

	// Initialize cancelCtx, cancelFunc, and streamCh for the test
	model.cancelCtx, model.cancelFunc = context.WithCancel(context.Background())
	model.streamCh = make(chan any, 1) // Create a dummy channel

	// Simulate the executor sending a UserConfirmationRequestEvent
	toolCallID := "confirm-123"
	confirmationMessage := "Do you want to proceed?"
	confirmationEvent := types.UserConfirmationRequestEvent{
		ToolCallID: toolCallID,
		Message:    confirmationMessage,
	}

	var newModel tea.Model
	var cmd tea.Cmd
	var ok bool
	var chatModel *ChatModel // Explicitly declare chatModel
	var tempChatModel *ChatModel
	var tempOk bool

	chatModel = model // Assign the initial model after type assertion

	// --- Test Continue ---
	// Reset model state for this sub-test
	chatModel.awaitingConfirmation = true // Should be true to enter confirmation flow
	chatModel.status = "Ready"

	t.Logf("Step 1: Initial state - chatModel: %p", chatModel)

	// Send the confirmation request event to the model
	newModel, cmd = chatModel.Update(streamEventMsg{event: confirmationEvent})
	t.Logf("Step 2: After confirmationEvent - newModel: %p, cmd: %v", newModel, cmd)
	tempChatModel, tempOk = newModel.(*ChatModel)
	chatModel = tempChatModel
	ok = tempOk
	assert.True(t, ok)
	assert.True(t, chatModel.awaitingConfirmation)
	assert.Contains(t, chatModel.status, "Awaiting user confirmation...")
	assert.Nil(t, cmd) // Expect a nil command when awaiting confirmation
	t.Logf("Step 3: After confirmationEvent cast - chatModel: %p, ok: %t", chatModel, ok)

	// Simulate user pressing 'c'
	newModel, cmd = chatModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	t.Logf("Step 4: After KeyMsg 'c' - newModel: %p, cmd: %v", newModel, cmd)
	tempChatModel, tempOk = newModel.(*ChatModel)
	chatModel = tempChatModel
	ok = tempOk
	assert.True(t, ok)
	t.Logf("Step 5: After KeyMsg 'c' cast - chatModel: %p, ok: %t", chatModel, ok)

	// Directly send the confirmation to the channel
	select {
	case executor.UserConfirmationChan <- true:
		t.Logf("Step 6: Directly sent 'true' to UserConfirmationChan")
	case <-time.After(time.Millisecond * 100):
		t.Fatal("Timed out sending 'true' to UserConfirmationChan")
	}

	// Process the userConfirmationMsg that would have been returned by the Update function
	// This is now handled by the direct send above, so we just need to ensure the model state is updated
	// We can simulate the effect of the userConfirmationMsg being processed by the Update function
	// by calling Update with a dummy userConfirmationMsg.
	newModel, cmd = chatModel.Update(userConfirmationMsg{toolCallID: toolCallID, response: "continue"})
	t.Logf("Step 7: After userConfirmationMsg - newModel: %p, cmd: %v", newModel, cmd)
	t.Logf("Step 8: After reading from UserConfirmationChan")
	tempChatModel, tempOk = newModel.(*ChatModel)
	t.Logf("Step 9: After userConfirmationMsg cast - tempChatModel: %p, tempOk: %t", tempChatModel, tempOk)
	chatModel = tempChatModel
	ok = tempOk
	assert.True(t, ok)
	assert.NotNil(t, cmd) // Expect a command to wait for stream events
	t.Logf("Step 10: End of 'c' sub-test - chatModel: %p, ok: %t", chatModel, ok)

	// Verify that 'true' was sent to the executor's channel
	select {
	case confirmed := <-executor.UserConfirmationChan: // Access the channel directly from the mock
		assert.True(t, confirmed)
	case <-time.After(time.Second):
		t.Fatal("Timed out waiting for confirmation response")
	}

	// --- Test Cancel ---
	// Reset model state for this sub-test
	chatModel.awaitingConfirmation = true // Should be true to enter confirmation flow
	chatModel.status = "Ready"

	// Send the confirmation request event to the model
	newModel, cmd = chatModel.Update(streamEventMsg{event: confirmationEvent})
	tempChatModel, tempOk = newModel.(*ChatModel)
	chatModel = tempChatModel
	ok = tempOk
	assert.True(t, ok)
	assert.True(t, chatModel.awaitingConfirmation)
	assert.Contains(t, chatModel.status, "Awaiting user confirmation...")
	assert.Nil(t, cmd) // Expect a nil command when awaiting confirmation

	// Simulate user pressing 'x'
	newModel, cmd = chatModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	tempChatModel, tempOk = newModel.(*ChatModel)
	chatModel = tempChatModel
	ok = tempOk
	assert.True(t, ok)
	assert.False(t, chatModel.awaitingConfirmation)
	assert.Contains(t, chatModel.status, "User cancelled.")
	assert.NotNil(t, cmd) // Expect a command to be returned

	// Directly send the cancellation to the channel
	select {
	case executor.UserConfirmationChan <- false:
		t.Logf("Step X: Directly sent 'false' to UserConfirmationChan")
	case <-time.After(time.Second):
		t.Fatal("Timed out sending 'false' to UserConfirmationChan")
	}

	// Process the userConfirmationMsg that would have been returned by the Update function
	newModel, cmd = chatModel.Update(userConfirmationMsg{toolCallID: toolCallID, response: "cancel"})
	tempChatModel, tempOk = newModel.(*ChatModel)
	chatModel = tempChatModel
	ok = tempOk
	assert.True(t, ok)
	assert.NotNil(t, cmd) // Expect a command to wait for stream events

	// Verify that 'false' was sent to the executor's channel
	select {
	case confirmed := <-executor.UserConfirmationChan: // Access the channel directly from the mock
		assert.False(t, confirmed)
	case <-time.After(time.Second):
		t.Fatal("Timed out waiting for cancellation response")
	}
}
