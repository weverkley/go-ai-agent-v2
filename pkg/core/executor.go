package core

import (
	"context"
	"go-ai-agent-v2/go-cli/pkg/types"
)

// Executor interface abstracts the behavior of different AI executors.
type Executor interface {
	Name() string // Returns the name/type of the executor (e.g., "gemini", "qwen", "mock")
	GenerateContent(contents ...*types.Content) (*types.GenerateContentResponse, error)
	GenerateContentWithTools(ctx context.Context, history []*types.Content, tools []types.Tool) (*types.GenerateContentResponse, error)
	ExecuteTool(ctx context.Context, fc *types.FunctionCall) (types.ToolResult, error)
	SendMessageStream(modelName string, messageParams types.MessageParams, promptId string) (<-chan types.StreamResponse, error)
	ListModels() ([]string, error)
	GetHistory() ([]*types.Content, error)
	SetHistory(history []*types.Content) error
	CompressChat(promptId string, force bool) (*types.ChatCompressionResult, error)
	StreamContent(ctx context.Context, contents ...*types.Content) (<-chan any, error)
	SetUserConfirmationChannel(chan bool) // New method for user confirmation
	SetToolConfirmationChannel(chan types.ToolConfirmationOutcome) // New method for rich tool confirmation
}
