package tools

import (
	"context"
	"fmt"

	"go-ai-agent-v2/go-cli/pkg/services"
	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/google/generative-ai-go/genai"
)

const WRITE_FILE_TOOL_NAME = "write_file"

// WriteFileTool is a tool that writes content to a specified file.
type WriteFileTool struct {
	*types.BaseDeclarativeTool
	fileSystemService services.FileSystemService
}

// NewWriteFileTool creates a new WriteFileTool.
func NewWriteFileTool(fileSystemService services.FileSystemService) *WriteFileTool {
	return &WriteFileTool{
		BaseDeclarativeTool: types.NewBaseDeclarativeTool(
			WRITE_FILE_TOOL_NAME,
			"Write File",
			"Writes content to a specified file in the local filesystem.",
			types.KindOther,
			&types.JsonSchemaObject{
				Type: "object",
				Properties: map[string]*types.JsonSchemaProperty{
					"file_path": &types.JsonSchemaProperty{
						Type:        "string",
						Description: "The absolute path to the file to write to (e.g., '/home/user/project/file.txt'). Relative paths are not supported.",
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

	err := t.fileSystemService.WriteFile(filePath, content)
	if err != nil {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: fmt.Sprintf("Failed to write to file %s: %v", filePath, err),
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("failed to write to file %s: %w", filePath, err)
	}

	output := fmt.Sprintf("Successfully wrote to file: %s", filePath)
	return types.ToolResult{
		LLMContent:    output,
		ReturnDisplay: output,
	}, nil
}

// Definition returns the Gemini tool definition.
func (t *WriteFileTool) Definition() *genai.Tool {
	return t.BaseDeclarativeTool.Definition()
}
