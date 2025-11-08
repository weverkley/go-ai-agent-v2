package cmd

import (
	"fmt"
	"os"

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
		value, found := SettingsService.Get(key)
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
		err := SettingsService.Set(key, value)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error setting '%s': %v\n", key, err)
			os.Exit(1)
		}
		err = SettingsService.Save()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error saving settings: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Setting '%s' updated to '%s'.\n", key, value)
	},
}

var settingsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available settings and their current values",
	Run: func(cmd *cobra.Command, args []string) {
		settings := SettingsService.AllSettings()
		if len(settings) == 0 {
			fmt.Println("No settings found.")
			return
		}
		fmt.Println("Current settings:")
		for key, value := range settings {
			fmt.Printf("- %s: %v\n", key, value)
		}
	},
}

var settingsResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset all settings to their default values",
	Run: func(cmd *cobra.Command, args []string) {
		err := SettingsService.Reset()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resetting settings: %v\n", err)
			os.Exit(1)
		}
		err = SettingsService.Save()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error saving settings: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("All settings reset to default values.")
	},
}
