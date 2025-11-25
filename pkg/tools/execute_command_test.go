package tools

import (
	"context"
	"fmt"
	"testing"

	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockShellExecutionService is a mock implementation of services.ShellExecutionService
type MockShellExecutionService struct {
	mock.Mock
}

func (m *MockShellExecutionService) ExecuteCommand(ctx context.Context, command string, dir string) (string, string, error) {
	args := m.Called(ctx, command, dir)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockShellExecutionService) ExecuteCommandInBackground(command string, dir string) (int, error) {
	args := m.Called(command, dir)
	return args.Int(0), args.Error(1)
}

func (m *MockShellExecutionService) KillAllProcesses() {
	m.Called()
}

func TestExecuteCommandTool_Execute(t *testing.T) {
	mockShellService := new(MockShellExecutionService)
	tool := NewExecuteCommandTool(mockShellService)

	tests := []struct {
		name                  string
		args                  map[string]any
		setupMock             func()
		expectedLLMContent    string
		expectedReturnDisplay string
		expectedError         string
	}{
		{
			name:          "missing command argument",
			args:          map[string]any{},
			setupMock:     func() {},
			expectedError: "missing or invalid 'command' argument",
		},
		{
			name:          "empty command argument",
			args:          map[string]any{"command": ""},
			setupMock:     func() {},
			expectedError: "missing or invalid 'command' argument",
		},
		{
			name: "successful command execution with stdout",
			args: map[string]any{"command": "echo hello", "dir": "/tmp"},
			setupMock: func() {
				mockShellService.On("ExecuteCommand", mock.Anything, "echo hello", "/tmp").Return("hello\n", "", nil).Once()
			},
			expectedLLMContent:    "Stdout:\nhello\n",
			expectedReturnDisplay: "Stdout:\nhello\n",
		},
		{
			name: "successful command execution with stderr",
			args: map[string]any{"command": "ls /nonexistent", "dir": "/tmp"},
			setupMock: func() {
				mockShellService.On("ExecuteCommand", mock.Anything, "ls /nonexistent", "/tmp").Return("", "ls: /nonexistent: No such file or directory\n", fmt.Errorf("exit status 1")).Once()
			},
			expectedLLMContent:    "Stderr:\nls: /nonexistent: No such file or directory\nError: exit status 1\n",
			expectedReturnDisplay: "Stderr:\nls: /nonexistent: No such file or directory\nError: exit status 1\n",
			expectedError:         "command execution failed: exit status 1",
		},
		{
			name: "successful command execution with no output",
			args: map[string]any{"command": "true", "dir": "/tmp"},
			setupMock: func() {
				mockShellService.On("ExecuteCommand", mock.Anything, "true", "/tmp").Return("", "", nil).Once()
			},
			expectedLLMContent:    "Command executed successfully with no output.",
			expectedReturnDisplay: "Command executed successfully with no output.",
		},
		{
			name: "command execution fails",
			args: map[string]any{"command": "false", "dir": "/tmp"},
			setupMock: func() {
				mockShellService.On("ExecuteCommand", mock.Anything, "false", "/tmp").Return("", "", fmt.Errorf("exit status 1")).Once()
			},
			expectedLLMContent:    "Error: exit status 1\n",
			expectedReturnDisplay: "Error: exit status 1\n",
			expectedError:         "command execution failed: exit status 1",
		},
		{
			name: "execute command in background",
			args: map[string]any{"command": "sleep 5", "dir": "/tmp", "background": true},
			setupMock: func() {
				mockShellService.On("ExecuteCommandInBackground", "sleep 5", "/tmp").Return(12345, nil).Once()
			},
			expectedLLMContent:    "Command 'sleep 5' started in background with PID 12345",
			expectedReturnDisplay: "Command 'sleep 5' started in background with PID 12345",
		},
		{
			name: "execute command in background fails",
			args: map[string]any{"command": "sleep 5", "dir": "/tmp", "background": true},
			setupMock: func() {
				mockShellService.On("ExecuteCommandInBackground", "sleep 5", "/tmp").Return(0, fmt.Errorf("failed to start")).Once()
			},
			expectedError: "failed to execute command in background: failed to start",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockShellService.Calls = []mock.Call{} // Reset calls for each test
			tt.setupMock()

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
				assert.Equal(t, tt.expectedLLMContent, result.LLMContent)
				assert.Equal(t, tt.expectedReturnDisplay, result.ReturnDisplay)
				mockShellService.AssertExpectations(t)
			}
		})
	}
}
