package lint

import (
	"testing"

	"github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"
)

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
