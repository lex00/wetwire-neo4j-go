package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGraphCLI_GenerateDOT(t *testing.T) {
	// Create temp directory with test definitions
	tmpDir := t.TempDir()

	// Create a schema file with dependencies
	schemaContent := `package schema

import "github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"

var Person = &schema.NodeType{
	Label: "Person",
}

var Company = &schema.NodeType{
	Label: "Company",
}

var WorksFor = &schema.RelationshipType{
	Label:  "WORKS_FOR",
	Source: "Person",
	Target: "Company",
}
`
	err := os.WriteFile(filepath.Join(tmpDir, "schema.go"), []byte(schemaContent), 0644)
	if err != nil {
		t.Fatalf("failed to write schema file: %v", err)
	}

	graph := NewGraphCLI()

	var buf bytes.Buffer
	err = graph.Generate(tmpDir, "dot", &buf)
	if err != nil {
		t.Errorf("generate DOT failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "digraph") {
		t.Errorf("expected DOT output to contain 'digraph', got: %s", output)
	}
	if !strings.Contains(output, "Person") {
		t.Errorf("expected DOT output to contain 'Person', got: %s", output)
	}
}

func TestGraphCLI_GenerateMermaid(t *testing.T) {
	// Create temp directory with test definitions
	tmpDir := t.TempDir()

	// Create a schema file
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

	graph := NewGraphCLI()

	var buf bytes.Buffer
	err = graph.Generate(tmpDir, "mermaid", &buf)
	if err != nil {
		t.Errorf("generate Mermaid failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "graph TD") {
		t.Errorf("expected Mermaid output to contain 'graph TD', got: %s", output)
	}
	if !strings.Contains(output, "Person") {
		t.Errorf("expected Mermaid output to contain 'Person', got: %s", output)
	}
}

func TestGraphCLI_InvalidFormat(t *testing.T) {
	tmpDir := t.TempDir()

	graph := NewGraphCLI()

	var buf bytes.Buffer
	err := graph.Generate(tmpDir, "invalid", &buf)
	if err == nil {
		t.Error("expected error for invalid format")
	}
	if !strings.Contains(err.Error(), "format") {
		t.Errorf("expected format error, got: %v", err)
	}
}

func TestGraphCLI_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	graph := NewGraphCLI()

	var buf bytes.Buffer
	err := graph.Generate(tmpDir, "dot", &buf)
	if err != nil {
		t.Errorf("generate failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "digraph") {
		t.Errorf("expected DOT output even for empty graph, got: %s", output)
	}
}
