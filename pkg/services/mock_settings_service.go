package services

import (
	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/stretchr/testify/mock"
)

// MockSettingsService is a mock implementation of the SettingsServiceIface.
type MockSettingsService struct {
	mock.Mock
}

// Get provides a mock function for Get.
func (m *MockSettingsService) Get(key string) (interface{}, bool) {
	args := m.Called(key)
	return args.Get(0), args.Bool(1)
}

// GetTelemetrySettings provides a mock function for GetTelemetrySettings.
func (m *MockSettingsService) GetTelemetrySettings() *types.TelemetrySettings {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*types.TelemetrySettings)
}

// GetGoogleCustomSearchSettings provides a mock function for GetGoogleCustomSearchSettings.
func (m *MockSettingsService) GetGoogleCustomSearchSettings() *types.GoogleCustomSearchSettings {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*types.GoogleCustomSearchSettings)
}

// GetWebSearchProvider provides a mock function for GetWebSearchProvider.
func (m *MockSettingsService) GetWebSearchProvider() types.WebSearchProvider {
	args := m.Called()
	return args.Get(0).(types.WebSearchProvider)
}

// GetTavilySettings provides a mock function for GetTavilySettings.
func (m *MockSettingsService) GetTavilySettings() *types.TavilySettings {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*types.TavilySettings)
}

// GetDangerousTools provides a mock function for GetDangerousTools.
func (m *MockSettingsService) GetDangerousTools() []string {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]string)
}

// GetWorkspaceDir provides a mock function for GetWorkspaceDir.
func (m *MockSettingsService) GetWorkspaceDir() string {
	args := m.Called()
	return args.String(0)
}

// GetTestWriterSettings provides a mock function for GetTestWriterSettings.
func (m *MockSettingsService) GetTestWriterSettings() *types.TestWriterSettings {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*types.TestWriterSettings)
}

// GetCodebaseInvestigatorSettings provides a mock function for GetCodebaseInvestigatorSettings.
func (m *MockSettingsService) GetCodebaseInvestigatorSettings() *types.CodebaseInvestigatorSettings {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*types.CodebaseInvestigatorSettings)
}

// Set provides a mock function for Set.
func (m *MockSettingsService) Set(key string, value interface{}) error {
	args := m.Called(key, value)
	return args.Error(0)
}

// AllSettings provides a mock function for AllSettings.
func (m *MockSettingsService) AllSettings() map[string]interface{} {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(map[string]interface{})
}

// Reset provides a mock function for Reset.
func (m *MockSettingsService) Reset() error {
	args := m.Called()
	return args.Error(0)
}

// Save provides a mock function for Save.
func (m *MockSettingsService) Save() error {
	args := m.Called()
	return args.Error(0)
}

// SetExtensionManager provides a mock function for SetExtensionManager.
func (m *MockSettingsService) SetExtensionManager(mgr types.ExtensionManager) {
	m.Called(mgr)
}
