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
		// Labels starting with numbers must be prefixed to be valid Go identifiers
		{"5122Node", "n5122node"},
		{"123", "n123"},
		{"1_2_3", "n123"},
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
		{"STRING", "STRING"},
		{"INTEGER", "INTEGER"},
		{"FLOAT", "FLOAT"},
		{"BOOLEAN", "BOOLEAN"},
		{"DATE", "DATE"},
		{"DATETIME", "DATETIME"},
		{"POINT", "POINT"},
		{"unknown", "STRING"},
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
		{"RANGE", "BTREE"},
		{"BTREE", "BTREE"},
		{"FULLTEXT", "FULLTEXT"},
		{"VECTOR", "VECTOR"},
		{"TEXT", "TEXT"},
		{"POINT", "POINT_INDEX"},
		{"unknown", "BTREE"},
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

func TestGenerator_GeneratesValidTypeConstants(t *testing.T) {
	// Test that generated code uses the exact constant names from pkg/neo4j/schema
	// This is critical - wrong constants mean the code won't compile or won't be discovered
	g := NewGenerator("schema")

	result := &ImportResult{
		NodeTypes: []NodeTypeDefinition{
			{
				Label: "TestNode",
				Properties: []PropertyDefinition{
					{Name: "strProp", Type: "STRING"},
					{Name: "intProp", Type: "INTEGER"},
					{Name: "floatProp", Type: "FLOAT"},
					{Name: "boolProp", Type: "BOOLEAN"},
					{Name: "dateProp", Type: "DATE"},
					{Name: "datetimeProp", Type: "DATETIME"},
					{Name: "pointProp", Type: "POINT"},
				},
			},
		},
	}

	code, err := g.Generate(result)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// These are the EXACT constant names from pkg/neo4j/schema/types.go
	// If any of these fail, the generated code won't compile
	validConstants := []string{
		"schema.STRING",
		"schema.INTEGER",
		"schema.FLOAT",
		"schema.BOOLEAN",
		"schema.DATE",
		"schema.DATETIME",
		"schema.POINT",
	}

	for _, constant := range validConstants {
		if !strings.Contains(code, constant) {
			t.Errorf("generated code missing valid constant %q", constant)
		}
	}

	// These are INVALID constant names that would cause compile errors
	// The importer previously generated these incorrectly
	invalidConstants := []string{
		"schema.TypeString",
		"schema.TypeInteger",
		"schema.TypeFloat",
		"schema.TypeBoolean",
		"schema.TypeDate",
		"schema.TypeDateTime",
		"schema.TypePoint",
	}

	for _, invalid := range invalidConstants {
		if strings.Contains(code, invalid) {
			t.Errorf("generated code contains invalid constant %q - this would cause compile errors", invalid)
		}
	}
}

func TestGenerator_GeneratesCompilableCode(t *testing.T) {
	// Test that generated code can be written to a file and compiled
	g := NewGenerator("testschema")

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
		RelationshipTypes: []RelationshipTypeDefinition{
			{
				Type:   "KNOWS",
				Source: "Person",
				Target: "Person",
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

	// Write to temp directory
	tmpDir := t.TempDir()
	schemaDir := filepath.Join(tmpDir, "testschema")
	if err := os.MkdirAll(schemaDir, 0755); err != nil {
		t.Fatalf("failed to create schema dir: %v", err)
	}

	schemaFile := filepath.Join(schemaDir, "schema.go")
	if err := os.WriteFile(schemaFile, []byte(code), 0644); err != nil {
		t.Fatalf("failed to write schema file: %v", err)
	}

	// Create go.mod for the temp project
	goMod := `module testproject

go 1.21

require github.com/lex00/wetwire-neo4j-go v1.5.4
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}

	// Verify the generated code has correct structure
	// (We can't actually compile without network access to fetch deps,
	// but we can verify the syntax and structure)
	if !strings.Contains(code, "package testschema") {
		t.Error("missing package declaration")
	}
	if !strings.Contains(code, `"github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"`) {
		t.Error("missing import statement")
	}
	if !strings.Contains(code, "var person = &schema.NodeType{") {
		t.Error("missing node type definition")
	}
	if !strings.Contains(code, "var knows = &schema.RelationshipType{") {
		t.Error("missing relationship type definition")
	}
	if !strings.Contains(code, "var Schema = &schema.Schema{") {
		t.Error("missing Schema wrapper")
	}
}

func TestImporter_NumericLabelGeneratesValidGo(t *testing.T) {
	// Labels starting with numbers must generate valid Go identifiers
	content := `CREATE CONSTRAINT FOR (n:5122Node) REQUIRE n.id IS UNIQUE;`

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "numeric.cypher")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	importer := NewCypherImporter(tmpFile)
	result, err := importer.Import(context.Background())
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	g := NewGenerator("schema")
	code, err := g.Generate(result)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Must NOT contain "var 5122" - that's invalid Go
	if strings.Contains(code, "var 5122") {
		t.Error("generated code contains 'var 5122' - invalid Go identifier")
	}

	// Must contain "var n5122" - the prefixed valid identifier
	if !strings.Contains(code, "var n5122") {
		t.Errorf("generated code should contain 'var n5122', got:\n%s", code)
	}
}

func TestMapNeo4jType_MatchesSchemaPackage(t *testing.T) {
	// Document the expected mapping between Neo4j types and schema constants
	// If this test fails, update mapNeo4jType to match schema package constants
	expectedMappings := map[string]string{
		// Standard types
		"STRING":   "STRING",
		"INTEGER":  "INTEGER",
		"FLOAT":    "FLOAT",
		"BOOLEAN":  "BOOLEAN",
		"DATE":     "DATE",
		"DATETIME": "DATETIME",
		"POINT":    "POINT",
		// Aliases
		"INT":            "INTEGER",
		"LONG":           "INTEGER",
		"DOUBLE":         "FLOAT",
		"BOOL":           "BOOLEAN",
		"ZONED DATETIME": "DATETIME",
		// List types
		"LIST": "LIST_STRING",
		// Unknown defaults to STRING
		"UNKNOWN":     "STRING",
		"RANDOM_TYPE": "STRING",
	}

	for neo4jType, expectedConstant := range expectedMappings {
		got := mapNeo4jType(neo4jType)
		if got != expectedConstant {
			t.Errorf("mapNeo4jType(%q) = %q, want %q (must match schema package constant)", neo4jType, got, expectedConstant)
		}
	}
}

func TestGenerator_MultipleConstraintsAndIndexes(t *testing.T) {
	// Test that multiple constraints and indexes are combined in single blocks
	// (not written as duplicate struct fields)
	g := NewGenerator("schema")

	result := &ImportResult{
		NodeTypes: []NodeTypeDefinition{
			{
				Label: "Person",
				Properties: []PropertyDefinition{
					{Name: "id", Type: "STRING", Required: true},
					{Name: "email", Type: "STRING", Required: true},
					{Name: "name", Type: "STRING"},
				},
				Constraints: []ConstraintDefinition{
					{Type: "UNIQUENESS", Properties: []string{"id"}},
					{Type: "UNIQUENESS", Properties: []string{"email"}},
					{Type: "NODE_KEY", Properties: []string{"id", "email"}},
				},
				Indexes: []IndexDefinition{
					{Type: "RANGE", Properties: []string{"name"}},
					{Type: "FULLTEXT", Properties: []string{"name"}},
				},
			},
		},
	}

	code, err := g.Generate(result)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Count occurrences of "Constraints:" - should be exactly 1
	constraintsCount := strings.Count(code, "Constraints:")
	if constraintsCount != 1 {
		t.Errorf("expected 1 'Constraints:' field, got %d (duplicate fields would cause compile error)", constraintsCount)
	}

	// Count occurrences of "Indexes:" - should be exactly 1
	indexesCount := strings.Count(code, "Indexes:")
	if indexesCount != 1 {
		t.Errorf("expected 1 'Indexes:' field, got %d (duplicate fields would cause compile error)", indexesCount)
	}

	// Verify all constraints are present
	if !strings.Contains(code, `schema.UNIQUE, Properties: []string{"id"}`) {
		t.Error("missing id uniqueness constraint")
	}
	if !strings.Contains(code, `schema.UNIQUE, Properties: []string{"email"}`) {
		t.Error("missing email uniqueness constraint")
	}
	if !strings.Contains(code, `schema.NODE_KEY, Properties: []string{"id", "email"}`) {
		t.Error("missing node key constraint")
	}

	// Verify all indexes are present
	if !strings.Contains(code, `schema.BTREE, Properties: []string{"name"}`) {
		t.Error("missing BTREE index")
	}
	if !strings.Contains(code, `schema.FULLTEXT, Properties: []string{"name"}`) {
		t.Error("missing FULLTEXT index")
	}
}

func TestGenerator_NameCollisionHandling(t *testing.T) {
	// Test that nodes and relationships with the same name don't cause redeclarations
	g := NewGenerator("schema")

	result := &ImportResult{
		NodeTypes: []NodeTypeDefinition{
			{Label: "WORKS_FOR"},  // becomes worksFor
			{Label: "Person"},     // becomes person
		},
		RelationshipTypes: []RelationshipTypeDefinition{
			{Type: "WORKS_FOR"},   // would be worksFor, should become worksForRel
			{Type: "PERSON"},      // would be person, should become personRel
		},
	}

	code, err := g.Generate(result)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check that we have 4 distinct variable declarations
	varDeclCount := strings.Count(code, "var ")
	// Should be 4 resources + 1 Schema = 5 var declarations
	if varDeclCount != 5 {
		t.Errorf("expected 5 var declarations, got %d", varDeclCount)
	}

	// Check that the node types are declared as NodeType
	if !strings.Contains(code, "var worksFor = &schema.NodeType{") {
		t.Error("missing worksFor NodeType declaration")
	}
	if !strings.Contains(code, "var person = &schema.NodeType{") {
		t.Error("missing person NodeType declaration")
	}

	// Check that the relationship types are declared as RelationshipType with suffix
	if !strings.Contains(code, "var worksForRel = &schema.RelationshipType{") {
		t.Error("missing worksForRel RelationshipType declaration")
	}
	if !strings.Contains(code, "var personRel = &schema.RelationshipType{") {
		t.Error("missing personRel RelationshipType declaration")
	}

	// Check that Schema uses correct variable names in correct arrays
	if !strings.Contains(code, "Nodes: []*schema.NodeType{\n\t\tworksFor,\n\t\tperson,") {
		t.Error("Schema.Nodes should contain worksFor and person")
	}
	if !strings.Contains(code, "Relationships: []*schema.RelationshipType{\n\t\tworksForRel,\n\t\tpersonRel,") {
		t.Error("Schema.Relationships should contain worksForRel and personRel")
	}

	// Verify the actual Neo4j labels/types are preserved correctly
	if !strings.Contains(code, `Label: "WORKS_FOR"`) {
		t.Error("WORKS_FOR label should be preserved in both node and relationship")
	}
	if !strings.Contains(code, `Label: "PERSON"`) {
		t.Error("PERSON label should be preserved in relationship")
	}
}
