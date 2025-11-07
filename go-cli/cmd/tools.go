package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// toolsCmd represents the tools command group
var toolsCmd = &cobra.Command{
	Use:   "tools",
	Short: "Manage AI tools",
	Long:  `The tools command group allows you to list and run AI tools.`, //nolint:staticcheck
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand is provided, print help
		cmd.Help()
	},
}

func init() {
	toolsCmd.AddCommand(toolsListCmd)
	toolsCmd.AddCommand(toolsRunCmd)
}

var toolsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available AI tools",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement actual listing of tools.
		fmt.Println("Listing all AI tools (not yet implemented).")
	},
}

var toolsRunCmd = &cobra.Command{
	Use:   "run <tool_name> [args]",
	Short: "Run a specific AI tool",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		toolName := args[0]
		toolArgs := args[1:]
		fmt.Printf("Running tool '%s' with args %v (not yet implemented).\n", toolName, toolArgs)
	},
}
