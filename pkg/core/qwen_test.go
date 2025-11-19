package core

import (
	"context" // New import
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
)

func TestNewQwenChat(t *testing.T) {
	os.Setenv("QWEN_API_KEY", "test-key")
	defer os.Unsetenv("QWEN_API_KEY")

	cfg := config.NewConfig(&config.ConfigParameters{
		ModelName: "qwen-turbo",
	})

	executor, err := NewQwenChat(cfg, types.GenerateContentConfig{}, nil)
	assert.NoError(t, err)
	assert.NotNil(t, executor)

	qwenChat, ok := executor.(*QwenChat)
	assert.True(t, ok)
	assert.Equal(t, "qwen-turbo", qwenChat.modelName)
}

func TestGenerateStream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintln(w, "data: {\"id\":\"chatcmpl-123\",\"object\":\"chat.completion.chunk\",\"created\":1694268190,\"model\":\"qwen-turbo\",\"choices\":[{\"index\":0,\"delta\":{\"role\":\"assistant\",\"content\":\"Hello\"},\"finish_reason\":\"\"}]}")
		fmt.Fprintln(w, "data: [DONE]")
	}))
	defer server.Close()

	os.Setenv("QWEN_API_KEY", "test-key")
	defer os.Unsetenv("QWEN_API_KEY")

	openaiConfig := openai.DefaultConfig("test-key")
	openaiConfig.BaseURL = server.URL
	client := openai.NewClientWithConfig(openaiConfig)

	qwenChat := &QwenChat{
		client:    client,
		modelName: "qwen-turbo", // Added missing field
	}

	eventChan, err := qwenChat.GenerateStream(context.Background(), &types.Content{Parts: []types.Part{{Text: "test"}}})
	assert.NoError(t, err)

	var events []any
	for event := range eventChan {
		events = append(events, event)
	}

	assert.Len(t, events, 3)
	assert.IsType(t, types.StreamingStartedEvent{}, events[0])
	assert.IsType(t, types.ThinkingEvent{}, events[1])
	assert.IsType(t, types.FinalResponseEvent{}, events[2])
	assert.Equal(t, "Hello", events[2].(types.FinalResponseEvent).Content)
}

type TestTool struct {
	*types.BaseDeclarativeTool
}

func (t *TestTool) Execute(ctx context.Context, args map[string]any) (types.ToolResult, error) {
	return types.ToolResult{ReturnDisplay: "tool result"}, nil
}

func TestGenerateStreamWithToolCalling(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		if callCount == 0 {
			// First call: return tool call
			fmt.Fprintln(w, "data: {\"id\":\"chatcmpl-123\",\"object\":\"chat.completion.chunk\",\"created\":1694268190,\"model\":\"qwen-turbo\",\"choices\":[{\"index\":0,\"delta\":{\"role\":\"assistant\",\"tool_calls\":[{\"index\":0,\"id\":\"call_123\",\"type\":\"function\",\"function\":{\"name\":\"test_tool\",\"arguments\":\"{\\\"arg1\\\":\\\"val1\\\"}\"}}]},\"finish_reason\":\"tool_calls\"}]}")
			fmt.Fprintln(w, "data: [DONE]")
		} else {
			// Second call: return final response
			fmt.Fprintln(w, "data: {\"id\":\"chatcmpl-456\",\"object\":\"chat.completion.chunk\",\"created\":1694268191,\"model\":\"qwen-turbo\",\"choices\":[{\"index\":0,\"delta\":{\"role\":\"assistant\",\"content\":\"Final response\"},\"finish_reason\":\"stop\"}]}")
			fmt.Fprintln(w, "data: [DONE]")
		}
		callCount++
	}))
	defer server.Close()

	os.Setenv("QWEN_API_KEY", "test-key")
	defer os.Unsetenv("QWEN_API_KEY")

	openaiConfig := openai.DefaultConfig("test-key")
	openaiConfig.BaseURL = server.URL
	client := openai.NewClientWithConfig(openaiConfig)

	toolRegistry := types.NewToolRegistry()
	tool := &TestTool{
		BaseDeclarativeTool: types.NewBaseDeclarativeTool(
			"test_tool",
			"Test Tool",
			"A tool for testing",
			types.KindOther,
			&types.JsonSchemaObject{
				Type: "object",
				Properties: map[string]*types.JsonSchemaProperty{
					"arg1": &types.JsonSchemaProperty{
						Type:        "string",
						Description: "An argument",
					},
				},
			},
			false,
			false,
			nil,
		),
	}
	toolRegistry.Register(tool)

	qwenChat := &QwenChat{
		client:       client,
		modelName:    "qwen-turbo",
		toolRegistry: toolRegistry,
	}

	eventChan, err := qwenChat.GenerateStream(context.Background(), &types.Content{Parts: []types.Part{{Text: "test"}}})
	assert.NoError(t, err)

	var events []any
	for event := range eventChan {
		events = append(events, event)
	}

	assert.Len(t, events, 5)
	assert.IsType(t, types.StreamingStartedEvent{}, events[0])
	assert.IsType(t, types.ThinkingEvent{}, events[1])
	assert.IsType(t, types.ToolCallStartEvent{}, events[2])
	assert.IsType(t, types.ToolCallEndEvent{}, events[3])
	assert.IsType(t, types.FinalResponseEvent{}, events[4])
	assert.Equal(t, "Final response", events[4].(types.FinalResponseEvent).Content)
}
