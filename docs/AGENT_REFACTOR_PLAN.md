# Agent Mode and WebSocket Refactoring Plan

## 1. Objective

This document outlines a plan to refactor the Go AI Agent to support two distinct operational modes:
1.  **CLI Mode**: The existing interactive terminal for local filesystem operations.
2.  **Agent Mode**: A non-interactive, headless service ideal for Docker deployments. This mode will not launch the UI but will instead start a server to handle tasks and broadcast events over a WebSocket connection.

This refactoring will enable the application to function both as a user-facing CLI and as a decoupled backend service.

## 2. Proposed Architecture

The core of this refactor is a new **run mode** concept controlled by a configuration setting.

-   **`runMode: "cli"` (Default)**: The application behaves as it does now, starting the interactive Bubble Tea UI.
-   **`runMode: "agent"`**: The application starts a background server and does not render a UI.

### Agent Mode Data Flow

1.  **Server Initialization**: In `agent` mode, the application will start an HTTP server. This server will host a WebSocket endpoint (e.g., `/ws`) and potentially RESTful endpoints for webhooks (e.g., `/api/v1/tasks`).
2.  **Client Connection**: A client connects to the `/ws` endpoint to initiate a task.
3.  **Task Orchestration**: Upon receiving a task request, the server instantiates the core `ChatService`, which remains unchanged.
4.  **Event Broadcasting**: A new **Broadcaster** component will subscribe to the event channel (`chan events.Event`) returned by the `ChatService`.
5.  **Event Forwarding**: The Broadcaster will serialize each event to JSON and send it over the WebSocket to the connected client(s), allowing for real-time streaming of the agent's activities.

This design reuses the existing `ChatService` and its event-driven nature, ensuring that the core logic remains consistent across both modes.

## 3. Detailed Implementation Steps

### Step 1: Add `RunMode` Configuration

1.  **Update Settings Struct**: In `pkg/config/config.go`, add a `RunMode` field to the `Settings` struct.
    ```go
    // In Settings struct
    RunMode string `json:"runMode"`
    ```
2.  **Update Settings Service**: In `pkg/services/settings_service.go`, ensure the new setting is loaded and has a default value.
    -   Add `RunMode: "cli"` to the default settings map.
    -   The existing `Get` method should already handle loading it.
3.  **Update `example.settings.json`**: Add `"runMode": "cli"` to the example configuration file.

### Step 2: Implement the Main Execution Switch

1.  **Modify `cmd/root.go`**: This is the central point for the change.
    -   In the `RootCmd.Run` function, access the new setting: `runMode := settingsService.Get("runMode")`.
    -   Implement a `switch` statement based on `runMode`.
    ```go
    // In cmd/root.go, inside RootCmd.Run
    switch settingsService.Get("runMode") {
    case "agent":
        // Call a new function to start the agent server
        runAgentCmd(cmd, args)
    case "cli":
        fallthrough
    default:
        // Call the existing function to start the UI
        runChatCmd(cmd, args)
    }
    ```

### Step 3: Abstract Telemetry for `stdout` Logging

1.  **Modify `pkg/telemetry/telemetry.go`**:
    -   Create a new logger struct `stdoutTelemetryLogger` that implements the `TelemetryLogger` interface. Its `Log` method will write formatted logs to `os.Stdout`.
    -   Update the `NewTelemetryLogger` factory to check the `runMode` or a new dedicated `telemetry.backend` setting. If in `agent` mode, return an instance of `stdoutTelemetryLogger`.
2.  **Update Configuration**: For better control, add a `backend` field to the `Telemetry` settings in `pkg/config/config.go` (`"file"` or `"stdout"`).

### Step 4: Create the Agent Server (`runAgentCmd`)

1.  **Create `cmd/agent.go`**: This new file will house the `runAgentCmd` function.
2.  **Implement `runAgentCmd`**:
    -   This function will be responsible for initializing and starting the web server.
    -   It will reuse the service initializations from `cmd/root.go` (`chatService`, `sessionService`, etc.).
3.  **Create `pkg/server/` package**: This package will contain the web server logic.
    -   `server.go`: Defines the server, routes, and handlers. Use a standard HTTP router (like `gorilla/mux`) and WebSocket library (`gorilla/websocket`).
    -   `handlers.go`: Contains the HTTP/WebSocket handler logic. The WebSocket handler will:
        - Upgrade the HTTP connection.
        - Instantiate a `Broadcaster`.
        - Read incoming task requests from the WebSocket.
        - Start the `ChatService.SendMessage` with the client's prompt.
        - Pass the returned event channel to the Broadcaster.
    -   `broadcaster.go`: Manages WebSocket connections and broadcasts events. It listens on the event channel, serializes events to JSON, and writes them to all registered client connections.

### Step 5: Define Event Schema and Task Handling

1.  **Update `pkg/ui/chat_events.go`**: Add `json:"..."` tags to the event structs so they can be serialized correctly.
2.  **Define Task Request**: Define a JSON structure for incoming tasks sent over WebSocket or via a webhook.
    ```json
    {
      "sessionId": "optional-session-id",
      "prompt": "Write a hello world app in Go",
      "settings_override": {
        "model": "gemini-1.5-pro-latest"
      }
    }
    ```
3.  **Implement Webhook Endpoint (Optional but Recommended)**:
    -   Add a `POST /api/v1/tasks` endpoint in `pkg/server/handlers.go`.
    -   This handler will parse the task request, start a `ChatService` session in a goroutine, and immediately return a `202 Accepted` response with a task ID. The events for this task would still be available via the WebSocket if the client connects.

### Step 6: Update Docker Configuration

1.  **Modify `Dockerfile`**:
    -   `EXPOSE 8080` (or the chosen port for the agent server).
2.  **Modify `docker-compose.yml`**:
    -   Map the host port to the container port (e.g., `ports: - "8080:8080"`).
    -   Set the default run mode via an environment variable:
        ```yaml
        environment:
          - GOAIAGENT_RUNMODE=agent
          # Recommended for agent mode
          - GOAIAGENT_APPROVALMODE=NEVER
          - GOAIAGENT_TELEMETRY_BACKEND=stdout
        ```

## 4. Summary of File Changes

-   **To Be Created**:
    -   `docs/AGENT_REFACTOR_PLAN.md` (This file)
    -   `cmd/agent.go` (Entry point for agent mode)
    -   `pkg/server/server.go` (Server setup)
    -   `pkg/server/handlers.go` (WebSocket and HTTP handlers)
    -   `pkg/server/broadcaster.go` (Event broadcasting logic)

-   **To Be Modified**:
    -   `cmd/root.go`: To add the CLI vs. Agent mode switch.
    -   `pkg/config/config.go`: To add the `RunMode` setting.
    -   `pkg/services/settings_service.go`: To add the default `RunMode` value.
    -   `pkg/telemetry/telemetry.go`: To add the `stdout` logger.
    -   `pkg/ui/chat_events.go`: To add JSON tags for serialization.
    -   `example.settings.json`: To include the new `runMode` setting.
    -   `Dockerfile`: To expose the new server port.
    -   `docker-compose.yml`: To configure the port mapping and set the default mode to `agent`.

## 5. Architectural Improvements & Considerations

-   **Authentication**: The new server endpoints (`/ws`, `/api/v1/tasks`) should be protected. Consider implementing a token-based authentication middleware.
-   **Structured Logging**: In `agent` mode, the `stdoutTelemetryLogger` should output logs in a structured format (e.g., JSON). This makes them easily parsable by log aggregation platforms like Splunk, Datadog, or the ELK stack.
-   **CORS**: If the WebSocket will be consumed by a web-based UI, Cross-Origin Resource Sharing (CORS) policies will need to be configured on the server.
-   **Graceful Shutdown**: The agent server should handle OS signals (e.g., `SIGINT`, `SIGTERM`) to shut down gracefully, closing active connections and finishing in-progress work where possible.
-   **Agent Configuration**: In `agent` mode, it is highly recommended to set `approvalMode` to `NEVER` to prevent the agent from stalling while waiting for user input that will never come. This should be the default in the Docker environment.
