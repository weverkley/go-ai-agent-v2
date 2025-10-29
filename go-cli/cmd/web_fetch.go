package cmd

import (
	"fmt"
	"os"

	"go-ai-agent-v2/go-cli/pkg/tools"
	"github.com/spf13/cobra"
)

var (
	webFetchUrls []string
	webFetchSummarize bool
	webFetchExtractKeyPoints bool
)

func init() {
	rootCmd.AddCommand(webFetchCmd)
	webFetchCmd.Flags().StringArrayVarP(&webFetchUrls, "url", "u", []string{}, "The URL(s) to fetch content from.")
	webFetchCmd.Flags().BoolVar(&webFetchSummarize, "summarize", false, "Whether to summarize the fetched content.")
	webFetchCmd.Flags().BoolVar(&webFetchExtractKeyPoints, "extract-key-points", false, "Whether to extract key points from the fetched content.")
	webFetchCmd.MarkFlagRequired("url")
}

var webFetchCmd = &cobra.Command{
	Use:   "web-fetch",
	Short: "Fetches content from URL(s)",
	Long:  `Fetches content from URL(s), including local and private network addresses (e.g., localhost), embedded in a prompt.`, 
	Run: func(cmd *cobra.Command, args []string) {
		webFetchTool := tools.NewWebFetchTool()
		result, err := webFetchTool.Execute(
			webFetchUrls,
			webFetchSummarize,
			webFetchExtractKeyPoints,
		)
		if err != nil {
			fmt.Printf("Error executing web-fetch command: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(result)
	},
}
