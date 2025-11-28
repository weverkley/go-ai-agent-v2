package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"go-ai-agent-v2/go-cli/pkg/commands" // Add commands import
	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/core"
	"go-ai-agent-v2/go-cli/pkg/core/agents" // Add this line
	"go-ai-agent-v2/go-cli/pkg/prompts"

	"go-ai-agent-v2/go-cli/pkg/extension"

	"go-ai-agent-v2/go-cli/pkg/services"

	"go-ai-agent-v2/go-cli/pkg/telemetry"

	"go-ai-agent-v2/go-cli/pkg/tools"
	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	*services.SessionService,
) {
	workspaceService := services.NewWorkspaceService(projectRoot)
	fsService := services.NewFileSystemService()
	shellService := services.NewShellExecutionService()
	extensionManager := extension.NewManager(projectRoot, fsService, services.NewGitService())
	settingsService := services.NewSettingsService(projectRoot, extensionManager)
	contextService := services.NewContextService(projectRoot)

	fileFilteringService, err := services.NewFileFilteringService(projectRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing FileFilteringService: %v\n", err)
		os.Exit(1)
	}

	var sessionStore services.SessionStore
	sessionStoreType := viper.GetString("sessionStore.type")
	switch sessionStoreType {
	case "redis":
		redisAddr := viper.GetString("sessionStore.redis.address")
		redisPassword := viper.GetString("sessionStore.redis.password")
		redisDB := viper.GetInt("sessionStore.redis.db")
		sessionStore, err = services.NewRedisSessionStore(redisAddr, redisPassword, redisDB)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing Redis session store: %v\n", err)
			os.Exit(1)
		}
	default:
		sessionsPath := filepath.Join(projectRoot, ".goaiagent", "sessions")
		sessionStore, err = services.NewFileSessionStore(sessionsPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing file session store: %v\n", err)
			os.Exit(1)
		}
	}

	sessionService, err := services.NewSessionService(sessionStore)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing SessionService: %v\n", err)
		os.Exit(1)
	}

	return workspaceService, fsService, shellService, extensionManager, settingsService, fileFilteringService, contextService, sessionService
}

func getTelemetrySettings(settingsService types.SettingsServiceIface) *types.TelemetrySettings {
	return settingsService.GetTelemetrySettings()
}

func createChatService(appConfig types.Config, settingsService types.SettingsServiceIface, sessionService *services.SessionService, contextService *services.ContextService) (*services.ChatService, error) {
	executorTypeVal, _ := settingsService.Get("executor")
	executorType, _ := executorTypeVal.(string)

	// Create ExecutorFactory
	executorFactory, err := core.NewExecutorFactory(executorType, appConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating executor factory: %w", err)
	}

	// Create Executor
	toolRegistryVal, _ := appConfig.Get("toolRegistry")
	toolRegistry, _ := toolRegistryVal.(types.ToolRegistryInterface)

	contextContent := contextService.GetContext()
	systemPrompt, err := prompts.GetCoreSystemPrompt(toolRegistry, appConfig, contextContent)
	if err != nil {
		return nil, fmt.Errorf("error creating system prompt: %w", err)
	}
	generationConfig := types.GenerateContentConfig{
		SystemInstruction: systemPrompt,
	}
	executor, err := executorFactory.NewExecutor(appConfig, generationConfig, []*types.Content{})
	if err != nil {
		return nil, fmt.Errorf("error creating executor: %w", err)
	}

	chatService, err := services.NewChatService(executor, toolRegistry, sessionService, settingsService, contextService, appConfig, generationConfig, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating chat service: %w", err)
	}

	return chatService, nil
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
		// This will be handled by the RootCmd.Run
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
		WorkspaceService, FSService, ShellService, ExtensionManager, SettingsService, fileFilteringService, ContextService, SessionService = initServices(projectRoot)
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

		telemetry.GlobalLogger = telemetry.NewTelemetryLogger(Cfg.Telemetry)
		extensionsCliCommand = commands.NewExtensionsCommand(ExtensionManager, SettingsService)

		chatService, err = createChatService(Cfg, SettingsService, SessionService, ContextService)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating chat service: %v\n", err)
			os.Exit(1)
		}
	}
	RootCmd.Run = func(cmd *cobra.Command, args []string) {
		runMode, _ := SettingsService.Get("runMode")
		switch runMode {
		case "agent":
			runAgentCmd(RootCmd, cmd, args)
		default:
			runChatCmd(RootCmd, cmd, args, SettingsService, ShellService)
		}
	}
	registerCommands()
}
