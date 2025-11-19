package tools

import (
	"context"
	"fmt"
	"os"
	"testing"

	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/stretchr/testify/assert"
)

// Mock functions for os operations
var (
	mockUserHomeDir func() (string, error)
	mockMkdirAll    func(path string, perm os.FileMode) error
	mockReadFile    func(name string) ([]byte, error)
	mockWriteFile   func(name string, data []byte, perm os.FileMode) error
	mockIsNotExist  func(err error) bool
)

func setupMemoryToolMocks() {
	// Assign mock implementations to the global variables in memory_tool.go
	osUserHomeDir = mockUserHomeDir
	osMkdirAll = mockMkdirAll
	osReadFile = mockReadFile
	osWriteFile = mockWriteFile
	osIsNotExist = mockIsNotExist
}

func teardownMemoryToolMocks() {
	// Restore original os functions
	osUserHomeDir = os.UserHomeDir
	osMkdirAll = os.MkdirAll
	osReadFile = os.ReadFile
	osWriteFile = os.WriteFile
	osIsNotExist = os.IsNotExist
}

func TestMemoryTool_Execute(t *testing.T) {
	tool := NewMemoryTool()

	tests := []struct {
		name          string
		args          map[string]any
		setupMock     func(tempDir string)
		expectedLLMContent string
		expectedReturnDisplay string
		expectedError string
	}{
		{
			name:          "missing fact argument",
			args:          map[string]any{},
			setupMock:     func(tempDir string) {},
			expectedError: "invalid or missing 'fact' argument",
		},
		{
			name:          "empty fact argument",
			args:          map[string]any{"fact": ""},
			setupMock:     func(tempDir string) {},
			expectedError: "invalid or missing 'fact' argument",
		},
		{
			name: "get user home dir fails",
			args: map[string]any{"fact": "test fact"},
			setupMock: func(tempDir string) {
				mockUserHomeDir = func() (string, error) { return "", fmt.Errorf("home dir error") }
			},
			expectedError: "failed to get user home directory: home dir error",
		},
		{
			name: "create memory directory fails",
			args: map[string]any{"fact": "test fact"},
			setupMock: func(tempDir string) {
				mockUserHomeDir = func() (string, error) { return tempDir, nil }
				mockMkdirAll = func(path string, perm os.FileMode) error { return fmt.Errorf("mkdir error") }
			},
			expectedError: "failed to create memory directory: mkdir error",
		},
		{
			name: "read memory file fails",
			args: map[string]any{"fact": "test fact"},
			setupMock: func(tempDir string) {
				mockUserHomeDir = func() (string, error) { return tempDir, nil }
				mockMkdirAll = func(path string, perm os.FileMode) error { return nil }
				mockReadFile = func(name string) ([]byte, error) { return nil, fmt.Errorf("read file error") }
				mockIsNotExist = func(err error) bool { return false }
			},
			expectedError: "failed to read memory file: read file error",
		},
		{
			name: "write memory file fails",
			args: map[string]any{"fact": "test fact"},
			setupMock: func(tempDir string) {
				mockUserHomeDir = func() (string, error) { return tempDir, nil }
				mockMkdirAll = func(path string, perm os.FileMode) error { return nil }
				mockReadFile = func(name string) ([]byte, error) { return nil, os.ErrNotExist }
				mockIsNotExist = func(err error) bool { return err == os.ErrNotExist }
				mockWriteFile = func(name string, data []byte, perm os.FileMode) error { return fmt.Errorf("write file error") }
			},
			expectedError: "failed to write memory file: write file error",
		},
		{
			name: "successful save - new file",
			args: map[string]any{"fact": "test fact"},
			setupMock: func(tempDir string) {
				mockUserHomeDir = func() (string, error) { return tempDir, nil }
				mockMkdirAll = func(path string, perm os.FileMode) error { return nil }
				mockReadFile = func(name string) ([]byte, error) { return nil, os.ErrNotExist }
				mockIsNotExist = func(err error) bool { return err == os.ErrNotExist }
				mockWriteFile = func(name string, data []byte, perm os.FileMode) error {
					expectedContent := fmt.Sprintf("%s\n- test fact\n", MEMORY_SECTION_HEADER)
					assert.Equal(t, expectedContent, string(data))
					return nil
				}
			},
			expectedLLMContent:    "Okay, I've remembered that: \"test fact\"",
			expectedReturnDisplay: "Okay, I've remembered that: \"test fact\"",
		},
		{
			name: "successful save - append to existing file",
			args: map[string]any{"fact": "another fact"},
			setupMock: func(tempDir string) {
				mockUserHomeDir = func() (string, error) { return tempDir, nil }
				mockMkdirAll = func(path string, perm os.FileMode) error { return nil }
				mockReadFile = func(name string) ([]byte, error) {
					return []byte(fmt.Sprintf("%s\n- existing fact\n", MEMORY_SECTION_HEADER)), nil
				}
				mockIsNotExist = func(err error) bool { return false }
				mockWriteFile = func(name string, data []byte, perm os.FileMode) error {
					expectedContent := fmt.Sprintf("%s\n- existing fact\n- another fact\n", MEMORY_SECTION_HEADER)
					assert.Equal(t, expectedContent, string(data))
					return nil
				}
			},
			expectedLLMContent:    "Okay, I've remembered that: \"another fact\"",
			expectedReturnDisplay: "Okay, I've remembered that: \"another fact\"",
		},
		{
			name: "fact already exists",
			args: map[string]any{"fact": "existing fact"},
			setupMock: func(tempDir string) {
				mockUserHomeDir = func() (string, error) { return tempDir, nil }
				mockMkdirAll = func(path string, perm os.FileMode) error { return nil }
				mockReadFile = func(name string) ([]byte, error) {
					return []byte(fmt.Sprintf("%s\n- existing fact\n", MEMORY_SECTION_HEADER)), nil
				}
				mockIsNotExist = func(err error) bool { return false }
				mockWriteFile = func(name string, data []byte, perm os.FileMode) error {
					t.Error("os.WriteFile should not be called when fact exists")
					return nil
				}
			},
			expectedLLMContent:    "I already know that: \"existing fact\"",
			expectedReturnDisplay: "I already know that: \"existing fact\"",
		},
		{
			name: "successful save - append to existing file without header",
			args: map[string]any{"fact": "another fact"},
			setupMock: func(tempDir string) {
				mockUserHomeDir = func() (string, error) { return tempDir, nil }
				mockMkdirAll = func(path string, perm os.FileMode) error { return nil }
				mockReadFile = func(name string) ([]byte, error) {
					return []byte("Some initial content.\n"), nil
				}
				mockIsNotExist = func(err error) bool { return false }
				mockWriteFile = func(name string, data []byte, perm os.FileMode) error {
					expectedContent := "Some initial content.\n\n" + MEMORY_SECTION_HEADER + "\n- another fact\n"
					assert.Equal(t, expectedContent, string(data))
					return nil
				}
			},
			expectedLLMContent:    "Okay, I've remembered that: \"another fact\"",
			expectedReturnDisplay: "Okay, I've remembered that: \"another fact\"",
		},
		{
			name: "successful save - fact with leading hyphen",
			args: map[string]any{"fact": "- test fact"},
			setupMock: func(tempDir string) {
				mockUserHomeDir = func() (string, error) { return tempDir, nil }
				mockMkdirAll = func(path string, perm os.FileMode) error { return nil }
				mockReadFile = func(name string) ([]byte, error) { return nil, os.ErrNotExist }
				mockIsNotExist = func(err error) bool { return err == os.ErrNotExist }
				mockWriteFile = func(name string, data []byte, perm os.FileMode) error {
					expectedContent := fmt.Sprintf("%s\n- test fact\n", MEMORY_SECTION_HEADER)
					assert.Equal(t, expectedContent, string(data))
					return nil
				}
			},
			expectedLLMContent:    "Okay, I've remembered that: \"- test fact\"",
			expectedReturnDisplay: "Okay, I've remembered that: \"- test fact\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			// First, set up the specific mocks for this test case.
			tt.setupMock(tempDir)

			// Then, apply the configured mocks to the package variables.
			setupMemoryToolMocks()
			t.Cleanup(teardownMemoryToolMocks)

			result, err := tool.Execute(context.Background(), tt.args)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				if result.Error != nil {
					assert.Equal(t, types.ToolErrorTypeExecutionFailed, result.Error.Type)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedLLMContent, result.LLMContent)
				assert.Equal(t, tt.expectedReturnDisplay, result.ReturnDisplay)
			}
		})
	}
}
