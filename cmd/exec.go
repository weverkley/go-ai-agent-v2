package cmd

import (
	"fmt"
	"os"

	"go-ai-agent-v2/go-cli/pkg/services"
	"github.com/spf13/cobra"
)

var execCommand string
var execWorkingDir string

func init() {
	execCmd.Flags().StringVarP(&execCommand, "command", "c", "", "The shell command to execute")
	execCmd.Flags().StringVarP(&execWorkingDir, "path", "p", ".", "The working directory for the command")
	_ = execCmd.MarkFlagRequired("command")
}

var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "Execute a shell command",
	Long:  `Execute a shell command in a specified working directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		shellService := services.NewShellExecutionService()
		stdout, stderr, err := shellService.ExecuteCommand(execCommand, execWorkingDir)
		if err != nil {
			fmt.Printf("Error executing command: %v\n", err)
			if stdout != "" {
				fmt.Printf("Stdout:\n%s\n", stdout)
			}
			if stderr != "" {
				fmt.Printf("Stderr:\n%s\n", stderr)
			}
			os.Exit(1)
		}
		if stdout != "" {
			fmt.Printf("Stdout:\n%s\n", stdout)
		}
		if stderr != "" {
			fmt.Printf("Stderr:\n%s\n", stderr)
		}
	},
}
