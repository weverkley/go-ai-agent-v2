package agents

import (
	"fmt"
)

// convertInputConfigToJsonSchema converts an internal InputConfig definition into a standard JSON Schema
// object suitable for a tool's FunctionDeclaration.
func convertInputConfigToJsonSchema(inputConfig InputConfig) (JsonSchemaObject, error) {
	properties := make(map[string]JsonSchemaProperty)
	required := []string{}

	for name, definition := range inputConfig.Inputs {
		schemaProperty := JsonSchemaProperty{
			Description: definition.Description,
		}

		switch definition.Type {
		case "string", "number", "integer", "boolean":
			schemaProperty.Type = definition.Type
		case "string[]":
			schemaProperty.Type = "array"
			schemaProperty.Items = &struct{ Type string }{Type: "string"}
		case "number[]":
			schemaProperty.Type = "array"
			schemaProperty.Items = &struct{ Type string }{Type: "number"}
		default:
			return JsonSchemaObject{}, fmt.Errorf("unsupported input type '%s' for parameter '%s'. Supported types: string, number, integer, boolean, string[], number[]", definition.Type, name)
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

	return JsonSchemaObject{
		Type:       "object",
		Properties: properties,
		Required:   requiredPtr,
	}, nil
}
