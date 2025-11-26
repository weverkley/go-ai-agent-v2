package cmd

import (
	"fmt"
	"go-ai-agent-v2/go-cli/pkg/tools"

	"github.com/spf13/cobra"
)

var codeGuideCmd = &cobra.Command{
	Use:   "code-guide [file_or_directory]",
	Short: "Provide a high-level guide to a codebase.",
	Long: `This command uses AI to analyze a specified file or directory and generate a high-level guide.
The guide includes summaries of functions, classes, and other code structures to help you quickly understand the codebase.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		toolRegistry := tools.RegisterAllTools(Cfg, FSService, ShellService, SettingsService, WorkspaceService)

		if len(args) > 0 {
			fmt.Printf("Code guide for: %s\n", args[0])
		} else {
			fmt.Println("Please provide a file or directory for the code guide.")
		}
		fmt.Println("Tool Registry contains:", toolRegistry.GetAllToolNames())
	},
}

func init() {
	codeGuideCmd.Flags().StringVarP(&executorType, "executor", "e", "gemini", "The type of AI executor to use (e.g., 'gemini', 'mock')")
}
