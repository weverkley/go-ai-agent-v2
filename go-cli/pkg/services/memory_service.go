package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const (
	memoryFile = ".gemini/memory.json"
)

// MemoryService manages the AI's long-term memory.
type MemoryService struct {
	mu      sync.RWMutex
	content string
	baseDir string // Base directory to resolve memoryFile
}

// NewMemoryService creates a new MemoryService instance.
func NewMemoryService(baseDir string) *MemoryService {
	ms := &MemoryService{
		baseDir: baseDir,
	}
	_ = ms.Load() // Ignore error for initial load
	return ms
}

// GetMemory returns the current memory content.
func (ms *MemoryService) GetMemory() string {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.content
}

// SetMemory sets the memory content.
func (ms *MemoryService) SetMemory(content string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.content = content
	return ms.Save()
}

// ClearMemory clears the memory content.
func (ms *MemoryService) ClearMemory() error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.content = ""
	return ms.Save()
}

// Save persists the current memory content to a file.
func (ms *MemoryService) Save() error {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	filePath := filepath.Join(ms.baseDir, memoryFile)
	data, err := json.MarshalIndent(ms.content, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal memory content: %w", err)
	}

	err = os.MkdirAll(filepath.Dir(filePath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory for memory file: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write memory file: %w", err)
	}
	return nil
}

// Load loads memory content from a file.
func (ms *MemoryService) Load() error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	filePath := filepath.Join(ms.baseDir, memoryFile)
	data, err := os.ReadFile(filePath)
	if os.IsNotExist(err) {
		ms.content = "" // File doesn't exist, start with empty
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to read memory file: %w", err)
	}

	if err := json.Unmarshal(data, &ms.content); err != nil {
		return fmt.Errorf("failed to unmarshal memory content: %w", err)
	}
	return nil
}