package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/stretchr/testify/assert"
)


func TestLsTool_Execute(t *testing.T) {
	mockWorkspace := new(MockWorkspaceService)
	mockWorkspace.On("GetProjectRoot").Return("") // Not directly used in this test, but good practice to set up
	tool := NewLsTool(mockWorkspace)

	// Create a temporary directory for testing file system operations
	tempDir := t.TempDir()

	// Create some dummy files and directories
	os.MkdirAll(filepath.Join(tempDir, "subdir1"), 0755)
	os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("content1"), 0644)
	os.WriteFile(filepath.Join(tempDir, "file2.go"), []byte("content2"), 0644)

	tests := []struct {
		name                  string
		args                  map[string]any
		expectedLLMContent    string
		expectedReturnDisplay string
		expectedError         string
	}{
		{
			name:          "missing path argument",
			args:          map[string]any{},
			expectedError: "path argument is required and must be a string",
		},
		{
			name:          "empty path argument",
			args:          map[string]any{"path": ""},
			expectedError: "path argument is required and must be a string",
		},
		{
			name:                  "successful list directory",
			args:                  map[string]any{"path": tempDir},
			expectedLLMContent:    "file1.txt\nfile2.go\nsubdir1",
			expectedReturnDisplay: "file1.txt\nfile2.go\nsubdir1",
		},
		{
			name:          "directory not found",
			args:          map[string]any{"path": "/nonexistent/dir"},
			expectedError: "failed to read directory /nonexistent/dir: open /nonexistent/dir: no such file or directory",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := tt.args
			if p, ok := args["path"].(string); ok && p == tempDir {
				// Replace the placeholder path with the actual tempDir
				args["path"] = tempDir
			}

			result, err := tool.Execute(context.Background(), args)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Equal(t, types.ToolErrorTypeExecutionFailed, result.Error.Type)
			} else {
				assert.NoError(t, err)
				// Sort the expected and actual output for comparison, as os.ReadDir order is not guaranteed
				expectedLines := strings.Split(tt.expectedLLMContent, "\n")
				actualLines := strings.Split(result.LLMContent.(string), "\n")
				assert.ElementsMatch(t, expectedLines, actualLines)
				assert.ElementsMatch(t, expectedLines, strings.Split(result.ReturnDisplay, "\n"))
			}
		})
	}
}
