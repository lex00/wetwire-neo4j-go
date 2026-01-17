package lint

import (
	"testing"

	"github.com/lex00/wetwire-neo4j-go/internal/algorithms"
	"github.com/lex00/wetwire-neo4j-go/internal/kg"
	"github.com/lex00/wetwire-neo4j-go/internal/pipelines"
	"github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"
)

// WN4001: damping_factor must be in [0, 1)
func TestLinter_WN4001_DampingFactor(t *testing.T) {
	l := NewLinter()

	tests := []struct {
		name          string
		dampingFactor float64
		expectError   bool
	}{
		{"valid 0.85", 0.85, false},
		{"valid 0", 0, false},
		{"valid 0.99", 0.99, false},
		{"invalid 1.0", 1.0, true},
		{"invalid negative", -0.1, true},
		{"invalid > 1", 1.5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			algo := &algorithms.PageRank{
				DampingFactor: tt.dampingFactor,
			}
			results := l.LintAlgorithm(algo)
			hasError := containsRule(results, "WN4001")
			if hasError != tt.expectError {
				t.Errorf("WN4001 error = %v, want %v", hasError, tt.expectError)
			}
		})
	}
}

// WN4002: max_iterations must be positive
func TestLinter_WN4002_MaxIterations(t *testing.T) {
	l := NewLinter()

	tests := []struct {
		name        string
		maxIter     int
		expectError bool
	}{
		{"valid 20", 20, false},
		{"valid 1", 1, false},
		{"invalid 0", 0, false}, // 0 means use default
		{"invalid negative", -5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			algo := &algorithms.PageRank{
				MaxIterations: tt.maxIter,
			}
			results := l.LintAlgorithm(algo)
			hasError := containsRule(results, "WN4002")
			if hasError != tt.expectError {
				t.Errorf("WN4002 error = %v, want %v", hasError, tt.expectError)
			}
		})
	}
}

// WN4006: embedding_dimension should be power of 2
func TestLinter_WN4006_EmbeddingDimension(t *testing.T) {
	l := NewLinter()

	tests := []struct {
		name       string
		dimension  int
		expectWarn bool
	}{
		{"power of 2: 128", 128, false},
		{"power of 2: 256", 256, false},
		{"power of 2: 64", 64, false},
		{"not power: 100", 100, true},
		{"not power: 150", 150, true},
		{"zero (default)", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			algo := &algorithms.FastRP{
				EmbeddingDimension: tt.dimension,
			}
			results := l.LintAlgorithm(algo)
			hasWarn := containsRule(results, "WN4006")
			if hasWarn != tt.expectWarn {
				t.Errorf("WN4006 warning = %v, want %v", hasWarn, tt.expectWarn)
			}
		})
	}
}

// WN4007: topK must be positive, warn if > 1000
func TestLinter_WN4007_TopK(t *testing.T) {
	l := NewLinter()

	tests := []struct {
		name       string
		topK       int
		expectWarn bool
	}{
		{"valid 10", 10, false},
		{"valid 1000", 1000, false},
		{"warn 1001", 1001, true},
		{"warn 5000", 5000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			algo := &algorithms.NodeSimilarity{
				TopK: tt.topK,
			}
			results := l.LintAlgorithm(algo)
			hasWarn := containsRule(results, "WN4007")
			if hasWarn != tt.expectWarn {
				t.Errorf("WN4007 warning = %v, want %v", hasWarn, tt.expectWarn)
			}
		})
	}
}

// WN4032: at least one model candidate required
func TestLinter_WN4032_ModelRequired(t *testing.T) {
	l := NewLinter()

	t.Run("no models", func(t *testing.T) {
		pipeline := &pipelines.NodeClassificationPipeline{
			BasePipeline: pipelines.BasePipeline{
				Name:   "test",
				Models: []pipelines.Model{},
			},
		}
		results := l.LintPipeline(pipeline)
		if !containsRule(results, "WN4032") {
			t.Error("expected WN4032 error for empty models")
		}
	})

	t.Run("has models", func(t *testing.T) {
		pipeline := &pipelines.NodeClassificationPipeline{
			BasePipeline: pipelines.BasePipeline{
				Name: "test",
				Models: []pipelines.Model{
					&pipelines.LogisticRegression{},
				},
			},
		}
		results := l.LintPipeline(pipeline)
		if containsRule(results, "WN4032") {
			t.Error("unexpected WN4032 error when models present")
		}
	})
}

// WN4040: Schema must have at least one EntityType
func TestLinter_WN4040_EntityTypeRequired(t *testing.T) {
	l := NewLinter()

	t.Run("no entity types", func(t *testing.T) {
		pipeline := &kg.SimpleKGPipeline{
			EntityTypes: []kg.EntityType{},
		}
		results := l.LintKGPipeline(pipeline)
		if !containsRule(results, "WN4040") {
			t.Error("expected WN4040 error for empty entity types")
		}
	})

	t.Run("has entity types", func(t *testing.T) {
		pipeline := &kg.SimpleKGPipeline{
			EntityTypes: []kg.EntityType{
				{Name: "Person"},
			},
		}
		results := l.LintKGPipeline(pipeline)
		if containsRule(results, "WN4040") {
			t.Error("unexpected WN4040 error when entity types present")
		}
	})
}

// WN4052: Node labels should be PascalCase
func TestLinter_WN4052_NodeLabelCase(t *testing.T) {
	l := NewLinter()

	tests := []struct {
		name       string
		label      string
		expectWarn bool
	}{
		{"PascalCase", "Person", false},
		{"PascalCase multi", "PersonDetails", false},
		{"lowercase", "person", true},
		{"snake_case", "person_details", true},
		{"SCREAMING", "PERSON", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &schema.NodeType{
				Label: tt.label,
			}
			results := l.LintNodeType(node)
			hasWarn := containsRule(results, "WN4052")
			if hasWarn != tt.expectWarn {
				t.Errorf("WN4052 warning = %v, want %v", hasWarn, tt.expectWarn)
			}
		})
	}
}

// WN4053: Relationship types should be SCREAMING_SNAKE_CASE
func TestLinter_WN4053_RelationshipTypeCase(t *testing.T) {
	l := NewLinter()

	tests := []struct {
		name       string
		relLabel   string
		expectWarn bool
	}{
		{"SCREAMING_SNAKE", "WORKS_FOR", false},
		{"SCREAMING single", "KNOWS", false},
		{"camelCase", "worksFor", true},
		{"PascalCase", "WorksFor", true},
		{"lowercase", "works_for", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rel := &schema.RelationshipType{
				Label: tt.relLabel,
			}
			results := l.LintRelationshipType(rel)
			hasWarn := containsRule(results, "WN4053")
			if hasWarn != tt.expectWarn {
				t.Errorf("WN4053 warning = %v, want %v", hasWarn, tt.expectWarn)
			}
		})
	}
}

// WN4043: Entity resolver threshold should be >= 0.8
func TestLinter_WN4043_ResolverThreshold(t *testing.T) {
	l := NewLinter()

	tests := []struct {
		name       string
		threshold  float64
		expectWarn bool
	}{
		{"valid 0.85", 0.85, false},
		{"valid 0.8", 0.8, false},
		{"valid 0.95", 0.95, false},
		{"warn 0.5", 0.5, true},
		{"warn 0.7", 0.7, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pipeline := &kg.SimpleKGPipeline{
				EntityTypes: []kg.EntityType{{Name: "Test"}},
				EntityResolver: &kg.FuzzyMatchResolver{
					Threshold: tt.threshold,
				},
			}
			results := l.LintKGPipeline(pipeline)
			hasWarn := containsRule(results, "WN4043")
			if hasWarn != tt.expectWarn {
				t.Errorf("WN4043 warning = %v, want %v", hasWarn, tt.expectWarn)
			}
		})
	}
}

// Test ArticleRank linting (WN4001, WN4002)
func TestLinter_LintArticleRank(t *testing.T) {
	l := NewLinter()

	tests := []struct {
		name          string
		algo          *algorithms.ArticleRank
		expectRule    string
		expectPresent bool
	}{
		{
			name:          "valid ArticleRank",
			algo:          &algorithms.ArticleRank{DampingFactor: 0.85, MaxIterations: 20},
			expectRule:    "WN4001",
			expectPresent: false,
		},
		{
			name:          "invalid damping factor",
			algo:          &algorithms.ArticleRank{DampingFactor: 1.5},
			expectRule:    "WN4001",
			expectPresent: true,
		},
		{
			name:          "negative max iterations",
			algo:          &algorithms.ArticleRank{MaxIterations: -1},
			expectRule:    "WN4002",
			expectPresent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := l.LintAlgorithm(tt.algo)
			hasRule := containsRule(results, tt.expectRule)
			if hasRule != tt.expectPresent {
				t.Errorf("rule %s present = %v, want %v", tt.expectRule, hasRule, tt.expectPresent)
			}
		})
	}
}

// Test Node2Vec linting (WN4006)
func TestLinter_LintNode2Vec(t *testing.T) {
	l := NewLinter()

	tests := []struct {
		name       string
		algo       *algorithms.Node2Vec
		expectWarn bool
	}{
		{
			name:       "valid dimension 64",
			algo:       &algorithms.Node2Vec{EmbeddingDimension: 64},
			expectWarn: false,
		},
		{
			name:       "valid dimension 128",
			algo:       &algorithms.Node2Vec{EmbeddingDimension: 128},
			expectWarn: false,
		},
		{
			name:       "non-power of 2",
			algo:       &algorithms.Node2Vec{EmbeddingDimension: 100},
			expectWarn: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := l.LintAlgorithm(tt.algo)
			hasWarn := containsRule(results, "WN4006")
			if hasWarn != tt.expectWarn {
				t.Errorf("WN4006 warning = %v, want %v", hasWarn, tt.expectWarn)
			}
		})
	}
}

// Test KNN linting (WN4007)
func TestLinter_LintKNN(t *testing.T) {
	l := NewLinter()

	tests := []struct {
		name       string
		algo       *algorithms.KNN
		expectWarn bool
	}{
		{
			name:       "valid K",
			algo:       &algorithms.KNN{K: 10},
			expectWarn: false,
		},
		{
			name:       "high K",
			algo:       &algorithms.KNN{K: 2000},
			expectWarn: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := l.LintAlgorithm(tt.algo)
			hasWarn := containsRule(results, "WN4007")
			if hasWarn != tt.expectWarn {
				t.Errorf("WN4007 warning = %v, want %v", hasWarn, tt.expectWarn)
			}
		})
	}
}

// Test split config validation (WN4031: test_fraction must be < 1.0)
func TestLinter_WN4031_SplitConfig(t *testing.T) {
	l := NewLinter()

	tests := []struct {
		name        string
		testFrac    float64
		expectError bool
	}{
		{"valid 0.2", 0.2, false},
		{"valid 0.0", 0.0, false}, // 0 is allowed (uses default)
		{"invalid 1", 1.0, true},
		{"invalid > 1", 1.5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pipeline := &pipelines.NodeClassificationPipeline{
				BasePipeline: pipelines.BasePipeline{
					Name:   "test",
					Models: []pipelines.Model{&pipelines.LogisticRegression{}},
				},
				SplitConfig: pipelines.SplitConfig{
					TestFraction: tt.testFrac,
				},
			}
			results := l.LintPipeline(pipeline)
			hasError := containsRule(results, "WN4031")
			if hasError != tt.expectError {
				t.Errorf("WN4031 error = %v, want %v", hasError, tt.expectError)
			}
		})
	}
}
