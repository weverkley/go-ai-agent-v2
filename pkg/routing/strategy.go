package routing

import (
	"context"
	"go-ai-agent-v2/go-cli/pkg/types" // Add this line
	"strings"
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
	IsFallback bool
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
	if !ctx.IsFallback {
		return nil, nil // Not a fallback scenario, pass to the next strategy.
	}

	currentModel, ok := cfg.Get("model")
	if !ok {
		return nil, nil // Cannot suggest if we don't know the current model.
	}

	suggestedModel, ok := getSuggestedModel(currentModel.(string))
	if !ok {
		return nil, nil // No suggestion available.
	}

	return &RoutingDecision{
		Model: suggestedModel,
		Metadata: map[string]interface{}{
			"source":    s.Name(),
			"reasoning": "Suggesting a fallback model due to an error.",
		},
	}, nil
}

func getSuggestedModel(currentModel string) (string, bool) {
	// This logic can be expanded for other executors.
	// For now, it's focused on Gemini.
	parts := strings.Split(currentModel, "/")
	modelName := parts[len(parts)-1]

	switch {
	case strings.Contains(modelName, "pro"):
		return "gemini-1.5-flash", true
	case strings.Contains(modelName, "flash"):
		return "gemini-1.5-flash-latest", true // Or some other lighter model
	default:
		return "", false
	}
}

