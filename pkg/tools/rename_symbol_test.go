package tools

import (
	"context"
	"fmt"
	"go-ai-agent-v2/go-cli/pkg/analysis" // Import the analysis package
	"go-ai-agent-v2/go-cli/pkg/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenameSymbolTool_Execute(t *testing.T) {
	tool := NewRenameSymbolTool()

	// Save original function and restore after test
	originalRenameSymbolFunc := analysis.RenameSymbolFunc
	defer func() {
		analysis.RenameSymbolFunc = originalRenameSymbolFunc
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
			args:          map[string]any{"line": float64(1), "column": float64(1), "new_name": "NewName"},
			setupMock:     func() {},
			expectedError: "missing or invalid 'file_path' argument",
		},
		{
			name:          "empty file_path argument",
			args:          map[string]any{"file_path": "", "line": float64(1), "column": float64(1), "new_name": "NewName"},
			setupMock:     func() {},
			expectedError: "missing or invalid 'file_path' argument",
		},
		{
			name:          "missing line argument",
			args:          map[string]any{"file_path": "/path/to/file.go", "column": float64(1), "new_name": "NewName"},
			setupMock:     func() {},
			expectedError: "missing or invalid 'line' argument",
		},
		{
			name:          "missing column argument",
			args:          map[string]any{"file_path": "/path/to/file.go", "line": float64(1), "new_name": "NewName"},
			setupMock:     func() {},
			expectedError: "missing or invalid 'column' argument",
		},
		{
			name:          "missing new_name argument",
			args:          map[string]any{"file_path": "/path/to/file.go", "line": float64(1), "column": float64(1)},
			setupMock:     func() {},
			expectedError: "missing or invalid 'new_name' argument",
		},
		{
			name:          "empty new_name argument",
			args:          map[string]any{"file_path": "/path/to/file.go", "line": float64(1), "column": float64(1), "new_name": ""},
			setupMock:     func() {},
			expectedError: "missing or invalid 'new_name' argument",
		},
		{
			name: "RenameSymbol returns error",
			args: map[string]any{"file_path": "/path/to/file.go", "line": float64(10), "column": float64(5), "new_name": "NewFunc"},
			setupMock: func() {
				analysis.RenameSymbolFunc = func(filePath string, line, column int, newName string) error {
					return fmt.Errorf("rename error")
				}
			},
			expectedError: "failed to rename symbol: rename error",
		},
		{
			name: "successful rename",
			args: map[string]any{"file_path": "/path/to/file.go", "line": float64(10), "column": float64(5), "new_name": "NewFunc"},
			setupMock: func() {
				analysis.RenameSymbolFunc = func(filePath string, line, column int, newName string) error {
					assert.Equal(t, "/path/to/file.go", filePath)
					assert.Equal(t, 10, line)
					assert.Equal(t, 5, column)
					assert.Equal(t, "NewFunc", newName)
					return nil
				}
			},
			expectedLLM: "Successfully renamed symbol at /path/to/file.go:10:5 to 'NewFunc'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			analysis.RenameSymbolFunc = originalRenameSymbolFunc
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
