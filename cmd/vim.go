package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// vimCmd represents the vim command
var vimCmd = &cobra.Command{
	Use:   "vim",
	Short: "Toggle vim mode on/off",
	Long:  `The vim command toggles Vim mode in the CLI for enhanced text editing. This functionality is not yet implemented. This feature may be available in a future version.`, //nolint:staticcheck
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Toggling Vim mode is not yet implemented. This feature may be available in a future version.")
	},
}
