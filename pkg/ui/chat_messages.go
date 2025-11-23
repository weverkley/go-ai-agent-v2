package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
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
	ids := make([]string, 0, len(msg.ToolCalls))
	for id := range msg.ToolCalls {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for i, id := range ids {
		tc := msg.ToolCalls[id]
		var boxContent strings.Builder
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
		case "smart_edit":
			actionWord = "Smart Editing:"
			if filePath, ok := tc.Args["file_path"].(string); ok {
				argument = filePath
			}
		case "user_confirm":
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
		if (tc.ToolName == "write_file" || tc.ToolName == "smart_edit") && tc.Args != nil {
			var contentToDisplay string
			var filePathForLexer string
			if tc.ToolName == "write_file" {
				contentToDisplay, _ = tc.Args["content"].(string)
				filePathForLexer, _ = tc.Args["file_path"].(string)
			} else if tc.ToolName == "smart_edit" {
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
					boxContent.WriteString("\n\n")
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
	return m.suggestionStyle.Width(m.viewport.Width - 10).Render(fmt.Sprintf("Tip: %s", msg.Content))
}
