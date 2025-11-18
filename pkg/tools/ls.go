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
					"List Directory Contents",
					"Lists the contents of a directory.",
					types.KindOther,
					(&types.JsonSchemaObject{
						Type: "object",
					}).SetProperties(map[string]*types.JsonSchemaProperty{
						"path": &types.JsonSchemaProperty{
							Type:        "string",
							Description: "The path to the directory to list.",
						},
					}),
					false, // isOutputMarkdown
					false, // canUpdateOutput
					nil,   // MessageBus
				),	}
}

// Execute runs the tool with the given arguments.
func (t *LsTool) Execute(ctx context.Context, args map[string]any) (types.ToolResult, error) {
	path, ok := args["path"].(string)
	if !ok || path == "" {
		return types.ToolResult{}, fmt.Errorf("path argument is required and must be a string")
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: fmt.Sprintf("Failed to read directory %s: %v", path, err),
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("failed to read directory %s: %w", path, err)
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
