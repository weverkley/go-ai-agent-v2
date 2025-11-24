package cmd

import (
	"context" // Add context import
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

func init() {
	findDocsCmd.Flags().StringVarP(&executorType, "executor", "e", "gemini", "The type of AI executor to use (e.g., 'gemini', 'mock')")
}

var findDocsCmd = &cobra.Command{
	Use:   "find-docs [question]",
	Short: "Find relevant documentation and output GitHub URLs.",
	Long: `Find relevant documentation within the current Git repository and output GitHub URLs.

	This command uses AI to search for documentation files related to your question and provides direct links to them on GitHub.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runFindDocsCmd(cmd, args, SettingsService, ShellService, WorkspaceService)
	},
}

// runFindDocsCmd handles the find-docs command logic.
func runFindDocsCmd(cmd *cobra.Command, args []string, settingsService types.SettingsServiceIface, shellService services.ShellExecutionService, workspaceService *services.WorkspaceService) {
	// Initialize the ToolRegistry
	toolRegistry := tools.RegisterAllTools(FSService, shellService, settingsService, workspaceService)

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

	question := strings.Join(args, " ")

	promptTemplate, err := ioutil.ReadFile("prompts/find_docs.md")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading prompt template: %v\n", err)
		os.Exit(1)
	}

	prompt := fmt.Sprintf(string(promptTemplate), question)

	// Initial content for the chat
	contents := []*types.Content{
		{
			Role:  "user",
			Parts: []types.Part{{Text: prompt}},
		},
	}

	// Main loop for tool calling
	for {
		resp, err := executor.GenerateContent(contents...)
		if err != nil {
			fmt.Printf("Error generating content: %v\n", err)
			os.Exit(1)
		}

		if resp == nil || len(resp.Candidates) == 0 {
			fmt.Println("No candidates returned from executor.")
			os.Exit(0)
		}

		candidate := resp.Candidates[0]
		if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
			fmt.Println("No content parts returned from executor.")
			os.Exit(0)
		}

		var toolCalls []*types.FunctionCall
		var textResponse string

		for _, part := range candidate.Content.Parts {
			if part.FunctionCall != nil { // Directly access FunctionCall field
				toolCalls = append(toolCalls, part.FunctionCall)
			} else if part.Text != "" { // Directly access Text field
				textResponse += part.Text
			}
		}

		if len(toolCalls) > 0 {
			// Execute tool calls
			var toolResponses []types.Part
			for _, fc := range toolCalls {
				toolResult, err := executor.ExecuteTool(context.Background(), fc)
				if err != nil {
					fmt.Printf("Error executing tool %s: %v\n", fc.Name, err)
					os.Exit(1)
				}
				// Create types.Part for function response
				toolResponses = append(toolResponses, types.Part{
					FunctionResponse: &types.FunctionResponse{
						Name:     fc.Name,
						Response: map[string]interface{}{"result": toolResult.LLMContent},
					},
				})
			}
			// Append tool calls and their responses to the conversation history
			contents = append(contents, &types.Content{
				Role:  "model",
				Parts: []types.Part{{FunctionCall: toolCalls[0]}}, // Simplified for now, assuming one tool call
			})
			contents = append(contents, &types.Content{
				Role:  "tool",
				Parts: toolResponses,
			})
		} else {
			// If no tool calls, it's the final answer
			fmt.Println(textResponse)
			return
		}
	}
}