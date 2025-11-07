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

func init() {
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authLogoutCmd)
}

var authLoginCmd = &cobra.Command{
	Use:   "login <API_KEY>",
	Short: "Set the Gemini API key",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := args[0]
		// For now, we'll just print the key. In a real scenario, this would be saved securely.
		// TODO: Implement secure storage of API key.
		fmt.Printf("Gemini API key set (not securely stored yet): %s\n", apiKey)
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
			fmt.Println("Gemini API key is NOT configured. Use 'auth login <API_KEY>' to set it.")
		}
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Clear the configured Gemini API key",
	Run: func(cmd *cobra.Command, args []string) {
		// In a real scenario, this would clear the securely stored API key.
		// For now, we'll just print a message.
		// TODO: Implement clearing of securely stored API key.
		fmt.Println("Gemini API key cleared (not securely cleared yet).")
	},
}
