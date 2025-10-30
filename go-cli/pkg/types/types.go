package types

import (
	"context"

	"go-ai-agent-v2/go-cli/pkg/tools"

	"github.com/google/generative-ai-go/genai"
)

// ApprovalMode defines the approval mode for tool calls.
type ApprovalMode string

const (
	ApprovalModeDefault  ApprovalMode = "default"
	ApprovalModeAutoEdit ApprovalMode = "autoEdit"
	ApprovalModeYOLO     ApprovalMode = "yolo"
)

// MCPServerStatus represents the connection status of an MCP server.
type MCPServerStatus string

const (
	CONNECTED    MCPServerStatus = "CONNECTED"
	CONNECTING   MCPServerStatus = "CONNECTING"
	DISCONNECTED MCPServerStatus = "DISCONNECTED"
)

// MCPOAuthConfig represents the OAuth configuration for an MCP server.
type MCPOAuthConfig struct {
	// Placeholder for now, add fields as needed.
}

// AuthProviderType represents the authentication provider type.
type AuthProviderType string

// MCPServerConfig represents the configuration for an MCP server.
type MCPServerConfig struct {
	Command              string            `json:"command,omitempty"`
	Args                 []string          `json:"args,omitempty"`
	Env                  map[string]string `json:"env,omitempty"`	
	Cwd                  string            `json:"cwd,omitempty"`
	Url                  string            `json:"url,omitempty"`
	HttpUrl              string            `json:"httpUrl,omitempty"`
	Headers              map[string]string `json:"headers,omitempty"`	
	Tcp                  string            `json:"tcp,omitempty"`
	Timeout              int               `json:"timeout,omitempty"`
	Trust                bool              `json:"trust,omitempty"`
	Description          string            `json:"description,omitempty"`
	IncludeTools         []string          `json:"includeTools,omitempty"`
	ExcludeTools         []string          `json:"excludeTools,omitempty"`
	Extension            *ExtensionInfo    `json:"extension,omitempty"` // Reference to the extension that provided this config
	Oauth                *MCPOAuthConfig   `json:"oauth,omitempty"`
	AuthProviderType     AuthProviderType  `json:"authProviderType,omitempty"`
	TargetAudience       string            `json:"targetAudience,omitempty"`
	TargetServiceAccount string            `json:"targetServiceAccount,omitempty"`
}

// ExtensionInfo represents basic information about an extension.
type ExtensionInfo struct {
	Name string `json:"name"`
}

// GenerateContentConfig represents the generation configuration for the model.
type GenerateContentConfig struct {
	Temperature     float32        `json:"temperature,omitempty"`
	TopP            float32        `json:"topP,omitempty"`
	ThinkingConfig  *ThinkingConfig `json:"thinkingConfig,omitempty"`
	SystemInstruction string         `json:"systemInstruction,omitempty"`
}

// ThinkingConfig represents the thinking configuration for the model.
type ThinkingConfig struct {
	IncludeThoughts bool `json:"includeThoughts,omitempty"`
	ThinkingBudget  int  `json:"thinkingBudget,omitempty"`
}

// MessageParams represents parameters for sending a message.
type MessageParams struct {
	Message     []Part
	Tools       []tools.FunctionDeclaration // Assuming tools.FunctionDeclaration is defined
	AbortSignal context.Context
}

// StreamEventType defines the type of event in the stream.
type StreamEventType string

const (
	StreamEventTypeChunk StreamEventType = "CHUNK"
	StreamEventTypeError StreamEventType = "ERROR"
	StreamEventTypeDone  StreamEventType = "DONE"
)

// StreamResponse represents a response from the stream.
type StreamResponse struct {
	Type  StreamEventType
	Value *genai.GenerateContentResponse // Or a custom struct that mirrors it
	Error error
}

// FileFilteringOptions for filtering files.
type FileFilteringOptions struct {
	RespectGitIgnore  *bool `json:"respectGitIgnore,omitempty"`
	RespectGeminiIgnore *bool `json:"respectGeminiIgnore,omitempty"`
}

// Part represents a part of a content message.
// This is a simplified version, will need to be expanded based on actual usage.
type Part struct {
	Text             string                 `json:"text,omitempty"`	
	FunctionResponse *FunctionResponse      `json:"functionResponse,omitempty"`
	InlineData       *InlineData            `json:"inlineData,omitempty"`
	FileData         *FileData              `json:"fileData,omitempty"`
	Thought          string                 `json:"thought,omitempty"` // For thought parts
}

// FunctionResponse represents a function response part.
type FunctionResponse struct {
	ID       string                 `json:"id,omitempty"`
	Name     string                 `json:"name"`
	Response map[string]interface{} `json:"response"`
}

// InlineData represents inline data part.
type InlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"` // Base64 encoded
}

// FileData represents file data part.
type FileData struct {
	MimeType string `json:"mimeType"`
	FileURL  string `json:"fileUri"`
}

// JsonOutput represents the JSON output structure.
type JsonOutput struct {
	Response *string        `json:"response,omitempty"`
	Stats    *SessionMetrics `json:"stats,omitempty"`
	Error    *JsonError     `json:"error,omitempty"`
}

// JsonError represents a JSON error structure.
type JsonError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Code    *string `json:"code,omitempty"`
}

// SessionMetrics represents session-related metrics for telemetry.
type SessionMetrics struct {
	// Add fields as needed based on uiTelemetry.js
	// For now, a placeholder.
	TotalTurns int `json:"totalTurns"`
	TotalTimeMs int `json:"totalTimeMs"`
}

// JsonStreamEvent represents a single event in the JSON stream.
type JsonStreamEvent struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}

// StreamStats represents simplified stats for streaming output.
type StreamStats struct {
	TotalTokens int `json:"total_tokens"`
	InputTokens int `json:"input_tokens"`	
	OutputTokens int `json:"output_tokens"`
	DurationMs  int `json:"duration_ms"`
	ToolCalls   int `json:"tool_calls"`
}

// ToolErrorType defines types of tool errors.
type ToolErrorType string

const (
	ToolErrorTypeToolNotRegistered ToolErrorType = "TOOL_NOT_REGISTERED"
	ToolErrorTypeInvalidToolParams ToolErrorType = "INVALID_TOOL_PARAMS"
	ToolErrorTypeUnhandledException ToolErrorType = "UNHANDLED_EXCEPTION"
	ToolErrorTypeExecutionFailed   ToolErrorType = "EXECUTION_FAILED"
)

// ToolResult represents the result of a tool execution.
type ToolResult struct {
	LLMContent  interface{} `json:"llmContent"` // Can be string or []types.Part
	ReturnDisplay string      `json:"returnDisplay"`
	Error       *ToolError  `json:"error,omitempty"`
}

// ToolError represents an error from a tool.
type ToolError struct {
	Message string        `json:"message"`
	Type    ToolErrorType `json:"type"`
}

// ToolCallRequestInfo represents the information for a tool call request.
type ToolCallRequestInfo struct {
	CallID          string                 `json:"callId"`
	Name            string                 `json:"name"`
	Args            map[string]interface{} `json:"args"`
	IsClientInitiated bool                   `json:"isClientInitiated"`
	PromptID        string                 `json:"prompt_id"`
}

// ToolResultDisplay represents the display information for a tool result.
type ToolResultDisplay struct {
	FileDiff        string `json:"fileDiff,omitempty"`
	FileName        string `json:"fileName,omitempty"`
	OriginalContent string `json:"originalContent,omitempty"`
	NewContent      string `json:"newContent,omitempty"`
}

// ToolConfirmationOutcome defines the outcome of a tool confirmation.
type ToolConfirmationOutcome string

const (
	ToolConfirmationOutcomeProceedAlways ToolConfirmationOutcome = "PROCEED_ALWAYS"
	ToolConfirmationOutcomeProceedOnce   ToolConfirmationOutcome = "PROCEED_ONCE"
	ToolConfirmationOutcomeCancel        ToolConfirmationOutcome = "CANCEL"
	ToolConfirmationOutcomeModifyWithEditor ToolConfirmationOutcome = "MODIFY_WITH_EDITOR"
)

// ToolCallConfirmationDetails represents details for tool call confirmation.
type ToolCallConfirmationDetails struct {
	Type              string            `json:"type"` // e.g., "edit", "shell"
	Message           string            `json:"message"`
	ToolName          string            `json:"toolName"`
	ToolArgs          map[string]interface{} `json:"toolArgs"`
	FileDiff          string            `json:"fileDiff,omitempty"`
	FileName          string            `json:"fileName,omitempty"`
	OriginalContent   string            `json:"originalContent,omitempty"`
	NewContent        string            `json:"newContent,omitempty"`
	IdeConfirmation   interface{}       `json:"ideConfirmation,omitempty"` // Placeholder for now
	OnConfirm         interface{}       `json:"onConfirm,omitempty"`       // Placeholder for now
	IsModifying       bool              `json:"isModifying,omitempty"`
}

// EditorType represents the type of editor.
type EditorType string

// ToolCallResponseInfo represents the response information for a tool call.
type ToolCallResponseInfo struct {
	CallID        string            `json:"callId"`
	Error         error             `json:"error,omitempty"`
	ResponseParts []Part            `json:"responseParts"`
	ResultDisplay *ToolResultDisplay `json:"resultDisplay,omitempty"`
	ErrorType     ToolErrorType     `json:"errorType,omitempty"`
	OutputFile    string            `json:"outputFile,omitempty"`
	ContentLength int               `json:"contentLength,omitempty"`
}

// JsonSchemaObject defines the structure for a JSON Schema object.
type JsonSchemaObject struct {
	Type       string                            `json:"type"` // "object"
	Properties map[string]JsonSchemaProperty `json:"properties"`
	Required   []string                          `json:"required,omitempty"`
}

// JsonSchemaProperty defines the structure for a property within a JsonSchemaObject.
type JsonSchemaProperty struct {
	Type        string `json:"type"` // "string", "number", "integer", "boolean", "array"
	Description string `json:"description"`
	Items       *struct {
		Type string `json:"type"` // "string", "number"
	} `json:"items,omitempty"`
}
