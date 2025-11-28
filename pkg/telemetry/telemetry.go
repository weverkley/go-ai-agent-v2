package telemetry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath" // Import filepath package
	"sync"
	"time"

	"go-ai-agent-v2/go-cli/pkg/pathutils"
	"go-ai-agent-v2/go-cli/pkg/types"
)

// TelemetryLogger defines the interface for telemetry logging.
type TelemetryLogger interface {
	LogAgentStart(event types.AgentStartEvent)
	LogAgentFinish(event types.AgentFinishEvent)
	LogErrorf(format string, args ...interface{})
	LogWarnf(format string, args ...interface{})
	LogInfof(format string, args ...interface{})
	LogDebugf(format string, args ...interface{})
	LogPrompt(prompt string)
	LogSubagentActivity(event types.SubagentActivityEvent) // New method
}

// noopTelemetryLogger is a no-operation implementation of TelemetryLogger.
type noopTelemetryLogger struct{}

func (l *noopTelemetryLogger) LogAgentStart(event types.AgentStartEvent)    {}
func (l *noopTelemetryLogger) LogAgentFinish(event types.AgentFinishEvent)  {}
func (l *noopTelemetryLogger) LogErrorf(format string, args ...interface{}) {}
func (l *noopTelemetryLogger) LogWarnf(format string, args ...interface{})  {}
func (l *noopTelemetryLogger) LogInfof(format string, args ...interface{})  {}
func (l *noopTelemetryLogger) LogDebugf(format string, args ...interface{}) {}
func (l *noopTelemetryLogger) LogPrompt(prompt string)                      {}
func (l *noopTelemetryLogger) LogSubagentActivity(event types.SubagentActivityEvent) {} // New method

// fileTelemetryLogger logs telemetry events to a specified file.
type stdoutTelemetryLogger struct {
	mu       sync.Mutex
	enabled  bool
	logLevel string
}

type fileTelemetryLogger struct {
	mu       sync.Mutex
	filePath string
	enabled  bool
	logLevel string
}

// NewFileTelemetryLogger creates a new fileTelemetryLogger.
func NewStdoutTelemetryLogger(enabled bool, logLevel string) *stdoutTelemetryLogger {
	return &stdoutTelemetryLogger{
		enabled:  enabled,
		logLevel: logLevel,
	}
}

func (l *stdoutTelemetryLogger) LogAgentStart(event types.AgentStartEvent) {
	if !l.enabled {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	logEntry := map[string]interface{}{
		"type":      "AgentStart",
		"agentID":   event.AgentID,
		"agentName": event.AgentName,
	}
	data, err := json.Marshal(logEntry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling AgentStart event: %v\n", err)
		return
	}

	l.writeLog(data)
}

func (l *stdoutTelemetryLogger) LogAgentFinish(event types.AgentFinishEvent) {
	if !l.enabled {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	logEntry := map[string]interface{}{
		"type":            "AgentFinish",
		"agentID":         event.AgentID,
		"agentName":       event.AgentName,
		"durationMs":      event.DurationMs,
		"turnCounter":     event.TurnCounter,
		"terminateReason": event.TerminateReason,
	}
	data, err := json.Marshal(logEntry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling AgentFinish event: %v\n", err)
		return
	}

	l.writeLog(data)
}

func (l *stdoutTelemetryLogger) writeLog(data []byte) {
	if _, err := os.Stdout.Write(data); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to stdout: %v\n", err)
		return
	}
	if _, err := os.Stdout.WriteString("\n"); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing newline to stdout: %v\n", err)
		return
	}
}

func (l *stdoutTelemetryLogger) LogErrorf(format string, args ...interface{}) {
	if !l.enabled {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	logEntry := map[string]interface{}{
		"type":      "Error",
		"message":   fmt.Sprintf(format, args...),
		"timestamp": time.Now().Format(time.RFC3339),
	}
	data, err := json.Marshal(logEntry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling error log entry: %v\n", err)
		return
	}
	l.writeLog(data)
}

func (l *stdoutTelemetryLogger) LogWarnf(format string, args ...interface{}) {
	if !l.enabled || (l.logLevel != "debug" && l.logLevel != "info" && l.logLevel != "warn") {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	logEntry := map[string]interface{}{
		"type":      "Warn",
		"message":   fmt.Sprintf(format, args...),
		"timestamp": time.Now().Format(time.RFC3339),
	}
	data, err := json.Marshal(logEntry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling warn log entry: %v\n", err)
		return
	}
	l.writeLog(data)
}

func (l *stdoutTelemetryLogger) LogInfof(format string, args ...interface{}) {
	if !l.enabled || (l.logLevel != "debug" && l.logLevel != "info") {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	logEntry := map[string]interface{}{
		"type":      "Info",
		"message":   fmt.Sprintf(format, args...),
		"timestamp": time.Now().Format(time.RFC3339),
	}
	data, err := json.Marshal(logEntry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling info log entry: %v\n", err)
		return
	}
	l.writeLog(data)
}

func (l *stdoutTelemetryLogger) LogDebugf(format string, args ...interface{}) {
	if !l.enabled || l.logLevel != "debug" {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	logEntry := map[string]interface{}{
		"type":      "Debug",
		"message":   fmt.Sprintf(format, args...),
		"timestamp": time.Now().Format(time.RFC3339),
	}
	data, err := json.Marshal(logEntry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling debug log entry: %v\n", err)
		return
	}
	l.writeLog(data)
}

func (l *stdoutTelemetryLogger) LogPrompt(prompt string) {
	if !l.enabled || l.logLevel != "debug" {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	logEntry := map[string]interface{}{
		"type":      "Prompt",
		"prompt":    prompt,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	data, err := json.Marshal(logEntry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling prompt log entry: %v\n", err)
		return
	}
	l.writeLog(data)
}

func (l *stdoutTelemetryLogger) LogSubagentActivity(event types.SubagentActivityEvent) {
	if !l.enabled || l.logLevel != "debug" { // Log subagent activity at debug level by default
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	// The event already contains necessary fields, just marshal and log
	data, err := json.Marshal(event)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling SubagentActivity event: %v\n", err)
		return
	}
	l.writeLog(data)
}

func NewFileTelemetryLogger(filePath string, enabled bool, logLevel string) *fileTelemetryLogger {
	return &fileTelemetryLogger{
		filePath: filePath,
		enabled:  enabled,
		logLevel: logLevel,
	}
}

func (l *fileTelemetryLogger) LogAgentStart(event types.AgentStartEvent) {
	if !l.enabled {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	logEntry := map[string]interface{}{
		"type":      "AgentStart",
		"agentID":   event.AgentID,
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
		"type":            "AgentFinish",
		"agentID":         event.AgentID,
		"agentName":       event.AgentName,
		"durationMs":      event.DurationMs,
		"turnCounter":     event.TurnCounter,
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
		"type":      "Error",
		"message":   fmt.Sprintf(format, args...),
		"timestamp": time.Now().Format(time.RFC3339),
	}
	data, err := json.Marshal(logEntry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling error log entry: %v\n", err)
		return
	}
	l.writeLog(data)
}

func (l *fileTelemetryLogger) LogWarnf(format string, args ...interface{}) {
	if !l.enabled || (l.logLevel != "debug" && l.logLevel != "info" && l.logLevel != "warn") {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	logEntry := map[string]interface{}{
		"type":      "Warn",
		"message":   fmt.Sprintf(format, args...),
		"timestamp": time.Now().Format(time.RFC3339),
	}
	data, err := json.Marshal(logEntry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling warn log entry: %v\n", err)
		return
	}
	l.writeLog(data)
}

func (l *fileTelemetryLogger) LogInfof(format string, args ...interface{}) {
	if !l.enabled || (l.logLevel != "debug" && l.logLevel != "info") {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	logEntry := map[string]interface{}{
		"type":      "Info",
		"message":   fmt.Sprintf(format, args...),
		"timestamp": time.Now().Format(time.RFC3339),
	}
	data, err := json.Marshal(logEntry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling info log entry: %v\n", err)
		return
	}
	l.writeLog(data)
}

func (l *fileTelemetryLogger) LogDebugf(format string, args ...interface{}) {
	if !l.enabled || l.logLevel != "debug" {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	logEntry := map[string]interface{}{
		"type":      "Debug",
		"message":   fmt.Sprintf(format, args...),
		"timestamp": time.Now().Format(time.RFC3339),
	}
	data, err := json.Marshal(logEntry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling debug log entry: %v\n", err)
		return
	}
	l.writeLog(data)
}

func (l *fileTelemetryLogger) LogPrompt(prompt string) {
	if !l.enabled || l.logLevel != "debug" {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	logEntry := map[string]interface{}{
		"type":      "Prompt",
		"prompt":    prompt,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	data, err := json.Marshal(logEntry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling prompt log entry: %v\n", err)
		return
	}
	l.writeLog(data)
}

func (l *fileTelemetryLogger) LogSubagentActivity(event types.SubagentActivityEvent) {
	if !l.enabled || l.logLevel != "debug" { // Log subagent activity at debug level by default
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	// The event already contains necessary fields, just marshal and log
	data, err := json.Marshal(event)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling SubagentActivity event: %v\n", err)
		return
	}
	l.writeLog(data)
}

// NewTelemetryLogger creates a TelemetryLogger based on the provided telemetry settings.
func NewTelemetryLogger(settings *types.TelemetrySettings) TelemetryLogger {
	if settings == nil || !settings.Enabled {
		return &noopTelemetryLogger{}
	}

	backend := settings.Backend
	if backend == "" {
		backend = "stdout" // Default to stdout if not specified
	}

	switch backend {
	case "stdout":
		return NewStdoutTelemetryLogger(true, settings.LogLevel)
	case "file":
		if settings.OutDir == "" {
			fmt.Fprintln(os.Stderr, "telemetry backend is 'file' but outdir is not set")
			return &noopTelemetryLogger{}
		}
		expandedPath, err := pathutils.ExpandPath(settings.OutDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error expanding telemetry output directory path %s: %v\n", settings.OutDir, err)
			return &noopTelemetryLogger{}
		}

		// Ensure the directory exists
		if err := os.MkdirAll(expandedPath, 0755); err != nil {
			// Log error to stderr as the logger isn't initialized yet
			fmt.Fprintf(os.Stderr, "Error creating telemetry output directory %s: %v\n", expandedPath, err)
			// Fallback to no-op logger
			return &noopTelemetryLogger{}
		}
		logFilePath := filepath.Join(expandedPath, "go-ai-agent.log")
		return NewFileTelemetryLogger(logFilePath, true, settings.LogLevel)
	default:
		fmt.Fprintf(os.Stderr, "unknown telemetry backend: %s, falling back to no-op logger\n", backend)
		return &noopTelemetryLogger{}
	}
}
