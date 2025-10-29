package cmd

import (
	"fmt"
	"os"

	"go-ai-agent-v2/go-cli/pkg/tools"
	"github.com/spf13/cobra"
)

var (
	grepPattern string
	grepPath string
	grepInclude string
)

func init() {
	rootCmd.AddCommand(grepCmd)
	grepCmd.Flags().StringVarP(&grepPattern, "pattern", "p", "", "The regular expression (regex) pattern to search for.")
	grepCmd.Flags().StringVar(&grepPath, "path", ".", "Optional: The absolute path to the directory to search within.")
	grepCmd.Flags().StringVar(&grepInclude, "include", "", "Optional: A glob pattern to filter which files are searched.")
	grepCmd.MarkFlagRequired("pattern")
}

var grepCmd = &cobra.Command{
	Use:   "grep",
	Short: "Searches for a regular expression pattern within file contents",
	Long:  `Searches for a regular expression pattern within the content of files in a specified directory (or current working directory). Can filter files by a glob pattern. Returns the lines containing matches, along with their file paths and line numbers.`,
	Run: func(cmd *cobra.Command, args []string) {
		grepTool := tools.NewGrepTool()
		result, err := grepTool.Execute(
			grepPattern,
			grepPath,
			grepInclude,
		)
		if err != nil {
			fmt.Printf("Error executing grep command: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(result)
	},
}
