package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// aboutCmd represents the about command
var aboutCmd = &cobra.Command{
	Use:   "about",
	Short: "Display information about the Gemini CLI",
	Long:  `The about command displays information about the Gemini CLI, including its purpose and version.`, //nolint:staticcheck
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Gemini CLI - A Go-based CLI for interacting with the Gemini API and managing extensions.")
		fmt.Println("Version: 0.1.0")
		fmt.Println("Developed by Google")
	},
}
