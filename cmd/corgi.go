package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// corgiCmd represents the corgi command
var corgiCmd = &cobra.Command{
	Use:   "corgi",
	Short: "Display a cute corgi ASCII art",
	Long:  `The corgi command displays a cute ASCII art representation of a corgi.`, //nolint:staticcheck
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(`
  / \__
 (    @\___
 /         O
/   (_____/
/_____/   U`)
	},
}
