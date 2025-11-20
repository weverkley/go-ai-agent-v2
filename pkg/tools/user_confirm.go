package tools

import (
	"context"
	"fmt"

	"go-ai-agent-v2/go-cli/pkg/types"
)

// UserConfirmTool represents the user_confirm tool.
type UserConfirmTool struct {
	*types.BaseDeclarativeTool
}

// NewUserConfirmTool creates a new instance of UserConfirmTool.
func NewUserConfirmTool() *UserConfirmTool {
	return &UserConfirmTool{
		types.NewBaseDeclarativeTool(
			"user_confirm",
			"User Confirm",
			"Asks the user for confirmation with a message and provides 'continue' or 'cancel' options.",
			types.KindOther, // Assuming KindOther for now
			(&types.JsonSchemaObject{
				Type: "object",
			}).SetProperties(map[string]*types.JsonSchemaProperty{
				"message": &types.JsonSchemaProperty{
					Type:        "string",
					Description: "The message to display to the user for confirmation.",
				},
			}).SetRequired([]string{"message"}),
			false, // isOutputMarkdown
			false, // canUpdateOutput
			nil,   // MessageBus
		),
	}
}

// Execute is not intended to be called directly in an interactive flow.
// The executor should intercept the 'user_confirm' tool call and handle it via the UI.
func (t *UserConfirmTool) Execute(ctx context.Context, args map[string]any) (types.ToolResult, error) {
	err := fmt.Errorf("the 'user_confirm' tool should not be executed directly in interactive mode; it must be handled by the executor")
	return types.ToolResult{
		Error: &types.ToolError{
			Message: err.Error(),
			Type:    types.ToolErrorTypeExecutionFailed,
		},
	}, err
}
