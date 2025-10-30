package agents

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/core"
	"go-ai-agent-v2/go-cli/pkg/tools"
	"go-ai-agent-v2/go-cli/pkg/types" // Added

	"github.com/google/generative-ai-go/genai"
)

// AgentTerminateMode defines the reasons an agent might terminate.
type AgentTerminateMode string

const (
	AgentTerminateModeError    AgentTerminateMode = "ERROR"
	AgentTerminateModeGoal     AgentTerminateMode = "GOAL"
	AgentTerminateModeMaxTurns AgentTerminateMode = "MAX_TURNS"
	AgentTerminateModeTimeout  AgentTerminateMode = "TIMEOUT"
	AgentTerminateModeAborted  AgentTerminateMode = "ABORTED"

	TASK_COMPLETE_TOOL_NAME = "complete_task"
)

// SubagentActivityEvent represents an activity event emitted by a subagent.
type SubagentActivityEvent struct {
	IsSubagentActivityEvent bool                   `json:"isSubagentActivityEvent"`
	AgentName               string                 `json:"agentName"`
	Type                    string                 `json:"type"`
	Data                    map[string]interface{} `json:"data"`
}

// ActivityCallback is a callback function to report on agent activity.
type ActivityCallback func(activity SubagentActivityEvent)

// AgentExecutor executes an agent loop based on an AgentDefinition.
type AgentExecutor struct {
	Definition     AgentDefinition
	AgentID        string
	ToolRegistry   *tools.ToolRegistry
	RuntimeContext *config.Config
	OnActivity     ActivityCallback
	parentPromptId string
}

// Run executes the agent loop.
func (ae *AgentExecutor) Run(inputs AgentInputs, ctx context.Context) (OutputObject, error) {
	startTime := time.Now()
	turnCounter := 0
	terminateReason := AgentTerminateModeError
	var finalResult *string = nil

	logAgentStart(
		ae.RuntimeContext,
		AgentStartEvent{AgentID: ae.AgentID, AgentName: ae.Definition.Name},
	)

	defer func() {
		logAgentFinish(
			ae.RuntimeContext,
			AgentFinishEvent{
				AgentID:         ae.AgentID,
				AgentName:       ae.Definition.Name,
				DurationMs:      time.Since(startTime).Milliseconds(),
				TurnCounter:     turnCounter,
				TerminateReason: terminateReason,
			},
		)
	}()

	chat, err := ae.createChatObject(inputs)
	if err != nil {
		ae.emitActivity("ERROR", map[string]interface{}{"error": err.Error()})
		return OutputObject{}, err
	}
	tools, err := ae.prepareToolsList()
	if err != nil {
		ae.emitActivity("ERROR", map[string]interface{}{"error": err.Error()})
		return OutputObject{}, err
	}

	query := "Get Started!"
	if ae.Definition.PromptConfig.Query != "" {
		query = templateString(ae.Definition.PromptConfig.Query, inputs)
	}
	currentMessage := []Part{{Text: query}}

	for {
		// Check for termination conditions like max turns or timeout.
		reason := ae.checkTermination(startTime, turnCounter)
		if reason != nil {
			terminateReason = *reason
			break
		}
		if ctx.Err() != nil { // Check for context cancellation (signal.aborted equivalent)
			terminateReason = AgentTerminateModeAborted
			break
		}

		promptId := fmt.Sprintf("%s#%d", ae.AgentID, turnCounter)
		turnCounter++

		functionCalls, textResponse, err := ae.callModel(chat, currentMessage, tools, ctx, promptId)
		if err != nil {
			ae.emitActivity("ERROR", map[string]interface{}{"error": err.Error()})
			return OutputObject{}, err
		}
		if ctx.Err() != nil {
			terminateReason = AgentTerminateModeAborted
			break
		}

		// If the model stops calling tools without calling complete_task, it's an error.
		if len(functionCalls) == 0 {
			terminateReason = AgentTerminateModeError
			finalResult = stringPtr(fmt.Sprintf("Agent stopped calling tools but did not call '%s' to finalize the session.", TASK_COMPLETE_TOOL_NAME))
			ae.emitActivity("ERROR", map[string]interface{}{
				"error":   *finalResult,
				"context": "protocol_violation",
			})
			break
		}

		nextMessageParts, submittedOutputVal, taskCompleted, err := ae.processFunctionCalls(functionCalls, ctx, promptId)
		if err != nil {
			ae.emitActivity("ERROR", map[string]interface{}{"error": err.Error()})
			return OutputObject{}, err
		}

		if taskCompleted {
			if submittedOutputVal != nil {
				finalResult = submittedOutputVal
			} else {
				temp := "Task completed successfully."
				finalResult = &temp
			}
			terminateReason = AgentTerminateModeGoal
			break
		}

		currentMessage = nextMessageParts
	}

	if terminateReason == AgentTerminateModeGoal {
		result := "Task completed."
		if finalResult != nil {
			result = *finalResult
		}
		return OutputObject{Result: result, TerminateReason: terminateReason}, nil
	}

	result := "Agent execution was terminated before completion."
	if finalResult != nil {
		result = *finalResult
	}
	return OutputObject{Result: result, TerminateReason: terminateReason}, nil
}

// createChatObject initializes a GeminiChat instance for the agent run.
func (ae *AgentExecutor) createChatObject(inputs AgentInputs) (*core.GeminiChat, error) {
	promptConfig := ae.Definition.PromptConfig
	modelConfig := ae.Definition.ModelConfig

	if promptConfig.SystemPrompt == "" && len(promptConfig.InitialMessages) == 0 {
		return nil, fmt.Errorf("PromptConfig must define either `systemPrompt` or `initialMessages`")
	}

	startHistory := ae.applyTemplateToInitialMessages(promptConfig.InitialMessages, inputs)

	// Build system instruction from the templated prompt string.
	var systemInstruction string
	if promptConfig.SystemPrompt != "" {
		var err error
		systemInstruction, err = ae.buildSystemPrompt(inputs)
		if err != nil {
			return nil, fmt.Errorf("failed to build system prompt: %w", err)
		}
	}

	generationConfig := types.GenerateContentConfig{
		Temperature: modelConfig.Temperature,
		TopP:        modelConfig.TopP,
		ThinkingConfig: &types.ThinkingConfig{
			IncludeThoughts: true,
			ThinkingBudget:  modelConfig.ThinkingBudget,
		},
		SystemInstruction: systemInstruction,
	}

	chat, err := core.NewGeminiChat(
		ae.RuntimeContext,
		generationConfig,
		startHistory,
	)
	if err != nil {
		// TODO: Implement reportError equivalent
		return nil, fmt.Errorf("failed to create chat object: %w", err)
	}

	return chat, nil
}

// applyTemplateToInitialMessages applies template strings to initial messages.
func (ae *AgentExecutor) applyTemplateToInitialMessages(
	initialMessages []Part,
	inputs AgentInputs,
) []genai.Content {
	templatedMessages := make([]genai.Content, len(initialMessages))
	for i, part := range initialMessages {
		var newGenaiParts []genai.Part
		if part.Text != "" {
			newGenaiParts = append(newGenaiParts, genai.Text(templateString(part.Text, inputs)))
		} else if part.FunctionResponse != nil {
			newGenaiParts = append(newGenaiParts, genai.FunctionResponse{
				Name:     part.FunctionResponse.Name,
				Response: part.FunctionResponse.Response,
			})
		} else if part.InlineData != nil {
			newGenaiParts = append(newGenaiParts, genai.Blob{
				MIMEType: part.InlineData.MimeType,
				Data:     []byte(part.InlineData.Data),
			})
		} else if part.FileData != nil {
			newGenaiParts = append(newGenaiParts, genai.Text(fmt.Sprintf("File data: %s (%s)", part.FileData.FileURL, part.FileData.MimeType)))
		}
		templatedMessages[i] = genai.Content{Parts: newGenaiParts}
	}
	return templatedMessages
}

// prepareToolsList prepares the list of tool function declarations to be sent to the model.
func (ae *AgentExecutor) prepareToolsList() ([]genai.FunctionDeclaration, error) {
	toolsList := []genai.FunctionDeclaration{}
	toolConfig := ae.Definition.ToolConfig
	outputConfig := ae.Definition.OutputConfig

	if toolConfig != nil {
		toolNamesToLoad := []string{}
		for _, toolRef := range toolConfig.Tools {
			// For now, we only handle tool names (strings).
			// The JS version also handles direct FunctionDeclaration objects and tool instances with schema.
			toolNamesToLoad = append(toolNamesToLoad, toolRef)
		}
		// Add schemas from tools that were registered by name.

		toolsList = append(toolsList, ae.ToolRegistry.GetFunctionDeclarationsFiltered(toolNamesToLoad)...)
	}

	// Always inject complete_task.
	// Configure its schema based on whether output is expected.
	completeTool := genai.FunctionDeclaration{
		Name:        TASK_COMPLETE_TOOL_NAME,
		Description: "Call this tool to signal that you have completed your task. This is the ONLY way to finish.",
		Parameters: &genai.Schema{
			Type:       genai.TypeObject,
			Properties: make(map[string]*genai.Schema),
			Required:   []string{},
		},
	}

	if outputConfig != nil {
		completeTool.Description = "Call this tool to submit your final answer and complete the task. This is the ONLY way to finish."
		// For now, we'll use a generic string schema for the output.
		// In a full implementation, this would involve converting outputConfig.Schema (Zod schema) to genai.Schema.
		completeTool.Parameters.Properties[outputConfig.OutputName] = &genai.Schema{Type: genai.TypeString}
		completeTool.Parameters.Required = append(completeTool.Parameters.Required, outputConfig.OutputName)
	}

	toolsList = append(toolsList, completeTool)

	return toolsList, nil
}

// checkTermination checks if the agent should terminate due to exceeding configured limits.
func (ae *AgentExecutor) checkTermination(startTime time.Time, turnCounter int) *AgentTerminateMode {
	runConfig := ae.Definition.RunConfig

	if runConfig.MaxTurns > 0 && turnCounter >= runConfig.MaxTurns {
		mode := AgentTerminateModeMaxTurns
		return &mode
	}

	elapsedMinutes := time.Since(startTime).Minutes()
	if runConfig.MaxTimeMinutes > 0 && elapsedMinutes >= float64(runConfig.MaxTimeMinutes) {
		mode := AgentTerminateModeTimeout
		return &mode
	}

	return nil
}

// callModel calls the generative model with the current context and tools.
func (ae *AgentExecutor) callModel(
	chat *core.GeminiChat,
	message []Part,
	tools []genai.FunctionDeclaration,
	ctx context.Context,
	promptId string,
) ([]FunctionCall, string, error) {
	messageParams := types.MessageParams{
		Message:     message,
		AbortSignal: ctx,
	}
	if len(tools) > 0 {
		messageParams.Tools = tools
	}

	responseStream, err := chat.SendMessageStream(
		ae.Definition.ModelConfig.Model,
		messageParams,
		promptId,
	)
	if err != nil {
		return nil, "", fmt.Errorf("failed to send message stream: %w", err)
	}

	functionCalls := []FunctionCall{}
	textResponse := ""

	for resp := range responseStream {
		if ctx.Err() != nil { // Check for context cancellation
			break
		}

		if resp.Type == types.StreamEventTypeChunk {
			chunk := resp.Value
			if chunk == nil || len(chunk.Candidates) == 0 || chunk.Candidates[0].Content == nil {
				continue
			}

			// Extract and emit any subject "thought" content from the model.
			for _, part := range chunk.Candidates[0].Content.Parts {
				if p, ok := part.(genai.Text); ok && strings.HasPrefix(string(p), "thought") { // Simplified thought detection
					thoughtResult := parseThought(string(p))
					if thoughtResult.Subject != "" {
						ae.emitActivity("THOUGHT_CHUNK", map[string]interface{}{"text": thoughtResult.Subject})
					}
				}
			}

			// Collect any function calls requested by the model.
			for _, part := range chunk.Candidates[0].Content.Parts {
				if fc, ok := part.(genai.FunctionCall); ok {
					// Convert genai.FunctionCall to agents.FunctionCall
					argsMap := make(map[string]interface{})
					for k, v := range fc.Args {
						argsMap[k] = v
					}
					functionCalls = append(functionCalls, FunctionCall{
						ID:   fc.ID,
						Name: fc.Name,
						Args: argsMap,
					})
				}
			}

			// Handle text response (non-thought text)
			for _, part := range chunk.Candidates[0].Content.Parts {
				if p, ok := part.(genai.Text); ok {
					// Check if it's not a thought part (simplified)
					if !strings.HasPrefix(string(p), "thought") {
						textResponse += string(p)
					}
				}
			}
		} else if resp.Type == types.StreamEventTypeError {
			return nil, "", resp.Error
		}
	}

	return functionCalls, textResponse, nil
}

// processFunctionCalls executes function calls requested by the model and returns the results.
func (ae *AgentExecutor) processFunctionCalls(
	functionCalls []FunctionCall,
	ctx context.Context,
	promptId string,
) ([]Part, *string, bool, error) {
	allowedToolNames := make(map[string]bool)
	for _, name := range ae.ToolRegistry.GetAllToolNames() {
		allowedToolNames[name] = true
	}
	// Always allow the completion tool
	allowedToolNames[TASK_COMPLETE_TOOL_NAME] = true

	var submittedOutput *string = nil
	taskCompleted := false

	// We'll collect results from tool executions
	toolResponseParts := []Part{}

	for i, functionCall := range functionCalls {
		callId := functionCall.ID
		if callId == "" {
			callId = fmt.Sprintf("%s-%d", promptId, i)
		}
		args := functionCall.Args

		ae.emitActivity("TOOL_CALL_START", map[string]interface{}{
			"name": functionCall.Name,
			"args": args,
		})

		if functionCall.Name == TASK_COMPLETE_TOOL_NAME {
			if taskCompleted {
				// We already have a completion from this turn. Ignore subsequent ones.
				errorMsg := "Task already marked complete in this turn. Ignoring duplicate call."
				toolResponseParts = append(toolResponseParts, Part{
					FunctionResponse: &FunctionResponse{
						Name:     TASK_COMPLETE_TOOL_NAME,
						Response: map[string]interface{}{"error": errorMsg},
						ID:       callId,
					},
				})
				ae.emitActivity("ERROR", map[string]interface{}{
					"context": "protocol_violation",
					"name":    functionCall.Name,
					"error":   errorMsg,
				})
				continue
			}

			outputConfig := ae.Definition.OutputConfig
			taskCompleted = true // Signal completion regardless of output presence

			if outputConfig != nil {
				outputName := outputConfig.OutputName
				outputValue, ok := args[outputName]
				if ok {
					// TODO: Implement schema validation (equivalent of Zod schema validation)
					// For now, assume validation passes.
					// const validationResult = outputConfig.schema.safeParse(outputValue);
					// if (!validationResult.success) { ... }

					// Simplified output processing
					// if ae.Definition.ProcessOutput != nil { // Assuming ProcessOutput is a field in AgentDefinition
					// 	// submittedOutput = ae.Definition.ProcessOutput(outputValue)
					// } else {
					if strVal, isString := outputValue.(string); isString {
						submittedOutput = &strVal
					} else {
						// Fallback to JSON stringify if not a string
						// jsonBytes, _ := json.MarshalIndent(outputValue, "", "  ")
						// strVal := string(jsonBytes)
						// submittedOutput = &strVal
					}
					// }

					toolResponseParts = append(toolResponseParts, Part{
						FunctionResponse: &FunctionResponse{
							Name:     TASK_COMPLETE_TOOL_NAME,
							Response: map[string]interface{}{"result": "Output submitted and task completed."},
							ID:       callId,
						},
					})
					ae.emitActivity("TOOL_CALL_END", map[string]interface{}{
						"name":   functionCall.Name,
						"output": "Output submitted and task completed.",
					})
				} else {
					// Failed to provide required output.
					taskCompleted = false // Revoke completion status
					errorMsg := fmt.Sprintf("Missing required argument '%s' for completion.", outputName)
					toolResponseParts = append(toolResponseParts, Part{
						FunctionResponse: &FunctionResponse{
							Name:     TASK_COMPLETE_TOOL_NAME,
							Response: map[string]interface{}{"error": errorMsg},
							ID:       callId,
						},
					})
					ae.emitActivity("ERROR", map[string]interface{}{
						"context": "tool_call",
						"name":    functionCall.Name,
						"error":   errorMsg,
					})
				}
			} else {
				// No output expected. Just signal completion.
				temp := "Task completed successfully."
				submittedOutput = &temp
				toolResponseParts = append(toolResponseParts, Part{
					FunctionResponse: &FunctionResponse{
						Name:     TASK_COMPLETE_TOOL_NAME,
						Response: map[string]interface{}{"status": "Task marked complete."},
						ID:       callId,
					},
				})
				ae.emitActivity("TOOL_CALL_END", map[string]interface{}{
					"name":   functionCall.Name,
					"output": "Task marked complete.",
				})
			}
			continue
		}

		// Handle standard tools
		if !allowedToolNames[functionCall.Name] {
			errorMsg := fmt.Sprintf("Unauthorized tool call: '%s' is not available to this agent.", functionCall.Name)
			// debugLogger.warn(fmt.Sprintf("[AgentExecutor] Blocked call: %s", errorMsg)) // TODO: Implement debugLogger

			toolResponseParts = append(toolResponseParts, Part{
				FunctionResponse: &FunctionResponse{
					Name:     functionCall.Name,
					ID:       callId,
					Response: map[string]interface{}{"error": errorMsg},
				},
			})

			ae.emitActivity("ERROR", map[string]interface{}{
				"context": "tool_call_unauthorized",
				"name":    functionCall.Name,
				"callId":  callId,
				"error":   errorMsg,
			})
			continue
		}

		requestInfo := ToolCallRequestInfo{
			CallID:            callId,
			Name:              functionCall.Name,
			Args:              args,
			IsClientInitiated: true,
			PromptID:          promptId,
		}

		// Execute the tool call
		completedCall, err := ExecuteToolCall(ae.RuntimeContext, requestInfo, ctx)
		if err != nil {
			errorMsg := fmt.Sprintf("tool execution failed: %v", err)
			toolResponseParts = append(toolResponseParts, Part{
				FunctionResponse: &FunctionResponse{
					Name:     functionCall.Name,
					ID:       callId,
					Response: map[string]interface{}{"error": errorMsg},
				},
			})
			ae.emitActivity("ERROR", map[string]interface{}{
				"context": "tool_call",
				"name":    functionCall.Name,
				"error":   errorMsg,
			})
			continue
		}

		if completedCall.GetResponse().Error != nil {
			ae.emitActivity("ERROR", map[string]interface{}{
				"context": "tool_call",
				"name":    functionCall.Name,
				"error":   completedCall.GetResponse().Error.Error(),
			})
		} else {
			ae.emitActivity("TOOL_CALL_END", map[string]interface{}{
				"name":   functionCall.Name,
				"output": completedCall.GetResponse().ResultDisplay, // Assuming ResultDisplay is a string
			})
		}
		toolResponseParts = append(toolResponseParts, completedCall.GetResponse().ResponseParts...)
	}

	// If all authorized tool calls failed (and task isn't complete), provide a generic error.
	if len(functionCalls) > 0 && len(toolResponseParts) == 0 && !taskCompleted {
		toolResponseParts = append(toolResponseParts, Part{
			Text: "All tool calls failed or were unauthorized. Please analyze the errors and try an alternative approach.",
		})
	}

	return []Part{{Text: "user", FunctionCall: nil, FunctionResponse: nil, InlineData: nil, FileData: nil, Thought: ""}}, submittedOutput, taskCompleted, nil // Simplified nextMessage
}

// Placeholder for emitActivity
func (ae *AgentExecutor) emitActivity(activityType string, data map[string]interface{}) {
	if ae.OnActivity != nil {
		event := SubagentActivityEvent{
			IsSubagentActivityEvent: true,
			AgentName:               ae.Definition.Name,
			Type:                    activityType,
			Data:                    data,
		}
		ae.OnActivity(event)
	}
}

// buildSystemPrompt builds the system prompt from the agent definition and inputs.
func (ae *AgentExecutor) buildSystemPrompt(inputs AgentInputs) (string, error) {
	promptConfig := ae.Definition.PromptConfig
	if promptConfig.SystemPrompt == "" {
		return "", nil
	}

	// Inject user inputs into the prompt template.
	finalPrompt := templateString(promptConfig.SystemPrompt, inputs)

	// Append environment context (CWD and folder structure).
	dirContext, err := getDirectoryContextString(ae.RuntimeContext)
	if err != nil {
		return "", fmt.Errorf("failed to get directory context string: %w", err)
	}
	finalPrompt += fmt.Sprintf("\n\n# Environment Context\n%s", dirContext)

	// Append standard rules for non-interactive execution.
	finalPrompt += "\nImportant Rules:\n* You are running in a non-interactive mode. You CANNOT ask the user for input or clarification.\n* Work systematically using available tools to complete your task.\n* Always use absolute paths for file operations. Construct them using the provided \"Environment Context\"."

	finalPrompt += fmt.Sprintf("\n* When you have completed your task, you MUST call the `%s` tool.\n* Do not call any other tools in the same turn as `%s`.\n* This is the ONLY way to complete your mission. If you stop calling tools without calling this, you have failed.", TASK_COMPLETE_TOOL_NAME, TASK_COMPLETE_TOOL_NAME)

	return finalPrompt, nil
}

// CreateAgentExecutor creates and validates a new AgentExecutor instance.
func CreateAgentExecutor(definition AgentDefinition, runtimeContext *config.Config, parentToolRegistry *tools.ToolRegistry, parentPromptId string, onActivity ActivityCallback) (*AgentExecutor, error) {
	// Create an isolated tool registry for this agent instance.
	agentToolRegistry := tools.NewToolRegistry()

	// Register tools specified in the agent's definition.
	if definition.ToolConfig != nil {
		for _, toolName := range definition.ToolConfig.Tools {
			tool := parentToolRegistry.GetTool(toolName)
			if tool == nil {
				return nil, fmt.Errorf("tool '%s' not found in parent registry", toolName)
			}
			if err := agentToolRegistry.Register(tool); err != nil {
				return nil, fmt.Errorf("failed to register tool %s: %w", toolName, err)
			}
		}

		// Validate that all registered tools are safe for non-interactive execution.
		if err := validateTools(agentToolRegistry, definition.Name); err != nil {
			return nil, err
		}
	}

	var parentPrefix string
	if parentPromptId != "" {
		parentPrefix = fmt.Sprintf("%s-", parentPromptId)
	}
	randomIDPart := fmt.Sprintf("%x", time.Now().UnixNano())
	agentID := fmt.Sprintf("%s%s-%s", parentPrefix, definition.Name, randomIDPart)

	return &AgentExecutor{
		Definition:     definition,
		AgentID:        agentID,
		ToolRegistry:   agentToolRegistry,
		RuntimeContext: runtimeContext,
		OnActivity:     onActivity,
		parentPromptId: parentPromptId,
	}, nil
}

// validateTools validates that all tools in a registry are safe for non-interactive use.
func validateTools(toolRegistry *tools.ToolRegistry, agentName string) error {
	// Tools that are non-interactive.
	allowlist := map[string]bool{
		tools.LS_TOOL_NAME:              true,
		tools.READ_FILE_TOOL_NAME:       true,
		tools.GREP_TOOL_NAME:            true,
		tools.GLOB_TOOL_NAME:            true,
		tools.READ_MANY_FILES_TOOL_NAME: true,
		tools.MEMORY_TOOL_NAME:          true,
		tools.WEB_SEARCH_TOOL_NAME:      true,
		tools.WEB_FETCH_TOOL_NAME:       true,
	}

	for _, tool := range toolRegistry.GetAllRegisteredTools() {
		if _, ok := allowlist[tool.Name()]; !ok {
			return fmt.Errorf("tool \"%s\" is not on the allow-list for non-interactive execution in agent \"%s\". Only tools that do not require user confirmation can be used in subagents.", tool.Name(), agentName)
		}
	}
	return nil
}
