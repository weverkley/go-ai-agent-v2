package tools

import (
	"context"
	"fmt"
	"testing"

	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGitCommitTool_Execute(t *testing.T) {
	mockGitService := new(MockGitService) // Reusing MockGitService
	tool := NewGitCommitTool(mockGitService)

	tests := []struct {
		name          string
		args          map[string]any
		setupMock     func()
		expectedError string
	}{
		{
			name:          "missing message argument",
			args:          map[string]any{},
			setupMock:     func() {},
			expectedError: "missing or invalid 'message' argument",
		},
		{
			name:          "empty message argument",
			args:          map[string]any{"message": ""},
			setupMock:     func() {},
			expectedError: "missing or invalid 'message' argument",
		},
		        {
		            name: "successful commit - no files to stage",
		            args: map[string]any{"message": "Initial commit", "dir": "/test/repo"},
		            setupMock: func() {
		                				mockGitService.On("StageFiles", "/test/repo", ([]string)(nil)).Return(nil).Once()
		                				mockGitService.On("Commit", "/test/repo", "Initial commit").Return(nil).Once()			},
		},
		{
			name: "successful commit - with specific files to stage",
			args: map[string]any{"message": "Add new feature", "files_to_stage": []any{"file1.txt", "file2.go"}, "dir": "/test/repo"},
			setupMock: func() {
				mockGitService.On("StageFiles", "/test/repo", []string{"file1.txt", "file2.go"}).Return(nil).Once()
				mockGitService.On("Commit", "/test/repo", "Add new feature").Return(nil).Once()
			},
		},
		{
			name: "stage files fails",
			args: map[string]any{"message": "Initial commit", "dir": "/test/repo"},
			setupMock: func() {
				mockGitService.On("StageFiles", "/test/repo", ([]string)(nil)).Return(fmt.Errorf("stage error")).Once()
			},
			expectedError: "failed to stage files in /test/repo: stage error",
		},
		{
			name: "commit fails",
			args: map[string]any{"message": "Initial commit", "dir": "/test/repo"},
			setupMock: func() {
				mockGitService.On("StageFiles", "/test/repo", ([]string)(nil)).Return(nil).Once()
				mockGitService.On("Commit", "/test/repo", "Initial commit").Return(fmt.Errorf("commit error")).Once()
			},
			expectedError: "failed to commit changes in /test/repo: commit error",
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
				assert.Contains(t, result.LLMContent, "Successfully committed changes")
				assert.Contains(t, result.ReturnDisplay, "Successfully committed changes")
			}
			mockGitService.AssertExpectations(t)
		})
	}
}
