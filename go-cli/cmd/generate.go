package cmd

import (
	"fmt"
	"os"

	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/core"
	"go-ai-agent-v2/go-cli/pkg/prompts"
	"go-ai-agent-v2/go-cli/pkg/types"
	"go-ai-agent-v2/go-cli/pkg/ui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/generative-ai-go/genai"
	"github.com/spf13/cobra"
)

var promptName string

func init() {
	generateCmd.Flags().StringVarP(&promptName, "prompt", "p", "default", "The name of the prompt to use for content generation")
	generateCmd.Flags().StringVarP(&executorType, "executor", "e", "gemini", "The type of AI executor to use (e.g., 'gemini', 'mock')")
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate content using a prompt",
	Long:  `Generate content using a specified prompt. If no prompt is provided, an interactive UI will be launched.`, 
	Run: func(cmd *cobra.Command, args []string) {
		// Load the configuration within the command's Run function
		workspaceDir, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error getting current working directory: %v\n", err)
			os.Exit(1)
		}
		loadedSettings := config.LoadSettings(workspaceDir)

		params := &config.ConfigParameters{
			Model: loadedSettings.Model,
		}
		appConfig := config.NewConfig(params)

		executorFactory := core.NewExecutorFactory()
		executor, err := executorFactory.CreateExecutor(executorType, appConfig, types.GenerateContentConfig{}, []*genai.Content{})
		if err != nil {
			fmt.Printf("Error creating executor: %v\n", err)
			os.Exit(1)
		}

		// If a prompt is provided as an argument, run in non-interactive mode
		if len(args) > 0 || promptName != "default" {
			promptManager := prompts.NewPromptManager()
			promptManager.AddPrompt(prompts.DiscoveredMCPPrompt{Name: "default", Description: "Translate the following Go code to Javascript:", ServerName: "cli"})

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

			resp, err := executor.GenerateContent(core.NewUserContent(finalPrompt))
			if err != nil {
				fmt.Printf("Error generating content: %v\n", err)
				os.Exit(1)
			}

			var textResponse string
			if resp != nil && len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
				for _, part := range resp.Candidates[0].Content.Parts {
					if txt, ok := part.(genai.Text); ok {
						textResponse += string(txt)
					}
				}
			}
			fmt.Println(textResponse)
		} else {
			// Launch interactive UI
			p := tea.NewProgram(ui.NewGenerateModel(executor)) // Pass executor here
			if _, err := p.Run(); err != nil {
				fmt.Printf("Error running interactive generate: %v\n", err)
				os.Exit(1)
			}
		}
	},
}