# Go AI Agent v2

Go AI Agent v2 is a powerful and extensible command-line interface (CLI) built with Go. It provides a rich, interactive terminal experience for chatting with different AI models, leveraging a sophisticated tool-calling system and a hierarchical agent framework to perform complex software engineering tasks.

## Core Features

-   **Multi-Executor Support**: Seamlessly switch between different AI providers (e.g., Google Gemini, Qwen) without changing the core logic, thanks to a unified `Executor` interface.
-   **Advanced Tooling System**: A robust, declarative system allows the agent to use a wide array of tools, from file system operations (`read_file`, `write_file`) to version control (`git_commit`) and code analysis (`find_unused_code`).
-   **Intelligent Model Routing**: Automatically switches from a primary model (like Gemini) to a fallback model (like Qwen) in case of API quota errors, ensuring resilience and uninterrupted workflow.
-   **Safe and Interactive Tool Execution**: Implements a user confirmation flow for "dangerous" tools that modify the file system or execute commands, giving you full control over the agent's actions.
-   **Hierarchical Sub-Agents**: A framework for delegating complex tasks to specialized, autonomous agents (like the `CodebaseInvestigator`) that can perform multi-step analysis or refactoring non-interactively.
-   **Rich Interactive UI**: A terminal UI powered by Bubble Tea that provides real-time streaming, session statistics, Git status, and a clear view of the agent's actions.
-   **Persistent Chat Sessions**: Your conversations are automatically saved, allowing you to list and resume previous sessions at any time.

---

## System Architecture

The application is designed around a decoupled, service-oriented architecture that promotes testability and extensibility.

### 1. High-Level Data Flow

The flow for an interactive chat session follows these steps:

1.  **UI (`chat_ui.go`)**: Captures user input and sends it to the `ChatService`. It subscribes to a channel of events from the service to render real-time updates (text, tool calls, errors).
2.  **Orchestrator (`chat_service.go`)**: The `ChatService` acts as the central brain. It manages the conversation history and orchestrates the multi-turn logic required for tool calls.
3.  **Executor Interface (`executor.go`)**: The `ChatService` communicates with the active AI model via a generic `Executor` interface, keeping it agnostic of the specific AI provider.
4.  **Concrete Executors (`gemini.go`, `qwen.go`)**: These are the specific implementations that handle the request/response logic for each AI provider (Gemini, Qwen). They are responsible for converting API-specific data types into the application's common internal types.
5.  **Tool Execution**: When an executor returns a `FunctionCall`, the `ChatService` intercepts it, executes the corresponding tool from the `ToolRegistry`, and sends the result back to the executor to get a final answer.

### 2. Multi-Executor and Model Support

The ability to use different AI models is achieved through the **Executor Interface**.

-   The `core.Executor` interface defines a standard set of capabilities, most importantly `StreamContent`, which all model implementations must provide.
-   A `core.ExecutorFactory` is used to instantiate the correct executor (`GeminiChat` or `QwenChat`) at runtime based on the user's settings.
-   This design means adding a new AI provider only requires creating a new struct that fulfills the `Executor` interface, without changing the core application logic.

### 3. Tool Handling and User Confirmation

The tool system is designed for both power and safety.

-   **Declaration & Registration**: Tools are defined declaratively with a name, description, and a JSON schema for their parameters. They are registered in a central `ToolRegistry`.
-   **Standard Execution**: For safe, read-only tools, the `ChatService` receives a `FunctionCall`, finds the tool in the registry, calls its `Execute` method, and sends the result back to the model.
-   **Confirmation Flow (Dangerous Tools)**:
    1.  Tools that can modify the system (e.g., `write_file`, `execute_command`) are marked as "dangerous" in the settings.
    2.  When the `ChatService` receives a call for a dangerous tool, it **intercepts** the call.
    3.  Instead of executing the tool, it sends a `ToolConfirmationRequestEvent` to the UI, which may include details like a code diff.
    4.  The UI displays a `(y/n/a)` prompt to the user.
    5.  The user's decision is sent back to the `ChatService` via a channel, which then either executes the tool, cancels it, or remembers the choice for future calls.

### 4. Model Fallback and Routing Strategy

To handle API limits gracefully, the application uses a routing service.

-   The `routing.ModelRouterService` contains a set of strategies.
-   The primary strategy is to handle `googleapi: Error 429` (Quota Exceeded) errors from the Gemini executor.
-   When this specific error is detected in the `ChatService`, it consults the router.
-   The router suggests a fallback model (e.g., from the `gemini` family).
-   The `ChatService` then automatically creates a new `GeminiChat` executor, swaps it for the current one, and re-tries the user's request, providing a seamless fallback experience.

---

## Getting Started

### Prerequisites

-   Go (version 1.21 or higher recommended)
-   API Keys for your desired AI models (e.g., `GEMINI_API_KEY`, `QWEN_API_KEY`) set as environment variables.

### Build & Run

1.  **Build the CLI:**
    ```bash
    go build -o main-agent .
    ```
    This creates the `main-agent` executable.

2.  **Run the Application:**
    Use the newly built executable to start the interactive chat.
    ```bash
    ./main-agent
    ```

3.  **Explore other commands:**
    ```bash
    # See all available commands
    ./main-agent --help
    ```

## Project Structure

-   **`cmd/`**: Contains the entry points for all CLI commands, powered by Cobra.
-   **`pkg/`**: Contains the core application logic.
    -   **`core/`**: The primary AI logic, including the `Executor` interface, concrete `gemini` and `qwen` implementations, and the `agents/` sub-agent framework.
    -   **`services/`**: Decoupled services. `chat_service.go` is the central orchestrator that manages history, tool calls, and the model-switching logic.
    -   **`tools/`**: Definitions for all agent tools (`read_file`, `execute_command`, etc.).
    -   **`ui/`**: The Bubble Tea interactive chat interface.
    -   **`routing/`**: Logic for dynamically routing requests to different models.
    -   **`types/`**: Centralized data structures and interfaces.

## TODO
- Validate fallback strategy for models in the current executor (e.g. `gemini` executor) will fallback in its models (e.g. `gemini-prod` to `gemini-flash-latest`).
- Validate switching between executors.