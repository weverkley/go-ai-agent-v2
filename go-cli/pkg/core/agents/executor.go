package agents

import (
	"context"
	"fmt"
	"time"

	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/tools"
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
				AgentID:       ae.AgentID,
				AgentName:     ae.Definition.Name,
				DurationMs:    time.Since(startTime).Milliseconds(),
				TurnCounter:   turnCounter,
				TerminateReason: terminateReason,
			},
		)
	}()

	// Placeholder for chat object
	// chat, err := ae.createChatObject(inputs)
	// if err != nil {
	// 	ae.emitActivity("ERROR", map[string]interface{}{"error": err.Error()})
	// 	return OutputObject{}, err
	// }
	// Placeholder for tools
	// tools := ae.prepareToolsList()

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

		// Placeholder for callModel
		// functionCalls, err := ae.callModel(chat, currentMessage, tools, ctx, promptId)
		// if err != nil {
		// 	ae.emitActivity("ERROR", map[string]interface{}{"error": err.Error()})
		// 	return OutputObject{}, err
		// }
		// if ctx.Err() != nil {
		// 	terminateReason = AgentTerminateModeAborted
		// 	break
		// }

		// Placeholder for processFunctionCalls
		// nextMessage, submittedOutput, taskCompleted, err := ae.processFunctionCalls(functionCalls, ctx, promptId)
		// if err != nil {
		// 	ae.emitActivity("ERROR", map[string]interface{}{"error": err.Error()})
		// 	return OutputObject{}, err
		// }

		// if taskCompleted {
		// 	if submittedOutput != nil {
		// 		finalResult = submittedOutput
		// 	} else {
		// 		temp := "Task completed successfully."
		// 		finalResult = &temp
		// 	}
		// 	terminateReason = AgentTerminateModeGoal
		// 	break
		// }

		// currentMessage = nextMessage
		break // Temporary break to avoid infinite loop with placeholders
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

	generationConfig := agents.GenerateContentConfig{
		Temperature: modelConfig.Temperature,
		TopP:        modelConfig.TopP,
		ThinkingConfig: &agents.ThinkingConfig{
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
	initialMessages []agents.Part,
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

// Placeholder for checkTermination
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

// Placeholder for callModel
func (ae *AgentExecutor) callModel(chat interface{}, currentMessage []Part, tools interface{}, ctx context.Context, promptId string) ([]interface{}, error) {
	return nil, nil
}

// Placeholder for processFunctionCalls
func (ae *AgentExecutor) processFunctionCalls(functionCalls []interface{}, ctx context.Context, promptId string) ([]Part, *string, bool, error) {
	return nil, nil, false, nil
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
