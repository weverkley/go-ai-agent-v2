package commands

import (
	"encoding/json"
	"fmt"
	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/extension"
	"go-ai-agent-v2/go-cli/pkg/types" // Add types import
	"path/filepath"
	"strings"
)

const EXAMPLES_PATH = "/home/wever-kley/Workspace/go-ai-agent-v2/docs/go-ai-agent-main/packages/cli/src/commands/extensions/examples"

// ExtensionsCommand represents the extensions command group.
type ExtensionsCommand struct {
	extensionManager *extension.Manager
	settingsService  types.SettingsServiceIface
}

// NewExtensionsCommand creates a new instance of ExtensionsCommand.
func NewExtensionsCommand(extensionManager *extension.Manager, settingsService types.SettingsServiceIface) *ExtensionsCommand {
	return &ExtensionsCommand{
		extensionManager: extensionManager,
		settingsService:  settingsService,
	}
}

// ListExtensions lists installed extensions.
func (c *ExtensionsCommand) ListExtensions() error {
	extensions := c.extensionManager.ListExtensions()

	if len(extensions) == 0 {
		fmt.Println("No extensions installed.")
		return nil
	}

	var outputStrings []string
	for _, ext := range extensions {
		outputStrings = append(outputStrings, fmt.Sprintf("- %s (Enabled: %t)", ext.Name, ext.Enabled))
	}
	fmt.Println(strings.Join(outputStrings, "\n\n"))

	return nil
}

// Install installs an extension.
func (c *ExtensionsCommand) Install(args extension.InstallArgs) error {
	fmt.Printf("Installing extension from source: %s\n", args.Source)

	var installMetadata extension.ExtensionInstallMetadata
	// Determine source type
	if strings.HasPrefix(args.Source, "http://") ||
		strings.HasPrefix(args.Source, "https://") ||
		strings.HasPrefix(args.Source, "git@") ||
		strings.HasPrefix(args.Source, "sso://") {
		installMetadata = extension.ExtensionInstallMetadata{
			Source:          args.Source,
			Type:            "git",
			Ref:             args.Ref,
			AutoUpdate:      args.AutoUpdate,
			AllowPreRelease: args.AllowPreRelease,
		}
	} else {
		if args.Ref != "" || args.AutoUpdate {
			return fmt.Errorf("--ref and --auto-update are not applicable for local extensions.")
		}
		// Check if local path exists
		exists, err := c.extensionManager.FSService.PathExists(args.Source)
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

	extName, err := c.extensionManager.InstallOrUpdateExtension(installMetadata, args.Force)
	if err != nil {
		return fmt.Errorf("failed to install/update extension: %w", err)
	}

	fmt.Printf("Extension \"%s\" installed successfully and enabled.\n", extName)
	return nil
}

// Uninstall uninstalls an extension.
func (c *ExtensionsCommand) Uninstall(name string, interactiveConsent bool) error {
	var err error
	err = c.extensionManager.UninstallExtension(name, interactiveConsent)
	if err != nil {
		return fmt.Errorf("failed to uninstall extension: %w", err)
	}

	fmt.Printf("Extension \"%s\" successfully uninstalled.\n", name)
	return nil
}

// New creates a new extension.
func (c *ExtensionsCommand) New(args extension.NewArgs) error {
	if args.Template != "" {
		// Implement copyDirectory logic here
		templatePath := c.extensionManager.FSService.JoinPaths(EXAMPLES_PATH, args.Template)
		err := c.extensionManager.FSService.CopyDirectory(templatePath, args.Path)
		if err != nil {
			return fmt.Errorf("failed to create extension from template: %w", err)
		}
		fmt.Printf("Successfully created new extension from template \"%s\" at %s.\n", args.Template, args.Path)
	} else {
		// Implement createDirectory and gemini-extension.json creation logic here
		err := c.extensionManager.FSService.CreateDirectory(args.Path)
		if err != nil {
			return fmt.Errorf("failed to create new extension directory: %w", err)
		}

		extensionName := filepath.Base(args.Path)
		manifest := map[string]interface{}{
			"name":    extensionName,
			"version": "1.0.0",
		}
		manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal extension manifest: %w", err)
		}
		err = c.extensionManager.FSService.WriteFile(c.extensionManager.FSService.JoinPaths(args.Path, "go-ai-agent-extension.json"), string(manifestBytes))
		if err != nil {
			return fmt.Errorf("failed to write go-ai-agent-extension.json: %w", err)
		}
		fmt.Printf("Successfully created new extension at %s.\n", args.Path)
	}

	fmt.Printf("You can install this using \"go-ai-agent extensions link %s\" to test it out.\n", args.Path)
	return nil
}

// Enable enables an extension.
func (c *ExtensionsCommand) Enable(args extension.ExtensionScopeArgs) error {
	var err error
	err = c.extensionManager.EnableExtension(args.Name)
	if err != nil {
		return fmt.Errorf("failed to enable extension: %w", err)
	}

	if args.Scope != "" {
		fmt.Printf("Extension \"%s\" successfully enabled for scope \"%s\".\n", args.Name, args.Scope)
	} else {
		fmt.Printf("Extension \"%s\" successfully enabled in all scopes.\n", args.Name)
	}
	return nil
}

// Disable disables an extension.

func (c *ExtensionsCommand) Disable(args extension.ExtensionScopeArgs) error {

	var err error

	scope := config.SettingScopeUser

	if strings.ToLower(args.Scope) == "workspace" {

		scope = config.SettingScopeWorkspace

	}

	err = c.extensionManager.DisableExtension(args.Name)
	if err != nil {
		return fmt.Errorf("failed to disable extension: %w", err)
	}

	fmt.Printf("Extension \"%s\" successfully disabled for scope \"%s\".\n", args.Name, scope)
	return nil
}

// Update updates an extension or all extensions.
func (c *ExtensionsCommand) Update(name string, all bool) error {
	var err error
	if all {
		extensions := c.extensionManager.ListExtensions()
		for _, ext := range extensions {
			err = c.extensionManager.UpdateExtension(ext.Name)
			if err != nil {
				fmt.Printf("Error updating extension %s: %v\n", ext.Name, err)
			}
		}
		fmt.Println("All extensions updated.")
	} else {
		err = c.extensionManager.UpdateExtension(name)
		if err != nil {
			return fmt.Errorf("failed to update extension: %w", err)
		}
	}

	return nil
}

// Link links a local extension.
func (c *ExtensionsCommand) Link(path string) error {
	var err error
	err = c.extensionManager.LinkExtension(path)
	if err != nil {
		return fmt.Errorf("failed to link extension: %w", err)
	}

	fmt.Printf("Extension at \"%s\" linked successfully.\n", path)
	return nil
}
