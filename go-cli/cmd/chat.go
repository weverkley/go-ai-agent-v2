package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"encoding/json"

	"github.com/google/generative-ai-go/genai"

	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// chatCmd represents the chat command
var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Manage conversation history",
	Long:  `The chat command group allows you to manage your conversation history, including listing, saving, resuming, deleting, and sharing chat checkpoints.`, //nolint:staticcheck
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand is provided, print help
		cmd.Help()
	},
}

func init() {
	// Add subcommands here
	chatCmd.AddCommand(chatListCmd)
	chatCmd.AddCommand(chatSaveCmd)
	chatCmd.AddCommand(chatResumeCmd)
	chatCmd.AddCommand(chatDeleteCmd)
	chatCmd.AddCommand(chatShareCmd)
}

// chatListCmd represents the chat list subcommand
var chatListCmd = &cobra.Command{
	Use:   "list",
	Short: "List saved conversation checkpoints",
	Long:  `List saved conversation checkpoints.`, //nolint:staticcheck
	Run: func(cmd *cobra.Command, args []string) {
		chatDetails, err := chatService.GetSavedChatTags(false) // false for ascending order
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing chat checkpoints: %v\n", err)
			os.Exit(1)
		}

		if len(chatDetails) == 0 {
			fmt.Println("No saved chat checkpoints found.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "TAG\tLAST MODIFIED")
		fmt.Fprintln(w, "---\t-------------")
		for _, detail := range chatDetails {
			fmt.Fprintf(w, "%s\t%s\n", detail.Name, detail.Mtime.Format("2006-01-02 15:04:05"))
		}
		w.Flush()
	},
}

// chatSaveCmd represents the chat save subcommand
var chatSaveCmd = &cobra.Command{
	Use:   "save <tag>",
	Short: "Save the current conversation as a checkpoint",
	Long:  `Save the current conversation as a checkpoint. Usage: /chat save <tag>`, //nolint:staticcheck
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		tag := args[0]

		exists, err := chatService.CheckpointExists(tag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error checking for existing checkpoint: %v\n", err)
			os.Exit(1)
		}

		if exists {
			prompt := promptui.Prompt{
				Label:     fmt.Sprintf("Checkpoint with tag '%s' already exists. Overwrite?", tag),
				IsConfirm: true,
			}

			result, err := prompt.Run()
			if err != nil {
				if err == promptui.ErrInterrupt {
					fmt.Println("Operation cancelled.")
					os.Exit(0)
				}
				fmt.Fprintf(os.Stderr, "Prompt failed %v\n", err)
				os.Exit(1)
			}

			if result != "y" && result != "Y" {
				fmt.Println("Operation cancelled.")
				os.Exit(0)
			}
		}

		history, err := executor.GetHistory()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting chat history: %v\n", err)
			os.Exit(1)
		}

		if len(history) <= 2 { // Assuming initial system prompts are 2 messages
			fmt.Println("No conversation found to save.")
			return
		}

		if err := chatService.SaveCheckpoint(history, tag); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving chat checkpoint: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Conversation checkpoint saved with tag: %s\n", tag)
	},
}

// chatResumeCmd represents the chat resume subcommand
var chatResumeCmd = &cobra.Command{
	Use:     "resume <tag>",
	Aliases: []string{"load"},
	Short:   "Resume a conversation from a checkpoint",
	Long:    `Resume a conversation from a checkpoint. Usage: /chat resume <tag>`, //nolint:staticcheck
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		tag := args[0]

		history, err := chatService.LoadCheckpoint(tag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading chat checkpoint: %v\n", err)
			os.Exit(1)
		}

		if len(history) == 0 {
			fmt.Printf("No saved checkpoint found with tag: %s.\n", tag)
			return
		}

		if err := executor.SetHistory(history); err != nil {
			fmt.Fprintf(os.Stderr, "Error setting executor history: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Conversation checkpoint '%s' resumed successfully.\n", tag)
	},
}

// chatDeleteCmd represents the chat delete subcommand
var chatDeleteCmd = &cobra.Command{
	Use:   "delete <tag>",
	Short: "Delete a conversation checkpoint",
	Long:  `Delete a conversation checkpoint. Usage: /chat delete <tag>`, //nolint:staticcheck
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		tag := args[0]

		deleted, err := chatService.DeleteCheckpoint(tag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error deleting chat checkpoint: %v\n", err)
			os.Exit(1)
		}

		if deleted {
			fmt.Printf("Conversation checkpoint '%s' has been deleted.\n", tag)
		} else {
			fmt.Printf("No checkpoint found with tag '%s'.\n", tag)
		}
	},
}

// chatShareCmd represents the chat share subcommand
var chatShareCmd = &cobra.Command{
	Use:   "share [file]",
	Short: "Share the current conversation to a markdown or json file",
	Long:  `Share the current conversation to a markdown or json file. Usage: /chat share [file]`, //nolint:staticcheck
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePathArg := ""
		if len(args) > 0 {
			filePathArg = args[0]
		}

		if filePathArg == "" {
			filePathArg = fmt.Sprintf("gemini-conversation-%d.json", time.Now().Unix())
		}

		filePath := filePathArg
		extension := filepath.Ext(filePath)

		if extension != ".md" && extension != ".json" {
			fmt.Fprintf(os.Stderr, "Error: Invalid file format. Only .md and .json are supported.\n")
			os.Exit(1)
		}

		historyGenai, err := executor.GetHistory()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting chat history: %v\n", err)
			os.Exit(1)
		}

		var userInitiatedHistoryGenai []*genai.Content
		if len(historyGenai) > 2 {
			userInitiatedHistoryGenai = historyGenai[2:] // Skip the first two messages (system/context)
		}

		var history []types.Content
		for _, item := range userInitiatedHistoryGenai { // Iterate over user-initiated history
			var parts []types.Part
			for _, part := range item.Parts {
				newPart := types.Part{}
				switch p := part.(type) {
				case genai.Text:
					newPart.Text = string(p)
				case *genai.FunctionCall:
					newPart.FunctionCall = &types.FunctionCall{Name: p.Name, Args: p.Args}
				case *genai.FunctionResponse:
					newPart.FunctionResponse = &types.FunctionResponse{Name: p.Name, Response: p.Response}
				}
				parts = append(parts, newPart)
			}
			history = append(history, types.Content{Role: item.Role, Parts: parts})
		}

		if len(history) == 0 { // Now check if there's any actual conversation
			fmt.Println("No conversation found to share.")
			return
		}

		var content string
		if extension == ".json" {
			jsonContent, err := json.MarshalIndent(history, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error marshalling chat history to JSON: %v\n", err)
				os.Exit(1)
			}
			content = string(jsonContent)
		} else { // .md
			content = serializeHistoryToMarkdown(history)
		}

		err = os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing conversation to file: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Conversation shared to %s\n", filePath)
	},
}

// serializeHistoryToMarkdown converts a slice of types.Content to a Markdown string.
func serializeHistoryToMarkdown(history []types.Content) string {
	var sb strings.Builder

	for i, item := range history {
		if i > 0 {
			sb.WriteString("\n\n---\n\n")
		}

		var textParts []string
		for _, part := range item.Parts {
			if part.Text != "" {
				textParts = append(textParts, part.Text)
			} else if part.FunctionCall != nil {
				jsonCall, err := json.MarshalIndent(part.FunctionCall, "", "  ")
				if err == nil {
					textParts = append(textParts, fmt.Sprintf("**Tool Command**:\n```json\n%s\n```", string(jsonCall)))
				}
			} else if part.FunctionResponse != nil {
				jsonResponse, err := json.MarshalIndent(part.FunctionResponse, "", "  ")
				if err == nil {
					textParts = append(textParts, fmt.Sprintf("**Tool Response**:\n```json\n%s\n```", string(jsonResponse)))
				}
			}
		}

		text := strings.Join(textParts, "")
		roleIcon := "✨"
		if item.Role == "user" {
			roleIcon = "🧑‍💻"
		}

		sb.WriteString(fmt.Sprintf("%s ## %s\n\n%s", roleIcon, strings.ToUpper(item.Role), text))
	}

	return sb.String()
}
