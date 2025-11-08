package cmd

import (
	"fmt"
	"strings" // Import strings package

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

		toolRegistry := Cfg.GetToolRegistry()
		if toolRegistry == nil {
			fmt.Println("Tool registry not initialized.")
			return
		}

		tool, err := toolRegistry.GetTool(toolName)
		if err != nil {
			fmt.Printf("Error: Tool '%s' not found: %v\n", toolName, err)
			return
		}

		parsedArgs := make(map[string]any)
		for _, arg := range toolArgs {
			parts := strings.SplitN(arg, "=", 2)
			if len(parts) == 2 {
				parsedArgs[parts[0]] = parts[1]
			} else {
				fmt.Printf("Warning: Invalid argument format '%s'. Expected key=value.\n", arg)
			}
		}

		result, err := tool.Execute(parsedArgs)
		if err != nil {
			fmt.Printf("Error executing tool '%s': %v\n", toolName, err)
			return
		}

		if result.Error != nil {
			fmt.Printf("Tool '%s' returned an error: %s\n", toolName, result.Error.Message)
		} else {
			fmt.Printf("Tool '%s' executed successfully.\n", toolName)
			if result.ReturnDisplay != "" {
				fmt.Println("Output:")
				fmt.Println(result.ReturnDisplay)
			}
		}
	},
}
