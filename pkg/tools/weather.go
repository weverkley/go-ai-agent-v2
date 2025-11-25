package tools

import (
	"context"
	"fmt"

	"go-ai-agent-v2/go-cli/pkg/types"
)

// WeatherTool implements the Tool interface for getting the weather.
type WeatherTool struct {
	*types.BaseDeclarativeTool
}

// NewWeatherTool creates a new WeatherTool.
func NewWeatherTool() *WeatherTool {
	return &WeatherTool{
		BaseDeclarativeTool: types.NewBaseDeclarativeTool(
			"get_weather",
			"Get Weather",
			"Gets the current weather for a given location.",
			types.KindOther,
			(&types.JsonSchemaObject{
				Type: "object",
			}).SetProperties(map[string]*types.JsonSchemaProperty{
				"location": {
					Type:        "string",
					Description: "The city for which to get the weather.",
				},
			}).SetRequired([]string{"location"}),
			false, // isOutputMarkdown
			false, // canUpdateOutput
			nil,   // MessageBus
		),
	}
}

// Execute performs the get_weather operation.
func (t *WeatherTool) Execute(ctx context.Context, args map[string]any) (types.ToolResult, error) {
	location, ok := args["location"].(string)
	if !ok || location == "" {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: "missing or invalid 'location' argument",
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("missing or invalid 'location' argument")
	}

	// Mocked weather data
	weather := fmt.Sprintf("The weather in %s is 28°C, Sunny, with a wind speed of 15 km/h.", location)

	return types.ToolResult{
		LLMContent:    weather,
		ReturnDisplay: weather,
	}, nil
}
