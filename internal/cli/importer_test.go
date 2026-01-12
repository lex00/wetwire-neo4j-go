package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestImporterCLI_ImportFromCypher(t *testing.T) {
	// Create temp directory with test Cypher file
	tmpDir := t.TempDir()

	// Create a Cypher schema file
	cypherContent := `-- Person constraints and indexes
CREATE CONSTRAINT person_id IF NOT EXISTS FOR (p:Person) REQUIRE p.id IS UNIQUE;
CREATE INDEX person_name IF NOT EXISTS FOR (p:Person) ON (p.name);
`
	cypherFile := filepath.Join(tmpDir, "schema.cypher")
	err := os.WriteFile(cypherFile, []byte(cypherContent), 0644)
	if err != nil {
		t.Fatalf("failed to write Cypher file: %v", err)
	}

	importer := NewImporterCLI()

	var buf bytes.Buffer
	err = importer.ImportFromCypher(cypherFile, "schema", &buf)
	if err != nil {
		t.Errorf("import from Cypher failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "package schema") {
		t.Errorf("expected output to contain 'package schema', got: %s", output)
	}
	if !strings.Contains(output, "Person") {
		t.Errorf("expected output to contain 'Person', got: %s", output)
	}
}

func TestImporterCLI_ImportFromCypherNotFound(t *testing.T) {
	importer := NewImporterCLI()

	var buf bytes.Buffer
	err := importer.ImportFromCypher("/nonexistent/file.cypher", "schema", &buf)
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestImporterCLI_ImportFromNeo4jRequiresURI(t *testing.T) {
	importer := NewImporterCLI()

	var buf bytes.Buffer
	err := importer.ImportFromNeo4j("", "", "", "", "schema", &buf)
	if err == nil {
		t.Error("expected error when no URI provided")
	}
	if !strings.Contains(err.Error(), "URI") {
		t.Errorf("expected URI error, got: %v", err)
	}
}

func TestImporterCLI_WriteToFile(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create a Cypher schema file
	cypherContent := `CREATE CONSTRAINT company_id IF NOT EXISTS FOR (c:Company) REQUIRE c.id IS UNIQUE;`
	cypherFile := filepath.Join(tmpDir, "input.cypher")
	err := os.WriteFile(cypherFile, []byte(cypherContent), 0644)
	if err != nil {
		t.Fatalf("failed to write Cypher file: %v", err)
	}

	outputFile := filepath.Join(tmpDir, "output.go")

	importer := NewImporterCLI()
	err = importer.ImportToFile(cypherFile, "", "", "", "", "myschema", outputFile)
	if err != nil {
		t.Errorf("import to file failed: %v", err)
	}

	// Verify output file was created
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	if !strings.Contains(string(content), "package myschema") {
		t.Errorf("expected output to contain 'package myschema', got: %s", string(content))
	}
	if !strings.Contains(string(content), "Company") {
		t.Errorf("expected output to contain 'Company', got: %s", string(content))
	}
}

func TestImporterCLI_DefaultPackageName(t *testing.T) {
	importer := NewImporterCLI()

	pkg := importer.DefaultPackageName()
	if pkg != "schema" {
		t.Errorf("expected default package to be 'schema', got: %s", pkg)
	}
}
