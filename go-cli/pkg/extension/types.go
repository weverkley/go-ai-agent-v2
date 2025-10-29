package extension

// InstallArgs represents the arguments for the install command.
type InstallArgs struct {
	Source        string
	Ref           string
	AutoUpdate    bool
	AllowPreRelease bool
	Consent       bool
}

// ExtensionInstallMetadata represents metadata about an extension to be installed.
type ExtensionInstallMetadata struct {
	Source        string
	Type          string // "git" or "local"
	Ref           string
	AutoUpdate    bool
	AllowPreRelease bool
}
