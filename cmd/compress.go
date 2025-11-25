package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// compressCmd represents the compress command
var compressCmd = &cobra.Command{
	Use:     "compress",
	Aliases: []string{"summarize"},
	Short:   "Compresses the context by replacing it with a summary",
	Long:    `The compress command compresses the current chat context by replacing it with a summary, reducing token count.`, //nolint:staticcheck
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Compressing chat history...")

		// // Check if there's a Go AI Agent client available
		// if executor == nil {
		// 	fmt.Fprintf(os.Stderr, "Error: Go AI Agent client not initialized. Cannot compress chat history.\n")
		// 	os.Exit(1)
		// }

		// // Call CompressChat method
		// result, err := executor.CompressChat("", false)
		// if err != nil {
		// 	fmt.Fprintf(os.Stderr, "Error compressing chat history: %v\n", err)
		// 	os.Exit(1)
		// }

		// fmt.Printf("Chat history compressed successfully. Original tokens: %d, New tokens: %d\n", result.OriginalTokenCount, result.NewTokenCount)
		fmt.Fprintf(os.Stderr, "Error: Compress command is not yet functional after refactoring. Use /clear in interactive chat for now.\n")
		os.Exit(1)
	},
}
