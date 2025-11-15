package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"go-ai-agent-v2/go-cli/pkg/utils" // Import utils package
)

// quitCmd represents the quit command
var quitCmd = &cobra.Command{
	Use:   "quit",
	Short: "Exit the Go AI Agent",
	Long:  `The quit command exits the Go AI Agent application.`, //nolint:staticcheck
	Run: func(cmd *cobra.Command, args []string) {
		duration := time.Since(SessionStartTime)
		fmt.Printf("Exiting Go AI Agent. Session duration: %s. Goodbye!\n", utils.FormatDuration(duration))
		os.Exit(0)
	},
}
