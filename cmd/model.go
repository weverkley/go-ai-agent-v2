package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// modelCmd represents the model command group
var modelCmd = &cobra.Command{
	Use:   "model",
	Short: "Manage AI models",
	Long:  `The model command group allows you to list, set, and get information about AI models.`, //nolint:staticcheck
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand is provided, print help
		cmd.Help()
	},
}

func init() {
	modelCmd.AddCommand(modelListCmd)
}

var modelListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available AI models",
	Long:  `List all available AI models that can be used with the generate command.`, //nolint:staticcheck
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(os.Stderr, "Error: Model list command is not yet functional after refactoring.\n")
		os.Exit(1)
		// models, err := executor.ListModels()
		// if err != nil {
		// 	fmt.Fprintf(os.Stderr, "Error listing models: %v\n", err)
		// 	os.Exit(1)
		// }
		// fmt.Println("Available AI Models:")
		// for _, model := range models {
		// 	fmt.Printf("- %s\n", model)
		// }
	},
}

