package mcp

// MCPServerStatus represents the connection status of an MCP server.
type MCPServerStatus string

const (
	CONNECTED    MCPServerStatus = "CONNECTED"
	CONNECTING   MCPServerStatus = "CONNECTING"
	DISCONNECTED MCPServerStatus = "DISCONNECTED"
)

// MCPOAuthConfig represents the OAuth configuration for an MCP server.
type MCPOAuthConfig struct {
	// Placeholder for now, add fields as needed.
}

// AuthProviderType represents the authentication provider type.
type AuthProviderType string

// MCPServerConfig represents the configuration for an MCP server.
type MCPServerConfig struct {
	Command       string            `json:"command,omitempty"`
	Args          []string          `json:"args,omitempty"`
	Env           map[string]string `json:"env,omitempty"`
	Cwd           string            `json:"cwd,omitempty"`
	Url           string            `json:"url,omitempty"`
	HttpUrl       string            `json:"httpUrl,omitempty"`
	Headers       map[string]string `json:"headers,omitempty"`
	Tcp           string            `json:"tcp,omitempty"`
	Timeout       int               `json:"timeout,omitempty"`
	Trust         bool              `json:"trust,omitempty"`
	Description   string            `json:"description,omitempty"`
	IncludeTools  []string          `json:"includeTools,omitempty"`
	ExcludeTools  []string          `json:"excludeTools,omitempty"`
	Extension     *ExtensionInfo    `json:"extension,omitempty"` // Reference to the extension that provided this config
	Oauth         *MCPOAuthConfig   `json:"oauth,omitempty"`
	AuthProviderType AuthProviderType `json:"authProviderType,omitempty"`
	TargetAudience string           `json:"targetAudience,omitempty"`
	TargetServiceAccount string     `json:"targetServiceAccount,omitempty"`
}

// ExtensionInfo represents basic information about an extension.
type ExtensionInfo struct {
	Name string `json:"name"`	
}
