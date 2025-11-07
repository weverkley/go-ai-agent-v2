package cmd

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

// editorCmd represents the editor command
var editorCmd = &cobra.Command{
	Use:   "editor <file_path>",
	Short: "Open a file in the default editor",
	Long:  `The editor command opens a specified file in the system's default text editor.`, //nolint:staticcheck
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]
		var err error

		switch runtime.GOOS {
		case "darwin": // macOS
			err = exec.Command("open", filePath).Run()
		case "linux": // Linux
			err = exec.Command("xdg-open", filePath).Run()
		case "windows": // Windows
			err = exec.Command("cmd", "/c", "start", filePath).Run()
		default:
			fmt.Printf("Unsupported operating system: %s\n", runtime.GOOS)
			return
		}

		if err != nil {
			fmt.Printf("Error opening file '%s' in editor: %v\n", filePath, err)
			return
		}

		fmt.Printf("Opened file '%s' in default editor.\n", filePath)
	},
}
