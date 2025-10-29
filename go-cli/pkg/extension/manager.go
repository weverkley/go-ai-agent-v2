package extension

import (
	"encoding/json"
	"fmt"
	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/mcp"
	"go-ai-agent-v2/go-cli/pkg/services"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
)

// Extension represents a simplified extension structure.
type Extension struct {
	Name          string `json:"name"`
	Path          string
	Description   string                         `json:"description"`
	McpServers    map[string]mcp.MCPServerConfig `json:"mcpServers,omitempty"`
	InstallType   string
	InstallSource string
}

// ExtensionConfig represents the structure of gemini-extension.json
type ExtensionConfig struct {
	Name        string                         `json:"name"`
	Description string                         `json:"description"`
	McpServers  map[string]mcp.MCPServerConfig `json:"mcpServers,omitempty"`
	// Add other fields as needed from the JSON config
}

// ExtensionManager manages the loading and interaction with extensions.
type ExtensionManager struct {
	workspaceDir string
	settings     *config.Settings
	fsService    *services.FileSystemService // Add FileSystemService dependency
	gitService   *services.GitService        // Add GitService dependency
	// Add other dependencies like consent handlers
}

// NewExtensionManager creates a new instance of ExtensionManager.
func NewExtensionManager(workspaceDir string) *ExtensionManager {
	settings := config.LoadSettings(workspaceDir)
	return &ExtensionManager{
		workspaceDir: workspaceDir,
		settings:     settings,
		fsService:    services.NewFileSystemService(), // Initialize FileSystemService
		gitService:   services.NewGitService(),        // Initialize GitService
	}
}

// LoadExtensions discovers and loads extensions from configured paths.
func (em *ExtensionManager) LoadExtensions() ([]Extension, error) {
	var loadedExtensions []Extension
	for _, extPath := range em.settings.ExtensionPaths {
		entries, err := em.fsService.ListDirectory(extPath)
		if err != nil {
			// Log error but continue with other paths
			fmt.Printf("Warning: Failed to list directory %s: %v\n", extPath, err)
			continue
		}

		for _, entry := range entries {
			fullPath := em.fsService.JoinPaths(extPath, entry)

			fileInfo, err := os.Lstat(fullPath) // Use Lstat to get info about the link itself
			if err != nil {
				fmt.Printf("Warning: Failed to get file info for %s: %v\n", fullPath, err)
				continue
			}

			var actualPath string
			var installType string
			var installSource string

			if fileInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
				// It's a symlink, resolve it
				resolvedPath, err := os.Readlink(fullPath)
				if err != nil {
					fmt.Printf("Warning: Failed to read symlink %s: %v\n", fullPath, err)
					continue
				}
				// If the symlink is relative, resolve it against the workspace directory
				if !filepath.IsAbs(resolvedPath) {
					resolvedPath = filepath.Join(em.workspaceDir, resolvedPath)
				}
				actualPath = resolvedPath
				installType = "link"
				installSource = actualPath // For linked extensions, source is the resolved path
			} else if fileInfo.IsDir() {
				actualPath = fullPath
				// Check if it's a git repository
				gitRepoPath := em.fsService.JoinPaths(actualPath, ".git")
				exists, err := em.fsService.PathExists(gitRepoPath)
				if err != nil {
					fmt.Printf("Warning: Failed to check if %s is a git repository: %v\n", actualPath, err)
					installType = "local"
					installSource = actualPath
				} else if exists {
					installType = "git"
					installSource = actualPath // For git, source is the path itself for now
				} else {
					installType = "local"
					installSource = actualPath
				}
			} else {
				continue // Not a directory or symlink to a directory
			}

			extensionConfigFile := em.fsService.JoinPaths(actualPath, "gemini-extension.json")
			exists, err := em.fsService.PathExists(extensionConfigFile)
			if err != nil {
				fmt.Printf("Warning: Failed to check existence of %s: %v\n", extensionConfigFile, err)
				continue
			}

			if exists {
				// Read and parse gemini-extension.json
				configBytes, err := ioutil.ReadFile(extensionConfigFile)
				if err != nil {
					fmt.Printf("Warning: Failed to read extension config file %s: %v\n", extensionConfigFile, err)
					continue
				}

				var extConfig ExtensionConfig
				if err = json.Unmarshal(configBytes, &extConfig); err != nil {
					fmt.Printf("Warning: Failed to parse extension config file %s: %v\n", extensionConfigFile, err)
					continue
				}

				loadedExtensions = append(loadedExtensions, Extension{
					Name:          extConfig.Name,
					Path:          actualPath,
					Description:   extConfig.Description,
					McpServers:    extConfig.McpServers,
					InstallType:   installType,
					InstallSource: installSource,
				})
			}
		}
	}

	return loadedExtensions, nil
}

// ToOutputString converts an Extension to a string for output.
func (em *ExtensionManager) ToOutputString(ext Extension) string {
	return fmt.Sprintf("Name: %s\nPath: %s\nDescription: %s", ext.Name, ext.Path, ext.Description)
}

// InstallOrUpdateExtension installs or updates an extension.
func (em *ExtensionManager) InstallOrUpdateExtension(metadata ExtensionInstallMetadata) (string, error) {
	fmt.Printf("Installing/updating extension from source: %s (type: %s)\n", metadata.Source, metadata.Type)

	// Determine installation path
	var extensionName string
	if metadata.Type == "git" {
		// Extract repository name from git URL
		repoName := filepath.Base(metadata.Source)
		if strings.HasSuffix(repoName, ".git") {
			repoName = strings.TrimSuffix(repoName, ".git")
		}
		extensionName = repoName
	} else {
		extensionName = filepath.Base(metadata.Source)
	}
	installPath := em.fsService.JoinPaths(em.settings.ExtensionPaths[0], extensionName) // For simplicity, use first extension path

	if metadata.Type == "git" {
		fmt.Printf("Cloning/fetching git repository %s to %s\n", metadata.Source, installPath)
		// Check if directory already exists
		exists, err := em.fsService.PathExists(installPath)
		if err != nil {
			return "", fmt.Errorf("failed to check existence of install path %s: %w", installPath, err)
		}

		if exists {
			// If exists, assume update (for now, just pull)
			fmt.Printf("Repository already exists, pulling latest changes in %s\n", installPath)
			r, err := git.PlainOpen(installPath)
			if err != nil {
				return "", fmt.Errorf("failed to open git repository at %s for update: %w", installPath, err)
			}
			w, err := r.Worktree()
			if err != nil {
				return "", fmt.Errorf("failed to get worktree for %s: %w", installPath, err)
			}
			err = w.Pull(&git.PullOptions{})
			if err != nil && err != git.NoErrAlreadyUpToDate {
				return "", fmt.Errorf("failed to pull latest changes for %s: %w", installPath, err)
			}
		} else {
			// If not exists, clone
			fmt.Printf("Cloning repository %s to %s\n", metadata.Source, installPath)
			_, err = git.PlainClone(installPath, false, &git.CloneOptions{
				URL:               metadata.Source,
				RecurseSubmodules: git.DefaultSubmoduleRecursionDepth, // Handle submodules
			})
			if err != nil {
				return "", fmt.Errorf("failed to clone git repository %s to %s: %w", metadata.Source, installPath, err)
			}
		}

	} else if metadata.Type == "local" {
		fmt.Printf("Copying local extension from %s to %s\n", metadata.Source, installPath)
		// Ensure destination directory is clean before copying
		exists, err := em.fsService.PathExists(installPath)
		if err != nil {
			return "", fmt.Errorf("failed to check existence of install path %s: %w", installPath, err)
		}
		if exists {
			fmt.Printf("Removing existing local extension at %s for update.\n", installPath)
			err = os.RemoveAll(installPath)
			if err != nil {
				return "", fmt.Errorf("failed to remove existing local extension at %s: %w", installPath, err)
			}
		}
		// Implement actual file copying
		err = em.fsService.CopyDirectory(metadata.Source, installPath)
		if err != nil {
			return "", fmt.Errorf("failed to copy local extension: %w", err)
		}
	}

	// Read the extension config from the source to get the real name
	configPath := em.fsService.JoinPaths(installPath, "gemini-extension.json")
	configBytes, err := em.fsService.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("failed to read extension config file from source: %w", err)
	}

	var extConfig ExtensionConfig
	err = json.Unmarshal([]byte(configBytes), &extConfig)
	if err != nil {
		return "", fmt.Errorf("failed to parse extension config file from source: %w", err)
	}

	// Potentially update the config file if needed (e.g., with installation metadata)
	// For now, we'll just use the name from the config
	err = em.fsService.WriteFile(configPath, configBytes)
	if err != nil {
		return "", fmt.Errorf("failed to write extension config file: %w", err)
	}

	// Rename the extension directory to the name from the config
	newInstallPath := em.fsService.JoinPaths(em.settings.ExtensionPaths[0], extConfig.Name)
	if installPath != newInstallPath {
		fmt.Printf("Renaming extension directory from %s to %s\n", installPath, newInstallPath)
		if err := os.Rename(installPath, newInstallPath); err != nil {
			return "", fmt.Errorf("failed to rename extension directory: %w", err)
		}
	}

	return extConfig.Name, nil
}

// UpdateExtension updates an installed extension.
func (em *ExtensionManager) UpdateExtension(name string) error {
	installedExtensions, err := em.LoadExtensions()
	if err != nil {
		return fmt.Errorf("failed to load installed extensions: %w", err)
	}

	var targetExtension *Extension
	for _, ext := range installedExtensions {
		if ext.Name == name {
			targetExtension = &ext
			break
		}
	}

	if targetExtension == nil {
		return fmt.Errorf("extension \"%s\" not found", name)
	}

	// Check if it's a git repository
	gitRepoPath := em.fsService.JoinPaths(targetExtension.Path, ".git")
	exists, err := em.fsService.PathExists(gitRepoPath)
	if err != nil {
		return fmt.Errorf("failed to check if %s is a git repository: %w", targetExtension.Path, err)
	}

	if exists {
		fmt.Printf("Updating git-based extension \"%s\" at %s\n", name, targetExtension.Path)
		r, err := git.PlainOpen(targetExtension.Path)
		if err != nil {
			return fmt.Errorf("failed to open git repository at %s for update: %w", targetExtension.Path, err)
		}
		w, err := r.Worktree()
		if err != nil {
			return fmt.Errorf("failed to get worktree for %s: %w", targetExtension.Path, err)
		}
		err = w.Pull(&git.PullOptions{})
		if err != nil && err != git.NoErrAlreadyUpToDate {
			return fmt.Errorf("failed to pull latest changes for %s: %w", targetExtension.Path, err)
		}
		fmt.Printf("Extension \"%s\" updated successfully.\n", name)
	} else {
		fmt.Printf("Extension \"%s\" at %s is a local extension and cannot be updated automatically.\n", name, targetExtension.Path)
	}

	return nil
}

// UninstallExtension removes an installed extension.
func (em *ExtensionManager) UninstallExtension(name string, _ bool) error {
	// For now, we ignore the `_ bool` (force) argument as we don't have interactive consent.

	extensionPath := em.fsService.JoinPaths(em.settings.ExtensionPaths[0], name)

	exists, err := em.fsService.PathExists(extensionPath)
	if err != nil {
		return fmt.Errorf("failed to check existence of extension path %s: %w", extensionPath, err)
	}
	if !exists {
		return fmt.Errorf("extension \"%s\" not found at %s", name, extensionPath)
	}

	fmt.Printf("Removing extension directory: %s\n", extensionPath)
	err = os.RemoveAll(extensionPath)
	if err != nil {
		return fmt.Errorf("failed to remove extension directory %s: %w", extensionPath, err)
	}

	return nil
}

// EnableExtension enables an extension.
func (em *ExtensionManager) EnableExtension(name string, scope config.SettingScope) error {
	// For now, this is a placeholder. In a real implementation, this would modify
	// a settings file to mark the extension as enabled for the given scope.
	fmt.Printf("Enabling extension \"%s\" for scope \"%s\" (placeholder)\n", name, scope)
	return nil
}

// DisableExtension disables an extension.
func (em *ExtensionManager) DisableExtension(name string, scope config.SettingScope) error {
	// For now, this is a placeholder. In a real implementation, this would modify
	// a settings file to mark the extension as disabled for the given scope.
	fmt.Printf("Disabling extension \"%s\" for scope \"%s\" (placeholder)\n", name, scope)
	return nil
}

// LinkExtension links a local extension.
func (em *ExtensionManager) LinkExtension(localPath string) error {
	// 1. Check if local path exists and is a directory
	exists, err := em.fsService.PathExists(localPath)
	if err != nil {
		return fmt.Errorf("failed to check existence of local path %s: %w", localPath, err)
	}
	if !exists {
		return fmt.Errorf("local extension path not found: %s", localPath)
	}
	isDir, err := em.fsService.IsDirectory(localPath)
	if err != nil {
		return fmt.Errorf("failed to check if %s is a directory: %w", localPath, err)
	}
	if !isDir {
		return fmt.Errorf("local extension path is not a directory: %s", localPath)
	}

	// 2. Read gemini-extension.json to get the extension name
	configPath := em.fsService.JoinPaths(localPath, "gemini-extension.json")
	configBytes, err := em.fsService.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read gemini-extension.json from %s: %w", localPath, err)
	}

	var extConfig ExtensionConfig
	if err := json.Unmarshal([]byte(configBytes), &extConfig); err != nil {
		return fmt.Errorf("failed to parse gemini-extension.json from %s: %w", localPath, err)
	}

	if extConfig.Name == "" {
		return fmt.Errorf("extension name not found in gemini-extension.json at %s", localPath)
	}

	// 3. Determine target link path
	installDir := em.settings.ExtensionPaths[0] // For simplicity, use the first extension path
	linkPath := em.fsService.JoinPaths(installDir, extConfig.Name)

	// Ensure the installation directory exists
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("failed to create extension installation directory %s: %w", installDir, err)
	}

	// Check if a link or directory already exists at linkPath
	exists, err = em.fsService.PathExists(linkPath)
	if err != nil {
		return fmt.Errorf("failed to check existence of link path %s: %w", linkPath, err)
	}
	if exists {
		// If it exists, and is a symlink, remove it. If it's a directory, return an error.
		fileInfo, err := os.Lstat(linkPath)
		if err != nil {
			return fmt.Errorf("failed to get file info for %s: %w", linkPath, err)
		}
		if fileInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
			fmt.Printf("Removing existing symlink at %s\n", linkPath)
			if err := os.Remove(linkPath); err != nil {
				return fmt.Errorf("failed to remove existing symlink at %s: %w", linkPath, err)
			}
		} else if fileInfo.IsDir() {
			return fmt.Errorf("a directory already exists at %s. Cannot link extension.", linkPath)
		} else {
			return fmt.Errorf("a file already exists at %s. Cannot link extension.", linkPath)
		}
	}

	// 4. Create symbolic link
	fmt.Printf("Creating symlink from %s to %s\n", localPath, linkPath)
	if err := os.Symlink(localPath, linkPath); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	return nil
}
