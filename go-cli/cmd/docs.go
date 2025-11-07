package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// docsCmd represents the docs command
var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Open the Gemini CLI documentation",
	Long:  `The docs command opens the official documentation for the Gemini CLI in your web browser.`, //nolint:staticcheck
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement opening documentation in a web browser.
		fmt.Println("Opening Gemini CLI documentation (not yet implemented).")
	},
}
