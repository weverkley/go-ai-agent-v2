package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// memoryCmd represents the memory command group
var memoryCmd = &cobra.Command{
	Use:   "memory",
	Short: "Manage user memory",
	Long:  `The memory command group allows you to manage user-specific memories that the AI can access.`, //nolint:staticcheck
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand is provided, print help
		cmd.Help()
	},
}

func init() {
	memoryCmd.AddCommand(memoryGetCmd)
	memoryCmd.AddCommand(memorySetCmd)
	memoryCmd.AddCommand(memoryClearCmd)
}

var memoryGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get the current user memory",
	Run: func(cmd *cobra.Command, args []string) {
		memoryContent := MemoryService.GetMemory()
		if memoryContent == "" {
			fmt.Println("Memory is currently empty.")
		} else {
			fmt.Printf("Current memory content:\n%s\n", memoryContent)
		}
	},
}

var memorySetCmd = &cobra.Command{
	Use:   "set <memory_content>",
	Short: "Set the user memory",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		memoryContent := args[0]
		err := MemoryService.SetMemory(memoryContent)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error setting user memory: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("User memory set to: '%s'.\n", memoryContent)
	},
}

var memoryClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear the user memory",
	Run: func(cmd *cobra.Command, args []string) {
		err := MemoryService.ClearMemory()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error clearing user memory: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("User memory cleared.")
	},
}