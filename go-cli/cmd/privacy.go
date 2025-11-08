package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// privacyCmd represents the privacy command
var privacyCmd = &cobra.Command{
	Use:   "privacy",
	Short: "Display the privacy notice",
	Long:  `The privacy command displays the privacy notice for the Gemini CLI.`, //nolint:staticcheck
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(`Gemini CLI Privacy Notice:

This CLI tool may collect anonymous usage data to help improve its features and performance. This data does not include any personal identifiable information.

Specifically, this tool may collect:
- Command usage statistics (e.g., which commands are run, how often)
- Error reports (e.g., stack traces, error messages)
- Performance metrics (e.g., command execution times)

This data is used solely for product improvement and debugging purposes. It is not shared with third parties.

You can disable telemetry collection in the settings.

For more detailed information, please refer to the full privacy policy available at [Link to full privacy policy, if applicable].`)
	},
}
