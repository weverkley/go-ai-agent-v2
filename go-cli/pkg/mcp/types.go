package mcp

// MCPServerStatus represents the connection status of an MCP server.
type MCPServerStatus string

const (
	CONNECTED    MCPServerStatus = "CONNECTED"
	CONNECTING   MCPServerStatus = "CONNECTING"
	DISCONNECTED MCPServerStatus = "DISCONNECTED"
)

// MCPServerConfig represents the configuration for an MCP server.
type MCPServerConfig struct {
	HttpUrl   string `json:"httpUrl,omitempty"`
	Url       string `json:"url,omitempty"`
	Command   string `json:"command,omitempty"`
	Args      []string `json:"args,omitempty"`
	Extension *ExtensionInfo `json:"extension,omitempty"` // Reference to the extension that provided this config
}

// ExtensionInfo represents basic information about an extension.
type ExtensionInfo struct {
	Name string `json:"name"`	
}
