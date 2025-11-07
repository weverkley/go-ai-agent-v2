package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// settingsCmd represents the settings command group
var settingsCmd = &cobra.Command{
	Use:   "settings",
	Short: "Manage CLI settings",
	Long:  `The settings command group allows you to view and modify various CLI settings.`, //nolint:staticcheck
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand is provided, print help
		cmd.Help()
	},
}

func init() {
	settingsCmd.AddCommand(settingsGetCmd)
	settingsCmd.AddCommand(settingsSetCmd)
	settingsCmd.AddCommand(settingsListCmd)
	settingsCmd.AddCommand(settingsResetCmd)
}

var settingsGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get the value of a specific setting",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value, found := cfg.Get(key)
		if !found {
			fmt.Printf("Setting '%s' not found.\n", key)
		} else {
			fmt.Printf("%s: %v\n", key, value)
		}
	},
}

var settingsSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set the value of a specific setting",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := args[1]
		// TODO: Implement actual setting of values in config. For now, just print.
		fmt.Printf("Setting '%s' to '%s' (not yet saved persistently).\n", key, value)
	},
}

var settingsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available settings and their current values",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement listing all settings. For now, just print a placeholder.
		fmt.Println("Listing all settings (not yet implemented).")
	},
}

var settingsResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset all settings to their default values",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement resetting settings. For now, just print a placeholder.
		fmt.Println("Resetting all settings to default values (not yet implemented).")
	},
}
