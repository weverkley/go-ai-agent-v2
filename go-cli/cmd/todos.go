package cmd

import (
	"fmt"
	"go-ai-agent-v2/go-cli/pkg/tools"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	todosList []string
)

func init() {
	rootCmd.AddCommand(todosCmd)
	todosCmd.Flags().StringArrayVarP(&todosList, "list", "l", []string{}, "The full list of todos, each in 'description:status' format.")
	todosCmd.MarkFlagRequired("list")
}

var todosCmd = &cobra.Command{
	Use:   "todos",
	Short: "Manage a list of subtasks (todos)",
	Long:  `This tool helps you list out the current subtasks that are required to be completed for a given user request.`,
	Run: func(cmd *cobra.Command, args []string) {
		var todos []tools.Todo
		for _, item := range todosList {
			parts := strings.SplitN(item, ":", 2)
			if len(parts) != 2 {
				fmt.Printf("Invalid todo format: %s. Expected 'description:status'\n", item)
				os.Exit(1)
			}
			todos = append(todos, tools.Todo{Description: parts[0], Status: parts[1]})
		}

		var todosForExecute []any
		for _, t := range todos {
			todosForExecute = append(todosForExecute, map[string]any{
				"description": t.Description,
				"status":      t.Status,
			})
		}

		writeTodosTool := tools.NewWriteTodosTool()
		result, err := writeTodosTool.Execute(map[string]any{
			"todos": todosForExecute,
		})
		if err != nil {
			fmt.Printf("Error executing todos command: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(result)
	},
}
