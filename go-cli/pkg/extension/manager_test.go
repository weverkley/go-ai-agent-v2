package extension

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"go-ai-agent-v2/go-cli/pkg/services" // Add services import

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockFileSystemService is a mock implementation of services.FileSystemService
type MockFileSystemService struct {
	mock.Mock
}

func (m *MockFileSystemService) ListDirectory(dirPath string, ignorePatterns []string, respectGitIgnore, respectGeminiIgnore bool) ([]string, error) {
	args := m.Called(dirPath, ignorePatterns, respectGitIgnore, respectGeminiIgnore)
	return args.Get(0).([]string), args.Error(1)
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

func (m *MockGitService) Clone(url string, directory string, ref string) error {
	args := m.Called(url, directory, ref)
	return args.Error(0)
}

// ManagerTestSuite is the test suite for the ExtensionManager
type ManagerTestSuite struct {
	suite.Suite
	manager *Manager
	mockFs  *MockFileSystemService
	mockGit *MockGitService
	tempDir string
}

func (s *ManagerTestSuite) SetupTest() {
	s.tempDir = s.T().TempDir()
	s.mockFs = new(MockFileSystemService)
	s.mockGit = new(MockGitService)
	var gitService services.GitService = s.mockGit

	// Mock os.ReadFile and os.WriteFile for settings.json
	s.mockFs.On("JoinPaths", mock.AnythingOfType("[]string")).Return("mocked/path/settings.json").Maybe()
	s.mockFs.On("CreateDirectory", filepath.Join(s.tempDir, ".gemini")).Return(nil).Maybe()
	s.mockFs.On("WriteFile", filepath.Join(s.tempDir, ".gemini", "settings.json"), mock.Anything).Return(nil).Once()
	// Mock ReadFile for initial LoadExtensionStatus call and subsequent calls
	s.mockFs.On("ReadFile", filepath.Join(s.tempDir, ".gemini", "settings.json")).Return("{}", nil).Once() // Specific mock for settings.json
	s.mockFs.On("MkdirAll", mock.Anything, mock.Anything).Return(nil).Maybe() // Added this as well, as it's called in SaveExtensionStatus
	s.mockFs.On("PathExists", filepath.Join(s.tempDir, ".gemini")).Return(true, nil).Once() // Mock PathExists for .gemini directory
	s.mockFs.On("PathExists", filepath.Join(s.tempDir, ".gemini", "settings.json")).Return(true, nil).Once() // Mock PathExists for settings.json

	s.manager = NewManager(s.tempDir, s.mockFs, gitService)
}

func (s *ManagerTestSuite) TearDownTest() {
	s.mockFs.AssertExpectations(s.T())
	s.mockGit.AssertExpectations(s.T())
	os.RemoveAll(s.tempDir)
}

func TestManagerTestSuite(t *testing.T) {
	suite.Run(t, new(ManagerTestSuite))
}

// TestInstallOrUpdateExtension_Git tests installing a git extension
func (s *ManagerTestSuite) TestInstallOrUpdateExtension_Git() {
	extName := "test-git-ext"
	source := "https://github.com/user/repo.git"
	ref := "main"
	extensionPath := filepath.Join(s.tempDir, ".gemini", "extensions", extName)
	manifestContent := fmt.Sprintf(`{"name": "%s"}`, extName)

	tempPath := filepath.Join(s.tempDir, ".gemini", "temp_extensions", filepath.Base(source))
	s.mockFs.On("PathExists", extensionPath).Return(false, nil).Once()
	s.mockGit.On("Clone", source, tempPath, ref).Return(nil).Once() // Clone to tempPath
	s.mockFs.On("ReadFile", filepath.Join(tempPath, "gemini-extension.json")).Return(manifestContent, nil).Once() // Read manifest from tempPath
	s.mockFs.On("Rename", tempPath, extensionPath).Return(nil).Once() // Rename from tempPath to final path
	s.mockFs.On("WriteFile", mock.Anything, mock.Anything).Return(nil).Once() // For saving status
	s.mockFs.On("PathExists", mock.Anything).Return(false, nil).Once() // Add PathExists mock for other calls
	s.mockFs.On("WriteFile", mock.Anything, mock.Anything).Return(nil).Maybe() // Add WriteFile mock for other calls

	metadata := ExtensionInstallMetadata{
		Source: source,
		Type:   "git",
		Ref:    ref,
	}

	name, err := s.manager.InstallOrUpdateExtension(metadata, false)
	s.NoError(err)
	s.Equal(extName, name)

	ext := s.manager.extensions[extName]
	s.NotNil(ext)
	s.True(ext.Enabled)
}

// TestInstallOrUpdateExtension_Local tests installing a local extension
func (s *ManagerTestSuite) TestInstallOrUpdateExtension_Local() {
	extName := "test-local-ext"
	source := "/path/to/local/ext"
	extensionPath := filepath.Join(s.tempDir, ".gemini", "extensions", extName)
	manifestContent := fmt.Sprintf(`{"name": "%s"}`, extName)

	s.mockFs.On("PathExists", extensionPath).Return(false, nil).Once()
	s.mockFs.On("ReadFile", filepath.Join(source, "gemini-extension.json")).Return(manifestContent, nil).Once() // Read manifest from source
	s.mockFs.On("Symlink", source, extensionPath).Return(nil).Once() // Symlink from source to final path
	s.mockFs.On("WriteFile", mock.Anything, mock.Anything).Return(nil).Once() // For saving status
	s.mockFs.On("PathExists", mock.Anything).Return(false, nil).Once() // Add PathExists mock for other calls
	s.mockFs.On("WriteFile", mock.Anything, mock.Anything).Return(nil).Maybe() // Add WriteFile mock for other calls

	metadata := ExtensionInstallMetadata{
		Source: source,
		Type:   "local",
	}

	name, err := s.manager.InstallOrUpdateExtension(metadata, false)
	s.NoError(err)
	s.Equal(extName, name)

	ext := s.manager.extensions[extName]
	s.NotNil(ext)
	s.True(ext.Enabled)
}

// TestUninstallExtension tests uninstalling an extension
func (s *ManagerTestSuite) TestUninstallExtension() {
	extName := "test-ext-to-uninstall"
	extensionPath := filepath.Join(s.tempDir, ".gemini", "extensions", extName)

	// Register a dummy extension first
	s.manager.RegisterExtension(&Extension{Name: extName, Enabled: true})
	s.manager.SaveExtensionStatus() // Persist it

	s.mockFs.On("PathExists", extensionPath).Return(true, nil).Once() // For os.RemoveAll check
	s.mockFs.On("RemoveAll", extensionPath).Return(nil).Once()
	s.mockFs.On("WriteFile", mock.Anything, mock.Anything).Return(nil).Once() // For saving status
	s.mockFs.On("PathExists", mock.Anything).Return(false, nil).Once() // Add PathExists mock for other calls
	s.mockFs.On("WriteFile", mock.Anything, mock.Anything).Return(nil).Once() // Add WriteFile mock for other calls

	err := s.manager.UninstallExtension(extName, false)
	s.NoError(err)

	_, ok := s.manager.extensions[extName]
	s.False(ok)
}

// TestUpdateExtension tests updating an extension
func (s *ManagerTestSuite) TestUpdateExtension() {
	extName := "test-ext-to-update"
	extensionPath := filepath.Join(s.tempDir, ".gemini", "extensions", extName)

	// Register a dummy extension first
	s.manager.RegisterExtension(&Extension{Name: extName, Enabled: true})
	s.manager.SaveExtensionStatus() // Persist it

	s.mockGit.On("Pull", extensionPath, "").Return(nil).Once()
	s.mockFs.On("WriteFile", mock.Anything, mock.Anything).Return(nil).Once() // For saving status
	s.mockFs.On("PathExists", mock.Anything).Return(false, nil).Maybe() // Add PathExists mock for other calls
	s.mockFs.On("WriteFile", mock.Anything, mock.Anything).Return(nil).Once() // Add WriteFile mock for other calls

	err := s.manager.UpdateExtension(extName)
	s.NoError(err)
}

// TestLinkExtension tests linking a local extension
func (s *ManagerTestSuite) TestLinkExtension() {
	extName := "test-linked-ext"
	source := "/path/to/local/ext"
	extensionPath := filepath.Join(s.tempDir, ".gemini", "extensions", extName)
	manifestContent := fmt.Sprintf(`{"name": "%s"}`, extName)

	s.mockFs.On("ReadFile", filepath.Join(source, "gemini-extension.json")).Return(manifestContent, nil).Once() // Read manifest from source
	s.mockFs.On("Symlink", source, extensionPath).Return(nil).Once() // Symlink from source to final path
	s.mockFs.On("WriteFile", mock.Anything, mock.Anything).Return(nil).Once() // For saving status
	s.mockFs.On("PathExists", mock.Anything).Return(false, nil).Maybe() // Add PathExists mock for other calls
	s.mockFs.On("WriteFile", mock.Anything, mock.Anything).Return(nil).Once() // Add WriteFile mock for other calls

	err := s.manager.LinkExtension(source)
	s.NoError(err)

	ext := s.manager.extensions[extName]
	s.NotNil(ext)
	s.True(ext.Enabled)
}
