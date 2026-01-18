package domain

import (
	"os"
	"path/filepath"
	"testing"

	coredomain "github.com/lex00/wetwire-core-go/domain"
)

// TestDomainInterface verifies that Neo4jDomain implements the Domain interface.
func TestDomainInterface(t *testing.T) {
	var _ coredomain.Domain = (*Neo4jDomain)(nil)
}

// TestListerDomainInterface verifies that Neo4jDomain implements the ListerDomain interface.
func TestListerDomainInterface(t *testing.T) {
	var _ coredomain.ListerDomain = (*Neo4jDomain)(nil)
}

// TestGrapherDomainInterface verifies that Neo4jDomain implements the GrapherDomain interface.
func TestGrapherDomainInterface(t *testing.T) {
	var _ coredomain.GrapherDomain = (*Neo4jDomain)(nil)
}

// TestDomainMetadata verifies that the domain returns correct metadata.
func TestDomainMetadata(t *testing.T) {
	d := &Neo4jDomain{}

	if d.Name() != "neo4j" {
		t.Errorf("expected name 'neo4j', got '%s'", d.Name())
	}

	if d.Version() == "" {
		t.Error("expected non-empty version")
	}
}

// TestDomainOperations verifies that all domain operations return non-nil implementations.
func TestDomainOperations(t *testing.T) {
	d := &Neo4jDomain{}

	if d.Builder() == nil {
		t.Error("Builder() returned nil")
	}

	if d.Linter() == nil {
		t.Error("Linter() returned nil")
	}

	if d.Initializer() == nil {
		t.Error("Initializer() returned nil")
	}

	if d.Validator() == nil {
		t.Error("Validator() returned nil")
	}

	if d.Lister() == nil {
		t.Error("Lister() returned nil")
	}

	if d.Grapher() == nil {
		t.Error("Grapher() returned nil")
	}
}

// TestLintOpts_Fields tests that LintOpts fields are correctly defined
func TestLintOpts_Fields(t *testing.T) {
	opts := LintOpts{
		Format:  "text",
		Fix:     true,
		Disable: []string{"WN4001", "WN4052"},
	}

	if opts.Format != "text" {
		t.Errorf("expected format 'text', got '%s'", opts.Format)
	}
	if !opts.Fix {
		t.Error("expected Fix to be true")
	}
	if len(opts.Disable) != 2 {
		t.Errorf("expected 2 disabled rules, got %d", len(opts.Disable))
	}
	if opts.Disable[0] != "WN4001" || opts.Disable[1] != "WN4052" {
		t.Errorf("unexpected disable values: %v", opts.Disable)
	}
}

// TestNeo4jLinter_Lint_Disable tests that disabled rules are skipped
func TestNeo4jLinter_Lint_Disable(t *testing.T) {
	// Create a temp directory with a Go file that has lint issues
	tmpDir := t.TempDir()

	// Create a schema file that would produce lint issues (lowercase label - WN4052)
	code := `package schema

import "github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"

// person has lowercase label which violates WN4052 (should be PascalCase)
var person = schema.NodeType{
	Label: "person",
	Properties: []schema.Property{
		{Name: "name", Type: schema.STRING, Required: true},
	},
}
`
	filePath := filepath.Join(tmpDir, "schema.go")
	if err := os.WriteFile(filePath, []byte(code), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	linter := &neo4jLinter{}
	ctx := &Context{}

	// Test with WN4052 disabled - should NOT find WN4052 issues
	result, err := linter.Lint(ctx, tmpDir, LintOpts{
		Disable: []string{"WN4052"},
	})
	if err != nil {
		t.Fatalf("Lint failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Should not have WN4052 issues since it's disabled
	for _, e := range result.Errors {
		if e.Code == "WN4052" {
			t.Error("WN4052 should be disabled but was found in results")
		}
	}
}

// TestNeo4jLinter_Lint_Fix tests that Fix mode is properly handled
func TestNeo4jLinter_Lint_Fix(t *testing.T) {
	// Create a temp directory with a Go file that has lint issues
	tmpDir := t.TempDir()

	code := `package schema

import "github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"

var person = schema.NodeType{
	Label: "person",
	Properties: []schema.Property{
		{Name: "name", Type: schema.STRING, Required: true},
	},
}
`
	filePath := filepath.Join(tmpDir, "schema.go")
	if err := os.WriteFile(filePath, []byte(code), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	linter := &neo4jLinter{}
	ctx := &Context{}

	// Test with Fix=true - should include "auto-fix" or "Fix" in message
	result, err := linter.Lint(ctx, tmpDir, LintOpts{
		Fix: true,
	})
	if err != nil {
		t.Fatalf("Lint failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// If there are issues, the message should mention auto-fix
	if len(result.Errors) > 0 {
		if result.Message == "" {
			t.Error("expected non-empty message")
		}
		// The message should indicate Fix mode was requested
		found := false
		if contains(result.Message, "auto-fix") || contains(result.Message, "Fix") {
			found = true
		}
		if !found {
			t.Errorf("expected message to mention auto-fix when Fix=true, got: %s", result.Message)
		}
	}
}

// contains checks if substr is in s
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
