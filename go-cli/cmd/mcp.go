package cmd

import (
	"fmt"
	"os"

	"go-ai-agent-v2/go-cli/pkg/commands"
	"go-ai-agent-v2/go-cli/pkg/config"
	"github.com/spf13/cobra"
)

var (
	mcpAddName string
	mcpAddCommandOrUrl string
	mcpAddArgs []string
	mcpAddScope string
	mcpAddTransport string
	mcpAddEnv []string
	mcpAddHeader []string
	mcpAddTimeout int
	mcpAddTrust bool
	mcpAddDescription string
	mcpAddIncludeTools []string
	mcpAddExcludeTools []string
	mcpRemoveName string
)

func init() {
	rootCmd.AddCommand(mcpCmd)

	mcpCmd.AddCommand(mcpListCmd)

	mcpCmd.AddCommand(mcpAddCmd)
	mcpAddCmd.Flags().StringArrayVar(&mcpAddArgs, "args", []string{}, "Arguments for the command (stdio transport).")
	mcpAddCmd.Flags().StringVar(&mcpAddScope, "scope", "project", "Configuration scope (user or project)")
	mcpAddCmd.Flags().StringVar(&mcpAddTransport, "transport", "stdio", "Transport type (stdio, sse, http)")
	mcpAddCmd.Flags().StringArrayVar(&mcpAddEnv, "env", []string{}, "Set environment variables (e.g. -e KEY=value)")
	mcpAddCmd.Flags().StringArrayVar(&mcpAddHeader, "header", []string{}, "Set HTTP headers for SSE and HTTP transports (e.g. -H \"X-Api-Key: abc123\")")
	mcpAddCmd.Flags().IntVar(&mcpAddTimeout, "timeout", 0, "Set connection timeout in milliseconds")
	mcpAddCmd.Flags().BoolVar(&mcpAddTrust, "trust", false, "Trust the server (bypass all tool call confirmation prompts)")
	mcpAddCmd.Flags().StringVar(&mcpAddDescription, "description", "", "Set the description for the server")
	mcpAddCmd.Flags().StringArrayVar(&mcpAddIncludeTools, "include-tools", []string{}, "A comma-separated list of tools to include")
	mcpAddCmd.Flags().StringArrayVar(&mcpAddExcludeTools, "exclude-tools", []string{}, "A comma-separated list of tools to exclude")

	mcpCmd.AddCommand(mcpRemoveCmd)
	mcpRemoveCmd.Flags().StringVar(&mcpRemoveName, "name", "", "The name of the MCP server to remove.")
	mcpRemoveCmd.MarkFlagRequired("name")
}

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Manage MCP servers",
	Long:  `Manage MCP servers for the Go Gemini CLI.`,
}

var mcpListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured MCP servers",
	Run: func(cmd *cobra.Command, args []string) {
		mcp := commands.NewMcpCommand()
		err := mcp.ListMcpItems()
		if err != nil {
			fmt.Printf("Error listing MCP items: %v\n", err)
			os.Exit(1)
		}
	},
}

var mcpAddCmd = &cobra.Command{
	Use:   "add [name] [commandOrUrl] [args...]",
	Short: "Add an MCP server",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		mcpAddName = args[0]
		mcpAddCommandOrUrl = args[1]
		if len(args) > 2 {
			mcpAddArgs = args[2:]
		}
		mcp := commands.NewMcpCommand()
		err := mcp.AddMcpItem(
			mcpAddName,
			mcpAddCommandOrUrl,
			mcpAddArgs,
			config.SettingScope(mcpAddScope),
			mcpAddTransport,
			mcpAddEnv,
			mcpAddHeader,
			mcpAddTimeout,
			mcpAddTrust,
			mcpAddDescription,
			mcpAddIncludeTools,
			mcpAddExcludeTools,
		)
		if err != nil {
			fmt.Printf("Error adding MCP item: %v\n", err)
			os.Exit(1)
		}
	},
}

var mcpRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove an MCP server",
	Run: func(cmd *cobra.Command, args []string) {
		mcp := commands.NewMcpCommand()
		err := mcp.RemoveMcpItem(mcpRemoveName)
		if err != nil {
			fmt.Printf("Error removing MCP item: %v\n", err)
			os.Exit(1)
		}
	},
}
