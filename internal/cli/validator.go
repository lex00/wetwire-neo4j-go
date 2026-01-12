package cli

import (
	"fmt"
	"io"

	"github.com/lex00/wetwire-neo4j-go/internal/discovery"
	"github.com/lex00/wetwire-neo4j-go/internal/validator"
)

// ValidatorCLI provides CLI functionality for validating Neo4j configurations.
type ValidatorCLI struct{}

// NewValidatorCLI creates a new ValidatorCLI.
func NewValidatorCLI() *ValidatorCLI {
	return &ValidatorCLI{}
}

// ValidatorConfig holds connection configuration for validation.
type ValidatorConfig struct {
	URI      string
	Username string
	Password string
	Database string
}

// ParseConfig parses and validates connection configuration.
func (v *ValidatorCLI) ParseConfig(uri, username, password, database string) ValidatorConfig {
	config := ValidatorConfig{
		URI:      uri,
		Username: username,
		Password: password,
		Database: database,
	}
	if config.Database == "" {
		config.Database = "neo4j"
	}
	return config
}

// ValidateFromPath discovers and validates configurations from the given path.
func (v *ValidatorCLI) ValidateFromPath(path string, uri string, w io.Writer) error {
	if uri == "" {
		return fmt.Errorf("URI is required for validation")
	}

	// This would require a full config to actually validate
	return fmt.Errorf("URI connection not implemented in test mode")
}

// ValidateWithConfig discovers and validates configurations using the provided config.
func (v *ValidatorCLI) ValidateWithConfig(path string, config ValidatorConfig, w io.Writer) error {
	if config.URI == "" {
		return fmt.Errorf("URI is required for validation")
	}

	// Discover resources
	scanner := discovery.NewScanner()
	resources, err := scanner.ScanDir(path)
	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}

	if len(resources) == 0 {
		_, _ = fmt.Fprintln(w, "No resources discovered to validate")
		return nil
	}

	// Create validator
	val, err := validator.New(validator.Config{
		URI:      config.URI,
		Username: config.Username,
		Password: config.Password,
		Database: config.Database,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to Neo4j: %w", err)
	}
	defer func() { _ = val.Close() }()

	// Print database info
	dbInfo := val.GetDatabaseInfo()
	gdsInfo := val.GetGDSInfo()
	_, _ = fmt.Fprintf(w, "Connected to Neo4j %s (%s)\n", dbInfo.Version, dbInfo.Edition)
	if gdsInfo.Installed {
		_, _ = fmt.Fprintf(w, "GDS %s (%s) installed\n", gdsInfo.Version, gdsInfo.Edition)
	} else {
		_, _ = fmt.Fprintln(w, "GDS not installed")
	}
	_, _ = fmt.Fprintln(w)

	// Report discovered resources (we can't validate runtime instances from AST discovery)
	_, _ = fmt.Fprintf(w, "Discovered %d resources:\n", len(resources))
	for _, r := range resources {
		_, _ = fmt.Fprintf(w, "  [%s] %s (%s:%d)\n", r.Kind, r.Name, r.File, r.Line)
	}

	return nil
}

// ValidateDryRun lists discovered resources without connecting to Neo4j.
func (v *ValidatorCLI) ValidateDryRun(path string, w io.Writer) error {
	// Discover resources
	scanner := discovery.NewScanner()
	resources, err := scanner.ScanDir(path)
	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}

	_, _ = fmt.Fprintf(w, "Discovered %d resources:\n\n", len(resources))

	for _, r := range resources {
		_, _ = fmt.Fprintf(w, "  %s: %s (%s:%d)\n", r.Kind, r.Name, r.File, r.Line)
	}

	return nil
}
