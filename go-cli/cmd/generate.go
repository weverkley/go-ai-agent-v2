package cmd

import (
	"fmt"
	"os"

	"go-ai-agent-v2/go-cli/pkg/core"
	"go-ai-agent-v2/go-cli/pkg/core/agents"
	"go-ai-agent-v2/go-cli/pkg/prompts"

	"github.com/google/generative-ai-go/genai"
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
		promptManager.AddPrompt(prompts.DiscoveredMCPPrompt{Name: "default", Description: "Translate the following Go code to Javascript:", ServerName: "cli"})

		// Initialize GeminiChat using the global config
		// For now, using default generation config and empty history.
		// This will be properly set up when agent execution is integrated.
		geminiClient, err := core.NewGeminiChat(cfg, agents.GenerateContentConfig{}, []genai.Content{})
		if err != nil {
			fmt.Printf("Error initializing GeminiChat: %v\n", err)
			os.Exit(1)
		}

		prompt, err := promptManager.GetPrompt(promptName)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		// If there are command-line arguments, use them as the prompt
		var finalPrompt string
		if len(args) > 0 {
			finalPrompt = args[0]
		} else {
			finalPrompt = prompt.Description
		}

		content, err := geminiClient.GenerateContent(finalPrompt)
		if err != nil {
			fmt.Printf("Error generating content: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(content)
	},
}
