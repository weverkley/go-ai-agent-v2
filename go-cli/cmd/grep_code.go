package cmd

import (
	"fmt"
	"os"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/core"
	"go-ai-agent-v2/go-cli/pkg/services"
	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/google/generative-ai-go/genai"
	"github.com/spf13/cobra"
)

var grepCodeCmd = &cobra.Command{
	Use:   "grep-code [pattern]",
	Short: "Summarize findings for a given code pattern.",
	Long: `This command uses grep to search for a code pattern and then uses AI to summarize the findings.`, 
	Args: cobra.MinimumNArgs(1), // Requires at least one argument for the pattern
	Run: func(cmd *cobra.Command, args []string) {
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
		promptTemplate := `Please summarize the findings for the pattern %s.

Search Results:
%s`
		
		finalPrompt := fmt.Sprintf(promptTemplate, pattern, grepOutput)

		// Load the configuration within the command's Run function
		workspaceDir, err := os.Getwd()
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
	rootCmd.AddCommand(grepCodeCmd)
}
