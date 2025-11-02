package cmd

import (
	"fmt"
	"os"

	"go-ai-agent-v2/go-cli/pkg/core"
	"go-ai-agent-v2/go-cli/pkg/types"
	"go-ai-agent-v2/go-cli/pkg/tools"
	"go-ai-agent-v2/go-cli/pkg/config"

	"github.com/google/generative-ai-go/genai"
	"github.com/spf13/cobra"
)


func init() {
	rootCmd.AddCommand(listModelsCmd)
	listModelsCmd.Flags().StringVarP(&executorType, "executor", "e", "gemini", "The type of AI executor to use (e.g., 'gemini', 'mock')")
}

var listModelsCmd = &cobra.Command{
	Use:   "list-models",
	Short: "List available Gemini models",
	Long:  `List all available Gemini models that can be used with the generate command.`,
	Run: func(cmd *cobra.Command, args []string) {
		workspaceDir, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error getting current working directory: %v\n", err)
			os.Exit(1)
		}
		loadedSettings := config.LoadSettings(workspaceDir)

		// Initialize the ToolRegistry
		toolRegistry := tools.RegisterAllTools()

		params := &config.ConfigParameters{
			Model: loadedSettings.Model,
			ToolRegistry: toolRegistry, // Use the initialized tool registry
		}
		appConfig := config.NewConfig(params)

		executorFactory := core.NewExecutorFactory()
		executor, err := executorFactory.CreateExecutor(executorType, appConfig, types.GenerateContentConfig{}, []*genai.Content{})
		if err != nil {
			fmt.Printf("Error creating executor: %v\n", err)
			os.Exit(1)
		}

		models, err := executor.ListModels()
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
