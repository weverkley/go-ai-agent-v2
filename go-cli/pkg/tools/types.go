package tools

import (
	"context"
	"fmt"

	"go-ai-agent-v2/go-cli/pkg/types" // Added

	"github.com/google/generative-ai-go/genai"
)

// Kind represents the type of tool.
type Kind string

const (
	KindThink Kind = "THINK"
	// Add other kinds as needed
)

// ToolInvocation represents an executable instance of a tool.
type ToolInvocation interface {
	Execute(ctx context.Context, updateOutput func(output string), shellExecutionConfig interface{}, setPidCallback func(int)) (types.ToolResult, error)
	ShouldConfirmExecute(ctx context.Context) (types.ToolCallConfirmationDetails, error)
	GetDescription() string
}

// BaseDeclarativeTool provides a base implementation for declarative tools.
type BaseDeclarativeTool struct {
	name            string
	displayName     string
	description     string
	kind            Kind
	parameterSchema types.JsonSchemaObject // Changed
	isOutputMarkdown bool
	canUpdateOutput bool
	messageBus      interface{} // Placeholder for MessageBus
}

// NewBaseDeclarativeTool creates a new BaseDeclarativeTool.
func NewBaseDeclarativeTool(
	name string,
	displayName string,
	description string,
	kind Kind,
	parameterSchema agents.JsonSchemaObject,
	isOutputMarkdown bool,
	canUpdateOutput bool,
	messageBus interface{},
) *BaseDeclarativeTool {
	return &BaseDeclarativeTool{
		name:            name,
		displayName:     displayName,
		description:     description,
		kind:            kind,
		parameterSchema: parameterSchema,
		isOutputMarkdown: isOutputMarkdown,
		canUpdateOutput: canUpdateOutput,
		messageBus:      messageBus,
	}
}

// Name returns the name of the tool.
func (bdt *BaseDeclarativeTool) Name() string {
	return bdt.name
}

// Definition returns the genai.Tool definition for the Gemini API.
func (bdt *BaseDeclarativeTool) Definition() *genai.Tool {
	// Convert JsonSchemaObject to genai.Schema
	properties := make(map[string]*genai.Schema)
	for k, v := range bdt.parameterSchema.Properties {
		propType := genai.Type(v.Type)
		var itemsSchema *genai.Schema
		if v.Items != nil {
			itemsSchema = &genai.Schema{Type: genai.Type(v.Items.Type)}
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
func (bdt *BaseDeclarativeTool) Execute(args map[string]any) (types.ToolResult, error) {
	return types.ToolResult{}, fmt.Errorf("Execute method not implemented for BaseDeclarativeTool")
}

// Tool is the interface that all tools must implement.
type Tool interface {
	// Name returns the name of the tool.
	Name() string
	// Definition returns the genai.Tool definition for the Gemini API.
	Definition() *genai.Tool
	// Execute runs the tool with the given arguments.
	Execute(args map[string]any) (types.ToolResult, error)
}
