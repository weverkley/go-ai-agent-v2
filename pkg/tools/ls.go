package tools

import (
	"context" // New import
	"fmt"
	"os"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/types"
)

// LsTool represents the ls tool.
type LsTool struct {
	*types.BaseDeclarativeTool
}

// NewLsTool creates a new instance of LsTool.
func NewLsTool() *LsTool {
	return &LsTool{
		BaseDeclarativeTool: types.NewBaseDeclarativeTool(
			types.LS_TOOL_NAME,
			"ls",
			"Lists the names of files and subdirectories directly within a specified directory path.",
			types.KindOther,
			types.JsonSchemaObject{
				Type: "object",
				Properties: map[string]types.JsonSchemaProperty{
					"path": {
						Type:        "string",
						Description: "The absolute path to the directory to list (must be absolute, not relative)",
					},
				},
				Required: []string{"path"},
			},
			false, // isOutputMarkdown
			false, // canUpdateOutput
			nil,   // MessageBus
		),
	}
}

// Execute runs the tool with the given arguments.
func (t *LsTool) Execute(ctx context.Context, args map[string]any) (types.ToolResult, error) {
	path, ok := args["path"].(string)
	if !ok || path == "" {
		return types.ToolResult{Error: &types.ToolError{Message: "path argument is required and must be a string"}}, nil
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return types.ToolResult{Error: &types.ToolError{Message: fmt.Sprintf("failed to read directory %s: %v", path, err)}}, nil
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
