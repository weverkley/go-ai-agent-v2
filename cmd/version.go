package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Go AI Agent",
	Long:  `All software has versions. This is Go AI Agent's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Go AI Agent Version: 0.1.0")
	},
}
