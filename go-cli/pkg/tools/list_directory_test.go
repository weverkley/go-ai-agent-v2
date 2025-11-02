package tools

import (
	"fmt"
	"testing"

	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockFileSystemService is a mock implementation of FileSystemService.
type MockFileSystemService struct {
	mock.Mock
}

// ListDirectory mocks the ListDirectory method.
func (m *MockFileSystemService) ListDirectory(path string, ignorePatterns []string, respectGitIgnore, respectGeminiIgnore bool) ([]string, error) {
	args := m.Called(path, ignorePatterns, respectGitIgnore, respectGeminiIgnore)
	return args.Get(0).([]string), args.Error(1)
}

// PathExists mocks the PathExists method.
func (m *MockFileSystemService) PathExists(path string) (bool, error) {
	args := m.Called(path)
	return args.Bool(0), args.Error(1)
}

// IsDirectory mocks the IsDirectory method.
func (m *MockFileSystemService) IsDirectory(path string) (bool, error) {
	args := m.Called(path)
	return args.Bool(0), args.Error(1)
}

// ReadFile mocks the ReadFile method.
func (m *MockFileSystemService) ReadFile(filePath string) (string, error) {
	args := m.Called(filePath)
	return args.String(0), args.Error(1)
}

// WriteFile mocks the WriteFile method.
func (m *MockFileSystemService) WriteFile(filePath, content string) error {
	args := m.Called(filePath, content)
	return args.Error(0)
}

// CreateDirectory mocks the CreateDirectory method.
func (m *MockFileSystemService) CreateDirectory(path string) error {
	args := m.Called(path)
	return args.Error(0)
}

// CopyDirectory mocks the CopyDirectory method.
func (m *MockFileSystemService) CopyDirectory(src, dst string) error {
	args := m.Called(src, dst)
	return args.Error(0)
}

// JoinPaths mocks the JoinPaths method.
func (m *MockFileSystemService) JoinPaths(elem ...string) string {
	args := m.Called(elem)
	return args.String(0)
}

func TestListDirectoryTool_Execute(t *testing.T) {
	// Setup mock FileSystemService
	mockFSS := new(MockFileSystemService)

	// Create a ListDirectoryTool with the mock service
	tool := &ListDirectoryTool{
		BaseDeclarativeTool: *types.NewBaseDeclarativeTool("list_directory", "", "", types.KindOther, types.JsonSchemaObject{}, false, false, nil),
		fileSystemService:   mockFSS, // Assign the mock directly
	}

	tests := []struct {
		name                string
		args                map[string]any
		setupMock           func()
		expectedLLMContent  string
		expectedReturnDisplay string
		expectedError       string
	}{
		{
			name:          "missing path argument",
			args:          map[string]any{},
			setupMock:     func() {},
			expectedError: "missing or invalid 'path' argument",
		},
		{
			name:          "empty path argument",
			args:          map[string]any{"path": ""},
			setupMock:     func() {},
			expectedError: "missing or invalid 'path' argument",
		},
		{
			name: "successful listing - no ignores",
			args: map[string]any{"path": "/test/dir"},
			setupMock: func() {
				mockFSS.On("ListDirectory", "/test/dir", []string{}, true, true).Return([]string{"file1.txt", "subdir"}, nil).Once()
			},
			expectedLLMContent:    "file1.txt\nsubdir",
			expectedReturnDisplay: "Listed directory /test/dir: 2 items",
		},
		{
			name: "successful listing - with ignores",
			args: map[string]any{"path": "/test/dir", "ignore": []interface{}{"*.log"}},
			setupMock: func() {
				mockFSS.On("ListDirectory", "/test/dir", []string{"*.log"}, true, true).Return([]string{"file1.txt"}, nil).Once()
			},
			expectedLLMContent:    "file1.txt",
			expectedReturnDisplay: "Listed directory /test/dir: 1 items",
		},
		{
			name: "fileSystemService returns error",
			args: map[string]any{"path": "/test/dir"},
			setupMock: func() {
				mockFSS.On("ListDirectory", "/test/dir", []string{}, true, true).Return([]string{}, fmt.Errorf("permission denied")).Once()
			},
			expectedError: "failed to list directory /test/dir: permission denied",
		},
		{
			name: "successful listing - respect_git_ignore false",
			args: map[string]any{"path": "/test/dir", "respect_git_ignore": false},
			setupMock: func() {
				mockFSS.On("ListDirectory", "/test/dir", []string{}, false, true).Return([]string{"file1.txt", "ignored.log"}, nil).Once()
			},
			expectedLLMContent:    "file1.txt\nignored.log",
			expectedReturnDisplay: "Listed directory /test/dir: 2 items",
		},
		{
			name: "successful listing - respect_gemini_ignore false",
			args: map[string]any{"path": "/test/dir", "respect_gemini_ignore": false},
			setupMock: func() {
				mockFSS.On("ListDirectory", "/test/dir", []string{}, true, false).Return([]string{"file1.txt", "ignored.gemini"}, nil).Once()
			},
			expectedLLMContent:    "file1.txt\nignored.gemini",
			expectedReturnDisplay: "Listed directory /test/dir: 2 items",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() {
				mockFSS.AssertExpectations(t)
				mockFSS.Calls = []mock.Call{} // Reset calls for next test
			})
			tt.setupMock()

			result, err := tool.Execute(tt.args)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedLLMContent, result.LLMContent)
				assert.Equal(t, tt.expectedReturnDisplay, result.ReturnDisplay)
			}
		})
	}
}