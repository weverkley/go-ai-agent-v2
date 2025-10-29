package extension

import (
	"encoding/json"
	"fmt"
	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/mcp"
	"go-ai-agent-v2/go-cli/pkg/services"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
)

// Extension represents a simplified extension structure.
type Extension struct {
	Name        string `json:"name"`
	Path        string
	Description string `json:"description"`
	McpServers  map[string]mcp.MCPServerConfig `json:"mcpServers,omitempty"`
	// Add other relevant fields from the JS Extension object
}

// ExtensionConfig represents the structure of gemini-extension.json
type ExtensionConfig struct {
	Name        string `json:"name"`
	Description string `json:"description"`
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
	fmt.Printf("Loading extensions from %s (placeholder)\n", em.workspaceDir)
	fmt.Printf("Using extension paths from settings: %v\n", em.settings.ExtensionPaths)

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
			isDir, err := em.fsService.IsDirectory(fullPath)
			if err != nil {
				fmt.Printf("Warning: Failed to check if %s is a directory: %v\n", fullPath, err)
				continue
			}

			if isDir {
				extensionConfigFile := em.fsService.JoinPaths(fullPath, "gemini-extension.json")
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
					err = json.Unmarshal(configBytes, &extConfig)
					if err != nil {
						fmt.Printf("Warning: Failed to parse extension config file %s: %v\n", extensionConfigFile, err)
						continue
					}

					loadedExtensions = append(loadedExtensions, Extension{
						Name:        extConfig.Name,
						Path:        fullPath,
						Description: extConfig.Description,
						McpServers:  extConfig.McpServers,
					})
				}
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
				URL:      metadata.Source,
				RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,  // Handle submodules
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
		err = copyDir(metadata.Source, installPath)
		if err != nil {
			return "", fmt.Errorf("failed to copy local extension: %w", err)
		}
	}

	// Simulate creating/updating gemini-extension.json
	dummyConfig := ExtensionConfig{Name: filepath.Base(metadata.Source), Description: "Installed via CLI"}
	configBytes, err := json.MarshalIndent(dummyConfig, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal dummy extension config: %w", err)
	}
	err = em.fsService.WriteFile(em.fsService.JoinPaths(installPath, "gemini-extension.json"), string(configBytes))
	if err != nil {
		return "", fmt.Errorf("failed to write extension config file: %w", err)
	}

	return filepath.Base(metadata.Source), nil
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

// copyDir recursively copies a directory from src to dst.
func copyDir(src string, dst string) error {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err = os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	dirents, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}

	for _, dirent := range dirents {
		srcPath := filepath.Join(src, dirent.Name())
		dstPath := filepath.Join(dst, dirent.Name())

		if dirent.IsDir() {
			err = copyDir(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			err = copyFile(srcPath, dstPath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// copyFile copies a file from src to dst.
func copyFile(src string, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}
