package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// privacyCmd represents the privacy command
var privacyCmd = &cobra.Command{
	Use:   "privacy",
	Short: "Display the privacy notice",
	Long:  `The privacy command displays the privacy notice for the Gemini CLI.`, //nolint:staticcheck
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement actual logic to display the privacy notice.
		fmt.Println("Displaying privacy notice (not yet implemented).")
	},
}
