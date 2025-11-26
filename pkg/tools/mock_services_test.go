package tools

import (
	"context"
	"os"

	"github.com/gobwas/glob"
	"github.com/stretchr/testify/mock"
)

// MockWorkspaceService is a mock implementation of types.WorkspaceServiceIface
type MockWorkspaceService struct {
	mock.Mock
}

func (m *MockWorkspaceService) GetProjectRoot() string {
	args := m.Called()
	return args.String(0)
}

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
