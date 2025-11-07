package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Gemini CLI project",
	Long:  `The init command initializes a new Gemini CLI project in the current directory, creating necessary configuration files.`, //nolint:staticcheck
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement actual project initialization (e.g., create .gemini directory, settings.json).
		projectRoot, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error getting current directory: %v\n", err)
			return
		}
		geminiDir := filepath.Join(projectRoot, ".gemini")

		// Check if .gemini directory already exists
		if _, err := os.Stat(geminiDir); !os.IsNotExist(err) {
			fmt.Printf("Gemini CLI project already initialized in %s.\n", projectRoot)
			return
		}

		// Create .gemini directory
		err = os.Mkdir(geminiDir, 0755)
		if err != nil {
			fmt.Printf("Error creating .gemini directory: %v\n", err)
			return
		}

		fmt.Printf("Gemini CLI project initialized in %s.\n", projectRoot)
	},
}
