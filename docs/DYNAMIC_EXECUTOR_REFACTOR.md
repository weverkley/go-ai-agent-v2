# Plan: Dynamic Executor and Model Switching

## 1. Objective

Refactor the application to allow a user to dynamically switch the active AI model and executor (e.g., from `gemini` to `qwen`) during an interactive chat session via a UI command (e.g., `/settings set executor qwen`). This requires moving from a static, startup-time initialization of the executor to a more dynamic, on-demand architecture, **while ensuring the current chat history and session context are seamlessly preserved.**

## 2. Architectural Analysis

- **Current State:** The `Executor` is created once at startup and stored as a global variable in `cmd/root.go`. The `ChatService` is initialized with this global `Executor` and holds a reference to it for its entire lifecycle. Changing a setting only updates the configuration file; it does not "hot-reload" any active components.
- **Existing Capability:** A sophisticated, strategy-based `ModelRouterService` already exists in `pkg/routing`. It can make dynamic decisions based on context (including error fallbacks, user overrides, and content analysis). However, it is currently only invoked once at startup when creating the `goaiagent` executor.

## 3. High-Level Plan

The core idea is to make the `ChatService` responsible for owning and managing its `Executor` instance. When a setting changes, instead of trying to hot-swap the executor, we will:
1.  Bundle the active session's state (history, session ID, user preferences) into a "state object".
2.  Create a *new* `ChatService` instance, initializing it from the state object and a new, dynamically-created `Executor`.
3.  Seamlessly transition the UI to use this new service, preserving the full conversation context.

---

## 4. Detailed Implementation Steps

### Phase 1: Decouple `ChatService` from Global Executor

**Objective:** Remove the reliance on the global `executor` variable in `cmd/root.go` and empower the chat session to create executors on-demand.

1.  **Modify `cmd/chat.go` (`runChatCmd`):**
    *   The `runChatCmd` function currently receives the global `executor`. Change this so it no longer takes the executor as an argument.
    *   Inside `runChatCmd`, add logic to:
        a.  Read the current `executor` setting from the `SettingsService`.
        b.  Create a new `ExecutorFactory` based on the setting.
        c.  Call the factory's `NewExecutor` method. The factory (specifically `GoaiagentExecutorFactory`) will internally use the `ModelRouterService` to select the best model.
        d.  Create the `ChatService` instance using this newly created executor.
    *   This makes `runChatCmd` responsible for creating the executor it needs each time a chat is started.

2.  **Modify `pkg/ui/chat_ui.go` (`ChatModel`):**
    *   Add a new method to the `ChatModel` struct: `SetChatService(newSvc *services.ChatService, newExecutorType string)`. This method will be responsible for updating the internal `chatService` and `executorType` fields to reflect the change in the UI footer.

3.  **Modify `cmd/root.go`:**
    *   Remove the global `executor core.Executor` variable. Its creation is now handled within the chat command's lifecycle.
    *   Remove the `PersistentPreRun` logic that creates the global executor instance. The `PersistentPreRun` should only initialize services like `SettingsService`, `ToolRegistry`, etc.

### Phase 2: Implement State Transfer and Hot-Reload

**Objective:** Create a robust mechanism for transferring session state and make the `/settings` command trigger the hot-reload.

1.  **Define a `ChatState` Object:**
    *   In `pkg/services/chat_service.go` (or a `types` file), define a new struct, e.g., `ChatState`, to act as a data transfer object.
    *   `ChatState` should contain:
        ```go
        type ChatState struct {
            History            []*types.Content
            SessionID          string
            ProceedAlwaysTools map[string]bool
            ToolCallCounter    int
            ToolErrorCounter   int
        }
        ```

2.  **Implement State Capture:**
    *   In `pkg/services/chat_service.go`, create a new method: `func (cs *ChatService) GetState() *ChatState`.
    *   This method will populate and return a `ChatState` struct with the `ChatService`'s current data.

3.  **Update `NewChatService` Constructor:**
    *   Modify the `NewChatService` constructor. Instead of taking just a `sessionID`, it should accept an optional `initialState *ChatState`.
    *   If `initialState` is provided, the new `ChatService` initializes its fields from the state object.
    *   If `initialState` is `nil` (for a new or resumed session), it proceeds with the existing logic of loading history from disk via the `SessionService`.

4.  **Create a Re-Initialization Command (in `pkg/ui/chat_ui.go`):**
    *   Define a new `tea.Cmd` function, e.g., `ReinitializeChatCmd(oldChatService *services.ChatService)`.
    *   This command function will:
        a.  Call `oldChatService.GetState()` to capture the full context of the current session.
        b.  Read the latest `executor` and `model` settings from the `SettingsService`.
        c.  Create a new `Executor` instance via the `ExecutorFactory` and router.
        d.  Create a new `ChatService` instance, passing it the new executor and the `ChatState` object from step (a).
        e.  Return a message, e.g., `chatServiceReloadedMsg`, containing both the new `ChatService` and the new `executorType` string.

5.  **Trigger the Reload from the UI:**
    *   In the `Update` method's `case commandFinishedMsg:`, find the `settings` command block.
    *   When the command is detected, return the `ReinitializeChatCmd(m.chatService)` to trigger the hot-reload.

6.  **Handle the Reload Message:**
    *   In the `Update` function, add a new `case chatServiceReloadedMsg:`.
    *   Inside this case, call `m.SetChatService(msg.newService, msg.newExecutorType)`.
    *   Add a `BotMessage` to the UI to inform the user of the successful switch.

### Phase 3: Verification

1.  **Testing:**
    *   Run the application and start a chat.
    *   Make several tool calls, including one where you select "allow always" (`a`).
    *   Issue the command `/settings set executor qwen`.
    *   **Verify History:** The entire conversation history must be present after the switch.
    *   **Verify Context:** Make another call with the tool you previously allowed always. It should execute without a prompt, confirming the `ProceedAlwaysTools` map was transferred.
    *   **Verify UI:** The footer should now show "qwen" as the executor.
    *   **Verify Functionality:** Subsequent prompts should be handled by the new Qwen executor.

## Expected Outcome

After this refactoring, a user will be able to type `/settings set executor qwen` in the chat UI. The application will then, without restarting, create a new executor, seamlessly transfer the **entire session state (history, preferences, etc.)**, and continue the conversation using the new executor. The user experience will be uninterrupted.
