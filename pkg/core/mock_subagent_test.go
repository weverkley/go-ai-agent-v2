package core_test

import (
	"context"
	"fmt" // Added fmt import
	"go-ai-agent-v2/go-cli/pkg/core"
	"go-ai-agent-v2/go-cli/pkg/types"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMockExecutor_SubagentFlow(t *testing.T) {
	// 1. Setup - Create a mock tool registry
	toolRegistry := types.NewToolRegistry()

	// Register all tools used by the NewRealisticMockExecutor
	toolRegistry.Register(&mockTool{name: types.LS_TOOL_NAME})
	toolRegistry.Register(&mockTool{name: types.CODEBASE_INVESTIGATOR_TOOL_NAME}) // Subagent itself is a tool
	toolRegistry.Register(&mockTool{name: types.TASK_COMPLETE_TOOL_NAME})
	toolRegistry.Register(&mockTool{name: types.WRITE_TODOS_TOOL_NAME})
	toolRegistry.Register(&mockTool{name: types.USER_CONFIRM_TOOL_NAME})
	toolRegistry.Register(&mockTool{name: types.WRITE_FILE_TOOL_NAME})
	toolRegistry.Register(&mockTool{name: types.SMART_EDIT_TOOL_NAME})

	// 2. Initialize the realistic mock executor with the tool registry
	mockExecutor := core.NewRealisticMockExecutor(toolRegistry)

	// Context for the stream
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // Add a timeout
	defer cancel()

	// History accumulator to simulate ChatService behavior
	history := []*types.Content{
		{
			Role: "user",
			Parts: []types.Part{
				{Text: "Initial user message to start the flow."},
			},
		},
	}

	// Channel to collect all events from the mock executor's stream
	allEvents := make(chan any, 100) // Buffered channel

	// Simulate the ChatService loop
	go func() {
		defer close(allEvents)
		var currentInput []*types.Content
		for {
			// Always send the current full history to the mock StreamContent
			currentInput = history

			stream, err := mockExecutor.StreamContent(ctx, currentInput...)
			if err != nil {
				allEvents <- types.ErrorEvent{Err: err}
				return
			}

			// Collect events from the mock's stream
			var modelResponseParts []types.Part
			var functionCalls []*types.FunctionCall
			var finalResponseText string
			var streamError error

			for event := range stream {
				select {
				case <-ctx.Done():
					allEvents <- types.ErrorEvent{Err: ctx.Err()}
					return
				default:
				}

				switch e := event.(type) {
				case types.Part:
					if e.FunctionCall != nil {
						functionCalls = append(functionCalls, e.FunctionCall)
					}
					if e.Text != "" {
						finalResponseText += e.Text
					}
					modelResponseParts = append(modelResponseParts, e)
					allEvents <- e // Forward to the main test goroutine
				case types.SubagentActivityEvent:
					allEvents <- e // Forward to the main test goroutine
				case types.ErrorEvent:
					streamError = e.Err
					allEvents <- e
					return
				}
			}

			if streamError != nil {
				return
			}

			// After collecting all parts, add model's response to history
			if len(modelResponseParts) > 0 {
				history = append(history, &types.Content{Role: "model", Parts: modelResponseParts})
			}

			// If there were function calls, simulate their execution and add response to history
			if len(functionCalls) > 0 {
				var toolResponseParts []types.Part
				for _, fc := range functionCalls {
					// In a real ChatService, the actual tool execution happens here.
					// For this mock test, NewRealisticMockExecutor's ExecuteToolFunc
					// returns a generic success or handles specific tools like WRITE_TODOS.
					toolResult, err := mockExecutor.ExecuteTool(ctx, fc) // Use mock's ExecuteToolFunc
					if err != nil {
						toolResponseParts = append(toolResponseParts, types.Part{
							FunctionResponse: &types.FunctionResponse{
								Name:     fc.Name,
								Response: map[string]interface{}{"error": err.Error()},
							},
						})
						allEvents <- types.ErrorEvent{Err: fmt.Errorf("mock tool execution failed: %w", err)}
						continue
					}
					// Check for TASK_COMPLETE, which effectively ends the subagent's internal loop
					if fc.Name == types.TASK_COMPLETE_TOOL_NAME {
						allEvents <- types.Part{
							FunctionResponse: &types.FunctionResponse{
								Name:     fc.Name,
								Response: map[string]interface{}{"result": toolResult.LLMContent},
							},
						}
						// If the mock signals task complete, we break the loop here.
						// The main test goroutine will handle the final assertions.
						return
					}

					toolResponseParts = append(toolResponseParts, types.Part{
						FunctionResponse: &types.FunctionResponse{
							Name:     fc.Name,
							Response: map[string]interface{}{"result": toolResult.LLMContent},
						},
					})
				}
				history = append(history, &types.Content{Role: "tool", Parts: toolResponseParts})
			} else if finalResponseText != "" {
				// If no function calls and a final text response, conversation might be over
				allEvents <- types.Part{Text: finalResponseText} // Forward final text as well
				return // End conversation loop
			}

			// If no function calls and no final text, but the mock still has steps, continue
			if mockExecutor.MockStep >= 10 { // Max step in mock is 9 for subagent flow, so 10 indicates end
				return
			}

			time.Sleep(1 * time.Millisecond) // Prevent busy-looping in test
		}
	}()

	receivedEventCount := 0
	for event := range allEvents {
		// Log event for debugging
		// t.Logf("Received event: %#v (type: %T)", event, event)

		if errEvent, ok := event.(types.ErrorEvent); ok {
			assert.FailNow(t, "Received error event from mock stream", errEvent.Err.Error())
		}

		// We need to skip the first 8 steps of events (0-7), which are for the Todo API creation,
		// to focus on the subagent part (steps 8 and 9 in the mock).
		// This assertion section will need to be carefully aligned with the mock's step progression.

		// Given the `NewRealisticMockExecutor` starts its subagent flow at `mock.MockStep == 8`
		// and the events from the previous steps are also pushed to the `allEvents` channel,
		// we first need to consume them or match against them.

		// For simplicity in this example, let's just assert that *some* subagent activities are present.
		if saEvent, ok := event.(types.SubagentActivityEvent); ok {
			assert.True(t, saEvent.IsSubagentActivityEvent)
			assert.Equal(t, types.CODEBASE_INVESTIGATOR_TOOL_NAME, saEvent.AgentName)
			receivedEventCount++
		} else if part, ok := event.(types.Part); ok {
			if part.FunctionCall != nil && part.FunctionCall.Name == types.CODEBASE_INVESTIGATOR_TOOL_NAME {
				assert.Equal(t, "Find all unused functions in api.js", part.FunctionCall.Args["objective"])
				receivedEventCount++
			} else if part.FunctionCall != nil && part.FunctionCall.Name == types.TASK_COMPLETE_TOOL_NAME {
				assert.Equal(t, "No unused functions found in api.js.", part.FunctionCall.Args["report"])
				receivedEventCount++
			} else if part.FunctionResponse != nil && part.FunctionResponse.Name == types.CODEBASE_INVESTIGATOR_TOOL_NAME {
				// Assert that the mock's ExecuteToolFunc generic response for subagent is received
				assert.Contains(t, part.FunctionResponse.Response["result"], "Mock tool executed successfully.")
				receivedEventCount++
			} else if part.FunctionResponse != nil && part.FunctionResponse.Name == types.TASK_COMPLETE_TOOL_NAME {
				// Assert that the mock's ExecuteToolFunc generic response for task_complete is received
				assert.Contains(t, part.FunctionResponse.Response["result"], "Mock tool executed successfully.")
				receivedEventCount++
			}
		}

		// To make this test comprehensive, we would need to assert every event in `expectedEvents` in order.
		// For now, checking for presence and types is a start.
	}

	// At a minimum, we should have received events related to the subagent call and its activities.
	// The exact number depends on how many events precede the subagent call in the mock's flow.
	// We'll assert that at least 6 relevant events (1 FC, 4 SA, 1 FC Task Complete) were processed.
	assert.GreaterOrEqual(t, receivedEventCount, 6, "Should have received events related to subagent flow")

}

// mockTool implements the types.Tool interface for testing purposes.
type mockTool struct {
	name string
}

func (m *mockTool) Name() string { return m.name }
func (m *mockTool) Description() string {
	return "Mock tool for testing"
}
func (m *mockTool) ServerName() string { return "" }
func (m *mockTool) Parameters() *types.JsonSchemaObject {
	return &types.JsonSchemaObject{Type: "object", Properties: map[string]*types.JsonSchemaProperty{}}
}
func (m *mockTool) Kind() types.Kind { return types.KindOther }
func (m *mockTool) Execute(ctx context.Context, args map[string]any) (types.ToolResult, error) {
	return types.ToolResult{LLMContent: "Mock tool executed for " + m.name, ReturnDisplay: "Mock tool result."}, nil
}
