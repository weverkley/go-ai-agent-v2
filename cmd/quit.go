package cmd

import (
	"github.com/spf13/cobra"
)

// quitCmd represents the quit command
var quitCmd = &cobra.Command{
	Use:   "quit",
	Short: "Exit the Go AI Agent",
	Long:  `The quit command exits the Go AI Agent application.`, //nolint:staticcheck
	Run: func(cmd *cobra.Command, args []string) {
		// This command is handled by the UI, which will send a tea.Quit message.
		// The Run function is left empty to prevent os.Exit(0) from being called here.
	},
}
