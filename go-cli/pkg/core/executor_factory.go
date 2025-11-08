package core

import (
	"fmt"

	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/google/generative-ai-go/genai"
)

// ExecutorFactory creates Executor instances.
type ExecutorFactory struct{}

// NewExecutorFactory creates a new ExecutorFactory.
func NewExecutorFactory() *ExecutorFactory {
	return &ExecutorFactory{}
}

// CreateExecutor creates an Executor based on the provided type.
func (ef *ExecutorFactory) CreateExecutor(executorType string, cfg types.GeminiConfigProvider, generationConfig types.GenerateContentConfig, startHistory []*genai.Content) (Executor, error) {
	switch executorType {
	case "gemini":
		return NewGeminiChat(cfg, generationConfig, startHistory)
	case "mock":
		return NewMockExecutor(nil, nil), nil
	// case "openai":
	// 	return NewOpenAIExecutor(cfg, generationConfig, startHistory)
	default:
		return nil, fmt.Errorf("unsupported executor type: %s", executorType)
	}
}
