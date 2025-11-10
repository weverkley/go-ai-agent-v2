package cmd

import (
	"fmt"
	"os"

	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/spf13/cobra"
)

var extractFunctionCmd = &cobra.Command{
	Use:   "extract-function",
	Short: "Extracts a code block into a new function or method.",
	Long: `The extract-function command refactors a specified code block into a new function or method.
It requires the file path, start and end line numbers of the code block, and the desired new function name.
Optionally, a receiver can be specified to create a method instead of a standalone function.`,
	Run: func(cmd *cobra.Command, args []string) {
		filePath, _ := cmd.Flags().GetString("file-path")
		startLine, _ := cmd.Flags().GetInt("start-line")
		endLine, _ := cmd.Flags().GetInt("end-line")
		newFunctionName, _ := cmd.Flags().GetString("new-function-name")
		receiver, _ := cmd.Flags().GetString("receiver")

		if filePath == "" || newFunctionName == "" || startLine == 0 || endLine == 0 {
			fmt.Fprintf(os.Stderr, "Error: --file-path, --start-line, --end-line, and --new-function-name are required.\n")
			_ = cmd.Help()
			os.Exit(1)
		}

		toolRegistryVal, found := Cfg.Get("toolRegistry")
		if !found || toolRegistryVal == nil {
			fmt.Fprintf(os.Stderr, "Error: Tool registry not found in config.\n")
			os.Exit(1)
		}
		toolRegistry, ok := toolRegistryVal.(types.ToolRegistryInterface)
		if !ok {
			fmt.Fprintf(os.Stderr, "Error: Tool registry in config is not of expected type.\n")
			os.Exit(1)
		}

		tool, err := toolRegistry.GetTool(types.EXTRACT_FUNCTION_TOOL_NAME)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		toolArgs := map[string]any{
			"file_path":         filePath,
			"start_line":        float64(startLine),
			"end_line":          float64(endLine),
			"new_function_name": newFunctionName,
		}
		if receiver != "" {
			toolArgs["receiver"] = receiver
		}

		result, err := tool.Execute(toolArgs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error executing extract-function tool: %v\n", err)
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
	RootCmd.AddCommand(extractFunctionCmd)

	extractFunctionCmd.Flags().StringP("file-path", "f", "", "The absolute path to the file to modify.")
	extractFunctionCmd.Flags().IntP("start-line", "s", 0, "The 1-based starting line number of the code block to extract.")
	extractFunctionCmd.Flags().IntP("end-line", "e", 0, "The 1-based ending line number of the code block to extract.")
	extractFunctionCmd.Flags().StringP("new-function-name", "n", "", "The desired name for the new function or method.")
	extractFunctionCmd.Flags().StringP("receiver", "r", "", "Optional: The receiver for the new method (e.g., 's *MyStruct').")
}
