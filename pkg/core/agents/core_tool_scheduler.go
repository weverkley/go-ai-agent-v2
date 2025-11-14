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

	outputUpdateHandler    func(toolCallID string, outputChunk string)
	onAllToolCallsComplete     func(completedToolCalls []CompletedToolCall)
	onToolCallsUpdate          func(toolCalls []ToolCall)
	getPreferredEditor         func() types.EditorType
	onEditorClose              func()
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
	toolRegistryVal, found := s.config.Get("toolRegistry")
	if !found || toolRegistryVal == nil {
		return fmt.Errorf("tool registry not found in config")
	}
	toolRegistry, ok := toolRegistryVal.(types.ToolRegistryInterface)
	if !ok {
		return fmt.Errorf("tool registry in config is not of expected type")
	}

	toolInstance, err := toolRegistry.GetTool(request.Name)
	if err != nil {
		return fmt.Errorf("tool \"%s\" not found in registry", request.Name)
	}

	startTime := time.Now()
	result, err := toolInstance.Execute(ctx, request.Args)
	durationMs := time.Since(startTime).Milliseconds()

	var completedCall CompletedToolCall

	if err != nil {
		// Even with an error, there might be partial results to display.
		// We create an ErroredToolCall.
		erroredCall := &ErroredToolCall{
			BaseToolCall: BaseToolCall{
				Request:    request,
				Tool:       nil, // Simplified, as we don't have AnyDeclarativeTool here
				Invocation: nil, // Simplified
				StartTime:  &startTime,
				Outcome:    types.ToolConfirmationOutcomeProceedAlways,
			},
			Response: types.ToolCallResponseInfo{
				CallID: request.CallID,
				Error:  err,
			},
			DurationMs: &durationMs,
		}
		if result.Error != nil {
			erroredCall.Response.ErrorType = result.Error.Type
		}
		if result.LLMContent != nil {
			if parts, ok := result.LLMContent.([]types.Part); ok {
				erroredCall.Response.ResponseParts = parts
			} else if text, ok := result.LLMContent.(string); ok {
				erroredCall.Response.ResponseParts = []types.Part{{Text: text}}
			}
		}
		completedCall = erroredCall
	} else {
		// Successful execution
		successfulCall := &SuccessfulToolCall{
			BaseToolCall: BaseToolCall{
				Request:    request,
				Tool:       nil, // Simplified
				Invocation: nil, // Simplified
				StartTime:  &startTime,
				Outcome:    types.ToolConfirmationOutcomeProceedAlways,
			},
			DurationMs: &durationMs,
		}
		if result.LLMContent != nil {
			if parts, ok := result.LLMContent.([]types.Part); ok {
				successfulCall.Response.ResponseParts = parts
			} else if text, ok := result.LLMContent.(string); ok {
				successfulCall.Response.ResponseParts = []types.Part{{Text: text}}
			}
		}
		if result.ReturnDisplay != "" {
			successfulCall.Response.ResultDisplay = &types.ToolResultDisplay{
				FileDiff: result.ReturnDisplay, // Assuming display is a diff for now
			}
			successfulCall.Response.ContentLength = len(result.ReturnDisplay)
		}
		completedCall = successfulCall
	}

	if s.onAllToolCallsComplete != nil {
		s.onAllToolCallsComplete([]CompletedToolCall{completedCall})
	}

	return nil
}
