package tools

import (
	"context"
	"fmt"

	"go-ai-agent-v2/go-cli/pkg/types"
)

// taskCompleteTool is a special tool used by agents to signal task completion.
type taskCompleteTool struct {
	*types.BaseDeclarativeTool
}

// NewTaskCompleteTool creates a new instance of the taskCompleteTool.
func NewTaskCompleteTool() types.Tool {
	return &taskCompleteTool{
		BaseDeclarativeTool: types.NewBaseDeclarativeTool(
			types.TASK_COMPLETE_TOOL_NAME,
			"Task Complete",
			"Signals that the agent has successfully completed its task. This is the ONLY way to finish an agent's execution.",
			types.KindOther,
			&types.JsonSchemaObject{
				Type:       "object",
				Properties: map[string]*types.JsonSchemaProperty{},
				Required:   []string{},
			},
			false,
			false,
			nil,
		),
	}
}

// Execute handles the execution of the taskCompleteTool.
func (t *taskCompleteTool) Execute(ctx context.Context, args map[string]any) (types.ToolResult, error) {
	// For logging and debugging, we can log the report if provided.
	report, ok := args["report"].(string)
	if !ok {
		report = "Task completed without a specific report."
	}
	return types.ToolResult{
		LLMContent:    fmt.Sprintf("Task completed: %s", report),
		ReturnDisplay: fmt.Sprintf("Task completed: %s", report),
	}, nil
}
