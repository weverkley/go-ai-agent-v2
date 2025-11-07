package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// terminalSetupCmd represents the terminal-setup command
var terminalSetupCmd = &cobra.Command{
	Use:   "terminal-setup",
	Short: "Set up terminal for Gemini CLI",
	Long:  `The terminal-setup command helps configure your terminal for optimal use with the Gemini CLI.`, //nolint:staticcheck
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement actual terminal setup logic.
		fmt.Println("Setting up terminal for Gemini CLI (not yet implemented).")
	},
}
