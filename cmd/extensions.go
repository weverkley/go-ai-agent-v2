package cmd

import (
	"fmt"
	"os"

	"go-ai-agent-v2/go-cli/pkg/extension"

	"github.com/spf13/cobra"
)

const EXAMPLES_PATH = "/home/wever-kley/Workspace/go-ai-agent-v2/docs/go-ai-agent-main/packages/cli/src/commands/extensions/examples"

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
		if err := extensionsCliCommand.ListExtensions(); err != nil {
			fmt.Fprintf(os.Stderr, "Error listing extensions: %v\n", err)
			os.Exit(1)
		}
	},
}

var extensionsEnableCmd = &cobra.Command{
	Use:   "enable <extension_name>",
	Short: "Enable a specific extension",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		extensionName := args[0]
		enableArgs := extension.ExtensionScopeArgs{
			Name: extensionName,
			// Scope is not currently used in Enable, but keeping the struct consistent
			Scope: "",
		}
		if err := extensionsCliCommand.Enable(enableArgs); err != nil {
			fmt.Fprintf(os.Stderr, "Error enabling extension '%s': %v\n", extensionName, err)
			os.Exit(1)
		}
	},
}

var extensionsDisableCmd = &cobra.Command{
	Use:   "disable <extension_name>",
	Short: "Disable a specific extension",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		extensionName := args[0]
		disableArgs := extension.ExtensionScopeArgs{
			Name: extensionName,
			// Scope is not currently used in Disable, but keeping the struct consistent
			Scope: "",
		}
		if err := extensionsCliCommand.Disable(disableArgs); err != nil {
			fmt.Fprintf(os.Stderr, "Error disabling extension '%s': %v\n", extensionName, err)
			os.Exit(1)
		}
	},
}

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
			Source:          source,
			Ref:             ref,
			AutoUpdate:      autoUpdate,
			AllowPreRelease: allowPreRelease,
			Force:           force,
			Consent:         consent,
		}

		if err := extensionsCliCommand.Install(installArgs); err != nil {
			fmt.Fprintf(os.Stderr, "Error installing extension: %v\n", err)
			os.Exit(1)
		}
	},
}

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

var uninstallCmd = &cobra.Command{
	Use:   "uninstall <extension_name>",
	Short: "Uninstall an extension",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		extensionName := args[0]
		consent, _ := cmd.Flags().GetBool("consent")
		if err := extensionsCliCommand.Uninstall(extensionName, consent); err != nil {
			fmt.Fprintf(os.Stderr, "Error uninstalling extension '%s': %v\n", extensionName, err)
			os.Exit(1)
		}
	},
}

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

func init() {
	ExtensionsCmd.AddCommand(extensionsListCmd)
	ExtensionsCmd.AddCommand(extensionsEnableCmd)
	ExtensionsCmd.AddCommand(extensionsDisableCmd)
	ExtensionsCmd.AddCommand(installCmd)
	ExtensionsCmd.AddCommand(newCmd)
	ExtensionsCmd.AddCommand(updateCmd)
	ExtensionsCmd.AddCommand(uninstallCmd)
	ExtensionsCmd.AddCommand(linkCmd)

	// Add flags for installCmd
	installCmd.Flags().String("ref", "", "Specify a ref (branch, tag, or commit) for git installations.")
	installCmd.Flags().Bool("auto-update", false, "Enable automatic updates for the extension.")
	installCmd.Flags().Bool("allow-prerelease", false, "Allow installation of pre-release versions.")
	installCmd.Flags().Bool("force", false, "Force installation, overwriting existing extensions.")
	installCmd.Flags().Bool("consent", false, "Provide consent for installation (e.g., for security warnings).")

	// Add flags for newCmd
	newCmd.Flags().String("template", "", "Specify a template to create the new extension from.")

	// Add flags for updateCmd
	updateCmd.Flags().Bool("all", false, "Update all installed extensions.")

	// Add flags for uninstallCmd
	uninstallCmd.Flags().Bool("consent", false, "Provide consent for uninstallation (e.g., for security warnings).")
}
