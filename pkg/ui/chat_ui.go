package ui

import (
	"bufio" // New import
	"context"
	"fmt"
	"os" // New import
	"path/filepath" // New import
	"sort"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/core"
	"go-ai-agent-v2/go-cli/pkg/routing"
	"go-ai-agent-v2/go-cli/pkg/services"
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
		case "smart_edit": // New case for smart_edit
			actionWord = "Smart Editing:"
			if filePath, ok := tc.Args["file_path"].(string); ok {
				argument = filePath
			}
		case "user_confirm": // New case for user_confirm
			actionWord = "Confirmation Required:"
			if message, ok := tc.Args["message"].(string); ok {
				argument = message
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

		// --- Special Content (for write_file, smart_edit, and user_confirm) ---
		if (tc.ToolName == "write_file" || tc.ToolName == "smart_edit") && tc.Args != nil {
			var contentToDisplay string
			var filePathForLexer string

			if tc.ToolName == "write_file" {
				contentToDisplay, _ = tc.Args["content"].(string)
				filePathForLexer, _ = tc.Args["file_path"].(string)
			} else if tc.ToolName == "smart_edit" {
				// For smart_edit, display the new_string and the instruction
				instruction, _ := tc.Args["instruction"].(string)
				newString, _ := tc.Args["new_string"].(string)
				filePathForLexer, _ = tc.Args["file_path"].(string)

				contentToDisplay = fmt.Sprintf("Instruction: %s\n\n--- NEW CONTENT ---\n%s", instruction, newString)
			}

			lines := strings.Split(contentToDisplay, "\n")
			if len(lines) > 6 {
				lines = lines[:6]
				lines = append(lines, "...")
			}
			truncatedContent := strings.Join(lines, "\n")

			lexer := lexers.Match(filePathForLexer)
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
		} else if tc.ToolName == "user_confirm" && tc.Args != nil {
			if message, ok := tc.Args["message"].(string); ok {
				boxContent.WriteString("\n\n")
				boxContent.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render(message))
				boxContent.WriteString("\n")
				boxContent.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render(" (c: continue, x: cancel)"))
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

type userConfirmationMsg struct {
	toolCallID string
	response   string // "continue" or "cancel"
}

func userConfirmationCmd(toolCallID, response string) tea.Cmd {
	return func() tea.Msg {
		return userConfirmationMsg{toolCallID: toolCallID, response: response}
	}
}

type ChatModel struct {
	viewport viewport.Model

	textarea textarea.Model

	spinner spinner.Model

	messages []Message

	executor core.Executor

	executorType string

	config types.Config

		router     *routing.ModelRouterService

		shellService *services.ShellExecutionService // New field

	

		streamCh    <-chan any

		isStreaming bool

		cancelCtx   context.Context    // New field

		cancelFunc  context.CancelFunc // New field

	status string

	err error

	title string

	commandExecutor func(args []string) (string, error) // New field for executing commands

	activeToolCalls map[string]*ToolCallStatus

	// Styles

	senderStyle lipgloss.Style

	botStyle lipgloss.Style

	toolStyle lipgloss.Style

	errorStyle lipgloss.Style

	statusStyle lipgloss.Style

	titleStyle lipgloss.Style

	suggestionStyle lipgloss.Style

	logFile *os.File
	logWriter *bufio.Writer

	awaitingConfirmation bool // New field to indicate if user confirmation is pending
	userConfirmationResponseChan chan bool // Channel to send user confirmation back to executor

	commandHistory []string // Stores previous commands
	historyIndex   int      // Current position in command history
}

func NewChatModel(executor core.Executor, executorType string, config types.Config, router *routing.ModelRouterService, commandExecutor func(args []string) (string, error), shellService *services.ShellExecutionService) *ChatModel {
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

	// Determine the log file path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		telemetry.LogErrorf("Error getting user home directory: %v", err)
		// Handle error, perhaps return nil or a model with a disabled logger
	}
	logDirPath := filepath.Join(homeDir, ".goaiagent", "tmp")
	logFilePath := filepath.Join(logDirPath, "chat-ui.log")

	// Ensure the log directory exists
	if err := os.MkdirAll(logDirPath, 0755); err != nil {
		telemetry.LogErrorf("Error creating log directory: %v", err)
		// Handle error
	}

	// Open log file
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		telemetry.LogErrorf("Error opening chat log file: %v", err)
		// Handle error, perhaps return nil or a model with a disabled logger
	}
	logWriter := bufio.NewWriter(logFile)

	model := &ChatModel{
		viewport:        vp,
		textarea:        ta,
		spinner:         s,
		messages:        []Message{},
		executor:        executor,
		executorType:    executorType,
		config:          config,
		router:          router,
		shellService:    shellService, // Initialize new field
		isStreaming:     false,
		status:          "Ready",
		title:           "Go AI Agent Chat",
		commandExecutor: commandExecutor, // Initialize the new field
		activeToolCalls: make(map[string]*ToolCallStatus),
		senderStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("5")),              // User
		botStyle:        lipgloss.NewStyle().Foreground(lipgloss.Color("6")),              // Bot
		toolStyle:       lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Italic(true), // Tool
		errorStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true),
		statusStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		titleStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("62")).Bold(true),
		suggestionStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Italic(true),
		logFile:         logFile,
		logWriter:       logWriter,
		userConfirmationResponseChan: make(chan bool, 1), // Initialize the channel
		commandHistory:  []string{},
		historyIndex:    0,
	}

	// Pass the confirmation channel to the executor
	model.executor.SetUserConfirmationChannel(model.userConfirmationResponseChan)

	return model
}

// Close closes the log file.
func (m *ChatModel) Close() error {
	if m.logWriter != nil {
		m.logWriter.Flush()
	}
	if m.logFile != nil {
		return m.logFile.Close()
	}
	return nil
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
		if m.awaitingConfirmation {
			switch msg.String() {
							case "c", "C": // Continue
								// Find the active user_confirm tool call
								var toolCallID string
								for id, tc := range m.activeToolCalls {
									if tc.ToolName == "user_confirm" && tc.Status == "Executing" {
										toolCallID = id
										break
									}
								}
								if toolCallID != "" {
									m.awaitingConfirmation = false
									m.status = "User confirmed. Resuming..."
									// Directly send the confirmation message
									return m, userConfirmationCmd(toolCallID, "continue")
								}
							case "x", "X": // Cancel
								// Find the active user_confirm tool call
								var toolCallID string
								for id, tc := range m.activeToolCalls {
									if tc.ToolName == "user_confirm" && tc.Status == "Executing" {
										toolCallID = id
										break
									}
								}
								if toolCallID != "" {
									m.awaitingConfirmation = false
									m.status = "User cancelled."
									// Directly send the cancellation message
									return m, userConfirmationCmd(toolCallID, "cancel")
								}			}
			return m, nil // Consume key presses while awaiting confirmation
		}

		switch msg.Type {
		case tea.KeyEsc: // Handle ESC for cancellation
			if m.isStreaming && m.cancelFunc != nil {
				m.cancelFunc() // Signal cancellation
				m.shellService.KillAllProcesses() // Kill any background shell processes
				m.status = "Cancelling..."
				m.isStreaming = false // Stop streaming immediately in UI
				m.streamCh = nil      // Clear stream channel
				m.activeToolCalls = make(map[string]*ToolCallStatus) // Clear active tool calls
				return m, nil
			}
			// If not streaming or no cancelFunc, then quit
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

			// Add to history if not empty and not a duplicate of the last entry
			if len(m.commandHistory) == 0 || m.commandHistory[len(m.commandHistory)-1] != userInput {
				m.commandHistory = append(m.commandHistory, userInput)
			}
			m.historyIndex = len(m.commandHistory) // Reset history index to the end

			if strings.HasPrefix(userInput, "/") {
				m.messages = append(m.messages, UserMessage{Content: userInput})
				m.updateViewport()
				return m.handleSlashCommand(userInput)
			}

			userMsg := UserMessage{Content: userInput}
			m.messages = append(m.messages, userMsg)
			m.logMessage(userMsg) // Log user message
			m.updateViewport()
			m.isStreaming = true
			m.status = "Sending..."
			telemetry.LogDebugf("Sending user input to executor: %s", userInput)
			return m, m.startStreaming(userInput) // Call updated startStreaming method

		case tea.KeyUp:
			if m.historyIndex > 0 {
				m.historyIndex--
				m.textarea.SetValue(m.commandHistory[m.historyIndex])
			}
		case tea.KeyDown:
			if m.historyIndex < len(m.commandHistory)-1 {
				m.historyIndex++
				m.textarea.SetValue(m.commandHistory[m.historyIndex])
			} else if m.historyIndex == len(m.commandHistory)-1 {
				// If at the last history item and pressing down, clear the input
				m.historyIndex = len(m.commandHistory)
				m.textarea.SetValue("")
			}		}

	// --- Streaming messages ---
	case streamChannelMsg:
		m.streamCh = msg.ch
		return m, waitForEvent(m.streamCh)

	case streamEventMsg:
		switch event := msg.event.(type) {
		case types.StreamingStartedEvent:
			m.status = "Stream started..."
			telemetry.LogDebugf("Received stream event: StreamingStartedEvent")
		case types.ThinkingEvent:
			m.status = "Thinking..."
			telemetry.LogDebugf("Received stream event: ThinkingEvent")
		case types.ToolCallStartEvent:
			m.status = "Executing tool..."
			telemetry.LogDebugf("Received stream event: ToolCallStartEvent (ID: %s, Name: %s, Args: %#v)", event.ToolCallID, event.ToolName, event.Args)
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
				m.logMessage(group) // Log tool call group message
			}
			group.ToolCalls[event.ToolCallID] = tcStatus

		case types.UserConfirmationRequestEvent:
			telemetry.LogDebugf("Received stream event: UserConfirmationRequestEvent (ID: %s, Message: %s)", event.ToolCallID, event.Message)
			// Add a message to the UI indicating confirmation is needed
			m.messages = append(m.messages, SuggestionMessage{Content: fmt.Sprintf("Confirmation required for tool '%s': %s (c: continue, x: cancel)", types.USER_CONFIRM_TOOL_NAME, event.Message)})
			m.logMessage(SuggestionMessage{Content: fmt.Sprintf("Confirmation required for tool '%s': %s", types.USER_CONFIRM_TOOL_NAME, event.Message)})
			m.awaitingConfirmation = true
			m.status = "Awaiting user confirmation..."

			// Add a dummy tool call status for user_confirm to activeToolCalls
			m.activeToolCalls[event.ToolCallID] = &ToolCallStatus{
				ToolName: types.USER_CONFIRM_TOOL_NAME,
				Args:     map[string]interface{}{"message": event.Message},
				Status:   "Executing", // Mark as executing to be found by confirmation logic
			}

			m.updateViewport()
			return m, nil // Stop processing stream events until user confirms

		case types.ToolCallEndEvent:
			m.status = "Got tool result..."
			telemetry.LogDebugf("Received stream event: ToolCallEndEvent (ID: %s, Name: %s, Result: %s, Err: %v)", event.ToolCallID, event.ToolName, event.Result, event.Err)
			if tc, ok := m.activeToolCalls[event.ToolCallID]; ok {
				tc.Status = "Completed"
				tc.Result = event.Result
				tc.Err = event.Err
			}
		case types.FinalResponseEvent:
			botMsg := BotMessage{Content: event.Content}
			m.messages = append(m.messages, botMsg)
			m.logMessage(botMsg) // Log bot message
			telemetry.LogDebugf("Received stream event: FinalResponseEvent (Content: %s)", event.Content)
		case types.ErrorEvent:
			telemetry.LogDebugf("Received stream event: ErrorEvent (Err: %#v)", event.Err)
			errMsg := ErrorMessage{Err: event.Err}
			m.messages = append(m.messages, errMsg)
			m.logMessage(errMsg) // Log error message
			// Check for a model suggestion
			routingCtx := &routing.RoutingContext{
				IsFallback:   true,
				ExecutorType: m.executorType,
			}
			decision, err := m.router.Route(routingCtx, m.config)
			if err == nil && decision != nil {
				suggestion := fmt.Sprintf("Model error. Suggestion: switch to '%s'. Use '/settings set model %s' to switch.", decision.Model, decision.Model)
				m.messages = append(m.messages, SuggestionMessage{Content: suggestion})
				m.logMessage(SuggestionMessage{Content: suggestion}) // Log suggestion message
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
		m.cancelFunc = nil // Clear cancel function
		m.cancelCtx = nil  // Clear context
		return m, nil

	case userConfirmationMsg:
		telemetry.LogDebugf("User confirmation received: ToolCallID: %s, Response: %s", msg.toolCallID, msg.response)
		// Send the user's confirmation back to the executor
		select {
		case m.userConfirmationResponseChan <- (msg.response == "continue"):
			telemetry.LogDebugf("Sent user confirmation to executor: %t", (msg.response == "continue"))
			// Successfully sent confirmation
		case <-m.cancelCtx.Done():
			// Context was cancelled, don't block sending
		default:
			// Channel might be full or closed, log an error or handle appropriately
			telemetry.LogErrorf("Failed to send user confirmation to executor: channel blocked or closed")
		}
		return m, waitForEvent(m.streamCh) // Continue waiting for events
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
	statusLine := m.statusStyle.Render(m.status)
	if m.isStreaming {
		// Add "Press ESC to stop" instruction for executing / streaming processes
		instruction := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(" (Press ESC to stop)")
		return m.spinner.View() + " " + statusLine + instruction
	} else if m.awaitingConfirmation {
		confirmationOptions := lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render(" (c: continue, x: cancel)")
		return statusLine + confirmationOptions
	}
	return statusLine
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
			err := fmt.Errorf("command '%s' is disabled in interactive chat to prevent freezing", args[0])
			m.messages = append(m.messages, ErrorMessage{Err: &types.ToolError{
				Message: err.Error(),
				Type:    types.ToolErrorTypeExecutionFailed,
			}})
			m.updateViewport()
			return m, nil
		}
	}
	// --- End Safety Check ---

	telemetry.LogDebugf("Executing command: %s with args: %v", commandString, args)

	output, err := m.commandExecutor(args)

	telemetry.LogDebugf("Command execution error: %v", err)
	telemetry.LogDebugf("Captured output: %s", output)

	if err != nil {
		errMsg := ErrorMessage{Err: err}
		m.messages = append(m.messages, errMsg)
		m.logMessage(errMsg) // Log error message
	}

	if output != "" {
		botMsg := BotMessage{Content: output}
		m.messages = append(m.messages, botMsg)
		m.logMessage(botMsg) // Log bot message
	} else if err == nil {
		botMsg := BotMessage{Content: "Command executed successfully."}
		m.messages = append(m.messages, botMsg)
		m.logMessage(botMsg) // Log bot message
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

// logMessage writes a message to the chat log file.
func (m *ChatModel) logMessage(msg Message) {
	if m.logWriter == nil {
		return
	}
	// Render the message to a string and write to log
	_, err := m.logWriter.WriteString(msg.Render(m) + "\n")
	if err != nil {
		telemetry.LogErrorf("Error writing to chat log: %v", err)
	}
	m.logWriter.Flush() // Ensure message is written immediately
}

// startStreaming initiates the stream and returns a message with the channel.
func (m *ChatModel) startStreaming(userInput string) tea.Cmd {
	return func() tea.Msg {
		// Create a new context for this streaming session
		m.cancelCtx, m.cancelFunc = context.WithCancel(context.Background())
		stream, err := m.executor.GenerateStream(m.cancelCtx, core.NewUserContent(userInput))
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
