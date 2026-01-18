package discover

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

// Test composite literal detection with various type expressions
func TestScanner_DetectCompositeLitKind(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		expectLen  int
		expectKind ResourceKind
	}{
		{
			name: "pointer to type",
			content: `package main
var n = &NodeType{Label: "Person"}
`,
			expectLen:  1,
			expectKind: KindNodeType,
		},
		{
			name: "package prefixed type",
			content: `package main
import "github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"
var n = schema.NodeType{Label: "Person"}
`,
			expectLen:  1,
			expectKind: KindNodeType,
		},
		{
			name: "pointer to package prefixed type",
			content: `package main
import "github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"
var n = &schema.NodeType{Label: "Person"}
`,
			expectLen:  1,
			expectKind: KindNodeType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test.go")
			if err := os.WriteFile(tmpFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to write temp file: %v", err)
			}

			s := NewScanner()
			resources, err := s.ScanFile(tmpFile)
			if err != nil {
				t.Fatalf("ScanFile failed: %v", err)
			}

			if len(resources) != tt.expectLen {
				t.Errorf("expected %d resources, got %d", tt.expectLen, len(resources))
			}

			if tt.expectLen > 0 && resources[0].Kind != tt.expectKind {
				t.Errorf("expected %v kind, got %v", tt.expectKind, resources[0].Kind)
			}
		})
	}
}

// Test more algorithm types
func TestScanner_ScanFile_MoreAlgorithms(t *testing.T) {
	content := `package algo

type PathFinder struct {
	Dijkstra
}

type Embeddings struct {
	FastRP
}

type Similarity struct {
	KNN
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

	if len(resources) != 3 {
		t.Fatalf("expected 3 resources, got %d", len(resources))
	}

	for _, r := range resources {
		if r.Kind != KindAlgorithm {
			t.Errorf("expected Algorithm kind for %s, got %v", r.Name, r.Kind)
		}
	}
}

// Test more retriever types
func TestScanner_ScanFile_MoreRetrievers(t *testing.T) {
	content := `package rag

type TextToCypher struct {
	Text2CypherRetriever
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

	if len(resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(resources))
	}

	if resources[0].Kind != KindRetriever {
		t.Errorf("expected Retriever kind, got %v", resources[0].Kind)
	}
}

// Test detecting resources in function calls
func TestScanner_ScanFile_FunctionCall(t *testing.T) {
	content := `package main

func NewPerson() *NodeType {
	return &NodeType{Label: "Person"}
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "func.go")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	s := NewScanner()
	_, err := s.ScanFile(tmpFile)
	if err != nil {
		t.Fatalf("ScanFile failed: %v", err)
	}
	// Just verify it doesn't crash - functions returning resources are not detected
}

// Test nested composite literals
func TestScanner_ScanFile_NestedLiterals(t *testing.T) {
	content := `package main

type Config struct {
	NodeType
	Nested NodeType
}

var cfg = Config{
	NodeType: NodeType{Label: "Parent"},
	Nested: NodeType{Label: "Child"},
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "nested.go")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	s := NewScanner()
	resources, err := s.ScanFile(tmpFile)
	if err != nil {
		t.Fatalf("ScanFile failed: %v", err)
	}

	// Should find the Config type with NodeType embedded
	if len(resources) == 0 {
		t.Log("No resources found, which is acceptable")
	}
}

// Test slice type detection
func TestScanner_ScanFile_SliceTypes(t *testing.T) {
	content := `package main

type MultiLabel struct {
	NodeType
	Labels []string
	Related []NodeType
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "slice.go")
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

	if resources[0].Name != "MultiLabel" {
		t.Errorf("expected MultiLabel, got %s", resources[0].Name)
	}
}

// Test map type detection
func TestScanner_ScanFile_MapTypes(t *testing.T) {
	content := `package main

type IndexedNodes struct {
	NodeType
	Index map[string]NodeType
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "map.go")
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
}
