package cmd

import (
	"fmt"
	"os"
	"time" // Import time package

	"go-ai-agent-v2/go-cli/pkg/commands" // Add commands import
	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/core"

	"go-ai-agent-v2/go-cli/pkg/extension"

	"go-ai-agent-v2/go-cli/pkg/services"

	"go-ai-agent-v2/go-cli/pkg/telemetry"

	"go-ai-agent-v2/go-cli/pkg/tools"

	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "go-cli",
	Short: "A Go-based CLI for Gemini",
	Long:  `A Go-based CLI for interacting with the Gemini API and managing extensions.`,
}

var Cfg *config.Config
var executorType string
var executor core.Executor                      // Declare package-level executor
var chatService *services.ChatService           // Declare package-level chatService
var WorkspaceService *services.WorkspaceService // Declare package-level workspaceService
var ExtensionManager *extension.Manager         // Declare package-level extensionManager
var MemoryService *services.MemoryService       // Declare package-level memoryService
var SessionStartTime time.Time                  // Declare sessionStartTime
var SettingsService *services.SettingsService   // Declare package-level settingsService

var FSService services.FileSystemService // Declare package-level FileSystemService
var ShellService services.ShellExecutionService   // Declare package-level ShellExecutionService
var extensionsCliCommand *commands.ExtensionsCommand // Declare package-level extensionsCliCommand

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initSessionStartTime() {
	SessionStartTime = time.Now()
}

func initServices(projectRoot string) (
	*services.WorkspaceService,
	services.FileSystemService,
	services.ShellExecutionService, // Changed to interface
	*extension.Manager,
	*services.SettingsService,
	*services.FileFilteringService,
) {
	workspaceService := services.NewWorkspaceService(projectRoot)
	fsService := services.NewFileSystemService()
	shellService := services.NewShellExecutionService()
	extensionManager := extension.NewManager(projectRoot, fsService, services.NewGitService())
	settingsService := services.NewSettingsService(projectRoot)

	fileFilteringService, err := services.NewFileFilteringService(projectRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing FileFilteringService: %v\n", err)
		os.Exit(1)
	}

	return workspaceService, fsService, shellService, extensionManager, settingsService, fileFilteringService
}

func getTelemetrySettings(settingsService *services.SettingsService) *types.TelemetrySettings {
	return settingsService.GetTelemetrySettings()
}

func registerTools(fsService services.FileSystemService, shellService services.ShellExecutionService, settingsService *services.SettingsService) *types.ToolRegistry {
	return tools.RegisterAllTools(fsService, shellService, settingsService)
}

func initConfig(
	toolRegistry *types.ToolRegistry,
	telemetrySettings *types.TelemetrySettings,
	workspaceService *services.WorkspaceService,
	fileFilteringService *services.FileFilteringService,
	settingsService *services.SettingsService, // Add settingsService
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
	grepCodeCmd.Run = func(cmd *cobra.Command, args []string) {
		runGrepCodeCmd(cmd, args, SettingsService, ShellService)
	}
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
		initSessionStartTime()

		projectRoot, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
			os.Exit(1)
		}

		var fileFilteringService *services.FileFilteringService
		WorkspaceService, FSService, ShellService, ExtensionManager, SettingsService, fileFilteringService = initServices(projectRoot)
		telemetrySettings := getTelemetrySettings(SettingsService)
		toolRegistry := registerTools(FSService, ShellService, SettingsService)
		Cfg = initConfig(toolRegistry, telemetrySettings, WorkspaceService, fileFilteringService, SettingsService)

		executorType = "gemini"
		telemetry.GlobalLogger = telemetry.NewTelemetryLogger(Cfg.Telemetry)
		extensionsCliCommand = commands.NewExtensionsCommand(ExtensionManager, SettingsService)
	}
	RootCmd.Run = func(cmd *cobra.Command, args []string) {
		runChatCmd(RootCmd, cmd, args, SettingsService, ShellService)
	}
	registerCommands()
}
