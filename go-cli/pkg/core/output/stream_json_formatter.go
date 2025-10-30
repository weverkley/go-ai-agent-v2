package output

import (
	"encoding/json"
	"fmt"
	"os"

	"go-ai-agent-v2/go-cli/pkg/types" // For JsonStreamEvent, StreamStats, SessionMetrics
)

// StreamJsonFormatter formats for streaming JSON output.
type StreamJsonFormatter struct{}

// NewStreamJsonFormatter creates a new instance of StreamJsonFormatter.
func NewStreamJsonFormatter() *StreamJsonFormatter {
	return &StreamJsonFormatter{}
}

// FormatEvent formats a single event as a JSON string with newline (JSONL format).
func (sjf *StreamJsonFormatter) FormatEvent(event types.JsonStreamEvent) string {
	jsonBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Sprintf("{\"type\":\"ERROR\",\"payload\":{\"message\":\"Error marshaling event: %v\"}}\n", err)
	}
	return string(jsonBytes) + "\n"
}

// EmitEvent emits an event directly to stdout in JSONL format.
func (sjf *StreamJsonFormatter) EmitEvent(event types.JsonStreamEvent) {
	os.Stdout.WriteString(sjf.FormatEvent(event))
}

// ConvertToStreamStats converts SessionMetrics to simplified StreamStats format.
func (sjf *StreamJsonFormatter) ConvertToStreamStats(
	metrics types.SessionMetrics,
	durationMs int,
) types.StreamStats {
	// TODO: Implement aggregation of token counts across all models if SessionMetrics is expanded.
	// For now, using placeholder values or direct mapping if available.
	return types.StreamStats{
		TotalTokens: metrics.TotalTurns, // Placeholder
		InputTokens: metrics.TotalTurns, // Placeholder
		OutputTokens: metrics.TotalTurns, // Placeholder
		DurationMs:  durationMs,
		ToolCalls:   0, // Placeholder
	}
}
