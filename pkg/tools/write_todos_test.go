package tools

import (
	"context"
	"fmt"
	"os"
	"testing"

	"go-ai-agent-v2/go-cli/pkg/services"
	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/stretchr/testify/assert"
)

func TestWriteTodosTool_Execute(t *testing.T) {
	tests := []struct {
		name                  string
		args                  map[string]any
		setupMock             func(mockService *services.MockSettingsService, tempDir string)
		expectedLLMContent    string
		expectedReturnDisplay string
		expectedError         string
	}{
		{
			name:          "missing todos argument",
			args:          map[string]any{},
			setupMock:     func(mockService *services.MockSettingsService, tempDir string) {},
			expectedError: "invalid or missing 'todos' argument",
		},
		{
			name:          "invalid todo item format",
			args:          map[string]any{"todos": []any{"not a map"}},
			setupMock:     func(mockService *services.MockSettingsService, tempDir string) {},
			expectedError: "invalid todo item format",
		},
		{
			name:          "todo with empty description",
			args:          map[string]any{"todos": []any{map[string]any{"description": "", "status": "pending"}}},
			setupMock:     func(mockService *services.MockSettingsService, tempDir string) {},
			expectedError: "each todo must have a non-empty description",
		},
		{
			name:          "todo with invalid status",
			args:          map[string]any{"todos": []any{map[string]any{"description": "Task 1", "status": "invalid"}}},
			setupMock:     func(mockService *services.MockSettingsService, tempDir string) {},
			expectedError: "invalid todo status: invalid",
		},
		{
			name: "multiple in_progress tasks",
			args: map[string]any{"todos": []any{
				map[string]any{"description": "Task 1", "status": "in_progress"},
				map[string]any{"description": "Task 2", "status": "in_progress"},
			}},
			setupMock:     func(mockService *services.MockSettingsService, tempDir string) {},
			expectedError: "only one task can be \"in_progress\" at a time",
		},
		{
			name: "fallback to home dir fails",
			args: map[string]any{"todos": []any{map[string]any{"description": "Task 1", "status": "pending"}}},
			setupMock: func(mockService *services.MockSettingsService, tempDir string) {
				mockService.On("GetWorkspaceDir").Return("").Once()
				testMockUserHomeDir = func() (string, error) { return "", fmt.Errorf("home dir error") }
			},
			expectedError: "failed to get user home directory: home dir error",
		},
		{
			name: "create todos directory fails",
			args: map[string]any{"todos": []any{map[string]any{"description": "Task 1", "status": "pending"}}},
			setupMock: func(mockService *services.MockSettingsService, tempDir string) {
				mockService.On("GetWorkspaceDir").Return(tempDir).Once()
				testMockMkdirAll = func(path string, perm os.FileMode) error { return fmt.Errorf("mkdir error") }
			},
			expectedError: "failed to create todos directory: mkdir error",
		},
		{
			name: "write todos file fails",
			args: map[string]any{"todos": []any{map[string]any{"description": "Task 1", "status": "pending"}}},
			setupMock: func(mockService *services.MockSettingsService, tempDir string) {
				mockService.On("GetWorkspaceDir").Return(tempDir).Once()
				testMockMkdirAll = func(path string, perm os.FileMode) error { return nil }
				testMockWriteFile = func(name string, data []byte, perm os.FileMode) error { return fmt.Errorf("write file error") }
			},
			expectedError: "failed to write todos file: write file error",
		},
		{
			name: "successful write - empty todos",
			args: map[string]any{"todos": []any{}},
			setupMock: func(mockService *services.MockSettingsService, tempDir string) {
				mockService.On("GetWorkspaceDir").Return(tempDir).Once()
				testMockMkdirAll = func(path string, perm os.FileMode) error { return nil }
				testMockWriteFile = func(name string, data []byte, perm os.FileMode) error {
					assert.Equal(t, "", string(data))
					return nil
				}
			},
			expectedLLMContent:    "Successfully cleared the todo list.",
			expectedReturnDisplay: "Successfully cleared the todo list.",
		},
		{
			name: "successful write - single todo",
			args: map[string]any{"todos": []any{map[string]any{"description": "Task 1", "status": "pending"}}},
			setupMock: func(mockService *services.MockSettingsService, tempDir string) {
				mockService.On("GetWorkspaceDir").Return(tempDir).Once()
				testMockMkdirAll = func(path string, perm os.FileMode) error { return nil }
				testMockWriteFile = func(name string, data []byte, perm os.FileMode) error {
					expectedContent := "# ToDo List\n\n1. [pending] Task 1\n"
					assert.Equal(t, expectedContent, string(data))
					return nil
				}
			},
			expectedLLMContent:    "Successfully updated the todo list. The current list is now:\n# ToDo List\n\n1. [pending] Task 1\n",
			expectedReturnDisplay: "Successfully updated the todo list. The current list is now:\n# ToDo List\n\n1. [pending] Task 1\n",
		},
		{
			name: "successful write - multiple todos",
			args: map[string]any{"todos": []any{
				map[string]any{"description": "Task 1", "status": "completed"},
				map[string]any{"description": "Task 2", "status": "in_progress"},
			}},
			setupMock: func(mockService *services.MockSettingsService, tempDir string) {
				mockService.On("GetWorkspaceDir").Return(tempDir).Once()
				testMockMkdirAll = func(path string, perm os.FileMode) error { return nil }
				testMockWriteFile = func(name string, data []byte, perm os.FileMode) error {
					expectedContent := "# ToDo List\n\n1. [completed] Task 1\n2. [in_progress] Task 2\n"
					assert.Equal(t, expectedContent, string(data))
					return nil
				}
			},
			expectedLLMContent:    "Successfully updated the todo list. The current list is now:\n# ToDo List\n\n1. [completed] Task 1\n2. [in_progress] Task 2\n",
			expectedReturnDisplay: "Successfully updated the todo list. The current list is now:\n# ToDo List\n\n1. [completed] Task 1\n2. [in_progress] Task 2\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			mockService := new(services.MockSettingsService)
			tool := NewWriteTodosTool(mockService)

			// Setup mocks
			tt.setupMock(mockService, tempDir)
			setupOsMocks()
			t.Cleanup(func() {
				teardownOsMocks()
				mockService.AssertExpectations(t)
			})

			result, err := tool.Execute(context.Background(), tt.args)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				if result.Error != nil {
					assert.Equal(t, types.ToolErrorTypeExecutionFailed, result.Error.Type)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedReturnDisplay, result.ReturnDisplay)
				// Assert the new structured LLMContent
				llmContentMap, ok := result.LLMContent.(map[string]interface{})
				assert.True(t, ok, "LLMContent should be a map")
				assert.Equal(t, true, llmContentMap["success"])
				assert.Contains(t, llmContentMap["message"], "Successfully")
			}
		})
	}
}
