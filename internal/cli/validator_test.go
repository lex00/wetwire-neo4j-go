package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidatorCLI_ValidateFromPath(t *testing.T) {
	// Create temp directory with test definitions
	tmpDir := t.TempDir()

	// Create a simple schema file
	schemaContent := `package schema

import "github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"

var Person = &schema.NodeType{
	Label: "Person",
	Properties: []schema.Property{
		{Name: "id", Type: schema.TypeString, Required: true},
	},
}
`
	err := os.WriteFile(filepath.Join(tmpDir, "schema.go"), []byte(schemaContent), 0644)
	if err != nil {
		t.Fatalf("failed to write schema file: %v", err)
	}

	validator := NewValidatorCLI()

	// Test without connection (should return connection required message)
	var buf bytes.Buffer
	err = validator.ValidateFromPath(tmpDir, "", &buf)
	if err == nil {
		t.Error("expected error when no URI provided")
	}
	if !strings.Contains(err.Error(), "URI") {
		t.Errorf("expected URI error, got: %v", err)
	}
}

func TestValidatorCLI_ValidateDryRun(t *testing.T) {
	// Create temp directory with test definitions
	tmpDir := t.TempDir()

	// Create a simple schema file
	schemaContent := `package schema

import "github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"

var Person = &schema.NodeType{
	Label: "Person",
}
`
	err := os.WriteFile(filepath.Join(tmpDir, "schema.go"), []byte(schemaContent), 0644)
	if err != nil {
		t.Fatalf("failed to write schema file: %v", err)
	}

	validator := NewValidatorCLI()

	// Test dry run (should list discovered resources without connecting)
	var buf bytes.Buffer
	err = validator.ValidateDryRun(tmpDir, &buf)
	if err != nil {
		t.Errorf("dry run failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Person") {
		t.Errorf("expected output to contain 'Person', got: %s", output)
	}
}

func TestValidatorCLI_FormatConfig(t *testing.T) {
	validator := NewValidatorCLI()

	config := validator.ParseConfig("bolt://localhost:7687", "neo4j", "password", "neo4j")
	if config.URI != "bolt://localhost:7687" {
		t.Errorf("expected URI to be bolt://localhost:7687, got: %s", config.URI)
	}
	if config.Username != "neo4j" {
		t.Errorf("expected Username to be neo4j, got: %s", config.Username)
	}
	if config.Password != "password" {
		t.Errorf("expected Password to be password, got: %s", config.Password)
	}
	if config.Database != "neo4j" {
		t.Errorf("expected Database to be neo4j, got: %s", config.Database)
	}
}

func TestValidatorCLI_ParseConfigDefaults(t *testing.T) {
	validator := NewValidatorCLI()

	config := validator.ParseConfig("bolt://localhost:7687", "", "", "")
	if config.Database != "neo4j" {
		t.Errorf("expected default Database to be neo4j, got: %s", config.Database)
	}
}
