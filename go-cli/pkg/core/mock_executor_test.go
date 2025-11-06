package core

import (
	"reflect"
	"testing"
	"time"

	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/google/generative-ai-go/genai"
)

func TestNewMockExecutor(t *testing.T) {
	// Test case 1: No default responses provided
	me := NewMockExecutor(nil, nil)

	if me == nil {
		t.Errorf("NewMockExecutor returned nil")
	}

	// Test case 2: Default GenerateContentResponse provided
	defaultGenContentResp := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: &genai.Content{
					Parts: []genai.Part{genai.Text("Custom GenerateContent response.")},
				},
			},
		},
	}
	me = NewMockExecutor(defaultGenContentResp, nil)
	if !reflect.DeepEqual(me.DefaultGenerateContentResponse, defaultGenContentResp) {
		t.Errorf("NewMockExecutor did not set DefaultGenerateContentResponse correctly")
	}

	// Test case 3: Default ExecuteToolResult provided
	defaultToolResult := &types.ToolResult{
		LLMContent:    "Custom tool result",
		ReturnDisplay: "Custom tool display",
	}
	me = NewMockExecutor(nil, defaultToolResult)
	if !reflect.DeepEqual(me.DefaultExecuteToolResult, defaultToolResult) {
		t.Errorf("NewMockExecutor did not set DefaultExecuteToolResult correctly")
	}
}

func TestMockExecutor_GenerateContent(t *testing.T) {
	// Test case 1: DefaultGenerateContentResponse is set
	expectedResp := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: &genai.Content{
					Parts: []genai.Part{genai.Text("Test response.")},
				},
			},
		},
	}
	me := NewMockExecutor(expectedResp, nil)
	resp, err := me.GenerateContent(&genai.Content{Parts: []genai.Part{genai.Text("prompt")}})
	if err != nil {
		t.Fatalf("GenerateContent returned an error: %v", err)
	}
	if !reflect.DeepEqual(resp, expectedResp) {
		t.Errorf("GenerateContent returned unexpected response. Got %v, want %v", resp, expectedResp)
	}

	// Test case 2: DefaultGenerateContentResponse is nil
	me = NewMockExecutor(nil, nil)
	resp, err = me.GenerateContent(&genai.Content{Parts: []genai.Part{genai.Text("prompt")}})
	if err != nil {
		t.Fatalf("GenerateContent returned an error: %v", err)
	}
	if resp.Candidates[0].Content.Parts[0].(genai.Text) != "Mocked response from GenerateContent." {
		t.Errorf("GenerateContent returned unexpected default response: %v", resp)
	}
}

func TestMockExecutor_ExecuteTool(t *testing.T) {
	// Test case 1: DefaultExecuteToolResult is set
	expectedResult := types.ToolResult{
		LLMContent:    "Custom tool result",
		ReturnDisplay: "Custom tool display",
	}
	me := NewMockExecutor(nil, &expectedResult)
	result, err := me.ExecuteTool(&genai.FunctionCall{Name: "test_tool"})
	if err != nil {
		t.Fatalf("ExecuteTool returned an error: %v", err)
	}
	if !reflect.DeepEqual(result, expectedResult) {
		t.Errorf("ExecuteTool returned unexpected result. Got %v, want %v", result, expectedResult)
	}

	// Test case 2: DefaultExecuteToolResult is nil
	me = NewMockExecutor(nil, nil)
	result, err = me.ExecuteTool(&genai.FunctionCall{Name: "test_tool"})
	if err != nil {
		t.Fatalf("ExecuteTool returned an error: %v", err)
	}
	if result.LLMContent != "Mocked result for tool test_tool with args map[]" {
		t.Errorf("ExecuteTool returned unexpected default result: %v", result)
	}
}

func TestMockExecutor_SendMessageStream(t *testing.T) {
	me := NewMockExecutor(nil, nil)
	respChan, err := me.SendMessageStream("mock-model", types.MessageParams{}, "prompt-123")
	if err != nil {
		t.Fatalf("SendMessageStream returned an error: %v", err)
	}

	// Read the first chunk
	select {
	case resp := <-respChan:
		if resp.Type != types.StreamEventTypeChunk {
			t.Errorf("Expected chunk type, got %v", resp.Type)
		}
		if resp.Value == nil || len(resp.Value.Candidates) == 0 || resp.Value.Candidates[0].Content.Parts[0].(genai.Text) != "Mocked streamed response chunk 1." {
			t.Errorf("Unexpected first chunk: %v", resp.Value)
		}
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for first stream chunk")
	}

	// Read the second chunk
	select {
	case resp := <-respChan:
		if resp.Type != types.StreamEventTypeChunk {
			t.Errorf("Expected chunk type, got %v", resp.Type)
		}
		if resp.Value == nil || len(resp.Value.Candidates) == 0 || resp.Value.Candidates[0].Content.Parts[0].(genai.Text) != "Mocked streamed response chunk 2." {
			t.Errorf("Unexpected second chunk: %v", resp.Value)
		}
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for second stream chunk")
	}

	// Ensure the channel is closed
	select {
	case _, ok := <-respChan:
		if ok {
			t.Errorf("Stream channel not closed")
		}
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for stream channel to close")
	}
}

func TestMockExecutor_ListModels(t *testing.T) {
	me := NewMockExecutor(nil, nil)
	models, err := me.ListModels()
	if err != nil {
		t.Fatalf("ListModels returned an error: %v", err)
	}

	expectedModels := []string{"mock-model-1", "mock-model-2"}
	if !reflect.DeepEqual(models, expectedModels) {
		t.Errorf("ListModels returned unexpected models. Got %v, want %v", models, expectedModels)
	}
}
