package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/commands"
	"go-ai-agent-v2/go-cli/pkg/core"
	"go-ai-agent-v2/go-cli/pkg/extension"
	"go-ai-agent-v2/go-cli/pkg/services"
)

func main() {
	version := flag.Bool("version", false, "Print version information")

	generateCmd := flag.NewFlagSet("generate", flag.ExitOnError)
	prompt := generateCmd.String("prompt", "", "The prompt for content generation")

	readCmd := flag.NewFlagSet("read", flag.ExitOnError)
	readFilePath := readCmd.String("file", "", "The path to the file to read")

	writeCmd := flag.NewFlagSet("write", flag.ExitOnError)
	writeFilePath := writeCmd.String("file", "", "The path to the file to write")
	writeContent := writeCmd.String("content", "", "The content to write to the file")

	execCmd := flag.NewFlagSet("exec", flag.ExitOnError)
	execCommand := execCmd.String("command", "", "The shell command to execute")
	execWorkingDir := execCmd.String("path", ".", "The working directory for the command")

	lsCmd := flag.NewFlagSet("ls", flag.ExitOnError)
	lsPath := lsCmd.String("path", ".", "The path to the directory to list")

	gitBranchCmd := flag.NewFlagSet("git-branch", flag.ExitOnError)
	gitBranchPath := gitBranchCmd.String("path", ".", "The path to the Git repository")

	extensionsCmd := flag.NewFlagSet("extensions", flag.ExitOnError)
	extensionsListCmd := flag.NewFlagSet("list", flag.ExitOnError)
	extensionsInstallCmd := flag.NewFlagSet("install", flag.ExitOnError)
	extensionsInstallSource := extensionsInstallCmd.String("source", "", "The git URL or local path of the extension to install.")
	extensionsInstallRef := extensionsInstallCmd.String("ref", "", "The git ref to install from.")
	extensionsInstallAutoUpdate := extensionsInstallCmd.Bool("auto-update", false, "Enable auto-update for this extension.")
	extensionsInstallAllowPreRelease := extensionsInstallCmd.Bool("pre-release", false, "Enable pre-release versions for this extension.")
	extensionsInstallConsent := extensionsInstallCmd.Bool("consent", false, "Acknowledge security risks and skip confirmation prompt.")

	extensionsUninstallCmd := flag.NewFlagSet("uninstall", flag.ExitOnError)
	extensionsUninstallName := extensionsUninstallCmd.String("name", "", "The name of the extension to uninstall.")

	mcpCmd := flag.NewFlagSet("mcp", flag.ExitOnError)
	mcpListCmd := flag.NewFlagSet("list", flag.ExitOnError)

	listModelsCmd := flag.NewFlagSet("list-models", flag.ExitOnError)

	flag.Parse()

	if *version {
		fmt.Println("Go Gemini CLI Version: 0.1.0")
		os.Exit(0)
	}

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "generate":
			generateCmd.Parse(os.Args[2:])
			if *prompt == "" {
				fmt.Println("Error: --prompt is required for generate command.")
				generateCmd.PrintDefaults()
				os.Exit(1)
			}

			// Initialize GeminiChat client (API key will be from env var later)
			geminiClient, err := core.NewGeminiChat()
			if err != nil {
				fmt.Printf("Error initializing GeminiChat: %v\n", err)
				os.Exit(1)
			}
			content, err := geminiClient.GenerateContent(*prompt)
			if err != nil {
				fmt.Printf("Error generating content: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(content)

		case "read":
			readCmd.Parse(os.Args[2:])
			if *readFilePath == "" {
				fmt.Println("Error: --file is required for read command.")
				readCmd.PrintDefaults()
				os.Exit(1)
			}
			fsService := services.NewFileSystemService()
			content, err := fsService.ReadFile(*readFilePath)
			if err != nil {
				fmt.Printf("Error reading file: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(content)

		case "write":
			writeCmd.Parse(os.Args[2:])
			if *writeFilePath == "" || *writeContent == "" {
				fmt.Println("Error: --file and --content are required for write command.")
				writeCmd.PrintDefaults()
				os.Exit(1)
			}
			fsService := services.NewFileSystemService()
			err := fsService.WriteFile(*writeFilePath, *writeContent)
			if err != nil {
				fmt.Printf("Error writing file: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Successfully wrote to %s\n", *writeFilePath)

		case "exec":
			execCmd.Parse(os.Args[2:])
			if *execCommand == "" {
				fmt.Println("Error: --command is required for exec command.")
				execCmd.PrintDefaults()
				os.Exit(1)
			}
			shellService := services.NewShellExecutionService()
			stdout, stderr, err := shellService.ExecuteCommand(*execCommand, *execWorkingDir)
			if err != nil {
				fmt.Printf("Error executing command: %v\n", err)
				if stdout != "" {
					fmt.Printf("Stdout:\n%s\n", stdout)
				}
				if stderr != "" {
					fmt.Printf("Stderr:\n%s\n", stderr)
				}
				os.Exit(1)
			}
			if stdout != "" {
				fmt.Printf("Stdout:\n%s\n", stdout)
			}
			if stderr != "" {
				fmt.Printf("Stderr:\n%s\n", stderr)
			}

		case "ls":
			lsCmd.Parse(os.Args[2:])
			fsService := services.NewFileSystemService()
			entries, err := fsService.ListDirectory(*lsPath)
			if err != nil {
				fmt.Printf("Error listing directory: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(strings.Join(entries, "\n"))

		case "git-branch":
			gitBranchCmd.Parse(os.Args[2:])
			gitService := services.NewGitService()
			branch, err := gitService.GetCurrentBranch(*gitBranchPath)
			if err != nil {
				fmt.Printf("Error getting git branch: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(branch)

		case "extensions":
			extensionsCmd.Parse(os.Args[2:])
			if len(extensionsCmd.Args()) == 0 {
				fmt.Println("Error: extensions command requires a subcommand.")
				extensionsCmd.PrintDefaults()
				os.Exit(1)
			}
			subcommand := extensionsCmd.Args()[0]
			switch subcommand {
			case "list":
				extensionsListCmd.Parse(extensionsCmd.Args()[1:])
				extensions := commands.NewExtensionsCommand()
				err := extensions.ListExtensions()
				if err != nil {
					fmt.Printf("Error listing extensions: %v\n", err)
					os.Exit(1)
				}
			case "install":
				extensionsInstallCmd.Parse(extensionsCmd.Args()[1:])
				if *extensionsInstallSource == "" {
					fmt.Println("Error: --source is required for install command.")
					extensionsInstallCmd.PrintDefaults()
					os.Exit(1)
				}
				extensions := commands.NewExtensionsCommand()
				err := extensions.Install(extension.InstallArgs{
					Source:          *extensionsInstallSource,
					Ref:             *extensionsInstallRef,
					AutoUpdate:      *extensionsInstallAutoUpdate,
					AllowPreRelease: *extensionsInstallAllowPreRelease,
					Consent:         *extensionsInstallConsent,
				})
				if err != nil {
					fmt.Printf("Error installing extension: %v\n", err)
					os.Exit(1)
				}
			case "uninstall":
				extensionsUninstallCmd.Parse(extensionsCmd.Args()[1:])
				if *extensionsUninstallName == "" {
					fmt.Println("Error: --name is required for uninstall command.")
					extensionsUninstallCmd.PrintDefaults()
					os.Exit(1)
				}
				extensions := commands.NewExtensionsCommand()
				err := extensions.Uninstall(*extensionsUninstallName)
				if err != nil {
					fmt.Printf("Error uninstalling extension: %v\n", err)
					os.Exit(1)
				}
			default:
				fmt.Printf("Unknown extensions subcommand: %s\n", subcommand)
				extensionsCmd.PrintDefaults()
				os.Exit(1)
			}

		case "mcp":
			mcpCmd.Parse(os.Args[2:])
			if len(mcpCmd.Args()) == 0 {
				fmt.Println("Error: mcp command requires a subcommand.")
				mcpCmd.PrintDefaults()
				os.Exit(1)
			}
			subcommand := mcpCmd.Args()[0]
			switch subcommand {
			case "list":
				mcpListCmd.Parse(mcpCmd.Args()[1:])
				mcp := commands.NewMcpCommand()
				err := mcp.ListMcpItems()
				if err != nil {
					fmt.Printf("Error listing MCP items: %v\n", err)
					os.Exit(1)
				}
			default:
				fmt.Printf("Unknown mcp subcommand: %s\n", subcommand)
				mcpCmd.PrintDefaults()
				os.Exit(1)
			}

		case "list-models":
			listModelsCmd.Parse(os.Args[2:])
			geminiClient, err := core.NewGeminiChat()
			if err != nil {
				fmt.Printf("Error initializing GeminiChat: %v\n", err)
				os.Exit(1)
			}
			models, err := geminiClient.ListModels()
			if err != nil {
				fmt.Printf("Error listing models: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Available Gemini Models:")
			for _, model := range models {
				fmt.Printf("- %s\n", model)
			}

		case "help":
			flag.PrintDefaults()
			fmt.Println("\nCommands:")
			fmt.Println("  generate --prompt <text>  Generate content based on a prompt.")
			fmt.Println("  read --file <path>        Read content from a file.")
			fmt.Println("  write --file <path> --content <text> Write content to a file.")
			fmt.Println("  exec --command <cmd>      Execute a shell command.")
			fmt.Println("  ls [--path <path>]        List contents of a directory.")
			fmt.Println("  git-branch [--path <path>] Get the current Git branch name.")
			fmt.Println("  extensions list           List installed extensions.")
			fmt.Println("  extensions install --source <src> [--ref <ref>] [--auto-update] [--pre-release] [--consent] Install an extension.")
			fmt.Println("  extensions uninstall --name <name> Uninstall an extension.")
			fmt.Println("  mcp list                  List MCP items.")
			fmt.Println("  list-models               List available Gemini models.")
			os.Exit(0)
		default:
			fmt.Printf("Unknown command: %s\n", os.Args[1])
			flag.PrintDefaults()
			os.Exit(1)
		}
	} else {
		fmt.Println("Go Gemini CLI is running! Use -h for help.")
	}
}
