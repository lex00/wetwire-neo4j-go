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
		name        string
		algo        *algorithms.Node2Vec
		expectWarn  bool
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

// Test FormatResults
func TestLinter_FormatResults(t *testing.T) {
	results := []LintResult{
		{Rule: "WN4001", Severity: Error, Message: "damping factor invalid", Location: "PageRank.DampingFactor"},
		{Rule: "WN4006", Severity: Warning, Message: "dimension not power of 2", Location: "FastRP.EmbeddingDimension"},
	}

	output := FormatResults(results)

	if output == "" {
		t.Error("FormatResults returned empty string")
	}

	// Check that results are formatted
	if !contains(output, "WN4001") {
		t.Error("FormatResults should include rule WN4001")
	}
	if !contains(output, "WN4006") {
		t.Error("FormatResults should include rule WN4006")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
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
		{"valid 0.0", 0.0, false},    // 0 is allowed (uses default)
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

// WN4010: Use typed NodeType/RelationshipType, not raw structs
func TestLinter_WN4010_TypedDefinitions(t *testing.T) {
	l := NewLinter()

	t.Run("NodeType with empty label", func(t *testing.T) {
		node := &schema.NodeType{
			Label: "", // Empty label indicates raw struct
		}
		results := l.LintNodeType(node)
		if !containsRule(results, "WN4010") {
			t.Error("expected WN4010 error for empty node label")
		}
	})

	t.Run("NodeType with valid label", func(t *testing.T) {
		node := &schema.NodeType{
			Label: "Person",
		}
		results := l.LintNodeType(node)
		if containsRule(results, "WN4010") {
			t.Error("unexpected WN4010 error for valid node label")
		}
	})

	t.Run("RelationshipType with empty label", func(t *testing.T) {
		rel := &schema.RelationshipType{
			Label:  "", // Empty label indicates raw struct
			Source: "Person",
			Target: "Company",
		}
		results := l.LintRelationshipType(rel)
		if !containsRule(results, "WN4010") {
			t.Error("expected WN4010 error for empty relationship label")
		}
	})

	t.Run("RelationshipType with valid label", func(t *testing.T) {
		rel := &schema.RelationshipType{
			Label:  "WORKS_FOR",
			Source: "Person",
			Target: "Company",
		}
		results := l.LintRelationshipType(rel)
		if containsRule(results, "WN4010") {
			t.Error("unexpected WN4010 error for valid relationship label")
		}
	})
}

// WN4011: Extract inline Property definitions to named vars
func TestLinter_WN4011_InlineProperties(t *testing.T) {
	l := NewLinter() // default maxInlineProperties = 5

	t.Run("NodeType with acceptable inline properties", func(t *testing.T) {
		node := &schema.NodeType{
			Label: "Person",
			Properties: []schema.Property{
				{Name: "name", Type: schema.STRING},
				{Name: "age", Type: schema.INTEGER},
				{Name: "email", Type: schema.STRING},
			},
		}
		results := l.LintNodeType(node)
		if containsRule(results, "WN4011") {
			t.Error("unexpected WN4011 warning for acceptable property count")
		}
	})

	t.Run("NodeType with too many inline properties", func(t *testing.T) {
		node := &schema.NodeType{
			Label: "Person",
			Properties: []schema.Property{
				{Name: "name", Type: schema.STRING},
				{Name: "age", Type: schema.INTEGER},
				{Name: "email", Type: schema.STRING},
				{Name: "phone", Type: schema.STRING},
				{Name: "address", Type: schema.STRING},
				{Name: "city", Type: schema.STRING},
			},
		}
		results := l.LintNodeType(node)
		if !containsRule(results, "WN4011") {
			t.Error("expected WN4011 warning for too many inline properties")
		}
	})

	t.Run("RelationshipType with too many inline properties", func(t *testing.T) {
		rel := &schema.RelationshipType{
			Label:  "WORKS_FOR",
			Source: "Person",
			Target: "Company",
			Properties: []schema.Property{
				{Name: "since", Type: schema.DATE},
				{Name: "salary", Type: schema.FLOAT},
				{Name: "title", Type: schema.STRING},
				{Name: "department", Type: schema.STRING},
				{Name: "location", Type: schema.STRING},
				{Name: "manager", Type: schema.STRING},
			},
		}
		results := l.LintRelationshipType(rel)
		if !containsRule(results, "WN4011") {
			t.Error("expected WN4011 warning for too many inline properties")
		}
	})

	t.Run("custom threshold", func(t *testing.T) {
		customLinter := NewLinter().WithMaxInlineProperties(2)
		node := &schema.NodeType{
			Label: "Person",
			Properties: []schema.Property{
				{Name: "name", Type: schema.STRING},
				{Name: "age", Type: schema.INTEGER},
				{Name: "email", Type: schema.STRING},
			},
		}
		results := customLinter.LintNodeType(node)
		if !containsRule(results, "WN4011") {
			t.Error("expected WN4011 warning with custom threshold of 2")
		}
	})
}

// WN4012: Prevent deeply nested schema definitions (max depth)
func TestLinter_WN4012_NestingDepth(t *testing.T) {
	l := NewLinter() // default maxNestingDepth = 3

	t.Run("NodeType at valid depth", func(t *testing.T) {
		node := &schema.NodeType{
			Label: "Person",
		}
		// Direct call at depth 0 should pass
		results := l.LintNodeType(node)
		if containsRule(results, "WN4012") {
			t.Error("unexpected WN4012 warning at depth 0")
		}
	})

	t.Run("NodeType exceeds max depth via internal call", func(t *testing.T) {
		// Test the internal function directly to verify depth checking
		l := NewLinter().WithMaxNestingDepth(2)
		node := &schema.NodeType{
			Label: "DeepNode",
		}
		// Call with depth > maxNestingDepth
		results := l.lintNodeTypeWithDepth(node, 3)
		if !containsRule(results, "WN4012") {
			t.Error("expected WN4012 warning when depth exceeds max")
		}
	})

	t.Run("RelationshipType exceeds max depth via internal call", func(t *testing.T) {
		l := NewLinter().WithMaxNestingDepth(2)
		rel := &schema.RelationshipType{
			Label:  "DEEP_REL",
			Source: "A",
			Target: "B",
		}
		// Call with depth > maxNestingDepth
		results := l.lintRelationshipTypeWithDepth(rel, 3)
		if !containsRule(results, "WN4012") {
			t.Error("expected WN4012 warning when depth exceeds max")
		}
	})

	t.Run("custom max depth", func(t *testing.T) {
		customLinter := NewLinter().WithMaxNestingDepth(1)
		node := &schema.NodeType{
			Label: "Person",
		}
		// At depth 2, should warn with maxNestingDepth of 1
		results := customLinter.lintNodeTypeWithDepth(node, 2)
		if !containsRule(results, "WN4012") {
			t.Error("expected WN4012 warning with custom max depth of 1")
		}
	})
}

// WN4013: Use direct references for relationship Source/Target
func TestLinter_WN4013_DirectReferences(t *testing.T) {
	l := NewLinter()

	t.Run("RelationshipType with valid Source and Target", func(t *testing.T) {
		rel := &schema.RelationshipType{
			Label:  "WORKS_FOR",
			Source: "Person",
			Target: "Company",
		}
		results := l.LintRelationshipType(rel)
		if containsRule(results, "WN4013") {
			t.Error("unexpected WN4013 error for valid Source/Target")
		}
	})

	t.Run("RelationshipType with empty Source", func(t *testing.T) {
		rel := &schema.RelationshipType{
			Label:  "WORKS_FOR",
			Source: "",
			Target: "Company",
		}
		results := l.LintRelationshipType(rel)
		if !containsRule(results, "WN4013") {
			t.Error("expected WN4013 error for empty Source")
		}
	})

	t.Run("RelationshipType with empty Target", func(t *testing.T) {
		rel := &schema.RelationshipType{
			Label:  "WORKS_FOR",
			Source: "Person",
			Target: "",
		}
		results := l.LintRelationshipType(rel)
		if !containsRule(results, "WN4013") {
			t.Error("expected WN4013 error for empty Target")
		}
	})

	t.Run("RelationshipType with both Source and Target empty", func(t *testing.T) {
		rel := &schema.RelationshipType{
			Label:  "WORKS_FOR",
			Source: "",
			Target: "",
		}
		results := l.LintRelationshipType(rel)
		// Should have two WN4013 errors
		count := 0
		for _, r := range results {
			if r.Rule == "WN4013" {
				count++
			}
		}
		if count != 2 {
			t.Errorf("expected 2 WN4013 errors, got %d", count)
		}
	})
}

// Test Linter configuration methods
func TestLinter_Configuration(t *testing.T) {
	t.Run("WithMaxInlineProperties", func(t *testing.T) {
		l := NewLinter().WithMaxInlineProperties(10)
		if l.maxInlineProperties != 10 {
			t.Errorf("maxInlineProperties = %d, want 10", l.maxInlineProperties)
		}
	})

	t.Run("WithMaxNestingDepth", func(t *testing.T) {
		l := NewLinter().WithMaxNestingDepth(5)
		if l.maxNestingDepth != 5 {
			t.Errorf("maxNestingDepth = %d, want 5", l.maxNestingDepth)
		}
	})

	t.Run("chained configuration", func(t *testing.T) {
		l := NewLinter().
			WithMaxInlineProperties(8).
			WithMaxNestingDepth(4)
		if l.maxInlineProperties != 8 {
			t.Errorf("maxInlineProperties = %d, want 8", l.maxInlineProperties)
		}
		if l.maxNestingDepth != 4 {
			t.Errorf("maxNestingDepth = %d, want 4", l.maxNestingDepth)
		}
	})
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
