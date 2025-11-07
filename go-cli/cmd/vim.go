package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// vimCmd represents the vim command
var vimCmd = &cobra.Command{
	Use:   "vim <file_path>",
	Short: "Open a file in Vim",
	Long:  `The vim command opens a specified file in the Vim text editor.`, //nolint:staticcheck
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]

		command := exec.Command("vim", filePath)
		command.Stdin = os.Stdin
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr

		err := command.Run()
		if err != nil {
			fmt.Printf("Error opening file '%s' in Vim: %v\n", filePath, err)
			return
		}

		fmt.Printf("Opened file '%s' in Vim.\n", filePath)
	},
}
