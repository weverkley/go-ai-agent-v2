package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/services"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// LsModel represents the state of the interactive ls UI.
type LsModel struct {
	fileSystemService services.FileSystemService
	currentPath       string
	entries           []string
	cursor            int
	err               error
	quitting          bool
	width             int
	height            int
}

// NewLsModel creates a new LsModel.
func NewLsModel(fs services.FileSystemService, initialPath string) *LsModel {
	return &LsModel{
		fileSystemService: fs,
		currentPath:       initialPath,
	}
}

// Init initializes the model.
func (m *LsModel) Init() tea.Cmd {
	return m.listDirectoryCmd()
}

// Update handles messages and updates the model.
func (m *LsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.entries)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.entries) == 0 {
				return m, nil
			}
			selectedEntry := m.entries[m.cursor]
			newPath := filepath.Join(m.currentPath, selectedEntry)

			isDir, err := m.fileSystemService.IsDirectory(newPath)
			if err != nil {
				m.err = err
				return m, nil
			}

			if isDir {
				m.currentPath = newPath
				m.cursor = 0 // Reset cursor when changing directory
				return m, m.listDirectoryCmd()
			} else {
				// For now, just print the file path and quit
				fmt.Printf("Selected file: %s\n", newPath)
				m.quitting = true
				return m, tea.Quit
			}
		case "backspace", "h":
			parentPath := filepath.Dir(m.currentPath)
			if parentPath != m.currentPath { // Prevent going above root
				m.currentPath = parentPath
				m.cursor = 0 // Reset cursor when changing directory
				return m, m.listDirectoryCmd()
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case errMsg:
		m.err = msg
		return m, nil
	case []string: // Message containing directory entries
		m.entries = msg
		if m.cursor >= len(m.entries) {
			m.cursor = len(m.entries) - 1
		}
		if m.cursor < 0 && len(m.entries) > 0 {
			m.cursor = 0
		}
		return m, nil
	}
	return m, nil
}

// View renders the UI.
func (m *LsModel) View() string {
	if m.quitting {
		return ""
	}

	s := strings.Builder{}
	s.WriteString(fmt.Sprintf("Current Path: %s\n\n", m.currentPath))

	if m.err != nil {
		s.WriteString(fmt.Sprintf("Error: %v\n\n", m.err))
	}

	if len(m.entries) == 0 {
		s.WriteString(" (empty directory)\n")
	} else {
		for i, entry := range m.entries {
			cursor := " " // no cursor
			if m.cursor == i {
				cursor = ">" // cursor!
			}

			entryPath := filepath.Join(m.currentPath, entry)
			isDir, err := m.fileSystemService.IsDirectory(entryPath)
			if err != nil {
				s.WriteString(fmt.Sprintf("%s %s (error: %v)\n", cursor, entry, err))
				continue
			}

			if isDir {
				s.WriteString(fmt.Sprintf("%s %s/\n", cursor, lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render(entry))) // Green for directories
			} else {
				s.WriteString(fmt.Sprintf("%s %s\n", cursor, entry))
			}
		}
	}

	s.WriteString("\nPress 'q' to quit, 'enter' to navigate/select, 'backspace' to go up.")
	return s.String()
}

// listDirectoryCmd is a tea.Cmd that lists the directory contents.
func (m *LsModel) listDirectoryCmd() tea.Cmd {
	return func() tea.Msg {
		entries, err := m.fileSystemService.ListDirectory(m.currentPath, []string{}, true, true)
		if err != nil {
			return errMsg(err)
		}
		return entries
	}
}

// errMsg is a custom error message type.
type errMsg error
