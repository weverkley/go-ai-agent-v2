package cmd

import (
	"fmt"
	"os"

	"go-ai-agent-v2/go-cli/pkg/core"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listModelsCmd)
}

var listModelsCmd = &cobra.Command{
	Use:   "list-models",
	Short: "List available Gemini models",
	Long:  `List all available Gemini models that can be used with the generate command.`,
	Run: func(cmd *cobra.Command, args []string) {
		geminiClient, err := core.NewGeminiChat()
		if err != nil {
			fmt.Printf("Error initializing GeminiChat: %v\n", err)
			os.Exit(1)
		}
		models, err := geminiClient.ListModels()
		if err != nil {
			fmt.Printf("Error listing models: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Available Gemini Models:")
		for _, model := range models {
			fmt.Printf("- %s\n", model)
		}
	},
}
