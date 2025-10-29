package cmd

import (
	"fmt"
	"os"

	"go-ai-agent-v2/go-cli/pkg/commands"
	"go-ai-agent-v2/go-cli/pkg/extension"
	"github.com/spf13/cobra"
)

var (
	extensionsInstallSource      string
	extensionsInstallRef         string
	extensionsInstallAutoUpdate  bool
	extensionsInstallAllowPreRelease bool
	extensionsInstallConsent     bool
	extensionsUninstallName    string
	extensionsNewPath          string
	extensionsNewTemplate      string
	extensionsEnableName        string
	extensionsEnableScope       string
	extensionsDisableName       string
	extensionsDisableScope      string
	extensionsUpdateName        string
	extensionsLinkPath          string
)

func init() {
	rootCmd.AddCommand(extensionsCmd)

	extensionsCmd.AddCommand(extensionsListCmd)

	extensionsCmd.AddCommand(extensionsInstallCmd)
	extensionsInstallCmd.Flags().StringVar(&extensionsInstallSource, "source", "", "The git URL or local path of the extension to install.")
	extensionsInstallCmd.Flags().StringVar(&extensionsInstallRef, "ref", "", "The git ref to install from.")
	extensionsInstallCmd.Flags().BoolVar(&extensionsInstallAutoUpdate, "auto-update", false, "Enable auto-update for this extension.")
	extensionsInstallCmd.Flags().BoolVar(&extensionsInstallAllowPreRelease, "pre-release", false, "Enable pre-release versions for this extension.")
	extensionsInstallCmd.Flags().BoolVar(&extensionsInstallConsent, "consent", false, "Acknowledge security risks and skip confirmation prompt.")
	extensionsInstallCmd.MarkFlagRequired("source")

	extensionsCmd.AddCommand(extensionsUninstallCmd)
	extensionsUninstallCmd.Flags().StringVar(&extensionsUninstallName, "name", "", "The name of the extension to uninstall.")
	extensionsUninstallCmd.MarkFlagRequired("name")

	extensionsCmd.AddCommand(extensionsNewCmd)
	extensionsNewCmd.Flags().StringVar(&extensionsNewPath, "path", "", "The path to create the extension in.")
	extensionsNewCmd.Flags().StringVar(&extensionsNewTemplate, "template", "", "The boilerplate template to use.")
	extensionsNewCmd.MarkFlagRequired("path")

	extensionsCmd.AddCommand(extensionsEnableCmd)
	extensionsEnableCmd.Flags().StringVar(&extensionsEnableName, "name", "", "The name of the extension to enable.")
	extensionsEnableCmd.Flags().StringVar(&extensionsEnableScope, "scope", "", "The scope to enable the extension in.")
	extensionsEnableCmd.MarkFlagRequired("name")

	extensionsCmd.AddCommand(extensionsDisableCmd)
	extensionsDisableCmd.Flags().StringVar(&extensionsDisableName, "name", "", "The name of the extension to disable.")
	extensionsDisableCmd.Flags().StringVar(&extensionsDisableScope, "scope", "", "The scope to disable the extension in.")
	extensionsDisableCmd.MarkFlagRequired("name")

	extensionsCmd.AddCommand(extensionsUpdateCmd)
	extensionsUpdateCmd.Flags().StringVar(&extensionsUpdateName, "name", "", "The name of the extension to update.")
	extensionsUpdateCmd.MarkFlagRequired("name")

	extensionsCmd.AddCommand(extensionsLinkCmd)
	extensionsLinkCmd.Flags().StringVar(&extensionsLinkPath, "path", "", "The path to the local extension to link.")
	extensionsLinkCmd.MarkFlagRequired("path")
}

var extensionsCmd = &cobra.Command{
	Use:   "extensions",
	Short: "Manage extensions",
	Long:  `Manage extensions for the Go Gemini CLI.`, 
}

var extensionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed extensions",
	Run: func(cmd *cobra.Command, args []string) {
		extensions := commands.NewExtensionsCommand()
		err := extensions.ListExtensions()
		if err != nil {
			fmt.Printf("Error listing extensions: %v\n", err)
			os.Exit(1)
		}
	},
}

var extensionsInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install an extension",
	Run: func(cmd *cobra.Command, args []string) {
		extensions := commands.NewExtensionsCommand()
		err := extensions.Install(extension.InstallArgs{
			Source:          extensionsInstallSource,
			Ref:             extensionsInstallRef,
			AutoUpdate:      extensionsInstallAutoUpdate,
			AllowPreRelease: extensionsInstallAllowPreRelease,
			Consent:         extensionsInstallConsent,
		})
		if err != nil {
			fmt.Printf("Error installing extension: %v\n", err)
			os.Exit(1)
		}
	},
}

var extensionsUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall an extension",
	Run: func(cmd *cobra.Command, args []string) {
		extensions := commands.NewExtensionsCommand()
		err := extensions.Uninstall(extensionsUninstallName)
		if err != nil {
			fmt.Printf("Error uninstalling extension: %v\n", err)
			os.Exit(1)
		}
	},
}

var extensionsNewCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new extension",
	Run: func(cmd *cobra.Command, args []string) {
		extensions := commands.NewExtensionsCommand()
		err := extensions.New(extension.NewArgs{
			Path:     extensionsNewPath,
			Template: extensionsNewTemplate,
		})
		if err != nil {
			fmt.Printf("Error creating new extension: %v\n", err)
			os.Exit(1)
		}
	},
}

var extensionsEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable an extension",
	Run: func(cmd *cobra.Command, args []string) {
		extensions := commands.NewExtensionsCommand()
		err := extensions.Enable(extension.ExtensionScopeArgs{
			Name:  extensionsEnableName,
			Scope: extensionsEnableScope,
		})
		if err != nil {
			fmt.Printf("Error enabling extension: %v\n", err)
			os.Exit(1)
		}
	},
}

var extensionsDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable an extension",
	Run: func(cmd *cobra.Command, args []string) {
		extensions := commands.NewExtensionsCommand()
		err := extensions.Disable(extension.ExtensionScopeArgs{
			Name:  extensionsDisableName,
			Scope: extensionsDisableScope,
		})
		if err != nil {
			fmt.Printf("Error disabling extension: %v\n", err)
			os.Exit(1)
		}
	},
}

var extensionsUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an extension",
	Run: func(cmd *cobra.Command, args []string) {
		extensions := commands.NewExtensionsCommand()
		err := extensions.Update(extensionsUpdateName)
		if err != nil {
			fmt.Printf("Error updating extension: %v\n", err)
			os.Exit(1)
		}
	},
}

var extensionsLinkCmd = &cobra.Command{
	Use:   "link",
	Short: "Link a local extension",
	Run: func(cmd *cobra.Command, args []string) {
		extensions := commands.NewExtensionsCommand()
		err := extensions.Link(extensionsLinkPath)
		if err != nil {
			fmt.Printf("Error linking extension: %v\n", err)
			os.Exit(1)
		}
	},
}
