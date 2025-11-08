package tools

import (
	"fmt"
	"os"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/google/generative-ai-go/genai"
)

// LsTool is a placeholder for the ls tool.
type LsTool struct{}

// NewLsTool creates a new instance of LsTool.
func NewLsTool() *LsTool {
	return &LsTool{}
}

// Name returns the name of the tool.
func (t *LsTool) Name() string {
	return types.LS_TOOL_NAME
}

// Definition returns the genai.Tool definition for the Gemini API.
func (t *LsTool) Definition() *genai.Tool {
	return &genai.Tool{
		FunctionDeclarations: []*genai.FunctionDeclaration{
			{
				Name:        types.LS_TOOL_NAME,
				Description: "Lists the names of files and subdirectories directly within a specified directory path.",
				Parameters: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"path": {
							Type:        genai.TypeString,
							Description: "The absolute path to the directory to list (must be absolute, not relative)",
						},
					},
					Required: []string{"path"},
				},
			},
		},
	}
}

// Execute runs the tool with the given arguments.
func (t *LsTool) Execute(args map[string]any) (types.ToolResult, error) {
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
