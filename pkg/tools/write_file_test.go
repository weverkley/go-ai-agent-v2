package tools

import (
	"context"
	"fmt"
	"testing"

	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWriteFileTool_Execute(t *testing.T) {
	mockFSS := new(MockFileSystemService) // Reusing MockFileSystemService from extract_function_test.go
	tool := NewWriteFileTool(mockFSS)

	tests := []struct {
		name          string
		args          map[string]any
		setupMock     func()
		expectedLLMContent string
		expectedReturnDisplay string
		expectedError string
	}{
		{
			name:          "missing file_path argument",
			args:          map[string]any{"content": "test content"},
			setupMock:     func() {},
			expectedError: "missing or invalid 'file_path' argument",
		},
		{
			name:          "empty file_path argument",
			args:          map[string]any{"file_path": "", "content": "test content"},
			setupMock:     func() {},
			expectedError: "missing or invalid 'file_path' argument",
		},
		{
			name:          "missing content argument",
			args:          map[string]any{"file_path": "/tmp/file.txt"},
			setupMock:     func() {},
			expectedError: "missing or invalid 'content' argument",
		},
		{
			name: "write file fails",
			args: map[string]any{"file_path": "/tmp/file.txt", "content": "test content"},
			setupMock: func() {
				mockFSS.On("WriteFile", "/tmp/file.txt", "test content").Return(fmt.Errorf("write error")).Once()
			},
			expectedError: "failed to write to file /tmp/file.txt: write error",
		},
		{
			name: "successful write file",
			args: map[string]any{"file_path": "/tmp/file.txt", "content": "test content"},
			setupMock: func() {
				mockFSS.On("WriteFile", "/tmp/file.txt", "test content").Return(nil).Once()
			},
			expectedLLMContent:    "Successfully wrote to file: /tmp/file.txt",
			expectedReturnDisplay: "Successfully wrote to file: /tmp/file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFSS.Calls = []mock.Call{} // Reset calls for each test
			tt.setupMock()

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
			mockFSS.AssertExpectations(t)
		})
	}
}
