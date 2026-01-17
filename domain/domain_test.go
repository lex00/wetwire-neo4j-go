package domain

import (
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
