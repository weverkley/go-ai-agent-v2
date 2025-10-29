package cmd

import (
	"fmt"
	"os"

	"go-ai-agent-v2/go-cli/pkg/tools"
	"github.com/spf13/cobra"
)

var (
	memoryFact string
)

func init() {
	rootCmd.AddCommand(memoryCmd)
	memoryCmd.Flags().StringVarP(&memoryFact, "fact", "f", "", "The specific fact or piece of information to remember.")
	memoryCmd.MarkFlagRequired("fact")
}

var memoryCmd = &cobra.Command{
	Use:   "memory",
	Short: "Saves a specific piece of information or fact to your long-term memory.",
	Long:  `Saves a specific piece of information or fact to your long-term memory. Use this when the user explicitly asks you to remember something, or when they state a clear, concise fact that seems important to retain for future interactions.`, 
	Run: func(cmd *cobra.Command, args []string) {
		memoryTool := tools.NewMemoryTool()
		result, err := memoryTool.Execute(
			memoryFact,
		)
		if err != nil {
			fmt.Printf("Error executing memory command: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(result)
	},
}
