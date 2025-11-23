# Plan: Dynamic Executor and Model Switching

## 1. Objective

Refactor the application to allow a user to dynamically switch the active AI model and executor (e.g., from `gemini` to `qwen`) during an interactive chat session via a UI command (e.g., `/settings set executor qwen`). This requires moving from a static, startup-time initialization of the executor to a more dynamic, on-demand architecture.

## 2. Architectural Challenges

- The `Executor` is currently created once at startup and stored as a global variable in `cmd/root.go`.
- The `ChatService` is initialized with this global `Executor` and holds a reference to it for its entire lifecycle.
- Changing the `executor` or `model` setting only updates the configuration file; it does not "hot-reload" any active components.

## 3. High-Level Plan

The core idea is to make the `ChatService` responsible for owning and managing its `Executor` instance. When a setting changes, instead of trying to hot-swap the executor inside the existing service, we will create a *new* `ChatService` instance with a new `Executor`, and then seamlessly transition the UI to use this new service.

---

## 4. Detailed Implementation Steps

### Phase 1: Decouple `ChatService` from Global Executor

**Objective:** Remove the reliance on the global `executor` variable in `cmd/root.go` and make the `ChatService` and `ChatUI` capable of re-initialization.

1.  **Modify `cmd/chat.go` (`runChatCmd`):**
    *   The `runChatCmd` function currently receives the global `executor`. Change this so it no longer takes the executor as an argument.
    *   Inside `runChatCmd`, add logic to:
        a.  Read the current `executor` and `model` from the `SettingsService`.
        b.  Create a new `ExecutorFactory` and a new `Executor` instance *inside* this function.
        c.  Create the `ChatService` instance using this newly created executor.
    *   This makes `runChatCmd` responsible for creating the executor it needs, each time a chat is started.

2.  **Modify `pkg/ui/chat_ui.go` (`NewChatModel`):**
    *   The `NewChatModel` function already takes `*services.ChatService` as an argument. This is good. We need to add a way to update it.
    *   Add a new method to the `ChatModel` struct: `SetChatService(newSvc *services.ChatService)`. This method will be responsible for updating the internal `chatService` field.

3.  **Modify `cmd/root.go`:**
    *   Remove the global `executor core.Executor` variable. Its creation is now handled by `runChatCmd`.
    *   The `PersistentPreRun` logic that creates the global executor can be removed.

### Phase 2: Implement the `/settings` Command Logic

**Objective:** Make the `/settings set ...` command trigger a "hot-reload" of the `ChatService`.

1.  **Modify `pkg/ui/chat_ui.go` (`Update` method):**
    *   Locate the `case commandFinishedMsg:` which handles the result of slash commands.
    *   Add a specific check for the `settings` command: `if msg.args[0] == "settings" && (msg.args[1] == "set" || msg.args[1] == "reset")`.
    *   Inside this block, instead of just showing a message, we will now trigger the re-initialization.

2.  **Create a Re-Initialization Command:**
    *   Define a new `tea.Cmd` function, e.g., `ReinitializeChatCmd()`.
    *   This command function will:
        a.  Read the latest `executor` and `model` settings from the `SettingsService`.
        b.  Create a new `Executor` instance (similar to the logic now in `runChatCmd`).
        c.  Create a new `ChatService` instance, passing it the new executor and the *history from the old `ChatService`*. A new method `GetCurrentHistory()` on the old `ChatService` might be needed.
        d.  Return a message, e.g., `chatServiceReloadedMsg`, containing the new `ChatService` instance.

3.  **Handle the Reload Message:**
    *   In the `Update` function, add a new case: `case chatServiceReloadedMsg:`.
    *   Inside this case, call `m.SetChatService(msg.newService)` on the `ChatModel`.
    *   Update the UI by adding a `BotMessage` like "Chat settings updated. Executor is now '...'.", and refresh the footer view.

### Phase 3: Refine and Verify

1.  **History Transfer:**
    *   Ensure that when the new `ChatService` is created, the history from the previous service is correctly passed to its constructor. This will provide a seamless transition for the user.

2.  **Update `settingsCmd`:**
    *   The `settingsCmd` in `cmd/settings.go` should now just call the `SettingsService.Set` and `Save` methods. The UI will handle the hot-reload.

3.  **Testing:**
    *   Run the application.
    *   Start a chat with the `mock` executor.
    *   Issue the command `/settings set executor gemini`.
    *   Verify that the UI updates and that subsequent prompts are handled by the Gemini executor.
    *   Check that the conversation history is preserved across the switch.
    *   Test switching back to the `mock` executor.
    *   Test changing the `model` setting.

## Expected Outcome

After this refactoring, a user will be able to type `/settings set executor qwen` or `/settings set model gemini-1.5-pro-latest` in the chat UI. The application will then, without restarting, create a new executor with the specified settings, transfer the current session's history to it, and continue the conversation using the new model/executor.
