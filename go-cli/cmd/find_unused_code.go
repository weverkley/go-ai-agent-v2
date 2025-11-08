package cmd

import (
	"fmt"
	"os"

	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/spf13/cobra"
)

var findUnusedCodeCmd = &cobra.Command{
	Use:   "find-unused-code",
	Short: "Finds unused functions and methods in Go files.",
	Long: `The find-unused-code command analyzes Go files in a specified directory
to identify functions and methods that are not called or referenced within the project.`,
	Run: func(cmd *cobra.Command, args []string) {
		directory, _ := cmd.Flags().GetString("directory")

		if directory == "" {
			fmt.Fprintf(os.Stderr, "Error: --directory is required.\n")
			_ = cmd.Help()
			os.Exit(1)
		}

		tool, err := Cfg.GetToolRegistry().GetTool(types.FIND_UNUSED_CODE_TOOL_NAME)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		toolArgs := map[string]any{
			"directory": directory,
		}

		result, err := tool.Execute(toolArgs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error executing find-unused-code tool: %v\n", err)
			os.Exit(1)
		}

		if result.Error != nil {
			fmt.Fprintf(os.Stderr, "Tool execution error: %s\n", result.Error.Message)
			os.Exit(1)
		}

		fmt.Println(result.ReturnDisplay)
	},
}

func init() {
	rootCmd.AddCommand(findUnusedCodeCmd)

	findUnusedCodeCmd.Flags().StringP("directory", "d", "", "The absolute path to the directory to analyze.")
}
