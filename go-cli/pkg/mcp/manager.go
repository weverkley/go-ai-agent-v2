package mcp

import (
	"fmt"
	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/types"
)

// McpClientManager manages the lifecycle of multiple MCP clients.
type McpClientManager struct {
	clients map[string]*McpClient
	toolRegistry *types.ToolRegistry // Placeholder for now
}

// NewMcpClientManager creates a new instance of McpClientManager.
func NewMcpClientManager(toolRegistry *types.ToolRegistry) *McpClientManager {
	return &McpClientManager{
		clients: make(map[string]*McpClient),
		toolRegistry: toolRegistry,
	}
}

// DiscoverAllMcpTools initiates the tool discovery process for all configured MCP servers.
func (m *McpClientManager) DiscoverAllMcpTools(cliConfig *config.Config) error {
	fmt.Println("Discovering MCP tools...")

	mcpServers := cliConfig.GetMcpServers()
	if len(mcpServers) == 0 {
		fmt.Println("No MCP servers configured.")
		return nil
	}

	for name, serverConfig := range mcpServers {
		fmt.Printf("Connecting to MCP server: %s (URL: %s)\n", name, serverConfig.Url)
		client := NewMcpClient(name, "v1.0", serverConfig) // Pass serverConfig
		
		// Simulate connection
		if err := client.Connect(5*time.Second); err != nil {
			fmt.Printf("Error connecting to MCP server %s: %v\n", name, err)
			continue
		}
		m.clients[name] = client

		// Simulate tool discovery and registration
		fmt.Printf("Simulating tool discovery for %s...\n", name)
		discoveredTools, err := client.GetTools()
		if err != nil {
			fmt.Printf("Error getting simulated tools from MCP server %s: %v\n", name, err)
			continue
		}

		for _, tool := range discoveredTools {
			if err := m.toolRegistry.Register(tool); err != nil {
				fmt.Printf("Error registering tool %s from MCP server %s: %v\n", tool.Name(), name, err)
			} else {
				fmt.Printf("Registered tool: %s from MCP server: %s\n", tool.Name(), name)
			}
		}
	}

	return nil
}

// Stop stops all running local MCP servers and closes all client connections.
func (m *McpClientManager) Stop() error {
	// For now, we will simulate the stop process.
	fmt.Println("Simulating stopping MCP clients...")

	// In a real scenario, this would involve:
	// 1. Iterating through all active McpClient's.
	// 2. Disconnecting from each client.

	return nil
}

// ListServers returns a list of configured MCP servers with their status.
func (m *McpClientManager) ListServers() []types.MCPServerStatus {
	// For now, return a dummy list.
	return []types.MCPServerStatus{
		{
			Name:   "dummy-server",
			Status: types.MCPServerStatusDisconnected,
			Url:    "http://localhost:8080",
			Description: "A dummy MCP server for testing.",
		},
	}
}