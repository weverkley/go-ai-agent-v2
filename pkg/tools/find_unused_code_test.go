package tools

import (
	"context"
	"testing"

	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/stretchr/testify/assert"
)

// MockFindUnusedFunctions is a mock for agents.FindUnusedFunctions
// var MockFindUnusedFunctions func(dir string) ([]agents.UnusedFunction, error)

// func init() {
// 	// Set default implementation to the actual function
// 	MockFindUnusedFunctions = agents.FindUnusedFunctions
// }

func TestFindUnusedCodeTool_Execute(t *testing.T) {
	tool := NewFindUnusedCodeTool()

	// originalFindUnusedFunctions := agents.FindUnusedFunctions
	// t.Cleanup(func() {
	// 	agents.FindUnusedFunctions = originalFindUnusedFunctions
	// })

	tests := []struct {
		name          string
		args          map[string]any
		setupMock     func()
		expectedLLMContent string
		expectedReturnDisplay string
		expectedError string
	}{
		{
			name:          "missing dir_path argument",
			args:          map[string]any{},
			setupMock:     func() {},
			expectedError: "missing or invalid 'dir_path' argument",
		},
		{
			name:          "empty dir_path argument",
			args:          map[string]any{"dir_path": ""},
			setupMock:     func() {},
			expectedError: "missing or invalid 'dir_path' argument",
		},
		// {
		// 	name: "find unused functions fails",
		// 	args: map[string]any{"directory": "/test/dir"},
		// 	setupMock: func() {
		// 		agents.FindUnusedFunctions = func(dir string) ([]agents.UnusedFunction, error) {
		// 			return nil, fmt.Errorf("analysis error")
		// 		}
		// 	},
		// 	expectedError: "failed to find unused functions in '/test/dir': analysis error",
		// },
		// {
		// 	name: "no unused functions found",
		// 	args: map[string]any{"directory": "/test/dir"},
		// 	setupMock: func() {
		// 		agents.FindUnusedFunctions = func(dir string) ([]agents.UnusedFunction, error) {
		// 			return []agents.UnusedFunction{}, nil
		// 		}
		// 	},
		// 	expectedLLMContent:    "No unused functions or methods found in '/test/dir'.",
		// 	expectedReturnDisplay: "No unused functions or methods found in '/test/dir'.",
		// },
		// {
		// 	name: "unused functions found",
		// 	args: map[string]any{"directory": "/test/dir"},
		// 	setupMock: func() {
		// 		agents.FindUnusedFunctions = func(dir string) ([]agents.UnusedFunction, error) {
		// 			return []agents.UnusedFunction{
		// 				{Name: "UnusedFunc", Type: "func", FilePath: "/test/dir/file.go"},
		// 				{Name: "UnusedMethod", Type: "method", FilePath: "/test/dir/another.go"},
		// 			}, nil
		// 		}
		// 	},
		// 	expectedLLMContent:    "Unused functions and methods found in '/test/dir':\n- UnusedFunc (func) in /test/dir/file.go\n- UnusedMethod (method) in /test/dir/another.go\n",
		// 	expectedReturnDisplay: "Unused functions and methods found in '/test/dir':\n- UnusedFunc (func) in /test/dir/file.go\n- UnusedMethod (method) in /test/dir/another.go\n",
		// },
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