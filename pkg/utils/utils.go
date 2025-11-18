package utils

import (
	"fmt"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/telemetry" // Import telemetry package
	"go-ai-agent-v2/go-cli/pkg/types"
)

// TemplateString replaces placeholders in a string with values from inputs.
func TemplateString(template string, inputs map[string]interface{}) string {
	result := template
	for key, value := range inputs {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}
	return result
}

// LogAgentStart logs the start of an agent's execution.
func LogAgentStart(event types.AgentStartEvent) {
	telemetry.GlobalLogger.LogAgentStart(event)
}

// LogAgentFinish logs the finish of an agent's execution.
func LogAgentFinish(event types.AgentFinishEvent) {
	telemetry.GlobalLogger.LogAgentFinish(event)
}
