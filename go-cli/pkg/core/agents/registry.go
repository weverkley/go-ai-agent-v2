package agents

import (
	"go-ai-agent-v2/go-cli/pkg/config"
)

// AgentRegistry manages the discovery, loading, validation, and registration of AgentDefinitions.
type AgentRegistry struct {
	agents map[string]AgentDefinition
	config *config.Config
}

// NewAgentRegistry creates a new instance of AgentRegistry.
func NewAgentRegistry(cfg *config.Config) *AgentRegistry {
	return &AgentRegistry{
		agents: make(map[string]AgentDefinition),
		config: cfg,
	}
}

// Initialize discovers and loads agents.
func (ar *AgentRegistry) Initialize() {
	ar.loadBuiltInAgents()

	if ar.config.GetDebugMode() {
		// debugLogger.log(fmt.Sprintf("[AgentRegistry] Initialized with %d agents.", len(ar.agents))) // TODO: Implement debugLogger
	}
}

// loadBuiltInAgents loads built-in agents.
func (ar *AgentRegistry) loadBuiltInAgents() {
	investigatorSettings := ar.config.GetCodebaseInvestigatorSettings()

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
}

// registerAgent registers an agent definition.
func (ar *AgentRegistry) registerAgent(definition AgentDefinition) {
	// Basic validation
	if definition.Name == "" || definition.Description == "" {
		// debugLogger.warn(fmt.Sprintf("[AgentRegistry] Skipping invalid agent definition. Missing name or description.")) // TODO: Implement debugLogger
		return
	}

	if _, exists := ar.agents[definition.Name]; exists && ar.config.GetDebugMode() {
		// debugLogger.log(fmt.Sprintf("[AgentRegistry] Overriding agent '%s'", definition.Name)) // TODO: Implement debugLogger
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
