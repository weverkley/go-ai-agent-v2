package tools

import (
	"context"
	"fmt"
	"go-ai-agent-v2/go-cli/pkg/analysis" // Import the analysis package
	"go-ai-agent-v2/go-cli/pkg/types"
	"testing"

	"github.com/stretchr/testify/assert"
)


func TestFindReferencesTool_Execute(t *testing.T) {
	mockWorkspace := new(MockWorkspaceService)
	// For this test, the project root can be empty as it's not directly used in the logic being tested,
	// but the dependency needs to be satisfied.
	mockWorkspace.On("GetProjectRoot").Return("")

	tool := NewFindReferencesTool(mockWorkspace)

	// Save original function and restore after test
	originalFindSymbolReferencesFunc := analysis.FindSymbolReferencesFunc
	defer func() {
		analysis.FindSymbolReferencesFunc = originalFindSymbolReferencesFunc
	}()

	tests := []struct {
		name          string
		args          map[string]any
		setupMock     func()
		expectedLLM   string
		expectedError string
	}{
		{
			name:          "missing file_path argument",
			args:          map[string]any{"line": float64(1), "column": float64(1)},
			setupMock:     func() {},
			expectedError: "missing or invalid 'file_path' argument",
		},
		{
			name:          "empty file_path argument",
			args:          map[string]any{"file_path": "", "line": float64(1), "column": float64(1)},
			setupMock:     func() {},
			expectedError: "missing or invalid 'file_path' argument",
		},
		{
			name:          "missing line argument",
			args:          map[string]any{"file_path": "/path/to/file.go", "column": float64(1)},
			setupMock:     func() {},
			expectedError: "missing or invalid 'line' argument",
		},
		{
			name:          "missing column argument",
			args:          map[string]any{"file_path": "/path/to/file.go", "line": float64(1)},
			setupMock:     func() {},
			expectedError: "missing or invalid 'column' argument",
		},
		{
			name: "FindSymbolReferences returns error",
			args: map[string]any{"file_path": "/path/to/file.go", "line": float64(10), "column": float64(5)},
			setupMock: func() {
				analysis.FindSymbolReferencesFunc = func(filePath string, line, column int) ([]string, error) {
					return nil, fmt.Errorf("analysis error")
				}
			},
			expectedError: "failed to find references: analysis error",
		},
		{
			name: "no references found",
			args: map[string]any{"file_path": "/path/to/file.go", "line": float64(10), "column": float64(5)},
			setupMock: func() {
				analysis.FindSymbolReferencesFunc = func(filePath string, line, column int) ([]string, error) {
					assert.Equal(t, "/path/to/file.go", filePath)
					assert.Equal(t, 10, line)
					assert.Equal(t, 5, column)
					return []string{}, nil
				}
			},
			expectedLLM: "No references found for symbol at /path/to/file.go:10:5",
		},
		{
			name: "references found",
			args: map[string]any{"file_path": "/path/to/file.go", "line": float64(10), "column": float64(5)},
			setupMock: func() {
				analysis.FindSymbolReferencesFunc = func(filePath string, line, column int) ([]string, error) {
					return []string{"/path/to/other.go:20:3", "/path/to/another.go:5:1"}, nil
				}
			},
			expectedLLM: "Found references for symbol at /path/to/file.go:10:5:\n/path/to/other.go:20:3\n/path/to/another.go:5:1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			analysis.FindSymbolReferencesFunc = originalFindSymbolReferencesFunc
			tt.setupMock()

			result, err := tool.Execute(context.Background(), tt.args)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Equal(t, types.ToolErrorTypeExecutionFailed, result.Error.Type)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedLLM, result.LLMContent)
				assert.Equal(t, tt.expectedLLM, result.ReturnDisplay)
			}
		})
	}
}
