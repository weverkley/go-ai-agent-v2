package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// bugCmd represents the bug command
var bugCmd = &cobra.Command{
	Use:   "bug",
	Short: "Report a bug or provide feedback",
	Long:  `The bug command opens a new issue in the Gemini CLI GitHub repository, allowing users to report bugs or provide feedback.`, //nolint:staticcheck
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Opening a new issue in the Gemini CLI GitHub repository (not yet implemented).")
	},
}
