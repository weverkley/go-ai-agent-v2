package agents

import (
	"context"
	"time"

	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/types"
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
	Name          string                   `json:"name"`
	DisplayName   string                   `json:"displayName"`
	Description   string                   `json:"description"`
	InputConfig   InputConfig              `json:"inputConfig"`
	OutputConfig  *OutputConfig            `json:"outputConfig,omitempty"`
	ProcessOutput func(interface{}) string `json:"-"` // Not directly translatable to JSON
	ModelConfig   ModelConfig              `json:"modelConfig"`
	RunConfig     RunConfig                `json:"runConfig"`
	ToolConfig    *ToolConfig              `json:"toolConfig,omitempty"`
	PromptConfig  PromptConfig             `json:"promptConfig"`
}

// InputConfig defines the input parameters for an agent.
type InputConfig struct {
	Inputs map[string]InputParameter `json:"inputs"`
}

// InputParameter defines a single input parameter.
type InputParameter struct {
	Description string `json:"description"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
}

// BaseToolInvocation provides common fields and methods for tool invocations.
type BaseToolInvocation struct {
	Params     AgentInputs
	MessageBus interface{} // Placeholder for MessageBus
}

// SubagentInvocation represents a validated, executable instance of a subagent tool.
type SubagentInvocation struct {
	BaseToolInvocation
	Definition AgentDefinition
	Config     *config.Config
}

// PromptConfig defines the prompting strategy for the agent.
type PromptConfig struct {
	SystemPrompt    string       `json:"systemPrompt,omitempty"`
	InitialMessages []types.Part `json:"initialMessages,omitempty"`
	Query           string       `json:"query,omitempty"`
}

// ModelConfig defines the model parameters for the agent.
type ModelConfig struct {
	Model          string  `json:"model"`
	Temperature    float32 `json:"temperature"`
	TopP           float32 `json:"topP"`
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

// AgentInputs represents the input parameters for an agent invocation.
type AgentInputs map[string]interface{}

// OutputObject represents the final output of an agent run.
type OutputObject struct {
	Result          string                   `json:"result"`
	TerminateReason types.AgentTerminateMode `json:"terminate_reason"`
}

// CodebaseInvestigationReportSchema represents the schema for the codebase investigation report.
type CodebaseInvestigationReportSchema struct {
	SummaryOfFindings string   `json:"SummaryOfFindings"`
	ExplorationTrace  []string `json:"ExplorationTrace"`
	RelevantLocations []struct {
		FilePath   string   `json:"FilePath"`
		Reasoning  string   `json:"Reasoning"`
		KeySymbols []string `json:"KeySymbols"`
	} `json:"RelevantLocations"`
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
	Execute(ctx context.Context, liveOutputCallback func(string), shellExecutionConfig interface{}, setPidCallback func(int)) (types.ToolResult, error)
	ShouldConfirmExecute(ctx context.Context) (types.ToolCallConfirmationDetails, error)
	// Add other methods as needed from the JS interface
}

// ToolCall represents a single tool call in its various states.
// This will be a discriminated union in Go using an interface and concrete types.
type ToolCall interface {
	GetStatus() string
	GetRequest() types.ToolCallRequestInfo
	GetTool() types.Tool // Changed to types.Tool
	GetInvocation() AnyToolInvocation
	GetOutcome() types.ToolConfirmationOutcome
	GetStartTime() *time.Time
	GetDurationMs() *int64
	GetResponse() *types.ToolCallResponseInfo
}

// BaseToolCall provides common fields for all ToolCall types.
type BaseToolCall struct {
	Request    types.ToolCallRequestInfo
	Tool       types.Tool // Changed to types.Tool
	Invocation AnyToolInvocation
	StartTime  *time.Time
	Outcome    types.ToolConfirmationOutcome
}

func (b *BaseToolCall) GetRequest() types.ToolCallRequestInfo     { return b.Request }
func (b *BaseToolCall) GetTool() types.Tool                       { return b.Tool }
func (b *BaseToolCall) GetInvocation() AnyToolInvocation          { return b.Invocation }
func (b *BaseToolCall) GetOutcome() types.ToolConfirmationOutcome { return b.Outcome }
func (b *BaseToolCall) GetStartTime() *time.Time                  { return b.StartTime }
func (b *BaseToolCall) GetDurationMs() *int64                     { return nil } // Default, overridden by completed calls
func (b *BaseToolCall) GetResponse() *types.ToolCallResponseInfo  { return nil } // Default, overridden by completed calls

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
	Response   types.ToolCallResponseInfo
	DurationMs *int64
}

func (e *ErroredToolCall) GetStatus() string                        { return "error" }
func (e *ErroredToolCall) GetResponse() *types.ToolCallResponseInfo { return &e.Response }
func (e *ErroredToolCall) GetDurationMs() *int64                    { return e.DurationMs }

// SuccessfulToolCall
type SuccessfulToolCall struct {
	BaseToolCall
	Response   types.ToolCallResponseInfo
	DurationMs *int64
}

func (s *SuccessfulToolCall) GetStatus() string                        { return "success" }
func (s *SuccessfulToolCall) GetResponse() *types.ToolCallResponseInfo { return &s.Response }
func (s *SuccessfulToolCall) GetDurationMs() *int64                    { return s.DurationMs }

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
	Response   types.ToolCallResponseInfo
	DurationMs *int64
}

func (c *CancelledToolCall) GetStatus() string                        { return "cancelled" }
func (c *CancelledToolCall) GetResponse() *types.ToolCallResponseInfo { return &c.Response }
func (c *CancelledToolCall) GetDurationMs() *int64                    { return c.DurationMs }

// WaitingToolCall
type WaitingToolCall struct {
	BaseToolCall
	ConfirmationDetails types.ToolCallConfirmationDetails
}

func (w *WaitingToolCall) GetStatus() string { return "awaiting_approval" }

// CompletedToolCall is an alias for ToolCall that has reached a terminal state.
type CompletedToolCall ToolCall
