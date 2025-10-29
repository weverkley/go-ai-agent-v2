package tools

// RegisterAllTools creates a new ToolRegistry and registers all the available tools.
func RegisterAllTools() *ToolRegistry {
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

	return registry
}
