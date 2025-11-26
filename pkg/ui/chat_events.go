package ui

import (
	"go-ai-agent-v2/go-cli/pkg/services"
	"go-ai-agent-v2/go-cli/pkg/types"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// --- Bubble Tea Message types ---
type tickMsg time.Time
type streamChannelMsg struct{ ch <-chan any }
type streamEventMsg struct{ event any }
type streamErrorMsg struct{ err error }
type streamFinishMsg struct{}

// commandFinishedMsg is sent when a slash command has finished executing.
type commandFinishedMsg struct {
	output string
	err    error
	args   []string // Keep track of the command that was run
}

// chatServiceReloadedMsg is sent when the chat service is reinitialized after a settings change.
type chatServiceReloadedMsg struct {
	newService      *services.ChatService
	newExecutorType string
}

// compressionResultMsg is sent when the chat history compression is complete.
type compressionResultMsg struct {
	result *types.ChatCompressionResult
	err    error
}

// executeCommandCmd runs the commandExecutor in a goroutine and returns a
// commandFinishedMsg when done.
func executeCommandCmd(executor func(args []string) (string, error), args []string) tea.Cmd {
	return func() tea.Msg {
		output, err := executor(args)
		return commandFinishedMsg{output: output, err: err, args: args}
	}
}

// waitForEvent listens on the channel for the next event.
func waitForEvent(ch <-chan any) tea.Cmd {
	return func() tea.Msg {
		event, ok := <-ch
		if !ok {
			return streamFinishMsg{}
		}
		return streamEventMsg{event: event}
	}
}