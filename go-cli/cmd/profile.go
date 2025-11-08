package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// profileCmd represents the profile command
var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Toggle the debug profile display",
	Long:  `The profile command toggles the debug profile display (only available in development mode). This functionality is not yet implemented.`, //nolint:staticcheck
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement actual logic to toggle the debug profile display.
		fmt.Println("Toggling debug profile display (only available in development mode, not yet implemented).")
	},
}
