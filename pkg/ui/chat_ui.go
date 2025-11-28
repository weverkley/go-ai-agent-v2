package ui

import (
	"bufio"
	"context"
	"fmt"

	// New import for core package
	"go-ai-agent-v2/go-cli/pkg/core"
	"go-ai-agent-v2/go-cli/pkg/pathutils"
	"go-ai-agent-v2/go-cli/pkg/services"
	"go-ai-agent-v2/go-cli/pkg/telemetry"
	"go-ai-agent-v2/go-cli/pkg/types"
	"go-ai-agent-v2/go-cli/pkg/utils"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ChatModel struct {
	viewport     viewport.Model
	textarea     textarea.Model
	spinner      spinner.Model
	messages     []Message
	executorType string
	config       types.Config
	// Services
	chatService      *services.ChatService
	sessionService   *services.SessionService
	contextService   *services.ContextService
	shellService     services.ShellExecutionService
	gitService       services.GitService
	workspaceService *services.WorkspaceService
	// UI state
	streamCh        <-chan interface{}
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
	toolConfirmationRequest  *types.ToolConfirmationRequestEvent
	awaitingToolConfirmation bool
	commandHistory           []string
	historyIndex             int
	// Session stats
	startTime      time.Time
	toolCallCount  int
	toolErrorCount int
	contextFile    string
	sessionID      string
	todosSummary   string
	subagentStatus string
}

func NewChatModel(
	chatService *services.ChatService,
	sessionService *services.SessionService,
	contextService *services.ContextService,
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
	ta.CharLimit = 0
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
	// ... (logging initialization)
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
	if logDirPath == "" {
		expandedPath, err := pathutils.ExpandPath("~/.goaiagent/tmp")
		if err != nil {
			telemetry.LogErrorf("Error expanding default log directory path: %v", err)
			logDirPath = ".goaiagent/tmp"
		} else {
			logDirPath = expandedPath
		}
	}
	logFilePath := filepath.Join(logDirPath, "chat-ui.log")
	if err := os.MkdirAll(logDirPath, 0755); err != nil {
		telemetry.LogErrorf("Error creating log directory: %v", err)
	}
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		telemetry.LogErrorf("Error opening chat log file: %v", err)
	}
	logWriter := bufio.NewWriter(logFile)
	model := &ChatModel{
		viewport:                 vp,
		textarea:                 ta,
		spinner:                  s,
		messages:                 createInitialMessages(),
		chatService:              chatService,
		sessionService:           sessionService,
		contextService:           contextService,
		executorType:             executorType,
		config:                   config,
		shellService:             shellService,
		gitService:               gitService,
		workspaceService:         workspaceService,
		isStreaming:              false,
		status:                   "Ready",
		commandExecutor:          commandExecutor,
		activeToolCalls:          make(map[string]*ToolCallStatus),
		senderStyle:              lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		botStyle:                 lipgloss.NewStyle().Foreground(lipgloss.Color("6")),
		toolStyle:                lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Italic(true),
		errorStyle:               lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true),
		statusStyle:              lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		footerStyle:              lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		suggestionStyle:          lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Italic(true),
		pathStyle:                lipgloss.NewStyle().Foreground(lipgloss.Color("12")),
		branchStyle:              lipgloss.NewStyle().Foreground(lipgloss.Color("10")),
		modelStyle:               lipgloss.NewStyle().Foreground(lipgloss.Color("13")),
		logFile:                  logFile,
		logWriter:                logWriter,
		toolConfirmationRequest:  nil,
		awaitingToolConfirmation: false,
		commandHistory:           []string{},
		historyIndex:             0,
		startTime:                time.Now(),
		toolCallCount:            0,
		toolErrorCount:           0,
		contextFile:              "",
		sessionID:                sessionID,
		todosSummary:             "",
		subagentStatus:           "",
	}
	if _, err := os.Stat("GOAIAGENT.md"); err == nil {
		model.contextFile = "GOAIAGENT.md"
	}
	model.updateViewport()
	return model
}
func (m *ChatModel) Close() error {
	if m.logWriter != nil {
		m.logWriter.Flush()
	}
	if m.logFile != nil {
		return m.logFile.Close()
	}
	return nil
}
func (m *ChatModel) GetStats() (int, int, time.Duration, map[string]*types.ModelTokenUsage) {
	return m.toolCallCount, m.toolErrorCount, time.Since(m.startTime), m.chatService.GetTokenUsage()
}

// SetChatService updates the ChatModel's chatService and executorType.
func (m *ChatModel) SetChatService(newSvc *services.ChatService, newExecutorType string) {
	m.chatService = newSvc
	m.executorType = newExecutorType
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
		case "a", "A": // Yes, allow always
			m.awaitingToolConfirmation = false
			m.status = fmt.Sprintf("User confirmed '%s' (always). Resuming...", m.toolConfirmationRequest.ToolName)
			m.chatService.GetToolConfirmationChannel() <- types.ToolConfirmationOutcomeProceedAlways
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
		// This is just to trigger a re-render for the timer
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
		// Log every event received for debugging
		m.logSystemMessage(fmt.Sprintf("Received stream event: %T", msg.event))
		switch event := msg.event.(type) {
		case types.StreamingStartedEvent:
			m.status = "Stream started..."
		case types.ThinkingEvent:
			m.status = "Thinking..."
		case types.TodosSummaryUpdateEvent:
			m.todosSummary = event.Summary
		case types.ToolCallStartEvent:
			m.status = "Executing tool..."
			m.toolCallCount++ // Increment tool call count
			m.logSystemMessage(fmt.Sprintf("Tool Started: %s with args %v", event.ToolName, event.Args))
			// Don't add user_confirm to the visual chat history as a tool call,
			// as it's handled separately in the footer.
			if event.ToolName == types.USER_CONFIRM_TOOL_NAME {
				// We still need to track it as active for the UI state.
				m.activeToolCalls[event.ToolCallID] = &ToolCallStatus{
					ToolName: event.ToolName,
					Args:     event.Args,
					Status:   "Awaiting Confirmation",
				}
				break // Exit the switch case for this event
			}
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
				// Add the error as a message to the chat history for persistence.
				m.messages = append(m.messages, ErrorMessage{Err: fmt.Errorf("tool '%s' failed: %w", event.ToolName, event.Err)})
			}
			m.logSystemMessage(fmt.Sprintf("Tool Ended: %s. Result: %s, Error: %v", event.ToolName, event.Result, event.Err))
			if tc, ok := m.activeToolCalls[event.ToolCallID]; ok {
				tc.Status = "Completed"
				tc.Result = event.Result
				tc.Err = event.Err
			}
		case types.SubagentActivityEvent:
			var status strings.Builder
			status.WriteString(fmt.Sprintf("🤖 %s", event.AgentName))
			if event.Type == types.ActivityTypeError {
				if errorMsg, ok := event.Data["error"].(string); ok {
					status.WriteString(fmt.Sprintf(" 💥 Error: %s", errorMsg))
				}
			} else {
				if toolNameVal, ok := event.Data["toolName"]; ok {
					if toolName, isString := toolNameVal.(string); isString && toolName != "" {
						status.WriteString(fmt.Sprintf(" -> %s", toolName))
					}
				}
				if thoughtVal, ok := event.Data["thought"]; ok {
					if thought, isString := thoughtVal.(string); isString && thought != "" {
						status.WriteString(fmt.Sprintf(" 🤔 %s", thought))
					}
				}
			}
			m.subagentStatus = status.String()
		case types.FinalResponseEvent:
			botMsg := BotMessage{Content: event.Content}
			m.messages = append(m.messages, botMsg)
			m.logUIMessage(botMsg) // Log bot message
		case types.ErrorEvent:
			errmsg := ErrorMessage{Err: event.Err} // Corrected errmsg to errMsg
			m.messages = append(m.messages, errmsg)
			m.logUIMessage(errmsg) // Log error message
		case types.ModelSwitchEvent:
			m.executorType = event.NewModel
			botMsg := BotMessage{Content: fmt.Sprintf("Automatically switched from **%s** to **%s** due to: %s", event.OldModel, event.NewModel, event.Reason)}
			m.messages = append(m.messages, botMsg)
			m.logUIMessage(botMsg)
		}
		m.updateViewport()
		return m, waitForEvent(m.streamCh) // Continue waiting for events
	case streamErrorMsg:
		m.err = msg.err
		m.status = "Error"
		errmsg := ErrorMessage{Err: msg.err} // Corrected errmsg to errMsg
		// m.messages = append(m.messages, errmsg) // Uncomment if you want to keep error messages in history
		m.logUIMessage(errmsg)
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
			errmsg := ErrorMessage{Err: msg.err} // Corrected errmsg to errMsg
			m.messages = append(m.messages, errmsg)
			m.logUIMessage(errmsg)
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
					// Trigger hot-reload of chat service
					m.status = "Settings changed, reinitializing chat service..."
					return m, ReinitializeChatCmd(m.chatService, m.sessionService, m.config)
				}
			}
		}
		m.updateViewport()
		return m, nil
	case compressionResultMsg:
		m.isStreaming = false
		m.status = "Ready"
		if msg.err != nil {
			errmsg := ErrorMessage{Err: msg.err}
			m.messages = append(m.messages, errmsg)
		} else {
			compressionMessage := fmt.Sprintf("History compressed successfully. Original tokens: %d, New tokens: %d.", msg.result.OriginalTokenCount, msg.result.NewTokenCount)
			botMsg := BotMessage{Content: compressionMessage}
			m.messages = append(m.messages, botMsg)
		}
		m.updateViewport()
		return m, nil
	case chatServiceReloadedMsg:
		m.SetChatService(msg.newService, msg.newExecutorType)
		m.messages = m.repopulateMessagesFromHistory(m.chatService.GetHistory()) // Repopulate messages from the new service's history
		m.status = fmt.Sprintf("Switched to executor: %s", msg.newExecutorType)
		botMsg := BotMessage{Content: fmt.Sprintf("Switched to executor: %s", msg.newExecutorType)}
		m.messages = append(m.messages, botMsg)
		m.logUIMessage(botMsg)
		m.updateViewport()
		return m, nil
	}
	return m, tea.Batch(cmds...)
}
func (m *ChatModel) View() string {
	var contextInfo string
	if m.contextFile != "" {
		contextInfo = lipgloss.NewStyle().
			PaddingLeft(1).
			Foreground(lipgloss.Color("240")).
			Render(fmt.Sprintf("Using: 1 %s file", m.contextFile))
	}
	if m.todosSummary != "" {
		if contextInfo != "" {
			contextInfo += lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(" | ")
		}
		contextInfo += lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Render(m.todosSummary)
	}
	if m.subagentStatus != "" {
		subagentLine := lipgloss.NewStyle().
			Foreground(lipgloss.Color("14")).
			Render(m.subagentStatus)
		contextInfo = lipgloss.JoinVertical(lipgloss.Left, contextInfo, subagentLine)
	}
	return fmt.Sprintf(
		"%s\n%s\n%s\n%s",
		m.viewport.View(),
		contextInfo,
		m.renderFooter(),
		m.textarea.View(),
	)
}
func (m *ChatModel) renderFooter() string {
	// While streaming or awaiting confirmation, show the spinner and status.
	if m.isStreaming || m.awaitingToolConfirmation {
		statusLine := lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Render(m.status) // Always blue in active state
		var finalRender string
		if m.awaitingToolConfirmation && m.toolConfirmationRequest != nil {
			options := "(y: allow once, a: allow always, n: cancel, m: modify)"
			messageStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Width(m.viewport.Width - lipgloss.Width(m.spinner.View()) - 1)
			var finalRender string
			if m.toolConfirmationRequest.Type == "edit" && m.toolConfirmationRequest.FileDiff != "" {
				// For edits, show message, options, and then the diff on subsequent lines.
				messageLine := messageStyle.Render(m.toolConfirmationRequest.Message)
				optionsLine := lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render(options)
				// Explicitly style the diff to ensure it's monospaced and readable
				diffStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("248"))
				styledDiff := diffStyle.Render(m.toolConfirmationRequest.FileDiff)
				finalRender = lipgloss.JoinVertical(lipgloss.Left, messageLine, optionsLine, styledDiff)
			} else {
				// For other confirmations, show message and options.
				fullMessage := fmt.Sprintf("%s %s", m.toolConfirmationRequest.Message, options)
				finalRender = messageStyle.Render(fullMessage)
			}
			return m.spinner.View() + " " + finalRender
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
		cwd = "???'"
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
	// ---
	// UI-only commands
	// ---
	if len(args) > 0 {
		switch args[0] {
		case "compress":
			m.status = "Compressing history..."
			m.isStreaming = true // Show spinner
			return m, compressHistoryCmd(m.chatService)
		case "clear":
			m.chatService.ClearHistory()
			m.messages = createInitialMessages()
			m.updateViewport()
			return m, nil
		case "help":
			helpText := `
**Available Commands**
**Session Management**
* ` + "`/clear`" + ` - Clears the current chat session history.
* ` + "`/compress`" + ` - Summarizes the current session to save tokens.
* ` + "`/sessions list`" + ` - Lists all saved chat sessions.
* ` + "`/sessions resume <id>`" + ` - Resumes a session by its ID or list number.
**Application**
* ` + "`/help`" + ` - Shows this help message.
* ` + "`/quit` or `/exit`" + ` - Exits the application.
**Settings**
* ` + "`/settings list`" + ` - Shows current settings.
* ` + "`/settings set <key> <value>`" + ` - Changes a setting (e.g., ` + "`/settings set executor qwen`" + `).
* ` + "`/settings get <key>`" + ` - Retrieves the value of a setting.
* ` + "`/settings reset`" + ` - Resets all settings to their default values.
*Most other commands from ` + "`main-agent --help`" + ` can also be run with a ` + "`/`" + ` prefix.*
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
				newChatService, err := services.NewChatService(m.chatService.GetExecutor(), m.chatService.GetToolRegistry(), m.sessionService, m.chatService.GetSettingsService(), m.contextService, m.config, m.chatService.GetGenerationConfig(), nil)
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
	// ---
	// Safety Check
	// ---
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
	// ---
	// End Safety Check
	// ---
	telemetry.LogDebugf("Executing command asynchronously: %s with args: %v", commandString, args)
	m.status = fmt.Sprintf("Executing `/%s`...", commandString)
	m.isStreaming = true
	return m, executeCommandCmd(m.commandExecutor, args)
}
func (m *ChatModel) repopulateMessagesFromHistory(history []*types.Content) []Message {
	newMessages := createInitialMessages()
	for _, content := range history {
		if content.Role == "user" {
			isToolResponse := false
			for _, part := range content.Parts {
				if part.FunctionResponse != nil {
					isToolResponse = true
					break
				}
			}
			if isToolResponse {
				continue
			}
			var textContent strings.Builder
			for _, part := range content.Parts {
				textContent.WriteString(part.Text)
			}
			if textContent.Len() > 0 {
				newMessages = append(newMessages, UserMessage{Content: textContent.String()})
			}
		} else if content.Role == "model" {
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
				group := &ToolCallGroupMessage{ToolCalls: make(map[string]*ToolCallStatus)}
				for i, fc := range functionCalls {
					toolCallID := fmt.Sprintf("resumed-tool-%d-%d", len(newMessages), i)
					status := &ToolCallStatus{
						ToolName: fc.Name,
						Args:     fc.Args,
						Status:   "Completed",
						Result:   "(resumed from history)",
						Err:      nil,
					}
					group.ToolCalls[toolCallID] = status
				}
				newMessages = append(newMessages, group)
			}
			if len(textParts) > 0 {
				fullText := strings.Join(textParts, "")
				newMessages = append(newMessages, BotMessage{Content: fullText})
			}
		}
	}
	return newMessages
}

// ---
// Commands
// ---
func (m *ChatModel) logUIMessage(msg Message) {
	if m.logWriter == nil {
		return
	}
	renderedString := msg.Render(m)
	cleanString := utils.StripAnsi(renderedString)
	_, err := m.logWriter.WriteString(cleanString + "\n")
	if err != nil {
		telemetry.LogErrorf("Error writing to chat log: %v", err)
	}
	m.logWriter.Flush()
}
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
func (m *ChatModel) startStreaming(userInput string) tea.Cmd {
	return func() tea.Msg {
		m.cancelCtx, m.cancelFunc = context.WithCancel(context.Background())
		stream, err := m.chatService.SendMessage(m.cancelCtx, m.sessionID, userInput)
		if err != nil {
			return streamErrorMsg{err}
		}
		return streamChannelMsg{ch: stream}
	}
}

// ReinitializeChatCmd creates a new chat service based on current settings and transfers state.
func ReinitializeChatCmd(oldChatService *services.ChatService, sessionService *services.SessionService, config types.Config) tea.Cmd {
	return func() tea.Msg {
		// 1. Capture the current state from the old chat service
		chatState := oldChatService.GetState()
		// 2. Read the latest executor and model settings
		settingsService := oldChatService.GetSettingsService()
		executorTypeVal, _ := settingsService.Get("executor")
		executorType, ok := executorTypeVal.(string)
		if !ok {
			executorType = "gemini" // Fallback
		}
		modelVal, _ := settingsService.Get("model")
		model, ok := modelVal.(string)
		if !ok {
			model = "gemini-pro" // Fallback
		}
		appConfig := config.WithModel(model)
		// 3. Create a new Executor instance
		executorFactory, err := core.NewExecutorFactory(executorType, appConfig)
		if err != nil {
			return streamErrorMsg{fmt.Errorf("error creating new executor factory: %w", err)}
		}
		newExecutor, err := executorFactory.NewExecutor(appConfig, types.GenerateContentConfig{}, chatState.History) // Pass history for context
		if err != nil {
			return streamErrorMsg{fmt.Errorf("error creating new executor: %w", err)}
		}
		// 4. Create a new ChatService instance, passing the captured state
		newChatService, err := services.NewChatService(
			newExecutor,
			oldChatService.GetToolRegistry(),
			sessionService,
			settingsService,
			oldChatService.GetContextService(),
			appConfig,
			oldChatService.GetGenerationConfig(), // Pass the existing generation config
			chatState,                            // Pass the captured state
		)
		if err != nil {
			return streamErrorMsg{fmt.Errorf("error creating new chat service: %w", err)}
		}
		return chatServiceReloadedMsg{newService: newChatService, newExecutorType: executorType}
	}
}
func compressHistoryCmd(cs *services.ChatService) tea.Cmd {
	return func() tea.Msg {
		res, err := cs.CompressHistory()
		return compressionResultMsg{result: res, err: err}
	}
}
