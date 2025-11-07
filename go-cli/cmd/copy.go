package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// copyCmd represents the copy command
var copyCmd = &cobra.Command{
	Use:   "copy <source> <destination>",
	Short: "Copy a file or directory",
	Long:  `The copy command copies a file or directory from a source to a destination.`, //nolint:staticcheck
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		source := args[0]
		destination := args[1]
		// TODO: Implement actual file/directory copying.
		fmt.Printf("Copying from '%s' to '%s' (not yet implemented).\n", source, destination)
	},
}
