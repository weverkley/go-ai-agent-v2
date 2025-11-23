# Go CLI Migration Plan

This document outlines the plan and current status of the Go AI Agent CLI. The goal is to create a generic, multi-executor CLI for advanced AI and agentic workflows.

### **IMPORTANT!**
This project is meant to be a generic CLI; it will use multiple AI executors (e.g., Gemini, OpenAI, etc.), not only Gemini which this tool is currently based on.

### **UI Package**
The UI for this CLI is implemented using `charmbracelet/bubbletea` for a rich, interactive terminal user interface. It features a dynamic footer that displays live session statistics (timer, tool calls), the current working directory, Git status, and the active AI model.

## 1. Current Architectural Status

The application is built on a modular, decoupled architecture. A central `PersistentPreRun` hook in `cmd/root.go` initializes and injects all necessary services (File System, Shell, Git, Settings) into the command context, ensuring a consistent environment.

- **`pkg/services` - Session Management**:
    - `SessionService` (`session_service.go`) provides file-based persistence for chat sessions.
    - Chat history is saved as a JSON file in a dedicated directory within the user's home (`~/.go-ai-agent/sessions`).
    - The `chat` command now supports session management via flags (`--session-id`, `--new-session`, `--latest`, `--list-sessions`, `--delete-session`), allowing users to resume previous conversations.
    - Each session is automatically saved after every user message and model response, ensuring no data is lost.

- **`pkg/core` - Multi-Executor Abstraction**:
    - The `Executor` interface defines a common contract for all AI backends (`StreamContent`, `ExecuteTool`, etc.).
    - `ExecutorFactory` (`executor_factory.go`) provides an abstract factory pattern, allowing the CLI to dynamically instantiate different AI backends (`gemini`, `qwen`, `mock`) by name.
    - `gemini.go` (`GoaiagentChat`) provides a concrete implementation for the Google Gemini API.
    - `qwen.go` (`QwenChat`) provides a concrete implementation for Qwen-compatible (OpenAI-style) APIs.

- **`pkg/core/agents` - Hierarchical Sub-Agent Framework**: A sophisticated, non-interactive agent execution layer has been implemented.
    - **`AgentDefinition`**: Serves as a declarative blueprint for specialized sub-agents, defining their purpose, allowed tools, prompts, and execution parameters.
    - **`CodebaseInvestigatorAgent`**: The first concrete implementation of a sub-agent. It is designed for deep, autonomous code analysis and is defined in `codebase_investigator.go` with prompts loaded from `codebase_investigator_prompts.md`.
    - **`AgentExecutor`**: The engine that runs an `AgentDefinition`. It executes a non-interactive, multi-turn loop of model calls and tool executions until the agent's goal is met or a limit is reached.
    - **`SubagentToolWrapper`**: A clever wrapper that makes a complex, multi-turn sub-agent (like the Codebase Investigator) appear as a single, callable tool to a parent agent.

- **`pkg/tools` - Declarative Tool System**:
    - All tools follow a consistent declarative pattern, defining their `name`, `description`, and argument `JsonSchemaObject`.
    - Tools are decoupled from system-level concerns via dependency injection of services (`FileSystemService`, `GitService`, etc.).
    - `register.go` acts as the central assembly point, instantiating all tools with their dependencies and adding them to a `ToolRegistry` that is used by the active AI executor.

- **`pkg/ui` - Interactive Chat UI**:
    - `chat_ui.go` implements a sophisticated UI based on the `bubbletea` framework.
    - It functions as an event stream visualizer. It initiates a request via the `Executor`'s `StreamContent` method and then renders the stream of events (`Thinking`, `ToolCallStart`, `ToolCallEnd`, `FinalResponse`, etc.) into a human-readable, interactive format.
    - Includes advanced rendering for tool calls, providing real-time insight into the agent's actions.
    - Features a dynamic footer with live session stats, Git status, the current session ID, and more.
    - It can execute other CLI commands from within the chat (e.g., `/settings`, `/clear`) via a `commandExecutor` function.

- **Command Status**: All commands listed in the previous version of this plan are considered functionally integrated into this new architecture.

## 2. Linter-Identified Issues

All previously identified linter issues have been resolved.

## 3. Remaining JavaScript Source Code to be Migrated

The core application logic has been successfully migrated to the new Go architecture. The remaining JS code primarily serves as a reference and is located in `docs/gemini-cli-main/packages/`. No further direct migration is planned; new features will be built in Go.

- `add`, `list`, and `remove` commands for MCP servers need to be fully implemented. Currently, they only provide informative messages.
- The logic in the `docs/gemini-cli-main/packages/core/src/fallback` directory may need to be migrated.

## 4. Next Steps

1.  **Qwen Executor Enhancements:**
    *   Fully implement tool-calling capabilities for the `QwenChat` executor. The current implementation has a basic loop but needs to be as robust as the Gemini one.
    *   Implement the remaining `Executor` interface methods: `SendMessageStream`, `CompressChat`, and `GenerateContentWithTools`.

2.  **Sub-Agent Framework Expansion:**
    *   Implement more specialized, built-in sub-agents using the `AgentDefinition` pattern.
    *   Expose the `codebase_investigator` sub-agent as a runnable tool/command from the main CLI.

3.  **UI/Tool Integration:**
    *   Implement scrolling and text selection/copying in the chat viewport.
    *   Allow resizing of the tool call code view.

4.  **Error Handling & Routing:**
    *   Improve error messages for API failures (e.g., distinguish between quota errors and other API errors).
    *   Refactor `getSuggestedModel` to be more generic and implement the `ClassifierStrategy` for more intelligent model routing.

5.  **Core Commands & Services:**
    - [COMPLETED] Add validation for `executor` and `model` settings in the `settings` command.
    *   Implement full CRUD operations for MCP server management (`mcp` command).
    *   Implement secure, OS-specific storage and clearing of API keys for the `auth` command.

6.  **Completing Features:**
    *   Implement full IDE integration (`ide` command).
    - [COMPLETED] Implement saving and restoring CLI state, including tool calls and conversation/file history, for the `restore` command.
    *   Implement folder trust management for the `permissions` command.

7.  **General**:
    *   Continue to add comprehensive unit and integration tests for all new components and commands. The test suite has been successfully refactored to pass with the new architecture.
