package tools

import (
	"context" // New import
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)


func TestReadFileTool_Execute(t *testing.T) {
	mockWorkspace := new(MockWorkspaceService)
	mockWorkspace.On("GetProjectRoot").Return("") // Project root is joined, so an empty root means paths are relative to current dir
	tool := NewReadFileTool(mockWorkspace)

	tests := []struct {
		name          string
		args          map[string]any
		setup         func(t *testing.T) (string, func())
		expectedLLMContent  string
		expectedReturnDisplay string
		expectedError string
	}{
		{
			name:          "missing file_path argument",
			args:          map[string]any{},
			setup:         func(t *testing.T) (string, func()) { return "", func() {} },
			expectedError: "invalid or missing 'file_path' argument",
		},
		{
			name:          "empty file_path argument",
			args:          map[string]any{"file_path": ""},
			setup:         func(t *testing.T) (string, func()) { return "", func() {} },
			expectedError: "invalid or missing 'file_path' argument",
		},
		{
			name:          "file not found",
			args:          map[string]any{"file_path": "/nonexistent/file.txt"},
			setup:         func(t *testing.T) (string, func()) { return "", func() {} },
			expectedError: "file not found: /nonexistent/file.txt",
		},
		{
			name: "path is a directory",
			args: map[string]any{"file_path": t.TempDir()},
			setup: func(t *testing.T) (string, func()) {
				dir := t.TempDir()
				return dir, func() {}
			},
			expectedError: "path is a directory, not a file",
		},
		{
			name: "binary file (png)",
			args: map[string]any{"file_path": "/path/to/image.png"},
			setup: func(t *testing.T) (string, func()) {
				tempFile := filepath.Join(t.TempDir(), "image.png")
				err := os.WriteFile(tempFile, []byte("dummy image data"), 0644)
				assert.NoError(t, err)
				return tempFile, func() {}
			},
			expectedLLMContent:    "Content of /path/to/image.png (binary file, not displayed)",
			expectedReturnDisplay: "Content of /path/to/image.png (binary file, not displayed)",
		},
		{
			name: "successful text file read - entire file",
			args: map[string]any{"file_path": "/path/to/file.txt"},
			setup: func(t *testing.T) (string, func()) {
				tempFile := filepath.Join(t.TempDir(), "file.txt")
				content := "line 1\nline 2\nline 3"
				err := os.WriteFile(tempFile, []byte(content), 0644)
				assert.NoError(t, err)
				return tempFile, func() {}
			},
			expectedLLMContent:    "line 1\nline 2\nline 3\n",
			expectedReturnDisplay: "line 1\nline 2\nline 3\n",
		},
		{
			name: "successful text file read - with offset and limit",
			args: map[string]any{"file_path": "/path/to/file.txt", "offset": float64(1), "limit": float64(1)},
			setup: func(t *testing.T) (string, func()) {
				tempFile := filepath.Join(t.TempDir(), "file.txt")
				content := "line 1\nline 2\nline 3"
				err := os.WriteFile(tempFile, []byte(content), 0644)
				assert.NoError(t, err)
				return tempFile, func() {}
			},
			expectedLLMContent:    "\nIMPORTANT: The file content has been truncated.\nStatus: Showing lines 2-2 of 3 total lines.\nAction: To read more of the file, you can use the 'offset' and 'limit' parameters in a subsequent 'read_file' call. For example, to read the next section of the file, use offset: 2.\n\n--- FILE CONTENT (truncated) ---\nline 2\n",
			expectedReturnDisplay: "\nIMPORTANT: The file content has been truncated.\nStatus: Showing lines 2-2 of 3 total lines.\nAction: To read more of the file, you can use the 'offset' and 'limit' parameters in a subsequent 'read_file' call. For example, to read the next section of the file, use offset: 2.\n\n--- FILE CONTENT (truncated) ---\nline 2\n",
		},
		{
			name: "offset beyond file end",
			args: map[string]any{"file_path": "/path/to/file.txt", "offset": float64(5), "limit": float64(1)},
			setup: func(t *testing.T) (string, func()) {
				tempFile := filepath.Join(t.TempDir(), "file.txt")
				content := "line 1\nline 2\nline 3"
				err := os.WriteFile(tempFile, []byte(content), 0644)
				assert.NoError(t, err)
				return tempFile, func() {}
			},
			expectedError: "offset 5 is beyond the end of the file (total lines: 3)",
		},
		{
			name: "limit exceeding file end",
			args: map[string]any{"file_path": "/path/to/file.txt", "offset": float64(1), "limit": float64(5)},
			setup: func(t *testing.T) (string, func()) {
				tempFile := filepath.Join(t.TempDir(), "file.txt")
				content := "line 1\nline 2\nline 3"
				err := os.WriteFile(tempFile, []byte(content), 0644)
				assert.NoError(t, err)
				return tempFile, func() {}
			},
			expectedLLMContent:    "\nIMPORTANT: The file content has been truncated.\nStatus: Showing lines 2-3 of 3 total lines.\nAction: To read more of the file, you can use the 'offset' and 'limit' parameters in a subsequent 'read_file' call. For example, to read the next section of the file, use offset: 3.\n\n--- FILE CONTENT (truncated) ---\nline 2\nline 3\n",
			expectedReturnDisplay: "\nIMPORTANT: The file content has been truncated.\nStatus: Showing lines 2-3 of 3 total lines.\nAction: To read more of the file, you can use the 'offset' and 'limit' parameters in a subsequent 'read_file' call. For example, to read the next section of the file, use offset: 3.\n\n--- FILE CONTENT (truncated) ---\nline 2\nline 3\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call setup function to create test files and get the actual path
			actualPath, cleanup := tt.setup(t)
			defer cleanup()

			if path, ok := tt.args["file_path"].(string); ok && strings.Contains(path, "/path/to/") {
				tempArgs := make(map[string]any)
				for k, v := range tt.args {
					tempArgs[k] = v
				}
				tempArgs["file_path"] = actualPath
				tt.args = tempArgs
				if strings.Contains(tt.expectedLLMContent, "/path/to/") {
					tt.expectedLLMContent = strings.ReplaceAll(tt.expectedLLMContent, "/path/to/image.png", actualPath)
					tt.expectedLLMContent = strings.ReplaceAll(tt.expectedLLMContent, "/path/to/file.txt", actualPath)
				}
				if strings.Contains(tt.expectedReturnDisplay, "/path/to/") {
					tt.expectedReturnDisplay = strings.ReplaceAll(tt.expectedReturnDisplay, "/path/to/image.png", actualPath)
					tt.expectedReturnDisplay = strings.ReplaceAll(tt.expectedReturnDisplay, "/path/to/file.txt", actualPath)
				}
			}

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