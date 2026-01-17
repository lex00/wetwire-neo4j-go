package cli

// BuildOptions contains options for building resources.
type BuildOptions struct {
	Output  string
	DryRun  bool
	Verbose bool
}

// LintOptions contains options for linting resources.
type LintOptions struct {
	Verbose bool
}

// InitOptions contains options for initializing a project.
type InitOptions struct {
	Template string
	Force    bool
}

// ValidationError represents a validation error.
type ValidationError struct {
	Path     string
	Line     int
	Message  string
	Severity string
	Code     string
}

// ValidateOptions contains options for validation.
type ValidateOptions struct {
	DryRun  bool
	Verbose bool
}

// Issue represents a lint issue.
type Issue struct {
	File     string
	Line     int
	Column   int
	Severity string
	Message  string
	Rule     string
}
