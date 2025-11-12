package tools

import (
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
			"list_directory",
			"List Directory",
			"Lists the names of files and subdirectories directly within a specified directory path. Can optionally ignore entries matching provided glob patterns.",
			types.KindOther,
			types.JsonSchemaObject{
				Type: "object",
				Properties: map[string]types.JsonSchemaProperty{
					"path": {
						Type:        "string",
						Description: "The absolute path to the directory to list (must be absolute, not relative)",
					},
					"ignore": {
						Type:        "array",
						Description: "List of glob patterns to ignore",
						Items:       &types.JsonSchemaPropertyItem{Type: "string"},
					},
					"file_filtering_options": {
						Type:        "object",
						Description: "Whether to respect ignore patterns from .gitignore or .geminiignore",
						Properties: map[string]types.JsonSchemaProperty{
							"respect_git_ignore": {
								Type:        "boolean",
								Description: "Optional: Whether to respect .gitignore patterns when listing files. Only available in git repositories. Defaults to true.",
							},
							"respect_gemini_ignore": {
								Type:        "boolean",
								Description: "Optional: Whether to respect .geminiignore patterns when listing files. Defaults to true.",
							},
						},
					},
				},
				Required: []string{"path"},
			},
			false,
			false,
			nil,
		),
		fileSystemService: fs,
	}
}

// Execute performs the list_directory operation.
func (t *ListDirectoryTool) Execute(args map[string]any) (types.ToolResult, error) {
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
	if val, ok := args["respect_git_ignore"].(bool); ok {
		respectGitIgnore = val
	}

	respectGeminiIgnore := true
	if val, ok := args["respect_gemini_ignore"].(bool); ok {
		respectGeminiIgnore = val
	}

	files, err := t.fileSystemService.ListDirectory(path, ignorePatterns, respectGitIgnore, respectGeminiIgnore)
	if err != nil {
		return types.ToolResult{}, fmt.Errorf("failed to list directory %s: %w", path, err)
	}

	return types.ToolResult{
		LLMContent:    strings.Join(files, "\n"),
		ReturnDisplay: fmt.Sprintf("Listed directory %s: %d items", path, len(files)),
	},
	nil
}