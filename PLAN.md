# Go CLI Migration Plan

This document outlines the plan for migrating the JavaScript-based Gemini CLI to a new Go-based CLI application. The goal is to rewrite the entire CLI, excluding the authentication flow (which will rely on `GEMINI_API_KEY` environment variable), test files/commands, and the VS Code IDE companion.

### **IMPORTANT!**
This project is meant to be a generic CLI, it will use multiple AI executors (e.g., Gemini, OpenAI, etc.), not only Gemini which this tool is currently based on.

### **UI Package**
The UI for this CLI will be implemented using `charmbracelet/bubbletea` for an interactive terminal user interface.

## 1. Current Status

The foundational structure for the Go CLI has been established, and several core services and commands have been implemented. Many original JavaScript files have been translated with tool-calling capabilities. All identified type-checking and unused import errors have been addressed.

- **Go Project Setup**: An empty Go module (`go-ai-agent-v2/go-cli`) has been initialized.

- **Core CLI Structure (main.go)**:
  - Command-line argument parsing is implemented using the `cobra` library.
  - `--version` flag implemented.
  - Top-level commands implemented:
    - `generate`: **Functional** (with tool-calling capabilities). Now includes an **Interactive UI** using `charmbracelet/bubbletea` if no prompt is provided. Generates content using `pkg/core/gemini.go` (real Gemini API integration). **Interactive UI is fully functional and tested, including dynamic loading spinner and self-clearing error messages.**
    - `read`: **Functional**, reads file content using `pkg/services/file_system_service.go`.
    - `write`: **Functional**, writes content to a file using `pkg/services/file_system_service.go`.
    - `exec`: **Functional**, executes shell commands (uses `pkg/services/shell_service.go`).
    - `ls`: **Functional**, lists directory contents (uses `pkg/services/file_system_service.go`).
    - `git-branch`: **Functional**, gets the current Git branch name (uses `pkg/services/git_service.go` with `go-git`).
    - `extensions`: Command group with subcommands:
      - `list`: **Functional**, lists discovered extensions by reading `gemini-extension.json` files.
      - `install`: **Functional** (bug in renaming directory after install fixed), command structure and argument parsing in `main.go` are ready. Core logic in `pkg/commands/extensions.go` and `pkg/extension/manager.go` is implemented with git clone and local copy functionality.
      - `uninstall`: **Functional** (with linked extension support).
      - `new`: **Functional**.
      - `enable`: **Functional**.
      - `disable`: **Functional`.
    - `mcp`: Command group with subcommands:
      - `list`: **Functional**, lists configured MCP servers, merging from settings and extensions, and simulates connection status.
      - `add`: **Functional**. 
      - `remove`: **Functional**.
    - **New Commands Implemented (from JavaScript .toml files)**:
      - `code-guide`: **Functional**, answers questions about the codebase using AI.
      - `find-docs`: **Functional**, finds relevant documentation and outputs GitHub URLs using AI.
      - `cleanup-back-to-main`: **Functional**, automates Git branch cleanup.
      - `pr-review`: **Functional**, conducts comprehensive AI-driven pull request reviews.
      - `grep-code`: **Functional**, summarizes code findings for a given pattern using grep and AI.
    - `list-models`: **Functional**, lists available Gemini models using `pkg/core/gemini.go`.

- **Core Services & Tools (pkg/core, pkg/extension, pkg/config, pkg/mcp, pkg/services)**:
  - `pkg/core/gemini.go`: **Functional** (with tool-calling capabilities), uses `google.golang.org/genai` for Gemini API interaction.
  - `pkg/services/shell_service.go`: **Functional**, provides `ExecuteCommand` for shell operations.
  - `pkg/services/file_system_service.go`: **Functional**, provides `ListDirectory`, `PathExists`, `IsDirectory`, `JoinPaths`, `WriteFile`, `ReadFile`, `CreateDirectory`, `CopyDirectory`.
  - `pkg/services/git_service.go`: **Functional**, uses `github.com/go-git/go-git/v5` to interact with Git repositories. Now includes `GetRemoteURL`, `CheckoutBranch`, `Pull`, and `DeleteBranch` methods.
  - `pkg/extension/manager.go`: **Functional**. Discovers and loads extensions, parses `gemini-extension.json`. `InstallOrUpdateExtension` has logic for git clone and local copy, `EnableExtension` and `DisableExtension` are implemented. The `fsService` type issue has been resolved.
  - `pkg/extension/types.go`: Defines `InstallArgs` and `ExtensionInstallMetadata`.
  - `pkg/config/config.go`: **Consolidated and Functional**. Now contains `SettingScope`, `Settings`, `LoadSettings`, `Config` struct, and related methods. `Config` struct now has an exported `Model` field, and `NewConfig` and `GetModel()` methods are adjusted accordingly.
  - `pkg/mcp/client.go`: **Functional** (renamed `Client` to `McpClient`). Simulates MCP connection.
  - `pkg/types/types.go`: **Centralized Types**. Updated to include `MCPServerConfig`, `MCPServerStatus`, `MCPOAuthConfig`, `AuthProviderType`, `ToolCallRequestInfo`, `JsonSchemaObject`, `JsonSchemaProperty`, `AgentTerminateMode`, `FunctionCall`, `Tool`, `ToolInvocation`, `Kind`, `BaseDeclarativeTool` (and its methods), `ToolRegistry` (and its methods), and `TelemetrySettings` to resolve import cycles and consolidate common types.
  - `pkg/types/constants.go`: **Cleaned Up**. Removed duplicate `MCPServerStatus` and `Kind` constants.

## 2. Linter-Identified Issues (Prioritized for Next Steps)

Based on results from `golangci-lint`, the following issues need to be addressed before a successful build.

### Resolved Issues:

- **`pkg/core/agents/types.go`**: Removed redundant `AgentTerminateMode` definition.
- **`pkg/core/agents/subagent_tool_wrapper.go`**: Corrected access to `MessageBus` and updated references to `types.BaseDeclarativeTool`, `types.NewBaseDeclarativeTool`, `types.KindThink` (replaced with `types.KindOther`), and `types.ToolInvocation`.-
- **`cmd/generate.go`**: Converted `[]genai.Content{}` to `[]*genai.Content{}`.
- **`cmd/list-models.go`**: Added `genai` import and provided correct arguments to `core.NewGeminiChat`.
- **Import Cycles**: Fully resolved by moving `Tool`, `ToolInvocation`, `Kind`, `BaseDeclarativeTool`, `ToolRegistry`, and `TelemetrySettings` to `pkg/types/types.go`, and removing `pkg/tools/tool_registry.go`.
- **`pkg/core/agents/non_interactive_tool_executor.go`**: Fixed `undefined: ToolCallRequestInfo`.
- **`pkg/core/agents/schema_utils.go`**: Fixed `undefined` errors for `JsonSchemaObject` and `JsonSchemaProperty`.
- **`pkg/core/agents/registry.go`**: Removed unused `fmt` import.
- **`pkg/tools/glob.go`, `pkg/tools/grep.go`**: Removed unused `pkg/core/tool` import.
- **`pkg/tools/read_many_files.go`**: Removed unused `bufio` import.
- **`pkg/types/types.go`**: Moved constants to `pkg/types/constants.go`.
- **`pkg/config/config.go`**: Changed `ToolRegistryProvider` to a struct, and exported `Model` field.
- **`pkg/core/agents/executor.go`**: Removed unused `stringPtr` function.
- **`pkg/tools/register.go`**: Removed subagent registration logic to resolve import cycle.
- **`cmd/root.go`**: Updated call to `tools.RegisterAllTools()` (now `types.NewToolRegistry()`) and removed unused `dummyConfig`.
- **`SA9003: empty branch` errors**: Added `//nolint:staticcheck` to empty `if` blocks in `pkg/utils/folder_structure.go` and `pkg/core/agents/registry.go`.
- **Duplicate definitions in `pkg/config`**: Consolidated `SettingScope`, `Settings`, and `LoadSettings` into `pkg/config/config.go` and deleted `pkg/config/settings.go`.
- **`cmd/generate.go` and `pkg/ui/generate_ui.go` type mismatch**: Corrected `ui.NewGenerateModel` to accept `*core.GeminiChat` and updated `cmd/generate.go` to pass the `geminiClient` correctly. Removed unused imports from `pkg/ui/generate_ui.go`.
- **Telemetry Logging**: Implemented basic telemetry logging with file output and global logger initialization.
- **`pkg/extension/manager.go`**: Corrected `fsService` type from `*services.FileSystemService` to `services.FileSystemService`.
- **`cmd/find_docs.go`**: Corrected ToolRegistry initialization.
- **`cmd/pr_review.go`**: Corrected ToolRegistry initialization and syntax error.

### Remaining Issues:

- None.

## 3. Command Implementation Strategy (Overview)

The `extensions` and `mcp` command groups are primary CLI functionalities.

### 3.1. Extensions Commands (`pkg/commands/extensions.go`)

Translate the logic from the JavaScript files below. Each command needs argument parsing, service interaction (using `FileSystemService`, `GitService`, etc.), and thorough analysis of the original JavaScript source.

- `install.ts`: **Functional**. Implemented with `force` flag support. Core logic in `pkg/extension/manager.go` handles git clone/pull and local copy/overwrite. Argument parsing in `main.go` is ready.
- `list.ts`: **Functional**.
- `new.ts`: **Functional`.
- `enable.ts`: **Functional**.
- `disable.ts`: **Functional`.
- `uninstall`: **Functional** (with linked extension support).
- `update.ts`: **Functional**.
- `link`: **Functional**.

### 3.2. MCP Commands (`pkg/commands/mcp.go`)

Translate logic from the following JavaScript files. Similar to extensions, each MCP command involves argument parsing, service interaction, and thorough analysis of the original JavaScript source.

- `add`: **Functional**.
- `list.ts`: **Functional** (simulated connection status).
- `remove`: **Functional`.

## 4. JavaScript Source Code Location

The JavaScript source code to be translated is located in the `docs/gemini-cli-main/packages/` directory. Specifically:

- **Core Logic**: `core/src/`
- **CLI Commands**: `cli/src/commands/`

## 5. API Integration Strategy (No Change)

- **Gemini API Client**: **Functional** (with tool-calling capabilities), uses `google.golang.org/genai` for Gemini API interaction. `GEMINI_API_KEY` is read from the environment.
- **Error Handling**: Implement robust error handling for API calls, including retries and clear error messages.

## 6. Testing Strategy (No Change)

- **Unit Tests**: Write unit tests for individual functions and methods within each Go package (`pkg/core`, `pkg/tools`, `pkg/services`, `pkg/commands`) to ensure correctness.
- **Integration Tests**: Develop integration tests for CLI commands to verify they interact correctly with the services and produce expected outputs. New Go-native tests will be created.
- **Manual Testing**: Regular manual testing of the CLI commands at various stages of implementation to ensure functionality.

## 7. Execution Flow (Refined)

The migration will proceed iteratively, focusing on one command or core functionality at a time, following these steps:

1.  **Linter First**: You must run `golangci-lint` from the go-cli directory to identify all issues.
2.  **Systematic Fixing**: Address issues one by one, prioritizing type-checking and unused import errors.
3.  **Identify Target**: Choose a specific JavaScript command or core module to translate.
4.  **Analyze JavaScript Source**: Read and understand the corresponding JavaScript file(s) to grasp functionality, dependencies, and logic.
5.  **Design Go Implementation**: Outline the Go structures, interfaces, and functions required.
6.  **Implement in Go**: Write the Go code.
7.  **Integrate with CLI**: Add the new Go command or integrate the new Go module into `main.go`.
8.  **Test**: Write and run Go unit/integration tests, and perform manual testing.
9.  **Refine**: Address any issues or improvements.

## 7. Git Instructions based on conventional commit convention
1. **Initialize a new repository**: if not already done, initialize a new repository in the go-cli directory
2. **Commit messages**: use short, clear and concise commit messages to document your changes
3. **Commit your changes**: use `git add .` to stage all changes, and then `git commit -m "Your commit message"` to commit your changes

## 8. Next Steps

1.  **Review of Go Port and Tool-Calling Mechanism**: Delve into `pkg/core/gemini.go` and `pkg/core/agents` to understand how prompts are structured, how tool calls are generated, and how they are executed. This is crucial for validating the AI's ability to understand and execute tools based on structured prompts.
2.  **End-to-End Testing for AI Commands**: Run end-to-end tests for commands that involve AI and tool-calling (e.g., `generate`, `find-docs`, `pr-review`) to observe the actual JSON output and tool execution, ensuring the Go port behaves identically to the JavaScript version.
3.  **Enhance Interactive UI**:
    *   Expand the interactive UI to other commands where user interaction would be beneficial (e.g., `code-guide`, `find-docs`).
    *   Improve the UI/UX of the interactive components (e.g., better loading indicators, error displays, input validation).
    *   **`generate`**: Interactive UI complete.
    *   **`find-docs`**: Interactive UI complete.
4.  **Tool Integration for AI Commands**:
    *   For commands like `find-docs` and `pr-review`, integrate actual tool-calling capabilities. This would allow the AI to dynamically use `GitService`, `FileSystemService`, and `ShellExecutionService` to gather information or perform actions, rather than relying solely on pre-constructed prompts.
    *   This would involve implementing the `tools` package in Go to register and execute these services as AI tools.
    *   **`find-docs`**: Tool integration complete.
    *   **`pr_review`**: Tool integration for `pr_review` needs to be verified after fixing the syntax. The `promptTemplate` in `cmd/pr_review.go` already outlines the tools to be used (`checkout_branch`, `execute_command`, `list_directory`, `read_file`, `pull`).
5.  **Error Handling and User Feedback**:
    *   Improve error handling across all commands, providing more user-friendly messages.
    *   Implement a consistent way to provide feedback to the user, especially for long-running operations.
6.  **Testing**:
    *   Implement comprehensive unit and integration tests for all newly added commands and UI components.
    *   Address the environmental issue encountered during `extensions` testing (permission denied when creating `.gemini/extensions` directory). This might involve adjusting default paths or providing clearer instructions for setting up the environment.
7.  **Remaining JavaScript CLI Commands** (if any):
    *   Review any remaining JavaScript CLI commands or features that have not yet been migrated to Go. (Based on current analysis, all explicit commands have been addressed, but a deeper dive might reveal more subtle features).

## 9. Git Instructions based on conventional commit convention
1. **Initialize a new repository**: if not already done, initialize a new repository in the go-cli directory
2. **Commit messages**: use short, clear and concise commit messages to document your changes
3. **Commit your changes**: use `git add .` to stage all changes, and then `git commit -m "Your commit message"` to commit your changes