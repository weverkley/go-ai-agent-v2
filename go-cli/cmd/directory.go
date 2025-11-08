package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// directoryCmd represents the directory command
var directoryCmd = &cobra.Command{
	Use:     "directory",
	Aliases: []string{"dir"},
	Short:   "Manage workspace directories",
	Long:    `The directory command allows you to manage directories within the workspace, including adding and showing them.`, //nolint:staticcheck
}

var directoryAddCmd = &cobra.Command{
	Use:   "add <paths>",
	Short: "Add directories to the workspace",
	Long:  `Add directories to the workspace. Use comma to separate multiple paths.`, //nolint:staticcheck
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pathsToAdd := strings.Split(strings.Join(args, " "), ",")
		for _, p := range pathsToAdd {
			path := expandHomeDir(strings.TrimSpace(p))
			err := WorkspaceService.AddDirectory(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error adding directory %s: %v\n", path, err)
				os.Exit(1)
			}
			fmt.Printf("Added directory to workspace: %s\n", path)
		}
	},
}

var directoryShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show all directories in the workspace",
	Long:  `Show all directories currently configured in the workspace.`, //nolint:staticcheck
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		dirs := WorkspaceService.GetDirectories()
		if len(dirs) == 0 {
			fmt.Println("No directories configured in the workspace.")
			return
		}
		fmt.Println("Current workspace directories:")
		for _, dir := range dirs {
			fmt.Println("- ", dir)
		}
	},
}

func init() {
	directoryCmd.AddCommand(directoryAddCmd)
	directoryCmd.AddCommand(directoryShowCmd)
}

// expandHomeDir expands the home directory in a path (e.g., ~ or %userprofile%)
func expandHomeDir(p string) string {
	if len(p) == 0 {
		return ""
	}

	if strings.HasPrefix(p, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return p // Return original path if home directory cannot be determined
		}
		return filepath.Join(homeDir, p[1:])
	} else if strings.HasPrefix(strings.ToLower(p), "%userprofile%") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return p // Return original path if home directory cannot be determined
		}
		return filepath.Join(homeDir, p[len("%userprofile%"):])
	}
	return p
}
