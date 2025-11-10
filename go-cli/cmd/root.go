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
	
	    
	
	    	"github.com/google/generative-ai-go/genai"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "go-cli",
	Short: "A Go-based CLI for Gemini",
	Long:  `A Go-based CLI for interacting with the Gemini API and managing extensions.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// This will run before any subcommand. We can use it to set up common configurations.
		// Initialize the executor here so it's available to all subcommands
		factory, err := core.NewExecutorFactory(executorType)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating executor factory: %v\n", err)
			os.Exit(1)
		}
		executor, err = factory.NewExecutor(Cfg, types.GenerateContentConfig{}, []*genai.Content{})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating executor: %v\n", err)
			os.Exit(1)
		}

		// Initialize chatService here, after executor is available
		chatService = services.NewChatService(Cfg, executor)
	},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			os.Exit(0)
		}
	},
}

var Cfg *config.Config
var executorType string
var executor core.Executor // Declare package-level executor
var chatService *services.ChatService // Declare package-level chatService
var WorkspaceService *services.WorkspaceService // Declare package-level workspaceService
var ExtensionManager *extension.Manager // Declare package-level extensionManager
var MemoryService *services.MemoryService // Declare package-level memoryService
var SessionStartTime time.Time // Declare sessionStartTime
var SettingsService *services.SettingsService // Declare package-level settingsService

var FSService services.FileSystemService // Declare package-level FileSystemService

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Initialize sessionStartTime
	SessionStartTime = time.Now()

	// Initialize workspaceService here
	projectRoot, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
		os.Exit(1)
	}
	WorkspaceService = services.NewWorkspaceService(projectRoot)

	// Initialize FileSystemService
	FSService = services.NewFileSystemService()

	// Initialize extensionManager here
	ExtensionManager = extension.NewManager(projectRoot, FSService, services.NewGitService())
	// Initialize settingsService here
	SettingsService = services.NewSettingsService(projectRoot)

	// Initialize extensionManager here
	ExtensionManager = extension.NewManager(projectRoot, FSService, services.NewGitService())
	// Initialize settingsService here
	SettingsService = services.NewSettingsService(projectRoot)

	// Initialize extensionsCliCommand here
	extensionsCliCommand = commands.NewExtensionsCommand(ExtensionManager, SettingsService)
	// Register all tools
	toolRegistry := tools.RegisterAllTools(FSService)

	// Initialize ConfigParameters
	params := &config.ConfigParameters{
		// Set default values or load from settings file
		DebugMode: false,
		ModelName: config.DEFAULT_GEMINI_MODEL,
		Telemetry: &types.TelemetrySettings{ // Initialize TelemetrySettings
			Enabled: false, // Default to disabled
			Outfile: "",    // Default to no outfile
		},
		// Add other parameters as needed
		ToolRegistry: toolRegistry, // Pass the populated toolRegistry
	}

	// Create the final Config instance
	Cfg = config.NewConfig(params)
	Cfg.WorkspaceContext = WorkspaceService // Set the workspace context

	// Initialize FileFilteringService
	fileFilteringService, err := services.NewFileFilteringService(projectRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing FileFilteringService: %v\n", err)
		os.Exit(1)
	}
	Cfg.FileFilteringService = fileFilteringService // Set the file filtering service directly

	// Initialize the global telemetry logger
	telemetry.GlobalLogger = telemetry.NewTelemetryLogger(params.Telemetry)

	RootCmd.AddCommand(todosCmd)
	RootCmd.AddCommand(chatCmd)
	RootCmd.AddCommand(authCmd)
	RootCmd.AddCommand(modelCmd)
	RootCmd.AddCommand(settingsCmd)
	RootCmd.AddCommand(memoryCmd)
	RootCmd.AddCommand(ExtensionsCmd)
	ExtensionsCmd.AddCommand(installCmd)
	ExtensionsCmd.AddCommand(extensionsListCmd)
	ExtensionsCmd.AddCommand(extensionsEnableCmd)
	ExtensionsCmd.AddCommand(extensionsDisableCmd)
	ExtensionsCmd.AddCommand(newCmd)
	ExtensionsCmd.AddCommand(updateCmd)
	ExtensionsCmd.AddCommand(linkCmd)

	// Add flags for installCmd
	installCmd.Flags().String("ref", "", "Specify a ref (branch, tag, or commit) for git installations.")
	installCmd.Flags().Bool("auto-update", false, "Enable automatic updates for the extension.")
	installCmd.Flags().Bool("allow-prerelease", false, "Allow installation of pre-release versions.")
	installCmd.Flags().Bool("force", false, "Force installation, overwriting existing extensions.")
	installCmd.Flags().Bool("consent", false, "Provide consent for installation (e.g., for security warnings).")

	// Add flags for newCmd
	newCmd.Flags().String("template", "", "Specify a template to create the new extension from.")

	// Add flags for updateCmd
	updateCmd.Flags().Bool("all", false, "Update all installed extensions.")
	RootCmd.AddCommand(mcpCmd)
	RootCmd.AddCommand(toolsCmd)
	RootCmd.AddCommand(grepCmd)
	RootCmd.AddCommand(lsCmd)
	RootCmd.AddCommand(webSearchCmd)
	RootCmd.AddCommand(webFetchCmd)
	RootCmd.AddCommand(readCmd)
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

