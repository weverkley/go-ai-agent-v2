package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// profileCmd represents the profile command
var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage Gemini CLI user profiles",
	Long:  `The profile command allows you to view and manage user profiles for the Gemini CLI.`, //nolint:staticcheck
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement actual profile management logic.
		fmt.Println("Managing Gemini CLI user profiles (not yet implemented).")
	},
}
