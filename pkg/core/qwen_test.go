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

	eventChan, err := qwenChat.StreamContent(context.Background(), &types.Content{Parts: []types.Part{{Text: "test"}}})
	assert.NoError(t, err)

	var receivedParts []types.Part
	for event := range eventChan {
		if part, ok := event.(types.Part); ok {
			receivedParts = append(receivedParts, part)
		}
	}

	assert.Len(t, receivedParts, 1)
	assert.Equal(t, "Hello", receivedParts[0].Text)
}

type TestTool struct {
	*types.BaseDeclarativeTool
}

func (t *TestTool) Execute(ctx context.Context, args map[string]any) (types.ToolResult, error) {
	return types.ToolResult{ReturnDisplay: "tool result"}, nil
}

func TestGenerateStreamWithToolCalling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		// The server now only needs to return the tool call delta.
		// The ChatService is responsible for executing it and re-prompting.
		fmt.Fprintln(w, "data: {\"id\":\"chatcmpl-123\",\"object\":\"chat.completion.chunk\",\"created\":1694268190,\"model\":\"qwen-turbo\",\"choices\":[{\"index\":0,\"delta\":{\"role\":\"assistant\",\"tool_calls\":[{\"index\":0,\"id\":\"call_123\",\"type\":\"function\",\"function\":{\"name\":\"test_tool\",\"arguments\":\"{\\\"arg1\\\":\\\"val1\\\"}\"}}]},\"finish_reason\":\"tool_calls\"}]}")
		fmt.Fprintln(w, "data: [DONE]")
	}))
	defer server.Close()

	os.Setenv("QWEN_API_KEY", "test-key")
	defer os.Unsetenv("QWEN_API_KEY")

	openaiConfig := openai.DefaultConfig("test-key")
	openaiConfig.BaseURL = server.URL
	client := openai.NewClientWithConfig(openaiConfig)

	toolRegistry := types.NewToolRegistry()
	// No need to register the tool here anymore for this test,
	// as the executor is not executing it.

	qwenChat := &QwenChat{
		client:       client,
		modelName:    "qwen-turbo",
		toolRegistry: toolRegistry,
	}

	eventChan, err := qwenChat.StreamContent(context.Background(), &types.Content{Parts: []types.Part{{Text: "test"}}})
	assert.NoError(t, err)

	var receivedParts []types.Part
	for event := range eventChan {
		if part, ok := event.(types.Part); ok {
			receivedParts = append(receivedParts, part)
		}
	}

	// We expect exactly one part: the function call.
	assert.Len(t, receivedParts, 1)
	assert.NotNil(t, receivedParts[0].FunctionCall)
	assert.Equal(t, "test_tool", receivedParts[0].FunctionCall.Name)
	assert.Equal(t, "call_123", receivedParts[0].FunctionCall.ID)
	assert.Equal(t, "val1", receivedParts[0].FunctionCall.Args["arg1"])
}
