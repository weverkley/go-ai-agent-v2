package cmd

import (
	"fmt"
	"os"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/services"

	"github.com/spf13/cobra"
)

var lsPath string

func init() {
	lsCmd.Flags().StringVarP(&lsPath, "path", "p", ".", "The path to the directory to list")
}

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List directory contents",
	Long:  `List the contents of a specified directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		fsService := services.NewFileSystemService()

		entries, err := fsService.ListDirectory(lsPath, []string{}, true, true)
		if err != nil {
			fmt.Printf("Error listing directory: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(strings.Join(entries, "\n"))
	},
}
