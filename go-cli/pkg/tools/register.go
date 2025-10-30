package tools

import (
	"fmt"

	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/core/agents"
)

// RegisterAllTools creates a new ToolRegistry and registers all the available tools.
func RegisterAllTools(cfg *config.Config) *ToolRegistry {
	registry := NewToolRegistry()

	registry.Register(NewGrepTool())
	registry.Register(NewGlobTool())
	registry.Register(NewReadFileTool())
	registry.Register(NewReadManyFilesTool())
	registry.Register(NewSmartEditTool())
	registry.Register(NewWebFetchTool())
	registry.Register(NewWebSearchTool())
	registry.Register(NewMemoryTool())
	registry.Register(NewWriteTodosTool())

	// Register subagents as tools
	subagentTool, err := agents.NewSubagentToolWrapper(agents.CodebaseInvestigatorAgent, cfg, nil) // messageBus is nil for now
	if err != nil {
		fmt.Printf("Error creating CodebaseInvestigatorAgent tool: %v\n", err)
	} else {
		registry.Register(subagentTool)
	}

	return registry
}
