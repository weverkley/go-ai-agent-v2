package extension

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const (
	settingsFile = ".gemini/settings.json"
)

// Extension represents a Gemini CLI extension.
type Extension struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	// Add other extension properties as needed
}

// ExtensionManager manages Gemini CLI extensions.
type Manager struct {
	mu        sync.RWMutex
	extensions map[string]*Extension
	baseDir   string // Base directory to resolve settingsFile
}

// NewManager creates a new ExtensionManager instance.
func NewManager(baseDir string) *Manager {
	em := &Manager{
		extensions: make(map[string]*Extension),
		baseDir:    baseDir,
	}
	// Attempt to load existing extension statuses on creation
	_ = em.LoadExtensionStatus() // Ignore error for initial load
	return em
}

// ListExtensions returns a list of all managed extensions.
func (em *Manager) ListExtensions() []*Extension {
	em.mu.RLock()
	defer em.mu.RUnlock()

	list := make([]*Extension, 0, len(em.extensions))
	for _, ext := range em.extensions {
		list = append(list, ext)
	}
	return list
}

// RegisterExtension registers a new extension with the manager.
func (em *Manager) RegisterExtension(ext *Extension) {
	em.mu.Lock()
	defer em.mu.Unlock()
	em.extensions[ext.Name] = ext
}

// EnableExtension enables a specific extension.
func (em *Manager) EnableExtension(name string) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	ext, ok := em.extensions[name]
	if !ok {
		return fmt.Errorf("extension '%s' not found", name)
	}
	ext.Enabled = true
	return em.SaveExtensionStatus()
}

// DisableExtension disables a specific extension.
func (em *Manager) DisableExtension(name string) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	ext, ok := em.extensions[name]
	if !ok {
		return fmt.Errorf("extension '%s' not found", name)
	}
	ext.Enabled = false
	return em.SaveExtensionStatus()
}

// SaveExtensionStatus persists the current extension statuses to a file.
func (em *Manager) SaveExtensionStatus() error {
	em.mu.RLock()
	defer em.mu.RUnlock()

	filePath := filepath.Join(em.baseDir, settingsFile)
	data, err := json.MarshalIndent(em.extensions, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal extension statuses: %w", err)
	}

	// Ensure the parent directory exists
	err = os.MkdirAll(filepath.Dir(filePath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory for settings file: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}
	return nil
}

// LoadExtensionStatus loads extension statuses from a file.
func (em *Manager) LoadExtensionStatus() error {
	em.mu.Lock()
	defer em.mu.Unlock()

	filePath := filepath.Join(em.baseDir, settingsFile)
	data, err := os.ReadFile(filePath)
	if os.IsNotExist(err) {
		em.extensions = make(map[string]*Extension) // File doesn't exist, start with empty
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to read settings file: %w", err)
	}

	var loadedExtensions map[string]*Extension
	if err := json.Unmarshal(data, &loadedExtensions); err != nil {
		return fmt.Errorf("failed to unmarshal extension statuses: %w", err)
	}
	em.extensions = loadedExtensions
	return nil
}