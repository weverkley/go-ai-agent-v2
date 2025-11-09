package core

import (
	"fmt"
	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/google/generative-ai-go/genai"
)

// ExecutorFactory abstracts the creation of different AI executors.
type ExecutorFactory interface {
	NewExecutor(cfg types.Config, generationConfig types.GenerateContentConfig, startHistory []*genai.Content) (Executor, error)
}

// GeminiExecutorFactory is an ExecutorFactory that creates GeminiChat instances.
type GeminiExecutorFactory struct{}

// NewExecutor creates a new GeminiChat executor.
func (f *GeminiExecutorFactory) NewExecutor(cfg types.Config, generationConfig types.GenerateContentConfig, startHistory []*genai.Content) (Executor, error) {
	return NewGeminiChat(cfg, generationConfig, startHistory)
}

// MockExecutorFactory is an ExecutorFactory that creates MockExecutor instances.
type MockExecutorFactory struct {
	Mock *MockExecutor
}

// NewExecutor creates a new MockExecutor.
func (f *MockExecutorFactory) NewExecutor(cfg types.Config, generationConfig types.GenerateContentConfig, startHistory []*genai.Content) (Executor, error) {
	if f.Mock != nil {
		return f.Mock, nil
	}
	return &MockExecutor{}, nil
}

// NewExecutorFactory creates an ExecutorFactory based on the provided type.
func NewExecutorFactory(executorType string) (ExecutorFactory, error) {
	switch executorType {
	case "gemini":
		return &GeminiExecutorFactory{}, nil
	case "mock":
		return &MockExecutorFactory{}, nil
	default:
		return nil, fmt.Errorf("unknown executor type: %s", executorType)
	}
}
