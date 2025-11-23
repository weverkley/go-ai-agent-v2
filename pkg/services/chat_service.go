package services

import (
	"context"
	"fmt"
	"os"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/core"
	"go-ai-agent-v2/go-cli/pkg/telemetry"
	"go-ai-agent-v2/go-cli/pkg/types"
)

// ChatService orchestrates the interactive chat session, handling the tool-calling loop.
type ChatService struct {
	executor             core.Executor
	toolRegistry         types.ToolRegistryInterface
	history              []*types.Content
	userConfirmationChan chan bool                          // Keep for the user_confirm tool
	ToolConfirmationChan chan types.ToolConfirmationOutcome // New channel for rich confirmation
	toolCallCounter      int
	toolErrorCounter     int
	sessionService       *SessionService
	sessionID            string
	settingsService      types.SettingsServiceIface // Change to interface
	proceedAlwaysTools   map[string]bool            // Store tool names for which "Proceed Always" is active
}

// NewChatService creates a new ChatService.
func NewChatService(executor core.Executor, toolRegistry types.ToolRegistryInterface, sessionService *SessionService, sessionID string, settingsService types.SettingsServiceIface) (*ChatService, error) {
	userConfirmationChan := make(chan bool, 1)
	toolConfirmationChan := make(chan types.ToolConfirmationOutcome, 1)
	executor.SetUserConfirmationChannel(userConfirmationChan) // The executor expects a chan bool for user_confirm
	executor.SetToolConfirmationChannel(toolConfirmationChan) // The executor expects a chan types.ToolConfirmationOutcome for rich tool confirmation

	initialHistory, err := sessionService.LoadHistory(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to load session history for ID %s: %w", sessionID, err)
	}

	cs := &ChatService{
		executor:             executor,
		toolRegistry:         toolRegistry,
		history:              initialHistory,
		userConfirmationChan: userConfirmationChan,
		ToolConfirmationChan: toolConfirmationChan,
		sessionService:       sessionService,
		sessionID:            sessionID,
		settingsService:      settingsService,
		proceedAlwaysTools:   make(map[string]bool), // Initialize the map
	}
	// Save the loaded (or empty) history immediately to ensure the session file exists if it's new.
	if err := cs.sessionService.SaveHistory(cs.sessionID, cs.history); err != nil {
		return nil, fmt.Errorf("failed to save initial session history: %w", err)
	}
	return cs, nil
}

// SendMessage starts the conversation loop for a user's message and returns a channel of events.
func (cs *ChatService) SendMessage(ctx context.Context, userInput string) (<-chan any, error) {
	eventChan := make(chan any)

	// The user's message is the first turn in this conversation sequence
	currentContent := &types.Content{
		Role:  "user",
		Parts: []types.Part{{Text: userInput}},
	}
	cs.history = append(cs.history, currentContent)
	if err := cs.sessionService.SaveHistory(cs.sessionID, cs.history); err != nil {
		eventChan <- types.ErrorEvent{Err: fmt.Errorf("failed to save history after user message: %w", err)}
		return nil, err
	}

	go func() {
		defer close(eventChan)

		eventChan <- types.StreamingStartedEvent{}
		telemetry.LogDebugf("Received stream event: StreamingStartedEvent")

		for {
			telemetry.LogDebugf("ChatService: Top of the loop.")
			// Check for cancellation at the start of each turn
			select {
			case <-ctx.Done():
				telemetry.LogDebugf("Context cancelled before calling executor.")
				eventChan <- types.ErrorEvent{Err: ctx.Err()}
				return
			default:
			}

			eventChan <- types.ThinkingEvent{}
			telemetry.LogDebugf("Received stream event: ThinkingEvent")

			// Call the executor to get the model's response stream
			stream, err := cs.executor.StreamContent(ctx, cs.history...)
			if err != nil {
				eventChan <- types.ErrorEvent{Err: err}
				return
			}

			var functionCalls []*types.FunctionCall
			var textResponse strings.Builder
			var modelResponseParts []types.Part

			// Process all events from the stream for this turn
			for event := range stream {
				switch e := event.(type) {
				case types.Part:
					modelResponseParts = append(modelResponseParts, e)
					if e.FunctionCall != nil {
						functionCalls = append(functionCalls, e.FunctionCall)
					}
					if e.Text != "" {
						textResponse.WriteString(e.Text)
					}
				case types.ErrorEvent:
					telemetry.LogDebugf("Received stream event: ErrorEvent (Err: %#v)", e.Err)
					return
				}
			}

			// After the stream is done, decide what to do next
			if len(functionCalls) > 0 {
				telemetry.LogDebugf("ChatService: Received %d tool call(s) from model.", len(functionCalls))

				cs.history = append(cs.history, &types.Content{Role: "model", Parts: modelResponseParts})
				if err := cs.sessionService.SaveHistory(cs.sessionID, cs.history); err != nil {
					eventChan <- types.ErrorEvent{Err: fmt.Errorf("failed to save history after model response with tool calls: %w", err)}
					return
				}

				var toolResponseParts []types.Part
				for _, fc := range functionCalls {
					cs.toolCallCounter++
					toolCallID := fmt.Sprintf("tool-call-%d", cs.toolCallCounter)

					// --- Generic confirmation for dangerous tools ---
					isDangerousTool := false
					for _, dt := range cs.settingsService.GetDangerousTools() {
						if fc.Name == dt {
							isDangerousTool = true
							break
						}
					}

					eventChan <- types.ToolCallStartEvent{ToolCallID: toolCallID, ToolName: fc.Name, Args: fc.Args}
					telemetry.LogDebugf("Received stream event: ToolCallStartEvent (ID: %s, Name: %s, Args: %#v)", toolCallID, fc.Name, fc.Args)

					var toolExecutionResult any
					var toolExecutionError error

					if isDangerousTool {
						if cs.proceedAlwaysTools[fc.Name] {
							telemetry.LogDebugf("ChatService: Proceeding automatically for tool '%s' (Proceed Always).", fc.Name)
							if fc.Name == types.USER_CONFIRM_TOOL_NAME {
								toolExecutionResult = map[string]any{"result": "continue"}
							} else {
								toolExecutionResult, toolExecutionError = executeTool(ctx, fc, cs.toolRegistry)
							}
						} else {
							confirmationEvent := types.ToolConfirmationRequestEvent{
								ToolCallID: toolCallID,
								ToolName:   fc.Name,
								ToolArgs:   fc.Args,
								Type:       "exec",
								Message:    fmt.Sprintf("Confirm execution of tool '%s'?", fc.Name),
							}
							// ... (rest of the switch for message formatting)
							switch fc.Name {
							case types.USER_CONFIRM_TOOL_NAME:
								confirmationEvent.Type = "info"
								if msg, ok := fc.Args["message"].(string); ok {
									confirmationEvent.Message = msg
								}
							case types.WRITE_FILE_TOOL_NAME:
								confirmationEvent.Type = "edit"
								confirmationEvent.Message = "Apply this change?"
								if filePath, ok := fc.Args["file_path"].(string); ok {
									confirmationEvent.FilePath = filePath
									if newContent, ok := fc.Args["content"].(string); ok {
										confirmationEvent.NewContent = newContent
										originalContentBytes, err := os.ReadFile(filePath)
										if err == nil {
											confirmationEvent.OriginalContent = string(originalContentBytes)
											confirmationEvent.FileDiff = generateDiff(string(originalContentBytes), newContent)
										}
									}
								}
							}

							eventChan <- confirmationEvent
							outcome := <-cs.ToolConfirmationChan

							switch outcome {
							case types.ToolConfirmationOutcomeProceedOnce, types.ToolConfirmationOutcomeProceedAlways:
								if outcome == types.ToolConfirmationOutcomeProceedAlways {
									cs.proceedAlwaysTools[fc.Name] = true
								}
								if fc.Name == types.USER_CONFIRM_TOOL_NAME {
									toolExecutionResult = map[string]any{"result": "continue"}
								} else {
									toolExecutionResult, toolExecutionError = executeTool(ctx, fc, cs.toolRegistry)
								}
							case types.ToolConfirmationOutcomeCancel:
								if fc.Name == types.USER_CONFIRM_TOOL_NAME {
									toolExecutionResult = map[string]any{"result": "cancel"}
								} else {
									toolExecutionResult = "Tool execution cancelled by user."
									toolExecutionError = fmt.Errorf("tool execution cancelled by user")
								}
								cs.toolErrorCounter++
							default:
								toolExecutionResult = "Unknown confirmation outcome."
								toolExecutionError = fmt.Errorf("unknown confirmation outcome")
								cs.toolErrorCounter++
							}
						}
					} else {
						toolExecutionResult, toolExecutionError = executeTool(ctx, fc, cs.toolRegistry)
					}

					if toolExecutionError == nil && fc.Name == types.WRITE_TODOS_TOOL_NAME {
						if todosData, ok := fc.Args["todos"].([]interface{}); ok {
							total := len(todosData)
							completed := 0
							for _, item := range todosData {
								if todoMap, ok := item.(map[string]interface{}); ok {
									if status, ok := todoMap["status"].(string); ok && status == "completed" {
										completed++
									}
								}
							}
							eventChan <- types.TodosSummaryUpdateEvent{
								Summary: fmt.Sprintf("Todos %d/%d", completed, total),
							}
						}
					}

					eventChan <- types.ToolCallEndEvent{
						ToolCallID: toolCallID,
						ToolName:   fc.Name,
						Result:     fmt.Sprintf("%v", toolExecutionResult),
						Err:        toolExecutionError,
					}

					toolResponseParts = append(toolResponseParts, types.Part{
						FunctionResponse: &types.FunctionResponse{
							Name:     fc.Name,
							Response: map[string]any{"result": toolExecutionResult, "error": toolExecutionError},
						},
					})
				}

				cs.history = append(cs.history, &types.Content{Role: "user", Parts: toolResponseParts})
				if err := cs.sessionService.SaveHistory(cs.sessionID, cs.history); err != nil {
					eventChan <- types.ErrorEvent{Err: fmt.Errorf("failed to save history after tool responses: %w", err)}
					return
				}

			} else { // No tool calls, this is the final answer
				cs.history = append(cs.history, &types.Content{Role: "model", Parts: modelResponseParts})
				if err := cs.sessionService.SaveHistory(cs.sessionID, cs.history); err != nil {
					eventChan <- types.ErrorEvent{Err: fmt.Errorf("failed to save history after final model response: %w", err)}
					return
				}
				eventChan <- types.FinalResponseEvent{Content: textResponse.String()}
				return
			}
		}
	}()

	return eventChan, nil
}

// Helper to execute a tool and return its result.
func executeTool(ctx context.Context, fc *types.FunctionCall, toolRegistry types.ToolRegistryInterface) (any, error) {
	tool, err := toolRegistry.GetTool(fc.Name)
	if err != nil {
		return "Tool not found", fmt.Errorf("tool %s not found: %w", fc.Name, err)
	}

	result, err := tool.Execute(ctx, fc.Args)
	if err != nil {
		return fmt.Sprintf("Error executing tool: %v", err), err
	}
	return result.LLMContent, nil
}

// generateDiff creates a simple diff string between two contents.
// This is a basic line-by-line diff for display purposes.
func generateDiff(oldContent, newContent string) string {
	oldLines := strings.Split(oldContent, "\n")
	newLines := strings.Split(newContent, "\n")

	// Max length to prevent excessive diffs
	maxLength := 10
	if len(oldLines) > maxLength || len(newLines) > maxLength {
		return "Diff too large to display. Please confirm manually."
	}

	diff := ""
	diffAdded := func(s string) string { return fmt.Sprintf("+ %s", s) }
	diffRemoved := func(s string) string { return fmt.Sprintf("- %s", s) }
	diffUnchanged := func(s string) string { return fmt.Sprintf("  %s", s) }

	// A very basic diff algorithm for demonstration
	i, j := 0, 0
	for i < len(oldLines) && j < len(newLines) {
		if oldLines[i] == newLines[j] {
			diff += diffUnchanged(oldLines[i]) + "\n"
			i++
			j++
		} else {
			// Find where newLines[j] matches in oldLines (if at all)
			found := false
			for k := i; k < len(oldLines); k++ {
				if oldLines[k] == newLines[j] {
					// new line was inserted before oldLines[k]
					for l := i; l < k; l++ {
						diff += diffRemoved(oldLines[l]) + "\n"
					}
					i = k
					found = true
					break
				}
			}
			if !found {
				diff += diffAdded(newLines[j]) + "\n"
				j++
			}
		}
	}
	for i < len(oldLines) {
		diff += diffRemoved(oldLines[i]) + "\n"
		i++
	}
	for j < len(newLines) {
		diff += diffAdded(newLines[j]) + "\n"
		j++
	}

	return diff
}

// GetHistory returns the current chat history.
func (cs *ChatService) GetHistory() []*types.Content {
	return cs.history
}

// ClearHistory resets the chat history.
func (cs *ChatService) ClearHistory() {
	cs.history = []*types.Content{}
	// Persist the empty history to clear the session file
	if err := cs.sessionService.SaveHistory(cs.sessionID, cs.history); err != nil {
		// Log the error but don't fail, as clearing history is a best-effort operation
		telemetry.LogErrorf("Failed to clear persisted history for session %s: %v", cs.sessionID, err)
	}
}

// GetUserConfirmationChannel returns the channel used for user confirmation responses.
func (cs *ChatService) GetUserConfirmationChannel() chan bool {
	return cs.userConfirmationChan
}

// GetToolConfirmationChannel returns the channel used for rich tool confirmation responses.
func (cs *ChatService) GetToolConfirmationChannel() chan types.ToolConfirmationOutcome {
	return cs.ToolConfirmationChan
}

// GetToolRegistry returns the tool registry instance.
func (cs *ChatService) GetToolRegistry() types.ToolRegistryInterface {
	return cs.toolRegistry
}

// GetExecutor returns the executor instance.
func (cs *ChatService) GetExecutor() core.Executor {
	return cs.executor
}

// GetSettingsService returns the settings service instance.
func (cs *ChatService) GetSettingsService() types.SettingsServiceIface {
	return cs.settingsService
}

// GetToolCallCount returns the total number of tool calls made in the session.
func (cs *ChatService) GetToolCallCount() int {
	return cs.toolCallCounter
}

// GetToolErrorCount returns the total number of tool calls that resulted in an error.
func (cs *ChatService) GetToolErrorCount() int {
	return cs.toolErrorCounter
}
