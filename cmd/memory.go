package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// memoryCmd represents the memory command group
var memoryCmd = &cobra.Command{
	Use:   "memory",
	Short: "Manage contextual memory files (GOAIAGENT.md)",
	Long:  `The memory command group allows you to manage the hierarchical GOAIAGENT.md context files.`, 
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	memoryCmd.AddCommand(memoryShowCmd)
	memoryCmd.AddCommand(memoryRefreshCmd)
	memoryCmd.AddCommand(memoryAddCmd)
	memoryCmd.AddCommand(memoryListCmd)
}

var memoryShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the full, concatenated context from all GOAIAGENT.md files",
	Run: func(cmd *cobra.Command, args []string) {
		contextContent := ContextService.ShowMemory()
		if contextContent == "" {
			fmt.Println("No context found.")
		} else {
			fmt.Printf("Current context:\n%s\n", contextContent)
		}
	},
}

var memoryRefreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Force a re-scan and reload of all GOAIAGENT.md files",
	Run: func(cmd *cobra.Command, args []string) {
		err := ContextService.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error refreshing context: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Context has been refreshed.")
	},
}

var memoryAddCmd = &cobra.Command{
	Use:   "add <text>",
	Short: "Append text to the global GOAIAGENT.md file",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		textToAdd := strings.Join(args, " ")
		err := ContextService.AddToGlobalMemory(textToAdd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error adding to global memory: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Added to global memory: '%s'.\n", textToAdd)
	},
}

var memoryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the paths of the GOAIAGENT.md files in use",
	Run: func(cmd *cobra.Command, args []string) {
		files := ContextService.ListMemoryFiles()
		if len(files) == 0 {
			fmt.Println("No GOAIAGENT.md files in use.")
		} else {
			fmt.Printf("There are %d GOAIAGENT.md file(s) in use:\n\n%s\n", len(files), strings.Join(files, "\n"))
		}
	},
}
