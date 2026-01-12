package discovery

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewScanner(t *testing.T) {
	s := NewScanner()
	if s == nil {
		t.Fatal("NewScanner returned nil")
	}
	if s.fset == nil {
		t.Error("fset is nil")
	}
	if len(s.typeAliases) == 0 {
		t.Error("typeAliases is empty")
	}
}

func TestScanner_ScanFile_NodeType(t *testing.T) {
	// Create a temporary file with a NodeType definition
	content := `package main

import "github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"

type Person struct {
	schema.NodeType
	Name string
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

	if len(resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(resources))
	}

	r := resources[0]
	if r.Name != "Person" {
		t.Errorf("Name = %v, want Person", r.Name)
	}
	if r.Kind != KindNodeType {
		t.Errorf("Kind = %v, want NodeType", r.Kind)
	}
	if r.Package != "main" {
		t.Errorf("Package = %v, want main", r.Package)
	}
	if r.Line != 5 {
		t.Errorf("Line = %v, want 5", r.Line)
	}
}

func TestScanner_ScanFile_RelationshipType(t *testing.T) {
	content := `package schema

type WorksFor struct {
	RelationshipType
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "rel.go")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	s := NewScanner()
	resources, err := s.ScanFile(tmpFile)
	if err != nil {
		t.Fatalf("ScanFile failed: %v", err)
	}

	if len(resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(resources))
	}

	r := resources[0]
	if r.Name != "WorksFor" {
		t.Errorf("Name = %v, want WorksFor", r.Name)
	}
	if r.Kind != KindRelationshipType {
		t.Errorf("Kind = %v, want RelationshipType", r.Kind)
	}
}

func TestScanner_ScanFile_Algorithm(t *testing.T) {
	content := `package algo

type InfluenceScore struct {
	PageRank
	DampingFactor float64
}

type Communities struct {
	Louvain
	MaxIterations int
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "algo.go")
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

	for _, r := range resources {
		if r.Kind != KindAlgorithm {
			t.Errorf("expected Algorithm kind, got %v", r.Kind)
		}
	}
}

func TestScanner_ScanFile_Pipeline(t *testing.T) {
	content := `package ml

type ClassifyNodes struct {
	NodeClassificationPipeline
}

type PredictLinks struct {
	LinkPredictionPipeline
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "ml.go")
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

	for _, r := range resources {
		if r.Kind != KindPipeline {
			t.Errorf("expected Pipeline kind, got %v", r.Kind)
		}
	}
}

func TestScanner_ScanFile_Retriever(t *testing.T) {
	content := `package rag

type SemanticSearch struct {
	VectorRetriever
}

type HybridSearch struct {
	HybridCypherRetriever
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "rag.go")
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

	for _, r := range resources {
		if r.Kind != KindRetriever {
			t.Errorf("expected Retriever kind, got %v", r.Kind)
		}
	}
}

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

	if !names["person"] {
		t.Error("expected to find 'person' variable")
	}
	if !names["company"] {
		t.Error("expected to find 'company' variable")
	}
}

func TestScanner_ScanFile_NoResources(t *testing.T) {
	content := `package main

type Foo struct {
	Bar string
}

func main() {}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	s := NewScanner()
	resources, err := s.ScanFile(tmpFile)
	if err != nil {
		t.Fatalf("ScanFile failed: %v", err)
	}

	if len(resources) != 0 {
		t.Errorf("expected 0 resources, got %d", len(resources))
	}
}

func TestScanner_ScanFile_InvalidFile(t *testing.T) {
	s := NewScanner()
	_, err := s.ScanFile("/nonexistent/file.go")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestScanner_ScanFile_InvalidSyntax(t *testing.T) {
	content := `package main

func invalid syntax {
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "invalid.go")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	s := NewScanner()
	_, err := s.ScanFile(tmpFile)
	if err == nil {
		t.Error("expected error for invalid syntax")
	}
}

func TestScanner_ScanDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple files
	files := map[string]string{
		"schema.go": `package schema
type Person struct { NodeType }
type Company struct { NodeType }
`,
		"relationships.go": `package schema
type WorksFor struct { RelationshipType }
`,
		"subdir/algo.go": `package algo
type Influence struct { PageRank }
`,
	}

	for path, content := range files {
		fullPath := filepath.Join(tmpDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
	}

	s := NewScanner()
	resources, err := s.ScanDir(tmpDir)
	if err != nil {
		t.Fatalf("ScanDir failed: %v", err)
	}

	if len(resources) != 4 {
		t.Errorf("expected 4 resources, got %d", len(resources))
	}

	// Check we found all types
	kinds := make(map[ResourceKind]int)
	for _, r := range resources {
		kinds[r.Kind]++
	}

	if kinds[KindNodeType] != 2 {
		t.Errorf("expected 2 NodeType, got %d", kinds[KindNodeType])
	}
	if kinds[KindRelationshipType] != 1 {
		t.Errorf("expected 1 RelationshipType, got %d", kinds[KindRelationshipType])
	}
	if kinds[KindAlgorithm] != 1 {
		t.Errorf("expected 1 Algorithm, got %d", kinds[KindAlgorithm])
	}
}

func TestScanner_ScanDir_SkipsTestFiles(t *testing.T) {
	tmpDir := t.TempDir()

	files := map[string]string{
		"schema.go": `package schema
type Person struct { NodeType }
`,
		"schema_test.go": `package schema
type TestPerson struct { NodeType }
`,
	}

	for path, content := range files {
		fullPath := filepath.Join(tmpDir, path)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
	}

	s := NewScanner()
	resources, err := s.ScanDir(tmpDir)
	if err != nil {
		t.Fatalf("ScanDir failed: %v", err)
	}

	// Should only find Person, not TestPerson
	if len(resources) != 1 {
		t.Errorf("expected 1 resource, got %d", len(resources))
	}
	if resources[0].Name != "Person" {
		t.Errorf("expected Person, got %s", resources[0].Name)
	}
}

func TestScanner_ScanDir_SkipsVendor(t *testing.T) {
	tmpDir := t.TempDir()

	files := map[string]string{
		"schema.go": `package schema
type Person struct { NodeType }
`,
		"vendor/other/schema.go": `package other
type VendorType struct { NodeType }
`,
	}

	for path, content := range files {
		fullPath := filepath.Join(tmpDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
	}

	s := NewScanner()
	resources, err := s.ScanDir(tmpDir)
	if err != nil {
		t.Fatalf("ScanDir failed: %v", err)
	}

	// Should only find Person, not VendorType
	if len(resources) != 1 {
		t.Errorf("expected 1 resource, got %d", len(resources))
	}
}

func TestDependencyGraph_TopologicalSort(t *testing.T) {
	resources := []DiscoveredResource{
		{Name: "C", Kind: KindNodeType, Dependencies: []string{"A", "B"}},
		{Name: "A", Kind: KindNodeType, Dependencies: []string{}},
		{Name: "B", Kind: KindNodeType, Dependencies: []string{"A"}},
	}

	g := NewDependencyGraph(resources)
	sorted, err := g.TopologicalSort()
	if err != nil {
		t.Fatalf("TopologicalSort failed: %v", err)
	}

	if len(sorted) != 3 {
		t.Fatalf("expected 3 resources, got %d", len(sorted))
	}

	// A should come before B, B should come before C
	positions := make(map[string]int)
	for i, r := range sorted {
		positions[r.Name] = i
	}

	if positions["A"] > positions["B"] {
		t.Error("A should come before B")
	}
	if positions["B"] > positions["C"] {
		t.Error("B should come before C")
	}
	if positions["A"] > positions["C"] {
		t.Error("A should come before C")
	}
}

func TestDependencyGraph_TopologicalSort_Cycle(t *testing.T) {
	resources := []DiscoveredResource{
		{Name: "A", Kind: KindNodeType, Dependencies: []string{"B"}},
		{Name: "B", Kind: KindNodeType, Dependencies: []string{"C"}},
		{Name: "C", Kind: KindNodeType, Dependencies: []string{"A"}},
	}

	g := NewDependencyGraph(resources)
	_, err := g.TopologicalSort()
	if err == nil {
		t.Error("expected error for circular dependency")
	}
}

func TestDependencyGraph_HasCycle(t *testing.T) {
	t.Run("no cycle", func(t *testing.T) {
		resources := []DiscoveredResource{
			{Name: "A", Kind: KindNodeType, Dependencies: []string{}},
			{Name: "B", Kind: KindNodeType, Dependencies: []string{"A"}},
		}
		g := NewDependencyGraph(resources)
		if g.HasCycle() {
			t.Error("should not have cycle")
		}
	})

	t.Run("has cycle", func(t *testing.T) {
		resources := []DiscoveredResource{
			{Name: "A", Kind: KindNodeType, Dependencies: []string{"B"}},
			{Name: "B", Kind: KindNodeType, Dependencies: []string{"A"}},
		}
		g := NewDependencyGraph(resources)
		if !g.HasCycle() {
			t.Error("should have cycle")
		}
	})
}

func TestDependencyGraph_GetDependencies(t *testing.T) {
	resources := []DiscoveredResource{
		{Name: "A", Kind: KindNodeType, Dependencies: []string{}},
		{Name: "B", Kind: KindNodeType, Dependencies: []string{"A"}},
		{Name: "C", Kind: KindNodeType, Dependencies: []string{"B"}},
	}

	g := NewDependencyGraph(resources)

	t.Run("leaf node", func(t *testing.T) {
		deps := g.GetDependencies("A")
		if len(deps) != 0 {
			t.Errorf("expected 0 deps, got %d", len(deps))
		}
	})

	t.Run("one level", func(t *testing.T) {
		deps := g.GetDependencies("B")
		if len(deps) != 1 || deps[0] != "A" {
			t.Errorf("expected [A], got %v", deps)
		}
	})

	t.Run("recursive", func(t *testing.T) {
		deps := g.GetDependencies("C")
		if len(deps) != 2 {
			t.Errorf("expected 2 deps, got %d", len(deps))
		}
	})
}

func TestIsPrimitiveType(t *testing.T) {
	primitives := []string{"bool", "string", "int", "int64", "float64", "byte", "error", "any"}
	for _, p := range primitives {
		if !isPrimitiveType(p) {
			t.Errorf("%s should be primitive", p)
		}
	}

	nonPrimitives := []string{"Person", "MyType", "NodeType"}
	for _, p := range nonPrimitives {
		if isPrimitiveType(p) {
			t.Errorf("%s should not be primitive", p)
		}
	}
}

func TestIsValidIdentifier(t *testing.T) {
	valid := []string{"foo", "Foo", "_foo", "foo123", "FooBar"}
	for _, v := range valid {
		if !isValidIdentifier(v) {
			t.Errorf("%s should be valid", v)
		}
	}

	invalid := []string{"", "123foo", "foo-bar", "foo.bar"}
	for _, v := range invalid {
		if isValidIdentifier(v) {
			t.Errorf("%s should be invalid", v)
		}
	}
}

func TestResourceKind_Constants(t *testing.T) {
	tests := []struct {
		kind ResourceKind
		want string
	}{
		{KindNodeType, "NodeType"},
		{KindRelationshipType, "RelationshipType"},
		{KindAlgorithm, "Algorithm"},
		{KindPipeline, "Pipeline"},
		{KindRetriever, "Retriever"},
	}

	for _, tt := range tests {
		if string(tt.kind) != tt.want {
			t.Errorf("ResourceKind = %v, want %v", tt.kind, tt.want)
		}
	}
}
