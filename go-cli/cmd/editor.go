package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// editorCmd represents the editor command
var editorCmd = &cobra.Command{
	Use:   "editor",
	Short: "Set external editor preference",
	Long:  `The editor command allows you to set your preferred external editor for opening files.`, //nolint:staticcheck
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Setting external editor preference is not yet implemented. This feature may be available in a future version.")
	},
}
