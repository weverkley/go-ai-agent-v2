package cmd

import (
	"fmt"
	"os"

	"go-ai-agent-v2/go-cli/pkg/tools"
	"github.com/spf13/cobra"
)

var (
	readManyFilesPaths []string
	readManyFilesInclude []string
	readManyFilesExclude []string
	readManyFilesRecursive bool
	readManyFilesUseDefaultExcludes bool
	readManyFilesRespectGitIgnore bool
	readManyFilesRespectGeminiIgnore bool
)

func init() {
	rootCmd.AddCommand(readManyFilesCmd)
	readManyFilesCmd.Flags().StringArrayVarP(&readManyFilesPaths, "paths", "p", []string{}, "An array of file paths or directory paths to search within.")
	readManyFilesCmd.Flags().StringArrayVar(&readManyFilesInclude, "include", []string{}, "Optional: Glob patterns for files to include.")
	readManyFilesCmd.Flags().StringArrayVar(&readManyFilesExclude, "exclude", []string{}, "Optional: Glob patterns for files/directories to exclude.")
	readManyFilesCmd.Flags().BoolVar(&readManyFilesRecursive, "recursive", true, "Optional: Whether to search recursively.")
	readManyFilesCmd.Flags().BoolVar(&readManyFilesUseDefaultExcludes, "use-default-excludes", true, "Optional: Whether to apply default exclusion patterns.")
	readManyFilesCmd.Flags().BoolVar(&readManyFilesRespectGitIgnore, "respect-git-ignore", true, "Optional: Whether to respect .gitignore patterns.")
	readManyFilesCmd.Flags().BoolVar(&readManyFilesRespectGeminiIgnore, "respect-gemini-ignore", true, "Optional: Whether to respect .geminiignore patterns.")
	readManyFilesCmd.MarkFlagRequired("paths")
}

var readManyFilesCmd = &cobra.Command{
	Use:   "read-many-files",
	Short: "Reads content from multiple files",
	Long:  `Reads content from multiple files specified by paths or glob patterns within a configured target directory. For text files, it concatenates their content into a single string.`, 
	Run: func(cmd *cobra.Command, args []string) {
		readManyFilesTool := tools.NewReadManyFilesTool()
		result, err := readManyFilesTool.Execute(
			readManyFilesPaths,
			readManyFilesInclude,
			readManyFilesExclude,
			readManyFilesRecursive,
			readManyFilesUseDefaultExcludes,
			readManyFilesRespectGitIgnore,
			readManyFilesRespectGeminiIgnore,
		)
		if err != nil {
			fmt.Printf("Error executing read-many-files command: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(result)
	},
}
