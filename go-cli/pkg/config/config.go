package config

import (
	"os"

	"go-ai-agent-v2/go-cli/pkg/core/agents"
	"go-ai-agent-v2/go-cli/pkg/types" // Import the new types package
	"go-ai-agent-v2/go-cli/pkg/tools"
)

// WorkspaceContext defines an interface for accessing workspace-related information.
type WorkspaceContext interface {
	GetDirectories() []string
}

// FileService defines an interface for file system operations.
type FileService interface {
	// shouldIgnoreFile checks if a file should be ignored based on filtering options.
	// This is a placeholder and needs a proper implementation.
	ShouldIgnoreFile(filePath string, options agents.FileFilteringOptions) bool
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
}

// Config represents the application configuration.
type Config struct {
	sessionID      string
	embeddingModel string
	targetDir      string
	debugMode      bool
	model          string
	mcpServers     map[string]types.MCPServerConfig
	approvalMode   types.ApprovalMode // Use ApprovalMode from types package
	telemetry      *TelemetrySettings
	output         *OutputSettings
}

// NewConfig creates a new Config instance.
func NewConfig(params *ConfigParameters) *Config {
	return &Config{
		sessionID:      params.SessionID,
		embeddingModel: params.EmbeddingModel,
		targetDir:      params.TargetDir,
		debugMode:      params.DebugMode,
		model:          params.Model,
		mcpServers:     params.McpServers,
		approvalMode:   params.ApprovalMode,
		telemetry:      params.Telemetry,
		output:         params.Output,
	}
}

// GetToolRegistry returns the global tool registry.
// This is a temporary placeholder and should be replaced with a proper
// mechanism to access the global tool registry.
func (c *Config) GetToolRegistry() *tools.ToolRegistry {
	// For now, return a new empty registry.
	// In a real scenario, this would return the application's main tool registry.
	return tools.NewToolRegistry()
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
	return &dummyFileService{}
}

// dummyFileService is a placeholder implementation of FileService.
type dummyFileService struct{}

func (d *dummyFileService) ShouldIgnoreFile(filePath string, options agents.FileFilteringOptions) bool {
	// For now, always return false.
	return false
}

