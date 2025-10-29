package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "go-cli",
	Short: "A Go-based CLI for Gemini",
	Long:  `A Go-based CLI for interacting with the Gemini API and managing extensions.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(readCmd)
	rootCmd.AddCommand(writeCmd)
	rootCmd.AddCommand(execCmd)
	rootCmd.AddCommand(lsCmd)
	rootCmd.AddCommand(gitBranchCmd)
	rootCmd.AddCommand(extensionsCmd)
	rootCmd.AddCommand(mcpCmd)
	rootCmd.AddCommand(listModelsCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(globCmd)
	rootCmd.AddCommand(grepCmd)
	rootCmd.AddCommand(webFetchCmd)
	rootCmd.AddCommand(memoryCmd)
	rootCmd.AddCommand(webSearchCmd)
	rootCmd.AddCommand(readManyFilesCmd)
	rootCmd.AddCommand(readFileCmd)
}
