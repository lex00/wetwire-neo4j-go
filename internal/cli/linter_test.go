package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/lex00/wetwire-core-go/cmd"
	"github.com/lex00/wetwire-neo4j-go/internal/algorithms"
	"github.com/lex00/wetwire-neo4j-go/internal/kg"
	"github.com/lex00/wetwire-neo4j-go/internal/pipelines"
	"github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"
)

func TestNewLinter(t *testing.T) {
	l := NewLinter()
	if l == nil {
		t.Fatal("NewLinter returned nil")
	}
	if l.scanner == nil {
		t.Error("scanner is nil")
	}
	if l.linter == nil {
		t.Error("linter is nil")
	}
}

func TestLinter_Lint_EmptyDir(t *testing.T) {
	l := NewLinter()

	// Create temp directory with no Go files
	tmpDir, err := os.MkdirTemp("", "linter-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	opts := cmd.LintOptions{}
	issues, err := l.Lint(context.Background(), tmpDir, opts)
	if err != nil {
		t.Errorf("Lint on empty dir failed: %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("expected 0 issues, got %d", len(issues))
	}
}

func TestLinter_Lint_WithResources(t *testing.T) {
	l := NewLinter()

	// Create temp directory with a Go file
	tmpDir, err := os.MkdirTemp("", "linter-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Write a Go file with naming issues
	goFile := filepath.Join(tmpDir, "schema.go")
	content := `package schema

import "github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"

var person = schema.NodeType{
	Label: "person",
}
`
	if err := os.WriteFile(goFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	opts := cmd.LintOptions{}
	issues, err := l.Lint(context.Background(), tmpDir, opts)
	if err != nil {
		t.Errorf("Lint failed: %v", err)
	}

	// The variable 'person' starts with lowercase, which may or may not trigger
	// based on discovery (discovery looks for uppercase-starting types)
	// This test verifies the lint runs without error
	_ = issues
}

func TestLinter_LintAlgorithm(t *testing.T) {
	l := NewLinter()

	tests := []struct {
		name        string
		algo        algorithms.Algorithm
		expectIssue bool
		rule        string
	}{
		{
			name:        "valid PageRank",
			algo:        &algorithms.PageRank{DampingFactor: 0.85},
			expectIssue: false,
		},
		{
			name:        "invalid PageRank dampingFactor",
			algo:        &algorithms.PageRank{DampingFactor: 1.5},
			expectIssue: true,
			rule:        "WN4001",
		},
		{
			name:        "negative maxIterations",
			algo:        &algorithms.PageRank{MaxIterations: -5},
			expectIssue: true,
			rule:        "WN4002",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := l.LintAlgorithm(tt.algo, "test.go", 10)
			hasIssue := len(issues) > 0
			if hasIssue != tt.expectIssue {
				t.Errorf("LintAlgorithm issue = %v, want %v", hasIssue, tt.expectIssue)
			}
			if tt.expectIssue && len(issues) > 0 {
				found := false
				for _, issue := range issues {
					if issue.Rule == tt.rule {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected rule %s not found", tt.rule)
				}
			}
		})
	}
}

func TestLinter_LintPipeline(t *testing.T) {
	l := NewLinter()

	tests := []struct {
		name        string
		pipeline    pipelines.Pipeline
		expectIssue bool
		rule        string
	}{
		{
			name: "pipeline with models",
			pipeline: &pipelines.NodeClassificationPipeline{
				BasePipeline: pipelines.BasePipeline{
					Name:   "test",
					Models: []pipelines.Model{&pipelines.LogisticRegression{}},
				},
			},
			expectIssue: false,
		},
		{
			name: "pipeline without models",
			pipeline: &pipelines.NodeClassificationPipeline{
				BasePipeline: pipelines.BasePipeline{
					Name:   "test",
					Models: []pipelines.Model{},
				},
			},
			expectIssue: true,
			rule:        "WN4032",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := l.LintPipeline(tt.pipeline, "test.go", 10)
			hasIssue := len(issues) > 0
			if hasIssue != tt.expectIssue {
				t.Errorf("LintPipeline issue = %v, want %v", hasIssue, tt.expectIssue)
			}
		})
	}
}

func TestLinter_LintKGPipeline(t *testing.T) {
	l := NewLinter()

	tests := []struct {
		name        string
		pipeline    kg.KGPipeline
		expectIssue bool
	}{
		{
			name: "pipeline with entity types",
			pipeline: &kg.SimpleKGPipeline{
				EntityTypes: []kg.EntityType{{Name: "Person"}},
			},
			expectIssue: false,
		},
		{
			name: "pipeline without entity types",
			pipeline: &kg.SimpleKGPipeline{
				EntityTypes: []kg.EntityType{},
			},
			expectIssue: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := l.LintKGPipeline(tt.pipeline, "test.go", 10)
			hasIssue := len(issues) > 0
			if hasIssue != tt.expectIssue {
				t.Errorf("LintKGPipeline issue = %v, want %v", hasIssue, tt.expectIssue)
			}
		})
	}
}

func TestLinter_LintNodeType(t *testing.T) {
	l := NewLinter()

	tests := []struct {
		name        string
		node        *schema.NodeType
		expectIssue bool
	}{
		{
			name:        "PascalCase label",
			node:        &schema.NodeType{Label: "Person"},
			expectIssue: false,
		},
		{
			name:        "lowercase label",
			node:        &schema.NodeType{Label: "person"},
			expectIssue: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := l.LintNodeType(tt.node, "test.go", 10)
			hasIssue := len(issues) > 0
			if hasIssue != tt.expectIssue {
				t.Errorf("LintNodeType issue = %v, want %v", hasIssue, tt.expectIssue)
			}
		})
	}
}

func TestLinter_LintRelationshipType(t *testing.T) {
	l := NewLinter()

	tests := []struct {
		name        string
		rel         *schema.RelationshipType
		expectIssue bool
	}{
		{
			name:        "SCREAMING_SNAKE_CASE",
			rel:         &schema.RelationshipType{Label: "WORKS_FOR"},
			expectIssue: false,
		},
		{
			name:        "camelCase",
			rel:         &schema.RelationshipType{Label: "worksFor"},
			expectIssue: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := l.LintRelationshipType(tt.rel, "test.go", 10)
			hasIssue := len(issues) > 0
			if hasIssue != tt.expectIssue {
				t.Errorf("LintRelationshipType issue = %v, want %v", hasIssue, tt.expectIssue)
			}
		})
	}
}

func TestIsPascalCase(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"Person", true},
		{"PersonDetails", true},
		{"person", false},
		{"PERSON", false},
		{"person_details", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isPascalCase(tt.input)
			if got != tt.want {
				t.Errorf("isPascalCase(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsScreamingSnakeCase(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"WORKS_FOR", true},
		{"KNOWS", true},
		{"PERSON_123", true},
		{"worksFor", false},
		{"Works_For", false},
		{"works_for", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isScreamingSnakeCase(tt.input)
			if got != tt.want {
				t.Errorf("isScreamingSnakeCase(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestLinter_LintAll(t *testing.T) {
	l := NewLinter()

	resources := []any{
		&algorithms.PageRank{DampingFactor: 1.5},      // Invalid
		&schema.NodeType{Label: "person"},             // Invalid
		&schema.RelationshipType{Label: "worksFor"},   // Invalid
	}

	issues := l.LintAll(resources, "test.go", 10)
	if len(issues) < 3 {
		t.Errorf("expected at least 3 issues, got %d", len(issues))
	}
}
