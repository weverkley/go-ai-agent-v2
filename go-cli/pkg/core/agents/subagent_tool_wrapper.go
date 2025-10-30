package agents

import (
	"context"
	"fmt"

	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/tools"
)

// SubagentToolWrapper is a tool wrapper that dynamically exposes a subagent as a standard,
// strongly-typed DeclarativeTool.
type SubagentToolWrapper struct {
	*tools.BaseDeclarativeTool
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

	baseTool := tools.NewBaseDeclarativeTool(
		definition.Name,
		definition.DisplayName,
		definition.Description,
		tools.KindThink, // Assuming subagents are "THINK" kind
		parameterSchema,
		true,        // isOutputMarkdown
		true,        // canUpdateOutput
		messageBus,
	)

	return &SubagentToolWrapper{
		BaseDeclarativeTool: baseTool,
		definition:          definition,
		config:              cfg,
	},
		nil
}

// CreateInvocation creates an invocation instance for executing the subagent.
func (stw *SubagentToolWrapper) CreateInvocation(params AgentInputs) tools.ToolInvocation {
	return NewSubagentInvocation(
		params,
		stw.definition,
		stw.config,
		stw.messageBus,
	)
}

// Execute is part of the tools.Tool interface. It delegates to the invocation.
func (stw *SubagentToolWrapper) Execute(args map[string]interface{}) (ToolResult, error) {
	invoCation := stw.CreateInvocation(args)
	// For now, we'll call execute with a dummy context and updateOutput.
	// The actual execution will happen within the AgentExecutor.
	result, err := invoCation.Execute(context.Background(), nil, nil, nil)
	if err != nil {
		return ToolResult{}, err
	}
	return result, nil
}
