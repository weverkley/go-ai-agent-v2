package cmd

import (
	"fmt"
	"os"

	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/core"
	"go-ai-agent-v2/go-cli/pkg/telemetry"
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
	},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			os.Exit(0)
		}
	},
}

var cfg *config.Config
var executorType string
var executor core.Executor // Declare package-level executor

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&executorType, "executor", "e", "gemini", "The type of AI executor to use (e.g., 'gemini', 'mock')")
	// Create a dummy config for initial tool registry creation
	toolRegistry := types.NewToolRegistry()

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
		ToolRegistry: toolRegistry, // Pass the toolRegistry directly
	}

	// Create the final Config instance
	cfg = config.NewConfig(params)

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

	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		// Initialize the executor here so it's available to all subcommands
		executorFactory := core.NewExecutorFactory()
		var err error
		executor, err = executorFactory.CreateExecutor(executorType, cfg, types.GenerateContentConfig{}, []*genai.Content{})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating executor: %v\n", err)
			os.Exit(1)
		}
	}
}

