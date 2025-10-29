package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Go Gemini CLI",
	Long:  `All software has versions. This is Go Gemini CLI's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Go Gemini CLI Version: 0.1.0")
	},
}
