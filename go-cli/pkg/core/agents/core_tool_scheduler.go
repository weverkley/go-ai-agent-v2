package agents

import (
	"context"
	"fmt"
	"time"

	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/types"
)

// CoreToolSchedulerOptions configures the CoreToolScheduler.
type CoreToolSchedulerOptions struct {
	Config                 *config.Config
	OutputUpdateHandler    func(toolCallID string, outputChunk string)
	OnAllToolCallsComplete func(completedToolCalls []CompletedToolCall)
	OnToolCallsUpdate      func(toolCalls []ToolCall)
	GetPreferredEditor     func() types.EditorType
	OnEditorClose          func()
}

// CoreToolScheduler manages the lifecycle of tool calls.
type CoreToolScheduler struct {
	config *config.Config

	toolCalls                  []ToolCall
	outputUpdateHandler        func(toolCallID string, outputChunk string)
	onAllToolCallsComplete     func(completedToolCalls []CompletedToolCall)
	onToolCallsUpdate          func(toolCalls []ToolCall)
	getPreferredEditor         func() types.EditorType
	onEditorClose              func()
	isFinalizingToolCalls      bool
	isScheduling               bool
	isCancelling               bool
	requestQueue               []*schedulerRequest
	toolCallQueue              []ToolCall
	completedToolCallsForBatch []CompletedToolCall
}

// schedulerRequest represents a request in the scheduler's queue.
type schedulerRequest struct {
	Request types.ToolCallRequestInfo
	Context context.Context
	Resolve func()
	Reject  func(error)
}

// NewCoreToolScheduler creates a new CoreToolScheduler instance.
func NewCoreToolScheduler(options CoreToolSchedulerOptions) *CoreToolScheduler {
	return &CoreToolScheduler{
		config:                 options.Config,
		outputUpdateHandler:    options.OutputUpdateHandler,
		onAllToolCallsComplete: options.OnAllToolCallsComplete,
		onToolCallsUpdate:      options.OnToolCallsUpdate,
		getPreferredEditor:     options.GetPreferredEditor,
		onEditorClose:          options.OnEditorClose,
	}
}

// Schedule schedules a tool call for execution.
func (s *CoreToolScheduler) Schedule(
	request types.ToolCallRequestInfo,
	ctx context.Context,
) error {
	// For now, a simplified implementation that directly processes the request.
	// The full queuing and state management will be added later.

	toolInstance, err := s.config.GetToolRegistry().GetTool(request.Name)
	if err != nil {
		return fmt.Errorf("tool \"%s\" not found in registry", request.Name)
	}

	// Assuming AnyDeclarativeTool has a Build method that returns AnyToolInvocation
	invocation, err := toolInstance.(AnyDeclarativeTool).Build(request.Args)
	if err != nil {
		return fmt.Errorf("failed to build tool invocation: %w", err)
	}

	// Simulate execution and completion
	startTime := time.Now()
	// In a real scenario, this would involve calling invocation.Execute()
	// and handling its output and potential errors.
	// For now, we'll create a dummy successful completion.
	

	durationMs := time.Since(startTime).Milliseconds()

	completedCall := &SuccessfulToolCall{
		BaseToolCall: BaseToolCall{
			Request:    request,
			Tool:       toolInstance.(AnyDeclarativeTool),
			Invocation: invocation,
			StartTime:  &startTime,
			Outcome:    types.ToolConfirmationOutcomeProceedAlways,
		},
		Response: types.ToolCallResponseInfo{
			CallID:        request.CallID,
			ResponseParts: []types.Part{{Text: "Tool executed successfully (dummy)."}},
			ResultDisplay: &types.ToolResultDisplay{FileDiff: "dummy"},
			ContentLength: len("Tool executed successfully (dummy)."),
		},
		DurationMs: &durationMs,
	}

	if s.onAllToolCallsComplete != nil {
		s.onAllToolCallsComplete([]CompletedToolCall{completedCall})
	}

	return nil
}
