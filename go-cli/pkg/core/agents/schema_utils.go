package agents

import (
	"fmt"

	"go-ai-agent-v2/go-cli/pkg/types"
)

// convertInputConfigToJsonSchema converts an internal InputConfig definition into a standard JSON Schema
// object suitable for a tool's FunctionDeclaration.
func convertInputConfigToJsonSchema(inputConfig InputConfig) (types.JsonSchemaObject, error) {
	properties := make(map[string]types.JsonSchemaProperty)
	required := []string{}

	for name, definition := range inputConfig.Inputs {
		schemaProperty := types.JsonSchemaProperty{
			Description: definition.Description,
		}

		switch definition.Type {
		case "string", "number", "integer", "boolean":
			schemaProperty.Type = definition.Type
			schemaProperty.Items = &types.JsonSchemaPropertyItem{Type: "string"}
		case "number[]":
			schemaProperty.Items = &types.JsonSchemaPropertyItem{Type: "number"}
		default:
			return types.JsonSchemaObject{}, fmt.Errorf("unsupported input type '%s' for parameter '%s'. Supported types: string, number, integer, boolean, string[], number[]", definition.Type, name)
		}

		properties[name] = schemaProperty

		if definition.Required {
			required = append(required, name)
		}
	}

	var requiredPtr []string = nil
	if len(required) > 0 {
		requiredPtr = required
	}

	return types.JsonSchemaObject{
		Type:       "object",
		Properties: properties,
		Required:   requiredPtr,
	}, nil
}
