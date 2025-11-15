package core_test

import (
	"context" // New import
	"os"
	"testing"

	"go-ai-agent-v2/go-cli/pkg/core"
	"go-ai-agent-v2/go-cli/pkg/routing" // Add this line
	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/google/generative-ai-go/genai"
	"github.com/stretchr/testify/assert"
)

// MockConfig implements types.Config for testing purposes.
type MockConfig struct {
	ModelName string
	ToolRegistry *types.ToolRegistry
	DebugMode bool
	CodebaseInvestigatorSettings *types.CodebaseInvestigatorSettings
}

func (m *MockConfig) WithModel(modelName string) types.Config {
	return &MockConfig{ModelName: modelName}
}

func (m *MockConfig) Get(key string) (interface{}, bool) {
	switch key {
	case "model": // Changed from "modelName" to "model"
		return m.ModelName, true
	case "toolRegistry":
		return m.ToolRegistry, true
	case "debugMode":
		return m.DebugMode, true
	case "codebaseInvestigatorSettings":
		return m.CodebaseInvestigatorSettings, true
	default:
		return nil, false
	}
}



func TestNewExecutorFactory(t *testing.T) {
	// Create a dummy config.Config instance for NewExecutorFactory
	dummyCfg := &MockConfig{}

	t.Run("should return GoaiagentExecutorFactory for 'goaiagent' type", func(t *testing.T) {
		factory, err := core.NewExecutorFactory("goaiagent", dummyCfg)
		assert.NoError(t, err)
		assert.IsType(t, &core.GoaiagentExecutorFactory{}, factory)
	})

	t.Run("should return MockExecutorFactory for 'mock' type", func(t *testing.T) {
		factory, err := core.NewExecutorFactory("mock", dummyCfg)
		assert.NoError(t, err)
		assert.IsType(t, &core.MockExecutorFactory{}, factory)
	})

	t.Run("should return error for unknown type", func(t *testing.T) {
		factory, err := core.NewExecutorFactory("unknown", dummyCfg)
		assert.Error(t, err)
		assert.Nil(t, factory)
		assert.Contains(t, err.Error(), "unknown executor type")
	})
}

func TestGeminiExecutorFactory_NewExecutor(t *testing.T) {
	t.Run("should create a GeminiChat instance", func(t *testing.T) {
		factory := &core.GoaiagentExecutorFactory{
			Router: routing.NewModelRouterService(&MockConfig{}), // Pass MockConfig here
		}
		mockConfig := &MockConfig{ModelName: "gemini-pro"}
		
		// Temporarily set GEMINI_API_KEY for this test
		os.Setenv("GEMINI_API_KEY", "test-api-key")
		defer os.Unsetenv("GEMINI_API_KEY")

		executor, err := factory.NewExecutor(mockConfig, types.GenerateContentConfig{}, []*genai.Content{})
		assert.NoError(t, err)
		assert.IsType(t, &core.GoaiagentChat{}, executor)
	})

	t.Run("should return error if GEMINI_API_KEY is not set", func(t *testing.T) {
		factory := &core.GoaiagentExecutorFactory{
			Router: routing.NewModelRouterService(&MockConfig{}), // Pass MockConfig here
		}
		mockConfig := &MockConfig{ModelName: "gemini-pro"}

		os.Unsetenv("GEMINI_API_KEY") // Ensure it's unset

		executor, err := factory.NewExecutor(mockConfig, types.GenerateContentConfig{}, []*genai.Content{})
		assert.Error(t, err)
		assert.Nil(t, executor)
		assert.Contains(t, err.Error(), "GEMINI_API_KEY environment variable not set")
	})
}

func TestMockExecutorFactory_NewExecutor(t *testing.T) {
	t.Run("should create a new MockExecutor instance", func(t *testing.T) {
		factory := &core.MockExecutorFactory{}
		mockConfig := &MockConfig{ModelName: "mock-model"}

		executor, err := factory.NewExecutor(mockConfig, types.GenerateContentConfig{}, []*genai.Content{})
		assert.NoError(t, err)
		assert.IsType(t, &core.MockExecutor{}, executor)
	})

	t.Run("should return the provided mock instance if set", func(t *testing.T) {
		expectedMock := &core.MockExecutor{
			GenerateContentFunc: func(contents ...*genai.Content) (*genai.GenerateContentResponse, error) {
				return &genai.GenerateContentResponse{}, nil
			},
		}
		factory := &core.MockExecutorFactory{Mock: expectedMock}
		mockConfig := &MockConfig{ModelName: "mock-model"}

		executor, err := factory.NewExecutor(mockConfig, types.GenerateContentConfig{}, []*genai.Content{})
		assert.NoError(t, err)
		assert.Same(t, expectedMock, executor)
	})
}

func TestMockExecutor(t *testing.T) {
	t.Run("GenerateContent should call the provided function", func(t *testing.T) {
		expectedResp := &genai.GenerateContentResponse{}
		mockExecutor := &core.MockExecutor{
			GenerateContentFunc: func(contents ...*genai.Content) (*genai.GenerateContentResponse, error) {
				return expectedResp, nil
			},
		}
		resp, err := mockExecutor.GenerateContent()
		assert.NoError(t, err)
		assert.Same(t, expectedResp, resp)
	})

	t.Run("GenerateContent should return error if function not provided", func(t *testing.T) {
		mockExecutor := &core.MockExecutor{}
		resp, err := mockExecutor.GenerateContent()
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "GenerateContent not implemented in mock")
	})

	t.Run("ExecuteTool should call the provided function", func(t *testing.T) {
		expectedResult := types.ToolResult{ReturnDisplay: "mocked tool result"}
		mockExecutor := &core.MockExecutor{
			ExecuteToolFunc: func(ctx context.Context, fc *genai.FunctionCall) (types.ToolResult, error) {
				return expectedResult, nil
			},
		}
		result, err := mockExecutor.ExecuteTool(context.Background(), &genai.FunctionCall{})
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
	})

	t.Run("ExecuteTool should return error if function not provided", func(t *testing.T) {
		mockExecutor := &core.MockExecutor{}
		result, err := mockExecutor.ExecuteTool(context.Background(), &genai.FunctionCall{})
		assert.Error(t, err)
		assert.Equal(t, types.ToolResult{}, result)
		assert.Contains(t, err.Error(), "ExecuteTool not implemented in mock")
	})

	t.Run("SendMessageStream should call the provided function", func(t *testing.T) {
		bidirectionalChan := make(chan types.StreamResponse)
		close(bidirectionalChan)
		expectedChan := (<-chan types.StreamResponse)(bidirectionalChan)
		mockExecutor := &core.MockExecutor{
			SendMessageStreamFunc: func(modelName string, messageParams types.MessageParams, promptId string) (<-chan types.StreamResponse, error) {
				return expectedChan, nil
			},
		}
		respChan, err := mockExecutor.SendMessageStream("", types.MessageParams{}, "")
		assert.NoError(t, err)
		assert.Equal(t, expectedChan, respChan)
	})

	t.Run("SendMessageStream should return error if function not provided", func(t *testing.T) {
		mockExecutor := &core.MockExecutor{}
		respChan, err := mockExecutor.SendMessageStream("", types.MessageParams{}, "")
		assert.Error(t, err)
		assert.NotNil(t, respChan) // Expect a non-nil, but closed channel
		assert.Contains(t, err.Error(), "SendMessageStream not implemented in mock")
	})

	t.Run("ListModels should call the provided function", func(t *testing.T) {
		expectedModels := []string{"model-a", "model-b"}
		mockExecutor := &core.MockExecutor{
			ListModelsFunc: func() ([]string, error) {
				return expectedModels, nil
			},
		}
		models, err := mockExecutor.ListModels()
		assert.NoError(t, err)
		assert.Equal(t, expectedModels, models)
	})

	t.Run("ListModels should return error if function not provided", func(t *testing.T) {
		mockExecutor := &core.MockExecutor{}
		models, err := mockExecutor.ListModels()
		assert.Error(t, err)
		assert.Nil(t, models)
		assert.Contains(t, err.Error(), "ListModels not implemented in mock")
	})

	t.Run("GetHistory should call the provided function", func(t *testing.T) {
		expectedHistory := []*genai.Content{{Role: "user"}}
		mockExecutor := &core.MockExecutor{
			GetHistoryFunc: func() ([]*genai.Content, error) {
				return expectedHistory, nil
			},
		}
		history, err := mockExecutor.GetHistory()
		assert.NoError(t, err)
		assert.Equal(t, expectedHistory, history)
	})

	t.Run("GetHistory should return error if function not provided", func(t *testing.T) {
		mockExecutor := &core.MockExecutor{}
		history, err := mockExecutor.GetHistory()
		assert.Error(t, err)
		assert.Nil(t, history)
		assert.Contains(t, err.Error(), "GetHistory not implemented in mock")
	})

	t.Run("SetHistory should call the provided function", func(t *testing.T) {
		mockExecutor := &core.MockExecutor{
			SetHistoryFunc: func(history []*genai.Content) error {
				return nil
			},
		}
		err := mockExecutor.SetHistory([]*genai.Content{})
		assert.NoError(t, err)
	})

	t.Run("SetHistory should return error if function not provided", func(t *testing.T) {
		mockExecutor := &core.MockExecutor{}
		err := mockExecutor.SetHistory([]*genai.Content{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "SetHistory not implemented in mock")
	})

	t.Run("CompressChat should call the provided function", func(t *testing.T) {
		expectedResult := &types.ChatCompressionResult{CompressionStatus: "compressed"}
		mockExecutor := &core.MockExecutor{
			CompressChatFunc: func(promptId string, force bool) (*types.ChatCompressionResult, error) {
				return expectedResult, nil
			},
		}
		result, err := mockExecutor.CompressChat("", false)
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
	})

	t.Run("CompressChat should return error if function not provided", func(t *testing.T) {
		mockExecutor := &core.MockExecutor{}
		result, err := mockExecutor.CompressChat("", false)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "CompressChat not implemented in mock")
	})
}
