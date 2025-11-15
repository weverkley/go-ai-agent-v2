package cmd

import (
	"fmt"
	"net/url"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

// bugCmd represents the bug command
var bugCmd = &cobra.Command{
	Use:   "bug",
	Short: "Report a bug or provide feedback",
	Long:  `The bug command opens a new issue in the Go AI Agent GitHub repository, allowing users to report bugs or provide feedback.`, //nolint:staticcheck
	Run: func(cmd *cobra.Command, args []string) {
		const githubNewIssueURL = "https://github.com/wever-kley/go-ai-agent-v2/issues/new"

		// Gather system information from the about command's details
		var bodyBuilder strings.Builder
		bodyBuilder.WriteString("\n\n---\n")
		bodyBuilder.WriteString(fmt.Sprintf("Go AI Agent Version: %s\n", version))
		bodyBuilder.WriteString(fmt.Sprintf("Build Date: %s\n", buildDate))
		bodyBuilder.WriteString(fmt.Sprintf("Git Commit: %s\n", gitCommit))
		bodyBuilder.WriteString(fmt.Sprintf("Go Version: %s\n", runtime.Version()))
		bodyBuilder.WriteString(fmt.Sprintf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH))
		bodyBuilder.WriteString("---\n")

		params := url.Values{}
		params.Add("labels", "bug")
		params.Add("title", "Bug Report: ")
		params.Add("body", bodyBuilder.String())

		fullURL := githubNewIssueURL + "?" + params.Encode()

		var err error
		switch runtime.GOOS {
		case "linux":
			err = exec.Command("xdg-open", fullURL).Start()
		case "windows":
			err = exec.Command("rundll32", "url.dll,FileProtocolHandler", fullURL).Start()
		case "darwin":
			err = exec.Command("open", fullURL).Start()
		default:
			err = fmt.Errorf("unsupported platform")
		}

		if err != nil {
			fmt.Printf("Error opening browser: %v\n", err)
			fmt.Printf("Please open a new issue manually at: %s\n", fullURL)
		} else {
			fmt.Println("Opening new issue in your default browser...")
		}
	},
}
