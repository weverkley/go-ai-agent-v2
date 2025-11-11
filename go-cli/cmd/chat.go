package cmd

import (
	"fmt"
	"os"

	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/core"
	"go-ai-agent-v2/go-cli/pkg/types"
	"go-ai-agent-v2/go-cli/pkg/ui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/generative-ai-go/genai"
	"github.com/spf13/cobra"
)

func init() {
	chatCmd.Flags().StringVarP(&executorType, "executor", "e", "gemini", "The type of AI executor to use (e.g., 'gemini', 'mock')")
}

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Start an interactive chat session with the AI agent",
	Run: func(cmd *cobra.Command, args []string) {
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

		params := &config.ConfigParameters{
			ModelName: model,
		}
		appConfig := config.NewConfig(params)

		factory, err := core.NewExecutorFactory(executorType, appConfig)
		if err != nil {
			fmt.Printf("Error creating executor factory: %v\n", err)
			os.Exit(1)
		}
		executor, err := factory.NewExecutor(appConfig, types.GenerateContentConfig{}, []*genai.Content{})
		if err != nil {
			fmt.Printf("Error creating executor: %v\n", err)
			os.Exit(1)
		}

		p := tea.NewProgram(ui.NewChatModel(executor, RootCmd)) // Pass RootCmd here
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error running interactive chat: %v\n", err)
			os.Exit(1)
		}
	},
}