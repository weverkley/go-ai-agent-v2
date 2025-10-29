package cmd

import (
	"fmt"
	"os"

	"go-ai-agent-v2/go-cli/pkg/services"
	"github.com/spf13/cobra"
)

var writeFilePath string
var writeContent string

func init() {
	rootCmd.AddCommand(writeCmd)
	writeCmd.Flags().StringVarP(&writeFilePath, "file", "f", "", "The path to the file to write")
	writeCmd.Flags().StringVarP(&writeContent, "content", "c", "", "The content to write to the file")
	writeCmd.MarkFlagRequired("file")
	writeCmd.MarkFlagRequired("content")
}

var writeCmd = &cobra.Command{
	Use:   "write",
	Short: "Write content to a file",
	Long:  `Write content to a specified file.`, 
	Run: func(cmd *cobra.Command, args []string) {
		fsService := services.NewFileSystemService()
		err := fsService.WriteFile(writeFilePath, writeContent)
		if err != nil {
			fmt.Printf("Error writing file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Successfully wrote to %s\n", writeFilePath)
	},
}
