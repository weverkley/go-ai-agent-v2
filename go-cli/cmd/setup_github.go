package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// setupGithubCmd represents the setup-github command
var setupGithubCmd = &cobra.Command{
	Use:   "setup-github",
	Short: "Set up GitHub integration",
	Long:  `The setup-github command guides you through setting up GitHub integration for the Gemini CLI.`, //nolint:staticcheck
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement actual GitHub setup logic.
		fmt.Println("Setting up GitHub integration (not yet implemented).")
	},
}
