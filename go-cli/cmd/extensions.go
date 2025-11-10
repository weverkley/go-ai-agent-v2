package cmd

import (
	"fmt"
	"os"

	"go-ai-agent-v2/go-cli/pkg/commands" // Add commands import
	"go-ai-agent-v2/go-cli/pkg/extension" // Add extension import
	"github.com/spf13/cobra"
)

// extensionsCmd represents the extensions command group
var ExtensionsCmd = &cobra.Command{
	Use:   "extensions",
	Short: "Manage CLI extensions",
	Long:  `The extensions command group allows you to list, enable, and disable CLI extensions.`, //nolint:staticcheck
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand is provided, print help
		cmd.Help()
	},
}

var extensionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available extensions",
	Run: func(cmd *cobra.Command, args []string) {
		extensions := ExtensionManager.ListExtensions()
		if len(extensions) == 0 {
			fmt.Println("No extensions found.")
			return
		}
		fmt.Println("Available extensions:")
		for _, ext := range extensions {
			status := "Disabled"
			if ext.Enabled {
				status = "Enabled"
			}
			fmt.Printf("- %s: %s (%s)\n", ext.Name, ext.Description, status)
		}
	},
}

var extensionsEnableCmd = &cobra.Command{
	Use:   "enable <extension_name>",
	Short: "Enable a specific extension",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		extensionName := args[0]
		err := ExtensionManager.EnableExtension(extensionName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error enabling extension '%s': %v\n", extensionName, err)
			os.Exit(1)
		}
		fmt.Printf("Extension '%s' enabled successfully.\n", extensionName)
	},
}

// extensionsDisableCmd represents the disable command
var extensionsDisableCmd = &cobra.Command{
	Use:   "disable <extension_name>",
	Short: "Disable a specific extension",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		extensionName := args[0]
		err := ExtensionManager.DisableExtension(extensionName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error disabling extension '%s': %v\n", extensionName, err)
			os.Exit(1)
		}
		fmt.Printf("Extension '%s' disabled successfully.\n", extensionName)
	},
}

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install <source>",
	Short: "Install a new extension",
	Long: `Install a new extension from a git repository or a local path.

Examples:
  gemini extensions install https://github.com/user/my-extension.git
  gemini extensions install /path/to/local/extension
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		source := args[0]
		ref, _ := cmd.Flags().GetString("ref")
		autoUpdate, _ := cmd.Flags().GetBool("auto-update")
		allowPreRelease, _ := cmd.Flags().GetBool("allow-prerelease")
		force, _ := cmd.Flags().GetBool("force")
		consent, _ := cmd.Flags().GetBool("consent")

		installArgs := extension.InstallArgs{
			Source:        source,
			Ref:           ref,
			AutoUpdate:    autoUpdate,
			AllowPreRelease: allowPreRelease,
			Force:         force,
			Consent:       consent,
		}

		if err := extensionsCliCommand.Install(installArgs); err != nil {
			fmt.Fprintf(os.Stderr, "Error installing extension: %v\n", err)
			os.Exit(1)
		}
	},
}

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new <path>",
	Short: "Create a new extension project",
	Long: `Create a new extension project at the specified path.
Optionally, you can specify a template to start from.

Examples:
  gemini extensions new my-new-extension
  gemini extensions new my-new-extension --template basic
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		template, _ := cmd.Flags().GetString("template")

		newArgs := extension.NewArgs{
			Path:     path,
			Template: template,
		}

		if err := extensionsCliCommand.New(newArgs); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating new extension: %v\n", err)
			os.Exit(1)
		}
	},
}

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update [extension_name]",
	Short: "Update an extension or all extensions",
	Long: `Update a specific extension or all installed extensions.

Examples:
  gemini extensions update my-extension
  gemini extensions update --all
`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		all, _ := cmd.Flags().GetBool("all")
		var name string
		if len(args) > 0 {
			name = args[0]
		}

		if all && name != "" {
			fmt.Fprintln(os.Stderr, "Error: Cannot specify both an extension name and --all flag.")
			os.Exit(1)
		}
		if !all && name == "" {
			fmt.Fprintln(os.Stderr, "Error: Must specify an extension name or use --all flag.")
			os.Exit(1)
		}

		if err := extensionsCliCommand.Update(name, all); err != nil {
			fmt.Fprintf(os.Stderr, "Error updating extension(s): %v\n", err)
			os.Exit(1)
		}
	},
}

// linkCmd represents the link command
var linkCmd = &cobra.Command{
	Use:   "link <path>",
	Short: "Link a local extension",
	Long: `Link a local directory as an extension. This is useful for developing extensions locally.

Example:
  gemini extensions link /path/to/my/local/extension
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		if err := extensionsCliCommand.Link(path); err != nil {
			fmt.Fprintf(os.Stderr, "Error linking extension: %v\n", err)
			os.Exit(1)
		}
	},
}

var extensionsCliCommand *commands.ExtensionsCommand

func init() {
}
