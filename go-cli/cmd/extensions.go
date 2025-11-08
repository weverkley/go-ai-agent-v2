package cmd

import (
	"fmt"
	"os"

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
		extensions := ExtensionManager.ListExtensions()
		if len(extensions) == 0 {
			fmt.Println("No extensions found.")
			return
		}
		fmt.Println("Available extensions:")
		for _, ext := range extensions {
			status := "Disabled"
			if ext.Enabled {
				status = "Enabled"
			}
			fmt.Printf("- %s: %s (%s)\n", ext.Name, ext.Description, status)
		}
	},
}

var extensionsEnableCmd = &cobra.Command{
	Use:   "enable <extension_name>",
	Short: "Enable a specific extension",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		extensionName := args[0]
		err := ExtensionManager.EnableExtension(extensionName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error enabling extension '%s': %v\n", extensionName, err)
			os.Exit(1)
		}
		fmt.Printf("Extension '%s' enabled successfully.\n", extensionName)
	},
}

var extensionsDisableCmd = &cobra.Command{
	Use:   "disable <extension_name>",
	Short: "Disable a specific extension",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		extensionName := args[0]
		err := ExtensionManager.DisableExtension(extensionName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error disabling extension '%s': %v\n", extensionName, err)
			os.Exit(1)
		}
		fmt.Printf("Extension '%s' disabled successfully.\n", extensionName)
	},
}