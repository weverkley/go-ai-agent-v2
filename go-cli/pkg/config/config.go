package config

import (
	"os"

	"go-ai-agent-v2/go-cli/pkg/types"
)

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

// TelemetrySettings represents the telemetry settings.
type TelemetrySettings struct {
	Enabled      bool   `json:"enabled,omitempty"`
	Target       string `json:"target,omitempty"`
	OtlpEndpoint string `json:"otlpEndpoint,omitempty"`
	OtlpProtocol string `json:"otlpProtocol,omitempty"`
	LogPrompts   bool   `json:"logPrompts,omitempty"`
	Outfile      string `json:"outfile,omitempty"`
	UseCollector bool   `json:"useCollector,omitempty"`
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
	Telemetry      *TelemetrySettings
	Output         *OutputSettings
	CodebaseInvestigator *CodebaseInvestigatorSettings
	ToolRegistryProvider *types.ToolRegistryProvider // Changed
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
	telemetry      *TelemetrySettings
	output         *OutputSettings
	codebaseInvestigatorSettings *CodebaseInvestigatorSettings
	toolRegistryProvider *types.ToolRegistryProvider // Changed
}

func NewConfig(params *ConfigParameters) *Config {
	return &Config{
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
		toolRegistryProvider: params.ToolRegistryProvider,
	}
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
	if c.toolRegistryProvider == nil {
		return nil // Or handle error appropriately
	}
	return c.toolRegistryProvider.GetToolRegistry()
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


