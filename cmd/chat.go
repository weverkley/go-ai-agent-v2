package cmd

import (
	"bytes" // New import for bytes.Buffer
	"fmt"
	"os"
	"strings" // New import for strings.TrimSpace
	"time"    // New import for time

	"go-ai-agent-v2/go-cli/pkg/services"
	"go-ai-agent-v2/go-cli/pkg/types"
	"go-ai-agent-v2/go-cli/pkg/ui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

// CommandExecutor is a function type that executes a command and returns its output and an error.
type CommandExecutor func(args []string) (string, error)

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Start an interactive chat session with the AI agent",
	Run: func(cmd *cobra.Command, args []string) {
		runChatCmd(RootCmd, cmd, args, SettingsService, ShellService)
	},
}

// runChatCmd contains the logic for the chat command, accepting necessary services.
func runChatCmd(rootCmd *cobra.Command, cmd *cobra.Command, args []string, settingsService *services.SettingsService, shellService services.ShellExecutionService) {
	// Use the global Cfg and initialized executor
	appConfig := Cfg
	executorTypeVal, _ := settingsService.Get("executor")
	executorType, _ := executorTypeVal.(string)

	toolRegistry, _ := appConfig.Get("toolRegistry")
	chatService := services.NewChatService(executor, toolRegistry.(types.ToolRegistryInterface), []*types.Content{})

	// Create a CommandExecutor function that wraps the Cobra command execution
	commandExecutor := func(cmdArgs []string) (string, error) {
		// Create a buffer to capture stdout and stderr
		var buffer bytes.Buffer
		rootCmd.SetOut(&buffer)
		rootCmd.SetErr(&buffer)

		// Find the command
		cmd, _, err := rootCmd.Find(cmdArgs)
		if err != nil {
			return "", err
		}

		// Execute the Cobra command
		rootCmd.SetArgs(cmdArgs)
		err = cmd.Execute()

		// Restore default output
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)

		output := strings.TrimSpace(buffer.String())
		return output, err
	}

	// Initialize GitService
	gitService := services.NewGitService()

	// Initialize WorkspaceService
	workspaceService := services.NewWorkspaceService(".") // Assuming "." is the project root

	p := tea.NewProgram(ui.NewChatModel(chatService, executorType, appConfig, commandExecutor, shellService, gitService, workspaceService), tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("Error running interactive chat: %v\n", err)
		os.Exit(1)
	}

	// Print final stats after the UI exits
	if m, ok := finalModel.(*ui.ChatModel); ok {
		toolCount, errorCount, duration := m.GetStats()
		fmt.Printf("\n\nSession ended.\n")
		fmt.Printf("Total tool calls: %d\n", toolCount)
		fmt.Printf("Failed tool calls: %d\n", errorCount)
		fmt.Printf("Total session time: %s\n", duration.Round(time.Second))
	}
}