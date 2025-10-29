package commands

import (
	"fmt"
	"go-ai-agent-v2/go-cli/pkg/extension"
	"go-ai-agent-v2/go-cli/pkg/services"
	"os"
	"strings"
)

// ExtensionsCommand represents the extensions command group.
type ExtensionsCommand struct {
	// Dependencies can be added here, e.g., FileSystemService, GitService
}

// NewExtensionsCommand creates a new instance of ExtensionsCommand.
func NewExtensionsCommand() *ExtensionsCommand {
	return &ExtensionsCommand{}
}

// ListExtensions lists installed extensions.
func (c *ExtensionsCommand) ListExtensions() error {
	workspaceDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	extensionManager := extension.NewExtensionManager(workspaceDir)
	extensions, err := extensionManager.LoadExtensions()
	if err != nil {
		return fmt.Errorf("failed to load extensions: %w", err)
	}

	if len(extensions) == 0 {
		fmt.Println("No extensions installed.")
		return nil
	}

	var outputStrings []string
	for _, ext := range extensions {
		outputStrings = append(outputStrings, extensionManager.ToOutputString(ext))
	}
	fmt.Println(strings.Join(outputStrings, "\n\n"))

	return nil
}

func (c *ExtensionsCommand) Install(args extension.InstallArgs) error {
	fmt.Printf("Installing extension from source: %s\n", args.Source)

	var installMetadata extension.ExtensionInstallMetadata
	// Determine source type
	if strings.HasPrefix(args.Source, "http://") ||
		strings.HasPrefix(args.Source, "https://") ||
		strings.HasPrefix(args.Source, "git@") ||
		strings.HasPrefix(args.Source, "sso://") {
		installMetadata = extension.ExtensionInstallMetadata{
			Source:        args.Source,
			Type:          "git",
			Ref:           args.Ref,
			AutoUpdate:    args.AutoUpdate,
			AllowPreRelease: args.AllowPreRelease,
		}
	} else {
		if args.Ref != "" || args.AutoUpdate {
			return fmt.Errorf("--ref and --auto-update are not applicable for local extensions.")
		}
		// Check if local path exists
		fsService := services.NewFileSystemService()
		exists, err := fsService.PathExists(args.Source)
		if err != nil {
			return fmt.Errorf("failed to check local path existence: %w", err)
		}
		if !exists {
			return fmt.Errorf("install source not found: %s", args.Source)
		}
		installMetadata = extension.ExtensionInstallMetadata{
			Source: args.Source,
			Type:   "local",
		}
	}

	// Consent handling
	if args.Consent {
		fmt.Println("You have consented to the installation.")
		// In a real scenario, you might log INSTALL_WARNING_MESSAGE here
	} else {
		// For now, we are skipping interactive consent. In a full implementation,
		// this would involve a prompt to the user.
		fmt.Println("Skipping interactive consent for now. Proceeding with installation.")
	}

	workspaceDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	extensionManager := extension.NewExtensionManager(workspaceDir)
	name, err := extensionManager.InstallOrUpdateExtension(installMetadata)
	if err != nil {
		return fmt.Errorf("failed to install/update extension: %w", err)
	}

	fmt.Printf("Extension \"%s\" installed successfully and enabled.\n", name)
	return nil
}

// Uninstall uninstalls an extension.
func (c *ExtensionsCommand) Uninstall(name string) error {
	workspaceDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	extensionManager := extension.NewExtensionManager(workspaceDir)
	err = extensionManager.UninstallExtension(name, false) // false for no interactive consent for now
	if err != nil {
		return fmt.Errorf("failed to uninstall extension: %w", err)
	}

	fmt.Printf("Extension \"%s\" successfully uninstalled.\n", name)
	return nil
}
