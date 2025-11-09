# Go AI Agent v2 CLI

This repository contains the Go-based Command Line Interface (CLI) for the AI Agent v2, designed to bring the power of various AI executors (e.g., Gemini, OpenAI) directly into your terminal. This project is a migration from an existing JavaScript-based CLI, focusing on a robust, extensible, and efficient Go implementation.

## Project Status

The foundational structure for the Go CLI is established, with several core services and commands implemented. Many original JavaScript functionalities have been translated, including tool-calling capabilities. All identified type-checking and unused import errors have been addressed.

### Key Features & Implemented Commands:

*   **`generate`**: Fully functional with tool-calling capabilities and an interactive UI (using `charmbracelet/bubbletea`) for generating content via the Gemini API.
*   **`read`**: Reads file content.
*   **`write`**: Writes content to a file, now utilizing the `write_file` tool.
*   **`exec`**: Executes shell commands.
*   **`ls`**: Lists directory contents.
*   **`git-branch`**: Gets the current Git branch name.
*   **`extract-function`**: Extracts a code block into a new function or method.
*   **`find-unused-code`**: Finds unused functions and methods in Go files.
*   **`extensions`**: Command group for managing extensions:
    *   `list`: Lists discovered extensions.
    *   `install`: Installs extensions (supports git clone/pull and local copy).
    *   `uninstall`: Uninstalls extensions.
    *   `new`: Creates new extensions.
    *   `enable`: Enables extensions.
    *   `disable`: Disables extensions.
*   **`mcp`**: Command group for managing Model Context Protocol (MCP) servers (currently provides informative messages about future implementation).
*   **AI-Powered Commands (migrated from JavaScript .toml files)**:
    *   `code-guide`: Answers codebase questions using AI.
    *   `find-docs`: Finds relevant documentation and outputs GitHub URLs using AI.
    *   `cleanup-back-to-main`: Automates Git branch cleanup.
    *   `pr-review`: Conducts comprehensive AI-driven pull request reviews.
    *   `grep-code`: Summarizes code findings for a given pattern using grep and AI.
*   **`list-models`**: Lists available Gemini models.
*   **Informative Commands**: `privacy`, `ide`, `theme`, `auth`, `setup-github`, `profile`, `vim`, `terminal-setup`, `editor`, `stats`, `restore`, `permissions` now provide informative messages about their future implementation.
*   **`tools run`**: Executes specific AI tools.

### Core Services & Tools:

*   **`pkg/core/gemini.go`**: Integrates with the Gemini API, including tool-calling and actual token counting.
*   **`pkg/services/shell_service.go`**: Provides shell command execution.
*   **`pkg/services/file_system_service.go`**: Offers file system operations like listing, reading, writing, and directory management.
*   **`pkg/services/git_service.go`**: Interacts with Git repositories for operations like getting remote URLs, checking out branches, pulling, and deleting branches.
*   **`pkg/extension/manager.go`**: Handles discovery, loading, installation, and management of extensions.
*   **`pkg/config/config.go`**: Manages application settings and configuration, including the `.gemini` directory path.
*   **`pkg/mcp/client.go`**: Simulates MCP connection.
*   **`pkg/types/types.go`**: Centralized definitions for various application types and interfaces.
*   **`pkg/utils/folder_structure.go`**: Implements logic for generating folder structures, including file ignoring.
*   **`pkg/utils/utils.go`**: Provides utility functions, now including actual telemetry logging.
*   **`pkg/tools/ls.go`**: Implements the `ls` tool functionality.
*   **`pkg/tools/write_file.go`**: Implements the `write_file` tool functionality.

## Getting Started

### Prerequisites

*   Go (version 1.21 or higher recommended)
*   `GEMINI_API_KEY` environment variable set with your Gemini API key.

### Building the CLI

Navigate to the `go-cli` directory and run:

```bash
go build -o gemini-cli .
```

This will compile the application and create an executable named `gemini-cli` in the current directory.

### Running Commands

You can run the compiled CLI using:

```bash
./gemini-cli [command] [flags]
```

For example:

```bash
./gemini-cli generate "Write a short story about a robot."
./gemini-cli ls .
./gemini-cli extensions list
./gemini-cli extract-function --file-path /path/to/file.go --start-line 10 --end-line 20 --new-function-name MyNewFunction
```

To see a list of all available commands and their options, run:

```bash
./gemini-cli --help
```

## Next Steps & Future Enhancements

The project is under active development. Future enhancements include:

*   **Mock Executor and Executor Factory**: For comprehensive testing and extensibility to other AI models.
*   **Improved Error Handling and User Feedback**: More user-friendly messages and consistent feedback for long-running operations.
*   **Comprehensive Testing**: Unit and integration tests for all new components.
*   **Secure API Key Management**: Robust OS-specific storage and clearing of API keys.
*   **IDE Integration**: Full integration with supported IDEs.
*   **Theme Customization**: Ability to change and persist CLI visual themes.
*   **Terminal Keybinding Configuration**: For enhanced multiline input.
*   **External Editor Preference**: Setting and using a preferred external editor.
*   **Detailed Usage Statistics**: Model and tool-specific metrics.
*   **Restore Functionality**: Saving and restoring CLI state.
*   **Folder Trust Management**: Security features for managing trusted folders.
*   **MCP Server Management**: Full CRUD operations for MCP servers.

## Contributing

Contributions are welcome! Please refer to the `PLAN.md` for the development strategy and `docs/gemini-cli-main/GEMINI.md` for coding conventions and guidelines.

## License

[Specify your project's license here]
