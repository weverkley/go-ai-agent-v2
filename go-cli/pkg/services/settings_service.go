package services

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/types"
)

// SettingsService manages application settings.
type SettingsService struct {
	mu        sync.RWMutex
	settings  *config.Settings
	baseDir   string // Base directory to resolve settings file
}

// NewSettingsService creates a new SettingsService instance.
func NewSettingsService(baseDir string) *SettingsService {
	ss := &SettingsService{
		baseDir: baseDir,
	}
	ss.settings = config.LoadSettings(baseDir) // Load initial settings
	return ss
}

// Get returns the value of a specific setting.
func (ss *SettingsService) Get(key string) (interface{}, bool) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	// Use reflection or a switch statement to get values
	switch key {
	case "model":
		return ss.settings.Model, true
	case "debugMode":
		return ss.settings.DebugMode, true
	case "userMemory":
		return ss.settings.UserMemory, true
	case "approvalMode":
		return ss.settings.ApprovalMode, true
	case "showMemoryUsage":
		return ss.settings.ShowMemoryUsage, true
	case "telemetryEnabled":
		return ss.settings.TelemetryEnabled, true
	case "proxy":
		return ss.settings.Proxy, true
	case "toolDiscoveryCommand":
		return ss.settings.ToolDiscoveryCommand, true
	case "toolCallCommand":
		return ss.settings.ToolCallCommand, true
	case "enabledExtensions":
		return ss.settings.EnabledExtensions, true
	case "extensionPaths":
		return ss.settings.ExtensionPaths, true
	case "mcpServers":
		return ss.settings.McpServers, true
	// Add other settings here
	default:
		return nil, false
	}
}

// Set sets the value of a specific setting.
func (ss *SettingsService) Set(key string, value interface{}) error {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	// Use reflection or a switch statement to set values
	switch key {
	case "model":
		if v, ok := value.(string); ok {
			ss.settings.Model = v
		} else {
			return fmt.Errorf("invalid type for model setting, expected string")
		}
	case "debugMode":
		if v, ok := value.(bool); ok {
			ss.settings.DebugMode = v
		} else {
			return fmt.Errorf("invalid type for debugMode setting, expected bool")
		}
	case "userMemory":
		if v, ok := value.(string); ok {
			ss.settings.UserMemory = v
		} else {
			return fmt.Errorf("invalid type for userMemory setting, expected string")
		}
	case "approvalMode":
		if v, ok := value.(types.ApprovalMode); ok {
			ss.settings.ApprovalMode = v
		} else {
			return fmt.Errorf("invalid type for approvalMode setting, expected types.ApprovalMode")
		}
	case "showMemoryUsage":
		if v, ok := value.(bool); ok {
			ss.settings.ShowMemoryUsage = v
		} else {
			return fmt.Errorf("invalid type for showMemoryUsage setting, expected bool")
		}
	case "telemetryEnabled":
		if v, ok := value.(bool); ok {
			ss.settings.TelemetryEnabled = v
		} else {
			return fmt.Errorf("invalid type for telemetryEnabled setting, expected bool")
		}
	case "proxy":
		if v, ok := value.(string); ok {
			ss.settings.Proxy = v
		} else {
			return fmt.Errorf("invalid type for proxy setting, expected string")
		}
	case "toolDiscoveryCommand":
		if v, ok := value.(string); ok {
			ss.settings.ToolDiscoveryCommand = v
		} else {
			return fmt.Errorf("invalid type for toolDiscoveryCommand setting, expected string")
		}
	case "toolCallCommand":
		if v, ok := value.(string); ok {
			ss.settings.ToolCallCommand = v
		} else {
			return fmt.Errorf("invalid type for toolCallCommand setting, expected string")
		}
	// Add other settings here
	default:
		return fmt.Errorf("setting '%s' not found or not settable", key)
	}
	return nil
}

// AllSettings returns a map of all settings and their current values.
func (ss *SettingsService) AllSettings() map[string]interface{} {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	all := make(map[string]interface{})
	all["model"] = ss.settings.Model
	all["debugMode"] = ss.settings.DebugMode
	all["userMemory"] = ss.settings.UserMemory
	all["approvalMode"] = ss.settings.ApprovalMode
	all["showMemoryUsage"] = ss.settings.ShowMemoryUsage
	all["telemetryEnabled"] = ss.settings.TelemetryEnabled
	all["proxy"] = ss.settings.Proxy
	all["toolDiscoveryCommand"] = ss.settings.ToolDiscoveryCommand
	all["toolCallCommand"] = ss.settings.ToolCallCommand
	all["enabledExtensions"] = ss.settings.EnabledExtensions
	all["extensionPaths"] = ss.settings.ExtensionPaths
	all["mcpServers"] = ss.settings.McpServers
	// Add other settings here
	return all
}

// Reset resets all settings to their default values.
func (ss *SettingsService) Reset() error {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	// Delete the settings file to ensure defaults are loaded
	settingsPath := filepath.Join(ss.baseDir, config.SettingsFileName)
	if err := os.Remove(settingsPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove settings file: %w", err)
	}

	ss.settings = config.LoadSettings(ss.baseDir) // Reload to get defaults
	return nil
}

// Save persists the current settings to a file.
func (ss *SettingsService) Save() error {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return config.SaveSettings(ss.baseDir, ss.settings)
}