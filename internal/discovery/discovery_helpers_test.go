package discovery

import (
	"testing"
)

func TestIsPrimitiveType(t *testing.T) {
	primitives := []string{"bool", "string", "int", "int64", "float64", "byte", "error", "any"}
	for _, p := range primitives {
		if !isPrimitiveType(p) {
			t.Errorf("%s should be primitive", p)
		}
	}

	nonPrimitives := []string{"Person", "MyType", "NodeType"}
	for _, p := range nonPrimitives {
		if isPrimitiveType(p) {
			t.Errorf("%s should not be primitive", p)
		}
	}
}

func TestIsValidIdentifier(t *testing.T) {
	valid := []string{"foo", "Foo", "_foo", "foo123", "FooBar"}
	for _, v := range valid {
		if !isValidIdentifier(v) {
			t.Errorf("%s should be valid", v)
		}
	}

	invalid := []string{"", "123foo", "foo-bar", "foo.bar"}
	for _, v := range invalid {
		if isValidIdentifier(v) {
			t.Errorf("%s should be invalid", v)
		}
	}
}

func TestResourceKind_Constants(t *testing.T) {
	tests := []struct {
		kind ResourceKind
		want string
	}{
		{KindNodeType, "NodeType"},
		{KindRelationshipType, "RelationshipType"},
		{KindAlgorithm, "Algorithm"},
		{KindPipeline, "Pipeline"},
		{KindRetriever, "Retriever"},
	}

	for _, tt := range tests {
		if string(tt.kind) != tt.want {
			t.Errorf("ResourceKind = %v, want %v", tt.kind, tt.want)
		}
	}
}
