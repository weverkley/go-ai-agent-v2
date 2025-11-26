package tools

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/analysis" // Import the new analysis package
	"go-ai-agent-v2/go-cli/pkg/types"
)

const FIND_REFERENCES_TOOL_NAME = "find_references"

// FindReferencesTool implements the Tool interface for finding references to a symbol.
type FindReferencesTool struct {
	*types.BaseDeclarativeTool
	workspaceService types.WorkspaceServiceIface
}

// NewFindReferencesTool creates a new FindReferencesTool.
func NewFindReferencesTool(workspaceService types.WorkspaceServiceIface) *FindReferencesTool {
	return &FindReferencesTool{
		BaseDeclarativeTool: types.NewBaseDeclarativeTool(
			FIND_REFERENCES_TOOL_NAME,
			"Find References",
			"Finds all usages of a specific symbol (function, variable, etc.) in the codebase.",
			types.KindOther,
			(&types.JsonSchemaObject{
				Type: "object",
			}).SetProperties(map[string]*types.JsonSchemaProperty{
				"file_path": {
					Type:        "string",
					Description: "Path to the file containing the symbol's definition, relative to the project root.",
				},
				"line": {
					Type:        "integer",
					Description: "The line number of the symbol (1-indexed).",
				},
				"column": {
					Type:        "integer",
					Description: "The column number of the symbol (1-indexed).",
				},
			}).SetRequired([]string{"file_path", "line", "column"}),
			false, // isOutputMarkdown
			false, // canUpdateOutput
			nil,   // MessageBus
		),
		workspaceService: workspaceService,
	}
}

// Execute performs the find_references operation.
func (t *FindReferencesTool) Execute(ctx context.Context, args map[string]any) (types.ToolResult, error) {
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

	projectRoot := t.workspaceService.GetProjectRoot()
	absolutePath := filepath.Join(projectRoot, filePath)

	references, err := analysis.FindSymbolReferences(absolutePath, line, column)
	if err != nil {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: fmt.Sprintf("failed to find references: %v", err),
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("failed to find references: %w", err)
	}

	if len(references) == 0 {
		return types.ToolResult{
			LLMContent:    fmt.Sprintf("No references found for symbol at %s:%d:%d", absolutePath, line, column),
			ReturnDisplay: fmt.Sprintf("No references found for symbol at %s:%d:%d", absolutePath, line, column),
		}, nil
	}

	output := fmt.Sprintf("Found references for symbol at %s:%d:%d:\n%s", absolutePath, line, column, strings.Join(references, "\n"))
	return types.ToolResult{
		LLMContent:    output,
		ReturnDisplay: output,
	}, nil
}
