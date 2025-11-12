package cmd

import (
	"fmt"
	"os" // Import os package
	"io/ioutil" // Import ioutil package

	"github.com/spf13/cobra"
)

// privacyCmd represents the privacy command
var privacyCmd = &cobra.Command{
	Use:   "privacy",
	Short: "Display the privacy notice",
	Long:  `The privacy command displays the privacy notice for the Gemini CLI.`, //nolint:staticcheck
	Run: func(cmd *cobra.Command, args []string) {
		privacyNoticePath := "docs/privacy_notice.md" // Path relative to project root
		content, err := ioutil.ReadFile(privacyNoticePath)
		if err != nil {
			fmt.Printf("Error reading privacy notice: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(content))
	},
}
