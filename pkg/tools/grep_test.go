package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/stretchr/testify/assert"
)

func TestGrepTool_Execute(t *testing.T) {
	tool := NewGrepTool()

	// Create a temporary directory for testing file system operations
	tempDir := t.TempDir()

	// Create some dummy files
	os.MkdirAll(filepath.Join(tempDir, "src"), 0755)
	os.WriteFile(filepath.Join(tempDir, "main.go"), []byte("package main\nfunc main() {\n\tfmt.Println(\"Hello, World!\")\n}\n"), 0644)
	os.WriteFile(filepath.Join(tempDir, "src", "utils.go"), []byte("package utils\nfunc Helper() {\n\t// Some helper function\n}\n"), 0644)
	os.WriteFile(filepath.Join(tempDir, "README.md"), []byte("# Project\nThis is a test project.\n"), 0644)
	os.WriteFile(filepath.Join(tempDir, "test.txt"), []byte("This is a test file.\nIt contains the word test.\n"), 0644)
	os.WriteFile(filepath.Join(tempDir, "ignored.log"), []byte("log entry\n"), 0644)

	tests := []struct {
		name                  string
		args                  map[string]any
		expectedLLMContent    string
		expectedReturnDisplay string
		expectedError         string
	}{
		{
			name:          "missing pattern argument",
			args:          map[string]any{},
			expectedError: "invalid or missing 'pattern' argument",
		},
		{
			name:          "empty pattern argument",
			args:          map[string]any{"pattern": ""},
			expectedError: "invalid or missing 'pattern' argument",
		},
		{
			name:          "invalid regex pattern",
			args:          map[string]any{"pattern": "["},
			expectedError: "invalid regex pattern: error parsing regexp: missing closing ]: `[`",
		},
		{
			name:                  "successful grep - single file",
			args:                  map[string]any{"pattern": "Hello", "path": tempDir},
			expectedLLMContent:    fmt.Sprintf("Found 1 matches for pattern \"Hello\" in path \"%s\":\n---\nFile: %s\nL3: fmt.Println(\"Hello, World!\")\n---\n\n", tempDir, filepath.Join(tempDir, "main.go")),
			expectedReturnDisplay: fmt.Sprintf("Found 1 matches for pattern \"Hello\" in path \"%s\":\n---\nFile: %s\nL3: fmt.Println(\"Hello, World!\")\n---\n\n", tempDir, filepath.Join(tempDir, "main.go")),
		},
		{
			name:                  "successful grep - multiple files",
			args:                  map[string]any{"pattern": "test", "path": tempDir},
			expectedLLMContent:    fmt.Sprintf("Found 3 matches for pattern \"test\" in path \"%s\":\n---\nFile: %s\nL2: This is a test project.\n---\n\n\nFile: %s\nL1: This is a test file.\nL2: It contains the word test.\n---\n\n", tempDir, filepath.Join(tempDir, "README.md"), filepath.Join(tempDir, "test.txt")),
			expectedReturnDisplay: fmt.Sprintf("Found 3 matches for pattern \"test\" in path \"%s\":\n---\nFile: %s\nL2: This is a test project.\n---\n\n\nFile: %s\nL1: This is a test file.\nL2: It contains the word test.\n---\n\n", tempDir, filepath.Join(tempDir, "README.md"), filepath.Join(tempDir, "test.txt")),
		},
		{
			name:                  "no matches found",
			args:                  map[string]any{"pattern": "NonExistent", "path": tempDir},
			expectedLLMContent:    fmt.Sprintf("No matches found for pattern \"NonExistent\" in path \"%s\"", tempDir),
			expectedReturnDisplay: fmt.Sprintf("No matches found for pattern \"NonExistent\" in path \"%s\"", tempDir),
		},
		{
			name:                  "filter by include glob",
			args:                  map[string]any{"pattern": "package", "path": tempDir, "include": "*.go"},
			expectedLLMContent:    fmt.Sprintf("Found 2 matches for pattern \"package\" in path \"%s\":\n---\nFile: %s\nL1: package main\n---\n\n\nFile: %s\nL1: package utils\n---\n\n", tempDir, filepath.Join(tempDir, "main.go"), filepath.Join(tempDir, "src", "utils.go")),
			expectedReturnDisplay: fmt.Sprintf("Found 2 matches for pattern \"package\" in path \"%s\":\n---\nFile: %s\nL1: package main\n---\n\n\nFile: %s\nL1: package utils\n---\n\n", tempDir, filepath.Join(tempDir, "main.go"), filepath.Join(tempDir, "src", "utils.go")),
		},
		{
			name:                  "filter by include glob - no match",
			args:                  map[string]any{"pattern": "package", "path": tempDir, "include": "*.md"},
			expectedLLMContent:    fmt.Sprintf("No matches found for pattern \"package\" in path \"%s\"", tempDir),
			expectedReturnDisplay: fmt.Sprintf("No matches found for pattern \"package\" in path \"%s\"", tempDir),
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
				assert.Equal(t, tt.expectedLLMContent, result.LLMContent)
				assert.Equal(t, tt.expectedReturnDisplay, result.ReturnDisplay)
			}
		})
	}
}
