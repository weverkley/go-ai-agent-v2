package tools

import (
	"context"
	"fmt"
	"testing"

	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSmartEditTool_Execute(t *testing.T) {
	mockFSS := new(MockFileSystemService) // Reusing MockFileSystemService from extract_function_test.go
	tool := NewSmartEditTool(mockFSS)

	tests := []struct {
		name                  string
		args                  map[string]any
		setupMock             func()
		expectedLLMContent    string
		expectedReturnDisplay string
		expectedError         string
	}{
		{
			name:          "missing file_path argument",
			args:          map[string]any{"instruction": "change", "old_string": "old", "new_string": "new"},
			setupMock:     func() {},
			expectedError: "invalid or missing 'file_path' argument",
		},
		{
			name:          "empty file_path argument",
			args:          map[string]any{"file_path": "", "instruction": "change", "old_string": "old", "new_string": "new"},
			setupMock:     func() {},
			expectedError: "invalid or missing 'file_path' argument",
		},
		{
			name:          "missing instruction argument",
			args:          map[string]any{"file_path": "/tmp/file.txt", "old_string": "old", "new_string": "new"},
			setupMock:     func() {},
			expectedError: "invalid or missing 'instruction' argument",
		},
		{
			name:          "missing new_string argument",
			args:          map[string]any{"file_path": "/tmp/file.txt", "instruction": "change", "old_string": "old"},
			setupMock:     func() {},
			expectedError: "invalid or missing 'new_string' argument",
		},
		{
			name: "read file fails",
			args: map[string]any{"file_path": "/tmp/file.txt", "instruction": "change", "old_string": "old", "new_string": "new"},
			setupMock: func() {
				mockFSS.On("ReadFile", "/tmp/file.txt").Return("", fmt.Errorf("read error")).Once()
			},
			expectedError: "failed to read file /tmp/file.txt: read error",
		},
		{
			name: "old_string is empty",
			args: map[string]any{"file_path": "/tmp/file.txt", "instruction": "change", "old_string": "", "new_string": "new"},
			setupMock: func() {
				mockFSS.On("ReadFile", "/tmp/file.txt").Return("content", nil).Once()
			},
			expectedError: "old_string cannot be empty for smart_edit. Use write_file to create new files.",
		},
		{
			name: "old_string not found",
			args: map[string]any{"file_path": "/tmp/file.txt", "instruction": "change", "old_string": "nonexistent", "new_string": "new"},
			setupMock: func() {
				mockFSS.On("ReadFile", "/tmp/file.txt").Return("original content", nil).Once()
			},
			expectedError: "old_string not found or no changes made in file /tmp/file.txt",
		},
		{
			name: "write file fails",
			args: map[string]any{"file_path": "/tmp/file.txt", "instruction": "change", "old_string": "old", "new_string": "new"},
			setupMock: func() {
				mockFSS.On("ReadFile", "/tmp/file.txt").Return("old content", nil).Once()
				mockFSS.On("WriteFile", "/tmp/file.txt", "new content").Return(fmt.Errorf("write error")).Once()
			},
			expectedError: "failed to write file /tmp/file.txt: write error",
		},
		{
			name: "successful smart edit",
			args: map[string]any{"file_path": "/tmp/file.txt", "instruction": "change", "old_string": "old", "new_string": "new"},
			setupMock: func() {
				mockFSS.On("ReadFile", "/tmp/file.txt").Return("old content", nil).Once()
				mockFSS.On("WriteFile", "/tmp/file.txt", "new content").Return(nil).Once()
			},
			expectedLLMContent:    "Successfully modified file: /tmp/file.txt",
			expectedReturnDisplay: "Successfully modified file: /tmp/file.txt",
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
