package tools

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/services"
	"go-ai-agent-v2/go-cli/pkg/types"
)

// SmartEditTool represents the smart-edit tool.
type SmartEditTool struct {
	*types.BaseDeclarativeTool
	fileSystemService services.FileSystemService
	workspaceService  types.WorkspaceServiceIface
}

// NewSmartEditTool creates a new instance of SmartEditTool.
func NewSmartEditTool(fileSystemService services.FileSystemService, workspaceService types.WorkspaceServiceIface) *SmartEditTool {
	return &SmartEditTool{
		BaseDeclarativeTool: types.NewBaseDeclarativeTool(
			types.SMART_EDIT_TOOL_NAME,
			types.SMART_EDIT_TOOL_DISPLAY_NAME,
			"Replaces text within a file. Replaces a single occurrence. This tool requires providing significant context around the change to ensure precise targeting. Always use the read_file tool to examine the file's current content before attempting a text replacement.",
			types.KindOther, // Assuming KindOther for now
			&types.JsonSchemaObject{
				Type: "object",
				Properties: map[string]*types.JsonSchemaProperty{
					"file_path": &types.JsonSchemaProperty{
						Type:        "string",
						Description: "The path to the file to modify, relative to the project root.",
					},
					"instruction": &types.JsonSchemaProperty{
						Type:        "string",
						Description: "A clear, semantic instruction for the code change, acting as a high-quality prompt for an expert LLM assistant.",
					},
					"old_string": &types.JsonSchemaProperty{
						Type:        "string",
						Description: "The exact literal text to replace (including all whitespace, indentation, newlines, and surrounding code etc.).",
					},
					"new_string": &types.JsonSchemaProperty{
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
		workspaceService:  workspaceService,
	}
}

// Execute performs a smart edit operation.
func (t *SmartEditTool) Execute(ctx context.Context, args map[string]any) (types.ToolResult, error) {
	filePath, ok := args["file_path"].(string)
	if !ok || filePath == "" {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: "Invalid or missing 'file_path' argument",
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("invalid or missing 'file_path' argument")
	}

	// instruction is mainly for the LLM, not used directly in this simplified Go version yet.
	_, ok = args["instruction"].(string)
	if !ok {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: "Invalid or missing 'instruction' argument",
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("invalid or missing 'instruction' argument")
	}

	oldString, _ := args["old_string"].(string)

	newString, ok := args["new_string"].(string)
	if !ok {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: "Invalid or missing 'new_string' argument",
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("invalid or missing 'new_string' argument")
	}

	projectRoot := t.workspaceService.GetProjectRoot()
	absolutePath := filepath.Join(projectRoot, filePath)

	// Read the file content
	content, err := t.fileSystemService.ReadFile(absolutePath)
	if err != nil {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: fmt.Sprintf("Failed to read file %s: %v", absolutePath, err),
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("failed to read file %s: %w", absolutePath, err)
	}

	if oldString == "" {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: "old_string cannot be empty for smart_edit. Use write_file to create new files.",
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("old_string cannot be empty for smart_edit. Use write_file to create new files.")
	}

	// Perform the replacement
	newContent := strings.Replace(content, oldString, newString, 1) // Replace only the first occurrence

	if newContent == content {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: fmt.Sprintf("old_string not found or no changes made in file %s", absolutePath),
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("old_string not found or no changes made in file %s", absolutePath)
	}

	// Write the new content back to the file
	err = t.fileSystemService.WriteFile(absolutePath, newContent)
	if err != nil {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: fmt.Sprintf("Failed to write file %s: %v", absolutePath, err),
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("failed to write file %s: %w", absolutePath, err)
	}

	successMessage := fmt.Sprintf("Successfully modified file: %s", absolutePath)
	return types.ToolResult{
		LLMContent:    successMessage,
		ReturnDisplay: successMessage,
	}, nil
}
