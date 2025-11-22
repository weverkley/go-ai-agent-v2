# Go AI Agent v2

Go AI Agent v2 is a powerful and extensible command-line interface (CLI) built with Go, designed to bring the power of various AI models and a hierarchical agent system directly into your terminal.

## Key Features

- **Interactive Chat Mode**: A rich, interactive chat UI powered by Bubble Tea, featuring:
    - A dynamic "GO AI AGENT" banner on startup.
    - A persistent footer displaying live session statistics (timer, tool call counts), current working directory, Git branch status, and the active AI model.
    - Real-time streaming of AI responses.
    - Advanced, color-coded rendering for tool calls in a clean, boxed layout.
    - Syntax highlighting for code within tool calls.
    - Automatic model fallback suggestions on API errors.

- **Multi-Executor Support**: A flexible, factory-based architecture allows the application to be independent of any single AI provider.
    - **Interface-Driven**: Core logic is built against a common `Executor` interface.
    - **Pluggable Backends**: Concrete implementations for Google Gemini (`gemini.go`) and Qwen-compatible endpoints (`qwen.go`) can be selected at runtime.

- **Extensible Tool System**: A robust, declarative system for adding new capabilities to the agent.
    - **Declarative by Design**: Each tool defines its own name, description, and argument schema, making it easy for the AI to understand and use.
    - **Centrally Registered**: All tools are instantiated and registered in a central `ToolRegistry`, which is then provided to the active executor.

- **Hierarchical Sub-Agent Framework**: A powerful system for delegating complex tasks to specialized, autonomous agents that run non-interactively.
    - **Agent Blueprints**: The `AgentDefinition` struct allows for creating detailed blueprints for sub-agents, specifying their purpose, allowed tools, and system prompts.
    - **Autonomous Execution**: The `AgentExecutor` runs these agents in an isolated, multi-turn loop until they complete their objective.
    - **Agents as Tools**: The `CodebaseInvestigatorAgent` is the first implementation of this, a specialized agent for deep code analysis that can be called like any other tool by a parent agent.

## Getting Started

### Prerequisites

- Go (version 1.21 or higher recommended)
- API Keys for your desired AI models (e.g., `GEMINI_API_KEY`, `QWEN_API_KEY`) set as environment variables.

### Build & Run

1.  **Build the CLI:**
    ```bash
    go build .
    ```
    This will create the `go-ai-agent-v2` executable.

2.  **Initialize the Project (First-Time Use):**
    Before starting a chat in a new project, run the `init` command. This analyzes your project and creates a `GOAIAGENT.md` file that provides essential context to the AI.
    ```bash
    ./go-ai-agent-v2 init
    ```

3.  **Start the Interactive Chat:**
    Running the executable without any arguments will start the interactive chat session.
    ```bash
    ./go-ai-agent-v2
    ```

4.  **Explore other commands:**
    ```bash
    # See all commands
    ./go-ai-agent-v2 --help

    # List available AI models
    ./go-ai-agent-v2 list-models

    # Run a one-shot generation
    ./go-ai-agent-v2 generate "Tell me a joke about Go."
    ```

## Project Structure

- **`cmd/`**: Contains the entry points for all CLI commands, powered by Cobra. `root.go` is the central initializer for all services and commands.
- **`pkg/`**: Contains the core logic, organized by feature:
    - **`core/`**: Contains the primary AI logic.
        - `executor.go`: The central `Executor` interface that abstracts different AI backends.
        - `executor_factory.go`: The factory for creating specific executor instances (Gemini, Qwen, Mock).
        - `gemini.go`, `qwen.go`: The concrete implementations for each AI provider.
        - **`agents/`**: The hierarchical sub-agent framework, including the `AgentDefinition` blueprint, the `AgentExecutor` engine, and the `CodebaseInvestigatorAgent`.
    - **`ui/`**: The Bubble Tea-based interactive chat interface (`chat_ui.go`), which renders the stream of events from the core executor and includes a dynamic footer for displaying real-time session statistics.
    - **`tools/`**: Definitions for all available agent tools (e.g., `read_file`, `execute_command`). `register.go` assembles the central tool registry.
    - **`services/`**: Shared, decoupled services for file system, Git, shell execution, etc.
    - **`config/`**: Application configuration and settings management.
    - **`routing/`**: Logic for dynamically routing requests to different AI models based on context or errors.
    - **`types/`**: Centralized data structures and interfaces used across the application.

## Contributing

Contributions are welcome! Please see `PLAN.md` for the development roadmap and `docs/gemini-cli-main/GEMINI.md` for coding conventions and guidelines.