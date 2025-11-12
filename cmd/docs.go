package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

// docsCmd represents the docs command
var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Open full Gemini CLI documentation in your browser",
	Long:  `The docs command opens the official documentation for the Gemini CLI in your web browser.`, //nolint:staticcheck
	Run: func(cmd *cobra.Command, args []string) {
		const docsUrl = "https://goo.gle/gemini-cli-docs"

		if os.Getenv("SANDBOX") != "" && os.Getenv("SANDBOX") != "sandbox-exec" {
			fmt.Printf("Please open the following URL in your browser to view the documentation:\n%s\n", docsUrl)
			return
		}

		fmt.Printf("Opening documentation in your browser: %s\n", docsUrl)

		var err error
		switch runtime.GOOS {
		case "linux":
			err = exec.Command("xdg-open", docsUrl).Start()
		case "windows":
			err = exec.Command("rundll32", "url.dll,FileProtocolHandler", docsUrl).Start()
		case "darwin":
			err = exec.Command("open", docsUrl).Start()
		default:
			err = fmt.Errorf("unsupported platform")
		}

		if err != nil {
			fmt.Printf("Error opening browser: %v\n", err)
			fmt.Printf("Please open the following URL in your browser to view the documentation manually:\n%s\n", docsUrl)
		}
	},
}
