package prompts

import "fmt"

// PromptManager manages prompts.
type PromptManager struct {
	prompts map[string]string
}

// NewPromptManager creates a new PromptManager.
func NewPromptManager() *PromptManager {
	return &PromptManager{
		prompts: make(map[string]string),
	}
}

// AddPrompt adds a prompt to the manager.
func (pm *PromptManager) AddPrompt(name, text string) {
	pm.prompts[name] = text
}

// GetPrompt retrieves a prompt by name.
func (pm *PromptManager) GetPrompt(name string) (string, error) {
	prompt, ok := pm.prompts[name]
	if !ok {
		return "", fmt.Errorf("prompt not found: %s", name)
	}
	return prompt, nil
}
