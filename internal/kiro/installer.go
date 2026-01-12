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
	if err := corekiro.Install(config); err != nil {
		return fmt.Errorf("install kiro config: %w", err)
	}
	return nil
}
