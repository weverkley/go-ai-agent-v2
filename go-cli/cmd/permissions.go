package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// permissionsCmd represents the permissions command
var permissionsCmd = &cobra.Command{
	Use:   "permissions",
	Short: "Manage Gemini CLI permissions",
	Long:  `The permissions command allows you to manage folder trust settings for the Gemini CLI.`, //nolint:staticcheck
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement actual folder trust management logic.
		fmt.Println("Managing folder trust settings (not yet implemented).")
	},
}
