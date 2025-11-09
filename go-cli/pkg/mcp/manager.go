package mcp

import (
	"fmt"
	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/types"
	"os/exec" // Import os/exec
	"time"
)

// McpClientManager manages the lifecycle of multiple MCP clients and local servers.
type McpClientManager struct {
	clients map[string]*McpClient
	toolRegistry *types.ToolRegistry
	runningServers map[string]*exec.Cmd // Map to store running local server processes
}

// NewMcpClientManager creates a new instance of McpClientManager.
func NewMcpClientManager(toolRegistry *types.ToolRegistry) *McpClientManager {
	return &McpClientManager{
		clients: make(map[string]*McpClient),
		toolRegistry: toolRegistry,
		runningServers: make(map[string]*exec.Cmd),
	}
}

// StartServer starts a local MCP server process.
func (m *McpClientManager) StartServer(name string, serverConfig types.MCPServerConfig) error {
	if serverConfig.Command == "" {
		return fmt.Errorf("MCP server '%s' has no command specified", name)
	}

	cmd := exec.Command(serverConfig.Command, serverConfig.Args...)
	if serverConfig.Cwd != "" {
		cmd.Dir = serverConfig.Cwd
	}
	if len(serverConfig.Env) > 0 {
		cmd.Env = os.Environ() // Start with current environment
		for k, v := range serverConfig.Env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	// Start the command in a new goroutine to avoid blocking
	go func() {
		fmt.Printf("Starting local MCP server '%s': %s %v\n", name, serverConfig.Command, serverConfig.Args)
		if err := cmd.Start(); err != nil {
			fmt.Printf("Error starting local MCP server '%s': %v\n", name, err)
			return
		}
		m.runningServers[name] = cmd
		fmt.Printf("Local MCP server '%s' started with PID %d\n", name, cmd.Process.Pid)

		// Wait for the command to finish (or be stopped)
		if err := cmd.Wait(); err != nil {
			fmt.Printf("Local MCP server '%s' exited with error: %v\n", name, err)
		} else {
			fmt.Printf("Local MCP server '%s' exited normally.\n", name)
		}
		delete(m.runningServers, name) // Clean up after exit
	}()

	return nil
}