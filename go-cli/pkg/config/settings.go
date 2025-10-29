package config

import (
	"encoding/json"
	"fmt"
	"go-ai-agent-v2/go-cli/pkg/mcp"
	"os"
	"path/filepath"
)

// SettingScope defines the scope of a setting.
type SettingScope string

const (
	SettingScopeUser      SettingScope = "user"
	SettingScopeWorkspace SettingScope = "workspace"
)

// Settings represents the application settings.
type Settings struct {
	ExtensionPaths []string                       `json:"extensionPaths"`
	McpServers     map[string]mcp.MCPServerConfig `json:"mcpServers,omitempty"`
}

func getSettingsPath(workspaceDir string) string {
	return filepath.Join(workspaceDir, ".gemini", "settings.json")
}

// LoadSettings loads the application settings from various sources.
func LoadSettings(workspaceDir string) *Settings {
	settingsPath := getSettingsPath(workspaceDir)
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		// Return default settings if file doesn't exist or can't be read
		return &Settings{
			ExtensionPaths: []string{filepath.Join(workspaceDir, ".gemini", "extensions")},
			McpServers:     make(map[string]mcp.MCPServerConfig),
		}
	}

	var settings Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		fmt.Printf("Warning: could not parse settings file, using defaults: %v\n", err)
		// Return default settings on parsing error
		return &Settings{
			ExtensionPaths: []string{filepath.Join(workspaceDir, ".gemini", "extensions")},
			McpServers:     make(map[string]mcp.MCPServerConfig),
		}
	}

	return &settings
}

// SaveSettings saves the application settings.
func SaveSettings(workspaceDir string, settings *Settings) error {
	settingsPath := getSettingsPath(workspaceDir)
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(settingsPath), 0755); err != nil {
		return fmt.Errorf("failed to create settings directory: %w", err)
	}

	return os.WriteFile(settingsPath, data, 0644)
}
