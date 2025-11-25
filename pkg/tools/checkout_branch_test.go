package tools

import (
	"context"
	"fmt"
	"testing"

	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockGitService is a mock implementation of services.GitService
type MockGitService struct {
	mock.Mock
}

func (m *MockGitService) GetCurrentBranch(dir string) (string, error) {
	args := m.Called(dir)
	return args.String(0), args.Error(1)
}

func (m *MockGitService) GetRemoteURL(dir string) (string, error) {
	args := m.Called(dir)
	return args.String(0), args.Error(1)
}

func (m *MockGitService) CheckoutBranch(dir string, branchName string) error {
	args := m.Called(dir, branchName)
	return args.Error(0)
}

func (m *MockGitService) Pull(dir string, ref string) error {
	args := m.Called(dir, ref)
	return args.Error(0)
}

func (m *MockGitService) DeleteBranch(dir string, branchName string) error {
	args := m.Called(dir, branchName)
	return args.Error(0)
}

func (m *MockGitService) StageFiles(dir string, files []string) error {
	args := m.Called(dir, files)
	return args.Error(0)
}

func (m *MockGitService) Commit(dir, message string) error {
	args := m.Called(dir, message)
	return args.Error(0)
}

func (m *MockGitService) Clone(url string, directory string, ref string) error {
	args := m.Called(url, directory, ref)
	return args.Error(0)
}

func TestCheckoutBranchTool_Execute(t *testing.T) {
	tests := []struct {
		name                  string
		args                  map[string]any
		setupMock             func(mockGitService *MockGitService)
		expectedLLMContent    string
		expectedReturnDisplay string
		expectedError         string
	}{
		{
			name:          "missing dir argument",
			args:          map[string]any{"branch_name": "feature"},
			setupMock:     func(mockGitService *MockGitService) {},
			expectedError: "missing or invalid 'dir' argument",
		},
		{
			name:          "empty dir argument",
			args:          map[string]any{"dir": "", "branch_name": "feature"},
			setupMock:     func(mockGitService *MockGitService) {},
			expectedError: "missing or invalid 'dir' argument",
		},
		{
			name:          "missing branch_name argument",
			args:          map[string]any{"dir": "/test/repo"},
			setupMock:     func(mockGitService *MockGitService) {},
			expectedError: "missing or invalid 'branch_name' argument",
		},
		{
			name:          "empty branch_name argument",
			args:          map[string]any{"dir": "/test/repo", "branch_name": ""},
			setupMock:     func(mockGitService *MockGitService) {},
			expectedError: "missing or invalid 'branch_name' argument",
		},
		{
			name: "successful checkout",
			args: map[string]any{"dir": "/test/repo", "branch_name": "feature"},
			setupMock: func(mockGitService *MockGitService) {
				mockGitService.On("CheckoutBranch", "/test/repo", "feature").Return(nil).Once()
			},
			expectedLLMContent:    "Successfully checked out branch feature in /test/repo.",
			expectedReturnDisplay: "Successfully checked out branch feature in /test/repo.",
		},
		{
			name: "checkout fails",
			args: map[string]any{"dir": "/test/repo", "branch_name": "feature"},
			setupMock: func(mockGitService *MockGitService) {
				mockGitService.On("CheckoutBranch", "/test/repo", "feature").Return(fmt.Errorf("git error")).Once()
			},
			expectedError: "failed to checkout branch feature in /test/repo: git error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGitService := new(MockGitService)
			tool := NewCheckoutBranchTool(mockGitService)

			tt.setupMock(mockGitService)

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
				mockGitService.AssertExpectations(t)
			}
		})
	}
}
