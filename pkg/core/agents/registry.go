package agents

import (
	"go-ai-agent-v2/go-cli/pkg/telemetry"
	"go-ai-agent-v2/go-cli/pkg/types"
)

// AgentRegistry manages the discovery, loading, validation, and registration of AgentDefinitions.
type AgentRegistry struct {
	agents map[string]AgentDefinition
	config types.Config
}

// NewAgentRegistry creates a new instance of AgentRegistry.
func NewAgentRegistry(cfg types.Config) *AgentRegistry {
	return &AgentRegistry{
		agents: make(map[string]AgentDefinition),
		config: cfg,
	}
}

// SetConfig updates the configuration for the AgentRegistry.
func (ar *AgentRegistry) SetConfig(cfg types.Config) {
	ar.config = cfg
}

// Initialize discovers and loads agents.
func (ar *AgentRegistry) Initialize() {
	ar.loadBuiltInAgents()

	debugModeVal, found := ar.config.Get("debugMode")
	if found && debugModeVal != nil {
		if debugMode, ok := debugModeVal.(bool); ok && debugMode {
			// debugLogger.Debug("Debug mode is enabled.")
		}
	}
}

// loadBuiltInAgents loads built-in agents.
func (ar *AgentRegistry) loadBuiltInAgents() {
	investigatorSettingsVal, found := ar.config.Get("codebaseInvestigatorSettings")
	var investigatorSettings *types.CodebaseInvestigatorSettings
	if found && investigatorSettingsVal != nil {
		if is, ok := investigatorSettingsVal.(*types.CodebaseInvestigatorSettings); ok {
			investigatorSettings = is
		}
	}

	// Only register the agent if it's enabled in the settings.
	if investigatorSettings != nil && investigatorSettings.Enabled {
		agentDef := CodebaseInvestigatorAgent // Start with the default definition

		// Override model config if settings are provided
		if investigatorSettings.Model != "" {
			agentDef.ModelConfig.Model = investigatorSettings.Model
		}
		if investigatorSettings.ThinkingBudget != nil {
			agentDef.ModelConfig.ThinkingBudget = *investigatorSettings.ThinkingBudget
		}

		// Override run config if settings are provided
		if investigatorSettings.MaxTimeMinutes != nil {
			agentDef.RunConfig.MaxTimeMinutes = *investigatorSettings.MaxTimeMinutes
		}
		if investigatorSettings.MaxNumTurns != nil {
			agentDef.RunConfig.MaxTurns = *investigatorSettings.MaxNumTurns
		}

		ar.registerAgent(agentDef)
	}

	testWriterSettingsVal, found := ar.config.Get("testWriterSettings")
	var testWriterSettings *types.TestWriterSettings
	if found && testWriterSettingsVal != nil {
		if tws, ok := testWriterSettingsVal.(*types.TestWriterSettings); ok {
			testWriterSettings = tws
		}
	}

	if testWriterSettings != nil && testWriterSettings.Enabled {
		agentDef := TestWriterAgent

		if testWriterSettings.Model != "" {
			agentDef.ModelConfig.Model = testWriterSettings.Model
		}
		if testWriterSettings.ThinkingBudget != nil {
			agentDef.ModelConfig.ThinkingBudget = *testWriterSettings.ThinkingBudget
		}
		if testWriterSettings.MaxTimeMinutes != nil {
			agentDef.RunConfig.MaxTimeMinutes = *testWriterSettings.MaxTimeMinutes
		}
		if testWriterSettings.MaxNumTurns != nil {
			agentDef.RunConfig.MaxTurns = *testWriterSettings.MaxNumTurns
		}

		ar.registerAgent(agentDef)
	}
}

// registerAgent registers an agent definition.
func (ar *AgentRegistry) registerAgent(definition AgentDefinition) {
	// Basic validation
	if definition.Name == "" || definition.Description == "" {
		telemetry.LogErrorf("Agent definition missing name or description.") // Changed to LogErrorf
		return
	}

	debugModeVal, found := ar.config.Get("debugMode")
	debugMode := false
	if found && debugModeVal != nil {
		if dm, ok := debugModeVal.(bool); ok {
			debugMode = dm
		}
	}

	if _, exists := ar.agents[definition.Name]; exists && debugMode {
		telemetry.LogDebugf("Overriding existing agent: %s", definition.Name)
	}

	ar.agents[definition.Name] = definition
}

// GetDefinition retrieves an agent definition by name.
func (ar *AgentRegistry) GetDefinition(name string) (AgentDefinition, bool) {
	agent, exists := ar.agents[name]
	return agent, exists
}

// GetAllDefinitions returns all active agent definitions.
func (ar *AgentRegistry) GetAllDefinitions() []AgentDefinition {
	definitions := make([]AgentDefinition, 0, len(ar.agents))
	for _, agent := range ar.agents {
		definitions = append(definitions, agent)
	}
	return definitions
}
