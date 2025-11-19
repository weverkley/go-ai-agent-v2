package tools

import (
	"context"
	"fmt"
	"testing"

	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPullTool_Execute(t *testing.T) {
	mockGitService := new(MockGitService) // Reusing MockGitService from checkout_branch_test.go
	tool := NewPullTool(mockGitService)

	tests := []struct {
		name          string
		args          map[string]any
		setupMock     func()
		expectedLLMContent string
		expectedReturnDisplay string
		expectedError string
	}{
		{
			name:          "missing dir argument",
			args:          map[string]any{},
			setupMock:     func() {},
			expectedError: "missing or invalid 'dir' argument",
		},
		{
			name:          "empty dir argument",
			args:          map[string]any{"dir": ""},
			setupMock:     func() {},
			expectedError: "missing or invalid 'dir' argument",
		},
		{
			name: "successful pull",
			args: map[string]any{"dir": "/test/repo"},
			setupMock: func() {
				mockGitService.On("Pull", "/test/repo", "").Return(nil).Once()
			},
			expectedLLMContent:    "Successfully pulled latest changes in /test/repo.",
			expectedReturnDisplay: "Successfully pulled latest changes in /test/repo.",
		},
		{
			name: "pull fails",
			args: map[string]any{"dir": "/test/repo"},
			setupMock: func() {
				mockGitService.On("Pull", "/test/repo", "").Return(fmt.Errorf("git error")).Once()
			},
			expectedError: "failed to pull changes in /test/repo: git error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGitService.Calls = []mock.Call{} // Reset calls for each test
			tt.setupMock()

			result, err := tool.Execute(context.Background(), tt.args)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				if result.Error != nil {
					assert.Equal(t, types.ToolErrorTypeExecutionFailed, result.Error.Type)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedLLMContent, result.LLMContent)
				assert.Equal(t, tt.expectedReturnDisplay, result.ReturnDisplay)
			}
			mockGitService.AssertExpectations(t)
		})
	}
}
