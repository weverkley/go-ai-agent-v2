package cmd

import (
	"fmt"
	"go-ai-agent-v2/go-cli/pkg/tools"

	"github.com/spf13/cobra"
)

func init() {
	findDocsCmd.Flags().StringVarP(&executorType, "executor", "e", "gemini", "The type of AI executor to use (e.g., 'gemini', 'mock')")
}

var findDocsCmd = &cobra.Command{
	Use:   "find-docs [question]",
	Short: "Find relevant documentation and output GitHub URLs.",
	Long: `Find relevant documentation within the current Git repository and output GitHub URLs.

	This command uses AI to search for documentation files related to your question and provides direct links to them on GitHub.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		toolRegistry := tools.RegisterAllTools(Cfg, FSService, ShellService, SettingsService, WorkspaceService)
		// This is a placeholder. You would implement the actual logic using the services and tools.
		fmt.Println("Tool Registry contains:", toolRegistry.GetAllToolNames())
		fmt.Println("Finding docs for:", args)
	},
}
