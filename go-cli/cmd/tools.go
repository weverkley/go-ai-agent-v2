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
		toolRegistry := Cfg.GetToolRegistry()
		if toolRegistry == nil {
			fmt.Println("Tool registry not initialized.")
			return
		}

		tools := toolRegistry.GetAllTools()
		if len(tools) == 0 {
			fmt.Println("No AI tools available.")
			return
		}

		fmt.Println("Available AI Tools:")
		for _, tool := range tools {
			// Filter out MCP tools (assuming MCP tools have a ServerName)
			if tool.ServerName() == "" {
				fmt.Printf("- %s: %s\n", tool.Name(), tool.Description())
			}
		}
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
