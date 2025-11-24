package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const (
	workspaceStateFile = ".goaiagent/workspace.json"
)

// WorkspaceService manages workspace directories.
type WorkspaceService struct {
	mu          sync.RWMutex
	directories []string
	baseDir     string // Base directory to resolve workspaceStateFile
}

// NewWorkspaceService creates a new WorkspaceService instance.
func NewWorkspaceService(baseDir string) *WorkspaceService {
	ws := &WorkspaceService{
		directories: []string{},
		baseDir:     baseDir,
	}
	// Attempt to load existing directories on creation
	_ = ws.Load() // Ignore error for initial load, as file might not exist
	return ws
}

// AddDirectory adds a new directory to the workspace.
func (ws *WorkspaceService) AddDirectory(path string) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	// Check if the directory already exists
	for _, dir := range ws.directories {
		if dir == path {
			return nil // Already exists, no-op
		}
	}

	// Validate path exists and is a directory
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", path)
	}
	if err != nil {
		return fmt.Errorf("error checking directory %s: %w", path, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}

	ws.directories = append(ws.directories, path)
	return ws.Save()
}

// GetDirectories returns all directories in the workspace.
func (ws *WorkspaceService) GetDirectories() []string {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	// Return a copy to prevent external modification
	dirs := make([]string, len(ws.directories))
	copy(dirs, ws.directories)
	return dirs
}

// Save persists the current workspace directories to a file.
func (ws *WorkspaceService) Save() error {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	filePath := filepath.Join(ws.baseDir, workspaceStateFile)
	data, err := json.MarshalIndent(ws.directories, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal workspace directories: %w", err)
	}

	// Ensure the parent directory exists
	err = os.MkdirAll(filepath.Dir(filePath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory for workspace state file: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write workspace state file: %w", err)
	}
	return nil
}

// Load loads workspace directories from a file.
func (ws *WorkspaceService) Load() error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	filePath := filepath.Join(ws.baseDir, workspaceStateFile)
	data, err := os.ReadFile(filePath)
	if os.IsNotExist(err) {
		ws.directories = []string{} // File doesn't exist, start with empty
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to read workspace state file: %w", err)
	}

	if err := json.Unmarshal(data, &ws.directories); err != nil {
		return fmt.Errorf("failed to unmarshal workspace directories: %w", err)
	}
	return nil
}

// GetProjectRoot returns the base directory of the workspace (project root).
func (ws *WorkspaceService) GetProjectRoot() string {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	return ws.baseDir
}