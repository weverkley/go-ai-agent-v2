package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// themeCmd represents the theme command
var themeCmd = &cobra.Command{
	Use:   "theme",
	Short: "Change the theme",
	Long:  `The theme command allows you to change the visual theme of the Go AI Agent.`, //nolint:staticcheck
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Theme changing is not yet implemented. This feature may be available in a future version.")
	},
}
