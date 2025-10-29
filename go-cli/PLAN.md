# Go CLI Migration Plan

This document outlines the plan for migrating the JavaScript-based Gemini CLI to a new Go-based CLI application. The goal is to rewrite the entire CLI, excluding the authentication flow (which will rely on `GEMINI_API_KEY` environment variable), test files/commands, and the VS Code IDE companion.

## 1. Current Status

The foundational structure for the Go CLI has been established, and several core services and commands have been implemented.

*   **Go Project Setup**: An empty Go module (`go-ai-agent-v2/go-cli`) has been initialized.
*   **Core CLI Structure (main.go)**:
    *   Basic command-line argument parsing is in place using Go's `flag` package.
    *   `--version` flag implemented.
    *   Top-level commands implemented:
        *   `generate`: **Functional**, generates content using `pkg/core/gemini.go` (real Gemini API integration).
        *   `read`: Reads file content (currently using `pkg/tools/file_tools.go`, will be moved to `pkg/services/file_system_service.go`).
        *   `write`: Writes content to a file (currently using `pkg/tools/file_tools.go`, will be moved to `pkg/services/file_system_service.go`).
        *   `exec`: Executes shell commands (uses `pkg/services/shell_service.go`).
        *   `ls`: Lists directory contents (uses `pkg/services/file_system_service.go`).
        *   `git-branch`: Gets the current Git branch name (uses `pkg/services/git_service.go` with `go-git`).
        *   `extensions`: Command group with subcommands:
            *   `list`: **Functional**, lists discovered extensions by reading `gemini-extension.json` files.
            *   `install`: **Partially implemented**, command structure and argument parsing in `main.go` are ready, but core logic in `pkg/commands/extensions.go` and `pkg/extension/manager.go` needs further refinement (specifically file operations).
        *   `mcp`: Command group with subcommands:
            *   `list`: **Functional**, lists configured MCP servers, merging from settings and extensions, and simulates connection status.
        *   `list-models`: **Functional**, lists available Gemini models using `pkg/core/gemini.go`.
*   **Core Services & Tools (pkg/core, pkg/extension, pkg/config, pkg/mcp, pkg/services)**:
    *   `pkg/core/gemini.go`: **Functional**, uses `google.golang.org/genai` for Gemini API interaction.
    *   `pkg/tools/file_tools.go`: **Deprecated**, `ReadFile` and `WriteFile` functionality will be moved to `pkg/services/file_system_service.go`.
    *   `pkg/services/shell_service.go`: Provides `ExecuteCommand` for shell operations.
    *   `pkg/services/file_system_service.go`: Provides `ListDirectory`, `PathExists`, `IsDirectory`, `JoinPaths`, and now `WriteFile` (will also include `ReadFile`).
    *   `pkg/services/git_service.go`: Uses `github.com/go-git/go-git/v5` to interact with Git repositories.
    *   `pkg/extension/manager.go`: Discovers and loads extensions, parses `gemini-extension.json`. Includes placeholder for `InstallOrUpdateExtension`.
    *   `pkg/extension/types.go`: Defines `InstallArgs` and `ExtensionInstallMetadata`.
    *   `pkg/config/settings.go`: Loads application settings, including extension paths and MCP server configurations.
    *   `pkg/mcp/types.go`: Defines `MCPServerStatus` and `MCPServerConfig`.
    *   `pkg/mcp/client.go`: Placeholder for MCP client interaction.

## 2. Remaining Core Components for Translation

Based on the analysis of `gemini-cli-main/packages/core/src/index.ts`, the following significant modules/functionalities still need to be translated into Go. Prioritization will depend on the specific CLI commands being implemented.

*   **AI Agent Logic (`pkg/core/agents`)**: Translation of the AI agent components that drive the intelligent behavior of the CLI.
*   **Code Assist (`pkg/core/code_assist`)**: Functionality related to code generation, completion, or analysis.
*   **Prompts (`pkg/core/prompts`)**: Management and rendering of prompts for AI models or user interactions.
*   **Tools (other than file I/O)**: Specialized tools such as `grep`, `glob`, `web-fetch`, `memoryTool`, `web-search`, `read-many-files`, etc., need to be implemented.
*   **Config (`pkg/core/config`)**: Robust configuration management beyond just reading an environment variable (e.g., loading from files).
*   **Output (`pkg/core/output`)**: Handling rich output formatting and display, potentially adapting existing JavaScript formatting logic.
*   **Policy (`pkg/core/policy`)**: Implementation of any policy enforcement or decision-making logic.
*   **Confirmation Bus (`pkg/core/confirmation-bus`)**: System for handling user confirmations or asynchronous operations.
*   **IDE Integration (`pkg/core/ide`)**: (Lower priority, as user explicitly excluded VS Code companion, but might include generic IDE-agnostic features).
*   **Telemetry (`pkg/core/telemetry`)**: Implementation of usage data collection.

## 3. Command Implementation Strategy

The `extensions` and `mcp` command groups are the primary CLI functionalities to be translated.

### 3.1. Extensions Commands (`pkg/commands/extensions.go`)

Translate the logic from the following JavaScript files into Go:

*   `install.ts`: **In Progress**. Core logic in `pkg/extension/manager.go` needs to be completed (git clone/local copy). Argument parsing in `main.go` is ready.
*   `list.ts`: **Functional**.
*   `new.ts`: Logic for creating new extensions.
*   `enable.ts`, `disable.ts`: Logic for enabling/disabling extensions.
*   `uninstall.ts`: Logic for uninstalling extensions.
*   `update.ts`: Logic for updating extensions.
*   `link.ts`: Logic for linking local extensions.

Each command will require:
1.  **Argument Parsing**: Define flags specific to the subcommand.
2.  **Service Interaction**: Utilize the core services (`FileSystemService`, `GitService`, etc.) as needed.
3.  **JavaScript File Analysis**: For each command, the corresponding JavaScript file will be read and analyzed to understand its functionality and internal logic before translation to Go.

### 3.2. MCP Commands (`pkg/commands/mcp.go`)

Translate the logic from the following JavaScript files into Go:

*   `add.ts`: Logic for adding MCP items.
*   `list.ts`: **Functional** (simulated connection status).
*   `remove.ts`: Logic for removing MCP items.

Similar to extensions, each MCP command will involve argument parsing, service interaction, and thorough analysis of the original JavaScript source.

## 4. API Integration Strategy

*   **Gemini API Client**: **Functional**, uses `google.golang.org/genai` for Gemini API interaction. `GEMINI_API_KEY` is read from the environment.
*   **Error Handling**: Implement robust error handling for API calls, including retries and clear error messages.

## 5. Testing Strategy

*   **Unit Tests**: Write unit tests for individual functions and methods within each Go package (`pkg/core`, `pkg/tools`, `pkg/services`, `pkg/commands`) to ensure correctness.
*   **Integration Tests**: Develop integration tests for CLI commands to verify they interact correctly with the services and produce expected outputs. Since the user explicitly excluded rewriting test *files*, new Go-native tests will be created.
*   **Manual Testing**: Regular manual testing of the CLI commands at various stages of implementation to ensure functionality.

## 6. Execution Flow

The migration will proceed iteratively, focusing on one command or core functionality at a time, following these steps:
1.  **Identify Target**: Choose a specific JavaScript command or core module to translate.
2.  **Analyze JavaScript Source**: Read and understand the corresponding JavaScript file(s) to grasp functionality, dependencies, and logic.
3.  **Design Go Implementation**: Outline the Go structures, interfaces, and functions required.
4.  **Implement in Go**: Write the Go code.
5.  **Integrate with CLI**: Add the new Go command or integrate the new Go module into `main.go`.
6.  **Test**: Write and run Go unit/integration tests, and perform manual testing.
7.  **Refine**: Address any issues or improvements.

This plan provides a roadmap for the Go CLI migration. I am now ready for your instructions on which specific command or core functionality to tackle next.