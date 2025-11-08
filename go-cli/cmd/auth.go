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
		fmt.Printf("Gemini API key received. For now, please set it as an environment variable: export GEMINI_API_KEY=%s\n", apiKey)
		fmt.Println("Secure storage of API keys is not yet implemented. This feature, crucial for security, may be available in a future version.")
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
		fmt.Println("Clearing the Gemini API key is not yet implemented. If you set it as an environment variable, you will need to unset it manually.")
		fmt.Println("Secure clearing of API keys, crucial for security, may be available in a future version.")
	},
}
