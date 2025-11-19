package ui

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/core"
	"go-ai-agent-v2/go-cli/pkg/core/agents"
	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type orchestratorState int

const (
	stateGetPrompt orchestratorState = iota
	stateProcessingModel
	stateExecutingTools
	stateDone
	stateError
)

type statusLog struct {
	message string
	isError bool
}

// --- Model ---

type OrchestratorModel struct {
	state         orchestratorState
	textarea      textarea.Model
	spinner       spinner.Model
	statusLogs    []statusLog
	finalContent  string
	err           error

	// Executor and Tools
	executor     core.Executor
	toolRegistry types.ToolRegistryInterface
	config       *config.Config

	// Loop State
	chatHistory      []*types.Content
	pendingToolCalls int
	toolResults      []types.Part
	toolResultsMutex sync.Mutex

	// Context
	ctx       context.Context
	cancelCtx context.CancelFunc
}

func NewOrchestratorModel(executor core.Executor, toolRegistry types.ToolRegistryInterface, cfg *config.Config) *OrchestratorModel {
	ta := textarea.New()
	ta.Placeholder = "Enter your prompt..."
	ta.Focus()
	ta.Prompt = "┃ "
	ta.CharLimit = 0
	ta.SetWidth(80)
	ta.SetHeight(1)
	ta.ShowLineNumbers = false
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	ctx, cancel := context.WithCancel(context.Background())

	return &OrchestratorModel{
		state:        stateGetPrompt,
		textarea:     ta,
		spinner:      s,
		statusLogs:   []statusLog{{message: "Waiting for prompt..."}},
		executor:     executor,
		toolRegistry: toolRegistry,
		config:       cfg,
		chatHistory:  []*types.Content{},
		ctx:          ctx,
		cancelCtx:    cancel,
	}
}

// --- Tea Commands & Messages ---

type modelResponseMsg struct {
	response *types.GenerateContentResponse
	err      error
}

type toolResultMsg struct {
	result types.Part
	err    error
}

func (m *OrchestratorModel) runModel() tea.Cmd {
	m.addLog("Calling model...", false)
	
	// Convert []types.Tool to []*types.ToolDefinition for the Executor call
	var toolDefinitions []*types.ToolDefinition
	if m.toolRegistry != nil {
		allTools := m.toolRegistry.GetAllTools()
		toolDefinitions = make([]*types.ToolDefinition, len(allTools))
		for i, t := range allTools {
			toolDefinitions[i] = &types.ToolDefinition{
				FunctionDeclarations: []*types.FunctionDeclaration{
					{
						Name:        t.Name(),
						Description: t.Description(),
						Parameters:  t.Parameters(),
					},
				},
			}
		}
	}

	return func() tea.Msg {
		resp, err := m.executor.GenerateContentWithTools(m.ctx, m.chatHistory, m.toolRegistry.GetAllTools())
		return modelResponseMsg{response: resp, err: err}
	}
}

func (m *OrchestratorModel) executeTool(fc *types.FunctionCall) tea.Cmd {
	return func() tea.Msg {
		m.addLog(fmt.Sprintf("Executing tool: %s", fc.Name), false)

		request := types.ToolCallRequestInfo{
			CallID: "not-implemented", // TODO: Generate unique ID
			Name:   fc.Name,
			Args:   fc.Args,
		}

		completedCall, err := agents.ExecuteToolCall(m.config, request, m.ctx)
		if err != nil {
			m.addLog(fmt.Sprintf("Tool execution failed for %s: %v", fc.Name, err), true)
			return toolResultMsg{
				result: types.Part{
					FunctionResponse: &types.FunctionResponse{
						Name:     fc.Name,
						Response: map[string]interface{}{"error": err.Error()},
					},
				},
				err: err,
			}
		}

		response := completedCall.GetResponse()
		var resultPart types.Part

		var toolOutputForModel map[string]interface{}
		if response != nil && response.Error != nil {
			toolOutputForModel = map[string]interface{}{"error": response.Error.Error()}
		} else if response != nil && len(response.ResponseParts) > 0 {
			resultPart = response.ResponseParts[0] // Use the first part directly
		} else {
			toolOutputForModel = map[string]interface{}{"status": "Tool executed successfully with no output."}
			resultPart = types.Part{FunctionResponse: &types.FunctionResponse{Name: fc.Name, Response: toolOutputForModel}}
		}

		return toolResultMsg{result: resultPart}
	}
}

// --- BubbleTea Implementation ---

func (m *OrchestratorModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m *OrchestratorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.cancelCtx()
			return m, tea.Quit
		case tea.KeyEnter:
			if m.state == stateGetPrompt {
				prompt := m.textarea.Value()
				m.textarea.Reset()
				m.state = stateProcessingModel
				m.addLog("Prompt received. Starting process...", false)
				m.chatHistory = append(m.chatHistory, &types.Content{
					Parts: []types.Part{{Text: prompt}},
					Role:  "user",
				})
				return m, m.runModel()
			}
		}

	case tea.WindowSizeMsg:
		m.textarea.SetWidth(msg.Width)
		return m, nil

	case modelResponseMsg:
		return m.handleModelResponse(msg)

	case toolResultMsg:
		return m.handleToolResult(msg)

	case error:
		m.err = msg
		m.state = stateError
		return m, tea.Quit
	}

	if m.state == stateProcessingModel || m.state == stateExecutingTools {
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *OrchestratorModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("An error occurred: %v\n\nPress Ctrl+C to exit.", m.err)
	}

	switch m.state {
	case stateGetPrompt:
		return fmt.Sprintf("Enter your request:\n\n%s", m.textarea.View())
	case stateProcessingModel, stateExecutingTools:
		var logLines strings.Builder
		for _, log := range m.statusLogs {
			if log.isError {
				logLines.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("Error: "+log.message) + "\n")
			} else {
				logLines.WriteString(log.message + "\n")
			}
		}
		return fmt.Sprintf("%s Processing...\n\n%s", m.spinner.View(), logLines.String())
	case stateDone:
		return fmt.Sprintf("✅ Processing complete.\n\nFinal Answer:\n%s\n\nPress Ctrl+C to exit.", m.finalContent)
	case stateError:
		return fmt.Sprintf("❌ An error occurred: %v\n\nPress Ctrl+C to exit.", m.err)
	default:
		return "Unknown state."
	}
}

// --- Loop Logic Handlers ---

func (m *OrchestratorModel) handleModelResponse(msg modelResponseMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.addLog(fmt.Sprintf("Model call failed: %v", msg.err), true)
		m.state = stateError
		m.err = msg.err
		return m, tea.Quit
	}

	if msg.response == nil || len(msg.response.Candidates) == 0 || msg.response.Candidates[0].Content == nil {
		m.addLog("Model returned no response.", true)
		m.state = stateError
		m.err = fmt.Errorf("empty response from model")
		return m, tea.Quit
	}

	m.chatHistory = append(m.chatHistory, msg.response.Candidates[0].Content)

	var toolCalls []*types.FunctionCall
	for _, part := range msg.response.Candidates[0].Content.Parts {
		if part.FunctionCall != nil {
			toolCalls = append(toolCalls, part.FunctionCall)
		}
	}

	if len(toolCalls) > 0 {
		m.state = stateExecutingTools
		m.pendingToolCalls = len(toolCalls)
		m.toolResults = make([]types.Part, 0, len(toolCalls))
		m.addLog(fmt.Sprintf("Model requested %d tool(s).", len(toolCalls)), false)

		var toolCmds []tea.Cmd
		for _, fc := range toolCalls {
			toolCmds = append(toolCmds, m.executeTool(fc))
		}
		return m, tea.Batch(toolCmds...)
	}

	m.state = stateDone
	for _, part := range msg.response.Candidates[0].Content.Parts {
		if part.Text != "" {
			m.finalContent += part.Text
		}
	}
	return m, tea.Quit
}

func (m *OrchestratorModel) handleToolResult(msg toolResultMsg) (tea.Model, tea.Cmd) {
	m.toolResultsMutex.Lock()
	defer m.toolResultsMutex.Unlock()

	if msg.err != nil {
		m.err = msg.err
		m.state = stateError
		return m, tea.Quit
	}
	m.toolResults = append(m.toolResults, msg.result)
	m.pendingToolCalls--

	if m.pendingToolCalls == 0 {
		m.addLog("All tools finished. Resuming model conversation...", false)
		m.chatHistory = append(m.chatHistory, &types.Content{
			Parts: m.toolResults,
			Role:  "tool",
		})
		m.state = stateProcessingModel
		return m, m.runModel()
	}

	return m, nil
}

func (m *OrchestratorModel) addLog(message string, isError bool) {
	m.statusLogs = append(m.statusLogs, statusLog{message: message, isError: isError})
}