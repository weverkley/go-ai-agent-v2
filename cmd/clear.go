package cmd

import (
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// clearCmd represents the clear command
var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear the screen and conversation history",
	Long:  `The clear command clears the terminal screen and resets the conversation history.`, //nolint:staticcheck
	Run: func(cmd *cobra.Command, args []string) {

		clearScreen() // Keep screen clearing
	},
}

func init() {
	// No rootCmd.AddCommand here, it's added in root.go
}

func clearScreen() {
	cmd := exec.Command("clear") // Linux/macOS
	cmd.Stdout = os.Stdout
	_ = cmd.Run()

	cmd = exec.Command("cmd", "/c", "cls") // Windows
	cmd.Stdout = os.Stdout
	_ = cmd.Run()
}
