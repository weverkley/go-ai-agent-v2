package tools

import (
	"fmt"

	"go-ai-agent-v2/go-cli/pkg/services"
	"go-ai-agent-v2/go-cli/pkg/types"
)

// RegisterAllTools creates a new ToolRegistry and registers all the available tools.
func RegisterAllTools(fs services.FileSystemService) *types.ToolRegistry {
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
	if err := registry.Register(NewListDirectoryTool(fs)); err != nil {
		fmt.Printf("Error registering ListDirectoryTool: %v\n", err)
	}
	if err := registry.Register(NewGetCurrentBranchTool()); err != nil {
		fmt.Printf("Error registering GetCurrentBranchTool: %v\n", err)
	}
	if err := registry.Register(NewGetRemoteURLTool()); err != nil {
		fmt.Printf("Error registering GetRemoteURLTool: %v\n", err)
	}
	if err := registry.Register(NewCheckoutBranchTool()); err != nil {
		fmt.Printf("Error registering CheckoutBranchTool: %v\n", err)
	}
	if err := registry.Register(NewPullTool()); err != nil {
		fmt.Printf("Error registering PullTool: %v\n", err)
	}
	if err := registry.Register(NewExecuteCommandTool()); err != nil {
		fmt.Printf("Error registering ExecuteCommandTool: %v\n", err)
	}
	if err := registry.Register(NewFindUnusedCodeTool()); err != nil {
		fmt.Printf("Error registering FindUnusedCodeTool: %v\n", err)
	}
	if err := registry.Register(NewExtractFunctionTool()); err != nil {
		fmt.Printf("Error registering ExtractFunctionTool: %v\n", err)
	}

	return registry
}
