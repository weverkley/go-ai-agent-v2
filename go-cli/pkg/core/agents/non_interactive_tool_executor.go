package agents

import (
	"context"
	"fmt"
	"time"

	"go-ai-agent-v2/go-cli/pkg/config"
)

// ExecuteToolCall executes a single tool call non-interactively.
func ExecuteToolCall(
	cfg *config.Config,
	toolCallRequest ToolCallRequestInfo,
	ctx context.Context,
) (CompletedToolCall, error) {
	// Create a channel to receive the completed tool calls.
	completedCallsChan := make(chan CompletedToolCall, 1)
	errorChan := make(chan error, 1)

	// Create a simplified CoreToolScheduler for non-interactive execution.
	// This scheduler will immediately resolve the tool call.
	scheduler := NewCoreToolScheduler(
		CoreToolSchedulerOptions{
			Config: cfg,
			OnAllToolCallsComplete: func(completedToolCalls []CompletedToolCall) {
				if len(completedToolCalls) > 0 {
					completedCallsChan <- completedToolCalls[0]
				} else {
					errorChan <- fmt.Errorf("no completed tool calls received")
				}
			},
		},
	)

	// Schedule the tool call.
	go func() {
		err := scheduler.Schedule(toolCallRequest, ctx)
		if err != nil {
			errorChan <- err
		}
	}()

	select {
	case completedCall := <-completedCallsChan:
		return completedCall, nil
	case err := <-errorChan:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(5 * time.Minute): // Add a timeout to prevent hanging
		return nil, fmt.Errorf("tool execution timed out")
	}
}
