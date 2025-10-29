package tools

import "github.com/google/generative-ai-go/genai"

// Tool is the interface that all tools must implement.
type Tool interface {
	// Name returns the name of the tool.
	Name() string
	// Definition returns the genai.Tool definition for the Gemini API.
	Definition() *genai.Tool
	// Execute runs the tool with the given arguments.
	Execute(args map[string]any) (string, error)
}
