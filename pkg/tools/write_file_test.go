package tools

import (
	"context"
	"fmt"
	"path/filepath" // Added import
	"testing"

	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockWorkspaceService implements types.WorkspaceServiceIface for testing
type MockWorkspaceService struct {
	mock.Mock
}

func (m *MockWorkspaceService) GetProjectRoot() string {
	args := m.Called()
	return args.String(0)
}

func TestWriteFileTool_Execute(t *testing.T) {
	mockFSS := new(MockFileSystemService)
	mockWSS := new(MockWorkspaceService) // Using the new mock

	tempProjectRoot := t.TempDir()
	mockWSS.On("GetProjectRoot").Return(tempProjectRoot).Maybe() // Use .Maybe() if not always called

	tool := NewWriteFileTool(mockFSS, mockWSS) // Pass the mock

	// Helper to resolve paths in test expectations
	resolvePath := func(relativePath string) string {
		return filepath.Join(tempProjectRoot, relativePath)
	}

	tests := []struct {
		name          string
		args          map[string]any
		setupMock     func()
		expectedLLMContent string
		expectedReturnDisplay string
		expectedError string
	}{
		{
			name: "missing file_path argument",
			args: map[string]any{"content": "test content"},
			setupMock: func() {},
			expectedError: "missing or invalid 'file_path' argument",
		},
		{
			name: "empty file_path argument",
			args: map[string]any{"file_path": "", "content": "test content"},
			setupMock: func() {},
			expectedError: "missing or invalid 'file_path' argument",
		},
		{
			name: "missing content argument",
			args: map[string]any{"file_path": "test.txt"},
			setupMock: func() {},
			expectedError: "missing or invalid 'content' argument",
		},
		{
			name: "write file fails",
			args: map[string]any{"file_path": "test.txt", "content": "test content"},
			setupMock: func() {
				mockFSS.On("WriteFile", resolvePath("test.txt"), "test content").Return(fmt.Errorf("write error")).Once()
			},
			expectedError: fmt.Sprintf("failed to write to file %s: write error", resolvePath("test.txt")),
		},
		{
			name: "successful write file",
			args: map[string]any{"file_path": "test.txt", "content": "test content"},
			setupMock: func() {
				mockFSS.On("WriteFile", resolvePath("test.txt"), "test content").Return(nil).Once()
			},
			expectedLLMContent:    fmt.Sprintf("Successfully wrote to file: %s", resolvePath("test.txt")),
			expectedReturnDisplay: fmt.Sprintf("Successfully wrote to file: %s", resolvePath("test.txt")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFSS.Calls = []mock.Call{} // Reset calls for each test
			mockWSS.Calls = []mock.Call{} // Reset calls for each test
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
			mockWSS.AssertExpectations(t) // Assert expectations for mockWSS
		})
	}
}
