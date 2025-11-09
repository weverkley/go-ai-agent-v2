package cmd

import (
	"fmt"
	"os"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/core"
	"go-ai-agent-v2/go-cli/pkg/types"
	"go-ai-agent-v2/go-cli/pkg/ui"
	"go-ai-agent-v2/go-cli/pkg/tools" // Add back the import

	"github.com/google/generative-ai-go/genai"
	"github.com/spf13/cobra"
	tea "github.com/charmbracelet/bubbletea"
)


func init() {
	findDocsCmd.Flags().StringVarP(&executorType, "executor", "e", "gemini", "The type of AI executor to use (e.g., 'gemini', 'mock')")
}

var findDocsCmd = &cobra.Command{
	Use:   "find-docs [question]",
	Short: "Find relevant documentation and output GitHub URLs.",
	Long: `Find relevant documentation within the current Git repository and output GitHub URLs.

This command uses AI to search for documentation files related to your question and provides direct links to them on GitHub.`, 
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

		factory, err := core.NewExecutorFactory(executorType)
		if err != nil {
			fmt.Printf("Error creating executor factory: %v\n", err)
			os.Exit(1)
		}
		executor, err := factory.NewExecutor(appConfig, types.GenerateContentConfig{}, []*genai.Content{})
		if err != nil {
			fmt.Printf("Error creating executor: %v\n", err)
			os.Exit(1)
		}

		// If no question is provided, launch interactive UI
		if len(args) == 0 {
			p := tea.NewProgram(ui.NewFindDocsModel(executor))
			if _, err := p.Run(); err != nil {
				fmt.Printf("Error running interactive find-docs: %v\n", err)
				os.Exit(1)
			}
			return
		}

		question := strings.Join(args, " ")

		// Construct a simpler prompt for the AI, guiding it to use tools
		prompt := fmt.Sprintf(`## Mission: Find Relevant Documentation

Your task is to find documentation files relevant to the user's question within the current git repository and provide a list of GitHub URLs to view them.

### Workflow:

1.  **Identify Repository Details**:
    *   Use the 'get_remote_url' and 'get_current_branch' tools to get the repository's remote URL and current branch.
    *   From the remote URL, parse and construct the base GitHub URL (e.g., https://github.com/user/repo). You must handle both HTTPS (https://github.com/user/repo.git) and SSH (git@github.com:user/repo.git) formats.

2.  **Search for Documentation**:
    *   Use the 'list_directory' tool to explore the repository for documentation files (e.g., .md, .mdx) that seem directly related to the user's question.
    *   If this initial search yields no relevant results, and a docs/ directory exists, use 'list_directory' and 'read_file' to read the content of all files within the docs/ directory to find relevant information.
    *   If you still can't find a direct match, broaden your search to include related concepts and synonyms of the keywords in the user's question.
    *   For each file you identify as potentially relevant, use 'read_file' to read its content to confirm it addresses the user's query.

3.  **Construct and Output URLs**:
    *   For each file you identify as relevant, construct the full GitHub URL by combining the base URL, branch, and file path. Do not use shell commands for this step.
    *   The URL format should be: {BASE_GITHUB_URL}/blob/{BRANCH_NAME}/{PATH_TO_FILE_FROM_REPO_ROOT}.
    *   Present the final list to the user as a markdown list. Each item in the list should be the URL to the document, followed by a short summary of its content.
    *   If, after all search attempts, you cannot find any relevant documentation, ask the user clarifying questions to better understand their needs. Do not return any URLs in this case.

### QUESTION:

%s`, question)

		// Initial content for the chat
		contents := []*genai.Content{core.NewUserContent(prompt)}

		// Main loop for tool calling
		for {
			resp, err := executor.GenerateContent(contents...)
			if err != nil {
				fmt.Printf("Error generating content: %v\n", err)
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
						fmt.Printf("Error executing tool %s: %v\n", fc.Name, err)
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
