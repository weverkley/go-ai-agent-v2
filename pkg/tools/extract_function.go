package tools

import (
	"context"
	"fmt"

	"go-ai-agent-v2/go-cli/pkg/core/agents"
	"go-ai-agent-v2/go-cli/pkg/services" // Added
	"go-ai-agent-v2/go-cli/pkg/types"
)

const EXTRACT_FUNCTION_TOOL_NAME = "extract_function"

// ExtractFunctionTool is a tool that extracts a code block into a new function or method.
type ExtractFunctionTool struct {
	*types.BaseDeclarativeTool
	fileSystemService services.FileSystemService
}

// NewExtractFunctionTool creates a new ExtractFunctionTool.
func NewExtractFunctionTool(fileSystemService services.FileSystemService) *ExtractFunctionTool {
	return &ExtractFunctionTool{
		BaseDeclarativeTool: types.NewBaseDeclarativeTool(
			EXTRACT_FUNCTION_TOOL_NAME,
			"Extract Function/Method",
			"Extracts a code block into a new function or method in a Go file.",
			types.KindOther,
			&types.JsonSchemaObject{
				Type: "object",
				Properties: map[string]*types.JsonSchemaProperty{
					"filePath": &types.JsonSchemaProperty{
						Type:        "string",
						Description: "The absolute path to the Go file.",
					},
					"startLine": &types.JsonSchemaProperty{
						Type:        "integer",
						Description: "The starting line number of the code block to extract (1-indexed).",
					},
					"endLine": &types.JsonSchemaProperty{
						Type:        "integer",
						Description: "The ending line number of the code block to extract (1-indexed).",
					},
					"newFunctionName": &types.JsonSchemaProperty{
						Type:        "string",
						Description: "The name for the new function or method.",
					},
					"receiver": &types.JsonSchemaProperty{
						Type:        "string",
						Description: "Optional: The receiver for the new method (e.g., 's *MyStruct'). Leave empty for a function.",
					},
				},
				Required: []string{"filePath", "startLine", "endLine", "newFunctionName"},
			},
			false, // isOutputMarkdown
			true,  // canUpdateOutput - This tool modifies files
			nil,   // MessageBus
		),
		fileSystemService: fileSystemService,
	}
}

// Execute implements the Tool interface.
func (t *ExtractFunctionTool) Execute(ctx context.Context, args map[string]any) (types.ToolResult, error) {
	filePath, ok := args["filePath"].(string)
	if !ok {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: "missing or invalid 'filePath' argument",
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("missing or invalid 'filePath' argument")
	}

	startLineFloat, ok := args["startLine"].(float64)
	if !ok {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: "missing or invalid 'startLine' argument",
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("missing or invalid 'startLine' argument")
	}
	startLine := int(startLineFloat)

	endLineFloat, ok := args["endLine"].(float64)
	if !ok {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: "missing or invalid 'endLine' argument",
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("missing or invalid 'endLine' argument")
	}
	endLine := int(endLineFloat)

	newFunctionName, ok := args["newFunctionName"].(string)
	if !ok {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: "missing or invalid 'newFunctionName' argument",
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("missing or invalid 'newFunctionName' argument")
	}

	receiver, _ := args["receiver"].(string) // Optional, can be empty

	extracted, err := agents.ExtractFunction(filePath, startLine, endLine, newFunctionName, receiver)
	if err != nil {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: fmt.Sprintf("Failed to extract function: %v", err),
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("failed to extract function: %w", err)
	}

	err = t.fileSystemService.WriteFile(filePath, extracted.NewCode)
	if err != nil {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: fmt.Sprintf("Failed to write extracted code to file: %v", err),
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("failed to write extracted code to file: %w", err)
	}

	return types.ToolResult{
		LLMContent:    fmt.Sprintf("Successfully extracted function and wrote to file. New code:\n```go\n%s\n```", extracted.NewCode),
		ReturnDisplay: fmt.Sprintf("Successfully extracted function and wrote to file. New code:\n```go\n%s\n```", extracted.NewCode),
	}, nil
}
