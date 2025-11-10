package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"go-ai-agent-v2/go-cli/pkg/services"
	"github.com/spf13/cobra"
)

var cleanupBackToMainCmd = &cobra.Command{
	Use:   "cleanup-back-to-main",
	Short: "Go back to main and clean up the branch.",
	Long:  `This command automates the process of going back to the main branch, pulling the latest changes, and deleting the feature branch.`, 
	Run: func(cmd *cobra.Command, args []string) {
		gitService := services.NewGitService()
		workspaceDir, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error getting current working directory: %v\n", err)
			os.Exit(1)
		}
		projectRoot := filepath.Dir(workspaceDir)

		// 1. Get Current Branch
		currentBranch, err := gitService.GetCurrentBranch(projectRoot)
		if err != nil {
			fmt.Printf("Error getting current branch: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Current branch: %s\n", currentBranch)

		// 2. Branch Check
		if currentBranch == "main" || currentBranch == "master" {
			fmt.Println("Already on main/master branch. Aborting cleanup.")
			os.Exit(0)
		}

		// 3. Go to Main
		fmt.Printf("Checking out main branch...\n")
		err = gitService.CheckoutBranch(projectRoot, "main")
		if err != nil {
			fmt.Printf("Error checking out main branch: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Successfully checked out main branch.")

		// 4. Pull Latest
		fmt.Printf("Pulling latest changes on main branch...\n")
		err = gitService.Pull(projectRoot, "")
		if err != nil {
			fmt.Printf("Error pulling latest changes: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Successfully pulled latest changes.")

		// 5. Branch Cleanup
		fmt.Printf("Deleting branch %s...\n", currentBranch)
		err = gitService.DeleteBranch(projectRoot, currentBranch)
		if err != nil {
			fmt.Printf("Error deleting branch %s: %v\n", currentBranch, err)
			os.Exit(1)
		}
		fmt.Printf("Successfully deleted branch %s.\n", currentBranch)

		fmt.Println("Cleanup complete.")
	},
}

func init() {
}
