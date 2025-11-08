package cmd

import (
	"fmt"
	"os"

	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/spf13/cobra"
)

var writeFilePath string
var writeContent string

func init() {
	writeCmd.Flags().StringVarP(&writeFilePath, "file", "f", "", "The path to the file to write")
	writeCmd.Flags().StringVarP(&writeContent, "content", "c", "", "The content to write to the file")
	_ = writeCmd.MarkFlagRequired("file")
	_ = writeCmd.MarkFlagRequired("content")
}

var writeCmd = &cobra.Command{
	Use:   "write",
	Short: "Write content to a file",
	Long:  `Write content to a specified file.`, 
	Run: func(cmd *cobra.Command, args []string) {
		tool, err := Cfg.GetToolRegistry().GetTool(types.WRITE_FILE_TOOL_NAME)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		toolArgs := map[string]any{
			"file_path": writeFilePath,
			"content":   writeContent,
		}

		result, err := tool.Execute(toolArgs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error executing write_file tool: %v\n", err)
			os.Exit(1)
		}

		if result.Error != nil {
			fmt.Fprintf(os.Stderr, "Tool execution error: %s\n", result.Error.Message)
			os.Exit(1)
		}

		fmt.Println(result.ReturnDisplay)
	},
}
