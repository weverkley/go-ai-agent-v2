package tools

import (
	"context"
	"os"
	"testing"

	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/gobwas/glob"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockFileSystemService is a mock implementation of services.FileSystemService
type MockFileSystemService struct {
	mock.Mock
}

func (m *MockFileSystemService) ListDirectory(dirPath string, ignorePatterns []string, respectGitIgnore, respectGeminiIgnore bool) ([]string, error) {
	args := m.Called(dirPath, ignorePatterns, respectGitIgnore, respectGeminiIgnore)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockFileSystemService) GetIgnorePatterns(searchDir string, respectGitIgnore, respectGoaiagentIgnore bool) ([]glob.Glob, error) {
	args := m.Called(searchDir, respectGitIgnore, respectGoaiagentIgnore)
	return args.Get(0).([]glob.Glob), args.Error(1)
}

func (m *MockFileSystemService) PathExists(path string) (bool, error) {
	args := m.Called(path)
	return args.Bool(0), args.Error(1)
}

func (m *MockFileSystemService) IsDirectory(path string) (bool, error) {
	args := m.Called(path)
	return args.Bool(0), args.Error(1)
}

func (m *MockFileSystemService) ReadFile(filePath string) (string, error) {
	args := m.Called(filePath)
	return args.String(0), args.Error(1)
}

func (m *MockFileSystemService) WriteFile(filePath string, content string) error {
	args := m.Called(filePath, content)
	return args.Error(0)
}

func (m *MockFileSystemService) CreateDirectory(path string) error {
	args := m.Called(path)
	return args.Error(0)
}

func (m *MockFileSystemService) CopyDirectory(src string, dst string) error {
	args := m.Called(src, dst)
	return args.Error(0)
}

func (m *MockFileSystemService) JoinPaths(elements ...string) string {
	args := m.Called(elements)
	return args.String(0)
}

func (m *MockFileSystemService) Symlink(oldname, newname string) error {
	args := m.Called(oldname, newname)
	return args.Error(0)
}

func (m *MockFileSystemService) RemoveAll(path string) error {
	args := m.Called(path)
	return args.Error(0)
}

func (m *MockFileSystemService) MkdirAll(path string, perm os.FileMode) error {
	args := m.Called(path, perm)
	return args.Error(0)
}

func (m *MockFileSystemService) Rename(oldpath, newpath string) error {
	args := m.Called(oldpath, newpath)
	return args.Error(0)
}

// MockExtractFunction is a mock for agents.ExtractFunction

// var MockExtractFunction func(filePath string, startLine, endLine int, newFunctionName, receiver string) (*agents.ExtractedFunction, error)

// func init() {

// 	// Set default implementation to the actual function

// 	MockExtractFunction = agents.ExtractFunction

// }

func TestExtractFunctionTool_Execute(t *testing.T) {

	mockFSS := new(MockFileSystemService)

	tool := NewExtractFunctionTool(mockFSS)

	// originalExtractFunction := agents.ExtractFunction

	// t.Cleanup(func() {

	// 	agents.ExtractFunction = originalExtractFunction

	// })

	tests := []struct {
		name string

		args map[string]any

		setupMock func()

		expectedLLMContent string

		expectedReturnDisplay string

		expectedError string
	}{

		{

			name: "missing filePath argument",

			args: map[string]any{"startLine": float64(1), "endLine": float64(5), "newFunctionName": "NewFunc"},

			setupMock: func() {},

			expectedError: "missing or invalid 'filePath' argument",
		},

		{

			name: "invalid startLine argument",

			args: map[string]any{"filePath": "/tmp/file.go", "startLine": "invalid", "endLine": float64(5), "newFunctionName": "NewFunc"},

			setupMock: func() {},

			expectedError: "missing or invalid 'startLine' argument",
		},

		// {

		// 	name:          "extract function fails",

		// 	args:          map[string]any{"filePath": "/tmp/file.go", "startLine": float64(1), "endLine": float64(5), "newFunctionName": "NewFunc"},

		// 	setupMock: func() {

		// 		agents.ExtractFunction = func(filePath string, startLine, endLine int, newFunctionName, receiver string) (*agents.ExtractedFunction, error) {

		// 			return nil, fmt.Errorf("extraction error")

		// 		}

		// 	},

		// 	expectedError: "failed to extract function: extraction error",

		// },

		// {

		// 	name: "write file fails",

		// 	args: map[string]any{"filePath": "/tmp/file.go", "startLine": float64(1), "endLine": float64(5), "newFunctionName": "NewFunc"},

		// 	setupMock: func() {

		// 		agents.ExtractFunction = func(filePath string, startLine, endLine int, newFunctionName, receiver string) (*agents.ExtractedFunction, error) {

		// 			return &agents.ExtractedFunction{NewCode: "package main\nfunc NewFunc() {}"},

		// 		}

		// 		mockFSS.On("WriteFile", "/tmp/file.go", "package main\nfunc NewFunc() {}").Return(fmt.Errorf("write error")).Once()

		// 	},

		// 	expectedError: "failed to write extracted code to file: write error",

		// },

		// {

		// 	name: "successful extraction",

		// 	args: map[string]any{"filePath": "/tmp/file.go", "startLine": float64(1), "endLine": float64(5), "newFunctionName": "NewFunc", "receiver": "r *Receiver"},

		// 	setupMock: func() {

		// 		agents.ExtractFunction = func(filePath string, startLine, endLine int, newFunctionName, receiver string) (*agents.ExtractedFunction, error) {

		// 			assert.Equal(t, "/tmp/file.go", filePath)

		// 			assert.Equal(t, 1, startLine)

		// 			assert.Equal(t, 5, endLine)

		// 			assert.Equal(t, "NewFunc", newFunctionName)

		// 			assert.Equal(t, "r *Receiver", receiver)

		// 			return &agents.ExtractedFunction{NewCode: "package main\nfunc (r *Receiver) NewFunc() {}"},

		// 		}

		// 		mockFSS.On("WriteFile", "/tmp/file.go", "package main\nfunc (r *Receiver) NewFunc() {}").Return(nil).Once()

		// 	},

		// 	expectedLLMContent:    "Successfully extracted function and wrote to file. New code:\n```go\npackage main\nfunc (r *Receiver) NewFunc() {}\n```",

		// 	expectedReturnDisplay: "Successfully extracted function and wrote to file. New code:\n```go\npackage main\nfunc (r *Receiver) NewFunc() {}\n```",

		// },

	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFSS.Calls = []mock.Call{} // Reset calls for each test
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
				mockFSS.AssertExpectations(t)
			}
		})
	}
}
