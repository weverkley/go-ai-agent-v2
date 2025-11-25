package ui

import (
	"regexp"

	"github.com/charmbracelet/lipgloss"
)

// The styles to use for formatting. These can be customized.
var (
	boldStyle      = lipgloss.NewStyle().Bold(true)
	codeStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("12")) // Using the existing 'pathStyle' color
	codeBlockStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240"))
)

// Regular expressions to find markdown-like patterns.
var (
	// Regex for code blocks, with optional language hint. `(?s)` allows `.` to match newlines.
	codeBlockRegex = regexp.MustCompile("(?s)```([a-zA-Z0-9]*)\\n(.*?)\\n```")
	boldRegex      = regexp.MustCompile(`\*\*(.*?)\*\*`)
	codeRegex      = regexp.MustCompile("`([^`]+)`")
)

// FormatMessage applies simple styling to a string.
func FormatMessage(content string) string {
	// 1. Process multi-line code blocks first to avoid conflicts.
	formatted := codeBlockRegex.ReplaceAllStringFunc(content, func(s string) string {
		matches := codeBlockRegex.FindStringSubmatch(s)
		if len(matches) < 3 {
			return s // Should not happen, but safeguard.
		}
		// lang := matches[1] // Language hint is available if needed for syntax highlighting later
		code := matches[2]
		return codeBlockStyle.Render(code)
	})

	// 2. Apply bold style on the remaining text
	formatted = boldRegex.ReplaceAllStringFunc(formatted, func(s string) string {
		innerContent := s[2 : len(s)-2]
		return boldStyle.Render(innerContent)
	})

	// 3. Apply inline code style
	formatted = codeRegex.ReplaceAllStringFunc(formatted, func(s string) string {
		innerContent := s[1 : len(s)-1]
		return codeStyle.Render(innerContent)
	})

	return formatted
}
