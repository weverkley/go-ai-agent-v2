package routing

import (
	"context"
	"fmt"
	"go-ai-agent-v2/go-cli/pkg/telemetry" // Re-add telemetry import
	"go-ai-agent-v2/go-cli/pkg/types"
	"strings"
)

// RoutingDecision is the output of a routing decision.
type RoutingDecision struct {
	Model    string
	Metadata map[string]interface{}
}

// RoutingContext is the context provided to the router.
type RoutingContext struct {
	History      []string
	Request      string
	Signal       context.Context
	IsFallback   bool
	ExecutorType string
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
	modelVal, ok := cfg.Get("model")
	if !ok || modelVal == nil {
		err := fmt.Errorf("model not found in config")
		telemetry.LogErrorf(err.Error())
		return nil, err
	}
	model, ok := modelVal.(string)
	if !ok {
		err := fmt.Errorf("model in config is not a string")
		telemetry.LogErrorf(err.Error())
		return nil, err
	}
	return &RoutingDecision{
		Model: model,
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
	modelVal, ok := cfg.Get("model")
	if !ok || modelVal == nil {
		return nil, nil // Pass to the next strategy
	}
	model, ok := modelVal.(string)
	if !ok {
		err := fmt.Errorf("model in config is not a string")
		telemetry.LogErrorf(err.Error())
		return nil, err
	}
	if model == "auto" {
		return nil, nil // Pass to the next strategy
	}

	return &RoutingDecision{
		Model: model,
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
	telemetry.LogDebugf("FallbackStrategy.Route called")
	if !ctx.IsFallback {
		return nil, nil // Not a fallback scenario, pass to the next strategy.
	}

	currentModelVal, ok := cfg.Get("model")
	if !ok || currentModelVal == nil {
		err := fmt.Errorf("current model not found in config for fallback strategy")
		telemetry.LogErrorf(err.Error())
		return nil, err
	}
	currentModel, ok := currentModelVal.(string)
	if !ok {
		err := fmt.Errorf("current model in config is not a string for fallback strategy")
		telemetry.LogErrorf(err.Error())
		return nil, err
	}

	suggester, ok := modelSuggesters[ctx.ExecutorType]
	telemetry.LogDebugf("FallbackStrategy: ctx.ExecutorType=%s, suggester found=%t", ctx.ExecutorType, ok)
	if !ok {
		return nil, nil // No suggester for this executor type.
	}

	suggestedModel, ok := suggester(currentModel)
	telemetry.LogDebugf("FallbackStrategy: currentModel=%s, suggestedModel=%s, suggestion made=%t", currentModel, suggestedModel, ok)
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

var modelSuggesters = map[string]func(string) (string, bool){
	"gemini": func(currentModel string) (string, bool) {
		telemetry.LogDebugf("Gemini Suggester: currentModel=%s", currentModel)
		parts := strings.Split(currentModel, "/")
		modelName := parts[len(parts)-1]
		telemetry.LogDebugf("Gemini Suggester: modelName=%s", modelName)

		isPro := strings.Contains(modelName, "pro")
		isFlash := strings.Contains(modelName, "flash")
		telemetry.LogDebugf("Gemini Suggester: isPro=%t, isFlash=%t", isPro, isFlash)

		switch {
		case isPro:
			return "gemini-flash", true
		case isFlash:
			return "gemini-flash-lite", true
		default:
			telemetry.LogDebugf("Gemini Suggester: No suggestion found for modelName=%s", modelName)
			return "", false
		}
	},
	"qwen": func(currentModel string) (string, bool) {
		switch {
		case strings.Contains(currentModel, "max"):
			return "qwen-plus", true
		case strings.Contains(currentModel, "plus"):
			return "qwen-turbo", true
		default:
			return "", false
		}
	},
}

// ClassifierStrategy suggests a model based on the content of the request.
type ClassifierStrategy struct{}

func (s *ClassifierStrategy) Name() string {
	return "classifier"
}

func (s *ClassifierStrategy) Route(ctx *RoutingContext, cfg types.Config) (*RoutingDecision, error) {
	// Simplified heuristic: if the request contains "code", use a more powerful model.
	if strings.Contains(strings.ToLower(ctx.Request), "code") {
		// This could be made more generic, e.g., by getting the "powerful" model from config.
		return &RoutingDecision{
			Model: "gemini-1.5-pro",
			Metadata: map[string]interface{}{
				"source":    s.Name(),
				"reasoning": "Request contains 'code', suggesting a more powerful model.",
			},
		}, nil
	}

	return nil, nil // Pass to the next strategy.
}
