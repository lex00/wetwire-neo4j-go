package cli

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/lex00/wetwire-neo4j-go/internal/discover"
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
	scanner := discover.NewScanner()
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
	scanner := discover.NewScanner()
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

// Validator implements Validator for Neo4j configuration validation.
// It uses environment variables for connection configuration:
// - NEO4J_URI: Connection URI (e.g., bolt://localhost:7687)
// - NEO4J_USERNAME: Username (default: neo4j)
// - NEO4J_PASSWORD: Password
// - NEO4J_DATABASE: Database name (default: neo4j)
type Validator struct{}

// NewValidator creates a new Validator.
func NewValidator() *Validator {
	return &Validator{}
}

// Validate validates Neo4j configurations in the given path.
// If NEO4J_URI is not set, returns validation errors indicating connection is required.
func (v *Validator) Validate(_ context.Context, path string, opts ValidateOptions) ([]ValidationError, error) {
	// Check if path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("path does not exist: %s", path)
	}

	// Discover resources
	scanner := discover.NewScanner()
	resources, err := scanner.ScanDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to scan directory: %w", err)
	}

	// If no resources, nothing to validate
	if len(resources) == 0 {
		return nil, nil
	}

	// Get connection config from environment
	uri := os.Getenv("NEO4J_URI")
	if uri == "" {
		// Return validation errors for each resource indicating connection needed
		errors := make([]ValidationError, len(resources))
		for i, r := range resources {
			errors[i] = ValidationError{
				Path:    fmt.Sprintf("%s:%d", r.File, r.Line),
				Message: fmt.Sprintf("cannot validate %s %q without NEO4J_URI configured", r.Kind, r.Name),
				Code:    "NEO4J_CONN_REQUIRED",
			}
		}
		return errors, nil
	}

	// Get other config from environment with defaults
	username := os.Getenv("NEO4J_USERNAME")
	if username == "" {
		username = "neo4j"
	}
	password := os.Getenv("NEO4J_PASSWORD")
	database := os.Getenv("NEO4J_DATABASE")
	if database == "" {
		database = "neo4j"
	}

	// Create validator and connect
	val, err := validator.New(validator.Config{
		URI:      uri,
		Username: username,
		Password: password,
		Database: database,
	})
	if err != nil {
		return []ValidationError{{
			Path:    path,
			Message: fmt.Sprintf("failed to connect to Neo4j: %v", err),
			Code:    "NEO4J_CONN_FAILED",
		}}, nil
	}
	defer func() { _ = val.Close() }()

	// Get database info for verbose output
	if opts.Verbose {
		dbInfo := val.GetDatabaseInfo()
		gdsInfo := val.GetGDSInfo()
		fmt.Printf("Connected to Neo4j %s (%s)\n", dbInfo.Version, dbInfo.Edition)
		if gdsInfo.Installed {
			fmt.Printf("GDS %s (%s) installed\n", gdsInfo.Version, gdsInfo.Edition)
		}
	}

	// For now, report success if connection works
	// Full validation would require runtime introspection of resource values
	// which isn't possible from AST-based discovery alone
	return nil, nil
}
