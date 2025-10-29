# Go CLI Migration Plan

This document outlines the plan for migrating the JavaScript-based Gemini CLI to a new Go-based CLI application. The goal is to rewrite the entire CLI, excluding the authentication flow (which will rely on `GEMINI_API_KEY` environment variable), test files/commands, and the VS Code IDE companion.

## 1. Current Status

The foundational structure for the Go CLI has been established, and several core services and commands have been implemented.

- **Go Project Setup**: An empty Go module (`go-ai-agent-v2/go-cli`) has been initialized.
- **Core CLI Structure (main.go)**:
  - Command-line argument parsing is implemented using the `cobra` library.
  - `--version` flag implemented.
  - Top-level commands implemented:
    - `generate`: **Functional**, generates content using `pkg/core/gemini.go` (real Gemini API integration).
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
      - `enable`: **Placeholder**.
      - `disable`: **Placeholder**.
    - `mcp`: Command group with subcommands:
      - `list`: **Functional**, lists configured MCP servers, merging from settings and extensions, and simulates connection status.
    - `list-models`: **Functional**, lists available Gemini models using `pkg/core/gemini.go`.
- **Core Services & Tools (pkg/core, pkg/extension, pkg/config, pkg/mcp, pkg/services)**:
  - `pkg/core/gemini.go`: **Functional**, uses `google.golang.org/genai` for Gemini API interaction.
  - `pkg/services/shell_service.go`: **Functional**, provides `ExecuteCommand` for shell operations.
  - `pkg/services/file_system_service.go`: **Functional**, provides `ListDirectory`, `PathExists`, `IsDirectory`, `JoinPaths`, `WriteFile`, and `ReadFile`.
  - `pkg/services/git_service.go`: **Functional**, uses `github.com/go-git/go-git/v5` to interact with Git repositories.
  - `pkg/extension/manager.go`: **Partially Implemented**. Discovers and loads extensions, parses `gemini-extension.json`. `InstallOrUpdateExtension` has logic for git clone and local copy, but `EnableExtension` and `DisableExtension` are placeholders.
  - `pkg/extension/types.go`: Defines `InstallArgs` and `ExtensionInstallMetadata`.
    *   `pkg/config/settings.go`: **Functional** (extended functionality). Loads default settings, can read/write settings.json, and includes fields for DebugMode, UserMemory, ApprovalMode, ShowMemoryUsage, TelemetryEnabled, Model, and Proxy.
  - `pkg/mcp/types.go`: Defines `MCPServerStatus`, `MCPServerConfig`, `MCPOAuthConfig`, and `AuthProviderType`.
  - `pkg/mcp/client.go`: **Placeholder**. Simulates MCP connection.

## 2. Remaining Core Components for Translation

Based on the analysis of `gemini-cli-main/packages/core/src/index.ts`, the following significant modules/functionalities still need to be translated into Go. Prioritization will depend on the specific CLI commands being implemented.

- **AI Agent Logic (`pkg/core/agents`)**: Translation of the AI agent components that drive the intelligent behavior of the CLI.
  - `codebase-investigator.ts`
  - `executor.ts`
  - `invocation.ts`
  - `registry.ts`
  - `schema-utils.ts`
  - `subagent-tool-wrapper.ts`
  - `types.ts`
  - `utils.ts`
- **Code Assist (`pkg/core/code_assist`)**: Functionality related to code generation, completion, or analysis.
  - `codeAssist.ts`
  - `converter.ts`
  - `oauth-credential-storage.ts`
  - `oauth2.ts`
  - `server.ts`
  - `setup.ts`
  - `types.ts`
- **Prompts (`pkg/core/prompts`)**: Management and rendering of prompts for AI models or user interactions.
  - `mcp-prompts.ts`: **Functional** (basic prompt management with `DiscoveredMCPPrompt` support and `GetPromptsByServer` method).
  - `prompt-registry.ts`: **Functional** (basic prompt management with `DiscoveredMCPPrompt` support).
- **Tools (other than file I/O)**: Specialized tools such as `grep`, `glob`, `web-fetch`, `memoryTool`, `web-search`, `read-many-files`, etc., need to be implemented.
  - `diffOptions.ts`
  - `edit.ts`
  - `glob.ts`: **Functional**.
  - `grep.ts`: **Functional**.
  - `ls.ts`
  - `mcp-client-manager.ts`
  - `mcp-client.ts`
  - `mcp-tool.ts`
  - `memoryTool.ts`: **Functional**.
  - `modifiable-tool.ts`
  - `read-file.ts`: **Functional**.
  - `read-many-files.ts`: **Functional**.
  - `ripGrep.ts`: **Functional** (implemented by `grep` tool).
  - `shell.ts`: **Functional** (implemented by `exec` command).
  - `smart-edit.ts`
  - `tool-error.ts`
  - `tool-names.ts`
  - `tool-registry.ts`
  - `tools.ts`
  - `web-fetch.ts`: **Functional**.
  - `web-search.ts`: **Functional**.
  - `write-file.ts`: **Functional** (implemented by `write` command).
  - `write-todos.ts`
- **Config (`pkg/core/config`)**: Robust configuration management beyond just reading an environment variable (e.g., loading from files).
  - `config.ts`
  - `constants.ts`
  - `flashFallback.ts`
  - `models.ts`
  - `storage.ts`
- **Output (`pkg/core/output`)**: Handling rich output formatting and display, potentially adapting existing JavaScript formatting logic.
  - `json-formatter.ts`
  - `stream-json-formatter.ts`
  - `types.ts`
- **Policy (`pkg/core/policy`)**: Implementation of any policy enforcement or decision-making logic.
  - `index.ts`
  - `policy-engine.ts`
  - `stable-stringify.ts`
  - `types.ts`
- **Confirmation Bus (`pkg/core/confirmation-bus`)**: System for handling user confirmations or asynchronous operations.
  - `index.ts`
  - `message-bus.ts`
  - `types.ts`
- **IDE Integration (`pkg/core/ide`)**: (Lower priority, as user explicitly excluded VS Code companion, but might include generic IDE-agnostic features).
  - `constants.ts`
  - `detect-ide.ts`
  - `ide-client.ts`
  - `ide-installer.ts`
  - `ideContext.ts`
  - `process-utils.ts`
  - `types.ts`
- **Telemetry (`pkg/core/telemetry`)**: Implementation of usage data collection.
  - `activity-detector.ts`
  - `activity-monitor.ts`
  - `activity-types.ts`
  - `config.ts`
  - `constants.ts`
  - `file-exporters.ts`
  - `gcp-exporters.ts`
  - `high-water-mark-tracker.ts`
  - `index.ts`
  - `loggers.ts`
  - `memory-monitor.ts`
  - `metrics.ts`
  - `rate-limiter.ts`
  - `sdk.ts`
  - `semantic.ts`
  - `telemetry-utils.ts`
  - `telemetry.ts`
  - `telemetryAttributes.ts`
  - `tool-call-decision.ts`
  - `trace.ts`
  - `types.ts`
  - `uiTelemetry.ts`

## 3. Command Implementation Strategy

The `extensions` and `mcp` command groups are the primary CLI functionalities to be translated.

### 3.1. Extensions Commands (`pkg/commands/extensions.go`)

Translate the logic from the following JavaScript files into Go:

- `install.ts`: **In Progress**. Core logic in `pkg/extension/manager.go` needs to be completed (git clone/local copy). Argument parsing in `main.go` is ready.
- `list.ts`: **Functional**.
- `new.ts`: **Functional**.
- `enable.ts`: **Placeholder**.
- `disable.ts`: **Placeholder**.
      - `uninstall`: **Functional** (with linked extension support).
      - `update.ts`: **Functional**.
      - `link`: **Functional**.

Each command will require:
1.  **Argument Parsing**: Define flags specific to the subcommand.
2.  **Service Interaction**: Utilize the core services (`FileSystemService`, `GitService`, etc.) as needed.
3.  **JavaScript File Analysis**: For each command, the corresponding JavaScript file will be read and analyzed to understand its functionality and internal logic before translation to Go.

### 3.2. MCP Commands (`pkg/commands/mcp.go`)

Translate the logic from the following JavaScript files into Go:

      - `add`: **Functional**.
      - `list.ts`: **Functional** (simulated connection status).
      - `remove`: **Functional**.

Similar to extensions, each MCP command will involve argument parsing, service interaction, and thorough analysis of the original JavaScript source.

## 4. JavaScript Source Code Location

The JavaScript source code to be translated is located in the `docs/gemini-cli-main/packages/` directory. Specifically:

- **Core Logic**: `core/src/`
- **CLI Commands**: `cli/src/commands/`

## 4. Next Steps

1.  **Implement Core Components**: Begin translating core components like `config`, `prompts`, and other tools.
2.  **Testing**: Write unit and integration tests for the new features.

## 5. API Integration Strategy

- **Gemini API Client**: **Functional**, uses `google.golang.org/genai` for Gemini API interaction. `GEMINI_API_KEY` is read from the environment.
- **Error Handling**: Implement robust error handling for API calls, including retries and clear error messages.

## 6. Testing Strategy

- **Unit Tests**: Write unit tests for individual functions and methods within each Go package (`pkg/core`, `pkg/tools`, `pkg/services`, `pkg/commands`) to ensure correctness.
- **Integration Tests**: Develop integration tests for CLI commands to verify they interact correctly with the services and produce expected outputs. Since the user explicitly excluded rewriting test _files_, new Go-native tests will be created.
- **Manual Testing**: Regular manual testing of the CLI commands at various stages of implementation to ensure functionality.

## 7. Execution Flow

The migration will proceed iteratively, focusing on one command or core functionality at a time, following these steps:

1.  **Identify Target**: Choose a specific JavaScript command or core module to translate.
2.  **Analyze JavaScript Source**: Read and understand the corresponding JavaScript file(s) to grasp functionality, dependencies, and logic.
3.  **Design Go Implementation**: Outline the Go structures, interfaces, and functions required.
4.  **Implement in Go**: Write the Go code.
5.  **Integrate with CLI**: Add the new Go command or integrate the new Go module into `main.go`.
6.  **Test**: Write and run Go unit/integration tests, and perform manual testing.
7.  **Refine**: Address any issues or improvements.

This plan provides a roadmap for the Go CLI migration. I am now ready for your instructions on which specific command or core functionality to tackle next.
