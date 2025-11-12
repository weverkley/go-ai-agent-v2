package routing

import (
	"context"
	"go-ai-agent-v2/go-cli/pkg/types" // Add this line
)

// RoutingDecision is the output of a routing decision.
type RoutingDecision struct {
	Model    string
	Metadata map[string]interface{}
}

// RoutingContext is the context provided to the router.
type RoutingContext struct {
	History    []string
	Request    string
	Signal     context.Context
}

// RoutingStrategy is the interface for all routing strategies.
type RoutingStrategy interface {
	Name() string
	Route(ctx *RoutingContext, cfg types.Config) (*RoutingDecision, error)
}

// TerminalStrategy is a strategy that is guaranteed to return a decision.
type TerminalStrategy interface {
	RoutingStrategy
}

// DefaultStrategy is the default routing strategy.
type DefaultStrategy struct{}

func (s *DefaultStrategy) Name() string {
	return "default"
}

func (s *DefaultStrategy) Route(ctx *RoutingContext, cfg types.Config) (*RoutingDecision, error) {
	model, _ := cfg.Get("model")
	return &RoutingDecision{
		Model: model.(string),
		Metadata: map[string]interface{}{
			"source":    s.Name(),
			"reasoning": "Default model selected",
		},
	}, nil
}

// OverrideStrategy handles cases where the user explicitly specifies a model.
type OverrideStrategy struct{}

func (s *OverrideStrategy) Name() string {
	return "override"
}

func (s *OverrideStrategy) Route(ctx *RoutingContext, cfg types.Config) (*RoutingDecision, error) {
	model, ok := cfg.Get("model")
	if !ok || model.(string) == "auto" {
		return nil, nil // Pass to the next strategy
	}

	return &RoutingDecision{
		Model: model.(string),
		Metadata: map[string]interface{}{
			"source":    s.Name(),
			"reasoning": "Model overridden by user",
		},
	}, nil
}

// FallbackStrategy handles cases where the application is in fallback mode.
type FallbackStrategy struct{}

func (s *FallbackStrategy) Name() string {
	return "fallback"
}

func (s *FallbackStrategy) Route(ctx *RoutingContext, cfg types.Config) (*RoutingDecision, error) {
	// TODO: Implement fallback logic. For now, we'll just pass.
	return nil, nil
}
