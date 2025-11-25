package tools

import (
	"context"
	"fmt"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/services"
	"go-ai-agent-v2/go-cli/pkg/types"
)

// ListDirectoryTool implements the Tool interface for listing directory contents.
type ListDirectoryTool struct {
	*types.BaseDeclarativeTool // Embed BaseDeclarativeTool directly
	fileSystemService services.FileSystemService // Use the interface
}

// NewListDirectoryTool creates a new ListDirectoryTool.
func NewListDirectoryTool(fs services.FileSystemService) *ListDirectoryTool {
	return &ListDirectoryTool{
		BaseDeclarativeTool: types.NewBaseDeclarativeTool(
		types.LIST_DIRECTORY_TOOL_NAME,
		"List Directory",
		"Lists the names of files and subdirectories directly within a specified directory path. Can optionally ignore entries matching provided glob patterns.",
		types.KindOther,
					&types.JsonSchemaObject{
				Type: "object",
				Properties: map[string]*types.JsonSchemaProperty{
					"path": &types.JsonSchemaProperty{
						Type:        "string",
						Description: "The path to the directory to list.",
					},
					"ignore": &types.JsonSchemaProperty{
						Type:        "array",
						Description: "List of glob patterns to ignore",
						Items:       &types.JsonSchemaObject{Type: "string"},
					},
					"file_filtering_options": &types.JsonSchemaProperty{
						Type:        "object",
						Description: "Optional: Whether to respect ignore patterns from .gitignore or .geminiignore",
						Properties: map[string]*types.JsonSchemaProperty{
							"respect_git_ignore": &types.JsonSchemaProperty{
								Type:        "boolean",
								Description: "Optional: Whether to respect .gitignore patterns when listing files. Only available in git repositories. Defaults to true.",
							},
							"respect_gemini_ignore": &types.JsonSchemaProperty{
								Type:        "boolean",
								Description: "Optional: Whether to respect .geminiignore patterns when listing files. Defaults to true.",
							},
						},
					},
				},
				Required: []string{"path"},
			},
		false, // isOutputMarkdown
		false, // canUpdateOutput
		nil,   // MessageBus
	),
		fileSystemService: fs,
	}
}

// Execute performs the list_directory operation.
func (t *ListDirectoryTool) Execute(ctx context.Context, args map[string]any) (types.ToolResult, error) {
	path, ok := args["path"].(string)
	if !ok || path == "" {
		return types.ToolResult{}, fmt.Errorf("missing or invalid 'path' argument")
	}

	ignorePatterns := []string{}
	if ignore, ok := args["ignore"].([]interface{}); ok {
		for _, p := range ignore {
			if pattern, isString := p.(string); isString {
				ignorePatterns = append(ignorePatterns, pattern)
			}
		}
	}

	respectGitIgnore := true
	respectGoaiagentIgnore := true

	if fileFilteringOptions, ok := args["file_filtering_options"].(map[string]any); ok {
		if val, ok := fileFilteringOptions["respect_git_ignore"].(bool); ok {
			respectGitIgnore = val
		}
		if val, ok := fileFilteringOptions["respect_goaiagent_ignore"].(bool); ok {
			respectGoaiagentIgnore = val
		}
	}

	files, err := t.fileSystemService.ListDirectory(path, ignorePatterns, respectGitIgnore, respectGoaiagentIgnore)
	if err != nil {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: fmt.Sprintf("Failed to list directory %s: %v", path, err),
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("failed to list directory %s: %w", path, err)
	}

	return types.ToolResult{
		LLMContent:    strings.Join(files, "\n"),
		ReturnDisplay: fmt.Sprintf("Listed directory %s: %d items", path, len(files)),
	},
	nil
}