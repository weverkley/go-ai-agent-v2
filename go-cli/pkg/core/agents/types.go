package agents

import (
	"context"
	"time"

	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/tools"
)

// AgentTerminateMode defines the reasons an agent might terminate.
type AgentTerminateMode string

const (
	AgentTerminateModeError    AgentTerminateMode = "ERROR"
	AgentTerminateModeGoal     AgentTerminateMode = "GOAL"
	AgentTerminateModeMaxTurns AgentTerminateMode = "MAX_TURNS"
	AgentTerminateModeTimeout  AgentTerminateMode = "TIMEOUT"
	AgentTerminateModeAborted  AgentTerminateMode = "ABORTED"

	TASK_COMPLETE_TOOL_NAME = "complete_task"
)

// SubagentActivityEvent represents an activity event emitted by a subagent.
type SubagentActivityEvent struct {
	IsSubagentActivityEvent bool                   `json:"isSubagentActivityEvent"`
	AgentName               string                 `json:"agentName"`
	Type                    string                 `json:"type"`
	Data                    map[string]interface{} `json:"data"`
}

// ActivityCallback is a callback function to report on agent activity.
type ActivityCallback func(activity SubagentActivityEvent)

// AgentDefinition defines the structure and behavior of an agent.
type AgentDefinition struct {
	Name        string       `json:"name"`
	PromptConfig PromptConfig `json:"promptConfig"`
	ModelConfig  ModelConfig  `json:"modelConfig"`
	ToolConfig   *ToolConfig  `json:"toolConfig,omitempty"`
	RunConfig    RunConfig    `json:"runConfig"`
	OutputConfig *OutputConfig `json:"outputConfig,omitempty"`
	// ProcessOutput func(interface{}) string `json:"-"` // Not directly translatable to JSON
}

// PromptConfig defines the prompting strategy for the agent.
type PromptConfig struct {
	SystemPrompt    string   `json:"systemPrompt,omitempty"`
	InitialMessages []Part   `json:"initialMessages,omitempty"`
	Query           string   `json:"query,omitempty"`
}

// ModelConfig defines the model parameters for the agent.
type ModelConfig struct {
	Model        string  `json:"model"`
	Temperature  float32 `json:"temperature"`
	TopP         float32 `json:"topP"`
	ThinkingBudget int     `json:"thinkingBudget,omitempty"`
}

// ToolConfig defines the tools available to the agent.
type ToolConfig struct {
	Tools []string `json:"tools"` // For now, just names. Will expand to AnyDeclarativeTool.
}

// RunConfig defines the execution parameters for the agent.
type RunConfig struct {
	MaxTurns       int `json:"maxTurns"`
	MaxTimeMinutes int `json:"maxTimeMinutes"`
}

// OutputConfig defines how the agent's output should be handled.
type OutputConfig struct {
	OutputName string `json:"outputName"`	
	// Schema     interface{} `json:"-"` // Placeholder for Zod schema equivalent
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

// AgentInputs represents the input parameters for an agent invocation.
type AgentInputs map[string]interface{}

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

// OutputObject represents the final output of an agent run.
type OutputObject struct {
	Result         string             `json:"result"`
	TerminateReason AgentTerminateMode `json:"terminate_reason"`
}

// FolderStructureOptions for customizing folder structure retrieval.
type FolderStructureOptions struct {
	MaxItems           *int    // Maximum number of files and folders combined to display. Defaults to 200.
	IgnoredFolders     *[]string // Set of folder names to ignore completely. Case-sensitive.
	FileIncludePattern *string // Optional regex to filter included files by name.
	// FileService        FileDiscoveryService // For filtering files.
	FileFilteringOptions *FileFilteringOptions // File filtering ignore options.
}

// FileFilteringOptions for filtering files.
type FileFilteringOptions struct {
	RespectGitIgnore  *bool `json:"respectGitIgnore,omitempty"`
	RespectGeminiIgnore *bool `json:"respectGeminiIgnore,omitempty"`
}

// FullFolderInfo represents the full, unfiltered information about a folder and its contents.
type FullFolderInfo struct {
	Name            string
	Path            string
	Files           []string
	SubFolders      []FullFolderInfo
	TotalChildren   int
	TotalFiles      int
	IsIgnored       bool // Flag to easily identify ignored folders later
	HasMoreFiles    bool // Indicates if files were truncated for this specific folder
	HasMoreSubfolders bool // Indicates if subfolders were truncated for this specific folder
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

// ToolErrorType defines types of tool errors.
type ToolErrorType string

const (
	ToolErrorTypeToolNotRegistered ToolErrorType = "TOOL_NOT_REGISTERED"
	ToolErrorTypeInvalidToolParams ToolErrorType = "INVALID_TOOL_PARAMS"
	ToolErrorTypeUnhandledException ToolErrorType = "UNHANDLED_EXCEPTION"
)

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

// ToolResult represents the result of a tool execution.
type ToolResult struct {
	LLMContent  interface{} `json:"llmContent"` // Can be string or []Part
	ReturnDisplay string      `json:"returnDisplay"`
	Error       *ToolError  `json:"error,omitempty"`
}

// ToolError represents an error from a tool.
type ToolError struct {
	Message string        `json:"message"`
	Type    ToolErrorType `json:"type"`
}

// AnsiOutput represents ANSI formatted output.
type AnsiOutput string

// AnyDeclarativeTool is an interface for any declarative tool.
type AnyDeclarativeTool interface {
	Name() string
	Build(args map[string]interface{}) (AnyToolInvocation, error)
	// Add other methods as needed from the JS interface
}

// AnyToolInvocation is an interface for any tool invocation.
type AnyToolInvocation interface {
	Execute(ctx context.Context, liveOutputCallback func(string), shellExecutionConfig interface{}, setPidCallback func(int)) (ToolResult, error)
	ShouldConfirmExecute(ctx context.Context) (ToolCallConfirmationDetails, error)
	// Add other methods as needed from the JS interface
}

// ToolCall represents a single tool call in its various states.
// This will be a discriminated union in Go using an interface and concrete types.
type ToolCall interface {
	GetStatus() string
	GetRequest() ToolCallRequestInfo
	GetTool() AnyDeclarativeTool
	GetInvocation() AnyToolInvocation
	GetOutcome() ToolConfirmationOutcome
	GetStartTime() *time.Time
	GetDurationMs() *int64
	GetResponse() *ToolCallResponseInfo
}

// BaseToolCall provides common fields for all ToolCall types.
type BaseToolCall struct {
	Request    ToolCallRequestInfo
	Tool       AnyDeclarativeTool
	Invocation AnyToolInvocation
	StartTime  *time.Time
	Outcome    ToolConfirmationOutcome
}

func (b *BaseToolCall) GetRequest() ToolCallRequestInfo { return b.Request }
func (b *BaseToolCall) GetTool() AnyDeclarativeTool     { return b.Tool }
func (b *BaseToolCall) GetInvocation() AnyToolInvocation { return b.Invocation }
func (b *BaseToolCall) GetOutcome() ToolConfirmationOutcome { return b.Outcome }
func (b *BaseToolCall) GetStartTime() *time.Time { return b.StartTime }
func (b *BaseToolCall) GetDurationMs() *int64 { return nil } // Default, overridden by completed calls
func (b *BaseToolCall) GetResponse() *ToolCallResponseInfo { return nil } // Default, overridden by completed calls

// ValidatingToolCall
type ValidatingToolCall struct {
	BaseToolCall
}

func (v *ValidatingToolCall) GetStatus() string { return "validating" }

// ScheduledToolCall
type ScheduledToolCall struct {
	BaseToolCall
}

func (s *ScheduledToolCall) GetStatus() string { return "scheduled" }

// ErroredToolCall
type ErroredToolCall struct {
	BaseToolCall
	Response ToolCallResponseInfo
	DurationMs *int64
}

func (e *ErroredToolCall) GetStatus() string { return "error" }
func (e *ErroredToolCall) GetResponse() *ToolCallResponseInfo { return &e.Response }
func (e *ErroredToolCall) GetDurationMs() *int64 { return e.DurationMs }

// SuccessfulToolCall
type SuccessfulToolCall struct {
	BaseToolCall
	Response ToolCallResponseInfo
	DurationMs *int64
}

func (s *SuccessfulToolCall) GetStatus() string { return "success" }
func (s *SuccessfulToolCall) GetResponse() *ToolCallResponseInfo { return &s.Response }
func (s *SuccessfulToolCall) GetDurationMs() *int64 { return s.DurationMs }

// ExecutingToolCall
type ExecutingToolCall struct {
	BaseToolCall
	LiveOutput *interface{} // string or AnsiOutput
	PID        *int
}

func (e *ExecutingToolCall) GetStatus() string { return "executing" }

// CancelledToolCall
type CancelledToolCall struct {
	BaseToolCall
	Response ToolCallResponseInfo
	DurationMs *int64
}

func (c *CancelledToolCall) GetStatus() string { return "cancelled" }
func (c *CancelledToolCall) GetResponse() *ToolCallResponseInfo { return &c.Response }
func (c *CancelledToolCall) GetDurationMs() *int64 { return c.DurationMs }

// WaitingToolCall
type WaitingToolCall struct {
	BaseToolCall
	ConfirmationDetails ToolCallConfirmationDetails
}

func (w *WaitingToolCall) GetStatus() string { return "awaiting_approval" }

// CompletedToolCall is an alias for ToolCall that has reached a terminal state.
type CompletedToolCall ToolCall