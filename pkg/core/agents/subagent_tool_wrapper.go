package agents

import (
	"context"
	"fmt"

	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/services"
	"go-ai-agent-v2/go-cli/pkg/types"
)

// SubagentToolWrapper is a tool wrapper that dynamically exposes a subagent as a standard,
// strongly-typed DeclarativeTool.
type SubagentToolWrapper struct {
	*types.BaseDeclarativeTool
	definition AgentDefinition
	config     *config.Config
}

// NewSubagentToolWrapper constructs the tool wrapper.
func NewSubagentToolWrapper(
	definition AgentDefinition,
	cfg *config.Config,
	messageBus interface{}, // Placeholder for MessageBus
) (*SubagentToolWrapper, error) {
	// Dynamically generate the JSON schema required for the tool definition.
	parameterSchema, err := convertInputConfigToJsonSchema(definition.InputConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to convert input config to JSON schema: %w", err)
	}

	baseTool := types.NewBaseDeclarativeTool(
		definition.Name,
		definition.DisplayName,
		definition.Description,
		types.KindOther, // Assuming KindOther for now
		parameterSchema,
		false,      // isOutputMarkdown
		false,      // canUpdateOutput
		messageBus, // MessageBus
	)

	return &SubagentToolWrapper{
			BaseDeclarativeTool: baseTool,
			definition:          definition,
			config:              cfg,
		},
		nil
}

// CreateInvocation creates an invocation instance for executing the subagent.
func (stw *SubagentToolWrapper) CreateInvocation(params AgentInputs, activityChan chan types.SubagentActivityEvent) types.ToolInvocation {
	return NewSubagentInvocation(
		params,
		stw.definition,
		stw.config,
		stw.BaseDeclarativeTool.MessageBus,
		activityChan,
	)
}

// Execute is part of the types.Tool interface. It delegates to the invocation.
func (stw *SubagentToolWrapper) Execute(ctx context.Context, args map[string]interface{}) (types.ToolResult, error) {
	eventChan, ok := ctx.Value(services.EventChanKey).(chan any)
	if !ok {
		// If the channel is not in the context, we can't stream activities.
		// We can proceed without it, but the UI won't show sub-agent activity.
		eventChan = nil
	}

	activityChan := make(chan types.SubagentActivityEvent) // Changed type here
	invocation := stw.CreateInvocation(args, activityChan)

	// Start a goroutine to listen for activity and forward it to the main event channel
	if eventChan != nil {
		go func() {
			for activity := range activityChan {
				eventChan <- activity
			}
		}()
	}

	// For now, we'll call execute with a dummy context and updateOutput.
	// The actual execution will happen within the AgentExecutor.
	result, err := invocation.Execute(ctx, nil, nil, nil)
	if err != nil {
		return result, err
	}
	return result, nil
}
