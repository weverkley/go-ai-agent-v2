package tools

import (
	"net/http"

	"go-ai-agent-v2/go-cli/pkg/services"
	"go-ai-agent-v2/go-cli/pkg/telemetry"
	"go-ai-agent-v2/go-cli/pkg/types"
)

// RegisterAllTools creates a new ToolRegistry and registers all the available tools.
func RegisterAllTools(fs services.FileSystemService, shellService *services.ShellExecutionService, settingsService types.SettingsServiceIface) *types.ToolRegistry {
	registry := types.NewToolRegistry()

	if err := registry.Register(NewGrepTool()); err != nil {
		telemetry.LogErrorf("Error registering GrepTool: %v", err)
	}
	if err := registry.Register(NewGlobTool(fs)); err != nil { // Updated
		telemetry.LogErrorf("Error registering GlobTool: %v", err)
	}
	if err := registry.Register(NewReadFileTool()); err != nil {
		telemetry.LogErrorf("Error registering ReadFileTool: %v", err)
	}
	if err := registry.Register(NewWriteFileTool(fs)); err != nil {
		telemetry.LogErrorf("Error registering WriteFileTool: %v", err)
	}
	if err := registry.Register(NewReadManyFilesTool(fs)); err != nil {
		telemetry.LogErrorf("Error registering ReadManyFilesTool: %v", err)
	}
	if err := registry.Register(NewSmartEditTool(fs)); err != nil {
		telemetry.LogErrorf("Error registering SmartEditTool: %v", err)
	}
	if err := registry.Register(NewWebFetchTool()); err != nil {
		telemetry.LogErrorf("Error registering WebFetchTool: %v", err)
	}
	if err := registry.Register(NewWebSearchTool(settingsService, http.DefaultClient, nil)); err != nil { // Updated
		telemetry.LogErrorf("Error registering WebSearchTool: %v", err)
	}
	if err := registry.Register(NewMemoryTool()); err != nil {
		telemetry.LogErrorf("Error registering MemoryTool: %v", err)
	}
	if err := registry.Register(NewWriteTodosTool()); err != nil {
		telemetry.LogErrorf("Error registering WriteTodosTool: %v", err)
	}
	if err := registry.Register(NewListDirectoryTool(fs)); err != nil {
		telemetry.LogErrorf("Error registering ListDirectoryTool: %v", err)
	}
	if err := registry.Register(NewGetCurrentBranchTool()); err != nil {
		telemetry.LogErrorf("Error registering GetCurrentBranchTool: %v", err)
	}
	if err := registry.Register(NewGetRemoteURLTool()); err != nil {
		telemetry.LogErrorf("Error registering GetRemoteURLTool: %v", err)
	}
	if err := registry.Register(NewCheckoutBranchTool()); err != nil {
		telemetry.LogErrorf("Error registering NewCheckoutBranchTool: %v", err)
	}
	if err := registry.Register(NewPullTool()); err != nil {
		telemetry.LogErrorf("Error registering PullTool: %v", err)
	}
	if err := registry.Register(NewExecuteCommandTool(shellService)); err != nil {
		telemetry.LogErrorf("Error registering ExecuteCommandTool: %v", err)
	}
	if err := registry.Register(NewFindUnusedCodeTool()); err != nil {
		telemetry.LogErrorf("Error registering FindUnusedCodeTool: %v", err)
	}
	if err := registry.Register(NewExtractFunctionTool(fs)); err != nil { // Updated
		telemetry.LogErrorf("Error registering ExtractFunctionTool: %v", err)
	}
	if err := registry.Register(NewLsTool()); err != nil {
		telemetry.LogErrorf("Error registering LsTool: %v", err)
	}
	if err := registry.Register(NewUserConfirmTool()); err != nil {
		telemetry.LogErrorf("Error registering UserConfirmTool: %v", err)
	}

	return registry
}
