package core

import (
	"context"
	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/google/generative-ai-go/genai"
)

// Executor interface abstracts the behavior of different AI executors.
type Executor interface {
	GenerateContent(contents ...*genai.Content) (*genai.GenerateContentResponse, error)
	GenerateContentWithTools(ctx context.Context, history []*genai.Content, tools []*genai.Tool) (*genai.GenerateContentResponse, error)
	ExecuteTool(ctx context.Context, fc *genai.FunctionCall) (types.ToolResult, error)
	SendMessageStream(modelName string, messageParams types.MessageParams, promptId string) (<-chan types.StreamResponse, error)
	ListModels() ([]string, error)
	GetHistory() ([]*genai.Content, error)
	SetHistory(history []*genai.Content) error
	CompressChat(promptId string, force bool) (*types.ChatCompressionResult, error)
	GenerateStream(ctx context.Context, contents ...*genai.Content) (<-chan any, error)
}
