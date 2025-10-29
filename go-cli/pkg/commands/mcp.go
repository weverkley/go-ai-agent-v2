package commands

import (
	"fmt"
	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/extension"
	"go-ai-agent-v2/go-cli/pkg/mcp"
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
func (c *McpCommand) getMcpServersFromConfig(workspaceDir string) (map[string]mcp.MCPServerConfig, error) {
	settings := config.LoadSettings(workspaceDir)
	extensionManager := extension.NewExtensionManager(workspaceDir)
	extensions, err := extensionManager.LoadExtensions()
	if err != nil {
		return nil, fmt.Errorf("failed to load extensions: %w", err)
	}

	mcpServers := make(map[string]mcp.MCPServerConfig)

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
func (c *McpCommand) testMCPConnection(serverName string, config mcp.MCPServerConfig) (mcp.MCPServerStatus, error) {
	client := mcp.NewClient("mcp-test-client", "0.0.1")

	// Simulate transport creation (for now, just a placeholder)
	// In a real implementation, createTransport would be called here.

	err := client.Connect(config, 5*time.Second) // 5s timeout
	if err != nil {
		client.Close()
		return mcp.DISCONNECTED, nil // Return nil error as status indicates disconnection
	}

	err = client.Ping()
	if err != nil {
		client.Close()
		return mcp.DISCONNECTED, nil // Return nil error as status indicates disconnection
	}

	client.Close()
	return mcp.CONNECTED, nil
}

// getServerStatus gets the status of an MCP server.
func (c *McpCommand) getServerStatus(serverName string, serverConfig mcp.MCPServerConfig) (mcp.MCPServerStatus, error) {
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

	fmt.Println("Configured MCP servers:\n")

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
		case mcp.CONNECTED:
			statusIndicator = COLOR_GREEN + "✓" + RESET_COLOR
			statusText = "Connected"
		case mcp.CONNECTING:
			statusIndicator = COLOR_YELLOW + "…" + RESET_COLOR
			statusText = "Connecting"
		case mcp.DISCONNECTED:
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
