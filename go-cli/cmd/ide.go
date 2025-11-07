package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

// ideCmd represents the ide command
var ideCmd = &cobra.Command{
	Use:   "ide",
	Short: "Open the current project in an IDE",
	Long:  `The ide command opens the current project in a configured IDE (e.g., VS Code).`, //nolint:staticcheck
	Run: func(cmd *cobra.Command, args []string) {
		// For now, assume VS Code is the default IDE and it's in the PATH.
		// TODO: Allow configuration of the IDE.
		var command *exec.Cmd

		projectPath, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error getting current working directory: %v\n", err)
			return
		}

		switch runtime.GOOS {
		case "darwin": // macOS
			command = exec.Command("code", projectPath) // Assumes VS Code is installed and in PATH
		case "linux": // Linux
			command = exec.Command("code", projectPath) // Assumes VS Code is installed and in PATH
		case "windows": // Windows
			command = exec.Command("cmd", "/c", "code", projectPath) // Assumes VS Code is installed and in PATH
		default:
			fmt.Printf("Unsupported operating system: %s\n", runtime.GOOS)
			return
		}

		command.Stdout = os.Stdout
		command.Stderr = os.Stderr

		err = command.Run()
		if err != nil {
			fmt.Printf("Error opening project in IDE: %v\n", err)
			return
		}

		fmt.Println("Opened project in IDE.")
	},
}
