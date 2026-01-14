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

func TestImporterCLI_WriteToFile_AutoAppendsGoExtension(t *testing.T) {
	// Test that output without .go extension gets it appended automatically
	tmpDir := t.TempDir()

	cypherContent := `CREATE CONSTRAINT test_id IF NOT EXISTS FOR (t:Test) REQUIRE t.id IS UNIQUE;`
	cypherFile := filepath.Join(tmpDir, "input.cypher")
	if err := os.WriteFile(cypherFile, []byte(cypherContent), 0644); err != nil {
		t.Fatalf("failed to write Cypher file: %v", err)
	}

	// Output path WITHOUT .go extension
	outputFile := filepath.Join(tmpDir, "schema")

	importer := NewImporterCLI()
	err := importer.ImportToFile(cypherFile, "", "", "", "", "schema", outputFile)
	if err != nil {
		t.Fatalf("import to file failed: %v", err)
	}

	// Should have created schema.go, not schema
	expectedFile := filepath.Join(tmpDir, "schema.go")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Errorf("expected file %s to exist (auto-appended .go)", expectedFile)
	}

	// The bare "schema" file should NOT exist
	if _, err := os.Stat(outputFile); err == nil {
		// Only fail if schema.go doesn't exist (if user specified schema.go, both checks pass)
		if outputFile != expectedFile {
			t.Errorf("bare file %s should not exist when .go was auto-appended", outputFile)
		}
	}
}

func TestImporterCLI_E2E_ImportThenList(t *testing.T) {
	// End-to-end test: import schema → scanner finds it
	tmpDir := t.TempDir()

	// Create cypher file with schema
	cypherContent := `CREATE CONSTRAINT person_id FOR (p:Person) REQUIRE p.id IS UNIQUE;
CREATE CONSTRAINT company_name FOR (c:Company) REQUIRE c.name IS UNIQUE;`
	cypherFile := filepath.Join(tmpDir, "schema.cypher")
	if err := os.WriteFile(cypherFile, []byte(cypherContent), 0644); err != nil {
		t.Fatalf("failed to write Cypher file: %v", err)
	}

	// Import to output file (without .go - should be auto-appended)
	outputFile := filepath.Join(tmpDir, "schema")
	importer := NewImporterCLI()
	if err := importer.ImportToFile(cypherFile, "", "", "", "", "schema", outputFile); err != nil {
		t.Fatalf("import failed: %v", err)
	}

	// Now use the scanner directly to find resources
	lister := NewLister()
	resources, err := lister.ScanDir(tmpDir)
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	// Should find at least 2 node types (Person and Company)
	if len(resources) < 2 {
		t.Errorf("expected at least 2 resources, got %d", len(resources))
	}

	// Verify we found Person and Company (now returns Neo4j labels, not Go variable names)
	foundPerson := false
	foundCompany := false
	for _, r := range resources {
		if r.Name == "Person" {
			foundPerson = true
		}
		if r.Name == "Company" {
			foundCompany = true
		}
	}

	if !foundPerson {
		t.Error("scanner did not find Person node type")
	}
	if !foundCompany {
		t.Error("scanner did not find Company node type")
	}
}
