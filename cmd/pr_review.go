package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/core"
	"go-ai-agent-v2/go-cli/pkg/tools"
	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/google/generative-ai-go/genai"
	"github.com/spf13/cobra"
)

var prReviewCmd = &cobra.Command{
	Use:   "pr-review [pr_identifier]",
	Short: "Review a specific pull request",
	Long: `This command uses AI to conduct a comprehensive review of a pull request.
It evaluates code quality, adherence to standards, and readiness for merging, providing detailed feedback or approval messages.`, 
	Args: cobra.MinimumNArgs(1),
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
			ModelName:    model,
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

		prIdentifier := strings.Join(args, " ")

		promptTemplate, err := ioutil.ReadFile("prompts/pr_review.md")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading prompt template: %v\n", err)
			os.Exit(1)
		}

		prompt := fmt.Sprintf(string(promptTemplate), prIdentifier, prIdentifier)

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
