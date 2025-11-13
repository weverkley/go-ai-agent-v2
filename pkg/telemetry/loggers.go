package telemetry

import (
	"go-ai-agent-v2/go-cli/pkg/types"
)

var GlobalLogger TelemetryLogger = &noopTelemetryLogger{}

// InitGlobalLogger initializes the global telemetry logger.
func InitGlobalLogger(settings *types.TelemetrySettings) {
	GlobalLogger = NewTelemetryLogger(settings)
}

// LogAgentStart logs the start of an agent's execution.
func LogAgentStart(event types.AgentStartEvent) {
	GlobalLogger.LogAgentStart(event)
}

// LogAgentFinish logs the finish of an agent's execution.
func LogAgentFinish(event types.AgentFinishEvent) {
	GlobalLogger.LogAgentFinish(event)
}

// LogDebugf logs a debug message.
func LogDebugf(format string, args ...interface{}) {
	GlobalLogger.LogDebugf(format, args...)
}