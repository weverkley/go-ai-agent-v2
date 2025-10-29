package config

import (
	"fmt"
	"go-ai-agent-v2/go-cli/pkg/mcp"
)

// Settings represents the application settings.
type Settings struct {
	// Add fields for various settings, e.g., extension paths, API keys, etc.
	ExtensionPaths []string
	McpServers     map[string]mcp.MCPServerConfig `json:"mcpServers,omitempty"`
}

// LoadSettings loads the application settings from various sources.
// For now, it returns default settings.
func LoadSettings(workspaceDir string) *Settings {
	fmt.Printf("Loading settings for workspace: %s (placeholder)\n", workspaceDir)
	// In a real implementation, this would load from config files, environment variables, etc.
	return &Settings{
		ExtensionPaths: []string{fmt.Sprintf("%s/.gemini/extensions", workspaceDir)},
		McpServers: map[string]mcp.MCPServerConfig{
			"default-mcp-server": {
				HttpUrl: "http://localhost:8080",
			},
		},
	}
}
