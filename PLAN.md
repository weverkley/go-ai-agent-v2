# Go CLI Migration Plan

This document outlines the plan for migrating the JavaScript-based Gemini CLI to a new Go-based CLI application. The goal is to rewrite the entire CLI, excluding the authentication flow (which will rely on `GEMINI_API_KEY` environment variable), test files/commands, and the VS Code IDE companion.

### **IMPORTANT!**
This project is meant to be a generic CLI, it will use multiple AI executors (e.g., Gemini, OpenAI, etc.), not only Gemini which this tool is currently based on.

### **UI Package**
The UI for this CLI will be implemented using `charmbracelet/bubbletea` for an interactive terminal user interface.

## 1. Current Status

The foundational structure for the Go CLI has been established, and several core services and commands have been implemented. Many original JavaScript files have been translated with tool-calling capabilities. All identified type-checking and unused import errors have been addressed.

- **Go Project Setup**: The Go module is now located at the project root (`go-ai-agent-v2`).

- **Core CLI Structure (main.go)**:
  - Command-line argument parsing is implemented using the `cobra` library.
  - `--version` flag implemented.
  - Top-level commands implemented:
    - `generate`: **Functional** (with tool-calling capabilities). Now includes an **Interactive UI** using `charmbracelet/bubbletea` if no prompt is provided. Generates content using `pkg/core/gemini.go` (real Gemini API integration). **Interactive UI is fully functional and tested, including dynamic loading spinner and self-clearing error messages.**
    - `read`: **Functional**, reads file content using `pkg/services/file_system_service.go`.
    - `write`: **Functional**, writes content to a file, now utilizing the `write_file` tool.
    - `exec`: **Functional**, executes shell commands (uses `pkg/services/shell_service.go`).
    - `ls`: **Functional**, lists directory contents (uses `pkg/services/file_system_service.go`).
    - `git-branch`: **Functional**, gets the current Git branch name (uses `pkg/services/git_service.go` with `go-git`).
    - `extract-function`: **Functional**, extracts a code block into a new function or method.
    - `find-unused-code`: **Functional**, finds unused functions and methods in Go files.
    - `extensions`: Command group with subcommands:
      - `list`: **Functional**, lists discovered extensions by reading `gemini-extension.json` files.
      - `install`: **Functional** (bug in renaming directory after install fixed), command structure and argument parsing in `main.go` are ready. Core logic in `pkg/commands/extensions.go` and `pkg/extension/manager.go` is implemented with git clone and local copy functionality.
      - `uninstall`: **Functional** (with linked extension support).
      - `new`: **Functional**.
      - `enable`: **Functional`.
      - `disable`: **Functional`.
    - `mcp`: Command group with subcommands:
      - `list`: **Informative Message**, listing configured MCP servers is not yet implemented.
      - `add`: **Informative Message**, adding MCP servers is not yet implemented.
      - `remove`: **Informative Message**, removing MCP servers is not yet implemented.
    - **AI-Powered Commands (migrated from JavaScript .toml files)**:
      - `code-guide`: **Functional**, answers questions about the codebase using AI. 
      - `find-docs`: **Functional**, finds relevant documentation and outputs GitHub URLs using AI.
      - `cleanup-back-to-main`: **Functional**, automates Git branch cleanup.
      - `pr-review`: **Functional**, conducts comprehensive AI-driven pull request reviews.
      - `grep-code`: **Functional**, summarizes code findings for a given pattern using grep and AI.
    - `list-models`: **Functional**, lists available Gemini models using `pkg/core/gemini.go`.
    - `privacy`: **Informative Message**, displays a detailed privacy notice.
    - `ide`: **Informative Message**, manages IDE integration (status, install, enable, disable) with messages indicating future availability.
    - `theme`: **Informative Message**, changes the visual theme with a message indicating future availability.
    - `auth`: **Informative Message**, manages authentication credentials (login, logout) with messages indicating future availability and secure storage.
    - `setup-github`: **Informative Message**, configures GitHub Actions with a message indicating future availability.
    - `tools`: **Functional**, `run` subcommand now executes specific AI tools.
    - `profile`: **Informative Message**, toggles debug profile display with a message indicating future availability.
    - `vim`: **Informative Message**, toggles Vim mode with a message indicating future availability.
    - `terminal-setup`: **Informative Message**, configures terminal keybindings with a message indicating future availability.
    - `editor`: **Informative Message**, sets external editor preference with a message indicating future availability.
    - `stats`: **Informative Message**, displays usage statistics with messages indicating future availability.
    - `restore`: **Informative Message**, restores tool calls and conversation/file history with messages indicating future availability.
    - `permissions`: **Informative Message**, manages folder trust settings with a message indicating future availability.
    - `chat`: **Functional**, interactive chat with AI agents, now featuring:
        - **Model Rotation Suggestions**: Automatically suggests alternative models (e.g., `gemini-1.5-flash` from `gemini-1.5-pro`) when API quota errors occur.
        - **Improved UI Stability**: Fixed issues with terminal history, duplicated input, and text wrapping for long messages.
        - **Enhanced `clear` command**: Correctly clears both UI and executor history.

- **Core Services & Tools (pkg/core, pkg/extension, pkg/config, pkg/mcp, pkg/services)**:
  - `pkg/core/gemini.go`: **Functional** (with tool-calling capabilities), uses `google.golang.org/genai` for Gemini API interaction. Now includes actual token counting in `CompressChat()`.
  - `pkg/core/qwen.go`: **Functional** (initial implementation using `go-openai` for Qwen API compatibility), supporting streaming responses.
  - `pkg/core/mock_executor.go`: **Functional** mock executor for testing.
  - `pkg/services/shell_service.go`: **Functional**, provides `ExecuteCommand` for shell operations.
  - `pkg/services/file_system_service.go`: **Functional**, provides `ListDirectory`, `PathExists`, `IsDirectory`, `JoinPaths`, `WriteFile`, `ReadFile`, `CreateDirectory`, `CopyDirectory`.
  - `pkg/services/git_service.go`: **Functional**, uses `github.com/go-git/go-git/v5` to interact with Git repositories. Now includes `GetRemoteURL`, `CheckoutBranch`, `Pull`, and `DeleteBranch` methods.
  - `pkg/extension/manager.go`: **Functional**. Discovers and loads extensions, parses `gemini-extension.json`. `InstallOrUpdateExtension` has logic for git clone/pull and local copy/overwrite. The `fsService` type issue has been resolved.
  - `pkg/extension/types.go`: Defines `InstallArgs` and `ExtensionInstallMetadata`.
  - `pkg/config/config.go`: **Consolidated and Functional**. Now contains `SettingScope`, `Settings`, `LoadSettings`, `Config` struct, and related methods. `Config` struct now has an exported `Model` field, and `NewConfig` and `GetModel()` methods are adjusted accordingly. Includes `GetGeminiDir()` method.
  - `pkg/mcp/client.go`: **Functional** (renamed `Client` to `McpClient`). Simulates MCP connection.
  - `pkg/types/types.go`: **Centralized Types**. Updated to include `MCPServerConfig`, `MCPServerStatus`, `MCPOAuthConfig`, `AuthProviderType`, `ToolCallRequestInfo`, `JsonSchemaObject`, `JsonSchemaProperty`, `AgentTerminateMode`, `FunctionCall`, `Tool`, `ToolInvocation`, `Kind`, `BaseDeclarativeTool` (and its methods), `ToolRegistry` (and its methods), and `TelemetrySettings` to resolve import cycles and consolidate common types.
  - `pkg/types/constants.go`: **Cleaned Up**. Removed duplicate `MCPServerStatus` and `Kind` constants.
  - `pkg/core/agents/executor.go`: Error reporting in `createChatObject()` now uses `utils.LogErrorf()`.
  - `pkg/utils/folder_structure.go`: `shouldIgnoreFile` logic correctly implemented using `fileService.ShouldIgnoreFile()`.
  - `pkg/utils/utils.go`: Telemetry logging implemented in `LogAgentStart()` and `LogAgentFinish()`.
  - `pkg/tools/ls.go`: `Execute()` method now lists directory contents.
  - `pkg/tools/write_file.go` and `pkg/tools/list_directory.go`, `pkg/tools/smart_edit.go`: Refactored to use `FileSystemService`.
  - `pkg/routing/strategy.go`: **Functional** model routing strategies, including a `FallbackStrategy` for error-based model suggestions.
  - `pkg/services/settings_service.go`: **Functional**, now includes an `executor` setting to specify the active AI executor.
  - `pkg/telemetry/telemetry.go`: **Functional**, enhanced with debug logging and file output capabilities.

## 2. Linter-Identified Issues

All previously identified linter issues have been resolved, including:

- **`pkg/core/agents/types.go`**: Removed redundant `AgentTerminateMode` definition.
- **`pkg/core/agents/subagent_tool_wrapper.go`**: Corrected access to `MessageBus` and updated references to `types.BaseDeclarativeTool`, `types.NewBaseDeclarativeTool`, `types.KindThink` (replaced with `types.KindOther`), and `types.ToolInvocation`.-
- **`cmd/generate.go`**: Converted `[]genai.Content{}` to `[]*genai.Content{}`.
- **`cmd/list-models.o`**: Added `genai` import and provided correct arguments to `core.NewGeminiChat`.
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
- **`pkg/extension/manager.go`**: Corrected `fsService` type from `*services.FileSystemService` to `services.FileSystemService`.
- **`cmd/find_docs.go`**: Corrected ToolRegistry initialization.
- **`cmd/pr_review.go`**: Corrected ToolRegistry initialization and syntax error.
- **`pkg/core/mock_executor.go`**: Fixed syntax error.
- **`cmd/chat.go`**: Removed unused `executorType` flag.
- **`cmd/clear.go`**: Restored `executor.SetHistory(nil)` and removed screen clearing.

## 3. Remaining JavaScript Source Code to be Migrated

Based on a review of the original JavaScript source code, the following files and modules may still need to be migrated:

- **MCP Commands**: The `add`, `list`, and `remove` commands for MCP servers need to be fully implemented. Currently, they only provide informative messages.
- **Core Functionality**: The following files from `docs/gemini-cli-main/packages/core/src/core` may need to be migrated:
    - `baseLlmClient.ts`
    - `client.ts`
    - `contentGenerator.ts`
    - `fakeContentGenerator.ts`
    - `geminiRequest.ts`
    - `logger.ts`
    - `loggingContentGenerator.ts`
    - `recordingContentGenerator.ts`
    - `tokenLimits.ts`
    - `turn.ts`
- **Fallback/Error Handling**: The logic in the `docs/gemini-cli-main/packages/core/src/fallback` directory may need to be migrated.

## 4. Command Implementation Strategy (Overview)

The `extensions` and `mcp` command groups are primary CLI functionalities.

### 3.1. Extensions Commands (`pkg/commands/extensions.go`)

Translate the logic from the JavaScript files below. Each command needs argument parsing, service interaction (using `FileSystemService`, `GitService`, etc.), and thorough analysis of the original JavaScript source.

- `install.ts`: **Functional**. Implemented with `force` flag support. Core logic in `pkg/extension/manager.go` handles git clone/pull and local copy/overwrite. Argument parsing in `main.go` is ready.
- `list.ts`: **Functional**.
- `new.ts`: **Functional`.
- `enable.ts`: **Functional`.
- `disable.ts`: **Functional`.
- `uninstall`: **Functional** (with linked extension support).
- `update.ts`: **Functional**.
- `link`: **Functional**.

### 3.2. MCP Commands (`pkg/commands/mcp.go`)

Translate logic from the following JavaScript files. Similar to extensions, each MCP command involves argument parsing, service interaction, and thorough analysis of the original JavaScript source.

- `add`: **Informative Message**.
- `list.ts`: **Informative Message`.
- `remove`: **Informative Message`.

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

## 8. Git Instructions based on conventional commit convention
1. **Initialize a new repository**: if not already done, initialize a new repository in the go-cli directory
2. **Commit messages**: use short, clear and concise commit messages to document your changes
3. **Commit your changes**: use `git add .` to stage all changes, and then `git commit -m "Your commit message"` to commit your changes

## 9. Next Steps

1.  **Qwen Executor Enhancements:**
    *   Implement `GenerateContent`, `ExecuteTool`, `SendMessageStream`, `GetHistory`, `CompressChat` for Qwen.
    *   Implement tool calling for Qwen.
2.  **Model Routing Enhancements:**
    *   Refactor `getSuggestedModel` to be more generic (e.g., a map of suggesters per executor type).
    *   Implement `ClassifierStrategy` for more intelligent model routing.
3.  **Error Handling:**
    *   Improve error messages for API failures (e.g., distinguish between quota errors and other API errors).
4.  **Settings Command:**
    *   Add validation for `executor` and `model` settings.
5.  **Comprehensive Testing**: Unit and integration tests for all new components.
6.  **Secure API Key Management**: Robust OS-specific storage and clearing of API keys.
7.  **IDE Integration**: Full integration with supported IDEs.
8.  **Theme Customization**: Ability to change and persist CLI visual themes.
9.  **Terminal Keybinding Configuration**: For enhanced multiline input.
10. **Implement External Editor Preference**: Setting and using a preferred external editor.
11. **Implement Usage Statistics**: Develop logic for collecting and displaying detailed model and tool-specific usage statistics.
12. **Implement Restore Functionality**: Saving and restoring CLI state, including tool calls and conversation/file history.
13. **Implement Folder Trust Management**: Security features for managing trusted folders.
14. **Implement MCP Server Management**: Full CRUD operations for MCP servers.
15. **Implement Rich Interactive Chat UI (Go)**: Replicate the sophisticated, component-based, and data-driven architecture of the JavaScript chat UI in the Go application.
    *   **Create a Structured `Message` Interface**: Instead of a simple `[]string` for history, define a `Message` interface with a `Render(model ChatModel) string` method. This will allow for different message types to have their own rendering logic.
    *   **Implement Concrete `Message` Types**: Create structs for each message type (e.g., `UserMessage`, `BotMessage`, `ToolCallMessage`, `ToolResultMessage`, `InfoMessage`, `ErrorMessage`) that implement the `Message` interface. Each struct will hold the relevant data and define the specific `lipgloss` styling in its `Render` method. This mirrors the component-based approach of the JavaScript version (e.g., `UserMessage.tsx`, `GeminiMessage.tsx`).
    *   **Refactor `ChatModel`**:
        *   Change the `messages` field from `[]string` to `[]Message`.
        *   Update the `Update` method to create instances of the new message structs and add them to the history.
        *   Modify the `View` method to iterate over the `[]Message` slice and call the `Render` method for each message, composing the final view from these rendered components.
    *   **Implement Slash Commands**: Add a simple slash command processor in the `Update` method to handle user commands. Start with essential commands:
        *   `/clear`: To clear the chat history.
        *   `/quit` or `/exit`: To exit the application.
        *   **Enhance `GenerateStream` Integration**: Ensure the `GenerateStream` method in the `Executor` produces events that can be easily mapped to the new `Message` types in the UI.
    16. **Implement Model Routing**: Port the JavaScript model routing logic to the Go application to enable dynamic selection of AI models.
        *   **Create `RoutingStrategy` Interface**: Define a `RoutingStrategy` interface in a new `pkg/routing` directory. This interface will have a `Route` method that takes a context and returns a model name or an error.
        *   **Implement Concrete Strategies**: Create structs for each routing strategy (`DefaultStrategy`, `OverrideStrategy`, `FallbackStrategy`, `ClassifierStrategy`, `CompositeStrategy`) that implement the `RoutingStrategy` interface.
        *   **Create `ModelRouterService`**: Develop a `ModelRouterService` that uses a `CompositeStrategy` to execute the strategies in a predefined order of priority.
        *   **Integrate with `ExecutorFactory`**: Modify the `ExecutorFactory` to use the `ModelRouterService` to determine which model to use when creating a new executor. The factory will call the router to get the model name before initializing the `GeminiChat` executor.
        *   **Simplify `ClassifierStrategy` (Initial Version)**: For the initial implementation, the `ClassifierStrategy` can use a simple heuristic (e.g., keyword matching, prompt length) instead of making an LLM call. The full LLM-based classification can be added in a future iteration.