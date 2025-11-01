package commands

import (
	"fmt"
	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/extension"
	"go-ai-agent-v2/go-cli/pkg/mcp"
	"go-ai-agent-v2/go-cli/pkg/types"
	"os"
	"strings"
)

// McpCommand provides functionalities for managing MCP servers.
type McpCommand struct {
	workspaceDir string
	settings     *config.Settings
	mcpManager   *mcp.McpClientManager
	extManager   *extension.ExtensionManager
	toolRegistry *types.ToolRegistry
}

// NewMcpCommand creates a new instance of McpCommand.
func NewMcpCommand(toolRegistry *types.ToolRegistry) *McpCommand {
	workspaceDir, _ := os.Getwd()
	settings := config.LoadSettings(workspaceDir)
	extManager := extension.NewExtensionManager(workspaceDir)
	return &McpCommand{
		workspaceDir: workspaceDir,
		settings:     settings,
		mcpManager:   mcp.NewMcpClientManager(toolRegistry),
		extManager:   extManager,
		toolRegistry: toolRegistry,
	}
}

// ListMcpServers lists all configured MCP servers.
func (c *McpCommand) ListMcpServers() error {
	servers := c.mcpManager.ListServers()
	if len(servers) == 0 {
		fmt.Println("No MCP servers configured.")
		return nil
	}

	fmt.Println("Configured MCP Servers:")
	for _, server := range servers {
		fmt.Printf("  - Name: %s\n", server.Name)
		fmt.Printf("    Status: %s\n", server.Status)
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
	scope config.SettingScope,
	transport string,
	env []string,
	header []string,
	timeout int,
	trust bool,
	description string,
	includeTools []string,
	excludeTools []string,
) error {
	// Load settings again to ensure we have the latest
	settings := config.LoadSettings(c.workspaceDir)

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

	if settings.McpServers == nil {
		settings.McpServers = make(map[string]types.MCPServerConfig)
	}
	settings.McpServers[name] = serverConfig

	// Save updated settings
	if err := config.SaveSettings(c.workspaceDir, settings); err != nil {
		return fmt.Errorf("failed to save settings: %w", err)
	}

	fmt.Printf("MCP server \"%s\" added/updated successfully (transport: %s).\n", name, transport)
	return nil
}

// RemoveMcpItem removes an MCP item.
func (c *McpCommand) RemoveMcpItem(name string) error {
	// Load settings again to ensure we have the latest
	settings := config.LoadSettings(c.workspaceDir)

	if _, exists := settings.McpServers[name]; !exists {
		return fmt.Errorf("MCP server \"%s\" not found", name)
	}

	delete(settings.McpServers, name)

	// Save updated settings
	if err := config.SaveSettings(c.workspaceDir, settings); err != nil {
		return fmt.Errorf("failed to save settings: %w", err)
	}

	fmt.Printf("MCP server \"%s\" removed successfully.\n", name)
	return nil
}
