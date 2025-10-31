package agents

import (
	"context"
	"fmt"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/types"
)

const (
	INPUT_PREVIEW_MAX_LENGTH = 50
	DESCRIPTION_MAX_LENGTH   = 200
)

// NewSubagentInvocation creates a new SubagentInvocation instance.
func NewSubagentInvocation(
	params AgentInputs,
	definition AgentDefinition,
	cfg *config.Config,
	messageBus interface{}, // Placeholder for MessageBus
) *SubagentInvocation {
	return &SubagentInvocation{
		BaseToolInvocation: BaseToolInvocation{
			Params:     params,
			MessageBus: messageBus,
		},
		Definition: definition,
		Config:     cfg,
	}
}

// GetDescription returns a concise, human-readable description of the invocation.
func (si *SubagentInvocation) GetDescription() string {
	inputSummaries := []string{}
	for key, value := range si.Params {
		strValue := fmt.Sprintf("%v", value)
		if len(strValue) > INPUT_PREVIEW_MAX_LENGTH {
			strValue = strValue[:INPUT_PREVIEW_MAX_LENGTH] + "..."
		}
		inputSummaries = append(inputSummaries, fmt.Sprintf("%s: %s", key, strValue))
	}

	description := fmt.Sprintf("Running subagent '%s' with inputs: { %s }", si.Definition.Name, strings.Join(inputSummaries, ", "))
	if len(description) > DESCRIPTION_MAX_LENGTH {
		description = description[:DESCRIPTION_MAX_LENGTH] + "..."
	}
	return description
}

// Execute executes the subagent.
func (si *SubagentInvocation) Execute(
	ctx context.Context,
	updateOutput func(output string), // Simplified from string | AnsiOutput
	shellExecutionConfig interface{}, // Not used for subagents, but part of AnyToolInvocation interface
	setPidCallback func(int), // Not used for subagents, but part of AnyToolInvocation interface
) (types.ToolResult, error) {
	if updateOutput != nil {
		updateOutput("Subagent starting...\n")
	}

	// Create an activity callback to bridge the executor's events to the
	// tool's streaming output.
	onActivity := func(activity SubagentActivityEvent) {
		if updateOutput == nil {
			return
		}

		if activity.Type == "THOUGHT_CHUNK" {
			if text, ok := activity.Data["text"].(string); ok {
				updateOutput(fmt.Sprintf("🤖💭 %s", text))
			}
		}
	}

	executor, err := CreateAgentExecutor(
		si.Definition,
		si.Config,
		si.Config.GetToolRegistry(), // Pass the main tool registry for subagent to discover tools
		"",                          // parentPromptId - SubagentInvocation is top-level for its own execution
		onActivity,
	)
	if err != nil {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: fmt.Sprintf("Failed to create agent executor: %v", err),
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, err
	}

	output, err := executor.Run(si.Params, ctx)
	if err != nil {
		errorMessage := err.Error()
		return types.ToolResult{
			LLMContent:    fmt.Sprintf("Subagent '%s' failed. Error: %s", si.Definition.Name, errorMessage),
			ReturnDisplay: fmt.Sprintf("Subagent Failed: %s\nError: %s", si.Definition.Name, errorMessage),
			Error: &types.ToolError{
				Message: errorMessage,
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, err
	}

	resultContent := fmt.Sprintf("Subagent '%s' finished.\nTermination Reason: %s\nResult:\n%s",
		si.Definition.Name, output.TerminateReason, output.Result)

	displayContent := fmt.Sprintf("\nSubagent %s Finished\n\nTermination Reason:\n %s\n\nResult:\n%s\n",
		si.Definition.Name, output.TerminateReason, output.Result)

	return types.ToolResult{
		LLMContent:    []types.Part{{Text: resultContent}}, // Assuming LLMContent can be []Part
		ReturnDisplay: displayContent,
	}, nil
}

// ShouldConfirmExecute always returns false for subagents as they are non-interactive.
func (si *SubagentInvocation) ShouldConfirmExecute(ctx context.Context) (types.ToolCallConfirmationDetails, error) {
	return types.ToolCallConfirmationDetails{}, nil
}
