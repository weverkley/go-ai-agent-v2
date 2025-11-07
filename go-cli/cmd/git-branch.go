package cmd

import (
	"fmt"
	"os"

	"go-ai-agent-v2/go-cli/pkg/services"
	"github.com/spf13/cobra"
)

var gitBranchPath string

func init() {
	gitBranchCmd.Flags().StringVarP(&gitBranchPath, "path", "p", ".", "The path to the Git repository")
}

var gitBranchCmd = &cobra.Command{
	Use:   "git-branch",
	Short: "Get the current Git branch name",
	Long:  `Get the current Git branch name of a specified repository.`,
	Run: func(cmd *cobra.Command, args []string) {
		gitService := services.NewGitService()
		branch, err := gitService.GetCurrentBranch(gitBranchPath)
		if err != nil {
			fmt.Printf("Error getting git branch: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(branch)
	},
}
