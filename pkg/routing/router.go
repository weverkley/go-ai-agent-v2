package routing

import (
	"fmt"
	"go-ai-agent-v2/go-cli/pkg/telemetry"
	"go-ai-agent-v2/go-cli/pkg/types"
)

// CompositeStrategy attempts a list of child strategies in order.
type CompositeStrategy struct {
	strategies []RoutingStrategy
}

func NewCompositeStrategy(strategies ...RoutingStrategy) *CompositeStrategy {
	return &CompositeStrategy{
		strategies: strategies,
	}
}

func (s *CompositeStrategy) Name() string {
	return "composite"
}

func (s *CompositeStrategy) Route(ctx *RoutingContext, cfg types.Config) (*RoutingDecision, error) {
	for _, strategy := range s.strategies {
		decision, err := strategy.Route(ctx, cfg)
		if err != nil {
			// Log the error and continue to the next strategy
			// telemetry.LogDebugf("Routing strategy %s failed: %v", strategy.Name(), err) // Assuming telemetry is available
			continue
		}
		if decision != nil {
			return decision, nil
		}
	}
	return nil, fmt.Errorf("no routing strategy made a decision")
}

// ModelRouterService is a centralized service for making model routing decisions.
type ModelRouterService struct {
	strategy RoutingStrategy
}

func NewModelRouterService(cfg types.Config) *ModelRouterService {
	telemetry.LogDebugf("NewModelRouterService called: Initializing routing strategies.")
	strategies := []RoutingStrategy{
		&FallbackStrategy{},
		&OverrideStrategy{},
		&ClassifierStrategy{},
		&DefaultStrategy{},
	}
	for i, s := range strategies {
		telemetry.LogDebugf("Strategy %d: %s", i, s.Name())
	}
	strategy := NewCompositeStrategy(strategies...)
	return &ModelRouterService{
		strategy: strategy,
	}
}

// Route determines which model to use for a given request context.
func (s *ModelRouterService) Route(ctx *RoutingContext, cfg types.Config) (*RoutingDecision, error) {
	return s.strategy.Route(ctx, cfg)
}
