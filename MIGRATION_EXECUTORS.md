# Refactoring Executors for Consistent Confirmation Flow

## Problem Statement

The user confirmation (and general tool execution) flow is inconsistent across the application's executors (`gemini`, `qwen`, `mock`). While the `mock` executor has been refined to correctly provide `FunctionCall` parts to the `ChatService` for orchestration, the `gemini` and `qwen` executors implement their own internal tool execution logic. This bypasses the central `ChatService`'s confirmation mechanism, leading to a lack of user prompts for dangerous operations when using these executors.

The `ChatService` is designed to be the single source of truth for the tool-calling loop, including confirmation handling. Executors should solely be responsible for communicating with their respective model APIs and streaming back model responses (text, function calls, function responses).

## Objective

Refactor the `gemini.go` and `qwen.go` executors to align with the intended architectural pattern: executors provide `FunctionCall` parts to the `ChatService`, and the `ChatService` handles tool execution, confirmation, and history management.

## Detailed Plan of Execution

### Phase 1: Refactor `pkg/core/gemini.go` (`GoaiagentChat`)

1.  **Add `ToolConfirmationChan` to `GoaiagentChat` struct:**
    *   Add `ToolConfirmationChan chan types.ToolConfirmationOutcome` to the `GoaiagentChat` struct to store the channel provided by the `ChatService`.

2.  **Implement `SetToolConfirmationChannel`:**
    *   Modify the `SetToolConfirmationChannel` method to correctly store the passed channel into the `GoaiagentChat` instance's `ToolConfirmationChan` field.

3.  **Refactor `GoaiagentChat.StreamContent`:**
    *   **Remove internal tool execution loop:** Currently, `GoaiagentChat.StreamContent` (and `SendMessageStream` which it uses) is designed to handle the entire streaming process, including receiving function calls and sending them back to the model. This needs to change.
    *   **Simplified streaming:** The `StreamContent` method should *only* call the Gemini API, convert the received parts (`genai.Text`, `genai.FunctionCall`, etc.) into `types.Part`, and send them directly to `eventChan`.
    *   **No internal tool execution:** It should **NOT**:
        *   Execute any `FunctionCall`s directly.
        *   Re-call the Gemini API after receiving `FunctionCall`s (this logic belongs entirely to `ChatService`).
        *   Store/manage internal chat history for multi-turn tool execution. The `ChatService` manages the full history (`cs.history`). The executor's `StreamContent` will receive the current `cs.history` as an argument.

4.  **Review and adjust related methods:**
    *   `GenerateContentWithTools`: This method might become redundant or need simplification if `StreamContent` becomes the primary interaction point. It's currently a direct call to the Gemini API, but the `ChatService`'s new loop structure primarily relies on `StreamContent`. Ensure it remains functional if used elsewhere or remove/refactor if obsolete.

### Phase 2: Refactor `pkg/core/qwen.go` (`QwenChat`)

1.  **Add `ToolConfirmationChan` to `QwenChat` struct:**
    *   Add `ToolConfirmationChan chan types.ToolConfirmationOutcome` to the `QwenChat` struct.

2.  **Implement `SetToolConfirmationChannel`:**
    *   Modify the `SetToolConfirmationChannel` method to correctly store the passed channel.

3.  **Refactor `QwenChat.StreamContent`:**
    *   **Remove internal tool execution loop:** Similar to `gemini.go`, the `QwenChat.StreamContent` method currently executes tool calls internally and re-prompts the model with tool responses. This entire logic needs to be removed.
    *   **Simplified streaming:** The `StreamContent` method should:
        *   Construct the `openai.ChatCompletionRequest` with the provided history and tools.
        *   Call `qc.client.CreateChatCompletionStream`.
        *   Iterate over the stream, convert `openai.ChatCompletionDelta` parts into `types.Part` (especially `FunctionCall` and `Text`).
        *   Send these `types.Part` objects to the `eventChan`.
    *   **No internal tool execution:** It should **NOT**:
        *   Call `qc.ExecuteTool` internally within `StreamContent`.
        *   Re-call `qc.client.CreateChatCompletionStream` with tool responses.

4.  **Review and adjust related methods:**
    *   `GenerateContentWithTools`: Review its usage and align it with the simplified `StreamContent` behavior or refactor as needed.

### Phase 3: Verification

1.  **Run `go build`:** Ensure the entire project compiles successfully after all changes.
2.  **Run `go test ./...`:** Verify that all existing unit tests (especially those for `ChatService` and mock executors) still pass. New tests for `gemini` and `qwen` executors interacting with the `ChatService`'s confirmation flow may be necessary.
3.  **Manual testing:**
    *   Run the application with the `gemini` executor (if configured).
    *   Run the application with the `qwen` executor (if configured).
    *   Confirm that `write_file` and `smart_edit` operations now correctly trigger user confirmation prompts.
    *   Confirm that the overall conversation flow proceeds as expected after confirmation.

## Expected Outcome

After this refactoring, all executors (`gemini`, `qwen`, and `mock`) will conform to the same architectural pattern: they will stream raw model responses (including `FunctionCall`s) back to the `ChatService`. The `ChatService` will then be solely responsible for executing tools, handling confirmation prompts, and managing the conversation history in a consistent and robust manner. This will ensure that dangerous tool confirmations are correctly displayed regardless of the chosen executor.
