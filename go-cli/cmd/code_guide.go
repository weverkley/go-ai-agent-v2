package cmd

import (
	"fmt"
	"os"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/core"
	"go-ai-agent-v2/go-cli/pkg/types"
	"go-ai-agent-v2/go-cli/pkg/ui"

	"github.com/google/generative-ai-go/genai"
	"github.com/spf13/cobra"
	"github.com/charmbracelet/bubbletea"
)

var codeGuideCmd = &cobra.Command{
	Use:   "code-guide [question]",
	Short: "Answer questions about the Gemini CLI codebase",
	Long: `Answer questions about the Gemini CLI codebase with explanations and code snippets.

This command acts as a specialized AI prompt to help new engineers understand the Gemini CLI codebase.
It provides clear explanations grounded in the actual source code, including full file paths and design choices.`,
	Args: cobra.MinimumNArgs(0), // Allow 0 arguments for interactive mode
	Run: func(cmd *cobra.Command, args []string) {
		// Load the configuration within the command's Run function
		// This ensures that cfg is initialized correctly based on the current working directory
		workspaceDir, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error getting current working directory: %v\n", err)
			os.Exit(1)
		}
		loadedSettings := config.LoadSettings(workspaceDir)

		// Create a ConfigParameters object
		params := &config.ConfigParameters{
			Model: loadedSettings.Model,
		}

		// Create a config.Config object
		appConfig := config.NewConfig(params)

		// Update params with ToolRegistry from appConfig
		params.ToolRegistry = appConfig.ToolRegistry

		geminiClient, err := core.NewGeminiChat(appConfig, types.GenerateContentConfig{}, []*genai.Content{})
		if err != nil {
			fmt.Printf("Error initializing GeminiChat: %v\n", err)
			os.Exit(1)
		}

		// If no question is provided, launch interactive UI
		if len(args) == 0 {
			p := tea.NewProgram(ui.NewCodeGuideModel(geminiClient))
			if _, err := p.Run(); err != nil {
				fmt.Printf("Error running interactive code-guide: %v\n", err)
				os.Exit(1)
			}
			return
		}

		question := strings.Join(args, " ")

		// The prompt content from code-guide.toml
		promptTemplate := `## Mission: Explain the Gemini CLI Codebase

Your primary task is to help a new engineer understand the Gemini CLI codebase by answering their questions about architecture, specific functions, and project structure.


### Objective:

Your primary task is to help a new engineer understand the Gemini CLI codebase. You will answer their questions about architecture, specific functions, and project structure by providing clear explanations grounded in the actual source code.


### Instructions:

1.  **Always Consult "Getting Started"**: Before providing any answer, you MUST first consult the getting started documentation located in the docs/get-started folder.

2.  **Consult Documentation and Specific Folders**: Before answering, you MUST first consult any relevant documentation within the docs folder. Base all your code-related answers exclusively on the contents of the following folders:  integration-tests, packages, and scripts.

3.  **Provide Specific Code Examples**: Always support your explanations with relevant code snippets. You MUST include the full file path (e.g., packages/gemini/core.py) so the user can easily locate the code.

4.  **Explain the "Why"**: Go beyond simply showing the code. Explain the design choices and the rationale behind the implementation. Discuss why a particular approach was taken and what trade-offs might have been considered.

5.  **Suggest a Learning Path**: Where appropriate, guide the user by suggesting related files to examine next or other relevant concepts to explore within the codebase to deepen their understanding.

6.  **Handle Unknowns Gracefully**: If the answer cannot be found in the provided folders and documentation, you must state that the information is unavailable and ask the user for clarification. Do not invent answers or speculate.


### Constraints:


1. No Hallucination: If the answer cannot be found in the provided context, you must state that the information is unavailable and ask the user for clarification. Do not invent answers or speculate.

2. Stay Focused: Only answer questions directly related to the Gemini CLI project within the specified folders.

### QUESTION:

%s`
		
		finalPrompt := fmt.Sprintf(promptTemplate, question)

		resp, err := geminiClient.GenerateContent(core.NewUserContent(finalPrompt))
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
	},
}

func init() {
	rootCmd.AddCommand(codeGuideCmd)
}
