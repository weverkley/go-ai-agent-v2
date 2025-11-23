package ui

import (
	"bufio" // New import
	"context"
	"fmt"
	"go-ai-agent-v2/go-cli/pkg/pathutils"
	"go-ai-agent-v2/go-cli/pkg/services"
	"go-ai-agent-v2/go-cli/pkg/telemetry"
	"go-ai-agent-v2/go-cli/pkg/types"
	"go-ai-agent-v2/go-cli/pkg/utils"
	"os"            // New import
	"path/filepath" // New import
	"sort"
	"strings"
	"time"

	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mbndr/figlet4go"
)

const tipsMessage = `Tips for getting started:
1. Ask questions, edit files, or run commands.
2. Be specific for the best results.
3. /help for more information.`

// --- Message Interface and Structs ---
type Message interface {
	Render(m *ChatModel) string
}

func createInitialMessages() []Message {
	ascii := figlet4go.NewAsciiRender()
	// You can change the font and color if you wish
	// options := figlet4go.NewRenderOptions()
	// options.FontColor = []figlet4go.Color{
	// 	// Colors here
	// }
	// ascii.LoadFont(options.FontName)
	renderStr, _ := ascii.Render("GO AI AGENT")
	return []Message{
		SystemMessage{Content: renderStr},
		SuggestionMessage{Content: tipsMessage},
	}
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

type SystemMessage struct {
	Content string
}

func (msg SystemMessage) Render(m *ChatModel) string {
	// A simple message type that just renders the content, useful for logos or announcements.
	return lipgloss.NewStyle().Width(m.viewport.Width).Render(msg.Content)
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

// executeCommandCmd runs the commandExecutor in a goroutine and returns a
// commandFinishedMsg when done.
func executeCommandCmd(executor func(args []string) (string, error), args []string) tea.Cmd {
	return func() tea.Msg {
		output, err := executor(args)
		return commandFinishedMsg{output: output, err: err, args: args}
	}
}

type ChatModel struct {
	viewport     viewport.Model
	textarea     textarea.Model
	spinner      spinner.Model
	messages     []Message
	executorType string
	config       types.Config
	// Services
	chatService      *services.ChatService
	sessionService   *services.SessionService // New field
	shellService     services.ShellExecutionService
	gitService       services.GitService
	workspaceService *services.WorkspaceService
	// UI state
	streamCh        <-chan any
	isStreaming     bool
	cancelCtx       context.Context
	cancelFunc      context.CancelFunc
	status          string
	err             error
	commandExecutor func(args []string) (string, error)
	activeToolCalls map[string]*ToolCallStatus
	// Styles
	senderStyle              lipgloss.Style
	botStyle                 lipgloss.Style
	toolStyle                lipgloss.Style
	errorStyle               lipgloss.Style
	statusStyle              lipgloss.Style
	footerStyle              lipgloss.Style
	suggestionStyle          lipgloss.Style
	pathStyle                lipgloss.Style
	branchStyle              lipgloss.Style
	modelStyle               lipgloss.Style
	logFile                  *os.File
	logWriter                *bufio.Writer
	awaitingConfirmation     bool                                // New field to indicate if user confirmation is pending
	toolConfirmationRequest  *types.ToolConfirmationRequestEvent // New field for rich tool confirmation
	awaitingToolConfirmation bool                                // New field
	commandHistory           []string                            // Stores previous commands
	historyIndex             int                                 // Current position in command history
	// Session stats
	startTime      time.Time
	toolCallCount  int
	toolErrorCount int
	contextFile    string
	sessionID      string // New field to store the current session ID
}

func NewChatModel(
	chatService *services.ChatService,
	sessionService *services.SessionService, // New parameter
	executorType string,
	config types.Config,
	commandExecutor func(args []string) (string, error),
	shellService services.ShellExecutionService,
	gitService services.GitService,
	workspaceService *services.WorkspaceService,
	sessionID string,
) *ChatModel {
	ta := textarea.New()
	ta.Placeholder = "Send a message or type a command (e.g. /clear)..."
	ta.Focus()
	ta.Prompt = "❯ "
	ta.CharLimit = 0 // No limit
	ta.SetWidth(80)
	ta.SetHeight(1)
	ta.ShowLineNumbers = false
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	vp := viewport.New(80, 20)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.HiddenBorder())
	s := spinner.New()
	s.Spinner = spinner.Meter
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	// Determine the log file path from telemetry settings
	var logDirPath string
	telemetrySettingsVal, ok := config.Get("telemetry")
	if ok {
		if telemetrySettings, ok := telemetrySettingsVal.(*types.TelemetrySettings); ok && telemetrySettings.Enabled && telemetrySettings.OutDir != "" {
			expandedPath, err := pathutils.ExpandPath(telemetrySettings.OutDir)
			if err != nil {
				telemetry.LogErrorf("Error expanding telemetry OutDir path '%s': %v", telemetrySettings.OutDir, err)
			} else {
				logDirPath = expandedPath
			}
		}
	}
	// Fallback to default if not specified in telemetry or if expansion failed
	if logDirPath == "" {
		expandedPath, err := pathutils.ExpandPath("~/.goaiagent/tmp")
		if err != nil {
			telemetry.LogErrorf("Error expanding default log directory path: %v", err)
			// As a last resort, use a relative path
			logDirPath = ".goaiagent/tmp"
		} else {
			logDirPath = expandedPath
		}
	}
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
		viewport:         vp,
		textarea:         ta,
		spinner:          s,
		messages:         createInitialMessages(),
		chatService:      chatService,
		sessionService:   sessionService, // Initialize new field
		executorType:     executorType,
		config:           config,
		shellService:     shellService,
		gitService:       gitService,
		workspaceService: workspaceService,
		isStreaming:      false,
		status:           "Ready",
		commandExecutor:  commandExecutor,
		activeToolCalls:  make(map[string]*ToolCallStatus),
		senderStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		botStyle:         lipgloss.NewStyle().Foreground(lipgloss.Color("6")),
		toolStyle:        lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Italic(true),
		errorStyle:       lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true),
		statusStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		footerStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		suggestionStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Italic(true),
		pathStyle:        lipgloss.NewStyle().Foreground(lipgloss.Color("12")), // Blue
		branchStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("10")), // Green
		modelStyle:       lipgloss.NewStyle().Foreground(lipgloss.Color("13")), // Pink
		logFile:          logFile,
		logWriter:        logWriter,
		commandHistory:   []string{},
		historyIndex:     0,
		startTime:        time.Now(),
		toolCallCount:    0,
		toolErrorCount:   0,
		sessionID:        sessionID,
	}
	// Check for context file
	if _, err := os.Stat("GOAIAGENT.md"); err == nil {
		model.contextFile = "GOAIAGENT.md"
	}
	model.updateViewport() // Ensure initial messages are displayed
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

// GetStats returns the final session statistics.
func (m *ChatModel) GetStats() (int, int, time.Duration) {
	return m.toolCallCount, m.toolErrorCount, time.Since(m.startTime)
}
func (m *ChatModel) Init() tea.Cmd {
	return tea.Batch(
		textarea.Blink,
		m.spinner.Tick,
		tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return tickMsg(t)
		}),
	)
}
func (m *ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	// Handle rich tool confirmation prompt first, blocking all other input.
	if keyMsg, ok := msg.(tea.KeyMsg); ok && m.awaitingToolConfirmation {
		switch keyMsg.String() {
		case "y", "Y": // Yes, allow once
			m.awaitingToolConfirmation = false
			m.status = fmt.Sprintf("User confirmed '%s' (once). Resuming...", m.toolConfirmationRequest.ToolName)
			m.chatService.GetToolConfirmationChannel() <- types.ToolConfirmationOutcomeProceedOnce
			return m, waitForEvent(m.streamCh)
		case "n", "N", "esc": // No, cancel
			m.awaitingToolConfirmation = false
			m.status = fmt.Sprintf("User cancelled '%s'. Resuming...", m.toolConfirmationRequest.ToolName)
			m.chatService.GetToolConfirmationChannel() <- types.ToolConfirmationOutcomeCancel
			return m, waitForEvent(m.streamCh)
		case "m", "M": // Modify with external editor
			m.awaitingToolConfirmation = false
			m.status = fmt.Sprintf("User chose to modify '%s' with editor. Resuming...", m.toolConfirmationRequest.ToolName)
			m.chatService.GetToolConfirmationChannel() <- types.ToolConfirmationOutcomeModifyWithEditor
			return m, waitForEvent(m.streamCh)
		default:
			// For any other key, just ignore it.
			return m, nil
		}
	}
	// Handle simple user_confirm prompt
	if keyMsg, ok := msg.(tea.KeyMsg); ok && m.awaitingConfirmation {
		switch keyMsg.String() {
		case "c", "C":
			m.awaitingConfirmation = false
			m.status = "User confirmed. Resuming..."
			m.chatService.GetUserConfirmationChannel() <- true
			return m, waitForEvent(m.streamCh)
		case "x", "X":
			m.awaitingConfirmation = false
			m.status = "User cancelled. Resuming..."
			m.chatService.GetUserConfirmationChannel() <- false
			return m, waitForEvent(m.streamCh)
		default:
			// For any other key, just ignore it and don't pass it to the text area.
			return m, nil
		}
	}
	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var spinnerCmd tea.Cmd
		m.spinner, spinnerCmd = m.spinner.Update(msg)
		return m, spinnerCmd
	case tickMsg:
		// This is just to trigger a re-render for the timer.
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 3 // Adjust for footer and input
		m.textarea.SetWidth(msg.Width)
		m.viewport.Style.Width(msg.Width)
		return m, nil
	case tea.KeyMsg:
		// NOTE: Confirmation logic is now handled at the top of the Update function.
		switch msg.Type {
		case tea.KeyEsc: // Handle ESC for cancellation
			if m.isStreaming && m.cancelFunc != nil {
				m.cancelFunc()                    // Signal cancellation
				m.shellService.KillAllProcesses() // Kill any background shell processes
				m.status = "Cancelling..."
				m.isStreaming = false                                // Stop streaming immediately in UI
				m.streamCh = nil                                     // Clear stream channel
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
				userMsg := UserMessage{Content: userInput}
				m.messages = append(m.messages, userMsg)
				m.logUIMessage(userMsg)
				m.updateViewport()
				return m.handleSlashCommand(userInput)
			}
			userMsg := UserMessage{Content: userInput}
			m.messages = append(m.messages, userMsg)
			m.logUIMessage(userMsg) // Log user message
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
			}
		}
	// --- Streaming messages ---
	case streamChannelMsg:
		m.streamCh = msg.ch
		return m, waitForEvent(m.streamCh)
	case streamEventMsg:
		switch event := msg.event.(type) {
		case types.StreamingStartedEvent:
			m.status = "Stream started..."
		case types.ThinkingEvent:
			m.status = "Thinking..."
		case types.ToolCallStartEvent:
			m.status = "Executing tool..."
			m.toolCallCount++ // Increment tool call count
			m.logSystemMessage(fmt.Sprintf("Tool Started: %s with args %v", event.ToolName, event.Args))
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
		case types.UserConfirmationRequestEvent:
			// Add a message to the UI indicating confirmation is needed
			suggestionMsg := SuggestionMessage{Content: fmt.Sprintf("Confirmation required for tool '%s': %s (c: continue, x: cancel)", types.USER_CONFIRM_TOOL_NAME, event.Message)}
			m.messages = append(m.messages, suggestionMsg)
			m.logUIMessage(suggestionMsg)
			m.awaitingConfirmation = true
			m.status = "Awaiting user confirmation..."
			// Add a dummy tool call status for user_confirm to activeToolCalls
			m.activeToolCalls[event.ToolCallID] = &ToolCallStatus{
				ToolName: types.USER_CONFIRM_TOOL_NAME,
				Args:     map[string]interface{}{"message": event.Message},
				Status:   "Executing", // Mark as executing to be found by confirmation logic
			}
			m.updateViewport()
			return m, nil // Stop waiting for stream events, but allow other ticks to continue
		case types.ToolConfirmationRequestEvent: // NEW: Rich tool confirmation
			telemetry.LogDebugf("Received stream event: ToolConfirmationRequestEvent (ID: %s, Name: %s)", event.ToolCallID, event.ToolName)
			m.toolConfirmationRequest = &event
			m.awaitingToolConfirmation = true
			m.status = "Awaiting tool confirmation..."
			m.updateViewport()
			return m, nil // Stop waiting for stream events, but allow other ticks to continue
		case types.ToolCallEndEvent:
			m.status = "Got tool result..."
			if event.Err != nil {
				m.toolErrorCount++
			}
			m.logSystemMessage(fmt.Sprintf("Tool Ended: %s. Result: %s, Error: %v", event.ToolName, event.Result, event.Err))
			if tc, ok := m.activeToolCalls[event.ToolCallID]; ok {
				tc.Status = "Completed"
				tc.Result = event.Result
				tc.Err = event.Err
			}
		case types.FinalResponseEvent:
			botMsg := BotMessage{Content: event.Content}
			m.messages = append(m.messages, botMsg)
			m.logUIMessage(botMsg) // Log bot message
		case types.ErrorEvent:
			errMsg := ErrorMessage{Err: event.Err}
			m.messages = append(m.messages, errMsg)
			m.logUIMessage(errMsg) // Log error message
		}
		m.updateViewport()
		return m, waitForEvent(m.streamCh) // Continue waiting for events
	case streamErrorMsg:
		m.err = msg.err
		m.status = "Error"
		errMsg := ErrorMessage{Err: msg.err}
		m.messages = append(m.messages, errMsg)
		m.logUIMessage(errMsg)
		m.updateViewport()
		m.isStreaming = false
		m.streamCh = nil
		return m, nil
	case streamFinishMsg:
		m.isStreaming = false
		m.status = "Ready"
		m.streamCh = nil
		m.activeToolCalls = make(map[string]*ToolCallStatus) // Clear active tool calls
		m.cancelFunc = nil                                   // Clear cancel function
		m.cancelCtx = nil                                    // Clear context
		// Save history after each completed turn
		if err := m.sessionService.SaveHistory(m.sessionID, m.chatService.GetHistory()); err != nil {
			telemetry.LogErrorf("Failed to save history after stream finish for session %s: %v", m.sessionID, err)
		}
		return m, nil
	case commandFinishedMsg:
		m.isStreaming = false
		m.status = "Ready"
		telemetry.LogDebugf("Command execution finished. Error: %v, Output: %s", msg.err, msg.output)
		if msg.err != nil {
			errMsg := ErrorMessage{Err: msg.err}
			m.messages = append(m.messages, errMsg)
			m.logUIMessage(errMsg)
		}
		if msg.output != "" {
			botMsg := BotMessage{Content: msg.output}
			m.messages = append(m.messages, botMsg)
			m.logUIMessage(botMsg)
		} else if msg.err == nil {
			botMsg := BotMessage{Content: "Command executed successfully."}
			m.messages = append(m.messages, botMsg)
			m.logUIMessage(botMsg)
		}
		// Handle specific commands that might require UI changes
		if len(msg.args) > 0 {
			switch msg.args[0] {
			case "quit", "exit":
				// Save history before quitting
				if err := m.sessionService.SaveHistory(m.sessionID, m.chatService.GetHistory()); err != nil {
					telemetry.LogErrorf("Failed to save history before quitting for session %s: %v", m.sessionID, err)
				}
				return m, tea.Quit
			case "settings":
				if len(msg.args) > 1 && (msg.args[1] == "set" || msg.args[1] == "reset") {
					suggestionMsg := SuggestionMessage{Content: "Configuration changed. Please restart the chat for the changes to take effect."}
					m.messages = append(m.messages, suggestionMsg)
					m.logUIMessage(suggestionMsg)
				}
			}
		}
		m.updateViewport()
		return m, nil
	}
	return m, tea.Batch(cmds...)
}
func (m *ChatModel) View() string {
	var contextInfo string
	if m.contextFile != "" {
		// Mimicking the "Using: 2 GEMINI.md files" format
		contextInfo = lipgloss.NewStyle().
			PaddingLeft(1). // Add some left padding
			Foreground(lipgloss.Color("240")).
			Render(fmt.Sprintf("Using: 1 %s file", m.contextFile))
	}
	// Place context info on the left side, above the footer
	return fmt.Sprintf(
		"%s\n%s\n%s\n%s",
		m.viewport.View(),
		contextInfo, // Render directly for left alignment
		m.renderFooter(),
		m.textarea.View(),
	)
}
func (m *ChatModel) renderFooter() string {
	// While streaming or awaiting confirmation, show the spinner and status.
	if m.isStreaming || m.awaitingConfirmation || m.awaitingToolConfirmation {
		statusLine := lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Render(m.status) // Always blue in active state
		var finalRender string
		if m.awaitingToolConfirmation && m.toolConfirmationRequest != nil {
			options := "(y: allow once, n: cancel, m: modify)"
			if m.toolConfirmationRequest.Type == "edit" {
				// Also show diff
				diffSummary := ""
				if m.toolConfirmationRequest.FileDiff != "" {
					diffSummary = "\n" + m.toolConfirmationRequest.FileDiff
				}
				finalRender = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render(m.toolConfirmationRequest.Message+diffSummary) + " " + lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render(options)
			} else {
				finalRender = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render(m.toolConfirmationRequest.Message) + " " + lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render(options)
			}
		} else if m.awaitingConfirmation {
			confirmationOptions := lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render(" (c: continue, x: cancel)")
			finalRender = statusLine + confirmationOptions
		} else {
			instruction := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(" (Press ESC to stop)")
			finalRender = statusLine + instruction
		}
		return m.spinner.View() + " " + finalRender
	}
	// When not busy, show the full stats footer.
	const separator = "  "
	// Left side: CWD and Git branch
	cwd, err := os.Getwd() // Use os.Getwd() directly
	if err != nil {
		cwd = "???"
	}
	baseCwd := filepath.Base(cwd) // Get only the current folder name
	branch, err := m.gitService.GetCurrentBranch(cwd)
	if err != nil {
		branch = "" // Don't show branch if not in a git repo or error
	}
	var left string
	if branch != "" {
		left = fmt.Sprintf("%s (%s)", m.pathStyle.Render(baseCwd), m.branchStyle.Render(branch))
	} else {
		left = m.pathStyle.Render(baseCwd)
	}
	// Center: Stats
	duration := time.Since(m.startTime)
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60
	var toolsStats string
	if m.toolErrorCount > 0 {
		errorString := lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(fmt.Sprintf("(%d errors)", m.toolErrorCount)) // Red color
		toolsStats = fmt.Sprintf("Tools: %d %s", m.toolCallCount, errorString)
	} else {
		toolsStats = fmt.Sprintf("Tools: %d", m.toolCallCount)
	}
	stats := fmt.Sprintf("%s | Time: %02d:%02d:%02d", toolsStats, hours, minutes, seconds)
	// Right side: Model name and Session ID
	right := fmt.Sprintf("%s | Session: %s", m.modelStyle.Render(m.executorType), m.modelStyle.Render(m.sessionID))
	// Calculate remaining space
	usedWidth := lipgloss.Width(left) + lipgloss.Width(right) + lipgloss.Width(stats) + 2*lipgloss.Width(separator)
	remainingWidth := m.viewport.Width - usedWidth
	if remainingWidth < 0 {
		remainingWidth = 0
	}
	spring := lipgloss.NewStyle().Width(remainingWidth).Render("")
	return m.footerStyle.Render(lipgloss.JoinHorizontal(lipgloss.Bottom,
		left,
		separator,
		stats,
		spring,
		right,
	))
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
	// --- UI-only commands ---
	if len(args) > 0 {
		switch args[0] {
		case "clear":
			m.chatService.ClearHistory()
			m.messages = createInitialMessages()
			m.updateViewport()
			return m, nil
		case "help":
			helpText := `
Available Commands:
  /clear                  - Clears the current chat session history.
  /sessions list          - Lists all saved chat sessions.
  /sessions resume <id>   - Resumes a specific chat session by its ID or number.
  /quit                   - Exits the application.
  /help                   - Shows this help message.
`
			m.messages = append(m.messages, BotMessage{Content: helpText})
			m.updateViewport()
			return m, nil
		case "sessions":
			if len(args) < 2 {
				m.messages = append(m.messages, ErrorMessage{Err: fmt.Errorf("usage: /sessions [list|resume]")})
				m.updateViewport()
				return m, nil
			}
			switch args[1] {
			case "list":
				sessions, err := m.sessionService.ListSessions()
				if err != nil {
					m.messages = append(m.messages, ErrorMessage{Err: err})
				} else if len(sessions) == 0 {
					m.messages = append(m.messages, BotMessage{Content: "No saved sessions found."})
				} else {
					var sessionList strings.Builder
					sessionList.WriteString("Available sessions (newest first):\n")
					for i, id := range sessions {
						sessionList.WriteString(fmt.Sprintf("  %d: %s\n", i+1, id))
					}
					m.messages = append(m.messages, BotMessage{Content: sessionList.String()})
				}
				m.updateViewport()
				return m, nil
			case "resume":
				if len(args) < 3 {
					m.messages = append(m.messages, ErrorMessage{Err: fmt.Errorf("usage: /sessions resume <session_id>")})
					m.updateViewport()
					return m, nil
				}
				sessionID := args[2]
				// Re-initialize chat service with the new session
				// This assumes the executor can be reused.
				newChatService, err := services.NewChatService(m.chatService.GetExecutor(), m.chatService.GetToolRegistry(), m.sessionService, sessionID, m.chatService.GetSettingsService())
				if err != nil {
					m.messages = append(m.messages, ErrorMessage{Err: fmt.Errorf("failed to resume session: %w", err)})
					m.updateViewport()
					return m, nil
				}
				m.chatService = newChatService
				m.sessionID = sessionID
				m.messages = m.repopulateMessagesFromHistory(newChatService.GetHistory())
				m.status = fmt.Sprintf("Resumed session %s", sessionID)
				m.updateViewport()
				return m, nil
			default:
				m.messages = append(m.messages, ErrorMessage{Err: fmt.Errorf("unknown /sessions command: %s", args[1])})
				m.updateViewport()
				return m, nil
			}
		}
	}
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
	telemetry.LogDebugf("Executing command asynchronously: %s with args: %v", commandString, args)
	m.status = fmt.Sprintf("Executing `/%s`...", commandString)
	m.isStreaming = true // Use the spinner to indicate the command is running
	return m, executeCommandCmd(m.commandExecutor, args)
}

// repopulateMessagesFromHistory converts the raw history from the service to UI messages.
func (m *ChatModel) repopulateMessagesFromHistory(history []*types.Content) []Message {
	newMessages := createInitialMessages()
	for _, content := range history {
		if content.Role == "user" {
			// A "user" role can be a text prompt or a set of tool responses.
			isToolResponse := false
			for _, part := range content.Parts {
				if part.FunctionResponse != nil {
					isToolResponse = true
					break
				}
			}
			// If it's a tool response, we don't render a separate message for it in the UI.
			if isToolResponse {
				continue
			}
			// It's a normal text message from the user.
			var textContent strings.Builder
			for _, part := range content.Parts {
				textContent.WriteString(part.Text)
			}
			if textContent.Len() > 0 {
				newMessages = append(newMessages, UserMessage{Content: textContent.String()})
			}
		} else if content.Role == "model" {
			// A "model" role can be a text response, a set of tool calls, or both.
			var functionCalls []*types.FunctionCall
			var textParts []string
			for _, part := range content.Parts {
				if part.FunctionCall != nil {
					functionCalls = append(functionCalls, part.FunctionCall)
				}
				if part.Text != "" {
					textParts = append(textParts, part.Text)
				}
			}
			if len(functionCalls) > 0 {
				// This was a tool-calling turn.
				group := &ToolCallGroupMessage{ToolCalls: make(map[string]*ToolCallStatus)}
				for i, fc := range functionCalls {
					// We don't have the original result or status here, as the history
					// doesn't store UI state. We'll render it as "Completed" since it's in the past.
					toolCallID := fmt.Sprintf("resumed-tool-%d-%d", len(newMessages), i)
					status := &ToolCallStatus{
						ToolName: fc.Name,
						Args:     fc.Args,
						Status:   "Completed",              // Assume completed as it's from history
						Result:   "(resumed from history)", // Placeholder result
						Err:      nil,
					}
					group.ToolCalls[toolCallID] = status
				}
				newMessages = append(newMessages, group)
			}
			if len(textParts) > 0 {
				// This was a text response turn.
				fullText := strings.Join(textParts, "")
				newMessages = append(newMessages, BotMessage{Content: fullText})
			}
		}
	}
	return newMessages
}

// --- Commands ---
// logUIMessage writes the rendered UI message to the log file.
func (m *ChatModel) logUIMessage(msg Message) {
	if m.logWriter == nil {
		return
	}
	renderedString := msg.Render(m)
	cleanString := utils.StripAnsi(renderedString)
	// Render the message to a string and write to log
	_, err := m.logWriter.WriteString(cleanString + "\n")
	if err != nil {
		telemetry.LogErrorf("Error writing to chat log: %v", err)
	}
	m.logWriter.Flush() // Ensure message is written immediately
}

// logSystemMessage writes a timestamped system message to the log file.
func (m *ChatModel) logSystemMessage(logMsg string) {
	if m.logWriter == nil {
		return
	}
	timestamp := time.Now().Format("15:04:05.000")
	formattedMsg := fmt.Sprintf("[%s] %s\n", timestamp, logMsg)
	_, err := m.logWriter.WriteString(formattedMsg)
	if err != nil {
		telemetry.LogErrorf("Error writing system message to chat log: %v", err)
	}
	m.logWriter.Flush()
}

// startStreaming initiates the stream and returns a message with the channel.
func (m *ChatModel) startStreaming(userInput string) tea.Cmd {
	return func() tea.Msg {
		m.cancelCtx, m.cancelFunc = context.WithCancel(context.Background())
		stream, err := m.chatService.SendMessage(m.cancelCtx, userInput)
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
