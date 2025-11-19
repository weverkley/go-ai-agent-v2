package cmd

import (
	"fmt"
	"os"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/core"
	"go-ai-agent-v2/go-cli/pkg/services" // Import services
	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/spf13/cobra"
)

func init() {
	generateCmd.Flags().StringVarP(&executorType, "executor", "e", "gemini", "The type of AI executor to use (e.g., 'gemini', 'mock')")
}

var generateCmd = &cobra.Command{
	Use:   "generate [prompt]",
	Short: "Generate content using a prompt",
	Long:  `Generate content using a specified prompt.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runGenerateCmd(cmd, args, SettingsService, ShellService)
	},
}

// runGenerateCmd contains the logic for the generate command, accepting necessary services.
func runGenerateCmd(cmd *cobra.Command, args []string, settingsService *services.SettingsService, shellService services.ShellExecutionService) {
	modelVal, ok := settingsService.Get("model")
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
	executor, err := factory.NewExecutor(appConfig, types.GenerateContentConfig{}, []*types.Content{})
	if err != nil {
		fmt.Printf("Error creating executor: %v\n", err)
		os.Exit(1)
	}

	prompt := strings.Join(args, " ")

	// Direct instantiation of types.Content
	userContent := &types.Content{
		Role:  "user",
		Parts: []types.Part{{Text: prompt}},
	}

	resp, err := executor.GenerateContent(userContent)
	if err != nil {
		fmt.Printf("Error generating content: %v\n", err)
		os.Exit(1)
	}

	var textResponse string
	if resp != nil && len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
		for _, part := range resp.Candidates[0].Content.Parts {
			// Access Text field directly
			if part.Text != "" {
				textResponse += part.Text
			}
		}
	}
	fmt.Println(textResponse)
}