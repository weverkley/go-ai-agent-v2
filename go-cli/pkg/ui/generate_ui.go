package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/core"
	"go-ai-agent-v2/go-cli/pkg/types"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/generative-ai-go/genai"
	lipgloss "github.com/charmbracelet/lipgloss"
)

// GenerateModel is the Bubble Tea model for the interactive generate command.
type GenerateModel struct {
	input        string
	output       string
	error        error
	quitting     bool
	generating   bool
	width        int
	height       int
	inputFocused bool
}

// NewGenerateModel creates a new GenerateModel.
func NewGenerateModel() GenerateModel {
	return GenerateModel{
		input:        "",
		output:       "",
		error:        nil,
		quitting:     false,
		generating:   false,
		inputFocused: true,
	}
}

// Init initializes the model.
func (m GenerateModel) Init() tea.Cmd {
	return tea.EnterAltScreen
}

// Update handles messages and updates the model.
func (m GenerateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			if m.inputFocused && !m.generating {
				m.generating = true
				// In a real implementation, this would trigger the AI generation.
				// For now, simulate a delay and set a dummy output.
				return m, tea.Batch(tea.Tick(2*time.Second, func(t time.Time) tea.Msg { return generateDoneMsg{} }), generateContentCmd(m.input))
			}
		case "backspace":
			if m.inputFocused && len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
		default:
			if m.inputFocused && !m.generating {
				m.input += msg.String()
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case generateDoneMsg:
		m.generating = false
		// This is where the actual AI output would be set.
		// For now, it's handled by generateContentCmd
	case generateContentResultMsg:
		m.output = string(msg)
		m.generating = false
		return m, nil
	case error:
		m.error = msg
		m.generating = false
		return m, nil
	}
	return m, nil
}

// View renders the UI.
func (m GenerateModel) View() string {
	if m.quitting {
		return ""
	}

	sb := strings.Builder{}
	sb.WriteString("\n")

	// Input area
	sb.WriteString(lipgloss.NewStyle().Bold(true).Render("Enter your prompt:"))
	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(0, 1).
		Width(m.width - 4).
		Render(m.input))
	sb.WriteString("\n\n")

	// Output area
	if m.generating {
		sb.WriteString(lipgloss.NewStyle().Faint(true).Render("Generating..."))
	} else if m.error != nil {
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(fmt.Sprintf("Error: %v", m.error)))
	} else if m.output != "" {
		sb.WriteString(lipgloss.NewStyle().Bold(true).Render("AI Response:"))
		sb.WriteString("\n")
		sb.WriteString(lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("205")).
			Padding(0, 1).
			Width(m.width - 4).
			Height(m.height - 10).
			Render(m.output))
	}

	return sb.String()
}

type generateDoneMsg struct{}

type generateContentResultMsg string

func generateContentCmd(prompt string) tea.Cmd {
	return func() tea.Msg {
		// Load the configuration within the command's Run function
		workspaceDir, err := os.Getwd()
		if err != nil {
			return err
		}
		loadedSettings := config.LoadSettings(workspaceDir)

		params := &config.ConfigParameters{
			Model: loadedSettings.Model,
		}
		appConfig := config.NewConfig(params)

		geminiClient, err := core.NewGeminiChat(appConfig, types.GenerateContentConfig{}, []*genai.Content{})
		if err != nil {
			return err
		}

		content, err := geminiClient.GenerateContent(prompt)
		if err != nil {
			return err
		}

		return generateContentResultMsg(content)
	}
}
