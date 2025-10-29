package cmd

import (
	"fmt"
	"os"

	"go-ai-agent-v2/go-cli/pkg/core"
	"go-ai-agent-v2/go-cli/pkg/prompts"
	"github.com/spf13/cobra"
)

var promptName string

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().StringVarP(&promptName, "prompt", "p", "default", "The name of the prompt to use for content generation")
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate content using a prompt",
	Long:  `Generate content using a specified prompt.`,
	Run: func(cmd *cobra.Command, args []string) {
		promptManager := prompts.NewPromptManager()
		promptManager.AddPrompt("default", "Translate the following Go code to Javascript:")

		geminiClient, err := core.NewGeminiChat()
		if err != nil {
			fmt.Printf("Error initializing GeminiChat: %v\n", err)
			os.Exit(1)
		}

		prompt, err := promptManager.GetPrompt(promptName)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		content, err := geminiClient.GenerateContent(prompt)
		if err != nil {
			fmt.Printf("Error generating content: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(content)
	},
}
