package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// quitCmd represents the quit command
var quitCmd = &cobra.Command{
	Use:   "quit",
	Short: "Exit the Gemini CLI",
	Long:  `The quit command exits the Gemini CLI application.`, //nolint:staticcheck
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Exiting Gemini CLI. Goodbye!")
		os.Exit(0)
	},
}
