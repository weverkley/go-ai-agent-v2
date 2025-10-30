package tools

import (
	"fmt"

	"github.com/google/generative-ai-go/genai"
)

// ToolRegistry holds all the registered tools.
type ToolRegistry struct {
	tools map[string]Tool
}

// NewToolRegistry creates a new instance of ToolRegistry.
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]Tool),
	}
}

// Register adds a tool to the registry.
func (r *ToolRegistry) Register(tool Tool) error {
	if _, exists := r.tools[tool.Name()]; exists {
		return fmt.Errorf("tool with name '%s' already registered", tool.Name())
	}
	r.tools[tool.Name()] = tool
	return nil
}

// GetTool retrieves a tool by its name.
func (r *ToolRegistry) GetTool(name string) (Tool, error) {
	tool, exists := r.tools[name]
	if !exists {
		return nil, fmt.Errorf("no tool found with name '%s'", name)
	}
	return tool, nil
}

// GetTools returns all registered tools as a slice of genai.Tool.
func (r *ToolRegistry) GetTools() []*genai.Tool {
	var toolDefs []*genai.Tool
	for _, tool := range r.tools {
		toolDefs = append(toolDefs, tool.Definition())
	}
	return toolDefs
}

// GetAllRegisteredTools returns all registered tools as a slice of tools.Tool.
func (r *ToolRegistry) GetAllRegisteredTools() []Tool {
	var registeredTools []Tool
	for _, tool := range r.tools {
		registeredTools = append(registeredTools, tool)
	}
	return registeredTools
}

// GetAllToolNames returns a slice of all registered tool names.
func (tr *ToolRegistry) GetAllToolNames() []string {
	names := make([]string, 0, len(tr.tools))
	for name := range tr.tools {
		names = append(names, name)
	}
	return names
}

// GetFunctionDeclarationsFiltered returns FunctionDeclarations for a given list of tool names.
func (tr *ToolRegistry) GetFunctionDeclarationsFiltered(toolNames []string) []genai.FunctionDeclaration {
	var declarations []genai.FunctionDeclaration
	for _, name := range toolNames {
		if tool, ok := tr.tools[name]; ok {
			if tool.Definition() != nil && len(tool.Definition().FunctionDeclarations) > 0 {
				declarations = append(declarations, *tool.Definition().FunctionDeclarations[0])
			}
		}
	}
	return declarations
}
