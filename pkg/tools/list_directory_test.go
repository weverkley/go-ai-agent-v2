package tools

import (
	"context" // New import
	"fmt"
	"testing"

	"go-ai-agent-v2/go-cli/pkg/services"
	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListDirectoryTool_Execute(t *testing.T) {
	// Setup mock FileSystemService
	mockFSS := new(services.MockFileSystemService)

	// Create a ListDirectoryTool with the mock service
	tool := &ListDirectoryTool{
		BaseDeclarativeTool: types.NewBaseDeclarativeTool("list_directory", "", "", types.KindOther, (&types.JsonSchemaObject{Type: "object"}).SetProperties(map[string]*types.JsonSchemaProperty{}).SetRequired([]string{}), false, false, nil),
		fileSystemService:   mockFSS, // Assign the mock directly
	}

	tests := []struct {
		name                string
		args                map[string]any
		setupMock           func()
		expectedLLMContent  string
		expectedReturnDisplay string
		expectedError       string
	}{
		{
			name:          "missing path argument",
			args:          map[string]any{},
			setupMock:     func() {},
			expectedError: "missing or invalid 'path' argument",
		},
		{
			name:          "empty path argument",
			args:          map[string]any{"path": ""},
			setupMock:     func() {},
			expectedError: "missing or invalid 'path' argument",
		},
		{
			name: "successful listing - no ignores",
			args: map[string]any{"path": "/test/dir"},
			setupMock: func() {
				mockFSS.On("ListDirectory", "/test/dir", []string{}, true, true).Return([]string{"file1.txt", "subdir"}, nil).Once()
			},
			expectedLLMContent:    "file1.txt\nsubdir",
			expectedReturnDisplay: "Listed directory /test/dir: 2 items",
		},
		{
			name: "successful listing - with ignores",
			args: map[string]any{"path": "/test/dir", "ignore": []interface{}{"*.log"}},
			setupMock: func() {
				mockFSS.On("ListDirectory", "/test/dir", []string{"*.log"}, true, true).Return([]string{"file1.txt"}, nil).Once()
			},
			expectedLLMContent:    "file1.txt",
			expectedReturnDisplay: "Listed directory /test/dir: 1 items",
		},
		{
			name: "fileSystemService returns error",
			args: map[string]any{"path": "/test/dir"},
			setupMock: func() {
				mockFSS.On("ListDirectory", "/test/dir", []string{}, true, true).Return([]string{}, fmt.Errorf("permission denied")).Once()
			},
			expectedError: "failed to list directory /test/dir: permission denied",
		},
		{
			name: "successful listing - respect_git_ignore false",
			args: map[string]any{"path": "/test/dir", "file_filtering_options": map[string]any{"respect_git_ignore": false}},
			setupMock: func() {
				mockFSS.On("ListDirectory", "/test/dir", []string{}, false, true).Return([]string{"file1.txt", "ignored.log"}, nil).Once()
			},
			expectedLLMContent:    "file1.txt\nignored.log",
			expectedReturnDisplay: "Listed directory /test/dir: 2 items",
		},
		{
			name: "successful listing - respect_goaiagent_ignore false",
			args: map[string]any{"path": "/test/dir", "file_filtering_options": map[string]any{"respect_goaiagent_ignore": false}},
			setupMock: func() {
				mockFSS.On("ListDirectory", "/test/dir", []string{}, true, false).Return([]string{"file1.txt", "ignored.goaiagent"}, nil).Once()
			},
			expectedLLMContent:    "file1.txt\nignored.goaiagent",
			expectedReturnDisplay: "Listed directory /test/dir: 2 items",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() {
				mockFSS.AssertExpectations(t)
				mockFSS.Calls = []mock.Call{} // Reset calls for next test
			})
			tt.setupMock()

			result, err := tool.Execute(context.Background(), tt.args)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedLLMContent, result.LLMContent)
				assert.Equal(t, tt.expectedReturnDisplay, result.ReturnDisplay)
			}
		})
	}
}