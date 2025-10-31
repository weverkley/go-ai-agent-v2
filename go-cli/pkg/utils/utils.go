package utils

import (
	"fmt"
	"strings"

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
func LogAgentStart(runtimeContext interface{}, event types.AgentStartEvent) {
	// TODO: Implement actual telemetry logging
	fmt.Printf("Telemetry: Agent %s started (placeholder). AgentID: %s\n", event.AgentName, event.AgentID)
}

// LogAgentFinish logs the finish of an agent's execution.
func LogAgentFinish(runtimeContext interface{}, event types.AgentFinishEvent) {
	// TODO: Implement actual telemetry logging
	fmt.Printf("Telemetry: Agent %s finished (placeholder). AgentID: %s, Duration: %dms, Turns: %d, Reason: %s\n", event.AgentName, event.AgentID, event.DurationMs, event.TurnCounter, event.TerminateReason)
}
