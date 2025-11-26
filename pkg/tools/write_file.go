package tools

import (
	"context"
	"fmt"
	"path/filepath" // Add this import

	"go-ai-agent-v2/go-cli/pkg/services"
	"go-ai-agent-v2/go-cli/pkg/types"
)

const WRITE_FILE_TOOL_NAME = "write_file"

// WriteFileTool is a tool that writes content to a specified file.
type WriteFileTool struct {
	*types.BaseDeclarativeTool
	fileSystemService services.FileSystemService
	workspaceService  types.WorkspaceServiceIface // Change to interface
}

// NewWriteFileTool creates a new WriteFileTool.
func NewWriteFileTool(fileSystemService services.FileSystemService, workspaceService types.WorkspaceServiceIface) *WriteFileTool {
	return &WriteFileTool{
		BaseDeclarativeTool: types.NewBaseDeclarativeTool(
			WRITE_FILE_TOOL_NAME,
			"Write File",
			"Writes content to a specified file in the local filesystem.",
			types.KindEdit,
			&types.JsonSchemaObject{
				Type: "object",
				Properties: map[string]*types.JsonSchemaProperty{
					"file_path": &types.JsonSchemaProperty{
						Type:        "string",
						Description: "The path to the file to write to, relative to the project root (e.g., 'hello.html', 'src/main.js').",
					},
					"content": &types.JsonSchemaProperty{
						Type:        "string",
						Description: "The content to write to the file.",
					},
				},
				Required: []string{"file_path", "content"},
			},
			false, // isOutputMarkdown
			true,  // canUpdateOutput - This tool modifies files
			nil,   // MessageBus
		),
		fileSystemService: fileSystemService,
		workspaceService:  workspaceService, // Initialize new field
	}
}

// Execute implements the Tool interface.
func (t *WriteFileTool) Execute(ctx context.Context, args map[string]any) (types.ToolResult, error) {
	filePath, ok := args["file_path"].(string)
	if !ok || filePath == "" {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: "Missing or invalid 'file_path' argument",
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("missing or invalid 'file_path' argument")
	}

	content, ok := args["content"].(string)
	if !ok {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: "Missing or invalid 'content' argument",
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("missing or invalid 'content' argument")
	}

	projectRoot := t.workspaceService.GetProjectRoot()
	resolvedPath := filepath.Join(projectRoot, filePath)

	err := t.fileSystemService.WriteFile(resolvedPath, content)
	if err != nil {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: fmt.Sprintf("Failed to write to file %s: %v", resolvedPath, err),
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("failed to write to file %s: %w", resolvedPath, err)
	}

	output := fmt.Sprintf("Successfully wrote to file: %s", resolvedPath)
	return types.ToolResult{
		LLMContent:    output,
		ReturnDisplay: output,
	}, nil
}
