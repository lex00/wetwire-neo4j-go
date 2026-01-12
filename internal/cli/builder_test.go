package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lex00/wetwire-core-go/cmd"
	"github.com/lex00/wetwire-neo4j-go/internal/algorithms"
	"github.com/lex00/wetwire-neo4j-go/internal/kg"
	"github.com/lex00/wetwire-neo4j-go/internal/pipelines"
	"github.com/lex00/wetwire-neo4j-go/internal/projections"
	"github.com/lex00/wetwire-neo4j-go/internal/retrievers"
	"github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"
)

func TestNewBuilder(t *testing.T) {
	b := NewBuilder()
	if b == nil {
		t.Fatal("NewBuilder returned nil")
	}
	if b.scanner == nil {
		t.Error("scanner is nil")
	}
	if b.cypherSerializer == nil {
		t.Error("cypherSerializer is nil")
	}
	if b.jsonSerializer == nil {
		t.Error("jsonSerializer is nil")
	}
}

func TestBuilder_DetectFormat(t *testing.T) {
	b := NewBuilder()

	tests := []struct {
		output string
		want   string
	}{
		{"", "cypher"},
		{"output.json", "json"},
		{"output.JSON", "json"},
		{"output.cypher", "cypher"},
		{"output.cql", "cypher"},
		{"output.txt", "cypher"},
	}

	for _, tt := range tests {
		t.Run(tt.output, func(t *testing.T) {
			got := b.detectFormat(tt.output)
			if got != tt.want {
				t.Errorf("detectFormat(%q) = %q, want %q", tt.output, got, tt.want)
			}
		})
	}
}

func TestBuilder_Build_EmptyDir(t *testing.T) {
	b := NewBuilder()

	// Create temp directory with no Go files
	tmpDir, err := os.MkdirTemp("", "builder-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	opts := cmd.BuildOptions{Verbose: true}
	err = b.Build(context.Background(), tmpDir, opts)
	if err != nil {
		t.Errorf("Build on empty dir failed: %v", err)
	}
}

func TestBuilder_Build_WithResources(t *testing.T) {
	b := NewBuilder()

	// Create temp directory with a Go file containing a resource
	tmpDir, err := os.MkdirTemp("", "builder-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Write a Go file with a NodeType definition
	goFile := filepath.Join(tmpDir, "schema.go")
	content := `package schema

import "github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"

var Person = schema.NodeType{
	Label: "Person",
}
`
	if err := os.WriteFile(goFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	opts := cmd.BuildOptions{DryRun: true, Verbose: true}
	err = b.Build(context.Background(), tmpDir, opts)
	if err != nil {
		t.Errorf("Build failed: %v", err)
	}
}

func TestBuilder_BuildFromResources_Cypher(t *testing.T) {
	b := NewBuilder()

	nodeTypes := []*schema.NodeType{
		{
			Label: "Person",
			Properties: []schema.Property{
				{Name: "name", Type: schema.STRING, Required: true},
			},
			Constraints: []schema.Constraint{
				{Name: "person_name_unique", Type: schema.UNIQUE, Properties: []string{"name"}},
			},
		},
	}

	relTypes := []*schema.RelationshipType{
		{
			Label:  "KNOWS",
			Source: "Person",
			Target: "Person",
		},
	}

	output, err := b.BuildFromResources(nodeTypes, relTypes, nil, nil, nil, nil, nil, "cypher")
	if err != nil {
		t.Fatalf("BuildFromResources failed: %v", err)
	}

	if !strings.Contains(output, "Person") {
		t.Error("output should contain Person")
	}
	if !strings.Contains(output, "UNIQUE") {
		t.Error("output should contain UNIQUE")
	}
}

func TestBuilder_BuildFromResources_JSON(t *testing.T) {
	b := NewBuilder()

	nodeTypes := []*schema.NodeType{
		{Label: "Person"},
	}
	algos := []algorithms.Algorithm{
		&algorithms.PageRank{DampingFactor: 0.85},
	}
	pipes := []pipelines.Pipeline{
		&pipelines.NodeClassificationPipeline{
			BasePipeline: pipelines.BasePipeline{
				Name: "test-pipeline",
			},
		},
	}
	projs := []projections.Projection{
		&projections.NativeProjection{
			BaseProjection: projections.BaseProjection{
				Name: "test-graph",
			},
		},
	}
	rets := []retrievers.Retriever{
		&retrievers.VectorRetriever{
			IndexName: "test-index",
		},
	}
	kgPipes := []kg.KGPipeline{
		&kg.SimpleKGPipeline{
			BasePipeline: kg.BasePipeline{Name: "test-kg"},
		},
	}

	output, err := b.BuildFromResources(nodeTypes, nil, algos, pipes, projs, rets, kgPipes, "json")
	if err != nil {
		t.Fatalf("BuildFromResources failed: %v", err)
	}

	if !strings.Contains(output, "Person") {
		t.Error("output should contain Person")
	}
	if !strings.Contains(output, "algorithms") {
		t.Error("output should contain algorithms")
	}
	if !strings.Contains(output, "pipelines") {
		t.Error("output should contain pipelines")
	}
	if !strings.Contains(output, "projections") {
		t.Error("output should contain projections")
	}
	if !strings.Contains(output, "retrievers") {
		t.Error("output should contain retrievers")
	}
	if !strings.Contains(output, "kgPipelines") {
		t.Error("output should contain kgPipelines")
	}
}

func TestBuilder_BuildFromResources_InvalidFormat(t *testing.T) {
	b := NewBuilder()

	_, err := b.BuildFromResources(nil, nil, nil, nil, nil, nil, nil, "invalid")
	if err == nil {
		t.Error("expected error for invalid format")
	}
}

func TestBuilder_GenerateJSON(t *testing.T) {
	b := NewBuilder()

	// This tests the internal generateJSON method indirectly
	output, err := b.BuildFromResources([]*schema.NodeType{{Label: "Test"}}, nil, nil, nil, nil, nil, nil, "json")
	if err != nil {
		t.Fatalf("BuildFromResources failed: %v", err)
	}

	if !strings.HasPrefix(output, "{") {
		t.Error("JSON output should start with {")
	}
	if !strings.HasSuffix(strings.TrimSpace(output), "}") {
		t.Error("JSON output should end with }")
	}
}
