package ui

import (
	"fmt"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/core"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// GenerateModel is the Bubble Tea model for the interactive generate UI.
type GenerateModel struct {
	textInput   textinput.Model
	err         error
	gemini      *core.GeminiChat
	history     []string
	awaitingGemini bool
	styles      *GenerateStyles
}

// GenerateStyles contains lipgloss styles for the UI.
type GenerateStyles struct {
	focusedPromptStyle lipgloss.Style
	blurredPromptStyle lipgloss.Style
	geminiResponseStyle lipgloss.Style
	userPromptStyle    lipgloss.Style
	errorStyle         lipgloss.Style
}

// NewGenerateModel creates a new GenerateModel with initialized text input.
func NewGenerateModel(gemini *core.GeminiChat) *GenerateModel {

ti := textinput.New()

ti.Placeholder = "Ask Gemini..."
ti.Focus()
ti.CharLimit = 256
ti.Width = 80
ti.Prompt = "> "

return &GenerateModel{
	textInput:   ti,
	gemini:      gemini,
	history:     []string{},
	awaitingGemini: false,
	styles: &GenerateStyles{
		focusedPromptStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("205")),
		blurredPromptStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		geminiResponseStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("63")).Bold(true),
		userPromptStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Italic(true),
		errorStyle:         lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true),
	},
}
}

// Init initializes the model.
func (m GenerateModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages and updates the model.
func (m GenerateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.awaitingGemini {
				return m, nil // Ignore input while awaiting response
			}
			query := m.textInput.Value()
			if strings.TrimSpace(query) == "" {
				return m, nil // Ignore empty queries
			}

			m.history = append(m.history, m.styles.userPromptStyle.Render("You: ")+query)
			m.textInput.Reset()
			m.awaitingGemini = true
			return m, m.generateContentCmd(query)
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}
	case geminiResponseMsg:
		m.history = append(m.history, m.styles.geminiResponseStyle.Render("Gemini: ")+string(msg))
		m.awaitingGemini = false
		return m, nil
	case errMsg:
		m.err = msg
		m.awaitingGemini = false
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// View renders the UI.
func (m GenerateModel) View() string {
	s := strings.Builder{}

	// History
	for _, item := range m.history {
		s.WriteString(item)
		s.WriteString("\n")
	}

	// Prompt
	s.WriteString("\n")
	if m.textInput.Focused() {
		s.WriteString(m.styles.focusedPromptStyle.Render(m.textInput.Prompt))
	} else {
		s.WriteString(m.styles.blurredPromptStyle.Render(m.textInput.Prompt))
	}
	s.WriteString(m.textInput.View())
	s.WriteString("\n")

	// Status/Error
	if m.awaitingGemini {
		s.WriteString("Gemini is thinking...")
	} else if m.err != nil {
		s.WriteString(m.styles.errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
	}

	return s.String()
}

// generateContentCmd is a command that sends the query to Gemini.
func (m GenerateModel) generateContentCmd(query string) tea.Cmd {
	return func() tea.Msg {
		// Use the geminiClient from the model
		content, err := m.gemini.GenerateContent(query)
		if err != nil {
			return errMsg{err}
		}
		return geminiResponseMsg(content)
	}
}

// geminiResponseMsg is a message type for Gemini's response.
type geminiResponseMsg string

// errMsg is a message type for errors.
type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }
