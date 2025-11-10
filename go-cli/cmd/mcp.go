package cmd

import (
	"fmt"
	"os" // Add os import
	"go-ai-agent-v2/go-cli/pkg/mcp"
	"go-ai-agent-v2/go-cli/pkg/types" // Add types import

	"github.com/spf13/cobra"
)

// mcpCmd represents the mcp command group
var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Manage Model Context Protocol (MCP) servers",
	Long:  `The mcp command group allows you to list, add, and remove Model Context Protocol (MCP) servers.`, //nolint:staticcheck
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand is provided, print help
		cmd.Help()
	},
}

func init() {
	mcpCmd.AddCommand(mcpListCmd)
	mcpCmd.AddCommand(mcpAddCmd)
	mcpCmd.AddCommand(mcpRemoveCmd)
}

var mcpListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured MCP servers",
	Run: func(cmd *cobra.Command, args []string) {
		toolRegistryVal, found := Cfg.Get("toolRegistry")
		if !found || toolRegistryVal == nil {
			fmt.Fprintf(os.Stderr, "Error: Tool registry not found in config.\n")
			return
		}
		toolRegistry, ok := toolRegistryVal.(types.ToolRegistryInterface)
		if !ok {
			fmt.Fprintf(os.Stderr, "Error: Tool registry in config is not of expected type.\n")
			return
		}
		manager := mcp.NewMcpClientManager(toolRegistry)
		servers := manager.ListServers(Cfg)

		if len(servers) == 0 {
			fmt.Println("No MCP servers configured.")
			return
		}

		fmt.Println("Configured MCP Servers:")
		for _, server := range servers {
			fmt.Printf("  - Name: %s, Status: %s, URL: %s, Description: %s\n", server.Name, server.Status, server.Url, server.Description)
		}
	},
}

var mcpAddCmd = &cobra.Command{
	Use:   "add <name> <url>",
	Short: "Add a new MCP server",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		url := args[1]
		fmt.Printf("Adding MCP server '%s' with URL '%s' is not yet implemented. This feature may be available in a future version.\n", name, url)
	},
}

var mcpRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove an MCP server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		fmt.Printf("Removing MCP server '%s' is not yet implemented. This feature may be available in a future version.\n", name)
	},
}
