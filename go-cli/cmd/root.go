package cmd

import (
	    "fmt"
	    "os"
	    "time" // Import time package
	
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

var rootCmd = &cobra.Command{
	Use:   "go-cli",
	Short: "A Go-based CLI for Gemini",
	Long:  `A Go-based CLI for interacting with the Gemini API and managing extensions.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// This will run before any subcommand. We can use it to set up common configurations.
		// Initialize the executor here so it's available to all subcommands
		executorFactory := core.NewExecutorFactory()
		var err error
		executor, err = executorFactory.CreateExecutor(executorType, Cfg, types.GenerateContentConfig{}, []*genai.Content{})
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

func Execute() {
	if err := rootCmd.Execute(); err != nil {
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

	// Initialize extensionManager here
	ExtensionManager = extension.NewManager(projectRoot)

	// Initialize memoryService here
	MemoryService = services.NewMemoryService(projectRoot)

	// Initialize settingsService here
	SettingsService = services.NewSettingsService(projectRoot)

	rootCmd.PersistentFlags().StringVarP(&executorType, "executor", "e", "gemini", "The type of AI executor to use (e.g., 'gemini', 'mock')")
	// Register all tools
	toolRegistry := tools.RegisterAllTools()

	// Initialize ConfigParameters
	params := &config.ConfigParameters{
		// Set default values or load from settings file
		DebugMode: false,
		Model:     config.DEFAULT_GEMINI_MODEL,
		Telemetry: &types.TelemetrySettings{ // Initialize TelemetrySettings
			Enabled: false, // Default to disabled
			Outfile: "",    // Default to no outfile
		},
		// Add other parameters as needed
		ToolRegistry: toolRegistry, // Pass the populated toolRegistry
	}

	// Initialize FileFilteringService
	fileFilteringService, err := services.NewFileFilteringService(projectRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing FileFilteringService: %v\n", err)
		os.Exit(1)
	}
	Cfg.SetConfiguredFileService(fileFilteringService)

	// Create the final Config instance
	Cfg = config.NewConfig(params)

	// Initialize the global telemetry logger
	telemetry.GlobalLogger = telemetry.NewTelemetryLogger(params.Telemetry)

	rootCmd.AddCommand(todosCmd)
	rootCmd.AddCommand(chatCmd)
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(modelCmd)
	rootCmd.AddCommand(settingsCmd)
	rootCmd.AddCommand(memoryCmd)
	rootCmd.AddCommand(extensionsCmd)
	rootCmd.AddCommand(mcpCmd)
	rootCmd.AddCommand(toolsCmd)
	rootCmd.AddCommand(grepCmd)
	rootCmd.AddCommand(lsCmd)
	rootCmd.AddCommand(webSearchCmd)
	rootCmd.AddCommand(webFetchCmd)
	rootCmd.AddCommand(readCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(smartEditCmd)
	rootCmd.AddCommand(grepCodeCmd)
	rootCmd.AddCommand(readManyFilesCmd)
	rootCmd.AddCommand(writeCmd)
	rootCmd.AddCommand(globCmd)
	rootCmd.AddCommand(readFileCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(execCmd)
	rootCmd.AddCommand(gitBranchCmd)
	rootCmd.AddCommand(codeGuideCmd)
	rootCmd.AddCommand(findDocsCmd)
	rootCmd.AddCommand(cleanupBackToMainCmd)
	rootCmd.AddCommand(prReviewCmd)
	rootCmd.AddCommand(aboutCmd)
	rootCmd.AddCommand(bugCmd)
	rootCmd.AddCommand(clearCmd)
	rootCmd.AddCommand(compressCmd)
	rootCmd.AddCommand(copyCmd)
	rootCmd.AddCommand(corgiCmd)
	rootCmd.AddCommand(directoryCmd)
	rootCmd.AddCommand(docsCmd)
	rootCmd.AddCommand(editorCmd)
	rootCmd.AddCommand(ideCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(permissionsCmd)
	rootCmd.AddCommand(privacyCmd)
	rootCmd.AddCommand(profileCmd)
	rootCmd.AddCommand(quitCmd)
	rootCmd.AddCommand(restoreCmd)
	rootCmd.AddCommand(setupGithubCmd)
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(terminalSetupCmd)
	rootCmd.AddCommand(themeCmd)
	rootCmd.AddCommand(vimCmd)
}

