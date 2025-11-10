package extension

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"go-ai-agent-v2/go-cli/pkg/services" // Add services import
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
	FSService services.FileSystemService // Add FileSystemService
	gitService services.GitService // Add GitService
}

// NewManager creates a new ExtensionManager instance.
func NewManager(baseDir string, fsService services.FileSystemService, gitService services.GitService) *Manager {
	em := &Manager{
		extensions: make(map[string]*Extension),
		baseDir:    baseDir,
		FSService: fsService,
		gitService: gitService,
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

// saveExtensionStatusLocked persists the current extension statuses to a file.
// It assumes the caller already holds the appropriate lock.
func (em *Manager) saveExtensionStatusLocked() error {
	filePath := filepath.Join(em.baseDir, settingsFile)
	data, err := json.MarshalIndent(em.extensions, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal extension statuses: %w", err)
	}

	// Ensure the parent directory exists
	err = em.FSService.MkdirAll(filepath.Dir(filePath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory for settings file: %w", err)
	}

	if err := em.FSService.WriteFile(filePath, string(data)); err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}
	return nil
}

// SaveExtensionStatus persists the current extension statuses to a file.
// This is the public method that acquires a read lock.
func (em *Manager) SaveExtensionStatus() error {
	em.mu.RLock()
	defer em.mu.RUnlock()
	return em.saveExtensionStatusLocked()
}

// LoadExtensionStatus loads extension statuses from a file.
func (em *Manager) LoadExtensionStatus() error {
	em.mu.Lock()
	defer em.mu.Unlock()

	filePath := filepath.Join(em.baseDir, settingsFile)
	dataStr, err := em.FSService.ReadFile(filePath)
	if os.IsNotExist(err) {
		em.extensions = make(map[string]*Extension) // File doesn't exist, start with empty
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to read settings file: %w", err)
	}
	data := []byte(dataStr)

	var loadedExtensions map[string]*Extension
	if err := json.Unmarshal(data, &loadedExtensions); err != nil {
		return fmt.Errorf("failed to unmarshal extension statuses: %w", err)
	}
	em.extensions = loadedExtensions
	return nil
}

// InstallOrUpdateExtension installs or updates an extension.
func (em *Manager) InstallOrUpdateExtension(metadata ExtensionInstallMetadata, force bool) (string, error) {
	em.mu.Lock()
	defer em.mu.Unlock()

	// Determine a temporary path for cloning/copying to read the manifest
	tempPath := filepath.Join(em.baseDir, ".gemini", "temp_extensions", filepath.Base(metadata.Source))

	// Perform clone/copy to temp path
	switch metadata.Type {
	case "git":
		// Clone to temp path
		if err := em.gitService.Clone(metadata.Source, tempPath, metadata.Ref); err != nil {
			return "", fmt.Errorf("failed to clone git repository to temp path: %w", err)
		}
	case "local":
		// For now, we'll assume direct manifest reading from source for local
		// and then symlink to final destination.
		// This part needs careful thought if local source is not a directory with manifest.
		// For simplicity, let's assume local source is a valid extension directory.
		tempPath = metadata.Source // Use source directly for manifest reading
	default:
		return "", fmt.Errorf("unsupported extension type: %s", metadata.Type)
	}

	// Read manifest to get extension name
	manifestPath := filepath.Join(tempPath, "gemini-extension.json")
	manifestDataStr, err := em.FSService.ReadFile(manifestPath)
	if err != nil {
		return "", fmt.Errorf("failed to read extension manifest from %s: %w", manifestPath, err)
	}
	manifestData := []byte(manifestDataStr)

	var manifest struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return "", fmt.Errorf("failed to unmarshal extension manifest from %s: %w", manifestPath, err)
	}

	// Now that we have the name, construct the final extension path
	finalExtensionPath := filepath.Join(em.baseDir, ".gemini", "extensions", manifest.Name)

	// If it was a git clone to temp, move it to final path
	if metadata.Type == "git" {
		// Remove existing if force is true
		if force {
			_ = em.FSService.RemoveAll(finalExtensionPath)
		}
		if err := em.FSService.Rename(tempPath, finalExtensionPath); err != nil {
			return "", fmt.Errorf("failed to move cloned extension from temp to final path: %w", err)
		}
	} else if metadata.Type == "local" {
		// For local, create symlink from source to final path
		if err := em.FSService.Symlink(metadata.Source, finalExtensionPath); err != nil {
			return "", fmt.Errorf("failed to create symlink for local extension: %w", err)
		}
	}

	ext := &Extension{
		Name:    manifest.Name,
		Enabled: true, // Enable by default on install
	}
	em.RegisterExtension(ext)
	if err := em.SaveExtensionStatus(); err != nil {
		return "", fmt.Errorf("failed to save extension status: %w", err)
	}

	return ext.Name, nil
}

// UninstallExtension uninstalls an extension.
func (em *Manager) UninstallExtension(name string, interactiveConsent bool) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	ext, ok := em.extensions[name]
	if !ok {
		return fmt.Errorf("extension '%s' not found", name)
	}

	extensionPath := filepath.Join(em.baseDir, ".gemini", "extensions", ext.Name)

	// Remove the extension directory or symlink
	if err := em.FSService.RemoveAll(extensionPath); err != nil {
		return fmt.Errorf("failed to remove extension files: %w", err)
	}

	delete(em.extensions, name)
	if err := em.saveExtensionStatusLocked(); err != nil {
		return fmt.Errorf("failed to save extension status: %w", err)
	}

	return nil
}

// UpdateExtension updates a specific extension.
func (em *Manager) UpdateExtension(name string) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	ext, ok := em.extensions[name]
	if !ok {
		return fmt.Errorf("extension '%s' not found", name)
	}

	extensionPath := filepath.Join(em.baseDir, ".gemini", "extensions", ext.Name)

	// For now, only git extensions can be updated
	// In a more complex scenario, metadata would store the type
	// and source for update logic.
	// Assuming git for now.
	if err := em.gitService.Pull(extensionPath, ""); err != nil { // Empty ref for default branch
		return fmt.Errorf("failed to pull git repository for extension '%s': %w", name, err)
	}

	return nil
}

// LinkExtension links a local extension.
func (em *Manager) LinkExtension(path string) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	// Read manifest from the source path to get extension name
	manifestPath := filepath.Join(path, "gemini-extension.json")
	manifestDataStr, err := em.FSService.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to read extension manifest from %s: %w", manifestPath, err)
	}
	manifestData := []byte(manifestDataStr)

	var manifest struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return fmt.Errorf("failed to parse extension manifest from %s: %w", manifestPath, err)
	}

	// Now that we have the name, construct the final extension path
	finalExtensionPath := filepath.Join(em.baseDir, ".gemini", "extensions", manifest.Name)

	// Create symlink from source to final path
	if err := em.FSService.Symlink(path, finalExtensionPath); err != nil {
		return fmt.Errorf("failed to create symlink for local extension: %w", err)
	}

	ext := &Extension{
		Name:    manifest.Name,
		Enabled: true, // Enable by default on link
	}
	em.RegisterExtension(ext)
	if err := em.SaveExtensionStatus(); err != nil {
		return fmt.Errorf("failed to save extension status: %w", err)
	}

	return nil
}