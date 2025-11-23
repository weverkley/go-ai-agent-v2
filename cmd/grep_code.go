package cmd

import (
	"context" // Add context import
	"fmt"
	"os"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/core"
	"go-ai-agent-v2/go-cli/pkg/services" // Import services
	"go-ai-agent-v2/go-cli/pkg/tools"
	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/spf13/cobra"
)

var grepCodeCmd = &cobra.Command{
	Use:   "grep-code [pattern]",
	Short: "Summarize findings for a given code pattern.",
	Long:  `This command uses grep to search for a code pattern and then uses AI to summarize the findings.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runGrepCodeCmd(cmd, args, SettingsService, ShellService)
	},
}

// runGrepCodeCmd contains the logic for the grep-code command, accepting necessary services.
func runGrepCodeCmd(cmd *cobra.Command, args []string, settingsService types.SettingsServiceIface, shellService services.ShellExecutionService) {
	// Initialize the ToolRegistry
	toolRegistry := tools.RegisterAllTools(FSService, shellService, settingsService)

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
		ModelName:    model,
		ToolRegistry: toolRegistry, // Use the initialized tool registry
	}

	appConfig := config.NewConfig(params)

	factory, err := core.NewExecutorFactory(executorType, appConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating executor factory: %v\n", err)
		os.Exit(1)
	}
	executor, err := factory.NewExecutor(appConfig, types.GenerateContentConfig{}, []*types.Content{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating executor: %v\n", err)
		os.Exit(1)
	}

	pattern := strings.Join(args, " ")

	// Execute grep command
	grepCommand := fmt.Sprintf("grep -r %s .", pattern)
	grepOutput, grepStderr, err := shellService.ExecuteCommand(context.Background(), grepCommand, ".")
	if err != nil {
		fmt.Printf("Error executing grep command: %v\n", err)
		if grepOutput != "" {
			fmt.Printf("Stdout:\n%s\n", grepOutput)
		}
		if grepStderr != "" {
			fmt.Printf("Stderr:\n%s\n", grepStderr)
		}
		os.Exit(1)
	}

	// The prompt content from grep-code.toml
	promptTemplate := `Please summarize the findings for the pattern %s.\n\nSearch Results:\n%s`

	finalPrompt := fmt.Sprintf(promptTemplate, pattern, grepOutput)

	// Direct instantiation of types.Content
	userContent := &types.Content{
		Role:  "user",
		Parts: []types.Part{{Text: finalPrompt}},
	}

	resp, err := executor.GenerateContent(userContent)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating content: %v\n", err)
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
