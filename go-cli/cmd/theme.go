package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// themeCmd represents the theme command
var themeCmd = &cobra.Command{
	Use:   "theme",
	Short: "Change the theme",
	Long:  `The theme command allows you to change the visual theme of the Gemini CLI.`, //nolint:staticcheck
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement actual logic to change the theme.
		fmt.Println("Changing the theme (not yet implemented).")
	},
}
