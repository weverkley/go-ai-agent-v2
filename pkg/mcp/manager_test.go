package mcp_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"go-ai-agent-v2/go-cli/pkg/mcp"
	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/stretchr/testify/assert"
)

// MockToolRegistry implements types.ToolRegistry for testing.
type MockToolRegistry struct {
	mu    sync.RWMutex
	tools map[string]types.Tool
}

func NewMockToolRegistry() types.ToolRegistryInterface {
	return &MockToolRegistry{
		tools: make(map[string]types.Tool),
	}
}

func (m *MockToolRegistry) Register(t types.Tool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.tools[t.Name()]; exists {
		return fmt.Errorf("tool with name '%s' already registered", t.Name())
	}
	m.tools[t.Name()] = t
	return nil
}

func (m *MockToolRegistry) GetTool(name string) (types.Tool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	t, exists := m.tools[name]
	if !exists {
		return nil, fmt.Errorf("no tool found with name '%s'", name)
	}
	return t, nil
}

func (m *MockToolRegistry) GetAllTools() []types.Tool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var allTools []types.Tool
	for _, t := range m.tools {
		allTools = append(allTools, t)
	}
	return allTools
}

func (m *MockToolRegistry) GetAllToolNames() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var names []string
	for name := range m.tools {
		names = append(names, name)
	}
	return names
}

func (m *MockToolRegistry) GetFunctionDeclarationsFiltered(toolNames []string) []*types.FunctionDeclaration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var declarations []*types.FunctionDeclaration
	for _, name := range toolNames {
		if t, ok := m.tools[name]; ok {
			declarations = append(declarations, &types.FunctionDeclaration{
				Name:        t.Name(),
				Description: t.Description(),
				Parameters:  t.Parameters(),
			})
		}
	}
	return declarations
}

// MockConfig implements types.Config for testing purposes.
type MockConfig struct {
	ModelValue string
	McpServers map[string]types.MCPServerConfig
	GetFunc    func(key string) (interface{}, bool)
}

func (m *MockConfig) WithModel(modelName string) types.Config {
	return &MockConfig{ModelValue: modelName}
}

func (m *MockConfig) GetModel() string {
	return m.ModelValue
}

func (m *MockConfig) GetAPIKey() string {
	return "mock-api-key"
}

func (m *MockConfig) CreateTempFile(prefix string) (string, error) {
	return "", nil
}

func (m *MockConfig) CleanupTempFile(path string) error {
	return nil
}

func (m *MockConfig) GetBasePath() string {
	return ""
}

func (m *MockConfig) SetBasePath(path string) {
}

func (m *MockConfig) NewToolConfig() types.ToolConfig {
	return nil
}

func (m *MockConfig) Get(key string) (interface{}, bool) {
	if m.GetFunc != nil {
		return m.GetFunc(key)
	}
	switch key {
	case "modelName":
		return m.ModelValue, true
	case "mcpServers":
		return m.McpServers, true
	default:
		return nil, false
	}
}

// MockMcpClient implements mcp.McpClientInterface for testing purposes.
type MockMcpClient struct {
	name         string
	ConnectFunc  func(timeout time.Duration) error
	CloseFunc    func() error
	GetToolsFunc func() ([]types.Tool, error)
}

func (m *MockMcpClient) Name() string {
	return m.name
}

func (m *MockMcpClient) Connect(timeout time.Duration) error {
	if m.ConnectFunc != nil {
		return m.ConnectFunc(timeout)
	}
	return nil // Default successful connection
}

func (m *MockMcpClient) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil // Default successful close
}

func (m *MockMcpClient) GetTools() ([]types.Tool, error) {
	if m.GetToolsFunc != nil {
		return m.GetToolsFunc()
	}
	return []types.Tool{}, nil // Default no tools
}

func TestMcpClientManager_DiscoverAllMcpTools(t *testing.T) {
	// Save original newMcpClientFactory and restore it after the test
	originalNewMcpClientFactory := mcp.NewMcpClientFactory
	t.Cleanup(func() {
		mcp.NewMcpClientFactory = originalNewMcpClientFactory
	})

	t.Run("should discover and register tools from configured MCP servers", func(t *testing.T) {
		mockToolRegistry := NewMockToolRegistry()
		manager := mcp.NewMcpClientManager(mockToolRegistry)

		mcpServers := map[string]types.MCPServerConfig{
			"server1": {
				Url:          "http://localhost:8080",
				IncludeTools: []string{"toolA", "toolB"},
			},
			"server2": {
				Url:          "http://localhost:8081",
				IncludeTools: []string{"toolC"},
			},
		}

		mockConfig := &MockConfig{
			ModelValue: "gemini-pro",
			McpServers: mcpServers, // Directly set McpServers
		}

		// Override newMcpClient to return a mock client
		mcp.NewMcpClientFactory = func(name, version string, serverConfig types.MCPServerConfig) mcp.McpClientInterface {
			return &MockMcpClient{
				name: name,
				ConnectFunc: func(timeout time.Duration) error {
					return nil // Simulate successful connection
				},
				GetToolsFunc: func() ([]types.Tool, error) {
					// Simulate returning some tools
					if name == "server1" {
						return []types.Tool{
							&MockTool{name: "toolA"},
							&MockTool{name: "toolB"}}, nil
					}
					if name == "server2" {
						return []types.Tool{
							&MockTool{name: "toolC"},
						}, nil
					}
					return []types.Tool{}, nil
				},
			}
		}

		err := manager.DiscoverAllMcpTools(mockConfig)
		assert.NoError(t, err)

		// Verify tools are registered
		_, err = mockToolRegistry.GetTool("toolA")
		assert.NoError(t, err)
		_, err = mockToolRegistry.GetTool("toolB")
		assert.NoError(t, err)
		_, err = mockToolRegistry.GetTool("toolC")
		assert.NoError(t, err)

		// Verify clients are stored
		assert.NotNil(t, manager.GetClient("server1"))
		assert.NotNil(t, manager.GetClient("server2"))
	})

	t.Run("should handle no configured MCP servers gracefully", func(t *testing.T) {
		mockToolRegistry := NewMockToolRegistry()
		manager := mcp.NewMcpClientManager(mockToolRegistry)

		mockConfig := &MockConfig{
			ModelValue: "gemini-pro",
			McpServers: map[string]types.MCPServerConfig{},
		}

		err := manager.DiscoverAllMcpTools(mockConfig)
		assert.NoError(t, err)
	})
}

func TestMcpClientManager_StartServer(t *testing.T) {
	t.Run("should start a local MCP server process", func(t *testing.T) {
		mockToolRegistry := NewMockToolRegistry()
		manager := mcp.NewMcpClientManager(mockToolRegistry)

		// Create a dummy executable for testing
		tempDir := t.TempDir()
		dummyExecutablePath := filepath.Join(tempDir, "dummy_server")
		err := os.WriteFile(dummyExecutablePath, []byte("#!/bin/bash\necho 'dummy server running' && sleep 10"), 0755)
		assert.NoError(t, err)

		serverConfig := types.MCPServerConfig{
			Command: dummyExecutablePath,
			Args:    []string{"--port", "8080"},
			Cwd:     tempDir,
		}

		err = manager.StartServer("test-server", serverConfig)
		assert.NoError(t, err)

		// Give it a moment to start
		time.Sleep(100 * time.Millisecond)

		// Verify the server is running
		assert.NotNil(t, manager.GetRunningServer("test-server"))
		assert.NotNil(t, manager.GetRunningServer("test-server").Process)

		// Clean up
		err = manager.Stop()
		assert.NoError(t, err)
	})

	t.Run("should return error if command is not specified", func(t *testing.T) {
		mockToolRegistry := NewMockToolRegistry()
		manager := mcp.NewMcpClientManager(mockToolRegistry)

		serverConfig := types.MCPServerConfig{
			Command: "",
		}

		err := manager.StartServer("test-server", serverConfig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no command specified")
	})
}

func TestMcpClientManager_Stop(t *testing.T) {
	t.Run("should stop all running local servers and close clients", func(t *testing.T) {
		mockToolRegistry := NewMockToolRegistry()
		manager := mcp.NewMcpClientManager(mockToolRegistry)

		// Start a dummy server
		tempDir := t.TempDir()
		dummyExecutablePath := filepath.Join(tempDir, "dummy_server")
		err := os.WriteFile(dummyExecutablePath, []byte("#!/bin/bash\necho 'dummy server running' && sleep 10"), 0755)
		assert.NoError(t, err)

		serverConfig := types.MCPServerConfig{
			Command: dummyExecutablePath,
			Args:    []string{"--port", "8080"},
			Cwd:     tempDir,
		}
		err = manager.StartServer("test-server", serverConfig)
		assert.NoError(t, err)
		time.Sleep(100 * time.Millisecond) // Give it time to start

		// Add a dummy client
		manager.AddClient("dummy-client", &MockMcpClient{name: "dummy-client"})

		err = manager.Stop()
		assert.NoError(t, err)

		// Verify no running servers or clients
		assert.Nil(t, manager.GetRunningServer("test-server"))
		assert.Nil(t, manager.GetClient("dummy-client"))
	})

	t.Run("should handle errors during client close or process kill", func(t *testing.T) {
		mockToolRegistry := NewMockToolRegistry()
		manager := mcp.NewMcpClientManager(mockToolRegistry)

		// Add a client that returns an error on Close
		mockClient := &MockMcpClient{
			name:      "error-client",
			CloseFunc: func() error { return errors.New("mock close error") },
		}
		manager.AddClient("error-client", mockClient)

		err := manager.Stop()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mock close error")
	})
}

func TestMcpClientManager_ListServers(t *testing.T) {
	// Save original newMcpClientFactory and restore it after the test
	originalNewMcpClientFactory := mcp.NewMcpClientFactory
	t.Cleanup(func() {
		mcp.NewMcpClientFactory = originalNewMcpClientFactory
	})

	t.Run("should list configured MCP servers with their status", func(t *testing.T) {
		mockToolRegistry := NewMockToolRegistry()
		manager := mcp.NewMcpClientManager(mockToolRegistry)

		mcpServers := map[string]types.MCPServerConfig{
			"server1": {
				Url:         "http://localhost:8080",
				Description: "Local server",
			},
			"server2": {
				Url:         "http://remote.com",
				Description: "Remote server",
			},
		}
		mockConfig := &MockConfig{
			ModelValue: "gemini-pro",
			McpServers: mcpServers,
		}

		// Simulate a running local server
		tempDir := t.TempDir()
		dummyExecutablePath := filepath.Join(tempDir, "dummy_server")
		err := os.WriteFile(dummyExecutablePath, []byte("#!/bin/bash\necho 'dummy server running' && sleep 10"), 0755)
		assert.NoError(t, err)
		serverConfig1 := mcpServers["server1"]
		serverConfig1.Command = dummyExecutablePath
		serverConfig1.Cwd = tempDir
		mcpServers["server1"] = serverConfig1 // Update config with command

		err = manager.StartServer("server1", serverConfig1)
		assert.NoError(t, err)
		time.Sleep(100 * time.Millisecond) // Give it time to start

		// Simulate a connected remote client
		manager.AddClient("server2", &MockMcpClient{name: "server2"})

		statuses := manager.ListServers(mockConfig)
		assert.Len(t, statuses, 2)

		// Check server1 status (running)
		s1Status := findServerStatus(statuses, "server1")
		assert.NotNil(t, s1Status)
		assert.Equal(t, "RUNNING", s1Status.Status)
		assert.Equal(t, "http://localhost:8080", s1Status.Url)

		// Check server2 status (connected)
		s2Status := findServerStatus(statuses, "server2")
		assert.NotNil(t, s2Status)
		assert.Equal(t, types.MCPServerStatusConnected, s2Status.Status)
		assert.Equal(t, "http://remote.com", s2Status.Url)

		// Clean up
		err = manager.Stop()
		assert.NoError(t, err)
	})

	t.Run("should list disconnected servers", func(t *testing.T) {
		mockToolRegistry := NewMockToolRegistry()
		manager := mcp.NewMcpClientManager(mockToolRegistry)

		mcpServers := map[string]types.MCPServerConfig{
			"server3": {
				Url:         "http://disconnected.com",
				Description: "Disconnected server",
			},
		}
		mockConfig := &MockConfig{
			ModelValue: "gemini-pro",
			McpServers: mcpServers,
		}

		statuses := manager.ListServers(mockConfig)
		assert.Len(t, statuses, 1)

		s3Status := findServerStatus(statuses, "server3")
		assert.NotNil(t, s3Status)
		assert.Equal(t, types.MCPServerStatusDisconnected, s3Status.Status)
	})
}

func findServerStatus(statuses []types.MCPServerStatus, name string) *types.MCPServerStatus {
	for _, s := range statuses {
		if s.Name == name {
			return &s
		}
	}
	return nil
}

// MockTool implements types.Tool for testing purposes.
type MockTool struct {
	name        string
	description string
	parameters  *types.JsonSchemaObject
}

func (m *MockTool) Name() string {
	return m.name
}

func (m *MockTool) Description() string {
	return m.description
}

func (m *MockTool) ServerName() string {
	return "mock-server"
}

func (m *MockTool) Parameters() *types.JsonSchemaObject {
	return m.parameters
}

func (m *MockTool) Kind() types.Kind {
	return types.KindOther
}

func (m *MockTool) Execute(ctx context.Context, args map[string]any) (types.ToolResult, error) {
	return types.ToolResult{LLMContent: "mock result", ReturnDisplay: "mock result"}, nil
}
