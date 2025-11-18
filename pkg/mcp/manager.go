package mcp

import (
	"fmt"
	"go-ai-agent-v2/go-cli/pkg/telemetry" // Added
	"go-ai-agent-v2/go-cli/pkg/types"
	"os"
	"os/exec"
	"sync" // Import sync package
	"time"
)

// McpClientManager manages the lifecycle of multiple MCP clients and local servers.
type McpClientManager struct {
	mu             sync.RWMutex // Mutex to protect concurrent access
	clients        map[string]*McpClient
	toolRegistry   types.ToolRegistryInterface
	runningServers map[string]*exec.Cmd // Map to store running local server processes
}

// NewMcpClientManager creates a new instance of McpClientManager.
func NewMcpClientManager(toolRegistry types.ToolRegistryInterface) *McpClientManager {
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
		telemetry.LogDebugf("Starting local MCP server '%s': %s %v", name, serverConfig.Command, serverConfig.Args)
		if err := cmd.Start(); err != nil {
			telemetry.LogErrorf("Error starting local MCP server '%s': %v", name, err)
			return
		}
		m.mu.Lock()
		m.runningServers[name] = cmd
		m.mu.Unlock()
		telemetry.LogDebugf("Local MCP server '%s' started with PID %d", name, cmd.Process.Pid)

		// Wait for the command to finish (or be stopped)
		if err := cmd.Wait(); err != nil {
			telemetry.LogErrorf("Local MCP server '%s' exited with error: %v", name, err)
		} else {
			telemetry.LogDebugf("Local MCP server '%s' exited normally.", name)
		}
		m.mu.Lock()
		delete(m.runningServers, name) // Clean up after exit
		m.mu.Unlock()
	}()

	return nil
}

// DiscoverAllMcpTools initiates the tool discovery process for all configured MCP servers.
func (m *McpClientManager) DiscoverAllMcpTools(cliConfig types.Config) error {
	telemetry.LogDebugf("Discovering MCP tools...")

	mcpServersVal, found := cliConfig.Get("mcpServers")
	if !found || mcpServersVal == nil {
		telemetry.LogDebugf("No MCP servers configured.")
		return nil
	}
	mcpServers, ok := mcpServersVal.(map[string]types.MCPServerConfig)
	if !ok {
		return fmt.Errorf("mcpServers in config is not of expected type")
	}

	for name, serverConfig := range mcpServers {
		telemetry.LogDebugf("Connecting to MCP server: %s (URL: %s)", name, serverConfig.Url)
		client := NewMcpClient(name, "v1.0", serverConfig) // Pass serverConfig
		
		// Simulate connection
		if err := client.Connect(5*time.Second); err != nil {
			telemetry.LogErrorf("Error connecting to MCP server %s: %v", name, err)
			continue
		}
		m.mu.Lock()
		m.clients[name] = client
		m.mu.Unlock()

		// Simulate tool discovery and registration
		telemetry.LogDebugf("Simulating tool discovery for %s...", name)
		discoveredTools, err := client.GetTools()
		if err != nil {
			telemetry.LogErrorf("Error getting simulated tools from MCP server %s: %v", name, err)
			continue
		}

		for _, tool := range discoveredTools {
			if err := m.toolRegistry.Register(tool); err != nil {
				telemetry.LogErrorf("Error registering tool %s from MCP server %s: %v", tool.Name(), name, err)
			} else {
				telemetry.LogDebugf("Registered tool: %s from MCP server: %s", tool.Name(), name)
			}
		}
	}

	return nil
}

// Stop stops all running local MCP servers and closes all client connections.
func (m *McpClientManager) Stop() error {
	telemetry.LogDebugf("Stopping MCP clients and local servers...")

	var allErrors []error

	m.mu.Lock()
	defer m.mu.Unlock()

	// Close client connections
	for name, client := range m.clients {
		if err := client.Close(); err != nil {
			allErrors = append(allErrors, fmt.Errorf("error closing client %s: %w", name, err))
		}
		delete(m.clients, name)
	}

	// Terminate local server processes
	for name, cmd := range m.runningServers {
		if cmd.Process != nil {
			telemetry.LogDebugf("Terminating local MCP server '%s' (PID: %d)...", name, cmd.Process.Pid)
			if err := cmd.Process.Kill(); err != nil {
				allErrors = append(allErrors, fmt.Errorf("error killing process for server %s (PID %d): %w", name, cmd.Process.Pid, err))
			} else {
				telemetry.LogDebugf("Local MCP server '%s' (PID: %d) terminated.", name, cmd.Process.Pid)
			}
		}
		delete(m.runningServers, name)
	}

	if len(allErrors) > 0 {
		return fmt.Errorf("multiple errors occurred during MCP stop: %v", allErrors)
	}

	telemetry.LogDebugf("All MCP clients and local servers stopped.")
	return nil
}

// ListServers returns a list of configured MCP servers with their status.
func (m *McpClientManager) ListServers(cliConfig types.Config) []types.MCPServerStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var statuses []types.MCPServerStatus

	mcpServersVal, found := cliConfig.Get("mcpServers")
	if !found || mcpServersVal == nil {
		return statuses
	}
	mcpServers, ok := mcpServersVal.(map[string]types.MCPServerConfig)
	if !ok {
		return statuses // Return empty if type assertion fails
	}
	for name, serverConfig := range mcpServers {
		currentStatus := types.MCPServerStatus{
			Name:        name,
			Url:         serverConfig.Url,
			Description: serverConfig.Description,
			Status:      types.MCPServerStatusDisconnected, // Default to disconnected
		}

		// Check if client is connected
		if _, ok := m.clients[name]; ok {
			currentStatus.Status = types.MCPServerStatusConnected
		}

		// Check if local server process is running
		if cmd, ok := m.runningServers[name]; ok && cmd.Process != nil {
			if currentStatus.Status == types.MCPServerStatusDisconnected { // Only if not already connected
				currentStatus.Status = "RUNNING" // Custom status for running local server
			} else {
				currentStatus.Status = types.MCPServerStatusConnected // Connected and running
			}
		}
		statuses = append(statuses, currentStatus)
	}

	return statuses
}

// GetClient returns a client by name for testing purposes.
func (m *McpClientManager) GetClient(name string) *McpClient {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.clients[name]
}

// AddClient adds a client to the manager for testing purposes.
func (m *McpClientManager) AddClient(name string, client *McpClient) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clients[name] = client
}

// GetRunningServer returns a running server command by name for testing purposes.
func (m *McpClientManager) GetRunningServer(name string) *exec.Cmd {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.runningServers[name]
}