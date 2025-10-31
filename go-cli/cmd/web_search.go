package cmd

import (
	"fmt"
	"os"

	"go-ai-agent-v2/go-cli/pkg/tools"
	"github.com/spf13/cobra"
)

var (
	webSearchQuery string
)

func init() {
	rootCmd.AddCommand(webSearchCmd)
	webSearchCmd.Flags().StringVarP(&webSearchQuery, "query", "q", "", "The search query to find information on the web.")
	_ = webSearchCmd.MarkFlagRequired("query")
}

var webSearchCmd = &cobra.Command{
	Use:   "web-search",
	Short: "Performs a web search using Google Search",
	Long:  `Performs a web search using Google Search (via the Gemini API) and returns the results. This tool is useful for finding information on the internet based on a query.`, 
	Run: func(cmd *cobra.Command, args []string) {
		webSearchTool := tools.NewWebSearchTool()
		result, err := webSearchTool.Execute(map[string]any{
			"query": webSearchQuery,
		})
		if err != nil {
			fmt.Printf("Error executing web-search command: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(result)
	},
}
