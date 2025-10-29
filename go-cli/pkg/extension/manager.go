package extension

import (
	"encoding/json"
	"fmt"
	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/mcp"
	"go-ai-agent-v2/go-cli/pkg/services"
	"io/ioutil"
	"path/filepath"
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

// installOrUpdateExtension installs or updates an extension.
func (em *ExtensionManager) InstallOrUpdateExtension(metadata ExtensionInstallMetadata) (string, error) {
	fmt.Printf("Installing/updating extension from source: %s (type: %s)\n", metadata.Source, metadata.Type)

	// Determine installation path
	installPath := em.fsService.JoinPaths(em.settings.ExtensionPaths[0], filepath.Base(metadata.Source)) // For simplicity, use first extension path

	if metadata.Type == "git" {
		fmt.Printf("Cloning/fetching git repository %s to %s\n", metadata.Source, installPath)
		// TODO: Implement actual git clone/fetch using go-git
		// For now, simulate success
	} else if metadata.Type == "local" {
		fmt.Printf("Copying local extension from %s to %s\n", metadata.Source, installPath)
		// TODO: Implement actual file copying
		// For now, simulate success
	}

	// Simulate creating/updating gemini-extension.json
	dummyConfig := ExtensionConfig{Name: filepath.Base(metadata.Source), Description: "Installed via CLI"}
	configBytes, err := json.MarshalIndent(dummyConfig, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal dummy extension config: %w", err)
	}
	em.fsService.WriteFile(em.fsService.JoinPaths(installPath, "gemini-extension.json"), string(configBytes))

	return filepath.Base(metadata.Source), nil
}
