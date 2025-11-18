package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/core"
	"go-ai-agent-v2/go-cli/pkg/types"
)

const (
	SettingsFileName = ".goaiagent/settings.json"
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
	Executor       string                         `json:"executor,omitempty"`
	Proxy          string                         `json:"proxy,omitempty"`
	EnabledExtensions map[types.SettingScope][]string `json:"enabledExtensions,omitempty"`
	ToolDiscoveryCommand string `json:"toolDiscoveryCommand,omitempty"`
	ToolCallCommand      string `json:"toolCallCommand,omitempty"`
	Telemetry      *types.TelemetrySettings       `json:"telemetry,omitempty"`
	GoogleCustomSearch *types.GoogleCustomSearchSettings `json:"googleCustomSearch,omitempty"`
	WebSearchProvider  types.WebSearchProvider    `json:"webSearchProvider,omitempty"`
	Tavily             *types.TavilySettings      `json:"tavily,omitempty"`
}

// LoadSettings loads the application settings from various sources.
func LoadSettings(workspaceDir string) *Settings {
	settingsPath := getSettingsPath(workspaceDir)
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		// Return default settings if file doesn't exist or can't be read
		return &Settings{
			ExtensionPaths: []string{filepath.Join(workspaceDir, ".goaiagent", "extensions")},
			McpServers:     make(map[string]types.MCPServerConfig),
			DebugMode:      false,
			UserMemory:     "",
			ApprovalMode:   types.ApprovalModeDefault,
			ShowMemoryUsage: false,
			TelemetryEnabled: false,
			Model:          "gemini-pro", // Default model
			Executor:       "gemini",     // Default executor
			Proxy:          "",
			EnabledExtensions: make(map[types.SettingScope][]string),
			ToolDiscoveryCommand: "",
			ToolCallCommand:      "",
			Telemetry: &types.TelemetrySettings{
				Enabled: false,
			},
			GoogleCustomSearch: &types.GoogleCustomSearchSettings{},
			WebSearchProvider:  types.WebSearchProviderGoogleCustomSearch, // Default to Google Custom Search
			Tavily:             &types.TavilySettings{},
		}
	}

	var settings Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		fmt.Printf("Warning: could not parse settings file, using defaults: %v\n", err)
		// Return default settings on parsing error
		return &Settings{
			ExtensionPaths: []string{filepath.Join(workspaceDir, ".goaiagent", "extensions")},
			McpServers:     make(map[string]types.MCPServerConfig),
			DebugMode:      false,
			UserMemory:     "",
			ApprovalMode:   types.ApprovalModeDefault,
			ShowMemoryUsage: false,
			TelemetryEnabled: false,
			Model:          "gemini-pro", // Default model
			Executor:       "gemini",     // Default executor
			Proxy:          "",
			EnabledExtensions: make(map[types.SettingScope][]string),
			ToolDiscoveryCommand: "",
			ToolCallCommand:      "",
			Telemetry: &types.TelemetrySettings{
				Enabled: false,
			},
			GoogleCustomSearch: &types.GoogleCustomSearchSettings{},
			WebSearchProvider:  types.WebSearchProviderGoogleCustomSearch, // Default to Google Custom Search
			Tavily:             &types.TavilySettings{},
		}
	}

	// Apply defaults if not set in the loaded settings
	if len(settings.ExtensionPaths) == 0 {
		settings.ExtensionPaths = []string{filepath.Join(workspaceDir, ".goaiagent", "extensions")}
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
	if settings.Executor == "" {
		settings.Executor = "gemini"
	}
	if settings.EnabledExtensions == nil {
		settings.EnabledExtensions = make(map[types.SettingScope][]string)
	}
	if settings.ToolDiscoveryCommand == "" {
		settings.ToolDiscoveryCommand = ""
	}
	if settings.ToolCallCommand == "" {
		settings.ToolCallCommand = ""
	}
	if settings.Telemetry == nil {
		settings.Telemetry = &types.TelemetrySettings{
			Enabled: false,
		}
	}
	if settings.GoogleCustomSearch == nil {
		settings.GoogleCustomSearch = &types.GoogleCustomSearchSettings{}
	}
	if settings.WebSearchProvider == "" {
		settings.WebSearchProvider = types.WebSearchProviderGoogleCustomSearch
	}
	if settings.Tavily == nil {
		settings.Tavily = &types.TavilySettings{}
	}

	return &settings
}

func getSettingsPath(workspaceDir string) string {
	return filepath.Join(workspaceDir, SettingsFileName)
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

// SettingsService manages application settings.
type SettingsService struct {
	mu        sync.RWMutex
	settings  *Settings
	baseDir   string // Base directory to resolve settings file
}

// NewSettingsService creates a new SettingsService instance.
func NewSettingsService(baseDir string) *SettingsService {
	ss := &SettingsService{
		baseDir: baseDir,
	}
	ss.settings = LoadSettings(baseDir) // Load initial settings
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
	case "executor":
		return ss.settings.Executor, true
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

// GetTelemetrySettings returns the telemetry settings.
func (ss *SettingsService) GetTelemetrySettings() *types.TelemetrySettings {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.settings.Telemetry
}

// GetGoogleCustomSearchSettings returns the Google Custom Search settings.
func (ss *SettingsService) GetGoogleCustomSearchSettings() *types.GoogleCustomSearchSettings {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.settings.GoogleCustomSearch
}

// GetWebSearchProvider returns the configured web search provider.
func (ss *SettingsService) GetWebSearchProvider() types.WebSearchProvider {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.settings.WebSearchProvider
}

// GetTavilySettings returns the Tavily settings.
func (ss *SettingsService) GetTavilySettings() *types.TavilySettings {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.settings.Tavily
}

// Set sets the value of a specific setting.
func (ss *SettingsService) Set(key string, value interface{}) error {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	// Use reflection or a switch statement to set values
	switch key {
	case "model":
		if v, ok := value.(string); ok {
			// Validate model against current executor's supported models
			currentExecutorVal, found := ss.Get("executor")
			if !found {
				return fmt.Errorf("cannot validate model: executor setting not found")
			}
			currentExecutorType, ok := currentExecutorVal.(string)
			if !ok {
				return fmt.Errorf("cannot validate model: executor setting is not a string")
			}

			// Create a temporary config for the executor factory
			tempConfigParams := &config.ConfigParameters{
				ModelName: v, // The model we are trying to set
			}
			tempConfig := config.NewConfig(tempConfigParams)

			factory, err := core.NewExecutorFactory(currentExecutorType, tempConfig)
			if err != nil {
				return fmt.Errorf("failed to create executor factory for validation: %w", err)
			}
			tempExecutor, err := factory.NewExecutor(tempConfig, types.GenerateContentConfig{}, nil)
			if err != nil {
				return fmt.Errorf("failed to create temporary executor for model validation: %w", err)
			}

			supportedModels, err := tempExecutor.ListModels()
			if err != nil {
				return fmt.Errorf("failed to list models for validation: %w", err)
			}

			foundModel := false
			for _, sm := range supportedModels {
				if sm == v {
					foundModel = true
					break
				}
			}
			if !foundModel {
				return fmt.Errorf("model '%s' is not supported by the current executor '%s'. Supported models: %v", v, currentExecutorType, supportedModels)
			}

			ss.settings.Model = v
		} else {
			return fmt.Errorf("invalid type for model setting, expected string")
		}
	case "executor":
		if v, ok := value.(string); ok {
			// Validate executor type
			supportedExecutors := map[string]bool{
				"gemini": true,
				"qwen":   true,
				"mock":   true,
			}
			if !supportedExecutors[v] {
				return fmt.Errorf("unsupported executor type '%s'. Supported types: gemini, qwen, mock", v)
			}
			ss.settings.Executor = v
		} else {
			return fmt.Errorf("invalid type for executor setting, expected string")
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
	case "enabledExtensions":
		if v, ok := value.(map[types.SettingScope][]string); ok {
			ss.settings.EnabledExtensions = v
		} else {
			return fmt.Errorf("invalid type for enabledExtensions setting, expected map[types.SettingScope][]string")
		}
	case "extensionPaths":
		if v, ok := value.([]string); ok {
			ss.settings.ExtensionPaths = v
		} else {
			return fmt.Errorf("invalid type for extensionPaths setting, expected []string")
		}
	case "mcpServers":
		if v, ok := value.(map[string]types.MCPServerConfig); ok {
			ss.settings.McpServers = v
		} else {
			return fmt.Errorf("invalid type for mcpServers setting, expected map[string]types.MCPServerConfig")
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
	all["executor"] = ss.settings.Executor
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
	settingsPath := filepath.Join(ss.baseDir, SettingsFileName)
	if err := os.Remove(settingsPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove settings file: %w", err)
	}

	ss.settings = LoadSettings(ss.baseDir) // Reload to get defaults
	return nil
}

// Save persists the current settings to a file.
func (ss *SettingsService) Save() error {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return SaveSettings(ss.baseDir, ss.settings)
}
