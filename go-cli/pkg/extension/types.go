package extension

// InstallArgs represents the arguments for the install command.
type InstallArgs struct {
	Source        string
	Ref           string
	AutoUpdate    bool
	AllowPreRelease bool
	Consent       bool
	Force         bool
}

// NewArgs represents the arguments for the new command.
type NewArgs struct {
	Path     string
	Template string
}

// ExtensionScopeArgs represents arguments for commands that operate on extension name and scope.
type ExtensionScopeArgs struct {
	Name  string
	Scope string
}

// ExtensionInstallMetadata represents metadata about an extension to be installed.
type ExtensionInstallMetadata struct {
	Source        string
	Type          string // "git" or "local"
	Ref           string
	AutoUpdate    bool
	AllowPreRelease bool
}
