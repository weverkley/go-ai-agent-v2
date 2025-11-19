package tools

import "os"

// These are global variables for os functions that can be overridden for testing purposes.
var (
	osUserHomeDir = os.UserHomeDir
	osMkdirAll    = os.MkdirAll
	osReadFile    = os.ReadFile
	osWriteFile   = os.WriteFile
	osIsNotExist  = os.IsNotExist
)
