package tools

import (
	"fmt"

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
	return LS_TOOL_NAME
}

// Definition returns the genai.Tool definition for the Gemini API.
func (t *LsTool) Definition() *genai.Tool {
	// TODO: Implement the actual tool definition
	return &genai.Tool{
		FunctionDeclarations: []*genai.FunctionDeclaration{
			{
				Name:        LS_TOOL_NAME,
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
func (t *LsTool) Execute(args map[string]any) (string, error) {
	// TODO: Implement the actual tool execution logic
	return fmt.Sprintf("LsTool executed with args: %v (placeholder)", args), nil
}
