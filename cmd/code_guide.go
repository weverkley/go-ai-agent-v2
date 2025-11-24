package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/core"
	"go-ai-agent-v2/go-cli/pkg/services" // Import services
	"go-ai-agent-v2/go-cli/pkg/tools"
	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/spf13/cobra"
)

var codeGuideCmd = &cobra.Command{
	Use:   "code-guide [question]",
	Short: "Answer questions about the Go AI Agent codebase",
	Long: `Answer questions about the Go AI Agent codebase with explanations and code snippets.
This command acts as a specialized AI prompt to help new engineers understand the Go AI Agent codebase.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runCodeGuideCmd(cmd, args, SettingsService, ShellService, WorkspaceService)
	},
}

// runCodeGuideCmd handles the code-guide command logic.
func runCodeGuideCmd(cmd *cobra.Command, args []string, settingsService types.SettingsServiceIface, shellService services.ShellExecutionService, workspaceService *services.WorkspaceService) {
	// Initialize FileSystemService and GitService
	fsService := services.NewFileSystemService()

	// Register tools
	toolRegistry := tools.RegisterAllTools(fsService, shellService, settingsService, workspaceService)

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
	executor, err := factory.NewExecutor(appConfig, types.GenerateContentConfig{}, []*types.Content{})
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

func init() {
	codeGuideCmd.Flags().StringVarP(&executorType, "executor", "e", "gemini", "The type of AI executor to use (e.g., 'gemini', 'mock')")
}

