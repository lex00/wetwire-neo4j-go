package schema

import (
	"strings"
	"testing"
)

func TestValidationError(t *testing.T) {
	t.Run("with field", func(t *testing.T) {
		err := ValidationError{
			Resource: "Person",
			Field:    "Label",
			Message:  "label is required",
		}
		got := err.Error()
		want := "Person.Label: label is required"
		if got != want {
			t.Errorf("Error() = %v, want %v", got, want)
		}
	})

	t.Run("without field", func(t *testing.T) {
		err := ValidationError{
			Resource: "Person",
			Message:  "invalid node type",
		}
		got := err.Error()
		want := "Person: invalid node type"
		if got != want {
			t.Errorf("Error() = %v, want %v", got, want)
		}
	})
}

func TestValidator_ValidateNodeType(t *testing.T) {
	t.Run("valid node type", func(t *testing.T) {
		v := NewValidator()
		node := &NodeType{
			Label: "Person",
			Properties: []Property{
				{Name: "name", Type: STRING, Required: true},
				{Name: "age", Type: INTEGER},
			},
		}
		result := v.ValidateNodeType(node)
		if !result.Valid {
			t.Errorf("expected valid, got errors: %v", result.Errors)
		}
	})

	t.Run("missing label", func(t *testing.T) {
		v := NewValidator()
		node := &NodeType{
			Properties: []Property{
				{Name: "name", Type: STRING},
			},
		}
		result := v.ValidateNodeType(node)
		if result.Valid {
			t.Error("expected invalid for missing label")
		}
		if len(result.Errors) == 0 {
			t.Error("expected errors for missing label")
		}
	})

	t.Run("non-PascalCase label warning", func(t *testing.T) {
		v := NewValidator()
		node := &NodeType{
			Label: "PERSON",
			Properties: []Property{
				{Name: "name", Type: STRING},
			},
		}
		result := v.ValidateNodeType(node)
		if len(result.Warnings) == 0 {
			t.Error("expected warning for non-PascalCase label")
		}
	})

	t.Run("duplicate property names", func(t *testing.T) {
		v := NewValidator()
		node := &NodeType{
			Label: "Person",
			Properties: []Property{
				{Name: "name", Type: STRING},
				{Name: "name", Type: STRING}, // duplicate
			},
		}
		result := v.ValidateNodeType(node)
		if result.Valid {
			t.Error("expected invalid for duplicate property names")
		}
		found := false
		for _, err := range result.Errors {
			if strings.Contains(err.Message, "duplicate property name") {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected error about duplicate property name")
		}
	})

	t.Run("constraint references unknown property", func(t *testing.T) {
		v := NewValidator()
		node := &NodeType{
			Label: "Person",
			Properties: []Property{
				{Name: "name", Type: STRING},
			},
			Constraints: []Constraint{
				{Name: "unknown_constraint", Type: UNIQUE, Properties: []string{"email"}},
			},
		}
		result := v.ValidateNodeType(node)
		if result.Valid {
			t.Error("expected invalid for constraint referencing unknown property")
		}
		found := false
		for _, err := range result.Errors {
			if strings.Contains(err.Message, "unknown property") {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected error about unknown property in constraint")
		}
	})

	t.Run("index references unknown property", func(t *testing.T) {
		v := NewValidator()
		node := &NodeType{
			Label: "Person",
			Properties: []Property{
				{Name: "name", Type: STRING},
			},
			Indexes: []Index{
				{Name: "unknown_idx", Type: BTREE, Properties: []string{"unknown"}},
			},
		}
		result := v.ValidateNodeType(node)
		if result.Valid {
			t.Error("expected invalid for index referencing unknown property")
		}
	})

	t.Run("property without name", func(t *testing.T) {
		v := NewValidator()
		node := &NodeType{
			Label: "Person",
			Properties: []Property{
				{Type: STRING}, // missing name
			},
		}
		result := v.ValidateNodeType(node)
		if result.Valid {
			t.Error("expected invalid for property without name")
		}
	})

	t.Run("property without type", func(t *testing.T) {
		v := NewValidator()
		node := &NodeType{
			Label: "Person",
			Properties: []Property{
				{Name: "name"}, // missing type
			},
		}
		result := v.ValidateNodeType(node)
		if result.Valid {
			t.Error("expected invalid for property without type")
		}
	})

	t.Run("invalid property type", func(t *testing.T) {
		v := NewValidator()
		node := &NodeType{
			Label: "Person",
			Properties: []Property{
				{Name: "name", Type: PropertyType("INVALID")},
			},
		}
		result := v.ValidateNodeType(node)
		if result.Valid {
			t.Error("expected invalid for invalid property type")
		}
	})
}

func TestValidator_ValidateRelationshipType(t *testing.T) {
	t.Run("valid relationship type", func(t *testing.T) {
		v := NewValidator()
		rel := &RelationshipType{
			Label:       "WORKS_FOR",
			Source:      "Person",
			Target:      "Company",
			Cardinality: MANY_TO_ONE,
			Properties: []Property{
				{Name: "since", Type: DATE},
			},
		}
		result := v.ValidateRelationshipType(rel)
		if !result.Valid {
			t.Errorf("expected valid, got errors: %v", result.Errors)
		}
	})

	t.Run("missing label", func(t *testing.T) {
		v := NewValidator()
		rel := &RelationshipType{
			Source: "Person",
			Target: "Company",
		}
		result := v.ValidateRelationshipType(rel)
		if result.Valid {
			t.Error("expected invalid for missing label")
		}
	})

	t.Run("non-SCREAMING_SNAKE_CASE label warning", func(t *testing.T) {
		v := NewValidator()
		rel := &RelationshipType{
			Label:  "worksFor",
			Source: "Person",
			Target: "Company",
		}
		result := v.ValidateRelationshipType(rel)
		if len(result.Warnings) == 0 {
			t.Error("expected warning for non-SCREAMING_SNAKE_CASE label")
		}
	})

	t.Run("missing source", func(t *testing.T) {
		v := NewValidator()
		rel := &RelationshipType{
			Label:  "WORKS_FOR",
			Target: "Company",
		}
		result := v.ValidateRelationshipType(rel)
		if result.Valid {
			t.Error("expected invalid for missing source")
		}
	})

	t.Run("missing target", func(t *testing.T) {
		v := NewValidator()
		rel := &RelationshipType{
			Label:  "WORKS_FOR",
			Source: "Person",
		}
		result := v.ValidateRelationshipType(rel)
		if result.Valid {
			t.Error("expected invalid for missing target")
		}
	})

	t.Run("source node type not found", func(t *testing.T) {
		v := NewValidator()
		// Register only Company, not Person
		v.Register(&NodeType{Label: "Company"})
		rel := &RelationshipType{
			Label:  "WORKS_FOR",
			Source: "Person",
			Target: "Company",
		}
		result := v.ValidateRelationshipType(rel)
		if result.Valid {
			t.Error("expected invalid when source node type not found")
		}
		found := false
		for _, err := range result.Errors {
			if strings.Contains(err.Message, "source node type not found") {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected error about source node type not found")
		}
	})

	t.Run("target node type not found", func(t *testing.T) {
		v := NewValidator()
		v.Register(&NodeType{Label: "Person"})
		rel := &RelationshipType{
			Label:  "WORKS_FOR",
			Source: "Person",
			Target: "Company",
		}
		result := v.ValidateRelationshipType(rel)
		if result.Valid {
			t.Error("expected invalid when target node type not found")
		}
	})

	t.Run("invalid cardinality", func(t *testing.T) {
		v := NewValidator()
		rel := &RelationshipType{
			Label:       "WORKS_FOR",
			Source:      "Person",
			Target:      "Company",
			Cardinality: Cardinality("INVALID"),
		}
		result := v.ValidateRelationshipType(rel)
		if result.Valid {
			t.Error("expected invalid for invalid cardinality")
		}
	})

	t.Run("duplicate property names", func(t *testing.T) {
		v := NewValidator()
		rel := &RelationshipType{
			Label:  "WORKS_FOR",
			Source: "Person",
			Target: "Company",
			Properties: []Property{
				{Name: "since", Type: DATE},
				{Name: "since", Type: DATE}, // duplicate
			},
		}
		result := v.ValidateRelationshipType(rel)
		if result.Valid {
			t.Error("expected invalid for duplicate property names")
		}
	})

	t.Run("constraint references unknown property", func(t *testing.T) {
		v := NewValidator()
		rel := &RelationshipType{
			Label:  "WORKS_FOR",
			Source: "Person",
			Target: "Company",
			Properties: []Property{
				{Name: "since", Type: DATE},
			},
			Constraints: []Constraint{
				{Name: "rel_key", Type: REL_KEY, Properties: []string{"unknown"}},
			},
		}
		result := v.ValidateRelationshipType(rel)
		if result.Valid {
			t.Error("expected invalid for constraint referencing unknown property")
		}
	})
}

func TestValidator_ValidateAll(t *testing.T) {
	t.Run("all valid", func(t *testing.T) {
		v := NewValidator()
		v.Register(
			&NodeType{
				Label: "Person",
				Properties: []Property{
					{Name: "name", Type: STRING},
				},
			},
			&NodeType{
				Label: "Company",
				Properties: []Property{
					{Name: "name", Type: STRING},
				},
			},
			&RelationshipType{
				Label:  "WORKS_FOR",
				Source: "Person",
				Target: "Company",
			},
		)
		result := v.ValidateAll()
		if !result.Valid {
			t.Errorf("expected valid, got errors: %v", result.Errors)
		}
	})

	t.Run("invalid node type", func(t *testing.T) {
		v := NewValidator()
		v.Register(
			&NodeType{
				Label: "", // invalid - missing label
			},
		)
		result := v.ValidateAll()
		if result.Valid {
			t.Error("expected invalid")
		}
	})

	t.Run("invalid relationship type", func(t *testing.T) {
		v := NewValidator()
		v.Register(
			&NodeType{Label: "Person"},
			&RelationshipType{
				Label:  "WORKS_FOR",
				Source: "Person",
				Target: "Company", // Company not registered
			},
		)
		result := v.ValidateAll()
		if result.Valid {
			t.Error("expected invalid")
		}
	})
}

func TestIsPascalCase(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"Person", true},
		{"MyClass", true},
		{"XMLParser", true},
		{"A", true},
		{"person", false},   // lowercase start
		{"PERSON", false},   // all caps (not PascalCase)
		{"my_class", false}, // snake_case
		{"MY_CLASS", false}, // SCREAMING_SNAKE_CASE
		{"", false},         // empty
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := isPascalCase(tt.input); got != tt.want {
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
		{"CREATED_BY", true},
		{"HAS", true},
		{"HAS_A", true},
		{"A", true},
		{"worksFor", false},   // camelCase
		{"works_for", false},  // snake_case
		{"WorksFor", false},   // PascalCase
		{"WORKS__FOR", false}, // double underscore
		{"_WORKS_FOR", false}, // leading underscore
		{"WORKS_FOR_", false}, // trailing underscore
		{"", false},           // empty
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := isScreamingSnakeCase(tt.input); got != tt.want {
				t.Errorf("isScreamingSnakeCase(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidator_Register(t *testing.T) {
	v := NewValidator()

	node := &NodeType{Label: "Person"}
	rel := &RelationshipType{Label: "KNOWS", Source: "Person", Target: "Person"}

	v.Register(node, rel)

	// Verify registration by checking if validation uses registered types
	result := v.ValidateRelationshipType(&RelationshipType{
		Label:  "FOLLOWS",
		Source: "Person",  // This should be found
		Target: "Unknown", // This should not be found
	})

	// Should have error for Unknown but not for Person
	hasUnknownError := false
	hasPersonError := false
	for _, err := range result.Errors {
		if strings.Contains(err.Message, "Unknown") {
			hasUnknownError = true
		}
		if strings.Contains(err.Message, "Person") {
			hasPersonError = true
		}
	}

	if !hasUnknownError {
		t.Error("expected error for unknown target type")
	}
	if hasPersonError {
		t.Error("should not have error for registered Person type")
	}
}
