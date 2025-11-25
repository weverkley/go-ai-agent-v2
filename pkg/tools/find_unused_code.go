package tools

import (
	"context"
	"fmt"
	"path/filepath"

	"go-ai-agent-v2/go-cli/pkg/core/agents"
	"go-ai-agent-v2/go-cli/pkg/types"
)

const FIND_UNUSED_CODE_TOOL_NAME = "find_unused_code"

// FindUnusedCodeTool is a tool that finds unused functions and methods in a given directory.
type FindUnusedCodeTool struct {
	*types.BaseDeclarativeTool
}

// NewFindUnusedCodeTool creates a new FindUnusedCodeTool.
func NewFindUnusedCodeTool() *FindUnusedCodeTool {
	return &FindUnusedCodeTool{
			BaseDeclarativeTool: types.NewBaseDeclarativeTool(
				types.FIND_UNUSED_CODE_TOOL_NAME,
				"Find Unused Code",
				"Finds unused code in the specified directory.",
				types.KindOther,
				&types.JsonSchemaObject{
					Type: "object",
									Properties: map[string]*types.JsonSchemaProperty{
										"dir_path": &types.JsonSchemaProperty{
											Type:        "string",
											Description: "The path to the directory to search for unused code.",
										},
									},					Required: []string{"dir_path"},
				},
				false, // isOutputMarkdown
				false, // canUpdateOutput
				nil,   // MessageBus
			),	}
}

// Execute implements the Tool interface.
func (t *FindUnusedCodeTool) Execute(ctx context.Context, args map[string]any) (types.ToolResult, error) {
	dirPath, ok := args["dir_path"].(string)
	if !ok || dirPath == "" {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: "missing or invalid 'dir_path' argument",
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("missing or invalid 'dir_path' argument")
	}

	// Resolve the absolute path
	absPath, err := filepath.Abs(dirPath)
	if err != nil {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: fmt.Sprintf("Failed to resolve absolute path for directory '%s': %v", dirPath, err),
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("failed to resolve absolute path for directory '%s': %w", dirPath, err)
	}

	unusedFunctions, err := agents.FindUnusedFunctions(absPath)
	if err != nil {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: fmt.Sprintf("Failed to find unused functions in '%s': %v", absPath, err),
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("failed to find unused functions in '%s': %w", absPath, err)
	}

	// Format the output
	var output string
	if len(unusedFunctions) == 0 {
		output = fmt.Sprintf("No unused functions or methods found in '%s'.", absPath)
	} else {
		output = fmt.Sprintf("Unused functions and methods found in '%s':\n", absPath)
		for _, fn := range unusedFunctions {
			output += fmt.Sprintf("- %s (%s) in %s\n", fn.Name, fn.Type, fn.FilePath)
		}
	}

	return types.ToolResult{
		LLMContent:    output,
		ReturnDisplay: output,
	}, nil
}

