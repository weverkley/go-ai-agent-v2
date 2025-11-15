package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	version   = "dev"
	buildDate = "unknown"
	gitCommit = "unknown"
)

// aboutCmd represents the about command
var aboutCmd = &cobra.Command{
	Use:   "about",
	Short: "Display information about the Go AI Agent",
	Long:  `The about command displays information about the Go AI Agent, including its purpose, version, build details, and environment.`, //nolint:staticcheck
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Go AI Agent - A Go-based CLI for interacting with the Gemini API and managing extensions.")
		fmt.Printf("Version: %s\n", version)
		fmt.Printf("Build Date: %s\n", buildDate)
		fmt.Printf("Git Commit: %s\n", gitCommit)
		fmt.Printf("Go Version: %s\n", runtime.Version())
		fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		fmt.Println("Developed by Google")
	},
}

func init() {
	// No rootCmd.AddCommand here, it's added in root.go
}
