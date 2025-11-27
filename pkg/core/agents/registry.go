package agents

import (
	"fmt"
	"sync" // Import the sync package

	"go-ai-agent-v2/go-cli/pkg/telemetry"
	"go-ai-agent-v2/go-cli/pkg/types"
)

// AgentRegistry manages the discovery, loading, validation, and registration of AgentDefinitions.
type AgentRegistry struct {
	mu     sync.RWMutex // Add mutex for thread safety
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
	ar.mu.Lock()
	defer ar.mu.Unlock()
	ar.config = cfg
}

// Register registers an agent.
func (ar *AgentRegistry) Register(a types.Agent) error {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	// Attempt to cast the types.Agent to an agentDefinitionWrapper
	wrapper, ok := a.(*agentDefinitionWrapper)
	if !ok {
		return fmt.Errorf("cannot register agent: provided agent does not contain an AgentDefinition")
	}

	if _, exists := ar.agents[wrapper.Name()]; exists {
		return fmt.Errorf("agent with name '%s' already registered", wrapper.Name())
	}
	ar.agents[wrapper.Name()] = wrapper.AgentDefinition
	return nil
}

// GetAgent retrieves an agent by its name.
func (ar *AgentRegistry) GetAgent(name string) (types.Agent, error) {
	ar.mu.RLock()
	defer ar.mu.RUnlock()

	agentDef, exists := ar.agents[name]
	if !exists {
		return nil, fmt.Errorf("no agent found with name '%s'", name)
	}
	return NewAgentDefinitionWrapper(agentDef), nil
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

// GetAllAgentDefinitions returns all active agent definitions as a slice of interface{}.
func (ar *AgentRegistry) GetAllAgentDefinitions() []interface{} {
	ar.mu.RLock()
	defer ar.mu.RUnlock()
	definitions := make([]interface{}, 0, len(ar.agents))
	for _, agent := range ar.agents {
		definitions = append(definitions, agent)
	}
	return definitions
}

// GetAllAgents returns all registered agents as a slice of types.Agent.
func (ar *AgentRegistry) GetAllAgents() []types.Agent {
	ar.mu.RLock()
	defer ar.mu.RUnlock()
	var registeredAgents []types.Agent
	for _, agentDef := range ar.agents {
		registeredAgents = append(registeredAgents, NewAgentDefinitionWrapper(agentDef))
	}
	return registeredAgents
}

// GetAllAgentNames returns a slice of all registered agent names.
func (ar *AgentRegistry) GetAllAgentNames() []string {
	ar.mu.RLock()
	defer ar.mu.RUnlock()
	names := make([]string, 0, len(ar.agents))
	for name := range ar.agents {
		names = append(names, name)
	}
	return names
}
