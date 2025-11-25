package tools

import (
	"context"
	"testing"

	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/stretchr/testify/assert"
)

func TestWeatherTool_Execute(t *testing.T) {
	tool := NewWeatherTool()

	tests := []struct {
		name                  string
		args                  map[string]any
		expectedLLMContent    string
		expectedReturnDisplay string
		expectedError         string
	}{
		{
			name:          "missing location argument",
			args:          map[string]any{},
			expectedError: "missing or invalid 'location' argument",
		},
		{
			name:          "empty location argument",
			args:          map[string]any{"location": ""},
			expectedError: "missing or invalid 'location' argument",
		},
		{
			name:                  "successful weather retrieval",
			args:                  map[string]any{"location": "Boston"},
			expectedLLMContent:    "The weather in Boston is 28°C, Sunny, with a wind speed of 15 km/h.",
			expectedReturnDisplay: "The weather in Boston is 28°C, Sunny, with a wind speed of 15 km/h.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tool.Execute(context.Background(), tt.args)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				if result.Error != nil {
					assert.Equal(t, types.ToolErrorTypeExecutionFailed, result.Error.Type)
				} else {
					assert.Fail(t, "Expected ToolResult.Error to be non-nil when an error is expected")
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedLLMContent, result.LLMContent)
				assert.Equal(t, tt.expectedReturnDisplay, result.ReturnDisplay)
			}
		})
	}
}
