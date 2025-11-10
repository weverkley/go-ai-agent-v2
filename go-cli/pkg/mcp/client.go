package mcp

import (
	"fmt"
	"go-ai-agent-v2/go-cli/pkg/types"
	"time"
)

// McpClient represents a simplified MCP client.
type McpClient struct {
	name    string
	version string
	config  types.MCPServerConfig // Store the config

	// For testing purposes, allow overriding Connect and Close behavior
	ConnectFunc func(timeout time.Duration) error
	CloseFunc   func() error
}

// NewMcpClient creates a new instance of McpClient.
func NewMcpClient(name, version string, config types.MCPServerConfig) *McpClient {
	client := &McpClient{name: name, version: version, config: config}
	// Set default implementations
	client.ConnectFunc = client.defaultConnect
	client.CloseFunc = client.defaultClose
	return client
}

// defaultConnect is the default implementation for Connect.
func (c *McpClient) defaultConnect(timeout time.Duration) error {
	fmt.Printf("Simulating connection to MCP server (name: %s, config: %+v) with timeout %v\n", c.name, c.config, timeout)
	// For now, always succeed.
	return nil
}

// Connect simulates connecting to an MCP server.
func (c *McpClient) Connect(timeout time.Duration) error {
	return c.ConnectFunc(timeout)
}

// Ping simulates pinging the MCP server.
func (c *McpClient) Ping() error {
	fmt.Println("Simulating ping to MCP server.")
	return nil
}

// defaultClose is the default implementation for Close.
func (c *McpClient) defaultClose() error {
	fmt.Println("Simulating closing MCP client.")
	return nil
}

// Close simulates closing the MCP client connection.
func (c *McpClient) Close() error {
	return c.CloseFunc()
}

// GetTools simulates retrieving tools from the MCP server.
func (c *McpClient) GetTools() ([]types.Tool, error) {
	var tools []types.Tool
	for _, toolName := range c.config.IncludeTools {
		dummyTool := types.NewBaseDeclarativeTool(
			toolName,
			toolName,
			fmt.Sprintf("Dummy tool from MCP server %s", c.name),
			types.KindOther,
			types.JsonSchemaObject{
				Type: "object",
				Properties: map[string]types.JsonSchemaProperty{},
			},
			false,
			false,
			nil,
		)
		tools = append(tools, dummyTool)
	}
	return tools, nil
}

