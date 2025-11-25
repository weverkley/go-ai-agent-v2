package cmd

import (
	"context" // Add context import
	"fmt"
	"os"

	"go-ai-agent-v2/go-cli/pkg/services"
	"go-ai-agent-v2/go-cli/pkg/tools"

	"github.com/spf13/cobra"
)

var (
	smartEditFilePath    string
	smartEditInstruction string
	smartEditOldString   string
	smartEditNewString   string
)

func init() {
	smartEditCmd.Flags().StringVarP(&smartEditFilePath, "file-path", "f", "", "The absolute path to the file to modify.")
	smartEditCmd.Flags().StringVarP(&smartEditInstruction, "instruction", "i", "", "A clear, semantic instruction for the code change.")
	smartEditCmd.Flags().StringVarP(&smartEditOldString, "old-string", "o", "", "The exact literal text to replace.")
	smartEditCmd.Flags().StringVarP(&smartEditNewString, "new-string", "n", "", "The exact literal text to replace old_string with.")
	_ = smartEditCmd.MarkFlagRequired("file-path")
	_ = smartEditCmd.MarkFlagRequired("instruction")
	_ = smartEditCmd.MarkFlagRequired("old-string")
	_ = smartEditCmd.MarkFlagRequired("new-string")
}

var smartEditCmd = &cobra.Command{
	Use:   "smart-edit",
	Short: "Replaces text within a file using smart strategies",
	Long:  `Replaces text within a file using smart strategies (exact, flexible, regex) and includes self-correction logic.`,
	Run: func(cmd *cobra.Command, args []string) {
		fileSystemService := services.NewFileSystemService()
		smartEditTool := tools.NewSmartEditTool(fileSystemService)
		result, err := smartEditTool.Execute(context.Background(), map[string]any{
			"file_path":   smartEditFilePath,
			"instruction": smartEditInstruction,
			"old_string":  smartEditOldString,
			"new_string":  smartEditNewString,
		})
		if err != nil {
			fmt.Printf("Error executing smart-edit command: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(result)
	},
}
