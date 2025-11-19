package extension

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sync"

	"go-ai-agent-v2/go-cli/pkg/services"
)

const (
	settingsFile = ".goaiagent/settings.json"
)

type Extension struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
}

type Manager struct {
	mu         sync.RWMutex
	extensions map[string]*Extension
	baseDir    string
	FSService  services.FileSystemService
	gitService services.GitService
}

func NewManager(baseDir string, fsService services.FileSystemService, gitService services.GitService) *Manager {
	em := &Manager{
		extensions: make(map[string]*Extension),
		baseDir:    baseDir,
		FSService:  fsService,
		gitService: gitService,
	}
	_ = em.LoadExtensionStatus()
	return em
}

func (em *Manager) ListExtensions() []*Extension {
	em.mu.RLock()
	defer em.mu.RUnlock()

	list := make([]*Extension, 0, len(em.extensions))
	for _, ext := range em.extensions {
		list = append(list, ext)
	}
	return list
}

func (em *Manager) RegisterExtension(ext *Extension) {
	em.extensions[ext.Name] = ext
}

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

func (em *Manager) saveExtensionStatusLocked() error {
	filePath := filepath.Join(em.baseDir, settingsFile)
	data, err := json.MarshalIndent(em.extensions, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal extension statuses: %w", err)
	}

	err = em.FSService.MkdirAll(filepath.Dir(filePath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory for settings file: %w", err)
	}

	if err := em.FSService.WriteFile(filePath, string(data)); err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}
	return nil
}

func (em *Manager) SaveExtensionStatus() error {
	return em.saveExtensionStatusLocked()
}

func (em *Manager) LoadExtensionStatus() error {
	em.mu.Lock()
	defer em.mu.Unlock()

	filePath := filepath.Join(em.baseDir, settingsFile)
	exists, err := em.FSService.PathExists(filePath)
	if err != nil {
		return fmt.Errorf("failed to check for settings file: %w", err)
	}
	if !exists {
		em.extensions = make(map[string]*Extension)
		if err := em.FSService.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for settings file: %w", err)
		}
		if err := em.FSService.WriteFile(filePath, "{}"); err != nil {
			return fmt.Errorf("failed to write initial empty settings file: %w", err)
		}
		return nil
	}

	dataStr, err := em.FSService.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read settings file: %w", err)
	}
	data := []byte(dataStr)

	if len(data) == 0 {
		em.extensions = make(map[string]*Extension)
		return nil
	}

	var loadedExtensions map[string]*Extension
	if err := json.Unmarshal(data, &loadedExtensions); err != nil {
		return fmt.Errorf("failed to unmarshal extension statuses: %w", err)
	}
	em.extensions = loadedExtensions
	return nil
}

func (em *Manager) InstallOrUpdateExtension(metadata ExtensionInstallMetadata, force bool) (string, error) {
	em.mu.Lock()
	defer em.mu.Unlock()

	tempPath := filepath.Join(em.baseDir, ".goaiagent", "temp_extensions", filepath.Base(metadata.Source))

	isGitRepo := metadata.Type == "git"

	if isGitRepo {
		defer func() {
			if err := em.FSService.RemoveAll(tempPath); err != nil {
				fmt.Printf("Warning: failed to clean up temporary extension path %s: %v\n", tempPath, err)
			}
		}()
	}

	if isGitRepo {
		if err := em.gitService.Clone(metadata.Source, tempPath, metadata.Ref); err != nil {
			return "", fmt.Errorf("failed to clone git repository to temp path: %w", err)
		}
	} else {
		tempPath = metadata.Source
	}

	manifestPath := filepath.Join(tempPath, "goaiagent-extension.json")
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

	finalExtensionPath := filepath.Join(em.baseDir, ".goaiagent", "extensions", manifest.Name)
	if isGitRepo {
		if force {
			_ = em.FSService.RemoveAll(finalExtensionPath)
		}
		if err := em.FSService.Rename(tempPath, finalExtensionPath); err != nil {
			return "", fmt.Errorf("failed to move cloned extension from temp to final path: %w", err)
		}
	} else {
		if err := em.FSService.Symlink(metadata.Source, finalExtensionPath); err != nil {
			return "", fmt.Errorf("failed to create symlink for local extension: %w", err)
		}
	}

	ext := &Extension{
		Name:    manifest.Name,
		Enabled: true,
	}
	em.RegisterExtension(ext)
	if err := em.SaveExtensionStatus(); err != nil {
		return "", fmt.Errorf("failed to save extension status: %w", err)
	}

	return ext.Name, nil
}

func (em *Manager) UninstallExtension(name string, interactiveConsent bool) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	delete(em.extensions, name)
	if err := em.saveExtensionStatusLocked(); err != nil {
		return fmt.Errorf("failed to save extension status: %w", err)
	}

	extensionPath := filepath.Join(em.baseDir, ".goaiagent", "extensions", name)
	if exists, err := em.FSService.PathExists(extensionPath); err != nil {
		return fmt.Errorf("failed to check if extension directory exists: %w", err)
	} else if exists {
		if err := em.FSService.RemoveAll(extensionPath); err != nil {
			return fmt.Errorf("failed to remove extension directory: %w", err)
		}
	}

	return nil
}

func (em *Manager) UpdateExtension(name string) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	ext, ok := em.extensions[name]
	if !ok {
		return fmt.Errorf("extension '%s' not found", name)
	}

	extensionPath := filepath.Join(em.baseDir, ".goaiagent", "extensions", ext.Name)

	if err := em.gitService.Pull(extensionPath, ""); err != nil {
		return fmt.Errorf("failed to pull git repository for extension '%s': %w", name, err)
	}

	return nil
}

func (em *Manager) LinkExtension(path string) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	manifestPath := filepath.Join(path, "goaiagent-extension.json")
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

	finalExtensionPath := filepath.Join(em.baseDir, ".goaiagent", "extensions", manifest.Name)

	if err := em.FSService.Symlink(path, finalExtensionPath); err != nil {
		return fmt.Errorf("failed to create symlink for local extension: %w", err)
	}

	ext := &Extension{
		Name:    manifest.Name,
		Enabled: true,
	}
	em.RegisterExtension(ext)
	if err := em.SaveExtensionStatus(); err != nil {
		return fmt.Errorf("failed to save extension status: %w", err)
	}

	return nil
}
