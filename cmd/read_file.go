package cmd

import (
	"context" // New import
	"fmt"
	"os"

	"go-ai-agent-v2/go-cli/pkg/tools"
	"github.com/spf13/cobra"
)

var (
	readFileAbsolutePath string
	readFileOffset int
	readFileLimit int
)

func init() {
	readFileCmd.Flags().StringVarP(&readFileAbsolutePath, "absolute-path", "a", "", "The absolute path to the file to read.")
	readFileCmd.Flags().IntVar(&readFileOffset, "offset", 0, "Optional: For text files, the 0-based line number to start reading from.")
	readFileCmd.Flags().IntVar(&readFileLimit, "limit", 0, "Optional: For text files, maximum number of lines to read.")
	_ = readFileCmd.MarkFlagRequired("absolute-path")
}

var readFileCmd = &cobra.Command{
	Use:   "read-file",
	Short: "Reads and returns the content of a specified file.",
	Long:  `Reads and returns the content of a specified file. If the file is large, the content will be truncated. The tool's response will clearly indicate if truncation has occurred and will provide details on how to read more of the file using the 'offset' and 'limit' parameters. Handles text, images (PNG, JPG, GIF, WEBP, SVG, BMP), and PDF files. For text files, it can read specific line ranges.`,
	Run: func(cmd *cobra.Command, args []string) {
		readFileTool := tools.NewReadFileTool()
		result, err := readFileTool.Execute(context.Background(), map[string]any{
			"absolute_path": readFileAbsolutePath,
			"offset":        readFileOffset,
			"limit":         readFileLimit,
		})
		if err != nil {
			fmt.Printf("Error executing read-file command: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(result)
	},
}
