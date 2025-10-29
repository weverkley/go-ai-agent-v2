package commands

import (
	"encoding/json"
	"fmt"
	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/extension"
	"go-ai-agent-v2/go-cli/pkg/services"
	"os"
	"path/filepath"
	"strings"
)

const EXAMPLES_PATH = "/home/wever-kley/Workspace/go-ai-agent-v2/docs/gemini-cli-main/packages/cli/src/commands/extensions/examples"

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
	name, err := extensionManager.InstallOrUpdateExtension(installMetadata, args.Force)
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

// New creates a new extension.
func (c *ExtensionsCommand) New(args extension.NewArgs) error {
	fsService := services.NewFileSystemService()

	if args.Template != "" {
		// Implement copyDirectory logic here
		fmt.Printf("Creating new extension from template \"%s\" at %s (placeholder)\n", args.Template, args.Path)
		// Placeholder for copyDirectory
		templatePath := fsService.JoinPaths(EXAMPLES_PATH, args.Template)
		err := fsService.CopyDirectory(templatePath, args.Path)
		if err != nil {
			return fmt.Errorf("failed to create extension from template: %w", err)
		}
		fmt.Printf("Successfully created new extension from template \"%s\" at %s.\n", args.Template, args.Path)
	} else {
		// Implement createDirectory and gemini-extension.json creation logic here
		fmt.Printf("Creating new extension at %s (placeholder)\n", args.Path)
		// Placeholder for createDirectory
		err := fsService.CreateDirectory(args.Path)
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
		err = fsService.WriteFile(fsService.JoinPaths(args.Path, "gemini-extension.json"), string(manifestBytes))
		if err != nil {
			return fmt.Errorf("failed to write gemini-extension.json: %w", err)
		}
		fmt.Printf("Successfully created new extension at %s.\n", args.Path)
	}

	fmt.Printf("You can install this using \"gemini extensions link %s\" to test it out.\n", args.Path)
	return nil
}

// Enable enables an extension.
func (c *ExtensionsCommand) Enable(args extension.ExtensionScopeArgs) error {
	workspaceDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	extensionManager := extension.NewExtensionManager(workspaceDir)
	_, err = extensionManager.LoadExtensions()
	if err != nil {
		return fmt.Errorf("failed to load extensions: %w", err)
	}

	scope := config.SettingScopeUser
	if strings.ToLower(args.Scope) == "workspace" {
		scope = config.SettingScopeWorkspace
	}

	err = extensionManager.EnableExtension(args.Name, scope)
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
	workspaceDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	extensionManager := extension.NewExtensionManager(workspaceDir)
	_, err = extensionManager.LoadExtensions()
	if err != nil {
		return fmt.Errorf("failed to load extensions: %w", err)
	}

	scope := config.SettingScopeUser
	if strings.ToLower(args.Scope) == "workspace" {
		scope = config.SettingScopeWorkspace
	}

	err = extensionManager.DisableExtension(args.Name, scope)
	if err != nil {
		return fmt.Errorf("failed to disable extension: %w", err)
	}

	fmt.Printf("Extension \"%s\" successfully disabled for scope \"%s\".\n", args.Name, scope)
	return nil
}

// Update updates an extension or all extensions.
func (c *ExtensionsCommand) Update(name string, all bool) error {
	workspaceDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	extensionManager := extension.NewExtensionManager(workspaceDir)

	if all {
		extensions, err := extensionManager.LoadExtensions()
		if err != nil {
			return fmt.Errorf("failed to load extensions: %w", err)
		}
		for _, ext := range extensions {
			err := extensionManager.UpdateExtension(ext.Name)
			if err != nil {
				fmt.Printf("Error updating extension %s: %v\n", ext.Name, err)
			}
		}
		fmt.Println("All extensions updated.")
	} else {
		err := extensionManager.UpdateExtension(name)
		if err != nil {
			return fmt.Errorf("failed to update extension: %w", err)
		}
	}

	return nil
}

// Link links a local extension.
func (c *ExtensionsCommand) Link(path string) error {
	workspaceDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	extensionManager := extension.NewExtensionManager(workspaceDir)
	err = extensionManager.LinkExtension(path)
	if err != nil {
		return fmt.Errorf("failed to link extension: %w", err)
	}

	fmt.Printf("Extension at \"%s\" linked successfully.\n", path)
	return nil
}

