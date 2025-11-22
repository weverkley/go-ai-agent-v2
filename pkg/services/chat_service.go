package services

import (
	"context"
	"fmt"
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
	userConfirmationChan chan bool
	toolCallCounter      int
	toolErrorCounter     int
	sessionService       *SessionService // New field
	sessionID            string          // New field
}

// NewChatService creates a new ChatService.
func NewChatService(executor core.Executor, toolRegistry types.ToolRegistryInterface, sessionService *SessionService, sessionID string) (*ChatService, error) {
	confirmationChan := make(chan bool, 1)
	executor.SetUserConfirmationChannel(confirmationChan)

	initialHistory, err := sessionService.LoadHistory(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to load session history for ID %s: %w", sessionID, err)
	}

	cs := &ChatService{
		executor:             executor,
		toolRegistry:         toolRegistry,
		history:              initialHistory,
		userConfirmationChan: confirmationChan,
		sessionService:       sessionService,
		sessionID:            sessionID,
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
				// Pass the event through to the UI
				eventChan <- event

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
					// If the stream returns an error, stop processing
					return
				}
			}

			// After the stream is done, decide what to do next
			if len(functionCalls) > 0 {
				telemetry.LogDebugf("ChatService: Received %d tool call(s) from model.", len(functionCalls))

				// Add the model's response (containing the tool calls) to history
				cs.history = append(cs.history, &types.Content{Role: "model", Parts: modelResponseParts})
				if err := cs.sessionService.SaveHistory(cs.sessionID, cs.history); err != nil {
					eventChan <- types.ErrorEvent{Err: fmt.Errorf("failed to save history after model response with tool calls: %w", err)}
					return
				}

				var toolResponseParts []types.Part

				// Execute all tool calls
				for _, fc := range functionCalls {
					cs.toolCallCounter++
					toolCallID := fmt.Sprintf("tool-call-%d", cs.toolCallCounter)

					// Intercept user_confirm before execution
					if fc.Name == types.USER_CONFIRM_TOOL_NAME {
						message := "Confirmation required."
						if msg, ok := fc.Args["message"].(string); ok {
							message = msg
						}

						telemetry.LogDebugf("ChatService: Emitting UserConfirmationRequestEvent for tool call %s", toolCallID)
						eventChan <- types.UserConfirmationRequestEvent{ToolCallID: toolCallID, Message: message}
						telemetry.LogDebugf("Received stream event: UserConfirmationRequestEvent (ID: %s, Message: %s)", toolCallID, message)

						confirmed := <-cs.userConfirmationChan
						telemetry.LogDebugf("ChatService: Received user confirmation response: %t", confirmed)

						result := "cancel"
						if confirmed {
							result = "continue"
						}

						toolResponseParts = append(toolResponseParts, types.Part{
							FunctionResponse: &types.FunctionResponse{
								Name:     fc.Name,
								Response: map[string]any{"result": result},
							},
						})
						continue // Go to next tool call
					}

					// For all other tools, execute them
					eventChan <- types.ToolCallStartEvent{ToolCallID: toolCallID, ToolName: fc.Name, Args: fc.Args}
					telemetry.LogDebugf("Received stream event: ToolCallStartEvent (ID: %s, Name: %s, Args: %#v)", toolCallID, fc.Name, fc.Args)

					tool, err := cs.toolRegistry.GetTool(fc.Name)
					if err != nil {
						telemetry.LogErrorf("Tool %s not found: %v", fc.Name, err)
						eventChan <- types.ToolCallEndEvent{ToolCallID: toolCallID, ToolName: fc.Name, Err: err}
						telemetry.LogDebugf("Received stream event: ToolCallEndEvent (ID: %s, Name: %s, Result: %s, Err: %v)", toolCallID, fc.Name, "", err)
						toolResponseParts = append(toolResponseParts, types.Part{
							FunctionResponse: &types.FunctionResponse{
								Name: fc.Name,
								Response: map[string]any{"error": "Tool not found"},
							},
						})
						continue
					}

					result, err := tool.Execute(ctx, fc.Args)
					if err != nil {
						telemetry.LogErrorf("Error executing tool %s: %v", fc.Name, err)
						cs.toolErrorCounter++
					}

					eventChan <- types.ToolCallEndEvent{ToolCallID: toolCallID, ToolName: fc.Name, Result: result.ReturnDisplay, Err: err}
					telemetry.LogDebugf("Received stream event: ToolCallEndEvent (ID: %s, Name: %s, Result: %s, Err: %v)", toolCallID, fc.Name, result.ReturnDisplay, err)

					toolResponseParts = append(toolResponseParts, types.Part{
						FunctionResponse: &types.FunctionResponse{
							Name:     fc.Name,
							Response: map[string]any{"result": result.LLMContent, "error": err},
						},
					})
				}
				telemetry.LogDebugf("ChatService: Finished tool call loop.")
				// Add the collected tool responses to history for the next turn
				cs.history = append(cs.history, &types.Content{Role: "user", Parts: toolResponseParts})
				if err := cs.sessionService.SaveHistory(cs.sessionID, cs.history); err != nil {
					eventChan <- types.ErrorEvent{Err: fmt.Errorf("failed to save history after tool responses: %w", err)}
					return
				}

			} else { // No tool calls, this is the final answer
				telemetry.LogDebugf("ChatService: Received final text response.")
				// The final text is already composed of the text parts sent to the UI
				// We just need to add the complete response to history
				cs.history = append(cs.history, &types.Content{Role: "model", Parts: modelResponseParts})
				if err := cs.sessionService.SaveHistory(cs.sessionID, cs.history); err != nil {
					eventChan <- types.ErrorEvent{Err: fmt.Errorf("failed to save history after final model response: %w", err)}
					return
				}
				eventChan <- types.FinalResponseEvent{Content: textResponse.String()}
				telemetry.LogDebugf("Received stream event: FinalResponseEvent (Content: %s)", textResponse.String())
				return // Exit the loop
			}
		}
	}()

	return eventChan, nil
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

// GetToolRegistry returns the tool registry instance.
func (cs *ChatService) GetToolRegistry() types.ToolRegistryInterface {
	return cs.toolRegistry
}

// GetExecutor returns the executor instance.
func (cs *ChatService) GetExecutor() core.Executor {
	return cs.executor
}

// GetToolCallCount returns the total number of tool calls made in the session.
func (cs *ChatService) GetToolCallCount() int {
	return cs.toolCallCounter
}

// GetToolErrorCount returns the total number of tool calls that resulted in an error.
func (cs *ChatService) GetToolErrorCount() int {
	return cs.toolErrorCounter
}
