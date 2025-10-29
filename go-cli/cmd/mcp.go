package cmd

import (
	"fmt"
	"os"

	"go-ai-agent-v2/go-cli/pkg/commands"
	"github.com/spf13/cobra"
)

var (
	mcpAddName string
	mcpAddUrl  string
	mcpRemoveName string
)

func init() {
	rootCmd.AddCommand(mcpCmd)

	mcpCmd.AddCommand(mcpListCmd)

	mcpCmd.AddCommand(mcpAddCmd)
	mcpAddCmd.Flags().StringVar(&mcpAddName, "name", "", "The name of the MCP server to add.")
	mcpAddCmd.Flags().StringVar(&mcpAddUrl, "url", "", "The URL of the MCP server to add.")
	mcpAddCmd.MarkFlagRequired("name")
	mcpAddCmd.MarkFlagRequired("url")

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
	Use:   "add",
	Short: "Add an MCP server",
	Run: func(cmd *cobra.Command, args []string) {
		mcp := commands.NewMcpCommand()
		err := mcp.AddMcpItem(mcpAddName, mcpAddUrl)
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
