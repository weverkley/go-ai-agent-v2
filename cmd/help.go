package cmd

import (
	"github.com/spf13/cobra"
)

// helpCmd represents the help command
var helpCmd = &cobra.Command{
	Use:   "help [command]",
	Short: "Help about any command",
	Long: `Help provides help for any command in the application.
Simply type ` + RootCmd.Name() + ` help [path to command] for full details.`, //nolint:staticcheck
	Run: func(cmd *cobra.Command, args []string) {
		RootCmd.Help()
	},
}
