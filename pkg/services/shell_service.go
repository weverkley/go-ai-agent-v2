package services

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"sync"
	"syscall"
)

// ShellExecutionService provides functionality to execute shell commands.
type ShellExecutionService struct {
	backgroundProcesses map[int]*os.Process
	mutex               sync.Mutex
}

// NewShellExecutionService creates a new instance of ShellExecutionService.
func NewShellExecutionService() *ShellExecutionService {
	return &ShellExecutionService{
		backgroundProcesses: make(map[int]*os.Process),
	}
}

// ExecuteCommand executes a given shell command in the specified working directory.
// It is a blocking call and can be cancelled via the provided context.
func (s *ShellExecutionService) ExecuteCommand(ctx context.Context, command string, workingDir string) (string, string, error) {
	cmd := exec.CommandContext(ctx, "bash", "-c", command)
	if workingDir != "" {
		cmd.Dir = workingDir
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	return stdout.String(), stderr.String(), err
}

// ExecuteCommandInBackground executes a command in the background and tracks its PID.
func (s *ShellExecutionService) ExecuteCommandInBackground(command string, workingDir string) (int, error) {
	cmd := exec.Command("bash", "-c", command)
	if workingDir != "" {
		cmd.Dir = workingDir
	}

	// Set a new process group ID to be able to kill the process and its children.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	err := cmd.Start()
	if err != nil {
		return -1, err
	}

	s.mutex.Lock()
	s.backgroundProcesses[cmd.Process.Pid] = cmd.Process
	s.mutex.Unlock()

	// We don't wait for the command to finish, just return the PID
	return cmd.Process.Pid, nil
}

// KillAllProcesses iterates over the tracked background processes and terminates them.
func (s *ShellExecutionService) KillAllProcesses() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for pid, process := range s.backgroundProcesses {
		// Use a negative PID to kill the entire process group.
		if err := syscall.Kill(-pid, syscall.SIGKILL); err == nil {
			process.Release()
			delete(s.backgroundProcesses, pid)
		}
	}
}
