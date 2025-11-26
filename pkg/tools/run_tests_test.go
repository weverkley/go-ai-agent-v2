package tools

import (

	"context"

	"fmt"

	"testing"



	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/mock"

)



func TestRunTestsTool_Execute(t *testing.T) {
	mockShellService := new(MockShellExecutionService)
	mockFsService := new(MockFileSystemService)
	mockWorkspace := new(MockWorkspaceService)
	mockWorkspace.On("GetProjectRoot").Return("") // Project root is joined, can be empty for this test.
	tool := NewRunTestsTool(mockShellService, mockFsService, mockWorkspace)

	tests := []struct {
		name          string
		args          map[string]any
		setupMocks    func()
		expectedCmd   string
		expectedError string
	}{
		{
			name: "Go project - all tests",
			args: map[string]any{"dir": "/go/project"},
			setupMocks: func() {
				mockFsService.On("PathExists", "/go/project/go.mod").Return(true, nil).Once()
				mockFsService.On("PathExists", "/go/project/package.json").Return(false, nil).Maybe()
				mockFsService.On("PathExists", "/go/project/requirements.txt").Return(false, nil).Maybe()
				mockShellService.On("ExecuteCommand", mock.Anything, "go test ./...", "/go/project").Return("PASS", "", nil).Once()
			},
			expectedCmd: "go test ./...",
		},
		{
			name: "Go project - with target and coverage",
			args: map[string]any{"dir": "/go/project", "target": "./mypackage", "coverage": true},
			setupMocks: func() {
				mockFsService.On("PathExists", "/go/project/go.mod").Return(true, nil).Once()
				mockFsService.On("PathExists", "/go/project/package.json").Return(false, nil).Maybe()
				mockFsService.On("PathExists", "/go/project/requirements.txt").Return(false, nil).Maybe()
				mockShellService.On("ExecuteCommand", mock.Anything, "go test ./mypackage -cover", "/go/project").Return("PASS", "", nil).Once()
			},
			expectedCmd: "go test ./mypackage -cover",
		},
		{
			name: "Node project - all tests with coverage",
			args: map[string]any{"dir": "/node/project", "coverage": true},
			setupMocks: func() {
				mockFsService.On("PathExists", "/node/project/go.mod").Return(false, nil).Once()
				mockFsService.On("PathExists", "/node/project/package.json").Return(true, nil).Once()
				mockFsService.On("PathExists", "/node/project/requirements.txt").Return(false, nil).Maybe()
				mockShellService.On("ExecuteCommand", mock.Anything, "npm test -- --coverage", "/node/project").Return("Tests passed", "", nil).Once()
			},
			expectedCmd: "npm test -- --coverage",
		},
		{
			name: "Python project - with target",
			args: map[string]any{"dir": "/python/project", "target": "tests/test_api.py"},
			setupMocks: func() {
				mockFsService.On("PathExists", "/python/project/go.mod").Return(false, nil).Once()
				mockFsService.On("PathExists", "/python/project/package.json").Return(false, nil).Once()
				mockFsService.On("PathExists", "/python/project/requirements.txt").Return(true, nil).Once()
				mockShellService.On("ExecuteCommand", mock.Anything, "pytest tests/test_api.py", "/python/project").Return("1 passed", "", nil).Once()
			},
			expectedCmd: "pytest tests/test_api.py",
		},
		{
			name: "Unknown project type",
			args: map[string]any{"dir": "/unknown/project"},
			setupMocks: func() {
				mockFsService.On("PathExists", "/unknown/project/go.mod").Return(false, nil).Once()
				mockFsService.On("PathExists", "/unknown/project/package.json").Return(false, nil).Once()
				mockFsService.On("PathExists", "/unknown/project/requirements.txt").Return(false, nil).Once()
			},
			expectedError: "could not determine project type",
		},
		{
			name: "Test command fails",
			args: map[string]any{"dir": "/go/project"},
			setupMocks: func() {
				mockFsService.On("PathExists", "/go/project/go.mod").Return(true, nil).Once()
				mockFsService.On("PathExists", "/go/project/package.json").Return(false, nil).Maybe()
				mockFsService.On("PathExists", "/go/project/requirements.txt").Return(false, nil).Maybe()
				mockShellService.On("ExecuteCommand", mock.Anything, "go test ./...", "/go/project").Return("FAIL", "compilation error", fmt.Errorf("exit status 1")).Once()
			},
			expectedCmd:   "go test ./...",
			expectedError: "exit status 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks for each test
			mockShellService.Calls = nil
			mockFsService.Calls = nil
			tt.setupMocks()

			_, err := tool.Execute(context.Background(), tt.args)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			// Verify that the mocks were called as expected
			mockFsService.AssertExpectations(t)
			mockShellService.AssertExpectations(t)
		})
	}
}
