package ui

import (
	"fmt"
	"strings"
	"time"

	"go-ai-agent-v2/go-cli/pkg/core"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/generative-ai-go/genai"
)

// FindDocsModel is the Bubble Tea model for the interactive find-docs UI.
type FindDocsModel struct {
	textInput      textinput.Model
	err            error
	gemini         *core.GeminiChat
	history        []string
	awaitingGemini bool
	spinner        spinner.Model // New field for the spinner
	styles         *FindDocsStyles
	currentError   string        // Stores the current error message
	errorTimer     *time.Timer   // Timer to clear the error message
}

// FindDocsStyles contains lipgloss styles for the UI.
type FindDocsStyles struct {
	focusedPromptStyle  lipgloss.Style
	blurredPromptStyle  lipgloss.Style
	geminiResponseStyle lipgloss.Style
	userPromptStyle     lipgloss.Style
	errorStyle          lipgloss.Style
}

// NewFindDocsModel creates a new FindDocsModel with initialized text input.
func NewFindDocsModel(gemini *core.GeminiChat) *FindDocsModel {

	ti := textinput.New()

	ti.Placeholder = "Ask Gemini to find documentation..."
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 80
	ti.Prompt = "> "

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))

	return &FindDocsModel{
		textInput:      ti,
		gemini:         gemini,
		history:        []string{},
		awaitingGemini: false,
		spinner:        s, // Initialize the spinner
		styles: &FindDocsStyles{
			focusedPromptStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("205")),
			blurredPromptStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
			geminiResponseStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("63")).Bold(true),
			userPromptStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Italic(true),
			errorStyle:          lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true),
		},
		currentError: "", // Initialize with no error
	}
}

// Init initializes the model.
func (m FindDocsModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}

// Update handles messages and updates the model.
func (m FindDocsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	// Update the spinner
	if m.awaitingGemini {
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.awaitingGemini {
				return m, nil // Ignore input while awaiting response
			}
			query := m.textInput.Value()
			if strings.TrimSpace(query) == "" {
				m.currentError = m.styles.errorStyle.Render("Error: Query cannot be empty.")
				if m.errorTimer != nil {
					m.errorTimer.Stop()
				}
				return m, ClearErrorCmd()			}

			m.history = append(m.history, m.styles.userPromptStyle.Render("You: ")+query)
			m.textInput.Reset()
			m.awaitingGemini = true
			return m, tea.Batch(append(cmds, m.findDocsCmd(query))...)
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}
	case findDocsResponseMsg:
		// Extract text from the response
		var textResponse string
		if msg.Value != nil && len(msg.Value.Candidates) > 0 && msg.Value.Candidates[0].Content != nil {
			for _, part := range msg.Value.Candidates[0].Content.Parts {
				if txt, ok := part.(genai.Text); ok {
					textResponse += string(txt)
				}
			}
		}
		m.history = append(m.history, m.styles.geminiResponseStyle.Render("Gemini: ")+textResponse)
		m.awaitingGemini = false
		return m, nil
	case ErrMsg:
		m.err = msg.Err // Store the actual error object
		m.currentError = m.styles.errorStyle.Render(fmt.Sprintf("Error: %v", msg.Err))
		m.awaitingGemini = false
		if m.errorTimer != nil {
			m.errorTimer.Stop()
		}
		m.errorTimer = time.NewTimer(5 * time.Second) // Start a timer to clear the error
		return m, ClearErrorCmd()
	case ClearErrorMsg:
		m.currentError = "" // Clear the error message
		m.err = nil         // Clear the error object
		if m.errorTimer != nil {
			m.errorTimer.Stop()
			m.errorTimer = nil
		}
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

// View renders the UI.
func (m FindDocsModel) View() string {
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
		s.WriteString(m.spinner.View()) // Render the spinner
		s.WriteString(" Gemini is thinking...")
	} else if m.currentError != "" { // Render currentError if present
		s.WriteString(m.currentError)
	}

	return s.String()
}

// findDocsCmd is a command that sends the query to Gemini.
func (m FindDocsModel) findDocsCmd(query string) tea.Cmd {
	return func() tea.Msg {
		// Use the geminiClient from the model
		content, err := m.gemini.GenerateContent(core.NewUserContent(query))
		if err != nil {
			return ErrMsg{Err: err}
		}
		return findDocsResponseMsg{Value: content}
	}
}

// findDocsResponseMsg is a message type for Gemini's response.
type findDocsResponseMsg struct {
	Value *genai.GenerateContentResponse
}
