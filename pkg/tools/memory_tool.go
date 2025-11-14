package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/types"
)

const (
	DEFAULT_CONTEXT_FILENAME = "GEMINI.md"
	MEMORY_SECTION_HEADER    = "## Gemini Added Memories"
)

// MemoryTool represents the memory tool.
type MemoryTool struct {
	*types.BaseDeclarativeTool
}

// NewMemoryTool creates a new instance of MemoryTool.
func NewMemoryTool() *MemoryTool {
	return &MemoryTool{
		types.NewBaseDeclarativeTool(
			"save_memory",
			"save_memory",
			"Saves a specific piece of information or fact to your long-term memory.",
			types.KindOther, // Assuming KindOther for now
			types.JsonSchemaObject{
				Type: "object",
				Properties: map[string]types.JsonSchemaProperty{
					"fact": {
						Type:        "string",
						Description: "The specific fact or piece of information to remember. Should be a clear, self-contained statement.",
					},
				},
				Required: []string{"fact"},
			},
			false, // isOutputMarkdown
			false, // canUpdateOutput
			nil,   // MessageBus
		),
	}
}

// getGlobalMemoryFilePath returns the path to the GEMINI.md file.
func getGlobalMemoryFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(homeDir, ".gemini", DEFAULT_CONTEXT_FILENAME), nil
}

// readMemoryFileContent reads the current content of the memory file.
func readMemoryFileContent() (string, error) {
	memoryFilePath, err := getGlobalMemoryFilePath()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(memoryFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // File doesn't exist, return empty content
		}
		return "", fmt.Errorf("failed to read memory file: %w", err)
	}
	return string(data), nil
}

// ensureNewlineSeparation ensures proper newline separation before appending content.
func ensureNewlineSeparation(currentContent string) string {
	if currentContent == "" {
		return ""
	}
	if strings.HasSuffix(currentContent, "\n\n") || strings.HasSuffix(currentContent, "\r\n\r\n") {
		return ""
	}
	if strings.HasSuffix(currentContent, "\n") || strings.HasSuffix(currentContent, "\r\n") {
		return "\n"
	}
	return "\n\n"
}

// computeNewContent computes the new content that would result from adding a memory entry.
func computeNewContent(currentContent, fact string) string {
	processedText := strings.TrimSpace(fact)
	processedText = strings.TrimPrefix(processedText, "- ") // Remove leading hyphen if present
	newMemoryItem := fmt.Sprintf("- %s", processedText)

	headerIndex := strings.Index(currentContent, MEMORY_SECTION_HEADER)

	if headerIndex == -1 {
		// Header not found, append header and then the entry
		separator := ensureNewlineSeparation(currentContent)
		return currentContent + separator + MEMORY_SECTION_HEADER + "\n" + newMemoryItem + "\n"
	} else {
		// Header found, find where to insert the new memory entry
		startOfSectionContent := headerIndex + len(MEMORY_SECTION_HEADER)
		endOfSectionIndex := strings.Index(currentContent[startOfSectionContent:], "\n## ")
		if endOfSectionIndex == -1 {
			endOfSectionIndex = len(currentContent) // End of file
		} else {
			endOfSectionIndex += startOfSectionContent
		}

		beforeSectionMarker := strings.TrimRight(currentContent[:startOfSectionContent], " \t\n\r")
		sectionContent := strings.TrimRight(currentContent[startOfSectionContent:endOfSectionIndex], " \t\n\r")
		afterSectionMarker := currentContent[endOfSectionIndex:]

		sectionContent += "\n" + newMemoryItem
		return strings.TrimRight(beforeSectionMarker+"\n"+strings.TrimLeft(sectionContent, " \t\n\r")+"\n"+afterSectionMarker, " \t\n\r") + "\n"
	}
}

// Execute saves a fact to long-term memory.
func (t *MemoryTool) Execute(ctx context.Context, args map[string]any) (types.ToolResult, error) {
	fact, ok := args["fact"].(string)
	if !ok || fact == "" {
		return types.ToolResult{}, fmt.Errorf("invalid or missing 'fact' argument")
	}

	memoryFilePath, err := getGlobalMemoryFilePath()
	if err != nil {
		return types.ToolResult{}, err
	}

	err = os.MkdirAll(filepath.Dir(memoryFilePath), 0755)
	if err != nil {
		return types.ToolResult{}, fmt.Errorf("failed to create memory directory: %w", err)
	}

	currentContent, err := readMemoryFileContent()
	if err != nil {
		return types.ToolResult{}, err
	}

	newContent := computeNewContent(currentContent, fact)

	err = os.WriteFile(memoryFilePath, []byte(newContent), 0644)
	if err != nil {
		return types.ToolResult{}, fmt.Errorf("failed to write memory file: %w", err)
	}

	successMessage := fmt.Sprintf("Okay, I've remembered that: \"%s\"", fact)
	return types.ToolResult{
		LLMContent:    successMessage,
		ReturnDisplay: successMessage,
	}, nil
}
