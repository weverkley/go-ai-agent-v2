package cmd

import (
	"fmt"
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
		if err := executor.SetHistory(nil); err != nil {
			fmt.Fprintf(os.Stderr, "Error clearing conversation history: %v\n", err)
			os.Exit(1)
		}
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
