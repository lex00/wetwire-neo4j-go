package lint

import (
	"testing"

	"github.com/lex00/wetwire-neo4j-go/internal/algorithms"
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

func TestLinter_LintAll(t *testing.T) {
	l := NewLinter()

	// Create some resources with issues
	algo := &algorithms.PageRank{DampingFactor: 1.5}   // Invalid
	node := &schema.NodeType{Label: "person"}          // Should be PascalCase
	rel := &schema.RelationshipType{Label: "worksFor"} // Should be SCREAMING_SNAKE

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

// Helper function used across test files
func containsRule(results []LintResult, rule string) bool {
	for _, r := range results {
		if r.Rule == rule {
			return true
		}
	}
	return false
}

// TestLinter_DisabledRules tests that disabled rules are properly skipped
func TestLinter_DisabledRules(t *testing.T) {
	l := NewLinter()

	// Create a PageRank with invalid damping factor (triggers WN4001)
	algo := &algorithms.PageRank{DampingFactor: 1.5}

	// Run lint without disabling any rules
	results := l.LintAlgorithm(algo)
	if !containsRule(results, "WN4001") {
		t.Error("expected WN4001 to be present when not disabled")
	}

	// Now test with DisabledRules - use LintAllWithOptions
	resultsFiltered := l.LintAllWithOptions([]any{algo}, LintOptions{
		DisabledRules: []string{"WN4001"},
	})
	if containsRule(resultsFiltered, "WN4001") {
		t.Error("WN4001 should be filtered out when disabled")
	}
}

// TestLinter_DisabledRules_Multiple tests disabling multiple rules
func TestLinter_DisabledRules_Multiple(t *testing.T) {
	l := NewLinter()

	// Create a node with lowercase label (triggers WN4052)
	node := &schema.NodeType{Label: "person"} // lowercase - violates WN4052

	// Create a PageRank with invalid damping factor (triggers WN4001)
	algo := &algorithms.PageRank{DampingFactor: 1.5}

	// Run lint without disabling
	results := l.LintAll([]any{algo, node})
	hasWN4001 := containsRule(results, "WN4001")
	hasWN4052 := containsRule(results, "WN4052")

	if !hasWN4001 {
		t.Error("expected WN4001 to be present")
	}
	if !hasWN4052 {
		t.Error("expected WN4052 to be present")
	}

	// Disable both rules
	filteredResults := l.LintAllWithOptions([]any{algo, node}, LintOptions{
		DisabledRules: []string{"WN4001", "WN4052"},
	})

	if containsRule(filteredResults, "WN4001") {
		t.Error("WN4001 should be filtered out")
	}
	if containsRule(filteredResults, "WN4052") {
		t.Error("WN4052 should be filtered out")
	}
}

// TestLintOptions_Fields tests that LintOptions struct has all expected fields
func TestLintOptions_Fields(t *testing.T) {
	opts := LintOptions{
		DisabledRules: []string{"WN4001", "WN4052"},
		Fix:           true,
	}

	if len(opts.DisabledRules) != 2 {
		t.Errorf("expected 2 disabled rules, got %d", len(opts.DisabledRules))
	}
	if !opts.Fix {
		t.Error("expected Fix to be true")
	}
}
