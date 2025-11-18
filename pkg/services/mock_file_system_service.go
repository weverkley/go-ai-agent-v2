package services

import (
	"os"

	"github.com/gobwas/glob"
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

// GetIgnorePatterns mocks the GetIgnorePatterns method.
func (m *MockFileSystemService) GetIgnorePatterns(searchDir string, respectGitIgnore, respectGoaiagentIgnore bool) ([]glob.Glob, error) {
	args := m.Called(searchDir, respectGitIgnore, respectGoaiagentIgnore)
	return args.Get(0).([]glob.Glob), args.Error(1)
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
