package core

import (
	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/google/generative-ai-go/genai"
)

// Executor interface abstracts the behavior of different AI executors.
type Executor interface {
	GenerateContent(contents ...*genai.Content) (*genai.GenerateContentResponse, error)
	ExecuteTool(fc *genai.FunctionCall) (types.ToolResult, error)
	SendMessageStream(modelName string, messageParams types.MessageParams, promptId string) (<-chan types.StreamResponse, error)
	ListModels() ([]string, error)
}
