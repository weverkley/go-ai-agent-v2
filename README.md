# Go AI Agent v2

Go AI Agent v2 is a powerful and extensible command-line interface (CLI) built with Go, designed to bring the power of various AI models directly into your terminal.

## Key Features

- **Interactive Chat Mode**: A rich, interactive chat UI powered by Bubble Tea, featuring:
    - Real-time streaming of AI responses.
    - Advanced, color-coded rendering for tool calls in a clean, boxed layout.
    - Syntax highlighting for code within tool calls like `write_file`.
    - Automatic model fallback suggestions on API errors.
- **Multi-Executor Support**: Easily switch between different AI backends, with initial support for Google's Gemini and Qwen models.
- **Extensible Tool System**: A robust system for adding new tools and capabilities to the agent.
- **AI-Powered Commands**: A suite of commands for common development tasks like code analysis, documentation search, and PR reviews.

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

2.  **Start the Interactive Chat:**
    ```bash
    ./go-ai-agent-v2 chat
    ```

3.  **Explore other commands:**
    ```bash
    # See all commands
    ./go-ai-agent-v2 --help

    # List available AI models
    ./go-ai-agent-v2 list-models

    # Run a one-shot generation
    ./go-ai-agent-v2 generate "Tell me a joke about Go."
    ```

## Project Structure

- **`cmd/`**: Contains the entry points for all CLI commands, powered by Cobra.
- **`pkg/`**: Contains the core logic, organized by feature:
    - **`core/`**: AI model executors (Gemini, Qwen) and core agent logic.
    - **`ui/`**: The Bubble Tea-based interactive chat interface.
    - **`tools/`**: Definitions for all available agent tools.
    - **`services/`**: Shared services for file system, Git, shell, etc.
    - **`config/`**: Application configuration and settings management.
    - **`routing/`**: Logic for dynamically routing requests to different AI models.

## Contributing

Contributions are welcome! Please see `PLAN.md` for the development roadmap and `docs/go-ai-agent-main/GEMINI.md` for coding conventions and guidelines.
