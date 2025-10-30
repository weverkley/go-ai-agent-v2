package output

import (
	"encoding/json"
	"fmt"
	"regexp"

	"go-ai-agent-v2/go-cli/pkg/types" // For JsonOutput, JsonError, SessionMetrics
)

// JsonFormatter formats responses, stats, and errors into a JSON string.
type JsonFormatter struct{}

// NewJsonFormatter creates a new instance of JsonFormatter.
func NewJsonFormatter() *JsonFormatter {
	return &JsonFormatter{}
}

// stripAnsi removes ANSI escape codes from a string.
func stripAnsi(str string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return re.ReplaceAllString(str, "")
}

// Format formats the response, stats, and error into a JSON string.
func (jf *JsonFormatter) Format(response *string, stats *types.SessionMetrics, err *types.JsonError) string {
	output := types.JsonOutput{}

	if response != nil {
		strippedResponse := stripAnsi(*response)
		output.Response = &strippedResponse
	}

	if stats != nil {
		output.Stats = stats
	}

	if err != nil {
		output.Error = err
	}

	jsonBytes, marshalErr := json.MarshalIndent(output, "", "  ")
	if marshalErr != nil {
		return fmt.Sprintf("Error marshaling JSON output: %v", marshalErr)
	}
	return string(jsonBytes)
}

// FormatError formats an error into a JsonError and then into a JSON string.
func (jf *JsonFormatter) FormatError(err error, code *string) string {
	jsonError := types.JsonError{
		Type:    fmt.Sprintf("%T", err), // Get type name
		Message: stripAnsi(err.Error()),
		Code:    code,
	}

	return jf.Format(nil, nil, &jsonError)
}
