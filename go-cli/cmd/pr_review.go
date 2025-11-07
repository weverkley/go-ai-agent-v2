package cmd

import (
	"fmt"
	"os"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/core"
	"go-ai-agent-v2/go-cli/pkg/types"
	"go-ai-agent-v2/go-cli/pkg/ui"
	"go-ai-agent-v2/go-cli/pkg/tools"

	"github.com/google/generative-ai-go/genai"
	"github.com/spf13/cobra"
	"github.com/charmbracelet/bubbletea"
)

var prReviewCmd = &cobra.Command{
	Use:   "pr-review [pr_identifier]",
	Short: "Review a specific pull request",
	Long: `This command uses AI to conduct a comprehensive review of a pull request.
It evaluates code quality, adherence to standards, and readiness for merging, providing detailed feedback or approval messages.`, 
	Args: cobra.MinimumNArgs(0), // Allow 0 arguments for interactive mode
	Run: func(cmd *cobra.Command, args []string) {
		// Load the configuration within the command's Run function
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
			fmt.Fprintf(os.Stderr, "Error creating executor: %v\n", err)
			os.Exit(1)
		}

		// If no question is provided, launch interactive UI
		if len(args) == 0 {
			p := tea.NewProgram(ui.NewPrReviewModel(executor))
			if _, err := p.Run(); err != nil {
				fmt.Printf("Error running interactive pr-review: %v\n", err)
				os.Exit(1)
			}
			return
		}

		prIdentifier := strings.Join(args, " ")

		// Construct a simpler prompt for the AI, guiding it to use tools
		prompt := fmt.Sprintf(`## Mission: Comprehensive Pull Request Review

Your task is to conduct a comprehensive review of the pull request identified by "%s".

### Workflow:

1.  **PR Preparation & Initial Assessment**:
    *   Use the 'checkout_branch' tool to check out the designated PR into a temporary branch.
    *   Use the 'execute_command' tool to run "npm run preflight". This includes building, linting, and running all unit tests.
    *   Analyze the output of these preflight checks, noting any failures, warnings, or linting issues.

2.  **In-Depth Code Review**:
    *   Use the 'list_directory' and 'read_file' tools to explore the codebase and review the changes introduced in the PR. Focus your analysis on:
        *   Correctness, Maintainability, Readability, Efficiency, Security, Edge Cases and Error Handling, Testability.
    *   Based on your analysis, determine if the PR is safe to merge.

3.  **Reviewing Previous Feedback**:
    *   If necessary, use tools to access and examine the PR's history to identify any outstanding requests or unresolved comments from previous reviews.

4.  **Decision and Output Generation**:
    *   If the PR is deemed safe to merge, draft a friendly, concise, and professional approval message.
    *   If the PR is NOT safe to merge, provide a clear, constructive, and detailed summary of the issues found, and suggest specific actionable changes.

### Post-PR Action:

*   After providing your review and decision, use the 'checkout_branch' tool to switch back to the main branch.
*   Use the 'execute_command' tool to clean up any temporary branches or files.
*   Use the 'pull' tool to ensure the main branch is synchronized with the latest upstream changes.

### PR Identifier:

%s`, prIdentifier, prIdentifier)

		// Initial content for the chat
		contents := []*genai.Content{core.NewUserContent(prompt)}

		// Main loop for tool calling
		for {
			resp, err := executor.GenerateContent(contents...)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error generating content: %v\n", err)
				os.Exit(1)
			}

			if len(resp.Candidates) == 0 {
				fmt.Println("No candidates returned from Gemini.")
				os.Exit(0)
			}

			candidate := resp.Candidates[0]
			if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
				fmt.Println("No content parts returned from Gemini.")
				os.Exit(0)
			}

			var toolCalls []*genai.FunctionCall
			var textResponse string

			for _, part := range candidate.Content.Parts {
				if fc, ok := part.(*genai.FunctionCall); ok && fc != nil {
					toolCalls = append(toolCalls, fc)
				} else if text, ok := part.(genai.Text); ok && text != "" {
					textResponse += string(text)
				}
			}

			if len(toolCalls) > 0 {
				// Execute tool calls
				var toolResponses []genai.Part
				for _, fc := range toolCalls {
					toolResult, err := executor.ExecuteTool(fc)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error executing tool %s: %v\n", fc.Name, err)
						os.Exit(1)
					}
					toolResponses = append(toolResponses, core.NewFunctionResponsePart(fc.Name, toolResult.LLMContent))
				}
				// Append tool calls and their responses to the conversation history
				contents = append(contents, core.NewFunctionCallContent(toolCalls...))
				contents = append(contents, core.NewToolContent(toolResponses...))
			} else {
				// If no tool calls, it's the final answer
				fmt.Println(textResponse)
				return
			}
		}
	},
}

func init() {
	prReviewCmd.Flags().StringVarP(&executorType, "executor", "e", "gemini", "The type of AI executor to use (e.g., 'gemini', 'mock')")
}