package routing

import (
	"fmt"
	"go-ai-agent-v2/go-cli/pkg/config"
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

func (s *CompositeStrategy) Route(ctx *RoutingContext, cfg *config.Config) (*RoutingDecision, error) {
	for _, strategy := range s.strategies {
		decision, err := strategy.Route(ctx, cfg)
		if err != nil {
			// Log the error and continue to the next strategy
			fmt.Printf("routing strategy %s failed: %v\n", strategy.Name(), err)
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

func NewModelRouterService(cfg *config.Config) *ModelRouterService {
	// Initialize the composite strategy with the desired priority order.
	strategy := NewCompositeStrategy(
		&FallbackStrategy{},
		&OverrideStrategy{},
		// &ClassifierStrategy{}, // To be implemented
		&DefaultStrategy{},
	)
	return &ModelRouterService{
		strategy: strategy,
	}
}

// Route determines which model to use for a given request context.
func (s *ModelRouterService) Route(ctx *RoutingContext, cfg *config.Config) (*RoutingDecision, error) {
	return s.strategy.Route(ctx, cfg)
}
