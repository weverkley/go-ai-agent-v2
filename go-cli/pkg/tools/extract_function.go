package tools

import (
	"fmt"
	"strconv"

	"go-ai-agent-v2/go-cli/pkg/core/agents"
	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/google/generative-ai-go/genai"
)

const EXTRACT_FUNCTION_TOOL_NAME = "extract_function"

// ExtractFunctionTool is a tool that extracts a code block into a new function or method.
type ExtractFunctionTool struct {
	*types.BaseDeclarativeTool
}

// NewExtractFunctionTool creates a new ExtractFunctionTool.
func NewExtractFunctionTool() *ExtractFunctionTool {
	return &ExtractFunctionTool{
		BaseDeclarativeTool: types.NewBaseDeclarativeTool(
			EXTRACT_FUNCTION_TOOL_NAME,
			"Extract Function/Method",
			"Extracts a code block into a new function or method in a Go file.",
			types.KindOther,
			types.JsonSchemaObject{
				Type: "object",
				Properties: map[string]types.JsonSchemaProperty{
					"filePath": {
						Type:        "string",
						Description: "The absolute path to the Go file.",
					},
					"startLine": {
						Type:        "integer",
						Description: "The starting line number of the code block to extract (1-indexed).",
					},
					"endLine": {
						Type:        "integer",
						Description: "The ending line number of the code block to extract (1-indexed).",
					},
					"newFunctionName": {
						Type:        "string",
						Description: "The name for the new function or method.",
					},
					"receiver": {
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
	}
}

// Execute implements the Tool interface.
func (t *ExtractFunctionTool) Execute(args map[string]any) (types.ToolResult, error) {
	filePath, ok := args["filePath"].(string)
	if !ok {
		return types.ToolResult{}, fmt.Errorf("missing or invalid 'filePath' argument")
	}

	startLineStr, ok := args["startLine"].(string)
	if !ok {
		return types.ToolResult{}, fmt.Errorf("missing or invalid 'startLine' argument")
	}
	startLine, err := strconv.Atoi(startLineStr)
	if err != nil {
		return types.ToolResult{}, fmt.Errorf("invalid 'startLine' argument: %w", err)
	}

	endLineStr, ok := args["endLine"].(string)
	if !ok {
		return types.ToolResult{}, fmt.Errorf("missing or invalid 'endLine' argument")
	}
	endLine, err := strconv.Atoi(endLineStr)
	if err != nil {
		return types.ToolResult{}, fmt.Errorf("invalid 'endLine' argument: %w", err)
	}

	newFunctionName, ok := args["newFunctionName"].(string)
	if !ok {
		return types.ToolResult{}, fmt.Errorf("missing or invalid 'newFunctionName' argument")
	}

	receiver, _ := args["receiver"].(string) // Optional, can be empty

	extracted, err := agents.ExtractFunction(filePath, startLine, endLine, newFunctionName, receiver)
	if err != nil {
		return types.ToolResult{}, fmt.Errorf("failed to extract function: %w", err)
	}

	// For now, we just return the new code. A real implementation would write to file.
	return types.ToolResult{
		LLMContent:    extracted.NewCode,
		ReturnDisplay: fmt.Sprintf("Successfully extracted function. New code:\n```go\n%s\n```", extracted.NewCode),
	}, nil
}

// Definition returns the Gemini tool definition.
func (t *ExtractFunctionTool) Definition() *genai.Tool {
	return t.BaseDeclarativeTool.Definition()
}
