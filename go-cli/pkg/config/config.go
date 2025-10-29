package config

import (
	"go-ai-agent-v2/go-cli/pkg/mcp"
)

// ApprovalMode defines the approval mode for tool calls.
type ApprovalMode string

const (
	ApprovalModeDefault  ApprovalMode = "default"
	ApprovalModeAutoEdit ApprovalMode = "autoEdit"
	ApprovalModeYOLO     ApprovalMode = "yolo"
)

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
	McpServers     map[string]mcp.MCPServerConfig
	ApprovalMode   ApprovalMode
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
	mcpServers     map[string]mcp.MCPServerConfig
	approvalMode   ApprovalMode
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
