package agents

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/core"
	"go-ai-agent-v2/go-cli/pkg/types"
	"go-ai-agent-v2/go-cli/pkg/utils"

	"github.com/google/generative-ai-go/genai"
)

// AgentExecutor executes an agent loop based on an AgentDefinition.
type AgentExecutor struct {
	Definition     AgentDefinition
	AgentID        string
	ToolRegistry   *types.ToolRegistry
	RuntimeContext *config.Config
	OnActivity     ActivityCallback
	parentPromptId string
}

// Run executes the agent loop.
func (ae *AgentExecutor) Run(inputs AgentInputs, ctx context.Context) (OutputObject, error) {
	startTime := time.Now()
	turnCounter := 0
	var terminateReason types.AgentTerminateMode
	var finalResult string

	utils.LogAgentStart(
		ae.RuntimeContext,
		types.AgentStartEvent{AgentID: ae.AgentID, AgentName: ae.Definition.Name},
	)

	chat, err := ae.createChatObject(inputs)
	if err != nil {
		ae.emitActivity("ERROR", map[string]interface{}{"error": err.Error()})
		return OutputObject{}, err
	}

toolsList, err := ae.prepareToolsList()
	if err != nil {
		ae.emitActivity("ERROR", map[string]interface{}{"error": err.Error()})
		return OutputObject{}, err
	}

	query := "Get Started!"
	if ae.Definition.PromptConfig.Query != "" {
		query = utils.TemplateString(ae.Definition.PromptConfig.Query, inputs)
	}
	currentMessage := &genai.Content{Parts: []genai.Part{genai.Text(query)}, Role: "user"}

MainLoop:
	for {
		reason := ae.checkTermination(startTime, turnCounter)
		if reason != nil {
			terminateReason = *reason
			break
		}
		if ctx.Err() != nil {
			terminateReason = types.AgentTerminateModeAborted
			break
		}

		promptId := fmt.Sprintf("%s#%d", ae.AgentID, turnCounter)
		turnCounter++

		functionCalls, _, err := ae.callModel(chat, currentMessage, toolsList, ctx, promptId)
		if err != nil {
			ae.emitActivity("ERROR", map[string]interface{}{"error": err.Error()})
			return OutputObject{}, err
		}

		if ctx.Err() != nil {
			terminateReason = types.AgentTerminateModeAborted
			break
		}

		if len(functionCalls) == 0 {
			terminateReason = types.AgentTerminateModeError
			finalResult = fmt.Sprintf("Agent stopped calling tools but did not call '%s' to finalize the session.", types.TASK_COMPLETE_TOOL_NAME)
			ae.emitActivity("ERROR", map[string]interface{}{
				"error":   finalResult,
				"context": "protocol_violation",
			})
			break
		}

		nextMessage, submittedOutput, taskCompleted, err := ae.processFunctionCalls(functionCalls, ctx, promptId)
		if err != nil {
			ae.emitActivity("ERROR", map[string]interface{}{"error": err.Error()})
			// In the JS version, some errors might not terminate the loop.
			// For now, we terminate on any error from processFunctionCalls.
			return OutputObject{}, err
		}

		if taskCompleted {
			if submittedOutput != "" {
				finalResult = submittedOutput
			} else {
				finalResult = "Task completed successfully."
			}
			terminateReason = types.AgentTerminateModeGoal
			break MainLoop
		}

		currentMessage = nextMessage
	}

	utils.LogAgentFinish(
		ae.RuntimeContext,
		types.AgentFinishEvent{
			AgentID:         ae.AgentID,
			AgentName:       ae.Definition.Name,
			DurationMs:      time.Since(startTime).Milliseconds(),
			TurnCounter:     turnCounter,
			TerminateReason: terminateReason,
		},
	)

	if terminateReason == types.AgentTerminateModeGoal {
		return OutputObject{Result: finalResult, TerminateReason: terminateReason}, nil
	}

	result := "Agent execution was terminated before completion."
	if finalResult != "" {
		result = finalResult
	}
	return OutputObject{Result: result, TerminateReason: terminateReason}, nil
}

// createChatObject initializes a GeminiChat instance for the agent run.
func (ae *AgentExecutor) createChatObject(inputs AgentInputs) (core.Executor, error) {
	promptConfig := ae.Definition.PromptConfig
	modelConfig := ae.Definition.ModelConfig

	if promptConfig.SystemPrompt == "" && len(promptConfig.InitialMessages) == 0 {
		return nil, fmt.Errorf("PromptConfig must define either `systemPrompt` or `initialMessages`")
	}

	startHistory := ae.applyTemplateToInitialMessages(promptConfig.InitialMessages, inputs)

	var systemInstruction *genai.Content
	if promptConfig.SystemPrompt != "" {
		instruction, err := ae.buildSystemPrompt(inputs)
		if err != nil {
			return nil, fmt.Errorf("failed to build system prompt: %w", err)
		}
		systemInstruction = &genai.Content{Parts: []genai.Part{genai.Text(instruction)}}
	}

	generationConfig := types.GenerateContentConfig{
		Temperature: modelConfig.Temperature,
		TopP:        modelConfig.TopP,
		ThinkingConfig: &types.ThinkingConfig{
			IncludeThoughts: true,
			ThinkingBudget:  modelConfig.ThinkingBudget,
		},
	}

	if systemInstruction != nil {
		// This is a simplified way to handle system instructions.
		// The JS version seems to handle it as part of GenerateContentConfig.
		// For now, we prepend it to the history.
		startHistory = append([]*genai.Content{systemInstruction}, startHistory...)
	}

	chat, err := core.NewGeminiChat(
		ae.RuntimeContext,
		generationConfig,
		startHistory,
	)
	if err != nil {
		utils.LogErrorf(ae.RuntimeContext, "Failed to create chat object: %v", err)
		return nil, fmt.Errorf("failed to create chat object: %w", err)
	}

	return chat, nil
}

// applyTemplateToInitialMessages applies template strings to initial messages.
func (ae *AgentExecutor) applyTemplateToInitialMessages(
	initialMessages []types.Part,
	inputs AgentInputs,
) []*genai.Content {
	templatedMessages := make([]*genai.Content, len(initialMessages))
	for i, part := range initialMessages {
		var newGenaiParts []genai.Part
		if part.Text != "" {
			newGenaiParts = append(newGenaiParts, genai.Text(utils.TemplateString(part.Text, inputs)))
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
		templatedMessages[i] = &genai.Content{Parts: newGenaiParts}
	}
	return templatedMessages
}

// buildSystemPrompt builds the system prompt from the agent definition and inputs.
func (ae *AgentExecutor) buildSystemPrompt(inputs AgentInputs) (string, error) {
	promptConfig := ae.Definition.PromptConfig
	if promptConfig.SystemPrompt == "" {
		return "", nil
	}

	finalPrompt := utils.TemplateString(promptConfig.SystemPrompt, inputs)

	dirContext, err := utils.GetDirectoryContextString(ae.RuntimeContext)
	if err != nil {
		return "", fmt.Errorf("failed to get directory context string: %w", err)
	}
	finalPrompt += fmt.Sprintf("\n\n# Environment Context\n%s", dirContext)

	finalPrompt += `
Important Rules:
* You are running in a non-interactive mode. You CANNOT ask the user for input or clarification.
* Work systematically using available tools to complete your task.
* Always use absolute paths for file operations. Construct them using the provided "Environment Context".`

	finalPrompt += fmt.Sprintf("\n* When you have completed your task, you MUST call the `%s` tool.\n* Do not call any other tools in the same turn as `%s`.\n* This is the ONLY way to complete your mission. If you stop calling tools without calling this, you have failed.", types.TASK_COMPLETE_TOOL_NAME, types.TASK_COMPLETE_TOOL_NAME)

	return finalPrompt, nil
}

// prepareToolsList prepares the list of tool function declarations to be sent to the model.
func (ae *AgentExecutor) prepareToolsList() ([]*genai.FunctionDeclaration, error) {
	toolsList := []*genai.FunctionDeclaration{}
	toolConfig := ae.Definition.ToolConfig
	outputConfig := ae.Definition.OutputConfig

	if toolConfig != nil {
		toolNamesToLoad := toolConfig.Tools
		// Add schemas from tools that were registered by name.
		declarations := ae.ToolRegistry.GetFunctionDeclarationsFiltered(toolNamesToLoad)
		for i := range declarations {
			toolsList = append(toolsList, &declarations[i])
		}
	}

	// Always inject complete_task.
	completeTool := &genai.FunctionDeclaration{
		Name:        types.TASK_COMPLETE_TOOL_NAME,
		Description: "Call this tool to signal that you have completed your task. This is the ONLY way to finish.",
		Parameters: &genai.Schema{
			Type:       genai.TypeObject,
			Properties: make(map[string]*genai.Schema),
			Required:   []string{},
		},
	}

	if outputConfig != nil {
		completeTool.Description = "Call this tool to submit your final answer and complete the task. This is the ONLY way to finish."
		// This is a simplified schema generation. A proper implementation would
		// convert the Zod-like schema from the definition to a genai.Schema.
		completeTool.Parameters.Properties[outputConfig.OutputName] = &genai.Schema{Type: genai.TypeString} // Assuming string output for now
		completeTool.Parameters.Required = append(completeTool.Parameters.Required, outputConfig.OutputName)
	}

	toolsList = append(toolsList, completeTool)

	return toolsList, nil
}

// checkTermination checks if the agent should terminate due to exceeding configured limits.
func (ae *AgentExecutor) checkTermination(startTime time.Time, turnCounter int) *types.AgentTerminateMode {
	runConfig := ae.Definition.RunConfig

	if runConfig.MaxTurns > 0 && turnCounter >= runConfig.MaxTurns {
		mode := types.AgentTerminateModeMaxTurns
		return &mode
	}

	elapsedMinutes := time.Since(startTime).Minutes()
	if runConfig.MaxTimeMinutes > 0 && elapsedMinutes >= float64(runConfig.MaxTimeMinutes) {
		mode := types.AgentTerminateModeTimeout
		return &mode
	}

	return nil
}

// callModel calls the generative model with the current context and tools.
func (ae *AgentExecutor) callModel(
	chat core.Executor,
	message *genai.Content,
	tools []*genai.FunctionDeclaration,
	ctx context.Context,
	promptId string,
) ([]*genai.FunctionCall, string, error) {
	var parts []types.Part
	for _, p := range message.Parts {
		if text, ok := p.(genai.Text); ok {
			parts = append(parts, types.Part{Text: string(text)})
		} else if fc, ok := p.(genai.FunctionCall); ok {
			argsMap := make(map[string]interface{})
			for k, v := range fc.Args {
				argsMap[k] = v
			}
			parts = append(parts, types.Part{FunctionCall: &types.FunctionCall{Name: fc.Name, Args: argsMap}})
		}
	}

	messageParams := types.MessageParams{
		Message:     parts,
		AbortSignal: ctx,
	}
	if len(tools) > 0 {
		messageParams.Tools = []*genai.Tool{{FunctionDeclarations: tools}}
	}

	responseStream, err := chat.SendMessageStream(
		ae.Definition.ModelConfig.Model,
		messageParams,
		promptId,
	)
	if err != nil {
		return nil, "", fmt.Errorf("failed to send message stream: %w", err)
	}

	var functionCalls []*genai.FunctionCall
	var textResponse strings.Builder

	for resp := range responseStream {
		if ctx.Err() != nil {
			break
		}

		if resp.Type == types.StreamEventTypeChunk {
			chunk := resp.Value
			if chunk == nil || len(chunk.Candidates) == 0 || chunk.Candidates[0].Content == nil {
				continue
			}

			for _, part := range chunk.Candidates[0].Content.Parts {
				if thought, ok := part.(genai.Text); ok {
					thoughtResult := utils.ParseThought(string(thought))
					if thoughtResult.Subject != "" {
						ae.emitActivity("THOUGHT_CHUNK", map[string]interface{}{"text": thoughtResult.Subject})
					}
				}

				if fcPart, ok := part.(genai.FunctionCall); ok {
					functionCalls = append(functionCalls, &fcPart)
				}

				if txt, ok := part.(genai.Text); ok {
					if !strings.HasPrefix(string(txt), "**") { // Simple check to filter out thoughts
						textResponse.WriteString(string(txt))
					}
				}
			}
		} else if resp.Type == types.StreamEventTypeError {
			return nil, "", resp.Error
		}
	}

	return functionCalls, textResponse.String(), nil
}

// processFunctionCalls executes function calls requested by the model and returns the results.
func (ae *AgentExecutor) processFunctionCalls(
	functionCalls []*genai.FunctionCall,
	ctx context.Context,
	promptId string,
) (*genai.Content, string, bool, error) {
	allowedToolNames := make(map[string]bool)
	for _, name := range ae.ToolRegistry.GetAllToolNames() {
		allowedToolNames[name] = true
	}
	allowedToolNames[types.TASK_COMPLETE_TOOL_NAME] = true

	var submittedOutput string
	taskCompleted := false

	var wg sync.WaitGroup
	toolResponseChan := make(chan []genai.Part, len(functionCalls))
	syncResponseParts := make([][]genai.Part, 0)

	for i, functionCall := range functionCalls {
		callId := fmt.Sprintf("%s-%d", promptId, i) // Reintroduce callId
		args := functionCall.Args

		ae.emitActivity("TOOL_CALL_START", map[string]interface{}{
			"name": functionCall.Name,
			"args": args,
		})

		if functionCall.Name == types.TASK_COMPLETE_TOOL_NAME {
			if taskCompleted {
				errorMsg := "Task already marked complete in this turn. Ignoring duplicate call."
				syncResponseParts = append(syncResponseParts, []genai.Part{genai.FunctionResponse{
						Name:     types.TASK_COMPLETE_TOOL_NAME,
						Response: map[string]interface{}{"error": errorMsg},
					}})
				ae.emitActivity("ERROR", map[string]interface{}{
					"context": "protocol_violation",
					"name":    functionCall.Name,
					"error":   errorMsg,
				})
				continue
			}

			outputConfig := ae.Definition.OutputConfig
			taskCompleted = true // Signal completion

			if outputConfig != nil {
				outputName := outputConfig.OutputName
				if outputValue, ok := args[outputName]; ok {
					// Simplified validation and processing
					validatedOutput := fmt.Sprintf("%v", outputValue)
					if ae.Definition.ProcessOutput != nil {
						submittedOutput = ae.Definition.ProcessOutput(validatedOutput)
					} else {
						submittedOutput = validatedOutput
					}
					syncResponseParts = append(syncResponseParts, []genai.Part{genai.FunctionResponse{
							Name:     types.TASK_COMPLETE_TOOL_NAME,
							Response: map[string]interface{}{"result": "Output submitted and task completed."},
						}})
					ae.emitActivity("TOOL_CALL_END", map[string]interface{}{
						"name":   functionCall.Name,
						"output": "Output submitted and task completed.",
					})
				} else {
					taskCompleted = false // Revoke completion
					errorMsg := fmt.Sprintf("Missing required argument '%s' for completion.", outputName)
					syncResponseParts = append(syncResponseParts, []genai.Part{genai.FunctionResponse{
							Name:     types.TASK_COMPLETE_TOOL_NAME,
							Response: map[string]interface{}{"error": errorMsg},
						}})
					ae.emitActivity("ERROR", map[string]interface{}{
						"context": "tool_call",
						"name":    functionCall.Name,
						"error":   errorMsg,
					})
				}
			} else {
				submittedOutput = "Task completed successfully."
				syncResponseParts = append(syncResponseParts, []genai.Part{genai.FunctionResponse{
						Name:     types.TASK_COMPLETE_TOOL_NAME,
						Response: map[string]interface{}{"status": "Task marked complete."},
					}})

				ae.emitActivity("TOOL_CALL_END", map[string]interface{}{
					"name":   functionCall.Name,
					"output": "Task marked complete.",
				})
			}
			continue
		}

		if !allowedToolNames[functionCall.Name] {
			errorMsg := fmt.Sprintf("Unauthorized tool call: '%s' is not available to this agent.", functionCall.Name)
			syncResponseParts = append(syncResponseParts, []genai.Part{genai.FunctionResponse{
					Name:     functionCall.Name,
					Response: map[string]interface{}{"error": errorMsg},
				}})
			ae.emitActivity("ERROR", map[string]interface{}{
				"context": "tool_call_unauthorized",
				"name":    functionCall.Name,
				"callId":  callId,
				"error":   errorMsg,
			})
			continue
		}

		wg.Add(1)
		go func(fc *genai.FunctionCall) {
			defer wg.Done()

			tool, err := ae.ToolRegistry.GetTool(fc.Name)
			if err != nil {
				// Handle error: tool not found
				return
			}

			result, err := tool.Execute(fc.Args)
			if err != nil {
				// Handle tool execution error
				return
			}

			if result.Error != nil {
				ae.emitActivity("ERROR", map[string]interface{}{
					"context": "tool_call",
					"name":    fc.Name,
					"error":   result.Error.Message,
				})
			} else {
				ae.emitActivity("TOOL_CALL_END", map[string]interface{}{
					"name":   fc.Name,
					"output": result.ReturnDisplay,
				})
			}

			// Assuming result.LLMContent is []genai.Part
			if parts, ok := result.LLMContent.([]genai.Part); ok {
				toolResponseChan <- parts
			}
		}(functionCall)
	}

	go func() {
		wg.Wait()
		close(toolResponseChan)
	}()

	var asyncResponseParts [][]genai.Part
	for parts := range toolResponseChan {
		asyncResponseParts = append(asyncResponseParts, parts)
	}

	toolResponseParts := make([]genai.Part, 0)
	for _, p := range syncResponseParts {
		toolResponseParts = append(toolResponseParts, p...)
	}
	for _, p := range asyncResponseParts {
		toolResponseParts = append(toolResponseParts, p...)
	}

	if len(functionCalls) > 0 && len(toolResponseParts) == 0 && !taskCompleted {
		toolResponseParts = append(toolResponseParts, genai.Text("All tool calls failed or were unauthorized. Please analyze the errors and try an alternative approach."))
	}

	return &genai.Content{Parts: toolResponseParts, Role: "user"}, submittedOutput, taskCompleted, nil
}

// emitActivity emits an activity event to the configured callback.
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

// CreateAgentExecutor creates and validates a new AgentExecutor instance.
func CreateAgentExecutor(definition AgentDefinition, runtimeContext *config.Config, parentToolRegistry *types.ToolRegistry, parentPromptId string, onActivity ActivityCallback) (*AgentExecutor, error) {
	agentToolRegistry := types.NewToolRegistry()

	if definition.ToolConfig != nil {
		for _, toolName := range definition.ToolConfig.Tools {
			tool, err := parentToolRegistry.GetTool(toolName)
			if err != nil {
				return nil, fmt.Errorf("tool '%s' not found in parent registry", toolName)
			}
			if err := agentToolRegistry.Register(tool); err != nil {
				return nil, fmt.Errorf("failed to register tool %s: %w", toolName, err)
			}
		}

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
func validateTools(toolRegistry *types.ToolRegistry, agentName string) error {
	allowlist := map[string]bool{
		types.LS_TOOL_NAME:              true,
		types.READ_FILE_TOOL_NAME:       true,
		types.GREP_TOOL_NAME:            true,
		types.GLOB_TOOL_NAME:            true,
		types.READ_MANY_FILES_TOOL_NAME: true,
		types.MEMORY_TOOL_NAME:          true,
		types.WEB_SEARCH_TOOL_NAME:      true,
		types.WEB_FETCH_TOOL_NAME:       true,
	}

	for _, tool := range toolRegistry.GetAllRegisteredTools() {
		if !allowlist[tool.Name()] {
			return fmt.Errorf("tool \"%s\" is not on the allow-list for non-interactive execution in agent \"%s\". Only tools that do not require user confirmation can be used in subagents.", tool.Name(), agentName)
		}
	}
	return nil
}