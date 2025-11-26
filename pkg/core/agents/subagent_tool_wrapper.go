package agents

import (
	"context"
	"fmt"

	"go-ai-agent-v2/go-cli/pkg/config"
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
func (stw *SubagentToolWrapper) CreateInvocation(params AgentInputs) types.ToolInvocation {
	return NewSubagentInvocation(
		params,
		stw.definition,
		stw.config,
		stw.BaseDeclarativeTool.MessageBus,
	)
}

// Execute is part of the types.Tool interface. It delegates to the invocation.
func (stw *SubagentToolWrapper) Execute(ctx context.Context, args map[string]interface{}) (types.ToolResult, error) {
	invocation := stw.CreateInvocation(args)
	// For now, we'll call execute with a dummy context and updateOutput.
	// The actual execution will happen within the AgentExecutor.
	result, err := invocation.Execute(ctx, nil, nil, nil)
	if err != nil {
		return types.ToolResult{}, err
	}
	return result, nil
}
