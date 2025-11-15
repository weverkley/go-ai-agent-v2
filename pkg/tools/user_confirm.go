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
			types.JsonSchemaObject{
				Type: "object",
				Properties: map[string]types.JsonSchemaProperty{
					"message": {
						Type:        "string",
						Description: "The message to display to the user for confirmation.",
					},
				},
				Required: []string{"message"},
			},
			false, // isOutputMarkdown
			false, // canUpdateOutput
			nil,   // MessageBus
		),
	}
}

// Execute performs the user_confirm operation.
func (t *UserConfirmTool) Execute(ctx context.Context, args map[string]any) (types.ToolResult, error) {
	message, ok := args["message"].(string)
	if !ok || message == "" {
		return types.ToolResult{}, fmt.Errorf("missing or invalid 'message' argument")
	}

	// In a real CLI, this would interact with the user.
	// For the mock executor, we'll simulate a "continue" response.
	// This will be replaced with actual UI interaction later.
	fmt.Printf("\n--- User Confirmation Required ---\n")
	fmt.Printf("Message: %s\n", message)
	fmt.Printf("Please run the command in another terminal and then choose an option.\n")
	fmt.Printf("Options: [continue, cancel]\n")
	fmt.Printf("----------------------------------\n")

	// Simulate user input for now. In a real scenario, this would block and wait for user input.
	// For the mock, we'll just assume continue.
	return types.ToolResult{
		LLMContent:    "continue",
		ReturnDisplay: "User chose to continue.",
	}, nil
}
