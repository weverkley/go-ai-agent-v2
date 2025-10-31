package cmd

import (
	"fmt"
	"os"

	"go-ai-agent-v2/go-cli/pkg/tools"
	"github.com/spf13/cobra"
)

var (
	globPattern string
	globPath string
	globCaseSensitive bool
	globRespectGitIgnore bool
	globRespectGeminiIgnore bool
)

func init() {
	rootCmd.AddCommand(globCmd)
	globCmd.Flags().StringVarP(&globPattern, "pattern", "p", "", "The glob pattern to match against.")
	globCmd.Flags().StringVar(&globPath, "path", ".", "The absolute path to the directory to search within.")
	globCmd.Flags().BoolVar(&globCaseSensitive, "case-sensitive", false, "Whether the search should be case-sensitive.")
	globCmd.Flags().BoolVar(&globRespectGitIgnore, "respect-git-ignore", true, "Whether to respect .gitignore patterns.")
	globCmd.Flags().BoolVar(&globRespectGeminiIgnore, "respect-gemini-ignore", true, "Whether to respect .geminiignore patterns.")
	_ = globCmd.MarkFlagRequired("pattern")
}

var globCmd = &cobra.Command{
	Use:   "glob",
	Short: "Efficiently finds files matching specific glob patterns",
	Long:  `Efficiently finds files matching specific glob patterns (e.g., src/**/*.ts, **/*.md), returning absolute paths sorted by modification time (newest first). Ideal for quickly locating files based on their name or path structure, especially in large codebases.`, 
	Run: func(cmd *cobra.Command, args []string) {
		globTool := tools.NewGlobTool()
		result, err := globTool.Execute(map[string]any{
			"pattern":               globPattern,
			"path":                  globPath,
			"case_sensitive":        globCaseSensitive,
			"respect_git_ignore":    globRespectGitIgnore,
			"respect_gemini_ignore": globRespectGeminiIgnore,
		})
		if err != nil {
			fmt.Printf("Error executing glob command: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(result)
	},
}
