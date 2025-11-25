package commands

import (
	"fmt"
	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/extension"
	"go-ai-agent-v2/go-cli/pkg/mcp"
	"go-ai-agent-v2/go-cli/pkg/services" // Add services import
	"go-ai-agent-v2/go-cli/pkg/types"
	"os"
	"strings"
)

// McpCommand provides functionalities for managing MCP servers.
type McpCommand struct {
	workspaceDir     string
	settingsService  *services.SettingsService
	mcpManager       *mcp.McpClientManager
	extensionManager *extension.Manager
	toolRegistry     *types.ToolRegistry
}

// NewMcpCommand creates a new instance of McpCommand.
func NewMcpCommand(toolRegistry *types.ToolRegistry, settingsService *services.SettingsService, extensionManager *extension.Manager) *McpCommand {
	workspaceDir, _ := os.Getwd()
	return &McpCommand{
		workspaceDir:     workspaceDir,
		settingsService:  settingsService,
		mcpManager:       mcp.NewMcpClientManager(toolRegistry),
		extensionManager: extensionManager,
		toolRegistry:     toolRegistry,
	}
}

// ListMcpServers lists all configured MCP servers.
func (c *McpCommand) ListMcpServers() error {
	mcpServersVal, found := c.settingsService.Get("mcpServers")
	if !found || mcpServersVal == nil {
		fmt.Println("No MCP servers configured.")
		return nil
	}
	mcpServers, ok := mcpServersVal.(map[string]types.MCPServerConfig)
	if !ok {
		return fmt.Errorf("mcpServers setting is not of expected type")
	}

	if len(mcpServers) == 0 {
		fmt.Println("No MCP servers configured.")
		return nil
	}

	fmt.Println("Configured MCP Servers:")
	for name, server := range mcpServers {
		fmt.Printf("  - Name: %s\n", name)
		fmt.Printf("    URL: %s\n", server.Url)
		if server.Description != "" {
			fmt.Printf("    Description: %s\n", server.Description)
		}
		fmt.Println()
	}
	return nil
}

// AddMcpServer adds or updates an MCP server configuration.
func (c *McpCommand) AddMcpServer(
	name, commandOrUrl string,
	args []string,
	scope config.SettingScope, // This parameter is currently unused, but kept for compatibility
	transport string,
	env []string,
	header []string,
	timeout int,
	trust bool,
	description string,
	includeTools []string,
	excludeTools []string,
) error {
	// Retrieve current MCP servers
	mcpServersVal, found := c.settingsService.Get("mcpServers")
	var mcpServers map[string]types.MCPServerConfig
	if found && mcpServersVal != nil {
		if ms, ok := mcpServersVal.(map[string]types.MCPServerConfig); ok {
			mcpServers = ms
		} else {
			return fmt.Errorf("mcpServers setting is not of expected type")
		}
	} else {
		mcpServers = make(map[string]types.MCPServerConfig)
	}

	// Create or update the server config
	serverConfig := types.MCPServerConfig{
		Description:  description,
		Trust:        trust,
		Args:         args,
		Env:          make(map[string]string),
		Headers:      make(map[string]string),
		Timeout:      timeout,
		IncludeTools: includeTools,
		ExcludeTools: excludeTools,
	}

	// Parse env variables
	for _, e := range env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			serverConfig.Env[parts[0]] = parts[1]
		}
	}

	// Parse headers
	for _, h := range header {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) == 2 {
			serverConfig.Headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	switch transport {
	case "stdio":
		serverConfig.Command = commandOrUrl
	case "http":
		serverConfig.HttpUrl = commandOrUrl
	case "tcp":
		serverConfig.Tcp = commandOrUrl
	default:
		return fmt.Errorf("unsupported transport type: %s", transport)
	}

	mcpServers[name] = serverConfig

	// Save updated MCP servers
	if err := c.settingsService.Set("mcpServers", mcpServers); err != nil {
		return fmt.Errorf("failed to update mcpServers setting: %w", err)
	}

	// Persist settings to file
	if err := c.settingsService.Save(); err != nil {
		return fmt.Errorf("failed to save settings: %w", err)
	}

	fmt.Printf("MCP server \"%s\" added/updated successfully (transport: %s).\n", name, transport)
	return nil
}

// RemoveMcpItem removes an MCP item.
func (c *McpCommand) RemoveMcpItem(name string) error {
	// Retrieve current MCP servers
	mcpServersVal, found := c.settingsService.Get("mcpServers")
	var mcpServers map[string]types.MCPServerConfig
	if found && mcpServersVal != nil {
		if ms, ok := mcpServersVal.(map[string]types.MCPServerConfig); ok {
			mcpServers = ms
		} else {
			return fmt.Errorf("mcpServers setting is not of expected type")
		}
	} else {
		return fmt.Errorf("no MCP servers configured")
	}

	if _, exists := mcpServers[name]; !exists {
		return fmt.Errorf("MCP server \"%s\" not found", name)
	}

	delete(mcpServers, name)

	// Save updated MCP servers
	if err := c.settingsService.Set("mcpServers", mcpServers); err != nil {
		return fmt.Errorf("failed to update mcpServers setting: %w", err)
	}

	// Persist settings to file
	if err := c.settingsService.Save(); err != nil {
		return fmt.Errorf("failed to save settings: %w", err)
	}

	fmt.Printf("MCP server \"%s\" removed successfully.\n", name)
	return nil
}
