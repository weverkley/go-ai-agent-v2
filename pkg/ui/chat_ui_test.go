package ui

import (
	"context" // New import
	"testing"

	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/core"
	"go-ai-agent-v2/go-cli/pkg/routing"
	"go-ai-agent-v2/go-cli/pkg/services" // New import
	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/charmbracelet/bubbletea"
	"github.com/google/generative-ai-go/genai"
	"github.com/stretchr/testify/assert"
)

func TestNewChatModel(t *testing.T) {
	executor := &core.MockExecutor{}
	// Define a dummy commandExecutor for testing
	dummyCommandExecutor := func(args []string) (string, error) {
		return "command executed", nil
	}
	dummyShellService := &services.ShellExecutionService{} // Create a dummy shell service
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
	dummyShellService := &services.ShellExecutionService{}
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
	dummyShellService := &services.ShellExecutionService{}
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
	dummyShellService := &services.ShellExecutionService{}
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
	dummyShellService := &services.ShellExecutionService{}
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
	chatModel := model

	// Execute & Assert for StreamingStartedEvent
	newModel, cmd = chatModel.Update(streamEventMsg{event: <-ch})
	chatModel, ok = newModel.(*ChatModel)
	assert.True(t, ok)
	assert.Equal(t, "Stream started...", chatModel.status)
	assert.NotNil(t, cmd)

	// Execute & Assert for ThinkingEvent
	newModel, cmd = chatModel.Update(streamEventMsg{event: <-ch})
	chatModel, ok = newModel.(*ChatModel)
	assert.True(t, ok)
	assert.Equal(t, "Thinking...", chatModel.status)
	assert.NotNil(t, cmd)

	// Execute & Assert for ToolCallStartEvent
	newModel, cmd = chatModel.Update(streamEventMsg{event: <-ch})
	chatModel, ok = newModel.(*ChatModel)
	assert.True(t, ok)
	assert.Equal(t, "Executing tool...", chatModel.status)
	assert.Len(t, chatModel.messages, 1)
	assert.NotNil(t, cmd)

	// Execute & Assert for ToolCallEndEvent
	newModel, cmd = chatModel.Update(streamEventMsg{event: <-ch})
	chatModel, ok = newModel.(*ChatModel)
	assert.True(t, ok)
	assert.Equal(t, "Got tool result...", chatModel.status)
	assert.NotNil(t, cmd)

	// Execute & Assert for FinalResponseEvent
	newModel, cmd = chatModel.Update(streamEventMsg{event: <-ch})
	chatModel, ok = newModel.(*ChatModel)
	assert.True(t, ok)
	assert.Len(t, chatModel.messages, 2)
	assert.NotNil(t, cmd)

	// Execute & Assert for streamFinishMsg
	newModel, cmd = chatModel.Update(streamFinishMsg{})
	chatModel, ok = newModel.(*ChatModel)
	assert.True(t, ok)
	assert.False(t, chatModel.isStreaming)
	assert.Equal(t, "Ready", chatModel.status)
	assert.Nil(t, cmd)
}
