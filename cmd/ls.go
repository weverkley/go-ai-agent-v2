package cmd

import (
	"fmt"
	"os"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/services"
	"go-ai-agent-v2/go-cli/pkg/ui" // Import the ui package

	"github.com/spf13/cobra"
	tea "github.com/charmbracelet/bubbletea" // Import bubbletea
)

var lsPath string
var interactive bool // New flag for interactive mode

func init() {
	lsCmd.Flags().StringVarP(&lsPath, "path", "p", ".", "The path to the directory to list")
	lsCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Run in interactive mode")
}

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List directory contents",
	Long:  `List the contents of a specified directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		fsService := services.NewFileSystemService()

		if interactive {
			// Launch interactive UI
			initialPath := lsPath
			if initialPath == "." {
				wd, err := os.Getwd()
				if err != nil {
					fmt.Printf("Error getting current working directory: %v\n", err)
					os.Exit(1)
				}
				initialPath = wd
			}

			p := tea.NewProgram(ui.NewLsModel(fsService, initialPath))
			if _, err := p.Run(); err != nil {
				fmt.Printf("Error running interactive ls: %v\n", err)
				os.Exit(1)
			}
		} else {
			// Existing non-interactive logic
			entries, err := fsService.ListDirectory(lsPath, []string{}, true, true)
			if err != nil {
				fmt.Printf("Error listing directory: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(strings.Join(entries, "\n"))
		}
	},
}
