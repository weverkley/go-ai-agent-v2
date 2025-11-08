package types

import (
	"context"
	"fmt"
	"sync" // Add sync import

	"github.com/google/generative-ai-go/genai"
)

// ApprovalMode defines the approval mode for tool calls.
type ApprovalMode string

// SettingScope defines the scope of a setting.
type SettingScope string

const (
	SettingScopeUser      SettingScope = "user"
	SettingScopeWorkspace SettingScope = "workspace"
)

// Tool names
const (
	LS_TOOL_NAME                = "ls"
	READ_FILE_TOOL_NAME         = "read_file"
	GLOB_TOOL_NAME              = "glob"
	GREP_TOOL_NAME              = "grep"
	SMART_EDIT_TOOL_NAME        = "smart_edit"
	WEB_FETCH_TOOL_NAME         = "web_fetch"
	WEB_SEARCH_TOOL_NAME        = "web_search"
	MEMORY_TOOL_NAME            = "memory"
	WRITE_TODOS_TOOL_NAME       = "write_todos"
	LIST_DIRECTORY_TOOL_NAME    = "list_directory"
	GET_CURRENT_BRANCH_TOOL_NAME = "get_current_branch"
	GET_REMOTE_URL_TOOL_NAME    = "get_remote_url"
	CHECKOUT_BRANCH_TOOL_NAME   = "checkout_branch"
	PULL_TOOL_NAME              = "pull"
	EXECUTE_COMMAND_TOOL_NAME   = "execute_command"
	READ_MANY_FILES_TOOL_NAME   = "read_many_files"
	FIND_UNUSED_CODE_TOOL_NAME  = "find_unused_code"
	EXTRACT_FUNCTION_TOOL_NAME  = "extract_function"
	WRITE_FILE_TOOL_NAME        = "write_file"
)

// MCPServerStatus represents the connection status of an MCP server.
type MCPServerStatus struct {
	Name        string
	Status      string
	Url         string
	Description string
}

const (
	MCPServerStatusConnected    string = "CONNECTED"
	MCPServerStatusDisconnected string = "DISCONNECTED"
	MCPServerStatusConnecting   string = "CONNECTING"
	MCPServerStatusError        string = "ERROR"
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
	Temperature       float32         `json:"temperature,omitempty"`
	TopP              float32         `json:"topP,omitempty"`
	ThinkingConfig    *ThinkingConfig `json:"thinkingConfig,omitempty"`
	SystemInstruction string          `json:"systemInstruction,omitempty"`
}

// ThinkingConfig represents the thinking configuration for the model.
type ThinkingConfig struct {
	IncludeThoughts bool `json:"includeThoughts,omitempty"`
	ThinkingBudget  int  `json:"thinkingBudget,omitempty"`
}

// MessageParams represents parameters for sending a message.
type MessageParams struct {
	Message     []Part
	Tools       []*genai.Tool
	AbortSignal context.Context
}

// StreamEventType defines the type of event in the stream.
type StreamEventType string

const (
	StreamEventTypeChunk StreamEventType = "chunk"
	StreamEventTypeError StreamEventType = "error"
)

// StreamResponse represents a response from the stream.
type StreamResponse struct {
	Type  StreamEventType
	Value *genai.GenerateContentResponse // Or a custom struct that mirrors it
	Error error
}

// FileFilteringOptions for filtering files.
type FileFilteringOptions struct {
	RespectGitIgnore    *bool `json:"respectGitIgnore,omitempty"`
	RespectGeminiIgnore *bool `json:"respectGeminiIgnore,omitempty"`
}

// Part represents a part of a content message.
// This is a simplified version, will need to be expanded based on actual usage.
type Part struct {
	Text             string            `json:"text,omitempty"`
	FunctionCall     *FunctionCall     `json:"functionCall,omitempty"` // Added this line
	FunctionResponse *FunctionResponse `json:"functionResponse,omitempty"`
	InlineData       *InlineData       `json:"inlineData,omitempty"`
	FileData         *FileData         `json:"fileData,omitempty"`
	Thought          string            `json:"thought,omitempty"` // For thought parts
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
	Response *string         `json:"response,omitempty"`
	Stats    *SessionMetrics `json:"stats,omitempty"`
	Error    *JsonError      `json:"error,omitempty"`
}

// JsonError represents a JSON error structure.
type JsonError struct {
	Type    string  `json:"type"`
	Message string  `json:"message"`
	Code    *string `json:"code,omitempty"`
}

// SessionMetrics represents session-related metrics for telemetry.
type SessionMetrics struct {
	TotalTurns   int `json:"totalTurns"`
	TotalTimeMs  int `json:"totalTimeMs"`
	InputTokens  int `json:"inputTokens"`
	OutputTokens int `json:"outputTokens"`
	TotalTokens  int `json:"totalTokens"`
}

// JsonStreamEvent represents a single event in the JSON stream.
type JsonStreamEvent struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}

// StreamStats represents simplified stats for streaming output.
type StreamStats struct {
	TotalTokens  int `json:"total_tokens"`
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	DurationMs   int `json:"duration_ms"`
	ToolCalls    int `json:"tool_calls"`
}

// ToolErrorType defines types of tool errors.
type ToolErrorType string

const (
	ToolErrorTypeExecutionFailed ToolErrorType = "EXECUTION_FAILED"
)

// FunctionCall represents a function call requested by the model.
type FunctionCall struct {
	ID   string                 `json:"id,omitempty"`
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args"`
}

// ToolResult represents the result of a tool execution.
type ToolResult struct {
	LLMContent    interface{} `json:"llmContent"` // Can be string or []types.Part
	ReturnDisplay string      `json:"returnDisplay"`
	Error         *ToolError  `json:"error,omitempty"`
}

// ToolError represents an error from a tool.
type ToolError struct {
	Message string        `json:"message"`
	Type    ToolErrorType `json:"type"`
}

// ToolCallRequestInfo represents the information for a tool call request.
type ToolCallRequestInfo struct {
	CallID            string                 `json:"callId"`
	Name              string                 `json:"name"`
	Args              map[string]interface{} `json:"args"`
	IsClientInitiated bool                   `json:"isClientInitiated"`
	PromptID          string                 `json:"prompt_id"`
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
)

// AgentTerminateMode defines the reasons an agent might terminate.
type AgentTerminateMode string

const (
	AgentTerminateModeAborted AgentTerminateMode = "ABORTED"
	AgentTerminateModeError   AgentTerminateMode = "ERROR"
	AgentTerminateModeGoal    AgentTerminateMode = "GOAL"
	AgentTerminateModeMaxTurns AgentTerminateMode = "MAX_TURNS"
	AgentTerminateModeTimeout AgentTerminateMode = "TIMEOUT"
)

const (
	TASK_COMPLETE_TOOL_NAME = "task_complete"
)

const (
	ApprovalModeDefault ApprovalMode = "DEFAULT"
)


// ToolCallConfirmationDetails represents details for tool call confirmation.
type ToolCallConfirmationDetails struct {
	Type            string                 `json:"type"` // e.g., "edit", "shell"
	Message         string                 `json:"message"`
	ToolName        string                 `json:"toolName"`
	ToolArgs        map[string]interface{} `json:"toolArgs"`
	FileDiff        string                 `json:"fileDiff,omitempty"`
	FileName        string                 `json:"fileName,omitempty"`
	OriginalContent string                 `json:"originalContent,omitempty"`
	NewContent      string                 `json:"newContent,omitempty"`
	IdeConfirmation interface{}            `json:"ideConfirmation,omitempty"` // Placeholder for now
	OnConfirm       interface{}            `json:"onConfirm,omitempty"`       // Placeholder for now
	IsModifying     bool                   `json:"isModifying,omitempty"`
}

// Content represents a message in the chat history, simplified from @google/genai Content type.
type Content struct {
	Parts []Part `json:"parts"`
	Role  string `json:"role"`
}

	// EditorType represents the type of editor.
	type EditorType string

	// ChatCompressionResult represents the result of a chat compression operation.
	type ChatCompressionResult struct {
		OriginalTokenCount int    `json:"originalTokenCount"`
		NewTokenCount      int    `json:"newTokenCount"`
		CompressionStatus  string `json:"compressionStatus"`
	}
// ToolCallResponseInfo represents the response information for a tool call.
type ToolCallResponseInfo struct {
	CallID        string             `json:"callId"`
	Error         error              `json:"error,omitempty"`
	ResponseParts []Part             `json:"responseParts"`
	ResultDisplay *ToolResultDisplay `json:"resultDisplay,omitempty"`
	ErrorType     ToolErrorType      `json:"errorType,omitempty"`
	OutputFile    string             `json:"outputFile,omitempty"`
	ContentLength int                `json:"contentLength,omitempty"`
}

// JsonSchemaObject defines the structure for a JSON Schema object.
type JsonSchemaObject struct {
	Type       string                        `json:"type"` // "object"
	Properties map[string]JsonSchemaProperty `json:"properties"`
	Required   []string                      `json:"required,omitempty"`
}

// JsonSchemaProperty defines the structure for a property within a JsonSchemaObject.
type JsonSchemaProperty struct {
	Type        string                        `json:"type"` // "string", "number", "integer", "boolean", "array", "object"
	Description string                        `json:"description"`
	Items       *JsonSchemaPropertyItem       `json:"items,omitempty"`
	Properties  map[string]JsonSchemaProperty `json:"properties,omitempty"` // Added Properties field
	Required    []string                      `json:"required,omitempty"`   // Added Required field for nested objects
}

// JsonSchemaPropertyItem defines the structure for items within a JsonSchemaProperty.
type JsonSchemaPropertyItem struct {
	Type string `json:"type"` // "string", "number"
}

// Tool is the interface that all tools must implement.
type Tool interface {
	Name() string
	Description() string
	ServerName() string // Add ServerName to the interface
	Definition() *genai.Tool
	Execute(args map[string]any) (ToolResult, error)
}

// ToolInvocation represents an executable instance of a tool.
type ToolInvocation interface {
	Execute(ctx context.Context, updateOutput func(output string), shellExecutionConfig interface{}, setPidCallback func(int)) (ToolResult, error)
	ShouldConfirmExecute(ctx context.Context) (ToolCallConfirmationDetails, error)
	GetDescription() string
}

// Kind represents the type of tool.
type Kind string

const (
	KindOther Kind = "OTHER"
)

// BaseDeclarativeTool provides a base implementation for declarative tools.
type BaseDeclarativeTool struct {
	name             string
	displayName      string
	description      string
	kind             Kind
	parameterSchema  JsonSchemaObject
	isOutputMarkdown bool
	canUpdateOutput  bool
	MessageBus       interface{}
	serverName       string // Add serverName field
}

// NewBaseDeclarativeTool creates a new BaseDeclarativeTool.
func NewBaseDeclarativeTool(
	name string,
	displayName string,
	description string,
	kind Kind,
	parameterSchema JsonSchemaObject,
	isOutputMarkdown bool,
	canUpdateOutput bool,
	MessageBus interface{},
) *BaseDeclarativeTool {
	return &BaseDeclarativeTool{
		name:             name,
		displayName:      displayName,
		description:      description,
		kind:             kind,
		parameterSchema:  parameterSchema,
		isOutputMarkdown: isOutputMarkdown,
		canUpdateOutput:  canUpdateOutput,
		MessageBus:       MessageBus,
		serverName:       "", // Default to empty
	}
}

// Name returns the name of the tool.
func (bdt *BaseDeclarativeTool) Name() string {
	return bdt.name
}

// Description returns the description of the tool.
func (bdt *BaseDeclarativeTool) Description() string {
	return bdt.description
}

// ServerName returns the server name of the tool.
func (bdt *BaseDeclarativeTool) ServerName() string {
	return bdt.serverName
}

// Definition returns the genai.Tool definition for the Gemini API.
func (bdt *BaseDeclarativeTool) Definition() *genai.Tool {
	// Convert JsonSchemaObject to genai.Schema
	properties := make(map[string]*genai.Schema)
	for k, v := range bdt.parameterSchema.Properties {
		var propType genai.Type
		switch v.Type {
		case "string":
			propType = genai.TypeString
		case "number":
			propType = genai.TypeNumber
		case "integer":
			propType = genai.TypeInteger
		case "boolean":
			propType = genai.TypeBoolean
		case "array":
			propType = genai.TypeArray
		default:
			propType = genai.TypeString // Default to string
		}

		var itemsSchema *genai.Schema
		if v.Items != nil {
			var itemType genai.Type
			switch v.Items.Type {
			case "string":
				itemType = genai.TypeString
			case "number":
				itemType = genai.TypeNumber
			default:
				itemType = genai.TypeString // Default to string
			}
			itemsSchema = &genai.Schema{Type: itemType}
		}

		properties[k] = &genai.Schema{
			Type:        propType,
			Description: v.Description,
			Items:       itemsSchema,
		}
	}

	return &genai.Tool{
		FunctionDeclarations: []*genai.FunctionDeclaration{
			{
				Name:        bdt.name,
				Description: bdt.description,
				Parameters: &genai.Schema{
					Type:       genai.TypeObject,
					Properties: properties,
					Required:   bdt.parameterSchema.Required,
				},
			},
		},
	}
}

// Execute is a placeholder and should be implemented by concrete tool types.
func (bdt *BaseDeclarativeTool) Execute(args map[string]any) (ToolResult, error) {
	return ToolResult{}, fmt.Errorf("Execute method not implemented for BaseDeclarativeTool")
}

// ToolRegistry manages the registration and retrieval of tools.
type ToolRegistry struct {
	mu    sync.RWMutex
	tools map[string]Tool
}

// FileService defines an interface for file system operations.
type FileService interface {
	// ShouldIgnoreFile checks if a file should be ignored based on filtering options.
	ShouldIgnoreFile(filePath string, options FileFilteringOptions) bool
}

// GeminiDirProvider provides the path to the .gemini directory.
type GeminiDirProvider interface {
	GetGeminiDir() string
}

// GeminiConfigProvider provides configuration for Gemini.
type GeminiConfigProvider interface {
	GetModel() string
	GetToolRegistry() *ToolRegistry
}

// NewToolRegistry creates a new instance of ToolRegistry.
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]Tool),
	}
}

// Register adds a tool to the registry.
func (r *ToolRegistry) Register(t Tool) error {
	if _, exists := r.tools[t.Name()]; exists {
		return fmt.Errorf("tool with name '%s' already registered", t.Name())
	}
	r.tools[t.Name()] = t
	return nil
}

// GetTool retrieves a tool by its name.
func (r *ToolRegistry) GetTool(name string) (Tool, error) {
	t, exists := r.tools[name]
	if !exists {
		return nil, fmt.Errorf("no tool found with name '%s'", name)
	}
	return t, nil
}

// GetTools returns all registered tools as a slice of genai.Tool.
func (r *ToolRegistry) GetTools() []*genai.Tool {
	var toolDefs []*genai.Tool
	for _, t := range r.tools {
		toolDefs = append(toolDefs, t.Definition())
	}
	return toolDefs
}

// GetAllTools returns all registered tools as a slice of Tool.
func (r *ToolRegistry) GetAllTools() []Tool {
	var registeredTools []Tool
	for _, t := range r.tools {
		registeredTools = append(registeredTools, t)
	}
	return registeredTools
}

// GetAllToolNames returns a slice of all registered tool names.
func (tr *ToolRegistry) GetAllToolNames() []string {
	names := make([]string, 0, len(tr.tools))
	for name := range tr.tools {
		names = append(names, name)
	}
	return names
}

// GetFunctionDeclarationsFiltered returns FunctionDeclarations for a given list of tool names.
func (tr *ToolRegistry) GetFunctionDeclarationsFiltered(toolNames []string) []genai.FunctionDeclaration {
	var declarations []genai.FunctionDeclaration
	for _, name := range toolNames {
		if t, ok := tr.tools[name]; ok {
			if t.Definition() != nil && len(t.Definition().FunctionDeclarations) > 0 {
				declarations = append(declarations, *t.Definition().FunctionDeclarations[0])
			}
		}
	}
	return declarations
}

// Config is an interface that represents the application configuration.
type Config interface {
	GetCodebaseInvestigatorSettings() *CodebaseInvestigatorSettings
	GetDebugMode() bool
	GetToolRegistry() *ToolRegistry
	Model() string
}

// CodebaseInvestigatorSettings represents settings for the Codebase Investigator agent.
type CodebaseInvestigatorSettings struct {
	Enabled        bool   `json:"enabled,omitempty"`
	Model          string `json:"model,omitempty"`
	ThinkingBudget *int   `json:"thinkingBudget,omitempty"`
	MaxTimeMinutes *int   `json:"maxTimeMinutes,omitempty"`
	MaxNumTurns    *int   `json:"maxNumTurns,omitempty"`
}

// AgentStartEvent is a telemetry event.
type AgentStartEvent struct {
	AgentID   string
	AgentName string
}

// AgentFinishEvent is a telemetry event.
type AgentFinishEvent struct {
	AgentID         string
	AgentName       string
	DurationMs      int64
	TurnCounter     int
	TerminateReason AgentTerminateMode
}

// FolderStructureOptions for customizing folder structure retrieval.
type FolderStructureOptions struct {
	MaxItems           *int      // Maximum number of files and folders combined to display. Defaults to 200.
	IgnoredFolders     *[]string // Set of folder names to ignore completely. Case-sensitive.
	FileIncludePattern *string   // Optional regex to filter included files by name.
	// FileService        FileDiscoveryService // For filtering files.
	FileFilteringOptions *FileFilteringOptions // File filtering ignore options.
}

// FullFolderInfo represents the full, unfiltered information about a folder and its contents.
type FullFolderInfo struct {
	Name              string
	Path              string
	Files             []string
	SubFolders        []FullFolderInfo
	TotalChildren     int
	TotalFiles        int
	IsIgnored         bool // Flag to easily identify ignored folders later
	HasMoreFiles      bool // Indicates if files were truncated for this specific folder
	HasMoreSubfolders bool // Indicates if subfolders were truncated for this specific folder
}

// TelemetrySettings represents the telemetry settings.
type TelemetrySettings struct {
	Enabled      bool   `json:"enabled,omitempty"`
	Target       string `json:"target,omitempty"`
	OtlpEndpoint string `json:"otlpEndpoint,omitempty"`
	OtlpProtocol string `json:"otlpProtocol,omitempty"`
	LogPrompts   bool   `json:"logPrompts,omitempty"`
	Outfile      string `json:"outfile,omitempty"`
	UseCollector bool   `json:"useCollector,omitempty"`
}
