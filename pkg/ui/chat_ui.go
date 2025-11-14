package ui

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/core"
	"go-ai-agent-v2/go-cli/pkg/routing"
	"go-ai-agent-v2/go-cli/pkg/telemetry"
	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra" // Add this line
)

// --- Message Interface and Structs ---

type Message interface {
	Render(m *ChatModel) string
}

type UserMessage struct {
	Content string
}

func (msg UserMessage) Render(m *ChatModel) string {
	return m.senderStyle.Render("You: ") + lipgloss.NewStyle().Width(m.viewport.Width-10).Render(msg.Content)
}

type BotMessage struct {
	Content string
}

func (msg BotMessage) Render(m *ChatModel) string {
	return m.botStyle.Render("Bot: ") + lipgloss.NewStyle().Width(m.viewport.Width-10).Render(msg.Content)
}

type ToolCallStatus struct {
	ToolName string
	Args     map[string]interface{}
	Result   string
	Err      error
	Status   string // "Executing", "Completed"
}

type ToolCallGroupMessage struct {
	ToolCalls map[string]*ToolCallStatus
}

func (msg *ToolCallGroupMessage) Render(m *ChatModel) string {
	var builder strings.Builder

	// Sort the keys for a stable order
	ids := make([]string, 0, len(msg.ToolCalls))
	for id := range msg.ToolCalls {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for i, id := range ids {
		tc := msg.ToolCalls[id]
		var boxContent strings.Builder

		// --- Legend Line ---
		iconColor := lipgloss.Color("208") // Default/Executing color (Orange)
		var statusIcon string
		switch tc.Status {
		case "Executing":
			statusIcon = "⏳"
		case "Completed":
			if tc.Err != nil {
				statusIcon = "❌"
				iconColor = lipgloss.Color("9") // Red
			} else {
				statusIcon = "✅"
				iconColor = lipgloss.Color("10") // Green
			}
		default:
			statusIcon = "•"
		}

		actionWordStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("6")) // Blue
		argumentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15"))  // White

		var actionWord, argument string
		switch tc.ToolName {
		case "execute_command":
			actionWord = "Running:"
			if cmd, ok := tc.Args["command"].(string); ok {
				argument = cmd
			}
		case "write_file":
			actionWord = "Writing to:"
			if path, ok := tc.Args["file_path"].(string); ok {
				argument = path
			}
		case "read_file":
			actionWord = "Reading:"
			if path, ok := tc.Args["file_path"].(string); ok {
				argument = path
			}
		case "web_search":
			actionWord = "Searching:"
			if query, ok := tc.Args["query"].(string); ok {
				argument = query
			}
		default:
			actionWord = tc.ToolName + ":"
			argument = fmt.Sprintf("%v", tc.Args)
			if len(argument) > 60 {
				argument = argument[:57] + "..."
			}
		}

		legend := fmt.Sprintf("%s %s %s",
			lipgloss.NewStyle().Foreground(iconColor).Render(statusIcon),
			actionWordStyle.Render(actionWord),
			argumentStyle.Render(argument),
		)
		boxContent.WriteString(legend)

		// --- Special Content (for write_file) ---
		if tc.ToolName == "write_file" && tc.Args != nil {
			content, _ := tc.Args["content"].(string)
			filePath, _ := tc.Args["file_path"].(string)

			lines := strings.Split(content, "\n")
			if len(lines) > 6 {
				lines = lines[:6]
				lines = append(lines, "...")
			}
			truncatedContent := strings.Join(lines, "\n")

			lexer := lexers.Match(filePath)
			if lexer == nil {
				lexer = lexers.Analyse(truncatedContent)
			}
			if lexer == nil {
				lexer = lexers.Fallback
			}
			style := styles.Get("monokai")
			if style == nil {
				style = styles.Fallback
			}
			formatter := formatters.Get("terminal256")
			iterator, err := lexer.Tokenise(nil, truncatedContent)
			if err == nil {
				var highlightedContent strings.Builder
				if formatter.Format(&highlightedContent, style, iterator) == nil {
					boxContent.WriteString("\n\n") // Add a newline to separate legend from code
					boxContent.WriteString(highlightedContent.String())
				}
			}
		}

		// --- Box Rendering ---
		contentWidth := m.viewport.Width - 6
		box := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1).
			Width(contentWidth).
			Render(boxContent.String())

		builder.WriteString(box)
		if i < len(ids)-1 {
			builder.WriteString("\n")
		}
	}
	return builder.String()
}

type ErrorMessage struct {
	Err error
}

func (msg ErrorMessage) Render(m *ChatModel) string {
	return m.errorStyle.Width(m.viewport.Width - 10).Render(fmt.Sprintf("Error: %v", msg.Err))
}

type SuggestionMessage struct {
	Content string
}

func (msg SuggestionMessage) Render(m *ChatModel) string {
	return m.suggestionStyle.Width(m.viewport.Width - 10).Render(msg.Content)
}

// --- Bubble Tea Message types ---
type streamChannelMsg struct{ ch <-chan any }
type streamEventMsg struct{ event any }
type streamErrorMsg struct{ err error }
type streamFinishMsg struct{}

type ChatModel struct {
	viewport viewport.Model

	textarea textarea.Model

	spinner spinner.Model

	messages []Message

	executor core.Executor

	executorType string

	config types.Config

	router *routing.ModelRouterService

	streamCh <-chan any

	isStreaming bool

	status string

	err error

	title string

	rootCmd *cobra.Command // Add this line

	activeToolCalls map[string]*ToolCallStatus

	// Styles

	senderStyle lipgloss.Style

	botStyle lipgloss.Style

	toolStyle lipgloss.Style

	errorStyle lipgloss.Style

	statusStyle lipgloss.Style

	titleStyle lipgloss.Style

	suggestionStyle lipgloss.Style
}

func NewChatModel(executor core.Executor, executorType string, config types.Config, router *routing.ModelRouterService, rootCmd *cobra.Command) *ChatModel {
	ta := textarea.New()
	ta.Placeholder = "Send a message or type a command (e.g. /clear)..."
	ta.Focus()
	ta.Prompt = "┃ "
	ta.CharLimit = 0 // No limit
	ta.SetWidth(80)
	ta.SetHeight(3)
	ta.ShowLineNumbers = false
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	vp := viewport.New(80, 20)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62"))

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return &ChatModel{
		viewport:        vp,
		textarea:        ta,
		spinner:         s,
		messages:        []Message{},
		executor:        executor,
		executorType:    executorType,
		config:          config,
		router:          router,
		isStreaming:     false,
		status:          "Ready",
		title:           "Go AI Agent Chat",
		rootCmd:         rootCmd, // Initialize the new field
		activeToolCalls: make(map[string]*ToolCallStatus),
		senderStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("5")),              // User
		botStyle:        lipgloss.NewStyle().Foreground(lipgloss.Color("6")),              // Bot
		toolStyle:       lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Italic(true), // Tool
		errorStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true),
		statusStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		titleStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("62")).Bold(true),
		suggestionStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Italic(true),
	}
}
func (m *ChatModel) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, m.spinner.Tick)
}

func (m *ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	if m.isStreaming {
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 8 // Adjust for title, status, and input
		m.textarea.SetWidth(msg.Width)
		m.viewport.Style.Width(msg.Width)
		m.titleStyle.Width(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			if m.isStreaming {
				return m, nil
			}
			userInput := strings.TrimSpace(m.textarea.Value())
			m.textarea.Reset()

			if userInput == "" {
				return m, nil
			}

			if strings.HasPrefix(userInput, "/") {
				m.messages = append(m.messages, UserMessage{Content: userInput})
				m.updateViewport()
				return m.handleSlashCommand(userInput)
			}

			m.messages = append(m.messages, UserMessage{Content: userInput})
			m.updateViewport()
			m.isStreaming = true
			m.status = "Sending..."
			telemetry.LogDebugf("Sending user input to executor: %s", userInput)
			return m, startStreaming(m.executor, userInput)
		}

	// --- Streaming messages ---
	case streamChannelMsg:
		m.streamCh = msg.ch
		return m, waitForEvent(m.streamCh)

	case streamEventMsg:
		telemetry.LogDebugf("Received stream event: %T", msg.event)
		switch event := msg.event.(type) {
		case types.StreamingStartedEvent:
			m.status = "Stream started..."
		case types.ThinkingEvent:
			m.status = "Thinking..."
		case types.ToolCallStartEvent:
			m.status = "Executing tool..."
			tcStatus := &ToolCallStatus{
				ToolName: event.ToolName,
				Args:     event.Args,
				Status:   "Executing",
			}
			m.activeToolCalls[event.ToolCallID] = tcStatus

			// Find or create the tool call group message
			var group *ToolCallGroupMessage
			if len(m.messages) > 0 {
				if g, ok := m.messages[len(m.messages)-1].(*ToolCallGroupMessage); ok {
					group = g
				}
			}
			if group == nil {
				group = &ToolCallGroupMessage{ToolCalls: make(map[string]*ToolCallStatus)}
				m.messages = append(m.messages, group)
			}
			group.ToolCalls[event.ToolCallID] = tcStatus

		case types.ToolCallEndEvent:
			m.status = "Got tool result..."
			if tc, ok := m.activeToolCalls[event.ToolCallID]; ok {
				tc.Status = "Completed"
				tc.Result = event.Result
				tc.Err = event.Err
			}
		case types.FinalResponseEvent:
			m.messages = append(m.messages, BotMessage{Content: event.Content})
		case types.ErrorEvent:
			telemetry.LogDebugf("Received error event: %#v", event.Err)
			m.messages = append(m.messages, ErrorMessage{Err: event.Err})
			// Check for a model suggestion
			routingCtx := &routing.RoutingContext{
				IsFallback:   true,
				ExecutorType: m.executorType,
			}
			decision, err := m.router.Route(routingCtx, m.config)
			if err == nil && decision != nil {
				suggestion := fmt.Sprintf("Model error. Suggestion: switch to '%s'. Use '/settings set model %s' to switch.", decision.Model, decision.Model)
				m.messages = append(m.messages, SuggestionMessage{Content: suggestion})
			}
		}
		m.updateViewport()
		return m, waitForEvent(m.streamCh) // Continue waiting for events

	case streamErrorMsg:
		m.err = msg.err
		m.status = "Error"
		m.messages = append(m.messages, ErrorMessage{Err: msg.err})
		m.updateViewport()
		m.isStreaming = false
		m.streamCh = nil
		return m, nil

	case streamFinishMsg:
		m.isStreaming = false
		m.status = "Ready"
		m.streamCh = nil
		m.activeToolCalls = make(map[string]*ToolCallStatus) // Clear active tool calls
		return m, nil
	}

	return m, tea.Batch(cmds...)
}

func (m *ChatModel) View() string {
	title := m.titleStyle.Render(m.title)
	return fmt.Sprintf(
		"%s\n%s\n\n%s\n%s",
		title,
		m.viewport.View(),
		m.renderStatus(),
		m.textarea.View(),
	)
}

func (m *ChatModel) renderStatus() string {
	if m.isStreaming {
		return m.spinner.View() + " " + m.statusStyle.Render(m.status)
	}
	return m.statusStyle.Render(m.status)
}

func (m *ChatModel) updateViewport() {
	var renderedMessages []string
	for _, msg := range m.messages {
		renderedMessages = append(renderedMessages, msg.Render(m))
	}
	m.viewport.SetContent(strings.Join(renderedMessages, "\n"))
	m.viewport.GotoBottom()
}

func (m *ChatModel) handleSlashCommand(input string) (*ChatModel, tea.Cmd) {
	// Remove the leading slash for Cobra
	commandString := strings.TrimPrefix(input, "/")
	args := strings.Fields(commandString)

	// --- Safety Check ---
	if len(args) > 0 {
		switch args[0] {
		case "chat", "exec", "generate":
			m.messages = append(m.messages, ErrorMessage{Err: fmt.Errorf("command '%s' is disabled in interactive chat to prevent freezing", args[0])})
			m.updateViewport()
			return m, nil
		}
	}
	// --- End Safety Check ---

	telemetry.LogDebugf("Executing command: %s with args: %v", commandString, args)

	// Create a buffer to capture stdout and stderr
	var buffer bytes.Buffer
	m.rootCmd.SetOut(&buffer)
	m.rootCmd.SetErr(&buffer)

	// Find the command
	cmd, _, err := m.rootCmd.Find(args)
	if err != nil {
		m.messages = append(m.messages, ErrorMessage{Err: err})
		m.updateViewport()
		return m, nil
	}

	// Execute the Cobra command
	m.rootCmd.SetArgs(args)
	err = cmd.Execute()

	// Restore default output
	m.rootCmd.SetOut(nil)
	m.rootCmd.SetErr(nil)

	telemetry.LogDebugf("Command execution error: %v", err)
	telemetry.LogDebugf("Captured output: %s", buffer.String())

	if err != nil {
		// The error is already in the buffer, but we can add a more specific message if we want.
		// For now, we'll just display the buffer's content.
	}

	output := strings.TrimSpace(buffer.String())
	if output != "" {
		m.messages = append(m.messages, BotMessage{Content: output})
	} else if err == nil {
		m.messages = append(m.messages, BotMessage{Content: "Command executed successfully."})
	}

	// Handle specific commands that might require UI changes
	if len(args) > 0 {
		switch args[0] {
		case "clear":
			m.messages = []Message{}
		case "quit", "exit":
			return m, tea.Quit
		case "settings":
			if len(args) > 1 && (args[1] == "set" || args[1] == "reset") {
				m.messages = append(m.messages, SuggestionMessage{Content: "Configuration changed. Please restart the chat for the changes to take effect."})
			}
		case "generate", "code-guide", "find-docs":
			// The output is already captured and displayed
		}
	}

	m.textarea.Reset()
	m.updateViewport()
	return m, nil
}

// --- Commands ---

// startStreaming initiates the stream and returns a message with the channel.
func startStreaming(executor core.Executor, userInput string) tea.Cmd {
	return func() tea.Msg {
		stream, err := executor.GenerateStream(core.NewUserContent(userInput))
		if err != nil {
			return streamErrorMsg{err}
		}
		return streamChannelMsg{ch: stream}
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
