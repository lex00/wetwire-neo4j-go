package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewListCommand(t *testing.T) {
	cmd := newListCommand()

	if cmd.Use != "list" {
		t.Errorf("expected Use 'list', got %s", cmd.Use)
	}
	if cmd.Short == "" {
		t.Error("expected Short description")
	}

	// Check flags exist
	pathFlag := cmd.Flags().Lookup("path")
	if pathFlag == nil {
		t.Error("expected path flag")
	}
	formatFlag := cmd.Flags().Lookup("format")
	if formatFlag == nil {
		t.Error("expected format flag")
	}
}

func TestNewListCommand_Execute(t *testing.T) {
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

	cmd := newListCommand()
	cmd.SetArgs([]string{"--path", tmpDir})

	// Command executes without error - output goes to stdout which we can't easily capture
	// in this test, but we verify the command runs successfully
	err = cmd.Execute()
	if err != nil {
		t.Errorf("list command failed: %v", err)
	}
}

func TestNewValidateCommand(t *testing.T) {
	cmd := newValidateCommand()

	if cmd.Use != "validate" {
		t.Errorf("expected Use 'validate', got %s", cmd.Use)
	}
	if cmd.Short == "" {
		t.Error("expected Short description")
	}

	// Check flags exist
	flags := []string{"path", "uri", "username", "password", "database", "dry-run"}
	for _, flag := range flags {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("expected %s flag", flag)
		}
	}
}

func TestNewValidateCommand_DryRun(t *testing.T) {
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

	cmd := newValidateCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--path", tmpDir, "--dry-run"})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("validate dry-run failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Person") {
		t.Errorf("expected output to contain 'Person', got: %s", output)
	}
}

func TestNewImportCommand(t *testing.T) {
	cmd := newImportCommand()

	if cmd.Use != "import" {
		t.Errorf("expected Use 'import', got %s", cmd.Use)
	}
	if cmd.Short == "" {
		t.Error("expected Short description")
	}

	// Check flags exist
	flags := []string{"file", "uri", "username", "password", "database", "package", "output"}
	for _, flag := range flags {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("expected %s flag", flag)
		}
	}
}

func TestNewImportCommand_FromCypher(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a Cypher file
	cypherContent := `CREATE CONSTRAINT person_id IF NOT EXISTS FOR (p:Person) REQUIRE p.id IS UNIQUE;
CREATE INDEX person_name IF NOT EXISTS FOR (p:Person) ON (p.name);
`
	cypherFile := filepath.Join(tmpDir, "schema.cypher")
	err := os.WriteFile(cypherFile, []byte(cypherContent), 0644)
	if err != nil {
		t.Fatalf("failed to write cypher file: %v", err)
	}

	cmd := newImportCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--file", cypherFile})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("import from cypher failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Person") {
		t.Errorf("expected output to contain 'Person', got: %s", output)
	}
}

func TestNewGraphCommand(t *testing.T) {
	cmd := newGraphCommand()

	if cmd.Use != "graph" {
		t.Errorf("expected Use 'graph', got %s", cmd.Use)
	}
	if cmd.Short == "" {
		t.Error("expected Short description")
	}

	// Check flags exist
	pathFlag := cmd.Flags().Lookup("path")
	if pathFlag == nil {
		t.Error("expected path flag")
	}
	formatFlag := cmd.Flags().Lookup("format")
	if formatFlag == nil {
		t.Error("expected format flag")
	}
}

func TestNewGraphCommand_DOT(t *testing.T) {
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

	cmd := newGraphCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--path", tmpDir, "--format", "dot"})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("graph dot command failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "digraph") {
		t.Errorf("expected DOT output to contain 'digraph', got: %s", output)
	}
}

func TestNewGraphCommand_Mermaid(t *testing.T) {
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

	cmd := newGraphCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--path", tmpDir, "--format", "mermaid"})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("graph mermaid command failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "graph TD") {
		t.Errorf("expected Mermaid output to contain 'graph TD', got: %s", output)
	}
}

func TestNewVersionCommand(t *testing.T) {
	cmd := newVersionCommand()

	if cmd.Use != "version" {
		t.Errorf("expected Use 'version', got %s", cmd.Use)
	}
	if cmd.Short == "" {
		t.Error("expected Short description")
	}
}

func TestNewVersionCommand_Execute(t *testing.T) {
	cmd := newVersionCommand()

	// Version command uses fmt.Printf directly, so we verify it executes without error
	err := cmd.Execute()
	if err != nil {
		t.Errorf("version command failed: %v", err)
	}
}

func TestNewListCommand_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()

	cmd := newListCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--path", tmpDir})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("list command failed on empty dir: %v", err)
	}
}

func TestNewGraphCommand_InvalidFormat(t *testing.T) {
	tmpDir := t.TempDir()

	cmd := newGraphCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--path", tmpDir, "--format", "invalid"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid format")
	}
}

func TestNewListCommand_JSONFormat(t *testing.T) {
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

	cmd := newListCommand()
	cmd.SetArgs([]string{"--path", tmpDir, "--format", "json"})

	// Command executes without error - output goes to stdout
	err = cmd.Execute()
	if err != nil {
		t.Errorf("list json command failed: %v", err)
	}
}
