package cmd

import (
	"bytes" // New import for bytes.Buffer
	"fmt"
	"os"
	"strings" // New import for strings.TrimSpace
	"time"    // New import for time

	"go-ai-agent-v2/go-cli/pkg/core" // New import for core package
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

func init() {
	chatCmd.Flags().String("session-id", "", "Specify an existing session ID to resume or create a new one with this ID")
	chatCmd.Flags().Bool("new-session", false, "Force creation of a new session (generates a new ID)")
	chatCmd.Flags().Bool("list-sessions", false, "List all available sessions and exit")
	chatCmd.Flags().String("delete-session", "", "Delete a specific session by ID and exit")
	chatCmd.Flags().Bool("latest", false, "Resume the latest session")
}

// runChatCmd contains the logic for the chat command, accepting necessary services.
func runChatCmd(rootCmd *cobra.Command, cmd *cobra.Command, args []string, settingsService types.SettingsServiceIface, shellService services.ShellExecutionService) {
	// Parse session flags
	listSessions, _ := cmd.Flags().GetBool("list-sessions")
	deleteSessionID, _ := cmd.Flags().GetString("delete-session")
	newSessionFlag, _ := cmd.Flags().GetBool("new-session")
	sessionIDFlag, _ := cmd.Flags().GetString("session-id")
	latestSessionFlag, _ := cmd.Flags().GetBool("latest")

	// Handle --list-sessions
	if listSessions {
		sessions, err := SessionService.ListSessions()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing sessions: %v\n", err)
			os.Exit(1)
		}
		if len(sessions) == 0 {
			fmt.Println("No chat sessions found.")
			os.Exit(0)
		}
		fmt.Println("Available chat sessions (newest first):")
		for i, id := range sessions {
			fmt.Printf("  %d: %s\n", i+1, id)
		}
		os.Exit(0)
	}

	// Handle --delete-session
	if deleteSessionID != "" {
		err := SessionService.DeleteSession(deleteSessionID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error deleting session '%s': %v\n", deleteSessionID, err)
			os.Exit(1)
		}
		fmt.Printf("Session '%s' deleted successfully.\n", deleteSessionID)
		os.Exit(0)
	}

	// Determine the current session ID
	var currentSessionID string
	if newSessionFlag {
		currentSessionID = SessionService.GenerateSessionID()
	} else if sessionIDFlag != "" {
		currentSessionID = sessionIDFlag
	} else if latestSessionFlag {
		sessions, err := SessionService.ListSessions()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting latest session: %v\n", err)
			os.Exit(1)
		}
		if len(sessions) > 0 {
			currentSessionID = sessions[0] // ListSessions returns newest first
		} else {
			fmt.Println("No previous sessions found. Starting a new session.")
			currentSessionID = SessionService.GenerateSessionID()
		}
	} else {
		// Default behavior: create a new session if no flags are specified
		currentSessionID = SessionService.GenerateSessionID()
	}

	// Use the global Cfg
	appConfig := Cfg
	executorTypeVal, _ := settingsService.Get("executor")
	executorType, _ := executorTypeVal.(string)

	// Create ExecutorFactory
	executorFactory, err := core.NewExecutorFactory(executorType, appConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating executor factory: %v\n", err)
		os.Exit(1)
	}

	// Create Executor
	// generationConfig and startHistory are not directly available here in the current flow
	// For now, we pass empty or default values. These will be properly managed in Phase 2.
	executor, err := executorFactory.NewExecutor(appConfig, types.GenerateContentConfig{}, []*types.Content{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating executor: %v\n", err)
		os.Exit(1)
	}

	toolRegistry, _ := appConfig.Get("toolRegistry")
	chatService, err := services.NewChatService(executor, toolRegistry.(types.ToolRegistryInterface), SessionService, currentSessionID, SettingsService, appConfig, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating chat service: %v\n", err)
		os.Exit(1)
	}

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

	p := tea.NewProgram(ui.NewChatModel(chatService, SessionService, executorType, appConfig, commandExecutor, shellService, gitService, workspaceService, currentSessionID), tea.WithAltScreen())
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