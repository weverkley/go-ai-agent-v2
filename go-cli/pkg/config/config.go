package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"go-ai-agent-v2/go-cli/pkg/telemetry"
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
	EnabledExtensions map[SettingScope][]string `json:"enabledExtensions,omitempty"`
	ToolDiscoveryCommand string `json:"toolDiscoveryCommand,omitempty"`
	ToolCallCommand      string `json:"toolCallCommand,omitempty"`
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
			EnabledExtensions: make(map[SettingScope][]string),
			ToolDiscoveryCommand: "",
			ToolCallCommand:      "",
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
			EnabledExtensions: make(map[SettingScope][]string),
			ToolDiscoveryCommand: "",
			ToolCallCommand:      "",
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
	if settings.EnabledExtensions == nil {
		settings.EnabledExtensions = make(map[SettingScope][]string)
	}
	if settings.ToolDiscoveryCommand == "" {
		settings.ToolDiscoveryCommand = ""
	}
	if settings.ToolCallCommand == "" {
		settings.ToolCallCommand = ""
	}

	return &settings
}

func getSettingsPath(workspaceDir string) string {
	return filepath.Join(workspaceDir, ".gemini", "settings.json")
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

// WorkspaceContext defines an interface for accessing workspace-related information.
type WorkspaceContext interface {
	GetDirectories() []string
}

// FileService defines an interface for file system operations.
type FileService interface {
	// shouldIgnoreFile checks if a file should be ignored based on filtering options.
	// This is a placeholder and needs a proper implementation.
	ShouldIgnoreFile(filePath string, options types.FileFilteringOptions) bool
}

// CodebaseInvestigatorSettings represents settings for the Codebase Investigator agent.
type CodebaseInvestigatorSettings struct {
	Enabled        bool    `json:"enabled,omitempty"`
	Model          string  `json:"model,omitempty"`
	ThinkingBudget *int    `json:"thinkingBudget,omitempty"`
	MaxTimeMinutes *int    `json:"maxTimeMinutes,omitempty"`
	MaxNumTurns    *int    `json:"maxNumTurns,omitempty"`
}

// OutputSettings represents the output settings.
type OutputSettings struct {
	Format string `json:"format,omitempty"`
}

// ConfigParameters represents the parameters for creating a new Config.
type ConfigParameters struct {
	SessionID      string
	EmbeddingModel string
	TargetDir      string
	DebugMode      bool
	Model          string
	McpServers     map[string]types.MCPServerConfig
	ApprovalMode   types.ApprovalMode // Use ApprovalMode from types package
	Telemetry      *types.TelemetrySettings
	Output         *OutputSettings
	CodebaseInvestigator *CodebaseInvestigatorSettings
	ToolRegistry *types.ToolRegistry // Changed to exported
	ToolDiscoveryCommand string
	ToolCallCommand      string
}

// Config represents the application configuration.
type Config struct {
	sessionID      string
	embeddingModel string
	targetDir      string
	debugMode      bool
	Model          string
	mcpServers     map[string]types.MCPServerConfig
	approvalMode   types.ApprovalMode // Use ApprovalMode from types package
	telemetry      *types.TelemetrySettings
	output         *OutputSettings
	codebaseInvestigatorSettings *CodebaseInvestigatorSettings
	ToolRegistry *types.ToolRegistry // Changed to exported
	toolDiscoveryCommand string
	toolCallCommand      string
	telemetryLogger telemetry.TelemetryLogger
}

func NewConfig(params *ConfigParameters) *Config {
	cfg := &Config{
		sessionID:      params.SessionID,
		embeddingModel: params.EmbeddingModel,
		targetDir:      params.TargetDir,
		debugMode:      params.DebugMode,
		Model:          params.Model,
		mcpServers:     params.McpServers,
		approvalMode:   params.ApprovalMode,
		telemetry:      params.Telemetry,
		output:         params.Output,
		codebaseInvestigatorSettings: params.CodebaseInvestigator,
		ToolRegistry: params.ToolRegistry,
		toolDiscoveryCommand: params.ToolDiscoveryCommand,
		toolCallCommand:      params.ToolCallCommand,
	}
	cfg.telemetryLogger = telemetry.NewTelemetryLogger(params.Telemetry) // Initialize here
	return cfg
}

// GetToolDiscoveryCommand returns the tool discovery command.
func (c *Config) GetToolDiscoveryCommand() string {
	return c.toolDiscoveryCommand
}

// GetToolCallCommand returns the tool call command.
func (c *Config) GetToolCallCommand() string {
	return c.toolCallCommand
}
// GetModel returns the configured model name.
func (c *Config) GetModel() string {
	return c.Model
}

// GetCodebaseInvestigatorSettings returns the Codebase Investigator settings.
func (c *Config) GetCodebaseInvestigatorSettings() *CodebaseInvestigatorSettings {
	return c.codebaseInvestigatorSettings
}

// GetDebugMode returns true if debug mode is enabled.
func (c *Config) GetDebugMode() bool {
	return c.debugMode
}

// GetToolRegistry returns the global tool registry.
func (c *Config) GetToolRegistry() *types.ToolRegistry {
	return c.ToolRegistry
}

// GetTelemetryLogger returns the initialized telemetry logger.
func (c *Config) GetTelemetryLogger() telemetry.TelemetryLogger {
	return c.telemetryLogger
}

// GetWorkspaceContext returns the workspace context.
// This is a placeholder and should be replaced with a proper implementation.
func (c *Config) GetWorkspaceContext() WorkspaceContext {
	// For now, return a dummy implementation.
	return &dummyWorkspaceContext{}
}

// dummyWorkspaceContext is a placeholder implementation of WorkspaceContext.
type dummyWorkspaceContext struct{}

func (d *dummyWorkspaceContext) GetDirectories() []string {
	// For now, return the current working directory.
	// In a real scenario, this would return configured workspace directories.
	cwd, err := os.Getwd()
	if err != nil {
		return []string{}
	}
	return []string{cwd}
}

// GetFileService returns the file service.
// This is a placeholder and should be replaced with a proper implementation.
func (c *Config) GetFileService() FileService {
	// For now, return a dummy implementation.
	return &DummyFileService{}
}

// DummyFileService is a placeholder implementation of FileService.
type DummyFileService struct{}

func (d *DummyFileService) ShouldIgnoreFile(filePath string, options types.FileFilteringOptions) bool {
	// For now, always return false.
	return false
}