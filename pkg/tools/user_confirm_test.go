package tools

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserConfirmTool_Execute(t *testing.T) {
	tool := NewUserConfirmTool()

	// The user_confirm tool should not be executed directly.
	// It is a special tool handled by the ChatService to manage the interactive flow.
	// Therefore, its Execute method should always return an error.
	expectedError := "the 'user_confirm' tool should not be executed directly in interactive mode; it must be handled by the executor"

	args := map[string]any{"message": "Do you want to continue?"}
	_, err := tool.Execute(context.Background(), args)

	assert.Error(t, err)
	assert.EqualError(t, err, expectedError)
}
