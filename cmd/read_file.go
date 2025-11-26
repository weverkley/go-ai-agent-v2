package cmd

import (
	"context" // New import
	"fmt"
	"os"

	"go-ai-agent-v2/go-cli/pkg/services"
	"go-ai-agent-v2/go-cli/pkg/tools"

	"github.com/spf13/cobra"
)

var (
	readFileAbsolutePath string
	readFileOffset       int
	readFileLimit        int
)

func init() {
	readFileCmd.Flags().StringVarP(&readFileAbsolutePath, "file_path", "f", "", "The path to the file to read, relative to the project root.")
	readFileCmd.Flags().IntVar(&readFileOffset, "offset", 0, "Optional: For text files, the 0-based line number to start reading from.")
	readFileCmd.Flags().IntVar(&readFileLimit, "limit", 0, "Optional: For text files, maximum number of lines to read.")
	_ = readFileCmd.MarkFlagRequired("file_path")
}

var readFileCmd = &cobra.Command{
	Use:   "read-file",
	Short: "Reads and returns the content of a specified file.",
	Long:  `Reads and returns the content of a specified file. If the file is large, the content will be truncated. The tool's response will clearly indicate if truncation has occurred and will provide details on how to read more of the file using the 'offset' and 'limit' parameters. Handles text, images (PNG, JPG, GIF, WEBP, SVG, BMP), and PDF files. For text files, it can read specific line ranges.`,
	Run: func(cmd *cobra.Command, args []string) {
		workspaceService := services.NewWorkspaceService(".")
		tool := tools.NewReadFileTool(workspaceService)
		result, err := tool.Execute(context.Background(), map[string]any{
			"file_path": readFileAbsolutePath,
			"offset":    readFileOffset,
			"limit":     readFileLimit,
		})
		if err != nil {
			fmt.Printf("Error executing read-file command: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(result)
	},
}
