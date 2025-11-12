package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/google/generative-ai-go/genai"

	"github.com/spf13/cobra"
)

// copyCmd represents the copy command
var copyCmd = &cobra.Command{
	Use:   "copy",
	Short: "Copy the last result or code snippet to clipboard",
	Long:  `The copy command copies the last AI generated output or code snippet to the system clipboard.`, //nolint:staticcheck
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		history, err := executor.GetHistory()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting chat history: %v\n", err)
			os.Exit(1)
		}

		var lastAiMessage string
		for i := len(history) - 1; i >= 0; i-- {
			item := history[i]
			if item.Role == "model" {
				for _, part := range item.Parts {
					if text, ok := part.(genai.Text); ok {
						lastAiMessage = string(text)
						break
					}
				}
			}
			if lastAiMessage != "" {
				break
			}
		}

		if lastAiMessage == "" {
			fmt.Println("No AI output found in history to copy.")
			return
		}

		err = copyToClipboard(lastAiMessage)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to copy to clipboard: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Last AI output copied to the clipboard.")
	},
}

// copyToClipboard copies the given text to the system clipboard.
func copyToClipboard(text string) error {
	var cmd *exec.Cmd
	fmt.Printf("DEBUG: OS is %s\n", runtime.GOOS)
	fmt.Printf("DEBUG: Text to copy: %s\n", text)

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "windows":
		cmd = exec.Command("cmd", "/c", "clip")
	case "linux":
		cmd = exec.Command("xclip", "-selection", "clipboard")
	default:
		return fmt.Errorf("unsupported operating system for clipboard operations: %s", runtime.GOOS)
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start clipboard command: %w", err)
	}

	_, err = stdin.Write([]byte(text))
	if err != nil {
		return fmt.Errorf("failed to write to clipboard: %w", err)
	}

	err = stdin.Close()
	if err != nil {
		return fmt.Errorf("failed to close stdin pipe: %w", err)
	}

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("clipboard command failed: %w", err)
	}
	fmt.Println("DEBUG: copyToClipboard finished successfully.")
	return nil
}
