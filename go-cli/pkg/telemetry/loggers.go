package telemetry

import (
	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/types"
)

var globalLogger TelemetryLogger = &noopTelemetryLogger{}

// LogAgentStart logs the start of an agent's execution.
func LogAgentStart(event types.AgentStartEvent) {
	globalLogger.LogAgentStart(event)
}

// LogAgentFinish logs the finish of an agent's execution.
func LogAgentFinish(event types.AgentFinishEvent) {
	globalLogger.LogAgentFinish(event)
}