package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/core"
	"go-ai-agent-v2/go-cli/pkg/tools"
	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/google/generative-ai-go/genai"
	"github.com/spf13/cobra"
)

var codeGuideCmd = &cobra.Command{
	Use:   "code-guide [question]",
	Short: "Answer questions about the Gemini CLI codebase",
	Long: `Answer questions about the Gemini CLI codebase with explanations and code snippets.

This command acts as a specialized AI prompt to help new engineers understand the Gemini CLI codebase.
It provides clear explanations grounded in the actual source code, including full file paths and design choices.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize the ToolRegistry
		toolRegistry := tools.RegisterAllTools(FSService)

		modelVal, ok := SettingsService.Get("model")
		if !ok {
			fmt.Printf("Error: 'model' setting not found.\n")
			os.Exit(1)
		}
		model, ok := modelVal.(string)
		if !ok {
			fmt.Printf("Error: 'model' setting is not a string.\n")
			os.Exit(1)
		}

		// Create a ConfigParameters object
		params := &config.ConfigParameters{
			ModelName:    model,
			ToolRegistry: toolRegistry, // Use the initialized tool registry
		}

		// Create a config.Config object
		appConfig := config.NewConfig(params)
		factory, err := core.NewExecutorFactory(executorType, appConfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating executor factory: %v\n", err)
			os.Exit(1)
		}
		executor, err := factory.NewExecutor(appConfig, types.GenerateContentConfig{}, []*genai.Content{})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating executor: %v\n", err)
			os.Exit(1)
		}

		question := strings.Join(args, " ")

		promptTemplate, err := ioutil.ReadFile("prompts/code_guide.md")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading prompt template: %v\n", err)
			os.Exit(1)
		}

		finalPrompt := fmt.Sprintf(string(promptTemplate), question)

		resp, err := executor.GenerateContent(core.NewUserContent(finalPrompt))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating content: %v\n", err)
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
	},
}

func init() {
	codeGuideCmd.Flags().StringVarP(&executorType, "executor", "e", "gemini", "The type of AI executor to use (e.g., 'gemini', 'mock')")
}

