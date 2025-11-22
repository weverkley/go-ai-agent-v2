package extension_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ExtensionFlowTestSuite struct {
	suite.Suite
	tempDir string
	cliPath string
}

func (s *ExtensionFlowTestSuite) SetupSuite() {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "extension-flow-test")
	s.Require().NoError(err)
	s.tempDir = tempDir

	// Build the CLI
	cliPath := filepath.Join(s.tempDir, "go-ai-agent-v2")
	cmd := exec.Command("go", "build", "-o", cliPath, "../../main.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	s.Require().NoError(err, "Failed to build the CLI")
	s.cliPath = cliPath
}

func (s *ExtensionFlowTestSuite) TearDownSuite() {
	// a temporary directory for the test
	os.RemoveAll(s.tempDir)
}

func (s *ExtensionFlowTestSuite) TestExtensionFlow() {
	extName := "spoon-knife-ext"
	repoURL := "https://github.com/octocat/Spoon-Knife.git"
	clonedRepoPath := filepath.Join(s.tempDir, "cloned-spoon-knife")

	// 1. Manually clone the repository
	cloneCmd := exec.Command("git", "clone", repoURL, clonedRepoPath)
	cloneCmd.Stdout = os.Stdout
	cloneCmd.Stderr = os.Stderr
	err := cloneCmd.Run()
	s.Require().NoError(err, "Failed to clone Spoon-Knife repository")

	// 2. Create a dummy goaiagent-extension.json in the cloned repository
	manifestContent := fmt.Sprintf(`{"name": "%s", "version": "1.0.0", "description": "A test extension"}`, extName)
	manifestPath := filepath.Join(clonedRepoPath, "goaiagent-extension.json")
	err = os.WriteFile(manifestPath, []byte(manifestContent), 0644)
	s.Require().NoError(err, "Failed to create goaiagent-extension.json")

	// Ensure the target directory for the symlink exists
	targetExtensionDir := filepath.Join(s.tempDir, ".goaiagent", "extensions")
	err = os.MkdirAll(targetExtensionDir, 0755)
	s.Require().NoError(err, "Failed to create target extension directory")

	// 3. Install the extension from the local path
	installCmd := exec.Command(s.cliPath, "extensions", "install", clonedRepoPath)
	installCmd.Dir = s.tempDir
	installCmd.Env = append(os.Environ(), "GO_AI_AGENT_TEST_EXECUTOR=mock")
	output, err := installCmd.CombinedOutput()
	fmt.Println(string(output))
	s.Require().NoError(err, "Failed to install git extension. Output: %s", string(output))

	// 4. List extensions and verify the new extension is there
	listCmd := exec.Command(s.cliPath, "extensions", "list")
	listCmd.Dir = s.tempDir
	listCmd.Env = append(os.Environ(), "GO_AI_AGENT_TEST_EXECUTOR=mock")
	output, err = listCmd.CombinedOutput()
	s.Require().NoError(err, "Failed to list extensions. Output: %s", string(output))
	s.Contains(string(output), extName, "The installed extension should be in the list")

	// 5. Disable the extension
	disableCmd := exec.Command(s.cliPath, "extensions", "disable", extName)
	disableCmd.Dir = s.tempDir
	disableCmd.Env = append(os.Environ(), "GO_AI_AGENT_TEST_EXECUTOR=mock")
	output, err = disableCmd.CombinedOutput()
	s.Require().NoError(err, "Failed to disable extension. Output: %s", string(output))

	// 6. List extensions and verify the extension is disabled
	listCmd = exec.Command(s.cliPath, "extensions", "list")
	listCmd.Dir = s.tempDir
	listCmd.Env = append(os.Environ(), "GO_AI_AGENT_TEST_EXECUTOR=mock")
	output, err = listCmd.CombinedOutput()
	s.Require().NoError(err, "Failed to list extensions. Output: %s", string(output))
	s.Contains(string(output), fmt.Sprintf("- %s (Enabled: false)", extName), "The extension should be disabled")

	// 7. Enable the extension
	enableCmd := exec.Command(s.cliPath, "extensions", "enable", extName)
	enableCmd.Dir = s.tempDir
	enableCmd.Env = append(os.Environ(), "GO_AI_AGENT_TEST_EXECUTOR=mock")
	output, err = enableCmd.CombinedOutput()
	s.Require().NoError(err, "Failed to enable extension. Output: %s", string(output))

	// 8. List extensions and verify the extension is enabled
	listCmd = exec.Command(s.cliPath, "extensions", "list")
	listCmd.Dir = s.tempDir
	listCmd.Env = append(os.Environ(), "GO_AI_AGENT_TEST_EXECUTOR=mock")
	output, err = listCmd.CombinedOutput()
	s.Require().NoError(err, "Failed to list extensions. Output: %s", string(output))
	s.NotContains(string(output), "disabled", "The extension should be enabled")

	// 9. Uninstall the extension
	uninstallCmd := exec.Command(s.cliPath, "extensions", "uninstall", extName)
	uninstallCmd.Dir = s.tempDir
	uninstallCmd.Env = append(os.Environ(), "GO_AI_AGENT_TEST_EXECUTOR=mock")
	output, err = uninstallCmd.CombinedOutput()
	s.Require().NoError(err, "Failed to uninstall extension. Output: %s", string(output))

	// 10. List extensions and verify the extension is gone
	listCmd = exec.Command(s.cliPath, "extensions", "list")
	listCmd.Dir = s.tempDir
	listCmd.Env = append(os.Environ(), "GO_AI_AGENT_TEST_EXECUTOR=mock")
	output, err = listCmd.CombinedOutput()
	s.Require().NoError(err, "Failed to list extensions. Output: %s", string(output))
	s.NotContains(string(output), extName, "The uninstalled extension should not be in the list")
}

func TestExtensionFlowTestSuite(t *testing.T) {
	suite.Run(t, new(ExtensionFlowTestSuite))
}