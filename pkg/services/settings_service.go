package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/spf13/viper"
)

const (
	SettingsFileName = ".goaiagent/settings.json"
)

// Settings represents the application settings.
type Settings struct {
	ExtensionPaths       []string                            `json:"extensionPaths" mapstructure:"extensionPaths"`
	McpServers           map[string]types.MCPServerConfig    `json:"mcpServers,omitempty" mapstructure:"mcpServers"`
	DebugMode            bool                                `json:"debugMode,omitempty" mapstructure:"debugMode"`
	ApprovalMode         types.ApprovalMode                  `json:"approvalMode,omitempty" mapstructure:"approvalMode"`
	DangerousTools       []string                            `json:"dangerousTools,omitempty" mapstructure:"dangerousTools"` // New field
	Model                string                              `json:"model,omitempty" mapstructure:"model"`
	Executor             string                              `json:"executor,omitempty" mapstructure:"executor"`
	Proxy                string                              `json:"proxy,omitempty" mapstructure:"proxy"`
	EnabledExtensions    map[types.SettingScope][]string     `json:"enabledExtensions,omitempty" mapstructure:"enabledExtensions"`
	ToolDiscoveryCommand string                              `json:"toolDiscoveryCommand,omitempty" mapstructure:"toolDiscoveryCommand"`
	ToolCallCommand      string                              `json:"toolCallCommand,omitempty" mapstructure:"toolCallCommand"`
	Telemetry            *types.TelemetrySettings            `json:"telemetry,omitempty" mapstructure:"telemetry"`
	GoogleCustomSearch   *types.GoogleCustomSearchSettings   `json:"googleCustomSearch,omitempty" mapstructure:"googleCustomSearch"`
	WebSearchProvider    types.WebSearchProvider             `json:"webSearchProvider,omitempty" mapstructure:"webSearchProvider"`
	Tavily               *types.TavilySettings               `json:"tavily,omitempty" mapstructure:"tavily"`
	CodebaseInvestigator *types.CodebaseInvestigatorSettings `json:"codebaseInvestigator,omitempty" mapstructure:"codebaseInvestigator"`
	TestWriter           *types.TestWriterSettings           `json:"testWriter,omitempty" mapstructure:"testWriter"`
}

func newDefaultSettings(workspaceDir string) {
	telemetryOutDir := filepath.Join(workspaceDir, ".goaiagent", "tmp")
	viper.SetDefault("extensionPaths", []string{filepath.Join(workspaceDir, ".goaiagent", "extensions")})
	viper.SetDefault("mcpServers", make(map[string]types.MCPServerConfig))
	viper.SetDefault("debugMode", false)
	viper.SetDefault("approvalMode", types.ApprovalModeDefault)
	viper.SetDefault("dangerousTools", []string{types.EXECUTE_COMMAND_TOOL_NAME, types.WRITE_FILE_TOOL_NAME, types.SMART_EDIT_TOOL_NAME, types.USER_CONFIRM_TOOL_NAME})
	viper.SetDefault("model", "mock-flash")
	viper.SetDefault("executor", types.ExecutorTypeMock)
	viper.SetDefault("proxy", "")
	viper.SetDefault("enabledExtensions", make(map[types.SettingScope][]string))
	viper.SetDefault("toolDiscoveryCommand", "")
	viper.SetDefault("toolCallCommand", "")
	viper.SetDefault("telemetry", &types.TelemetrySettings{
		Enabled:  true,
		OutDir:   telemetryOutDir,
		LogLevel: "info",
	})
	viper.SetDefault("googleCustomSearch", &types.GoogleCustomSearchSettings{
		ApiKey: "API_KEY_GOES_HERE",
		CxId:   "CX_ID_GOES_HERE",
	})
	viper.SetDefault("webSearchProvider", types.WebSearchProviderGoogleCustomSearch)
	viper.SetDefault("tavily", &types.TavilySettings{
		ApiKey: "API_KEY_GOES_HERE",
	})
	viper.SetDefault("codebaseInvestigator", &types.CodebaseInvestigatorSettings{Enabled: true})
	viper.SetDefault("testWriter", &types.TestWriterSettings{Enabled: true})
}

// SettingsService manages application settings.
type SettingsService struct {
	mu       sync.RWMutex
	settings *Settings
	baseDir  string // Base directory to resolve settings file
}

// NewSettingsService creates a new SettingsService instance.
func NewSettingsService(baseDir string, extensionManager types.ExtensionManager) types.SettingsServiceIface {
	ss := &SettingsService{
		baseDir: baseDir,
	}

	viper.AddConfigPath(filepath.Join(baseDir, ".goaiagent"))
	viper.SetConfigName("settings")
	viper.SetConfigType("json")

	viper.SetEnvPrefix("GOAIAGENT")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	newDefaultSettings(baseDir)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			fmt.Println("No settings file found, using defaults.")
			// You might want to create a default config file here
			if err := viper.SafeWriteConfig(); err != nil {
				fmt.Printf("Warning: failed to write default settings: %v\n", err)
			}
		} else {
			// Config file was found but another error was produced
			fmt.Printf("Warning: could not parse settings file, using defaults: %v\n", err)
		}
	}

	var settings Settings
	if err := viper.Unmarshal(&settings); err != nil {
		fmt.Printf("Warning: could not unmarshal settings, using defaults: %v\n", err)
	}
	ss.settings = &settings

	if err := extensionManager.LoadExtensionStatus(); err != nil {
		fmt.Printf("Warning: failed to load extension status: %v\n", err)
	}
	return ss
}


// Get returns the value of a specific setting.
func (ss *SettingsService) Get(key string) (interface{}, bool) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return viper.Get(key), viper.IsSet(key)
}

// GetTelemetrySettings returns the telemetry settings.
func (ss *SettingsService) GetTelemetrySettings() *types.TelemetrySettings {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	var telemetrySettings types.TelemetrySettings
	if err := viper.UnmarshalKey("telemetry", &telemetrySettings); err != nil {
		return nil
	}
	return &telemetrySettings
}

// GetGoogleCustomSearchSettings returns the Google Custom Search settings.
func (ss *SettingsService) GetGoogleCustomSearchSettings() *types.GoogleCustomSearchSettings {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	var googleCustomSearchSettings types.GoogleCustomSearchSettings
	if err := viper.UnmarshalKey("googleCustomSearch", &googleCustomSearchSettings); err != nil {
		return nil
	}
	return &googleCustomSearchSettings
}

// GetWebSearchProvider returns the configured web search provider.
func (ss *SettingsService) GetWebSearchProvider() types.WebSearchProvider {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return types.WebSearchProvider(viper.GetString("webSearchProvider"))
}

// GetTavilySettings returns the Tavily settings.
func (ss *SettingsService) GetTavilySettings() *types.TavilySettings {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	var tavilySettings types.TavilySettings
	if err := viper.UnmarshalKey("tavily", &tavilySettings); err != nil {
		return nil
	}
	return &tavilySettings
}

// GetWorkspaceDir returns the base directory for the settings service.
func (ss *SettingsService) GetWorkspaceDir() string {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.baseDir
}

// GetDangerousTools returns the list of tools that require confirmation.
func (ss *SettingsService) GetDangerousTools() []string {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return viper.GetStringSlice("dangerousTools")
}

// GetCodebaseInvestigatorSettings returns the codebase investigator settings.
func (ss *SettingsService) GetCodebaseInvestigatorSettings() *types.CodebaseInvestigatorSettings {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	var codebaseInvestigatorSettings types.CodebaseInvestigatorSettings
	if err := viper.UnmarshalKey("codebaseInvestigator", &codebaseInvestigatorSettings); err != nil {
		return nil
	}
	return &codebaseInvestigatorSettings
}

// GetTestWriterSettings returns the test writer settings.
func (ss *SettingsService) GetTestWriterSettings() *types.TestWriterSettings {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	var testWriterSettings types.TestWriterSettings
	if err := viper.UnmarshalKey("testWriter", &testWriterSettings); err != nil {
		return nil
	}
	return &testWriterSettings
}

// GetTelemetryLogPath returns the configured telemetry log file path.
func (ss *SettingsService) GetTelemetryLogPath() string {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	if viper.IsSet("telemetry.outdir") {
		return filepath.Join(viper.GetString("telemetry.outdir"), "go-ai-agent.log")
	}
	// Fallback to a default path if not explicitly configured
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "" // Handle error or return a more robust default
	}
	return filepath.Join(homeDir, ".goaiagent", "logs", "go-ai-agent.log")
}

// Set sets the value of a specific setting.
func (ss *SettingsService) Set(key string, value interface{}) error {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	viper.Set(key, value)
	return nil
}

// AllSettings returns a map of all settings and their current values.
func (ss *SettingsService) AllSettings() map[string]interface{} {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return viper.AllSettings()
}

// Reset resets all settings to their default values.
func (ss *SettingsService) Reset() error {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	// Delete the settings file to ensure defaults are loaded
	settingsPath := viper.ConfigFileUsed()
	if err := os.Remove(settingsPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove settings file: %w", err)
	}

	// Re-read config to load defaults
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config: %w", err)
		}
	}
	var settings Settings
	if err := viper.Unmarshal(&settings); err != nil {
		return fmt.Errorf("failed to unmarshal settings: %w", err)
	}
	ss.settings = &settings

	return nil
}

// Save persists the current settings to a file.
func (ss *SettingsService) Save() error {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return viper.WriteConfig()
}
