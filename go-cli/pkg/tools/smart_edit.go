package tools

import (
	"fmt"
	"io/ioutil"
	"strings"
)

// SmartEditTool represents the smart-edit tool.
type SmartEditTool struct {
}

// NewSmartEditTool creates a new instance of SmartEditTool.
func NewSmartEditTool() *SmartEditTool {
	return &SmartEditTool{}
}

// Execute performs a smart edit operation.
func (t *SmartEditTool) Execute(
	filePath string,
	instruction string,
	oldString string,
	newString string,
) (string, error) {
	// Read the file content
	contentBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	content := string(contentBytes)

	// Perform the replacement
	newContent := strings.Replace(content, oldString, newString, 1) // Replace only the first occurrence

	if newContent == content {
		return "", fmt.Errorf("old_string not found or no changes made in file %s", filePath)
	}

	// Write the new content back to the file
	err = ioutil.WriteFile(filePath, []byte(newContent), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	return fmt.Sprintf("Successfully modified file: %s", filePath), nil
}
