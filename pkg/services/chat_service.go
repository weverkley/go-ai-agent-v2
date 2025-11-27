package services

import (
	"context"
	"errors"
	"fmt"
	"go-ai-agent-v2/go-cli/pkg/telemetry"
	"os"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/core"
	"go-ai-agent-v2/go-cli/pkg/routing"
	"go-ai-agent-v2/go-cli/pkg/types"

	"google.golang.org/api/googleapi"
)

var EventChanKey = struct{}{}

// ChatService orchestrates the interactive chat session, handling the tool-calling loop.
type ChatService struct {
	executor             core.Executor
	toolRegistry         types.ToolRegistryInterface
	history              []*types.Content
	sessionService       *SessionService
	sessionID            string
	settingsService      types.SettingsServiceIface
	contextService       *ContextService
	appConfig            types.Config
	generationConfig     types.GenerateContentConfig
	tokenUsage           map[string]*types.ModelTokenUsage
	proceedAlwaysTools   map[string]bool
	toolCallCounter      int
	toolErrorCounter     int
	ToolConfirmationChan chan types.ToolConfirmationOutcome
	userConfirmationChan chan bool
}

// NewChatService creates a new ChatService.
func NewChatService(executor core.Executor, toolRegistry types.ToolRegistryInterface, sessionService *SessionService, sessionID string, settingsService types.SettingsServiceIface, contextService *ContextService, appConfig types.Config, generationConfig types.GenerateContentConfig, initialState *types.ChatState) (*ChatService, error) {
	// Prepend context to system instruction
	contextContent := contextService.GetContext()
	if contextContent != "" {
		generationConfig.SystemInstruction = contextContent + "\n\n" + generationConfig.SystemInstruction
	}

	cs := &ChatService{
		executor:             executor,
		toolRegistry:         toolRegistry,
		sessionService:       sessionService,
		sessionID:            sessionID,
		settingsService:      settingsService,
		contextService:       contextService,
		appConfig:            appConfig,
		generationConfig:     generationConfig,
		tokenUsage:           make(map[string]*types.ModelTokenUsage),
		proceedAlwaysTools:   make(map[string]bool),
		ToolConfirmationChan: make(chan types.ToolConfirmationOutcome, 1),
		userConfirmationChan: make(chan bool, 1),
	}

	executor.SetToolConfirmationChannel(cs.ToolConfirmationChan)
	executor.SetUserConfirmationChannel(cs.userConfirmationChan)

	if initialState != nil {
		cs.history = initialState.History
		cs.proceedAlwaysTools = initialState.ProceedAlwaysTools
		cs.toolCallCounter = initialState.ToolCallCounter
		cs.toolErrorCounter = initialState.ToolErrorCounter
	} else {
		initialHistory, err := sessionService.LoadHistory(sessionID)
		if err != nil {
			return nil, fmt.Errorf("failed to load session history for ID %s: %w", sessionID, err)
		}
		cs.history = initialHistory
	}

	if err := cs.sessionService.SaveHistory(cs.sessionID, cs.history); err != nil {
		return nil, fmt.Errorf("failed to save initial session history: %w", err)
	}
	return cs, nil
}

// SendMessage starts the conversation loop for a user's message and returns a channel of events.
func (cs *ChatService) SendMessage(ctx context.Context, userInput string) (<-chan any, error) {
	eventChan := make(chan any)

	cs.history = append(cs.history, &types.Content{
		Role:  "user",
		Parts: []types.Part{{Text: userInput}},
	})

	go func() {
		defer close(eventChan)
		eventChan <- types.StreamingStartedEvent{}

		for { // Main loop for multi-turn tool calls
			select {
			case <-ctx.Done():
				eventChan <- types.ErrorEvent{Err: ctx.Err()}
				return
			default:
			}

			eventChan <- types.ThinkingEvent{}

			stream, err := cs.executor.StreamContent(ctx, cs.history...)
			if err != nil {
				eventChan <- types.ErrorEvent{Err: err}
				return
			}

			var modelResponseParts []types.Part
			var functionCalls []*types.FunctionCall
			var textResponse strings.Builder
			var streamErr error

			for event := range stream {
				switch e := event.(type) {
				case types.Part:
					modelResponseParts = append(modelResponseParts, e)
					if e.FunctionCall != nil {
						functionCalls = append(functionCalls, e.FunctionCall)
					}
					if e.Text != "" {
						textResponse.WriteString(e.Text)
						eventChan <- e
					}
				case types.TokenCountEvent:
					if _, ok := cs.tokenUsage[cs.executor.Name()]; !ok {
						cs.tokenUsage[cs.executor.Name()] = &types.ModelTokenUsage{}
					}
					cs.tokenUsage[cs.executor.Name()].InputTokens += e.InputTokens
					cs.tokenUsage[cs.executor.Name()].OutputTokens += e.OutputTokens
				case types.ErrorEvent:
					streamErr = e.Err
					goto EndStream
				}
			}

		EndStream:
			if streamErr != nil {
				currentExecutorTypeVal, _ := cs.settingsService.Get("executor")
				currentExecutorType, ok := currentExecutorTypeVal.(string)
				if !ok {
					currentExecutorType = "gemini"
				}

				if strings.HasPrefix(currentExecutorType, "gemini") {
					var apiErr *googleapi.Error
					if errors.As(streamErr, &apiErr) && apiErr.Code == 429 {
						telemetry.LogDebugf("Gemini Quota Exceeded error detected: %v", streamErr)

						router := routing.NewModelRouterService(cs.appConfig)
						routingCtx := &routing.RoutingContext{
							Request:      userInput,
							Signal:       ctx,
							IsFallback:   true,
							ExecutorType: "gemini",
						}

						decision, routeErr := router.Route(routingCtx, cs.appConfig)
						if routeErr != nil || decision == nil || strings.HasPrefix(decision.Model, "gemini") {
							eventChan <- types.ErrorEvent{
								Err: fmt.Errorf("Quota Exceeded for Gemini. No fallback model found. Please try again later or switch models manually."),
							}
							return
						}

						eventChan <- types.ModelSwitchEvent{
							OldModel: currentExecutorType,
							NewModel: decision.Model,
							Reason:   "Quota Exceeded",
						}

						fallbackExecutorType := "qwen"
						executorFactory, factoryErr := core.NewExecutorFactory(fallbackExecutorType, cs.appConfig)
						if factoryErr != nil {
							eventChan <- types.ErrorEvent{Err: fmt.Errorf("failed to create factory for fallback model %s: %w", decision.Model, factoryErr)}
							return
						}
						newExecutor, executorErr := executorFactory.NewExecutor(cs.appConfig, types.GenerateContentConfig{}, cs.history)
						if executorErr != nil {
							eventChan <- types.ErrorEvent{Err: fmt.Errorf("failed to create fallback executor %s: %w", decision.Model, executorErr)}
							return
						}

						cs.executor = newExecutor
						cs.settingsService.Set("executor", decision.Model)
						cs.settingsService.Set("model", decision.Model)
						continue // Re-attempt streaming with the new executor
					}
				}
				eventChan <- types.ErrorEvent{Err: streamErr}
				return
			}

			cs.history = append(cs.history, &types.Content{Role: "model", Parts: modelResponseParts})

			if len(functionCalls) > 0 {
				var toolResponseParts []types.Part
				for _, fc := range functionCalls {
					var toolExecutionResult any = nil
					var toolExecutionError error = nil
					cs.toolCallCounter++
					toolCallID := fc.ID

					isDangerousTool := false
					for _, dt := range cs.settingsService.GetDangerousTools() {
						if fc.Name == dt {
							isDangerousTool = true
							break
						}
					}

					eventChan <- types.ToolCallStartEvent{ToolCallID: toolCallID, ToolName: fc.Name, Args: fc.Args}

					if isDangerousTool && !cs.proceedAlwaysTools[fc.Name] {
						confirmationEvent := types.ToolConfirmationRequestEvent{
							ToolCallID: toolCallID,
							ToolName:   fc.Name,
							ToolArgs:   fc.Args,
							Type:       "exec",
							Message:    fmt.Sprintf("Confirm execution of tool '%s'?", fc.Name),
						}
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
									if originalContentBytes, err := os.ReadFile(filePath); err == nil {
										confirmationEvent.OriginalContent = string(originalContentBytes)
										confirmationEvent.FileDiff = generateDiff(string(originalContentBytes), newContent)
									}
								}
							}
						}

						eventChan <- confirmationEvent
						outcome := <-cs.ToolConfirmationChan

						switch outcome {
						case types.ToolConfirmationOutcomeProceedOnce:
							if fc.Name == types.USER_CONFIRM_TOOL_NAME {
								toolExecutionResult = "continue"
							} else {
								toolExecutionResult, toolExecutionError = executeTool(context.WithValue(ctx, EventChanKey, eventChan), fc, cs.toolRegistry, telemetry.GlobalLogger)
							}
						case types.ToolConfirmationOutcomeProceedAlways:
							cs.proceedAlwaysTools[fc.Name] = true
							if fc.Name == types.USER_CONFIRM_TOOL_NAME {
								toolExecutionResult = "continue"
							} else {
								toolExecutionResult, toolExecutionError = executeTool(context.WithValue(ctx, EventChanKey, eventChan), fc, cs.toolRegistry, telemetry.GlobalLogger)
							}
						case types.ToolConfirmationOutcomeCancel:
							toolExecutionResult = "Tool execution cancelled by user."
							toolExecutionError = fmt.Errorf("tool execution cancelled by user")
							cs.toolErrorCounter++
						default:
							toolExecutionResult = "Unknown confirmation outcome."
							toolExecutionError = fmt.Errorf("unknown confirmation outcome")
							cs.toolErrorCounter++
						}
					} else {
						toolExecutionResult, toolExecutionError = executeTool(context.WithValue(ctx, EventChanKey, eventChan), fc, cs.toolRegistry, telemetry.GlobalLogger)
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
							eventChan <- types.TodosSummaryUpdateEvent{Summary: fmt.Sprintf("Todos %d/%d", completed, total)}
						}
					}

					eventChan <- types.ToolCallEndEvent{ToolCallID: toolCallID, ToolName: fc.Name, Result: fmt.Sprintf("%v", toolExecutionResult), Err: toolExecutionError}

					toolResponseParts = append(toolResponseParts, types.Part{
						FunctionResponse: &types.FunctionResponse{
							Name:     fc.Name,
							Response: map[string]any{"result": toolExecutionResult},
						},
					})
				}
				cs.history = append(cs.history, &types.Content{Role: "tool", Parts: toolResponseParts})
				continue
			}

			if textResponse.Len() > 0 {
				eventChan <- types.FinalResponseEvent{Content: textResponse.String()}
			}

			break
		}

		if err := cs.sessionService.SaveHistory(cs.sessionID, cs.history); err != nil {
			telemetry.LogErrorf("Failed to save history: %v", err)
		}
	}()

	return eventChan, nil
}

func executeTool(ctx context.Context, fc *types.FunctionCall, toolRegistry types.ToolRegistryInterface, logger telemetry.TelemetryLogger) (any, error) {
	logger.LogDebugf("Executing tool '%s' with args: %v", fc.Name, fc.Args)

	tool, err := toolRegistry.GetTool(fc.Name)
	if err != nil {
		logger.LogErrorf("Error getting tool '%s': %v", fc.Name, err)
		return nil, fmt.Errorf("tool %s not found: %w", fc.Name, err)
	}

	result, err := tool.Execute(ctx, fc.Args)
	if err != nil {
		logger.LogErrorf("Tool '%s' execution failed: %v", fc.Name, err)
		// If the result contains a ToolError, we should propagate its details.
		if result.Error != nil {
			return result.LLMContent, result.Error
		}
		return nil, err
	}

	// Also handle cases where the tool execution itself doesn't fail but returns a business logic error.
	if result.Error != nil {
		logger.LogErrorf("Tool '%s' executed with a ToolError: %v", fc.Name, result.Error)
		return result.LLMContent, result.Error
	}

	logger.LogInfof("Tool '%s' executed successfully. Result: %v", fc.Name, result.LLMContent)
	return result.LLMContent, nil
}

func generateDiff(oldContent, newContent string) string {
	oldLines := strings.Split(oldContent, "\n")
	newLines := strings.Split(newContent, "\n")
	maxLength := 20
	if len(oldLines) > maxLength || len(newLines) > maxLength {
		return "Diff too large to display."
	}
	var diff strings.Builder
	// This is a placeholder for a real diff algorithm
	diff.WriteString("--- a\n")
	diff.WriteString("+++ b\n")
	for _, line := range oldLines {
		diff.WriteString(fmt.Sprintf("-%s\n", line))
	}
	for _, line := range newLines {
		diff.WriteString(fmt.Sprintf("+%s\n", line))
	}
	return diff.String()
}

func (cs *ChatService) GetHistory() []*types.Content {
	return cs.history
}
func (cs *ChatService) ClearHistory() {
	cs.history = []*types.Content{}
	if err := cs.sessionService.SaveHistory(cs.sessionID, cs.history); err != nil {
		telemetry.LogErrorf("Failed to clear persisted history for session %s: %v", cs.sessionID, err)
	}
}
func (cs *ChatService) GetState() *types.ChatState {
	return &types.ChatState{
		History:            cs.history,
		SessionID:          cs.sessionID,
		ProceedAlwaysTools: cs.proceedAlwaysTools,
		ToolCallCounter:    cs.toolCallCounter,
		ToolErrorCounter:   cs.toolErrorCounter,
	}
}
func (cs *ChatService) GetToolRegistry() types.ToolRegistryInterface { return cs.toolRegistry }
func (cs *ChatService) GetExecutor() core.Executor                   { return cs.executor }
func (cs *ChatService) GetSettingsService() types.SettingsServiceIface {
	return cs.settingsService
}

func (cs *ChatService) GetContextService() *ContextService {
	return cs.contextService
}
func (cs *ChatService) GetToolConfirmationChannel() chan types.ToolConfirmationOutcome {
	return cs.ToolConfirmationChan
}
func (cs *ChatService) GetUserConfirmationChannel() chan bool {
	return cs.userConfirmationChan
}

// CompressHistory compresses the current chat history.
func (cs *ChatService) CompressHistory() (*types.ChatCompressionResult, error) {
	if cs.executor == nil {
		return nil, fmt.Errorf("executor is not initialized")
	}

	promptID := "compress-" + cs.sessionID // Or generate a unique ID
	result, err := cs.executor.CompressChat(cs.history, promptID)
	if err != nil {
		return nil, err
	}

	if _, ok := cs.tokenUsage[cs.executor.Name()]; !ok {
		cs.tokenUsage[cs.executor.Name()] = &types.ModelTokenUsage{}
	}
	cs.tokenUsage[cs.executor.Name()].InputTokens += result.InputTokens
	cs.tokenUsage[cs.executor.Name()].OutputTokens += result.OutputTokens

	// Create a new history with the system prompt and the summary
	newHistory := []*types.Content{}
	if cs.generationConfig.SystemInstruction != "" {
		newHistory = append(newHistory, &types.Content{
			Role:  "system",
			Parts: []types.Part{{Text: cs.generationConfig.SystemInstruction}},
		})
	}
	newHistory = append(newHistory, &types.Content{
		Role:  "user",
		Parts: []types.Part{{Text: "Summary of previous conversation:\n" + result.Summary}},
	})
	newHistory = append(newHistory, &types.Content{
		Role:  "model",
		Parts: []types.Part{{Text: "Okay, I have reviewed the summary and am ready to continue."}},
	})

	cs.history = newHistory

	// Save the new compressed history
	if err := cs.sessionService.SaveHistory(cs.sessionID, cs.history); err != nil {
		// Log the error but don't fail the whole operation, as the compression itself succeeded.
		telemetry.LogErrorf("Failed to save compressed history for session %s: %v", cs.sessionID, err)
	}

	return result, nil
}
func (cs *ChatService) GetGenerationConfig() types.GenerateContentConfig {
	return cs.generationConfig
}

func (cs *ChatService) GetTokenUsage() map[string]*types.ModelTokenUsage {
	return cs.tokenUsage
}
