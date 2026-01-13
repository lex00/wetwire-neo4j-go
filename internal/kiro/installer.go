package kiro

import (
	"fmt"

	corekiro "github.com/lex00/wetwire-core-go/kiro"
)

// EnsureInstalled installs the Kiro agent configuration if not already present.
func EnsureInstalled() error {
	return EnsureInstalledWithForce(false)
}

// EnsureInstalledWithForce installs the Kiro agent configuration.
// If force is true, overwrites any existing configuration.
func EnsureInstalledWithForce(force bool) error {
	config := NewConfig()
	return InstallConfig(config)
}

// InstallConfig installs a specific Kiro configuration.
// This allows installing configurations with custom schema context.
func InstallConfig(config corekiro.Config) error {
	if err := corekiro.Install(config); err != nil {
		return fmt.Errorf("install kiro config: %w", err)
	}
	return nil
}
