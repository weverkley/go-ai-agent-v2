package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// statsCmd represents the stats command
var statsCmd = &cobra.Command{
	Use:     "stats",
	Aliases: []string{"usage"},
	Short:   "Check session stats",
	Long:    `The stats command displays various usage statistics for the current session, including overall duration and model/tool specific metrics.`, //nolint:staticcheck
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Session statistics are now shown at the end of an interactive 'chat' session.")
	},
}

var statsModelCmd = &cobra.Command{
	Use:   "model",
	Short: "Show model-specific usage statistics",
	Long:  `Displays statistics related to model usage during the session.`, //nolint:staticcheck
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Displaying model-specific usage statistics is not yet implemented. This feature may be available in a future version.")
	},
}

var statsToolsCmd = &cobra.Command{
	Use:   "tools",
	Short: "Show tool-specific usage statistics",
	Long:  `Displays statistics related to tool usage during the session.`, //nolint:staticcheck
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Displaying tool-specific usage statistics is not yet implemented. This feature may be available in a future version.")
	},
}

func init() {
	statsCmd.AddCommand(statsModelCmd)
	statsCmd.AddCommand(statsToolsCmd)
}
