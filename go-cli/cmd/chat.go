package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"go-ai-agent-v2/go-cli/pkg/services"

	"github.com/spf13/cobra"
)

var chatService *services.ChatService

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
	chatService = services.NewChatService(cfg) // cfg is from root.go

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
			// TODO: Implement interactive overwrite confirmation using bubbletea
			// For now, just print a message and exit.
			fmt.Fprintf(os.Stderr, "Error: Checkpoint with tag '%s' already exists. Please use a different tag or delete the existing one.\n", tag)
			os.Exit(1)
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
		fmt.Printf("Sharing chat with tag: %s (Not yet implemented)\n", args[0])
		// TODO: Implement chat share logic (e.g., generate a shareable URL or format)
		// nolint:staticcheck
	},
}
