package telemetry

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"go-ai-agent-v2/go-cli/pkg/types"
)

// TelemetryLogger defines the interface for telemetry logging.
type TelemetryLogger interface {
	LogAgentStart(event types.AgentStartEvent)
	LogAgentFinish(event types.AgentFinishEvent)
	LogErrorf(format string, args ...interface{})
}

// noopTelemetryLogger is a no-operation implementation of TelemetryLogger.
type noopTelemetryLogger struct{}

func (l *noopTelemetryLogger) LogAgentStart(event types.AgentStartEvent) {}
func (l *noopTelemetryLogger) LogAgentFinish(event types.AgentFinishEvent) {}
func (l *noopTelemetryLogger) LogErrorf(format string, args ...interface{}) {}

// fileTelemetryLogger logs telemetry events to a specified file.
type fileTelemetryLogger struct {
	mu       sync.Mutex
	filePath string
	enabled  bool
}

// NewFileTelemetryLogger creates a new fileTelemetryLogger.
func NewFileTelemetryLogger(filePath string, enabled bool) *fileTelemetryLogger {
	return &fileTelemetryLogger{
		filePath: filePath,
		enabled:  enabled,
	}
}

func (l *fileTelemetryLogger) LogAgentStart(event types.AgentStartEvent) {
	if !l.enabled {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	logEntry := map[string]interface{}{
		"type":    "AgentStart",
		"agentID": event.AgentID,
		"agentName": event.AgentName,
	}
	data, err := json.Marshal(logEntry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling AgentStart event: %v\n", err)
		return
	}

	l.writeLog(data)
}

func (l *fileTelemetryLogger) LogAgentFinish(event types.AgentFinishEvent) {
	if !l.enabled {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	logEntry := map[string]interface{}{
		"type":          "AgentFinish",
		"agentID":       event.AgentID,
		"agentName":     event.AgentName,
		"durationMs":    event.DurationMs,
		"turnCounter":   event.TurnCounter,
		"terminateReason": event.TerminateReason,
	}
	data, err := json.Marshal(logEntry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling AgentFinish event: %v\n", err)
		return
	}

	l.writeLog(data)
}

func (l *fileTelemetryLogger) writeLog(data []byte) {
	file, err := os.OpenFile(l.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening telemetry log file %s: %v\n", l.filePath, err)
		return
	}
	defer file.Close()

	if _, err := file.Write(data); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to telemetry log file %s: %v\n", l.filePath, err)
		return
	}
	if _, err := file.WriteString("\n"); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing newline to telemetry log file %s: %v\n", l.filePath, err)
		return
	}
}

func (l *fileTelemetryLogger) LogErrorf(format string, args ...interface{}) {
	if !l.enabled {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	logEntry := map[string]interface{}{
		"type":    "Error",
		"message": fmt.Sprintf(format, args...),
		"timestamp": time.Now().Format(time.RFC3339),
	}
	data, err := json.Marshal(logEntry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling error log entry: %v\n", err)
		return
	}
	l.writeLog(data)
}

// NewTelemetryLogger creates a TelemetryLogger based on the provided telemetry settings.
func NewTelemetryLogger(settings *types.TelemetrySettings) TelemetryLogger {
	if settings == nil || !settings.Enabled {
		return &noopTelemetryLogger{}
	}

	if settings.Outfile != "" {
		return NewFileTelemetryLogger(settings.Outfile, true)
	}

	// Default to no-op logger if no specific logger is configured
	return &noopTelemetryLogger{}
}
