package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// profileCmd represents the profile command
var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Toggle the debug profile display",
	Long:  `The profile command toggles the debug profile display (only available in development mode). This functionality is not yet implemented. This feature may be available in a future version.`, //nolint:staticcheck
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Toggling debug profile display is not yet implemented. This feature may be available in a future version.")
	},
}
