package tools

import (
	"fmt"

	"go-ai-agent-v2/go-cli/pkg/types"
)

// RegisterAllTools creates a new ToolRegistry and registers all the available tools.
func RegisterAllTools() *types.ToolRegistry {
	registry := types.NewToolRegistry()

	if err := registry.Register(NewGrepTool()); err != nil {
		fmt.Printf("Error registering GrepTool: %v\n", err)
	}
	if err := registry.Register(NewGlobTool()); err != nil {
		fmt.Printf("Error registering GlobTool: %v\n", err)
	}
	if err := registry.Register(NewReadFileTool()); err != nil {
		fmt.Printf("Error registering ReadFileTool: %v\n", err)
	}
	if err := registry.Register(NewReadManyFilesTool()); err != nil {
		fmt.Printf("Error registering ReadManyFilesTool: %v\n", err)
	}
	if err := registry.Register(NewSmartEditTool()); err != nil {
		fmt.Printf("Error registering SmartEditTool: %v\n", err)
	}
	if err := registry.Register(NewWebFetchTool()); err != nil {
		fmt.Printf("Error registering WebFetchTool: %v\n", err)
	}
	if err := registry.Register(NewWebSearchTool()); err != nil {
		fmt.Printf("Error registering WebSearchTool: %v\n", err)
	}
	if err := registry.Register(NewMemoryTool()); err != nil {
		fmt.Printf("Error registering MemoryTool: %v\n", err)
	}
	if err := registry.Register(NewWriteTodosTool()); err != nil {
		fmt.Printf("Error registering WriteTodosTool: %v\n", err)
	}

	return registry
}