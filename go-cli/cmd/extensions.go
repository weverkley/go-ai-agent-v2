package cmd

import (
	"go-ai-agent-v2/go-cli/pkg/commands" // Add commands import

	"github.com/spf13/cobra"
)

// extensionsCmd represents the extensions command group
var ExtensionsCmd = &cobra.Command{
	Use:   "extensions",
	Short: "Manage CLI extensions",
	Long:  `The extensions command group allows you to list, enable, and disable CLI extensions.`, //nolint:staticcheck
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand is provided, print help
		cmd.Help()
	},
}


var extensionsCliCommand *commands.ExtensionsCommand

func init() {
}
