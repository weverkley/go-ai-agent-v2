package prompts

import (
	"fmt"
)

// DiscoveredMCPPrompt represents a prompt discovered from an MCP server.
type DiscoveredMCPPrompt struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	ServerName  string `json:"serverName"`
}

// PromptManager manages prompts.
type PromptManager struct {
	prompts map[string]DiscoveredMCPPrompt
}

// NewPromptManager creates a new PromptManager.
func NewPromptManager() *PromptManager {
	return &PromptManager{
		prompts: make(map[string]DiscoveredMCPPrompt),
	}
}

// AddPrompt adds a prompt to the manager.
func (pm *PromptManager) AddPrompt(prompt DiscoveredMCPPrompt) {
	pm.prompts[prompt.Name] = prompt
}

// GetPrompt retrieves a prompt by name.
func (pm *PromptManager) GetPrompt(name string) (DiscoveredMCPPrompt, error) {
	prompt, ok := pm.prompts[name]
	if !ok {
		return DiscoveredMCPPrompt{}, fmt.Errorf("prompt not found: %s", name)
	}
	return prompt, nil
}
