package services

import (
	"bytes"
	"os/exec"
)

// ShellExecutionService provides functionality to execute shell commands.
type ShellExecutionService struct{}

// NewShellExecutionService creates a new instance of ShellExecutionService.
func NewShellExecutionService() *ShellExecutionService {
	return &ShellExecutionService{}
}

// ExecuteCommand executes a given shell command in the specified working directory.
// If workingDir is empty, the command is executed in the current process's working directory.
func (s *ShellExecutionService) ExecuteCommand(command string, workingDir string) (string, string, error) {
	cmd := exec.Command("bash", "-c", command)
	if workingDir != "" {
		cmd.Dir = workingDir
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	return stdout.String(), stderr.String(), err
}
