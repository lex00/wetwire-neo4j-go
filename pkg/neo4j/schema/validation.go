package schema

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidationError represents a validation error with context.
type ValidationError struct {
	// Resource is the resource name where the error occurred.
	Resource string
	// Field is the field name where the error occurred.
	Field string
	// Message is the error message.
	Message string
}

func (e ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s.%s: %s", e.Resource, e.Field, e.Message)
	}
	return fmt.Sprintf("%s: %s", e.Resource, e.Message)
}

// ValidationResult contains the results of validating a schema.
type ValidationResult struct {
	// Valid is true if no errors were found.
	Valid bool
	// Errors is a list of validation errors.
	Errors []ValidationError
	// Warnings is a list of validation warnings (non-fatal issues).
	Warnings []ValidationError
}

// Validator validates schema definitions.
type Validator struct {
	// nodeTypes maps label to NodeType for reference validation.
	nodeTypes map[string]*NodeType
	// relationshipTypes maps label to RelationshipType.
	relationshipTypes map[string]*RelationshipType
}

// NewValidator creates a new schema validator.
func NewValidator() *Validator {
	return &Validator{
		nodeTypes:         make(map[string]*NodeType),
		relationshipTypes: make(map[string]*RelationshipType),
	}
}

// Register registers resources for cross-reference validation.
func (v *Validator) Register(resources ...Resource) {
	for _, r := range resources {
		switch res := r.(type) {
		case *NodeType:
			v.nodeTypes[res.Label] = res
		case *RelationshipType:
			v.relationshipTypes[res.Label] = res
		}
	}
}

// ValidateNodeType validates a NodeType definition.
func (v *Validator) ValidateNodeType(n *NodeType) ValidationResult {
	result := ValidationResult{Valid: true}

	// Validate label
	if n.Label == "" {
		result.Errors = append(result.Errors, ValidationError{
			Resource: "NodeType",
			Field:    "Label",
			Message:  "label is required",
		})
	} else if !isPascalCase(n.Label) {
		result.Warnings = append(result.Warnings, ValidationError{
			Resource: n.Label,
			Field:    "Label",
			Message:  "node labels should be PascalCase",
		})
	}

	// Validate properties
	propNames := make(map[string]bool)
	for i, prop := range n.Properties {
		propErrors := v.validateProperty(n.Label, prop, i)
		result.Errors = append(result.Errors, propErrors...)

		if propNames[prop.Name] {
			result.Errors = append(result.Errors, ValidationError{
				Resource: n.Label,
				Field:    fmt.Sprintf("Properties[%d].Name", i),
				Message:  fmt.Sprintf("duplicate property name: %s", prop.Name),
			})
		}
		propNames[prop.Name] = true
	}

	// Validate constraints reference existing properties
	for i, c := range n.Constraints {
		for _, propName := range c.Properties {
			if !propNames[propName] {
				result.Errors = append(result.Errors, ValidationError{
					Resource: n.Label,
					Field:    fmt.Sprintf("Constraints[%d].Properties", i),
					Message:  fmt.Sprintf("constraint references unknown property: %s", propName),
				})
			}
		}
	}

	// Validate indexes reference existing properties
	for i, idx := range n.Indexes {
		for _, propName := range idx.Properties {
			if !propNames[propName] {
				result.Errors = append(result.Errors, ValidationError{
					Resource: n.Label,
					Field:    fmt.Sprintf("Indexes[%d].Properties", i),
					Message:  fmt.Sprintf("index references unknown property: %s", propName),
				})
			}
		}
	}

	result.Valid = len(result.Errors) == 0
	return result
}

// ValidateRelationshipType validates a RelationshipType definition.
func (v *Validator) ValidateRelationshipType(r *RelationshipType) ValidationResult {
	result := ValidationResult{Valid: true}

	// Validate label
	if r.Label == "" {
		result.Errors = append(result.Errors, ValidationError{
			Resource: "RelationshipType",
			Field:    "Label",
			Message:  "label is required",
		})
	} else if !isScreamingSnakeCase(r.Label) {
		result.Warnings = append(result.Warnings, ValidationError{
			Resource: r.Label,
			Field:    "Label",
			Message:  "relationship labels should be SCREAMING_SNAKE_CASE",
		})
	}

	// Validate source
	if r.Source == "" {
		result.Errors = append(result.Errors, ValidationError{
			Resource: r.Label,
			Field:    "Source",
			Message:  "source node type is required",
		})
	} else if len(v.nodeTypes) > 0 {
		if _, ok := v.nodeTypes[r.Source]; !ok {
			result.Errors = append(result.Errors, ValidationError{
				Resource: r.Label,
				Field:    "Source",
				Message:  fmt.Sprintf("source node type not found: %s", r.Source),
			})
		}
	}

	// Validate target
	if r.Target == "" {
		result.Errors = append(result.Errors, ValidationError{
			Resource: r.Label,
			Field:    "Target",
			Message:  "target node type is required",
		})
	} else if len(v.nodeTypes) > 0 {
		if _, ok := v.nodeTypes[r.Target]; !ok {
			result.Errors = append(result.Errors, ValidationError{
				Resource: r.Label,
				Field:    "Target",
				Message:  fmt.Sprintf("target node type not found: %s", r.Target),
			})
		}
	}

	// Validate cardinality
	validCardinalities := map[Cardinality]bool{
		ONE_TO_ONE:   true,
		ONE_TO_MANY:  true,
		MANY_TO_ONE:  true,
		MANY_TO_MANY: true,
	}
	if r.Cardinality != "" && !validCardinalities[r.Cardinality] {
		result.Errors = append(result.Errors, ValidationError{
			Resource: r.Label,
			Field:    "Cardinality",
			Message:  fmt.Sprintf("invalid cardinality: %s", r.Cardinality),
		})
	}

	// Validate properties
	propNames := make(map[string]bool)
	for i, prop := range r.Properties {
		propErrors := v.validateProperty(r.Label, prop, i)
		result.Errors = append(result.Errors, propErrors...)

		if propNames[prop.Name] {
			result.Errors = append(result.Errors, ValidationError{
				Resource: r.Label,
				Field:    fmt.Sprintf("Properties[%d].Name", i),
				Message:  fmt.Sprintf("duplicate property name: %s", prop.Name),
			})
		}
		propNames[prop.Name] = true
	}

	// Validate constraints reference existing properties
	for i, c := range r.Constraints {
		for _, propName := range c.Properties {
			if !propNames[propName] {
				result.Errors = append(result.Errors, ValidationError{
					Resource: r.Label,
					Field:    fmt.Sprintf("Constraints[%d].Properties", i),
					Message:  fmt.Sprintf("constraint references unknown property: %s", propName),
				})
			}
		}
	}

	result.Valid = len(result.Errors) == 0
	return result
}

// validateProperty validates a single property definition.
func (v *Validator) validateProperty(resource string, prop Property, index int) []ValidationError {
	var errors []ValidationError

	if prop.Name == "" {
		errors = append(errors, ValidationError{
			Resource: resource,
			Field:    fmt.Sprintf("Properties[%d].Name", index),
			Message:  "property name is required",
		})
	}

	if prop.Type == "" {
		errors = append(errors, ValidationError{
			Resource: resource,
			Field:    fmt.Sprintf("Properties[%d].Type", index),
			Message:  "property type is required",
		})
	} else {
		validTypes := map[PropertyType]bool{
			STRING:       true,
			INTEGER:      true,
			FLOAT:        true,
			BOOLEAN:      true,
			DATE:         true,
			DATETIME:     true,
			POINT:        true,
			LIST_STRING:  true,
			LIST_INTEGER: true,
			LIST_FLOAT:   true,
		}
		if !validTypes[prop.Type] {
			errors = append(errors, ValidationError{
				Resource: resource,
				Field:    fmt.Sprintf("Properties[%d].Type", index),
				Message:  fmt.Sprintf("invalid property type: %s", prop.Type),
			})
		}
	}

	return errors
}

// ValidateAll validates all registered resources.
func (v *Validator) ValidateAll() ValidationResult {
	result := ValidationResult{Valid: true}

	for _, n := range v.nodeTypes {
		nodeResult := v.ValidateNodeType(n)
		result.Errors = append(result.Errors, nodeResult.Errors...)
		result.Warnings = append(result.Warnings, nodeResult.Warnings...)
	}

	for _, r := range v.relationshipTypes {
		relResult := v.ValidateRelationshipType(r)
		result.Errors = append(result.Errors, relResult.Errors...)
		result.Warnings = append(result.Warnings, relResult.Warnings...)
	}

	result.Valid = len(result.Errors) == 0
	return result
}

// isPascalCase checks if a string is in PascalCase.
func isPascalCase(s string) bool {
	if s == "" {
		return false
	}
	// Must start with uppercase letter
	if s[0] < 'A' || s[0] > 'Z' {
		return false
	}
	// Must not contain underscores or be all uppercase
	if strings.Contains(s, "_") {
		return false
	}
	// Should contain at least one lowercase letter (not all caps)
	hasLower := false
	for _, c := range s {
		if c >= 'a' && c <= 'z' {
			hasLower = true
			break
		}
	}
	return hasLower || len(s) == 1
}

// isScreamingSnakeCase checks if a string is in SCREAMING_SNAKE_CASE.
func isScreamingSnakeCase(s string) bool {
	if s == "" {
		return false
	}
	// Use regex for SCREAMING_SNAKE_CASE: all uppercase letters, numbers, and underscores
	matched, _ := regexp.MatchString(`^[A-Z][A-Z0-9]*(_[A-Z0-9]+)*$`, s)
	return matched
}
