package tools

import (
	"os"
)

// These are the *mock functions* that test cases can set.
// They are assigned to the `os` package's global variables via setupOsMocks.
var (
	testMockUserHomeDir func() (string, error)
	testMockMkdirAll    func(path string, perm os.FileMode) error
	testMockReadFile    func(name string) ([]byte, error)
	testMockWriteFile   func(name string, data []byte, perm os.FileMode) error
	testMockIsNotExist  func(err error) bool
)

// setupOsMocks assigns the test-specific mock functions to the global os function variables in the `tools` package.
func setupOsMocks() {
	if testMockUserHomeDir != nil {
		osUserHomeDir = testMockUserHomeDir
	} else {
		osUserHomeDir = os.UserHomeDir
	}
	if testMockMkdirAll != nil {
		osMkdirAll = testMockMkdirAll
	} else {
		osMkdirAll = os.MkdirAll
	}
	if testMockReadFile != nil {
		osReadFile = testMockReadFile
	} else {
		osReadFile = os.ReadFile
	}
	if testMockWriteFile != nil {
		osWriteFile = testMockWriteFile
	} else {
		osWriteFile = os.WriteFile
	}
	if testMockIsNotExist != nil {
		osIsNotExist = testMockIsNotExist
	} else {
		osIsNotExist = os.IsNotExist
	}
}

// teardownOsMocks restores the original os functions and resets mock test functions.
func teardownOsMocks() {
	osUserHomeDir = os.UserHomeDir
	osMkdirAll = os.MkdirAll
	osReadFile = os.ReadFile
	osWriteFile = os.WriteFile
	osIsNotExist = os.IsNotExist

	testMockUserHomeDir = nil
	testMockMkdirAll = nil
	testMockReadFile = nil
	testMockWriteFile = nil
	testMockIsNotExist = nil
}