package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// extensionsCmd represents the extensions command group
var extensionsCmd = &cobra.Command{
	Use:   "extensions",
	Short: "Manage CLI extensions",
	Long:  `The extensions command group allows you to list, enable, and disable CLI extensions.`, //nolint:staticcheck
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand is provided, print help
		cmd.Help()
	},
}

func init() {
	extensionsCmd.AddCommand(extensionsListCmd)
	extensionsCmd.AddCommand(extensionsEnableCmd)
	extensionsCmd.AddCommand(extensionsDisableCmd)
}

var extensionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available extensions",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement actual listing of extensions.
		fmt.Println("Listing all extensions (not yet implemented).")
	},
}

var extensionsEnableCmd = &cobra.Command{
	Use:   "enable <extension_name>",
	Short: "Enable a specific extension",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		extensionName := args[0]
		// TODO: Implement actual enabling of extensions.
		fmt.Printf("Enabling extension: '%s' (not yet implemented).\n", extensionName)
	},
}

var extensionsDisableCmd = &cobra.Command{
	Use:   "disable <extension_name>",
	Short: "Disable a specific extension",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		extensionName := args[0]
		// TODO: Implement actual disabling of extensions.
		fmt.Printf("Disabling extension: '%s' (not yet implemented).\n", extensionName)
	},
}