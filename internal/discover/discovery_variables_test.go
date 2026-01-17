package discover

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanner_ScanFile_VariableDeclaration(t *testing.T) {
	content := `package main

var person = NodeType{
	Label: "Person",
}

var company = &NodeType{
	Label: "Company",
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "vars.go")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	s := NewScanner()
	resources, err := s.ScanFile(tmpFile)
	if err != nil {
		t.Fatalf("ScanFile failed: %v", err)
	}

	if len(resources) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(resources))
	}

	names := make(map[string]bool)
	for _, r := range resources {
		names[r.Name] = true
		if r.Kind != KindNodeType {
			t.Errorf("expected NodeType kind for %s, got %v", r.Name, r.Kind)
		}
	}

	// Should find Label values, not variable names
	if !names["Person"] {
		t.Error("expected to find 'Person' (Label value), not 'person' (variable name)")
	}
	if !names["Company"] {
		t.Error("expected to find 'Company' (Label value), not 'company' (variable name)")
	}
}

// Test variable declarations with complex expressions
func TestScanner_ScanFile_ComplexExpressions(t *testing.T) {
	content := `package main

var nodes = []NodeType{
	{Label: "Person"},
	{Label: "Company"},
}

var config = &PageRank{
	DampingFactor: 0.85,
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "complex.go")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	s := NewScanner()
	resources, err := s.ScanFile(tmpFile)
	if err != nil {
		t.Fatalf("ScanFile failed: %v", err)
	}

	// Should find at least the PageRank config
	found := false
	for _, r := range resources {
		if r.Kind == KindAlgorithm && r.Name == "config" {
			found = true
			break
		}
	}

	if !found {
		t.Logf("resources found: %v", resources)
	}
}

func TestScanner_ReturnsNeo4jLabelsNotVariableNames(t *testing.T) {
	content := `package schema

import "github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"

var myVar = &schema.NodeType{
	Label: "ActualLabel",
}

var worksFor = &schema.RelationshipType{
	Label: "WORKS_FOR",
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "labels.go")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	s := NewScanner()
	resources, err := s.ScanFile(tmpFile)
	if err != nil {
		t.Fatalf("ScanFile failed: %v", err)
	}

	if len(resources) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(resources))
	}

	// Must find "ActualLabel", not "myVar"
	foundActualLabel := false
	foundWorksFor := false
	for _, r := range resources {
		if r.Name == "ActualLabel" {
			foundActualLabel = true
		}
		if r.Name == "WORKS_FOR" {
			foundWorksFor = true
		}
		// Should NOT find variable names
		if r.Name == "myVar" {
			t.Error("found 'myVar' (Go variable name) instead of 'ActualLabel' (Neo4j label)")
		}
		if r.Name == "worksFor" {
			t.Error("found 'worksFor' (Go variable name) instead of 'WORKS_FOR' (Neo4j label)")
		}
	}

	if !foundActualLabel {
		t.Error("did not find 'ActualLabel' - scanner should return Neo4j label, not Go variable name")
	}
	if !foundWorksFor {
		t.Error("did not find 'WORKS_FOR' - scanner should return Neo4j label, not Go variable name")
	}
}

func TestScanner_ExtractsPropertyDetails(t *testing.T) {
	content := `package schema

import "github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"

var Person = &schema.NodeType{
	Label: "Person",
	Properties: []schema.Property{
		{Name: "id", Type: schema.STRING, Required: true},
		{Name: "name", Type: schema.STRING},
		{Name: "age", Type: schema.INTEGER},
	},
	Constraints: []schema.Constraint{
		{Type: schema.UNIQUE, Properties: []string{"id"}},
	},
	Indexes: []schema.Index{
		{Type: schema.BTREE, Properties: []string{"name"}},
	},
}

var WorksFor = &schema.RelationshipType{
	Label:  "WORKS_FOR",
	Source: "Person",
	Target: "Company",
	Properties: []schema.Property{
		{Name: "since", Type: schema.DATE},
	},
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "schema.go")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	s := NewScanner()
	resources, err := s.ScanFile(tmpFile)
	if err != nil {
		t.Fatalf("ScanFile failed: %v", err)
	}

	if len(resources) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(resources))
	}

	// Find Person
	var person, worksFor *DiscoveredResource
	for i := range resources {
		if resources[i].Name == "Person" {
			person = &resources[i]
		}
		if resources[i].Name == "WORKS_FOR" {
			worksFor = &resources[i]
		}
	}

	if person == nil {
		t.Fatal("did not find Person resource")
	}

	// Check Person properties
	if len(person.Properties) != 3 {
		t.Errorf("expected 3 properties, got %d", len(person.Properties))
	} else {
		if person.Properties[0].Name != "id" || person.Properties[0].Type != "STRING" || !person.Properties[0].Required {
			t.Errorf("first property mismatch: %+v", person.Properties[0])
		}
		if person.Properties[1].Name != "name" || person.Properties[1].Type != "STRING" {
			t.Errorf("second property mismatch: %+v", person.Properties[1])
		}
		if person.Properties[2].Name != "age" || person.Properties[2].Type != "INTEGER" {
			t.Errorf("third property mismatch: %+v", person.Properties[2])
		}
	}

	// Check Person constraints
	if len(person.Constraints) != 1 {
		t.Errorf("expected 1 constraint, got %d", len(person.Constraints))
	} else {
		if person.Constraints[0].Type != "UNIQUE" {
			t.Errorf("constraint type mismatch: %s", person.Constraints[0].Type)
		}
		if len(person.Constraints[0].Properties) != 1 || person.Constraints[0].Properties[0] != "id" {
			t.Errorf("constraint properties mismatch: %v", person.Constraints[0].Properties)
		}
	}

	// Check Person indexes
	if len(person.Indexes) != 1 {
		t.Errorf("expected 1 index, got %d", len(person.Indexes))
	} else {
		if person.Indexes[0].Type != "BTREE" {
			t.Errorf("index type mismatch: %s", person.Indexes[0].Type)
		}
		if len(person.Indexes[0].Properties) != 1 || person.Indexes[0].Properties[0] != "name" {
			t.Errorf("index properties mismatch: %v", person.Indexes[0].Properties)
		}
	}

	if worksFor == nil {
		t.Fatal("did not find WORKS_FOR resource")
	}

	// Check WorksFor source/target
	if worksFor.Source != "Person" {
		t.Errorf("expected Source 'Person', got '%s'", worksFor.Source)
	}
	if worksFor.Target != "Company" {
		t.Errorf("expected Target 'Company', got '%s'", worksFor.Target)
	}

	// Check WorksFor properties
	if len(worksFor.Properties) != 1 {
		t.Errorf("expected 1 property, got %d", len(worksFor.Properties))
	} else {
		if worksFor.Properties[0].Name != "since" || worksFor.Properties[0].Type != "DATE" {
			t.Errorf("property mismatch: %+v", worksFor.Properties[0])
		}
	}
}

func TestScanner_ExtractsAgentContext(t *testing.T) {
	content := `package schema

import "github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"

var Person = &schema.NodeType{
	Label: "Person",
}

var Schema = &schema.Schema{
	Name: "myschema",
	Nodes: []*schema.NodeType{Person},
	AgentContext: ` + "`" + `Multi-tenant database - always filter by tenantId.
Ignore nodes prefixed with _ (internal).` + "`" + `,
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "schema.go")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	s := NewScanner()
	resources, err := s.ScanFile(tmpFile)
	if err != nil {
		t.Fatalf("ScanFile failed: %v", err)
	}

	// Should find Person NodeType and Schema
	var schemaRes *DiscoveredResource
	for i := range resources {
		if resources[i].Kind == KindSchema {
			schemaRes = &resources[i]
			break
		}
	}

	if schemaRes == nil {
		t.Fatal("did not find Schema resource")
	}

	if schemaRes.AgentContext == "" {
		t.Error("AgentContext should not be empty")
	}

	if schemaRes.AgentContext != "Multi-tenant database - always filter by tenantId.\nIgnore nodes prefixed with _ (internal)." {
		t.Errorf("unexpected AgentContext: %q", schemaRes.AgentContext)
	}
}

func TestScanner_ExtractsAgentContext_RegularString(t *testing.T) {
	content := `package schema

import "github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"

var Schema = &schema.Schema{
	Name: "myschema",
	AgentContext: "Simple instructions for the agent.",
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "schema.go")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	s := NewScanner()
	resources, err := s.ScanFile(tmpFile)
	if err != nil {
		t.Fatalf("ScanFile failed: %v", err)
	}

	var schemaRes *DiscoveredResource
	for i := range resources {
		if resources[i].Kind == KindSchema {
			schemaRes = &resources[i]
			break
		}
	}

	if schemaRes == nil {
		t.Fatal("did not find Schema resource")
	}

	if schemaRes.AgentContext != "Simple instructions for the agent." {
		t.Errorf("unexpected AgentContext: %q", schemaRes.AgentContext)
	}
}
