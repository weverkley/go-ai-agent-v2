package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"go-ai-agent-v2/go-cli/pkg/types"
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
	McpServers     map[string]types.MCPServerConfig `json:"mcpServers,omitempty"`
	DebugMode      bool                           `json:"debugMode,omitempty"`
	UserMemory     string                         `json:"userMemory,omitempty"`
	ApprovalMode   types.ApprovalMode             `json:"approvalMode,omitempty"`
	ShowMemoryUsage bool                          `json:"showMemoryUsage,omitempty"`
	TelemetryEnabled bool                          `json:"telemetryEnabled,omitempty"`
	Model          string                         `json:"model,omitempty"`
	Proxy          string                         `json:"proxy,omitempty"`
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
			McpServers:     make(map[string]types.MCPServerConfig),
			DebugMode:      false,
			UserMemory:     "",
			ApprovalMode:   types.ApprovalModeDefault,
			ShowMemoryUsage: false,
			TelemetryEnabled: false,
			Model:          "gemini-pro", // Default model
			Proxy:          "",
		}
	}

	var settings Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		fmt.Printf("Warning: could not parse settings file, using defaults: %v\n", err)
		// Return default settings on parsing error
		return &Settings{
			ExtensionPaths: []string{filepath.Join(workspaceDir, ".gemini", "extensions")},
			McpServers:     make(map[string]types.MCPServerConfig),
			DebugMode:      false,
			UserMemory:     "",
			ApprovalMode:   types.ApprovalModeDefault,
			ShowMemoryUsage: false,
			TelemetryEnabled: false,
			Model:          "gemini-pro", // Default model
			Proxy:          "",
		}
	}

	// Apply defaults if not set in the loaded settings
	if len(settings.ExtensionPaths) == 0 {
		settings.ExtensionPaths = []string{filepath.Join(workspaceDir, ".gemini", "extensions")}
	}
	if settings.McpServers == nil {
		settings.McpServers = make(map[string]types.MCPServerConfig)
	}
	if settings.ApprovalMode == "" {
		settings.ApprovalMode = types.ApprovalModeDefault
	}
	if settings.Model == "" {
		settings.Model = "gemini-pro"
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
