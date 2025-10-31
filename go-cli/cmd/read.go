package cmd

import (
	"fmt"
	"os"

	"go-ai-agent-v2/go-cli/pkg/services"
	"github.com/spf13/cobra"
)

var readFilePath string

func init() {
	rootCmd.AddCommand(readCmd)
	readCmd.Flags().StringVarP(&readFilePath, "file", "f", "", "The path to the file to read")
	_ = readCmd.MarkFlagRequired("file")
}

var readCmd = &cobra.Command{
	Use:   "read",
	Short: "Read content from a file",
	Long:  `Read the content of a specified file.`, 
	Run: func(cmd *cobra.Command, args []string) {
		fsService := services.NewFileSystemService()
		content, err := fsService.ReadFile(readFilePath)
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(content)
	},
}
