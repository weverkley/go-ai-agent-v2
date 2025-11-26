package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// compressCmd represents the compress command
var compressCmd = &cobra.Command{
	Use:   "compress",
	Short: "Compress the current chat history to save tokens",
	Long: `This command summarizes the current conversation, replacing the existing
history with the summary. This helps to reduce the number of tokens sent to the
AI model in subsequent requests, which can save costs and prevent context
window limits.

This command is intended to be run from within the interactive chat as /compress.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("This command is only available within the interactive chat. Use `/compress`.")
		return nil
	},
}

func init() {
	RootCmd.AddCommand(compressCmd)
}