package tools

import (
	"fmt"
	"path/filepath"

	"go-ai-agent-v2/go-cli/pkg/core/agents"
	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/google/generative-ai-go/genai"
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
			FIND_UNUSED_CODE_TOOL_NAME,
			"Find Unused Code",
			"Finds unused functions and methods in a given directory.",
			types.KindOther,
			types.JsonSchemaObject{
				Type: "object",
				Properties: map[string]types.JsonSchemaProperty{
					"directory": {
						Type:        "string",
						Description: "The absolute path to the directory to scan for unused code.",
					},
				},
				Required: []string{"directory"},
			},
			false, // isOutputMarkdown
			false, // canUpdateOutput
			nil,   // MessageBus
		),
	}
}

// Execute implements the Tool interface.
func (t *FindUnusedCodeTool) Execute(args map[string]any) (types.ToolResult, error) {
	directory, ok := args["directory"].(string)
	if !ok {
		return types.ToolResult{}, fmt.Errorf("missing or invalid 'directory' argument")
	}

	// Resolve the absolute path
	absPath, err := filepath.Abs(directory)
	if err != nil {
		return types.ToolResult{}, fmt.Errorf("failed to resolve absolute path for directory '%s': %w", directory, err)
	}

	unusedFunctions, err := agents.FindUnusedFunctions(absPath)
	if err != nil {
		return types.ToolResult{}, fmt.Errorf("failed to find unused functions in '%s': %w", absPath, err)
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

// Definition returns the Gemini tool definition.
func (t *FindUnusedCodeTool) Definition() *genai.Tool {
	return t.BaseDeclarativeTool.Definition()
}
