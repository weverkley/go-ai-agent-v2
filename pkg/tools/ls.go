package tools

import (
	"context" // New import
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/types"
)

// LsTool represents the ls tool.
type LsTool struct {
	*types.BaseDeclarativeTool
	workspaceService types.WorkspaceServiceIface
}

// NewLsTool creates a new instance of LsTool.
func NewLsTool(workspaceService types.WorkspaceServiceIface) *LsTool {
	return &LsTool{
		BaseDeclarativeTool: types.NewBaseDeclarativeTool(
			types.LS_TOOL_NAME,
			"List Directory Contents",
			"Lists the contents of a directory.",
			types.KindOther,
			(&types.JsonSchemaObject{
				Type: "object",
			}).SetProperties(map[string]*types.JsonSchemaProperty{
				"path": &types.JsonSchemaProperty{
					Type:        "string",
					Description: "The path to the directory to list, relative to the project root.",
				},
			}).SetRequired([]string{"path"}),
			false, // isOutputMarkdown
			false, // canUpdateOutput
			nil,   // MessageBus
		),
		workspaceService: workspaceService,
	}
}

// Execute runs the tool with the given arguments.
func (t *LsTool) Execute(ctx context.Context, args map[string]any) (types.ToolResult, error) {
	// Extract arguments
	path, ok := args["path"].(string)
	if !ok || path == "" {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: "path argument is required and must be a string",
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("path argument is required and must be a string")
	}

	projectRoot := t.workspaceService.GetProjectRoot()
	absolutePath := filepath.Join(projectRoot, path)

	entries, err := os.ReadDir(absolutePath)
	if err != nil {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: fmt.Sprintf("failed to read directory %s: %v", absolutePath, err),
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("failed to read directory %s: %w", absolutePath, err)
	}

	var fileNames []string
	for _, entry := range entries {
		fileNames = append(fileNames, entry.Name())
	}

	output := strings.Join(fileNames, "\n")

	return types.ToolResult{
		LLMContent:    output,
		ReturnDisplay: output,
	}, nil
}
