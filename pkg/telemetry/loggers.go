package telemetry

import (
	"go-ai-agent-v2/go-cli/pkg/types"
)

var GlobalLogger TelemetryLogger = &noopTelemetryLogger{}

// InitGlobalLogger initializes the global telemetry logger.
func InitGlobalLogger(settings *types.TelemetrySettings, runMode string) {
	GlobalLogger = NewTelemetryLogger(settings, runMode)
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

// LogErrorf logs an error message.
func LogErrorf(format string, args ...interface{}) {
	GlobalLogger.LogErrorf(format, args...)
}

// LogWarnf logs a warning message.
func LogWarnf(format string, args ...interface{}) {
	GlobalLogger.LogWarnf(format, args...)
}
