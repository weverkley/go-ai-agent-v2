package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go-ai-agent-v2/go-cli/pkg/services"
	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadManyFilesTool_Execute(t *testing.T) {
	fs := services.NewFileSystemService()
	tool := NewReadManyFilesTool(fs)

	tempDir, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)

	// Create some dummy files and directories
	require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "src"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "docs"), 0755))

	mainGoPath := filepath.Join(tempDir, "main.go")
	appGoPath := filepath.Join(tempDir, "src", "app.go")
	readmePath := filepath.Join(tempDir, "docs", "README.md")

	require.NoError(t, os.WriteFile(mainGoPath, []byte("package main\n"), 0644))
	require.NoError(t, os.WriteFile(appGoPath, []byte("package src\n"), 0644))
	require.NoError(t, os.WriteFile(readmePath, []byte("# README\n"), 0644))

	// Change working directory to the temp directory for the duration of the test
	originalCwd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tempDir))
	defer os.Chdir(originalCwd)

	t.Run("missing paths argument", func(t *testing.T) {
		result, err := tool.Execute(context.Background(), map[string]any{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid or missing 'paths' argument")
		require.NotNil(t, result.Error)
		assert.Equal(t, types.ToolErrorTypeExecutionFailed, result.Error.Type)
	})

	t.Run("empty paths argument", func(t *testing.T) {
		result, err := tool.Execute(context.Background(), map[string]any{"paths": []any{}})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid or missing 'paths' argument")
		require.NotNil(t, result.Error)
		assert.Equal(t, types.ToolErrorTypeExecutionFailed, result.Error.Type)
	})

	t.Run("successful read - go files", func(t *testing.T) {
		result, err := tool.Execute(context.Background(), map[string]any{"paths": []any{"**/*.go", "*.go"}})
		require.NoError(t, err)

		assert.Contains(t, result.LLMContent, fmt.Sprintf("--- %s ---\npackage main\n", mainGoPath))
		assert.Contains(t, result.LLMContent, fmt.Sprintf("--- %s ---\npackage src\n", appGoPath))
		assert.True(t, strings.HasSuffix(result.LLMContent.(string), "\n--- End of content ---\n\n"+result.ReturnDisplay))

		// Check display message
		assert.Contains(t, result.ReturnDisplay, "Successfully read and concatenated content from **2 file(s)**.")
		assert.Contains(t, result.ReturnDisplay, fmt.Sprintf("- `%s`", mainGoPath))
		assert.Contains(t, result.ReturnDisplay, fmt.Sprintf("- `%s`", appGoPath))
	})

	t.Run("successful read - with exclude", func(t *testing.T) {
		result, err := tool.Execute(context.Background(), map[string]any{
			"paths":   []any{"**/*.go", "*.go"},
			"exclude": []any{"src/app.go"},
		})
		require.NoError(t, err)

		// Check that only main.go is included
		assert.Contains(t, result.LLMContent, fmt.Sprintf("--- %s ---\npackage main\n", mainGoPath))
		assert.NotContains(t, result.LLMContent, fmt.Sprintf("--- %s ---", appGoPath))

		// Check display message
		assert.Contains(t, result.ReturnDisplay, "Successfully read and concatenated content from **1 file(s)**.")
		assert.Contains(t, result.ReturnDisplay, fmt.Sprintf("- `%s`", mainGoPath))
		assert.NotContains(t, result.ReturnDisplay, fmt.Sprintf("- `%s`", appGoPath))
	})

	t.Run("no files found", func(t *testing.T) {
		result, err := tool.Execute(context.Background(), map[string]any{"paths": []any{"**/*.txt"}})
		require.NoError(t, err)

		expectedDisplay := "### ReadManyFiles Result\n\nNo files were read and concatenated based on the criteria.\n"
		expectedLLMContent := "\n--- End of content ---\n\n" + expectedDisplay

		assert.Equal(t, expectedLLMContent, result.LLMContent)
		assert.Equal(t, expectedDisplay, result.ReturnDisplay)
	})
}
