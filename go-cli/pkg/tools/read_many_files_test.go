package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockFileInfo is a mock implementation of os.FileInfo.
type MockFileInfo struct {
	mock.Mock
	name  string
	isDir bool
}

func (m *MockFileInfo) Name() string       { return m.name }
func (m *MockFileInfo) Size() int64        { return 0 }
func (m *MockFileInfo) Mode() os.FileMode  { return 0 }
func (m *MockFileInfo) ModTime() time.Time { return time.Now() }
func (m *MockFileInfo) IsDir() bool        { return m.isDir }
func (m *MockFileInfo) Sys() any           { return nil }

// MockWalkFunc is a type for mocking filepath.WalkFunc.
type MockWalkFunc func(path string, info os.FileInfo, err error) error

// MockFileSystemService for ReadManyFilesTool
type MockFileSystemServiceReadMany struct {
	mock.Mock
}

func (m *MockFileSystemServiceReadMany) ListDirectory(dirPath string, ignorePatterns []string, respectGitIgnore, respectGeminiIgnore bool) ([]string, error) {
	args := m.Called(dirPath, ignorePatterns, respectGitIgnore, respectGeminiIgnore)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockFileSystemServiceReadMany) PathExists(path string) (bool, error) {
	args := m.Called(path)
	return args.Bool(0), args.Error(1)
}

func (m *MockFileSystemServiceReadMany) IsDirectory(path string) (bool, error) {
	args := m.Called(path)
	return args.Bool(0), args.Error(1)
}

func (m *MockFileSystemServiceReadMany) ReadFile(filePath string) (string, error) {
	args := m.Called(filePath)
	return args.String(0), args.Error(1)
}

func (m *MockFileSystemServiceReadMany) WriteFile(filePath, content string) error {
	args := m.Called(filePath, content)
	return args.Error(0)
}

func (m *MockFileSystemServiceReadMany) CreateDirectory(path string) error {
	args := m.Called(path)
	return args.Error(0)
}

func (m *MockFileSystemServiceReadMany) CopyDirectory(src, dst string) error {
	args := m.Called(src, dst)
	return args.Error(0)
}

func (m *MockFileSystemServiceReadMany) JoinPaths(elem ...string) string {
	args := m.Called(elem)
	return args.String(0)
}

func TestReadManyFilesTool_Execute(t *testing.T) {

	tests := []struct {
		name          string
		args          map[string]any
		setup         func(t *testing.T, tempDir string)
		expectedLLMContent  string
		expectedReturnDisplay string
		expectedError string
	}{
		{
			name:          "missing paths argument",
			args:          map[string]any{},
			setup:         func(t *testing.T, tempDir string) {},
			expectedError: "invalid or missing 'paths' argument",
		},
		{
			name:          "empty paths argument",
			args:          map[string]any{"paths": []any{}},
			setup:         func(t *testing.T, tempDir string) {},
			expectedError: "invalid or missing 'paths' argument",
		},
		{
			name: "single file path",
			args: map[string]any{"paths": []any{"file1.txt"}},
			setup: func(t *testing.T, tempDir string) {
				filePath := filepath.Join(tempDir, "file1.txt")
				err := os.WriteFile(filePath, []byte("content of file1"), 0644)
				assert.NoError(t, err)
			},
			expectedLLMContent:    "--- %s ---\ncontent of file1\n\n--- End of content ---\n\n### ReadManyFiles Result\n\nSuccessfully read and concatenated content from **1 file(s)**.\n\n**Processed Files:**\n- `%s`\n",
			expectedReturnDisplay: "### ReadManyFiles Result\n\nSuccessfully read and concatenated content from **1 file(s)**.\n\n**Processed Files:**\n- `%s`\n",
		},
		{
			name: "multiple file paths",
			args: map[string]any{"paths": []any{"file1.txt", "file2.txt"}},
			setup: func(t *testing.T, tempDir string) {
				filePath1 := filepath.Join(tempDir, "file1.txt")
				filePath2 := filepath.Join(tempDir, "file2.txt")
				err := os.WriteFile(filePath1, []byte("content of file1"), 0644)
				assert.NoError(t, err)
				err = os.WriteFile(filePath2, []byte("content of file2"), 0644)
				assert.NoError(t, err)
			},
			expectedLLMContent:    "--- %s ---\ncontent of file1\n--- %s ---\ncontent of file2\n\n--- End of content ---\n\n### ReadManyFiles Result\n\nSuccessfully read and concatenated content from **2 file(s)**.\n\n**Processed Files:**\n- `%s`\n- `%s`\n",
			expectedReturnDisplay: "### ReadManyFiles Result\n\nSuccessfully read and concatenated content from **2 file(s)**.\n\n**Processed Files:**\n- `%s`\n- `%s`\n",
		},
		{
			name: "glob pattern",
			args: map[string]any{"paths": []any{"*.txt"}},
			setup: func(t *testing.T, tempDir string) {
				filePath1 := filepath.Join(tempDir, "file1.txt")
				filePath2 := filepath.Join(tempDir, "file2.log")
				filePath3 := filepath.Join(tempDir, "file3.txt")
				err := os.WriteFile(filePath1, []byte("content of file1"), 0644)
				assert.NoError(t, err)
				err = os.WriteFile(filePath2, []byte("content of file2"), 0644)
				assert.NoError(t, err)
				err = os.WriteFile(filePath3, []byte("content of file3"), 0644)
				assert.NoError(t, err)
			},
			expectedLLMContent:    "--- %s ---\ncontent of file1\n--- %s ---\ncontent of file3\n\n--- End of content ---\n\n### ReadManyFiles Result\n\nSuccessfully read and concatenated content from **2 file(s)**.\n\n**Processed Files:**\n- `%s`\n- `%s`\n",
			expectedReturnDisplay: "### ReadManyFiles Result\n\nSuccessfully read and concatenated content from **2 file(s)**.\n\n**Processed Files:**\n- `%s`\n- `%s`\n",
		},
		{
			name: "with include patterns",
			args: map[string]any{"paths": []any{"*.log"}, "include": []any{"*.txt"}},
			setup: func(t *testing.T, tempDir string) {
				filePath1 := filepath.Join(tempDir, "file1.txt")
				filePath2 := filepath.Join(tempDir, "file2.log")
				err := os.WriteFile(filePath1, []byte("content of file1"), 0644)
				assert.NoError(t, err)
				err = os.WriteFile(filePath2, []byte("content of file2"), 0644)
				assert.NoError(t, err)
			},
			expectedLLMContent:    "--- %s ---\ncontent of file1\n--- %s ---\ncontent of file2\n\n--- End of content ---\n\n### ReadManyFiles Result\n\nSuccessfully read and concatenated content from **2 file(s)**.\n\n**Processed Files:**\n- `%s`\n- `%s`\n",
			expectedReturnDisplay: "### ReadManyFiles Result\n\nSuccessfully read and concatenated content from **2 file(s)**.\n\n**Processed Files:**\n- `%s`\n- `%s`\n",
		},
		{
			name: "with exclude patterns",
			args: map[string]any{"paths": []any{"*.txt"}, "exclude": []any{"file1.txt"}},
			setup: func(t *testing.T, tempDir string) {
				filePath1 := filepath.Join(tempDir, "file1.txt")
				filePath2 := filepath.Join(tempDir, "file2.txt")
				err := os.WriteFile(filePath1, []byte("content of file1"), 0644)
				assert.NoError(t, err)
				err = os.WriteFile(filePath2, []byte("content of file2"), 0644)
				assert.NoError(t, err)
			},
			expectedLLMContent:    "--- %s ---\ncontent of file2\n\n--- End of content ---\n\n### ReadManyFiles Result\n\nSuccessfully read and concatenated content from **1 file(s)**.\n\n**Processed Files:**\n- `%s`\n",
			expectedReturnDisplay: "### ReadManyFiles Result\n\nSuccessfully read and concatenated content from **1 file(s)**.\n\n**Processed Files:**\n- `%s`\n",
		},
		{
			name: "recursive set to false",
			args: map[string]any{"paths": []any{"**/*.txt"}, "recursive": false},
			setup: func(t *testing.T, tempDir string) {
				filePath1 := filepath.Join(tempDir, "file1.txt")
				subDir := filepath.Join(tempDir, "sub")
				filePath2 := filepath.Join(subDir, "file2.txt")
				err := os.Mkdir(subDir, 0755)
				assert.NoError(t, err)
				err = os.WriteFile(filePath1, []byte("content of file1"), 0644)
				assert.NoError(t, err)
				err = os.WriteFile(filePath2, []byte("content of file2"), 0644)
				assert.NoError(t, err)
			},
			expectedLLMContent:    "--- %s ---\ncontent of file1\n\n--- End of content ---\n\n### ReadManyFiles Result\n\nSuccessfully read and concatenated content from **1 file(s)**.\n\n**Processed Files:**\n- `%s`\n",
			expectedReturnDisplay: "### ReadManyFiles Result\n\nSuccessfully read and concatenated content from **1 file(s)**.\n\n**Processed Files:**\n- `%s`\n",
		},
		{
			name: "useDefaultExcludes set to false",
			args: map[string]any{"paths": []any{"**/*.txt"}, "useDefaultExcludes": false},
			setup: func(t *testing.T, tempDir string) {
				filePath1 := filepath.Join(tempDir, "file1.txt")
				gitDir := filepath.Join(tempDir, ".git")
				filePath2 := filepath.Join(gitDir, "config.txt")
				err := os.Mkdir(gitDir, 0755)
				assert.NoError(t, err)
				err = os.WriteFile(filePath1, []byte("content of file1"), 0644)
				assert.NoError(t, err)
				err = os.WriteFile(filePath2, []byte("content of git config"), 0644)
				assert.NoError(t, err)
			},
			expectedLLMContent:    "--- %s ---\ncontent of file1\n--- %s ---\ncontent of git config\n\n--- End of content ---\n\n### ReadManyFiles Result\n\nSuccessfully read and concatenated content from **2 file(s)**.\n\n**Processed Files:**\n- `%s`\n- `%s`\n",
			expectedReturnDisplay: "### ReadManyFiles Result\n\nSuccessfully read and concatenated content from **2 file(s)**.\n\n**Processed Files:**\n- `%s`\n- `%s`\n",
		},
		{
			name: "file read error",
			args: map[string]any{"paths": []any{"file1.txt"}},
			setup: func(t *testing.T, tempDir string) {
				filePath := filepath.Join(tempDir, "file1.txt")
				// Create file but make it unreadable
				err := os.WriteFile(filePath, []byte("content"), 0000) // No read permissions
				assert.NoError(t, err)
			},
			expectedLLMContent:    "\n--- End of content ---\n\n### ReadManyFiles Result\n\nSuccessfully read and concatenated content from **0 file(s)**.\n\n**Skipped 1 item(s):**\n- `%s` (Reason: failed to read: open %s: permission denied)\n",
			expectedReturnDisplay: "### ReadManyFiles Result\n\nSuccessfully read and concatenated content from **0 file(s)**.\n\n**Skipped 1 item(s):**\n- `%s` (Reason: failed to read: open %s: permission denied)\n",
		},
		{
			name: "no files found",
			args: map[string]any{"paths": []any{"*.nonexistent"}},
			setup: func(t *testing.T, tempDir string) {},
			expectedLLMContent:    "\n--- End of content ---\n\n### ReadManyFiles Result\n\nNo files were read and concatenated based on the criteria.\n",
			expectedReturnDisplay: "### ReadManyFiles Result\n\nNo files were read and concatenated based on the criteria.\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := NewReadManyFilesTool() // Initialize tool inside the test run
			tempDir := t.TempDir()
			currentDir, err := os.Getwd()
			assert.NoError(t, err)
			err = os.Chdir(tempDir)
			assert.NoError(t, err)
			defer os.Chdir(currentDir) // Restore original working directory

			tt.setup(t, tempDir)

			result, err := tool.Execute(tt.args)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)

				// Dynamically format expected strings with tempDir. Relative paths are used here.
				var formattedExpectedLLMContent string
				var formattedExpectedReturnDisplay string

				switch tt.name {
				case "single file path":
					formattedExpectedLLMContent = fmt.Sprintf(tt.expectedLLMContent, filepath.Join(tempDir, "file1.txt"), filepath.Join(tempDir, "file1.txt"))
					formattedExpectedReturnDisplay = fmt.Sprintf(tt.expectedReturnDisplay, filepath.Join(tempDir, "file1.txt"))
				case "multiple file paths":
					formattedExpectedLLMContent = fmt.Sprintf(tt.expectedLLMContent, filepath.Join(tempDir, "file1.txt"), filepath.Join(tempDir, "file2.txt"), filepath.Join(tempDir, "file1.txt"), filepath.Join(tempDir, "file2.txt"))
					formattedExpectedReturnDisplay = fmt.Sprintf(tt.expectedReturnDisplay, filepath.Join(tempDir, "file1.txt"), filepath.Join(tempDir, "file2.txt"))
				case "glob pattern":
					formattedExpectedLLMContent = fmt.Sprintf(tt.expectedLLMContent, filepath.Join(tempDir, "file1.txt"), filepath.Join(tempDir, "file3.txt"), filepath.Join(tempDir, "file1.txt"), filepath.Join(tempDir, "file3.txt"))
					formattedExpectedReturnDisplay = fmt.Sprintf(tt.expectedReturnDisplay, filepath.Join(tempDir, "file1.txt"), filepath.Join(tempDir, "file3.txt"))
				case "with include patterns":
					formattedExpectedLLMContent = fmt.Sprintf(tt.expectedLLMContent, filepath.Join(tempDir, "file1.txt"), filepath.Join(tempDir, "file2.log"), filepath.Join(tempDir, "file1.txt"), filepath.Join(tempDir, "file2.log"))
					formattedExpectedReturnDisplay = fmt.Sprintf(tt.expectedReturnDisplay, filepath.Join(tempDir, "file1.txt"), filepath.Join(tempDir, "file2.log"))
				case "with exclude patterns":
					formattedExpectedLLMContent = fmt.Sprintf(tt.expectedLLMContent, filepath.Join(tempDir, "file2.txt"), filepath.Join(tempDir, "file2.txt"))
					formattedExpectedReturnDisplay = fmt.Sprintf(tt.expectedReturnDisplay, filepath.Join(tempDir, "file2.txt"))
				case "recursive set to false":
					formattedExpectedLLMContent = fmt.Sprintf(tt.expectedLLMContent, filepath.Join(tempDir, "file1.txt"), filepath.Join(tempDir, "file1.txt"))
					formattedExpectedReturnDisplay = fmt.Sprintf(tt.expectedReturnDisplay, filepath.Join(tempDir, "file1.txt"))
				case "useDefaultExcludes set to false":
					formattedExpectedLLMContent = fmt.Sprintf(tt.expectedLLMContent, filepath.Join(tempDir, "file1.txt"), filepath.Join(tempDir, ".git", "config.txt"), filepath.Join(tempDir, "file1.txt"), filepath.Join(tempDir, ".git", "config.txt"))
					formattedExpectedReturnDisplay = fmt.Sprintf(tt.expectedReturnDisplay, filepath.Join(tempDir, "file1.txt"), filepath.Join(tempDir, ".git", "config.txt"))
				case "file read error":
					formattedExpectedLLMContent = fmt.Sprintf(tt.expectedLLMContent, filepath.Join(tempDir, "file1.txt"), filepath.Join(tempDir, "file1.txt"))
					formattedExpectedReturnDisplay = fmt.Sprintf(tt.expectedReturnDisplay, filepath.Join(tempDir, "file1.txt"), filepath.Join(tempDir, "file1.txt"))
				case "no files found":
					formattedExpectedLLMContent = tt.expectedLLMContent
					formattedExpectedReturnDisplay = tt.expectedReturnDisplay
				default:
					t.Fatalf("unknown test name: %s", tt.name)
				}

				assert.Equal(t, formattedExpectedLLMContent, result.LLMContent)
				assert.Equal(t, formattedExpectedReturnDisplay, result.ReturnDisplay)
			}
		})
	}
}
