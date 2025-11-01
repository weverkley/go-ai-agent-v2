package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/core"
	"go-ai-agent-v2/go-cli/pkg/services"
	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/google/generative-ai-go/genai"
	"github.com/spf13/cobra"
)

var findDocsCmd = &cobra.Command{
	Use:   "find-docs [question]",
	Short: "Find relevant documentation and output GitHub URLs.",
	Long: `Find relevant documentation within the current Git repository and output GitHub URLs.

This command uses AI to search for documentation files related to your question and provides direct links to them on GitHub.`, 
	Args: cobra.MinimumNArgs(1), // Requires at least one argument for the question
	Run: func(cmd *cobra.Command, args []string) {
		question := strings.Join(args, " ")

						// Get Git repository details

						gitService := services.NewGitService()

						workspaceDir, err := os.Getwd() // Re-introduce this line

						if err != nil {

							fmt.Printf("Error getting current working directory: %v\n", err)

							os.Exit(1)

						}

						// Assume project root is the parent directory of go-cli

						projectRoot := filepath.Dir(workspaceDir)

				

						remoteURL, err := gitService.GetRemoteURL(projectRoot)

						if err != nil {

							fmt.Printf("Error getting Git remote URL: %v\n", err)

							os.Exit(1)

						}

				

						currentBranch, err := gitService.GetCurrentBranch(projectRoot)

						if err != nil {
			fmt.Printf("Error getting current Git branch: %v\n", err)
			os.Exit(1)
		}

		// Construct the prompt for the AI
		promptTemplate := `## Mission: Find Relevant Documentation

Your task is to find documentation files relevant to the user's question within the current git repository and provide a list of GitHub URLs to view them.

### Repository Details:
- Remote URL: %s
- Current Branch: %s

### Workflow:

1.  **Identify Repository Details**:
    *   You have been provided with the remote URL and current branch.
    *   From the remote URL, parse and construct the base GitHub URL (e.g., https://github.com/user/repo). You must handle both HTTPS (https://github.com/user/repo.git) and SSH (git@github.com:user/repo.git) formats.

2.  **Search for Documentation**:
    *   First, perform a targeted search across the repository for documentation files (e.g., .md, .mdx) that seem directly related to the user's question.
    *   If this initial search yields no relevant results, and a docs/ directory exists, read the content of all files within the docs/ directory to find relevant information.
    *   If you still can't find a direct match, broaden your search to include related concepts and synonyms of the keywords in the user's question.
    *   For each file you identify as potentially relevant, read its content to confirm it addresses the user's query.

3.  **Construct and Output URLs**:
    *   For each file you identify as relevant, construct the full GitHub URL by combining the base URL, branch, and file path. Do not use shell commands for this step.
    *   The URL format should be: {BASE_GITHUB_URL}/blob/{BRANCH_NAME}/{PATH_TO_FILE_FROM_REPO_ROOT}.
    *   Present the final list to the user as a markdown list. Each item in the list should be the URL to the document, followed by a short summary of its content.
    *   If, after all search attempts, you cannot find any relevant documentation, ask the user clarifying questions to better understand their needs. Do not return any URLs in this case.

### QUESTION:

%s`

		finalPrompt := fmt.Sprintf(promptTemplate, remoteURL, currentBranch, question)

		// Load the configuration within the command's Run function
		workspaceDir, err = os.Getwd()
		if err != nil {
			fmt.Printf("Error getting current working directory: %v\n", err)
			os.Exit(1)
		}
		loadedSettings := config.LoadSettings(workspaceDir)

		params := &config.ConfigParameters{
			Model: loadedSettings.Model,
		}
		appConfig := config.NewConfig(params)

		geminiClient, err := core.NewGeminiChat(appConfig, types.GenerateContentConfig{}, []*genai.Content{})
		if err != nil {
			fmt.Printf("Error initializing GeminiChat: %v\n", err)
			os.Exit(1)
		}

		content, err := geminiClient.GenerateContent(finalPrompt)
		if err != nil {
			fmt.Printf("Error generating content: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(content)
	},
}

func init() {
	rootCmd.AddCommand(findDocsCmd)
}
