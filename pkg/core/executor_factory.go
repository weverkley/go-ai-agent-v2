package core

import (
	"context"
	"fmt"
	"go-ai-agent-v2/go-cli/pkg/routing"
	"go-ai-agent-v2/go-cli/pkg/telemetry"
	"go-ai-agent-v2/go-cli/pkg/types"
)

// ExecutorFactory abstracts the creation of different AI executors.
type ExecutorFactory interface {
	NewExecutor(cfg types.Config, generationConfig types.GenerateContentConfig, startHistory []*types.Content) (Executor, error)
}

// GeminiExecutorFactory is an ExecutorFactory that creates GeminiExecutor instances.
type GeminiExecutorFactory struct {
	Router *routing.ModelRouterService
}

// NewExecutor creates a new executor.
func (f *GeminiExecutorFactory) NewExecutor(cfg types.Config, generationConfig types.GenerateContentConfig, startHistory []*types.Content) (Executor, error) {
	// For now, we'll create a simple context. This will need to be updated
	// to use the actual chat history and user request.
	routingCtx := &routing.RoutingContext{
		History:      []string{},
		Request:      "dummy request",
		Signal:       context.Background(),
		ExecutorType: types.ExecutorTypeGemini,
	}

	decision, err := f.Router.Route(routingCtx, cfg)
	if err != nil {
		return nil, fmt.Errorf("error routing model: %w", err)
	}

	routedCfg := cfg.WithModel(decision.Model)

	return NewGeminiChat(routedCfg, generationConfig, startHistory, telemetry.GlobalLogger)
}

// MockExecutorFactory is an ExecutorFactory that creates MockExecutor instances.
type MockExecutorFactory struct {
	Mock *MockExecutor
}

// NewExecutor creates a new MockExecutor.
func (f *MockExecutorFactory) NewExecutor(cfg types.Config, generationConfig types.GenerateContentConfig, startHistory []*types.Content) (Executor, error) {
	if f.Mock != nil {
		return f.Mock, nil
	}

	toolRegistryVal, ok := cfg.Get("toolRegistry")
	if !ok {
		return nil, fmt.Errorf("tool registry not found in config for mock executor")
	}
	toolRegistry, ok := toolRegistryVal.(types.ToolRegistryInterface)
	if !ok {
		return nil, fmt.Errorf("tool registry in config is not of expected type")
	}
	return NewRealisticMockExecutor(toolRegistry), nil
}

// QwenExecutorFactory is an ExecutorFactory that creates QwenChat instances.
type QwenExecutorFactory struct{}

// NewExecutor creates a new QwenChat executor.
func (f *QwenExecutorFactory) NewExecutor(cfg types.Config, generationConfig types.GenerateContentConfig, startHistory []*types.Content) (Executor, error) {
	return NewQwenChat(cfg, generationConfig, startHistory, telemetry.GlobalLogger)
}

// NewExecutorFactory creates an ExecutorFactory based on the provided type.
func NewExecutorFactory(executorType string, cfg types.Config) (ExecutorFactory, error) {
	switch executorType {
	case types.ExecutorTypeGemini:
		return &GeminiExecutorFactory{
			Router: routing.NewModelRouterService(cfg),
		}, nil
	case types.ExecutorTypeQwen:
		return &QwenExecutorFactory{}, nil
	case types.ExecutorTypeMock:
		return &MockExecutorFactory{}, nil
	default:
		return nil, fmt.Errorf("unknown executor type: %s", executorType)
	}
}
