package cmd

import (
	"fmt"
	"os"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/core"
	"go-ai-agent-v2/go-cli/pkg/services" // Keep this import for shellService
	"go-ai-agent-v2/go-cli/pkg/types"
	"go-ai-agent-v2/go-cli/pkg/ui"
	"go-ai-agent-v2/go-cli/pkg/tools" // Add back the import

	"github.com/google/generative-ai-go/genai"
	"github.com/spf13/cobra"
	"github.com/charmbracelet/bubbletea"
)

var grepCodeCmd = &cobra.Command{
	Use:   "grep-code [pattern]",
	Short: "Summarize findings for a given code pattern.",
	Long: `This command uses grep to search for a code pattern and then uses AI to summarize the findings.`, 
	Args: cobra.MinimumNArgs(0), // Allow 0 arguments for interactive mode
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

		params := &config.ConfigParameters{
			ModelName: model,
			ToolRegistry: toolRegistry, // Use the initialized tool registry
		}

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

		// If no question is provided, launch interactive UI
		if len(args) == 0 {
			p := tea.NewProgram(ui.NewGrepCodeModel(executor))
			if _, err := p.Run(); err != nil {
				fmt.Printf("Error running interactive grep-code: %v\n", err)
				os.Exit(1)
			}
			return
		}

		pattern := strings.Join(args, " ")

		// Execute grep command
		shellService := services.NewShellExecutionService()
		grepCommand := fmt.Sprintf("grep -r %s .", pattern)
		grepOutput, grepStderr, err := shellService.ExecuteCommand(grepCommand, ".")
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
	grepCodeCmd.Flags().StringVarP(&executorType, "executor", "e", "gemini", "The type of AI executor to use (e.g., 'gemini', 'mock')")
}
