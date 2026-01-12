package importer

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewGenerator(t *testing.T) {
	g := NewGenerator("mypackage")
	if g == nil {
		t.Fatal("NewGenerator returned nil")
	}
	if g.PackageName != "mypackage" {
		t.Errorf("PackageName = %q, want %q", g.PackageName, "mypackage")
	}
}

func TestGenerator_Generate_NodeTypes(t *testing.T) {
	g := NewGenerator("schema")

	result := &ImportResult{
		NodeTypes: []NodeTypeDefinition{
			{
				Label: "Person",
				Properties: []PropertyDefinition{
					{Name: "name", Type: "STRING", Required: true},
					{Name: "age", Type: "INTEGER"},
				},
				Constraints: []ConstraintDefinition{
					{Type: "UNIQUENESS", Properties: []string{"name"}},
				},
			},
		},
	}

	code, err := g.Generate(result)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check package declaration
	if !strings.Contains(code, "package schema") {
		t.Error("missing package declaration")
	}

	// Check import
	if !strings.Contains(code, `"github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"`) {
		t.Error("missing import")
	}

	// Check node type variable
	if !strings.Contains(code, "var person = &schema.NodeType{") {
		t.Error("missing node type variable")
	}

	// Check label
	if !strings.Contains(code, `Label: "Person"`) {
		t.Error("missing label")
	}

	// Check properties
	if !strings.Contains(code, `Name: "name"`) {
		t.Error("missing name property")
	}
	if !strings.Contains(code, "Required: true") {
		t.Error("missing required flag")
	}
}

func TestGenerator_Generate_RelationshipTypes(t *testing.T) {
	g := NewGenerator("schema")

	result := &ImportResult{
		RelationshipTypes: []RelationshipTypeDefinition{
			{
				Type:   "WORKS_FOR",
				Source: "Person",
				Target: "Company",
				Properties: []PropertyDefinition{
					{Name: "since", Type: "DATE"},
				},
			},
		},
	}

	code, err := g.Generate(result)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check relationship type variable
	if !strings.Contains(code, "var worksFor = &schema.RelationshipType{") {
		t.Error("missing relationship type variable")
	}

	// Check source and target
	if !strings.Contains(code, `Source: "Person"`) {
		t.Error("missing source")
	}
	if !strings.Contains(code, `Target: "Company"`) {
		t.Error("missing target")
	}
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Person", "person"},
		{"WORKS_FOR", "worksFor"},
		{"MY_TEST_TYPE", "myTestType"},
		{"simple", "simple"},
	}

	for _, tt := range tests {
		got := toCamelCase(tt.input)
		if got != tt.want {
			t.Errorf("toCamelCase(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestMapNeo4jType(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"STRING", "TypeString"},
		{"INTEGER", "TypeInteger"},
		{"FLOAT", "TypeFloat"},
		{"BOOLEAN", "TypeBoolean"},
		{"DATE", "TypeDate"},
		{"DATETIME", "TypeDateTime"},
		{"POINT", "TypePoint"},
		{"unknown", "TypeString"},
	}

	for _, tt := range tests {
		got := mapNeo4jType(tt.input)
		if got != tt.want {
			t.Errorf("mapNeo4jType(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestMapIndexType(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"RANGE", "RangeIndex"},
		{"FULLTEXT", "FullTextIndex"},
		{"VECTOR", "VectorIndex"},
		{"TEXT", "TextIndex"},
		{"unknown", "RangeIndex"},
	}

	for _, tt := range tests {
		got := mapIndexType(tt.input)
		if got != tt.want {
			t.Errorf("mapIndexType(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFormatStringSlice(t *testing.T) {
	tests := []struct {
		input []string
		want  string
	}{
		{[]string{}, "[]string{}"},
		{[]string{"a"}, `[]string{"a"}`},
		{[]string{"a", "b"}, `[]string{"a", "b"}`},
	}

	for _, tt := range tests {
		got := formatStringSlice(tt.input)
		if got != tt.want {
			t.Errorf("formatStringSlice(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestCypherImporter_Import(t *testing.T) {
	content := `
-- Schema definitions
CREATE CONSTRAINT person_id_unique FOR (p:Person) REQUIRE p.id IS UNIQUE;
CREATE CONSTRAINT company_name_unique FOR (c:Company) REQUIRE c.name IS UNIQUE;
CREATE INDEX person_name_idx FOR (p:Person) ON (p.name);
CREATE RANGE INDEX document_idx FOR (d:Document) ON (d.title);
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "schema.cypher")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	importer := NewCypherImporter(tmpFile)
	result, err := importer.Import(context.Background())
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	// Check constraints
	if len(result.Constraints) != 2 {
		t.Errorf("expected 2 constraints, got %d", len(result.Constraints))
	}

	// Check indexes
	if len(result.Indexes) < 1 {
		t.Errorf("expected at least 1 index, got %d", len(result.Indexes))
	}

	// Check node types were built
	if len(result.NodeTypes) < 2 {
		t.Errorf("expected at least 2 node types, got %d", len(result.NodeTypes))
	}
}

func TestCypherImporter_Import_NodeKey(t *testing.T) {
	content := `CREATE CONSTRAINT order_key FOR (o:Order) REQUIRE (o.id, o.region) IS NODE KEY;`

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "nodekey.cypher")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	importer := NewCypherImporter(tmpFile)
	result, err := importer.Import(context.Background())
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	if len(result.Constraints) != 1 {
		t.Fatalf("expected 1 constraint, got %d", len(result.Constraints))
	}

	c := result.Constraints[0]
	if c.Type != "NODE_KEY" {
		t.Errorf("Type = %q, want NODE_KEY", c.Type)
	}
	if c.Label != "Order" {
		t.Errorf("Label = %q, want Order", c.Label)
	}
	if len(c.Properties) != 2 {
		t.Errorf("expected 2 properties, got %d: %v", len(c.Properties), c.Properties)
	}
}

func TestCypherImporter_Import_Existence(t *testing.T) {
	content := `CREATE CONSTRAINT person_name_exists FOR (p:Person) REQUIRE p.name IS NOT NULL;`

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "existence.cypher")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	importer := NewCypherImporter(tmpFile)
	result, err := importer.Import(context.Background())
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	if len(result.Constraints) != 1 {
		t.Fatalf("expected 1 constraint, got %d", len(result.Constraints))
	}

	c := result.Constraints[0]
	if c.Type != "NODE_PROPERTY_EXISTENCE" {
		t.Errorf("Type = %q, want NODE_PROPERTY_EXISTENCE", c.Type)
	}
}

func TestCypherImporter_Import_FileNotFound(t *testing.T) {
	importer := NewCypherImporter("/nonexistent/file.cypher")
	_, err := importer.Import(context.Background())
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestParsePropertyList(t *testing.T) {
	tests := []struct {
		propStr string
		varName string
		want    []string
	}{
		{"n.name", "n", []string{"name"}},
		{"n.id, n.name", "n", []string{"id", "name"}},
		{"name", "", []string{"name"}},
		{"id, name", "", []string{"id", "name"}},
	}

	for _, tt := range tests {
		got := parsePropertyList(tt.propStr, tt.varName)
		if len(got) != len(tt.want) {
			t.Errorf("parsePropertyList(%q, %q) = %v, want %v", tt.propStr, tt.varName, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("parsePropertyList(%q, %q)[%d] = %q, want %q", tt.propStr, tt.varName, i, got[i], tt.want[i])
			}
		}
	}
}
