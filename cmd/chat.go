package cmd

import (
	"fmt"
	"os"

	"go-ai-agent-v2/go-cli/pkg/core"
	"go-ai-agent-v2/go-cli/pkg/routing"
	"go-ai-agent-v2/go-cli/pkg/types"
	"go-ai-agent-v2/go-cli/pkg/ui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/generative-ai-go/genai"
	"github.com/spf13/cobra"
)

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

		executorVal, ok := SettingsService.Get("executor")
		if !ok {
			fmt.Printf("Error: 'executor' setting not found.\n")
			os.Exit(1)
		}
		executorType, ok := executorVal.(string)
		if !ok {
			fmt.Printf("Error: 'executor' setting is not a string.\n")
			os.Exit(1)
		}

		// Use the global Cfg and override the model
		appConfig := Cfg.WithModel(model)

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

		router := routing.NewModelRouterService(appConfig)
		p := tea.NewProgram(ui.NewChatModel(executor, executorType, appConfig, router, RootCmd), tea.WithAltScreen()) // Pass RootCmd here
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error running interactive chat: %v\n", err)
			os.Exit(1)
		}
	},
}