package tools

import (
	"context"
	"fmt"
	"os"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/services"
	"go-ai-agent-v2/go-cli/pkg/types"
)

// SmartEditTool represents the smart-edit tool.
type SmartEditTool struct {
	*types.BaseDeclarativeTool
	fileSystemService services.FileSystemService
}

// NewSmartEditTool creates a new instance of SmartEditTool.
func NewSmartEditTool(fileSystemService services.FileSystemService) *SmartEditTool {
	return &SmartEditTool{
		BaseDeclarativeTool: types.NewBaseDeclarativeTool(
			"smart_edit",
			"smart_edit",
			"Replaces text within a file. Replaces a single occurrence. This tool requires providing significant context around the change to ensure precise targeting. Always use the read_file tool to examine the file's current content before attempting a text replacement.",
			types.KindOther, // Assuming KindOther for now
			types.JsonSchemaObject{
				Type: "object",
				Properties: map[string]types.JsonSchemaProperty{
					"file_path": {
						Type:        "string",
						Description: "The absolute path to the file to modify. Must start with '/'.",
					},
					"instruction": {
						Type:        "string",
						Description: "A clear, semantic instruction for the code change, acting as a high-quality prompt for an expert LLM assistant.",
					},
					"old_string": {
						Type:        "string",
						Description: "The exact literal text to replace (including all whitespace, indentation, newlines, and surrounding code etc.).",
					},
					"new_string": {
						Type:        "string",
						Description: "The exact literal text to replace `old_string` with (also including all whitespace, indentation, newlines, and surrounding code etc.). Ensure the resulting code is correct and idiomatic and that `old_string` and `new_string` are different.",
					},
				},
				Required: []string{"file_path", "instruction", "old_string", "new_string"},
			},
			false, // isOutputMarkdown
			false, // canUpdateOutput
			nil,   // MessageBus
		),
		fileSystemService: fileSystemService,
	}
}

// Execute performs a smart edit operation.
func (t *SmartEditTool) Execute(ctx context.Context, args map[string]any) (types.ToolResult, error) {
	filePath, ok := args["file_path"].(string)
	if !ok || filePath == "" {
		return types.ToolResult{}, fmt.Errorf("invalid or missing 'file_path' argument")
	}

	// instruction is mainly for the LLM, not used directly in this simplified Go version yet.
	_, ok = args["instruction"].(string)
	if !ok {
		return types.ToolResult{}, fmt.Errorf("invalid or missing 'instruction' argument")
	}

	oldString, _ := args["old_string"].(string)

	newString, ok := args["new_string"].(string)
	if !ok {
		return types.ToolResult{}, fmt.Errorf("invalid or missing 'new_string' argument")
	}

	// Read the file content
	content, err := t.fileSystemService.ReadFile(filePath)
	if err != nil {
		// If file doesn't exist and old_string is empty, it's a new file creation
		if os.IsNotExist(err) && oldString == "" {
			err = t.fileSystemService.WriteFile(filePath, newString)
			if err != nil {
				return types.ToolResult{}, fmt.Errorf("failed to create new file %s: %w", filePath, err)
			}
			successMessage := fmt.Sprintf("Successfully created new file: %s", filePath)
			return types.ToolResult{
				LLMContent:    successMessage,
				ReturnDisplay: successMessage,
			}, nil
		}
		return types.ToolResult{}, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// If old_string is empty but file exists, it's an error
	if oldString == "" {
		return types.ToolResult{}, fmt.Errorf("file already exists, cannot create: %s. Use a non-empty old_string to modify.", filePath)
	}

	// Perform the replacement
	newContent := strings.Replace(content, oldString, newString, 1) // Replace only the first occurrence

	if newContent == content {
		return types.ToolResult{}, fmt.Errorf("old_string not found or no changes made in file %s", filePath)
	}

	// Write the new content back to the file
	err = t.fileSystemService.WriteFile(filePath, newContent)
	if err != nil {
		return types.ToolResult{}, fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	successMessage := fmt.Sprintf("Successfully modified file: %s", filePath)
	return types.ToolResult{
		LLMContent:    successMessage,
		ReturnDisplay: successMessage,
	}, nil
}
