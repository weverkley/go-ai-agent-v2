package tools

import (
	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/core/agents"
	"go-ai-agent-v2/go-cli/pkg/services"
	"go-ai-agent-v2/go-cli/pkg/telemetry"
	"go-ai-agent-v2/go-cli/pkg/types"
	"net/http"
)

// RegisterAllTools creates a new ToolRegistry and registers all the available tools.
func RegisterAllTools(cfg types.Config, fs services.FileSystemService, shellService services.ShellExecutionService, settingsService types.SettingsServiceIface, workspaceService *services.WorkspaceService) *types.ToolRegistry {
	registry := types.NewToolRegistry()

	// Register standard tools
	// ... (all existing tool registrations remain here)

	// Web Searching tools
	if err := registry.Register(NewWebFetchTool()); err != nil {
		telemetry.LogErrorf("Error registering WebFetchTool: %v", err)
	}
	if err := registry.Register(NewWebSearchTool(settingsService, http.DefaultClient, nil)); err != nil {
		telemetry.LogErrorf("Error registering WebSearchTool: %v", err)
	}
	// Application specific tools
	if err := registry.Register(NewWeatherTool()); err != nil {
		telemetry.LogErrorf("Error registering WeatherTool: %v", err)
	}
	if err := registry.Register(NewWriteTodosTool(settingsService)); err != nil {
		telemetry.LogErrorf("Error registering WriteTodosTool: %v", err)
	}
	if err := registry.Register(NewMemoryTool()); err != nil {
		telemetry.LogErrorf("Error registering MemoryTool: %v", err)
	}
	if err := registry.Register(NewExecuteCommandTool(shellService)); err != nil {
		telemetry.LogErrorf("Error registering ExecuteCommandTool: %v", err)
	}
	if err := registry.Register(NewRunTestsTool(shellService, fs, workspaceService)); err != nil {
		telemetry.LogErrorf("Error registering RunTestsTool: %v", err)
	}
	// File system tools
	if err := registry.Register(NewGrepTool(workspaceService)); err != nil {
		telemetry.LogErrorf("Error registering GrepTool: %v", err)
	}
	if err := registry.Register(NewGlobTool(fs, workspaceService)); err != nil {
		telemetry.LogErrorf("Error registering GlobTool: %v", err)
	}
	if err := registry.Register(NewLsTool(workspaceService)); err != nil {
		telemetry.LogErrorf("Error registering LsTool: %v", err)
	}
	if err := registry.Register(NewReadFileTool(workspaceService)); err != nil {
		telemetry.LogErrorf("Error registering ReadFileTool: %v", err)
	}
	if err := registry.Register(NewWriteFileTool(fs, workspaceService)); err != nil {
		telemetry.LogErrorf("Error registering WriteFileTool: %v", err)
	}
	if err := registry.Register(NewReadManyFilesTool(fs)); err != nil {
		telemetry.LogErrorf("Error registering ReadManyFilesTool: %v", err)
	}
	if err := registry.Register(NewListDirectoryTool(fs)); err != nil {
		telemetry.LogErrorf("Error registering ListDirectoryTool: %v", err)
	}
	if err := registry.Register(NewSmartEditTool(fs, workspaceService)); err != nil {
		telemetry.LogErrorf("Error registering SmartEditTool: %v", err)
	}
	// Agents related tools
	if err := registry.Register(NewFindUnusedCodeTool()); err != nil {
		telemetry.LogErrorf("Error registering FindUnusedCodeTool: %v", err)
	}
	if err := registry.Register(NewFindReferencesTool(workspaceService)); err != nil {
		telemetry.LogErrorf("Error registering FindReferencesTool: %v", err)
	}
	if err := registry.Register(NewRenameSymbolTool(workspaceService)); err != nil {
		telemetry.LogErrorf("Error registering RenameSymbolTool: %v", err)
	}
	if err := registry.Register(NewExtractFunctionTool(fs, workspaceService)); err != nil {
		telemetry.LogErrorf("Error registering ExtractFunctionTool: %v", err)
	}
	// Behavioral tools
	if err := registry.Register(NewUserConfirmTool()); err != nil {
		telemetry.LogErrorf("Error registering UserConfirmTool: %v", err)
	}
	if err := registry.Register(NewTaskCompleteTool()); err != nil {
		telemetry.LogErrorf("Error registering NewTaskCompleteTool: %v", err)
	}
	// Git related tools
	if err := registry.Register(NewGetCurrentBranchTool(services.NewGitService())); err != nil {
		telemetry.LogErrorf("Error registering GetCurrentBranchTool: %v", err)
	}
	if err := registry.Register(NewGetRemoteURLTool(services.NewGitService())); err != nil {
		telemetry.LogErrorf("Error registering GetRemoteURLTool: %v", err)
	}
	if err := registry.Register(NewCheckoutBranchTool(services.NewGitService())); err != nil {
		telemetry.LogErrorf("Error registering NewCheckoutBranchTool: %v", err)
	}
	if err := registry.Register(NewPullTool(services.NewGitService())); err != nil {
		telemetry.LogErrorf("Error registering PullTool: %v", err)
	}
	if err := registry.Register(NewGitCommitTool(services.NewGitService())); err != nil {
		telemetry.LogErrorf("Error registering GitCommitTool: %v", err)
	}

	// Register agents as tools
	agentRegistryVal, ok := cfg.Get("agentRegistry")
	if !ok || agentRegistryVal == nil {
		telemetry.LogErrorf("Agent registry not found in config. Subagents will not be registered.")
		return registry // Continue without subagents
	}
	agentRegistry := agentRegistryVal.(types.AgentRegistryInterface)
	for _, agentDefVal := range agentRegistry.GetAllAgentDefinitions() {
		agentDef, ok := agentDefVal.(agents.AgentDefinition)
		if !ok {
			telemetry.LogErrorf("Failed to cast agent definition to agents.AgentDefinition: %v", agentDefVal)
			continue
		}
		wrappedAgent, err := agents.NewSubagentToolWrapper(agentDef, cfg.(*config.Config), nil) // Assuming nil for messageBus for now
		if err != nil {
			telemetry.LogErrorf("Error wrapping agent %s: %v", agentDef.Name, err)
			continue
		}
		if err := registry.Register(wrappedAgent); err != nil {
			telemetry.LogErrorf("Error registering wrapped agent %s: %v", agentDef.Name, err)
		}
	}

	return registry
}
