package commands

import (
	"fmt"
	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/extension"
	"go-ai-agent-v2/go-cli/pkg/mcp"
	"go-ai-agent-v2/go-cli/pkg/types"
	"os"
	"strings"
	"time"
)

const (
	COLOR_GREEN = "\u001b[32m"
	COLOR_YELLOW = "\u001b[33m"
	COLOR_RED = "\u001b[31m"
	RESET_COLOR = "\u001b[0m"
)

// McpCommand represents the MCP command group.
type McpCommand struct {
	// Dependencies can be added here
}

// NewMcpCommand creates a new instance of McpCommand.
func NewMcpCommand() *McpCommand {
	return &McpCommand{}
}

// getMcpServersFromConfig loads and merges MCP server configurations.
func (c *McpCommand) getMcpServersFromConfig(workspaceDir string) (map[string]types.MCPServerConfig, error) {
	settings := config.LoadSettings(workspaceDir)
	extensionManager := extension.NewExtensionManager(workspaceDir)
	extensions, err := extensionManager.LoadExtensions()
	if err != nil {
		return nil, fmt.Errorf("failed to load extensions: %w", err)
	}

	mcpServers := make(map[string]types.MCPServerConfig)

	// Merge MCP servers from settings
	for k, v := range settings.McpServers {
		mcpServers[k] = v
	}

	// Merge MCP servers from extensions
	for _, ext := range extensions {
		for k, v := range ext.McpServers {
			if _, exists := mcpServers[k]; !exists {
				// Only add if not already defined in settings
				mcpServers[k] = v
			}
		}
	}
	return mcpServers, nil
}

// testMCPConnection simulates testing an MCP connection.
func (c *McpCommand) testMCPConnection(serverName string, config types.MCPServerConfig) (types.MCPServerStatus, error) {
	client := mcp.NewMcpClient("mcp-test-client", "0.0.1")

	// Simulate transport creation (for now, just a placeholder)
	// In a real implementation, createTransport would be called here.

	err := client.Connect(config, 5*time.Second) // 5s timeout
	if err != nil {
		client.Close()
		return types.DISCONNECTED, nil // Return nil error as status indicates disconnection
	}

	err = client.Ping()
	if err != nil {
		client.Close()
		return types.DISCONNECTED, nil // Return nil error as status indicates disconnection
	}

	client.Close()
	return types.CONNECTED, nil
}

// getServerStatus gets the status of an MCP server.
func (c *McpCommand) getServerStatus(serverName string, serverConfig types.MCPServerConfig) (types.MCPServerStatus, error) {
	return c.testMCPConnection(serverName, serverConfig)
}

// ListMcpItems lists configured MCP items.
func (c *McpCommand) ListMcpItems() error {
	workspaceDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	mcpServers, err := c.getMcpServersFromConfig(workspaceDir)
	if err != nil {
		return fmt.Errorf("failed to get MCP servers from config: %w", err)
	}

	serverNames := make([]string, 0, len(mcpServers))
	for name := range mcpServers {
		serverNames = append(serverNames, name)
	}
	// Sort server names for consistent output
	// sort.Strings(serverNames)

	if len(serverNames) == 0 {
		fmt.Println("No MCP servers configured.")
		return nil
	}

	fmt.Println("Configured MCP servers:")

	for _, serverName := range serverNames {
		server := mcpServers[serverName]

		status, err := c.getServerStatus(serverName, server)
		if err != nil {
			fmt.Printf("Error getting status for %s: %v\n", serverName, err)
			continue
		}

		var statusIndicator string
		var statusText string
		switch status {
		case types.CONNECTED:
			statusIndicator = COLOR_GREEN + "✓" + RESET_COLOR
			statusText = "Connected"
		case types.CONNECTING:
			statusIndicator = COLOR_YELLOW + "…" + RESET_COLOR
			statusText = "Connecting"
		case types.DISCONNECTED:
			statusIndicator = COLOR_RED + "✗" + RESET_COLOR
			statusText = "Disconnected"
		default:
			statusIndicator = "?"
			statusText = "Unknown"
		}

		serverInfo := serverName
		// if server.Extension != nil && server.Extension.Name != "" {
		// 	serverInfo += fmt.Sprintf(" (from %s)", server.Extension.Name)
		// }
		serverInfo += ": "

		if server.HttpUrl != "" {
			serverInfo += fmt.Sprintf("%s (http)", server.HttpUrl)
		} else if server.Url != "" {
			serverInfo += fmt.Sprintf("%s (sse)", server.Url)
		} else if server.Command != "" {
			serverInfo += fmt.Sprintf("%s %s (stdio)", server.Command, strings.Join(server.Args, " "))
		}

		fmt.Printf("%s %s - %s\n", statusIndicator, serverInfo, statusText)
	}

	return nil
}

// AddMcpItem adds a new MCP item.
func (c *McpCommand) AddMcpItem(
	name string,
	commandOrUrl string,
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
	workspaceDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	settings := config.LoadSettings(workspaceDir)
	if settings.McpServers == nil {
		settings.McpServers = make(map[string]types.MCPServerConfig)
	}

	if _, exists := settings.McpServers[name]; exists {
		// If it exists, we are updating it.
		fmt.Printf("MCP server \"%s\" already exists, updating.\n", name)
	}

	// Parse env variables
	envMap := make(map[string]string)
	for _, e := range env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	// Parse headers
	headersMap := make(map[string]string)
	for _, h := range header {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) == 2 {
			headersMap[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	newServer := types.MCPServerConfig{
		Timeout:      timeout,
		Trust:        trust,
		Description:  description,
		IncludeTools: includeTools,
		ExcludeTools: excludeTools,
	}

	switch transport {
	case "sse":
		newServer.Url = commandOrUrl
		newServer.Headers = headersMap
	case "http":
		newServer.HttpUrl = commandOrUrl
		newServer.Headers = headersMap
	case "stdio":
		newServer.Command = commandOrUrl
		newServer.Args = args
		newServer.Env = envMap
	default:
		return fmt.Errorf("unsupported transport type: %s", transport)
	}

	settings.McpServers[name] = newServer

	if err := config.SaveSettings(workspaceDir, settings); err != nil {
		return fmt.Errorf("failed to save settings: %w", err)
	}

	fmt.Printf("MCP server \"%s\" added/updated successfully (transport: %s).\n", name, transport)
	return nil
}

// RemoveMcpItem removes an MCP item.
func (c *McpCommand) RemoveMcpItem(name string) error {
	workspaceDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	settings := config.LoadSettings(workspaceDir)

	if _, exists := settings.McpServers[name]; !exists {
		return fmt.Errorf("MCP server \"%s\" not found", name)
	}

	delete(settings.McpServers, name)

	if err := config.SaveSettings(workspaceDir, settings); err != nil {
		return fmt.Errorf("failed to save settings: %w", err)
	}

	fmt.Printf("MCP server \"%s\" removed successfully.\n", name)
	return nil
}
