package tools

import (
	"context"
	"fmt"
	"go-ai-agent-v2/go-cli/pkg/analysis"
	"go-ai-agent-v2/go-cli/pkg/types"
)

const RENAME_SYMBOL_TOOL_NAME = "rename_symbol"

// RenameSymbolTool implements the Tool interface for safely renaming a symbol.
type RenameSymbolTool struct {
	*types.BaseDeclarativeTool
}

// NewRenameSymbolTool creates a new RenameSymbolTool.
func NewRenameSymbolTool() *RenameSymbolTool {
	return &RenameSymbolTool{
		BaseDeclarativeTool: types.NewBaseDeclarativeTool(
			RENAME_SYMBOL_TOOL_NAME,
			"Rename Symbol",
			"Safely renames a symbol (variable, function, struct, etc.) across all its usages in the codebase.",
			types.KindOther,
			(&types.JsonSchemaObject{
				Type: "object",
			}).SetProperties(map[string]*types.JsonSchemaProperty{
				"file_path": {
					Type:        "string",
					Description: "Path to the file containing the symbol's definition.",
				},
				"line": {
					Type:        "integer",
					Description: "The line number of the symbol (1-indexed).",
				},
				"column": {
					Type:        "integer",
					Description: "The column number of the symbol (1-indexed).",
				},
				"new_name": {
					Type:        "string",
					Description: "The new name for the symbol.",
				},
			}).SetRequired([]string{"file_path", "line", "column", "new_name"}),
			false, // isOutputMarkdown
			true,  // canUpdateOutput - This tool modifies files
			nil,   // MessageBus
		),
	}
}

// Execute performs the rename_symbol operation.
func (t *RenameSymbolTool) Execute(ctx context.Context, args map[string]any) (types.ToolResult, error) {
	filePath, ok := args["file_path"].(string)
	if !ok || filePath == "" {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: "missing or invalid 'file_path' argument",
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("missing or invalid 'file_path' argument")
	}

	lineFloat, ok := args["line"].(float64)
	if !ok {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: "missing or invalid 'line' argument",
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("missing or invalid 'line' argument")
	}
	line := int(lineFloat)

	columnFloat, ok := args["column"].(float64)
	if !ok {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: "missing or invalid 'column' argument",
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("missing or invalid 'column' argument")
	}
	column := int(columnFloat)

	newName, ok := args["new_name"].(string)
	if !ok || newName == "" {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: "missing or invalid 'new_name' argument",
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("missing or invalid 'new_name' argument")
	}

	err := analysis.RenameSymbol(filePath, line, column, newName)
	if err != nil {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: fmt.Sprintf("failed to rename symbol: %v", err),
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("failed to rename symbol: %w", err)
	}

	output := fmt.Sprintf("Successfully renamed symbol at %s:%d:%d to '%s'", filePath, line, column, newName)
	return types.ToolResult{
		LLMContent:    output,
		ReturnDisplay: output,
	}, nil
}
