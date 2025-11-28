package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication credentials",
	Long:  `The auth command group allows you to manage your authentication credentials, such as setting and clearing API keys.`, //nolint:staticcheck
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand is provided, print help
		cmd.Help()
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check if a Gemini API key is configured",
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey != "" {
			fmt.Println("Gemini API key is configured.")
		} else {
			fmt.Println("Gemini API key is NOT configured. Please set it as an environment variable (e.g., export GEMINI_API_KEY=your_key_here).")
		}
	},
}

func init() {
	authCmd.AddCommand(authStatusCmd)
}
