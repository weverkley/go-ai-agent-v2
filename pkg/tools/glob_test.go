package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/gobwas/glob"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)


func TestGlobTool_Execute(t *testing.T) {
	// Create a temporary directory for testing file system operations
	tempDir := t.TempDir()

	// Create some dummy files and directories
	os.MkdirAll(filepath.Join(tempDir, "src"), 0755)
	os.MkdirAll(filepath.Join(tempDir, "docs"), 0755)
	os.MkdirAll(filepath.Join(tempDir, ".git"), 0755)         // Simulate a git repo
	os.MkdirAll(filepath.Join(tempDir, "node_modules"), 0755) // Simulate node_modules

	os.WriteFile(filepath.Join(tempDir, "main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tempDir, "src", "app.go"), []byte("package src"), 0644)
	os.WriteFile(filepath.Join(tempDir, "docs", "README.md"), []byte("# README"), 0644)
	os.WriteFile(filepath.Join(tempDir, "test.log"), []byte("log data"), 0644)
	os.WriteFile(filepath.Join(tempDir, "src", "test.log"), []byte("log data"), 0644)
	os.WriteFile(filepath.Join(tempDir, "node_modules", "package.json"), []byte("{}"), 0644)

	// Create a .gitignore file
	os.WriteFile(filepath.Join(tempDir, ".gitignore"), []byte("*.log\nnode_modules/"), 0644)

	tests := []struct {
		name                  string
		args                  map[string]any
		setupMock             func(*MockFileSystemService)
		expectedLLMContent    string
		expectedReturnDisplay string
		expectedError         string
	}{
		{
			name: "missing pattern argument",
			args: map[string]any{},
			setupMock: func(mockFSS *MockFileSystemService) {
				mockFSS.On("GetIgnorePatterns", mock.AnythingOfType("string"), true, true).Return([]glob.Glob{}, nil).Maybe()
			},
			expectedError: "invalid or missing 'pattern' argument",
		},
		{
			name: "empty pattern argument",
			args: map[string]any{"pattern": ""},
			setupMock: func(mockFSS *MockFileSystemService) {
				mockFSS.On("GetIgnorePatterns", mock.AnythingOfType("string"), true, true).Return([]glob.Glob{}, nil).Maybe()
			},
			expectedError: "invalid or missing 'pattern' argument",
		},
		{
			name: "successful glob - all go files",
			args: map[string]any{"pattern": "**/*.go", "path": "."},
			setupMock: func(mockFSS *MockFileSystemService) {
				mockFSS.On("GetIgnorePatterns", mock.AnythingOfType("string"), true, true).Return([]glob.Glob{}, nil).Once()
			},
			expectedLLMContent:    "Found 2 file(s) matching \"**/*.go\" in path \".\":",
			expectedReturnDisplay: "Found 2 file(s) matching \"**/*.go\" in path \".\":",
		},
		{
			name: "successful glob - markdown files",
			args: map[string]any{"pattern": "docs/*.md", "path": "."},
			setupMock: func(mockFSS *MockFileSystemService) {
				mockFSS.On("GetIgnorePatterns", mock.AnythingOfType("string"), true, true).Return([]glob.Glob{}, nil).Once()
			},
			expectedLLMContent:    fmt.Sprintf("Found 1 file(s) matching \"docs/*.md\" in path \".\":\n%s", filepath.Join(tempDir, "docs", "README.md")),
			expectedReturnDisplay: fmt.Sprintf("Found 1 file(s) matching \"docs/*.md\" in path \".\":\n%s", filepath.Join(tempDir, "docs", "README.md")),
		},
		{
			name: "no files found",
			args: map[string]any{"pattern": "**/*.txt", "path": "."},
			setupMock: func(mockFSS *MockFileSystemService) {
				mockFSS.On("GetIgnorePatterns", mock.AnythingOfType("string"), true, true).Return([]glob.Glob{}, nil).Once()
			},
			expectedLLMContent:    "No files found matching pattern \"**/*.txt\" in path \".\"",
			expectedReturnDisplay: "No files found matching pattern \"**/*.txt\" in path \".\"",
		},
		{
			name: "invalid glob pattern",
			args: map[string]any{"pattern": "[", "path": "."},
			setupMock: func(mockFSS *MockFileSystemService) {
				mockFSS.On("GetIgnorePatterns", mock.AnythingOfType("string"), true, true).Return([]glob.Glob{}, nil).Maybe() // Maybe because it might not be called if pattern is invalid
			},
			expectedError: "failed to compile glob pattern [: unexpected end of input",
		},
		{
			name: "respect .gitignore",
			args: map[string]any{"pattern": "**/*.log", "path": ".", "respect_git_ignore": true},
			setupMock: func(mockFSS *MockFileSystemService) {
				// Mock GetIgnorePatterns to return a glob for *.log and node_modules/
				logGlob, _ := glob.Compile("*.log")
				nodeModulesGlob, _ := glob.Compile("node_modules/")
				mockFSS.On("GetIgnorePatterns", mock.AnythingOfType("string"), true, true).Return([]glob.Glob{logGlob, nodeModulesGlob}, nil).Once()
			},
			expectedLLMContent:    "No files found matching pattern \"**/*.log\" in path \".\"",
			expectedReturnDisplay: "No files found matching pattern \"**/*.log\" in path \".\"",
		},
		{
			name: "do not respect .gitignore",
			args: map[string]any{"pattern": "**/*.log", "path": ".", "respect_git_ignore": false},
			setupMock: func(mockFSS *MockFileSystemService) {
				// Mock GetIgnorePatterns to return an empty slice (no .gitignore respected)
				mockFSS.On("GetIgnorePatterns", mock.AnythingOfType("string"), false, true).Return([]glob.Glob{}, nil).Once()
			},
			expectedLLMContent:    fmt.Sprintf("Found 2 file(s) matching \"**/*.log\" in path \".\":\n%s\n%s", filepath.Join(tempDir, "src", "test.log"), filepath.Join(tempDir, "test.log")),
			expectedReturnDisplay: fmt.Sprintf("Found 2 file(s) matching \"**/*.log\" in path \".\":\n%s\n%s", filepath.Join(tempDir, "src", "test.log"), filepath.Join(tempDir, "test.log")),
		},
		{
			name: "case insensitive glob",
			args: map[string]any{"pattern": "docs/*.md", "path": "."},
			setupMock: func(mockFSS *MockFileSystemService) {
				mockFSS.On("GetIgnorePatterns", mock.AnythingOfType("string"), true, true).Return([]glob.Glob{}, nil).Once()
			},
			expectedLLMContent:    fmt.Sprintf("Found 1 file(s) matching \"docs/*.md\" in path \".\":\n%s", filepath.Join(tempDir, "docs", "README.md")),
			expectedReturnDisplay: fmt.Sprintf("Found 1 file(s) matching \"docs/*.md\" in path \".\":\n%s", filepath.Join(tempDir, "docs", "README.md")),
		},
		{
			name: "case sensitive glob - no match",
			args: map[string]any{"pattern": "docs/*.MD", "path": ".", "case_sensitive": true},
			setupMock: func(mockFSS *MockFileSystemService) {
				mockFSS.On("GetIgnorePatterns", mock.AnythingOfType("string"), true, true).Return([]glob.Glob{}, nil).Once()
			},
			expectedLLMContent:    "No files found matching pattern \"docs/*.MD\" in path \".\"",
			expectedReturnDisplay: "No files found matching pattern \"docs/*.MD\" in path \".\"",
		},
		{
			name: "GetIgnorePatterns returns error",
			args: map[string]any{"pattern": "**/*.go", "path": "."},
			setupMock: func(mockFSS *MockFileSystemService) {
				mockFSS.On("GetIgnorePatterns", mock.AnythingOfType("string"), true, true).Return([]glob.Glob{}, fmt.Errorf("ignore error")).Once()
			},
			expectedError: "failed to get ignore patterns: ignore error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFSS := new(MockFileSystemService) // Create a new mock for each subtest
			mockWorkspace := new(MockWorkspaceService)
			mockWorkspace.On("GetProjectRoot").Return(tempDir)
			tool := NewGlobTool(mockFSS, mockWorkspace) // Pass the new mock to the tool

			tt.setupMock(mockFSS) // Pass the mock to the setup function

			result, err := tool.Execute(context.Background(), tt.args)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				if result.Error != nil {
					assert.Equal(t, types.ToolErrorTypeExecutionFailed, result.Error.Type)
				} else {
					assert.Fail(t, "Expected ToolResult.Error to be non-nil when an error is expected")
				}
			} else {
				assert.NoError(t, err)
				assert.Contains(t, result.LLMContent, tt.expectedLLMContent)
				assert.Contains(t, result.ReturnDisplay, tt.expectedReturnDisplay)
				mockFSS.AssertExpectations(t)
			}
		})
	}
}
