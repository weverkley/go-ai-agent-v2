package tools

import (
	"context"
	"testing"

	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/stretchr/testify/assert"
)

func TestUserConfirmTool_Execute(t *testing.T) {
	tool := NewUserConfirmTool()

	tests := []struct {
		name          string
		args          map[string]any
		expectedLLMContent string
		expectedReturnDisplay string
		expectedError string
	}{
		{
			name:          "missing message argument",
			args:          map[string]any{},
			expectedError: "missing or invalid 'message' argument",
		},
		{
			name:          "empty message argument",
			args:          map[string]any{"message": ""},
			expectedError: "missing or invalid 'message' argument",
		},
		{
			name: "successful confirmation",
			args: map[string]any{"message": "Do you want to continue?"},
			expectedLLMContent:    "continue",
			expectedReturnDisplay: "User confirmation requested: Do you want to continue?. (Simulated 'continue' for mock executor)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tool.Execute(context.Background(), tt.args)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Equal(t, types.ToolErrorTypeExecutionFailed, result.Error.Type)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedLLMContent, result.LLMContent)
				assert.Equal(t, tt.expectedReturnDisplay, result.ReturnDisplay)
			}
		})
	}
}
