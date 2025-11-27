package cmd

import (
	"fmt"
	"os"

	"go-ai-agent-v2/go-cli/pkg/commands" // Add commands import
	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/core/agents" // Add this line

	"go-ai-agent-v2/go-cli/pkg/extension"

	"go-ai-agent-v2/go-cli/pkg/services"

	"go-ai-agent-v2/go-cli/pkg/telemetry"

	"go-ai-agent-v2/go-cli/pkg/tools"
	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "go-ai-agent-v2/go-cli",
	Short: "A Go-based CLI for multimodal AI interactions",
	Long:  `A Go-based CLI for multimodal AI interactions, managing extensions and expose MCP servers.`,
}

var Cfg *config.Config
var executorType string
var chatService *services.ChatService           // Declare package-level chatService
var WorkspaceService *services.WorkspaceService // Declare package-level workspaceService
var ExtensionManager *extension.Manager         // Declare package-level extensionManager
var ContextService *services.ContextService     // Declare package-level contextService
var SettingsService types.SettingsServiceIface  // Declare package-level settingsService as interface
var SessionService *services.SessionService     // Declare package-level sessionService

var FSService services.FileSystemService             // Declare package-level FileSystemService
var ShellService services.ShellExecutionService      // Declare package-level ShellExecutionService
var extensionsCliCommand *commands.ExtensionsCommand // Declare package-level extensionsCliCommand

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initServices(projectRoot string) (
	*services.WorkspaceService,
	services.FileSystemService,
	services.ShellExecutionService, // Changed to interface
	*extension.Manager,
	types.SettingsServiceIface, // Changed to interface
	*services.FileFilteringService,
	*services.ContextService,
) {
	workspaceService := services.NewWorkspaceService(projectRoot)
	fsService := services.NewFileSystemService()
	shellService := services.NewShellExecutionService()
	extensionManager := extension.NewManager(projectRoot, fsService, services.NewGitService())
	settingsService := services.NewSettingsService(projectRoot)
	contextService := services.NewContextService(projectRoot)

	fileFilteringService, err := services.NewFileFilteringService(projectRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing FileFilteringService: %v\n", err)
		os.Exit(1)
	}

	return workspaceService, fsService, shellService, extensionManager, settingsService, fileFilteringService, contextService
}

func getTelemetrySettings(settingsService types.SettingsServiceIface) *types.TelemetrySettings {
	return settingsService.GetTelemetrySettings()
}

func registerTools(cfg types.Config, fsService services.FileSystemService, shellService services.ShellExecutionService, settingsService types.SettingsServiceIface, workspaceService *services.WorkspaceService) *types.ToolRegistry {
	return tools.RegisterAllTools(cfg, fsService, shellService, settingsService, workspaceService)
}

func initConfig(
	toolRegistry *types.ToolRegistry,
	agentRegistry types.AgentRegistryInterface,
	telemetrySettings *types.TelemetrySettings,
	codebaseInvestigatorSettings *types.CodebaseInvestigatorSettings,
	testWriterSettings *types.TestWriterSettings,
	workspaceService *services.WorkspaceService,
	fileFilteringService *services.FileFilteringService,
	settingsService types.SettingsServiceIface, // Add settingsService
) *config.Config {
	debugModeVal, _ := settingsService.Get("debugMode")
	debugMode, ok := debugModeVal.(bool)
	if !ok {
		debugMode = false // Fallback if type assertion fails or not found
	}

	modelNameVal, _ := settingsService.Get("model")
	modelName, ok := modelNameVal.(string)
	if !ok {
		modelName = config.DEFAULT_GEMINI_MODEL // Fallback
	}

	params := &config.ConfigParameters{
		DebugMode:    debugMode,
		ModelName:    modelName,
		Telemetry:    telemetrySettings,
		ToolRegistry: toolRegistry,
		AgentRegistry: agentRegistry,
		CodebaseInvestigator: codebaseInvestigatorSettings,
		TestWriterSettings:   testWriterSettings,
	}

	cfg := config.NewConfig(params)
	cfg.WorkspaceContext = workspaceService
	cfg.FileFilteringService = fileFilteringService
	return cfg
}

func registerCommands() {
	RootCmd.AddCommand(todosCmd)
	chatCmd.Run = func(cmd *cobra.Command, args []string) {
		runChatCmd(RootCmd, cmd, args, SettingsService, ShellService)
	}
	RootCmd.AddCommand(chatCmd)
	RootCmd.AddCommand(authCmd)
	RootCmd.AddCommand(modelCmd)
	RootCmd.AddCommand(settingsCmd)
	RootCmd.AddCommand(memoryCmd)
	RootCmd.AddCommand(ExtensionsCmd)

	RootCmd.AddCommand(mcpCmd)
	RootCmd.AddCommand(toolsCmd)
	RootCmd.AddCommand(grepCmd)
	RootCmd.AddCommand(webSearchCmd)
	RootCmd.AddCommand(webFetchCmd)
	RootCmd.AddCommand(readCmd)
	generateCmd.Run = func(cmd *cobra.Command, args []string) {
		runGenerateCmd(cmd, args, SettingsService, ShellService)
	}
	RootCmd.AddCommand(generateCmd)
	RootCmd.AddCommand(smartEditCmd)
	RootCmd.AddCommand(grepCodeCmd)
	RootCmd.AddCommand(readManyFilesCmd)
	RootCmd.AddCommand(writeCmd)
	RootCmd.AddCommand(globCmd)
	RootCmd.AddCommand(readFileCmd)
	RootCmd.AddCommand(versionCmd)
	RootCmd.AddCommand(execCmd)
	RootCmd.AddCommand(gitBranchCmd)
	RootCmd.AddCommand(codeGuideCmd)
	RootCmd.AddCommand(findDocsCmd)
	RootCmd.AddCommand(cleanupBackToMainCmd)
	RootCmd.AddCommand(prReviewCmd)
	RootCmd.AddCommand(aboutCmd)
	RootCmd.AddCommand(bugCmd)
	RootCmd.AddCommand(clearCmd)
	RootCmd.AddCommand(compressCmd)
	RootCmd.AddCommand(copyCmd)
	RootCmd.AddCommand(corgiCmd)
	RootCmd.AddCommand(directoryCmd)
	RootCmd.AddCommand(docsCmd)
	RootCmd.AddCommand(editorCmd)
	RootCmd.AddCommand(ideCmd)
	RootCmd.AddCommand(initCmd)
	RootCmd.AddCommand(permissionsCmd)
	RootCmd.AddCommand(privacyCmd)
	RootCmd.AddCommand(profileCmd)
	RootCmd.AddCommand(quitCmd)
	RootCmd.AddCommand(restoreCmd)
	RootCmd.AddCommand(setupGithubCmd)
	RootCmd.AddCommand(statsCmd)
	RootCmd.AddCommand(terminalSetupCmd)
	RootCmd.AddCommand(themeCmd)
	RootCmd.AddCommand(vimCmd)
}

func init() {
	RootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {

		projectRoot, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
			os.Exit(1)
		}

		var fileFilteringService *services.FileFilteringService
		WorkspaceService, FSService, ShellService, ExtensionManager, SettingsService, fileFilteringService, ContextService = initServices(projectRoot)
		telemetrySettings := getTelemetrySettings(SettingsService)

		// Retrieve agent-specific settings
		codebaseInvestigatorSettings := SettingsService.GetCodebaseInvestigatorSettings()
		testWriterSettings := SettingsService.GetTestWriterSettings()

		// 1. Initialize Cfg with minimal parameters first
		Cfg = initConfig(nil, types.AgentRegistryInterface(nil), telemetrySettings, codebaseInvestigatorSettings, testWriterSettings, WorkspaceService, fileFilteringService, SettingsService) // Cfg is a *config.Config

		// 2. Create AgentRegistry using the initial Cfg
		agentRegistry := agents.NewAgentRegistry(Cfg) // agentRegistry is *agents.AgentRegistry

		// 3. Set the AgentRegistry into Cfg immediately
		Cfg.AgentRegistry = agentRegistry // Here Cfg.AgentRegistry is set

		// 4. Initialize AgentRegistry (now it's populated with definitions)
		agentRegistry.Initialize()

		// 5. Register all tools (standard tools and wrapped agents)
		// tools.RegisterAllTools will now receive a Cfg with an initialized AgentRegistry
		toolRegistry := registerTools(Cfg, FSService, ShellService, SettingsService, WorkspaceService)

		// 6. Set the fully populated ToolRegistry into Cfg
		Cfg.ToolRegistry = toolRegistry
		// Initialize SessionService now that Cfg is available
		SessionService, err = services.NewSessionService(Cfg.GetGoaiagentDir())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing SessionService: %v\n", err)
			os.Exit(1)
		}

		telemetry.GlobalLogger = telemetry.NewTelemetryLogger(Cfg.Telemetry)
		extensionsCliCommand = commands.NewExtensionsCommand(ExtensionManager, SettingsService)
	}
	RootCmd.Run = func(cmd *cobra.Command, args []string) {
		runChatCmd(RootCmd, cmd, args, SettingsService, ShellService)
	}
	registerCommands()
}
