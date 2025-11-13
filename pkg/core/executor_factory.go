package core

import (
	"context"
	"fmt"
	"go-ai-agent-v2/go-cli/pkg/routing"
	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/google/generative-ai-go/genai"
)

// ExecutorFactory abstracts the creation of different AI executors.
type ExecutorFactory interface {
	NewExecutor(cfg types.Config, generationConfig types.GenerateContentConfig, startHistory []*genai.Content) (Executor, error)
}

// GeminiExecutorFactory is an ExecutorFactory that creates GeminiChat instances.
type GeminiExecutorFactory struct {
	Router *routing.ModelRouterService
}

// NewExecutor creates a new GeminiChat executor.
func (f *GeminiExecutorFactory) NewExecutor(cfg types.Config, generationConfig types.GenerateContentConfig, startHistory []*genai.Content) (Executor, error) {
	// For now, we'll create a simple context. This will need to be updated
	// to use the actual chat history and user request.
	routingCtx := &routing.RoutingContext{
		History:      []string{},
		Request:      "dummy request",
		Signal:       context.Background(),
		ExecutorType: "gemini",
	}

	decision, err := f.Router.Route(routingCtx, cfg)
	if err != nil {
		return nil, fmt.Errorf("error routing model: %w", err)
	}

	routedCfg := cfg.WithModel(decision.Model)

	return NewGeminiChat(routedCfg, generationConfig, startHistory)
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

// QwenExecutorFactory is an ExecutorFactory that creates QwenChat instances.
type QwenExecutorFactory struct{}

// NewExecutor creates a new QwenChat executor.
func (f *QwenExecutorFactory) NewExecutor(cfg types.Config, generationConfig types.GenerateContentConfig, startHistory []*genai.Content) (Executor, error) {
	return NewQwenChat(cfg, generationConfig, startHistory)
}

// NewExecutorFactory creates an ExecutorFactory based on the provided type.
func NewExecutorFactory(executorType string, cfg types.Config) (ExecutorFactory, error) {
	switch executorType {
	case "gemini":
		return &GeminiExecutorFactory{
			Router: routing.NewModelRouterService(cfg),
		}, nil
	case "qwen":
		return &QwenExecutorFactory{}, nil
	case "mock":
		return &MockExecutorFactory{}, nil
	default:
		return nil, fmt.Errorf("unknown executor type: %s", executorType)
	}
}
