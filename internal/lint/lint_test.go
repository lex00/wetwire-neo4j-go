package lint

import (
	"testing"

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
}

func TestSeverity_Values(t *testing.T) {
	tests := []struct {
		s    Severity
		want string
	}{
		{Error, "error"},
		{Warning, "warning"},
		{Info, "info"},
	}

	for _, tt := range tests {
		if string(tt.s) != tt.want {
			t.Errorf("Severity = %v, want %v", string(tt.s), tt.want)
		}
	}
}

func TestLintResult_Structure(t *testing.T) {
	r := LintResult{
		Rule:     "WN4001",
		Severity: Error,
		Message:  "damping_factor must be in [0, 1)",
		Location: "PageRank.DampingFactor",
	}

	if r.Rule != "WN4001" {
		t.Errorf("Rule = %v, want WN4001", r.Rule)
	}
	if r.Severity != Error {
		t.Errorf("Severity = %v, want Error", r.Severity)
	}
}

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
		name        string
		dimension   int
		expectWarn  bool
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
		name        string
		label       string
		expectWarn  bool
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
		name        string
		relLabel    string
		expectWarn  bool
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
		name        string
		threshold   float64
		expectWarn  bool
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

func TestLinter_LintAll(t *testing.T) {
	l := NewLinter()

	// Create some resources with issues
	algo := &algorithms.PageRank{DampingFactor: 1.5} // Invalid
	node := &schema.NodeType{Label: "person"}        // Should be PascalCase
	rel := &schema.RelationshipType{Label: "worksFor"}  // Should be SCREAMING_SNAKE

	results := l.LintAll([]any{algo, node, rel})

	if len(results) < 3 {
		t.Errorf("expected at least 3 lint results, got %d", len(results))
	}
}

func TestLinter_HasErrors(t *testing.T) {
	results := []LintResult{
		{Severity: Warning},
		{Severity: Info},
	}

	if HasErrors(results) {
		t.Error("HasErrors should return false for warnings only")
	}

	results = append(results, LintResult{Severity: Error})
	if !HasErrors(results) {
		t.Error("HasErrors should return true when error present")
	}
}

func TestLinter_FilterBySeverity(t *testing.T) {
	results := []LintResult{
		{Severity: Error},
		{Severity: Warning},
		{Severity: Warning},
		{Severity: Info},
	}

	errors := FilterBySeverity(results, Error)
	if len(errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(errors))
	}

	warnings := FilterBySeverity(results, Warning)
	if len(warnings) != 2 {
		t.Errorf("expected 2 warnings, got %d", len(warnings))
	}
}

// Helper function
func containsRule(results []LintResult, rule string) bool {
	for _, r := range results {
		if r.Rule == rule {
			return true
		}
	}
	return false
}
