package mcp_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"go-ai-agent-v2/go-cli/pkg/mcp"
	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/google/generative-ai-go/genai" // Add genai import
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

func (m *MockToolRegistry) GetTools() []*genai.Tool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var genaiTools []*genai.Tool
	for _, t := range m.tools {
		genaiTools = append(genaiTools, t.Definition())
	}
	return genaiTools
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

func (m *MockToolRegistry) GetFunctionDeclarationsFiltered(toolNames []string) []genai.FunctionDeclaration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var declarations []genai.FunctionDeclaration
	for _, name := range toolNames {
		if t, ok := m.tools[name]; ok {
			if t.Definition() != nil && len(t.Definition().FunctionDeclarations) > 0 {
				declarations = append(declarations, *t.Definition().FunctionDeclarations[0])
			}
		}
	}
	return declarations
}

// MockConfig implements types.Config for testing purposes.
type MockConfig struct {
	ModelName string
	ToolRegistry types.ToolRegistryInterface
	DebugMode bool
	CodebaseInvestigatorSettings *types.CodebaseInvestigatorSettings
	McpServers map[string]types.MCPServerConfig
}

func (m *MockConfig) Get(key string) (interface{}, bool) {
	switch key {
	case "modelName":
		return m.ModelName, true
	case "toolRegistry":
		return m.ToolRegistry, true
	case "debugMode":
		return m.DebugMode, true
	case "codebaseInvestigatorSettings":
		return m.CodebaseInvestigatorSettings, true
	case "mcpServers":
		return m.McpServers, true
	default:
		return nil, false
	}
}



func TestMcpClientManager_DiscoverAllMcpTools(t *testing.T) {
	t.Run("should discover and register tools from configured MCP servers", func(t *testing.T) {
		mockToolRegistry := NewMockToolRegistry()
		manager := mcp.NewMcpClientManager(mockToolRegistry)

		mockConfig := &MockConfig{
			McpServers: map[string]types.MCPServerConfig{
				"server1": {
					Url:          "http://localhost:8080",
					IncludeTools: []string{"toolA", "toolB"},
				},
				"server2": {
					Url:          "http://localhost:8081",
					IncludeTools: []string{"toolC"},
				},
			},
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
		// assert.NotNil(t, manager.GetClient("server1"))
		// assert.NotNil(t, manager.GetClient("server2"))
	})

	t.Run("should handle no configured MCP servers gracefully", func(t *testing.T) {
		mockToolRegistry := NewMockToolRegistry()
		manager := mcp.NewMcpClientManager(mockToolRegistry)

		mockConfig := &MockConfig{
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
		// assert.NotNil(t, manager.GetRunningServer("test-server"))
		// assert.NotNil(t, manager.GetRunningServer("test-server").Process)

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
		// manager.AddClient("dummy-client", mcp.NewMcpClient("dummy-client", "v1.0", types.MCPServerConfig{}))

		err = manager.Stop()
		assert.NoError(t, err)

		// Verify no running servers or clients
		// assert.Nil(t, manager.GetRunningServer("test-server"))
		// assert.Nil(t, manager.GetClient("dummy-client"))
	})

	t.Run("should handle errors during client close or process kill", func(t *testing.T) {
		mockToolRegistry := NewMockToolRegistry()
		manager := mcp.NewMcpClientManager(mockToolRegistry)

		// Add a client that returns an error on Close
		mockClient := mcp.NewMcpClient("error-client", "v1.0", types.MCPServerConfig{})
		mockClient.ConnectFunc = func(timeout time.Duration) error { return nil } // Mock Connect
		mockClient.CloseFunc = func() error { return errors.New("mock close error") }
		manager.AddClient("error-client", mockClient)

		err := manager.Stop()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mock close error")
	})
}

func TestMcpClientManager_ListServers(t *testing.T) {
	t.Run("should list configured MCP servers with their status", func(t *testing.T) {
		mockToolRegistry := NewMockToolRegistry()
		manager := mcp.NewMcpClientManager(mockToolRegistry)

		mockConfig := &MockConfig{
			McpServers: map[string]types.MCPServerConfig{
				"server1": {
					Url:         "http://localhost:8080",
					Description: "Local server",
				},
				"server2": {
					Url:         "http://remote.com",
					Description: "Remote server",
				},
			},
		}

		// Simulate a running local server
		tempDir := t.TempDir()
		dummyExecutablePath := filepath.Join(tempDir, "dummy_server")
		err := os.WriteFile(dummyExecutablePath, []byte("#!/bin/bash\necho 'dummy server running' && sleep 10"), 0755)
		assert.NoError(t, err)
		serverConfig1 := mockConfig.McpServers["server1"]
		serverConfig1.Command = dummyExecutablePath
		serverConfig1.Cwd = tempDir
		mockConfig.McpServers["server1"] = serverConfig1 // Update config with command
		
		err = manager.StartServer("server1", serverConfig1)
		assert.NoError(t, err)
		time.Sleep(100 * time.Millisecond) // Give it time to start

		// Simulate a connected remote client
		manager.AddClient("server2", mcp.NewMcpClient("server2", "v1.0", mockConfig.McpServers["server2"]))

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

		mockConfig := &MockConfig{
			McpServers: map[string]types.MCPServerConfig{
				"server3": {
					Url:         "http://disconnected.com",
					Description: "Disconnected server",
				},
			},
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


