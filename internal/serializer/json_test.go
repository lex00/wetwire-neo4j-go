package serializer

import (
	"encoding/json"
	"testing"

	"github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"
)

func TestNewJSONSerializer(t *testing.T) {
	s := NewJSONSerializer()
	if s == nil {
		t.Fatal("NewJSONSerializer returned nil")
	}
}

func TestJSONSerializer_SerializeNodeType(t *testing.T) {
	s := NewJSONSerializer()
	node := &schema.NodeType{
		Label:       "Person",
		Description: "A person in the system",
		Properties: []schema.Property{
			{Name: "id", Type: schema.STRING, Required: true, Unique: true},
			{Name: "name", Type: schema.STRING, Required: true},
			{Name: "age", Type: schema.INTEGER, Description: "Age in years"},
		},
		Constraints: []schema.Constraint{
			{Name: "person_key", Type: schema.NODE_KEY, Properties: []string{"id"}},
		},
		Indexes: []schema.Index{
			{Name: "person_name_idx", Type: schema.BTREE, Properties: []string{"name"}},
		},
	}

	result, err := s.SerializeNodeType(node)
	if err != nil {
		t.Fatalf("SerializeNodeType failed: %v", err)
	}

	// Verify it's valid JSON
	var parsed NodeTypeJSON
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}

	if parsed.Label != "Person" {
		t.Errorf("Label = %v, want Person", parsed.Label)
	}
	if parsed.Description != "A person in the system" {
		t.Errorf("Description = %v, want 'A person in the system'", parsed.Description)
	}
	if len(parsed.Properties) != 3 {
		t.Errorf("Properties count = %v, want 3", len(parsed.Properties))
	}
	if len(parsed.Constraints) != 1 {
		t.Errorf("Constraints count = %v, want 1", len(parsed.Constraints))
	}
	if len(parsed.Indexes) != 1 {
		t.Errorf("Indexes count = %v, want 1", len(parsed.Indexes))
	}
}

func TestJSONSerializer_SerializeNodeType_Properties(t *testing.T) {
	s := NewJSONSerializer()
	node := &schema.NodeType{
		Label: "Person",
		Properties: []schema.Property{
			{Name: "id", Type: schema.STRING, Required: true, Unique: true, Description: "Unique ID"},
			{Name: "score", Type: schema.FLOAT, DefaultValue: 0.0},
		},
	}

	result, err := s.SerializeNodeType(node)
	if err != nil {
		t.Fatalf("SerializeNodeType failed: %v", err)
	}

	var parsed NodeTypeJSON
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}

	// Check first property
	if parsed.Properties[0].Name != "id" {
		t.Errorf("Property name = %v, want id", parsed.Properties[0].Name)
	}
	if parsed.Properties[0].Type != "STRING" {
		t.Errorf("Property type = %v, want STRING", parsed.Properties[0].Type)
	}
	if !parsed.Properties[0].Required {
		t.Error("Property should be required")
	}
	if !parsed.Properties[0].Unique {
		t.Error("Property should be unique")
	}
	if parsed.Properties[0].Description != "Unique ID" {
		t.Errorf("Property description = %v, want 'Unique ID'", parsed.Properties[0].Description)
	}

	// Check second property with default
	if parsed.Properties[1].DefaultValue != 0.0 {
		t.Errorf("Property defaultValue = %v, want 0.0", parsed.Properties[1].DefaultValue)
	}
}

func TestJSONSerializer_SerializeRelationshipType(t *testing.T) {
	s := NewJSONSerializer()
	rel := &schema.RelationshipType{
		Label:       "WORKS_FOR",
		Source:      "Person",
		Target:      "Company",
		Cardinality: schema.MANY_TO_ONE,
		Description: "Employment relationship",
		Properties: []schema.Property{
			{Name: "since", Type: schema.DATE, Required: true},
			{Name: "role", Type: schema.STRING},
		},
		Constraints: []schema.Constraint{
			{Name: "works_for_since_required", Type: schema.EXISTS, Properties: []string{"since"}},
		},
	}

	result, err := s.SerializeRelationshipType(rel)
	if err != nil {
		t.Fatalf("SerializeRelationshipType failed: %v", err)
	}

	var parsed RelationshipTypeJSON
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}

	if parsed.Label != "WORKS_FOR" {
		t.Errorf("Label = %v, want WORKS_FOR", parsed.Label)
	}
	if parsed.Source != "Person" {
		t.Errorf("Source = %v, want Person", parsed.Source)
	}
	if parsed.Target != "Company" {
		t.Errorf("Target = %v, want Company", parsed.Target)
	}
	if parsed.Cardinality != "MANY_TO_ONE" {
		t.Errorf("Cardinality = %v, want MANY_TO_ONE", parsed.Cardinality)
	}
	if len(parsed.Properties) != 2 {
		t.Errorf("Properties count = %v, want 2", len(parsed.Properties))
	}
	if len(parsed.Constraints) != 1 {
		t.Errorf("Constraints count = %v, want 1", len(parsed.Constraints))
	}
}

func TestJSONSerializer_SerializeAll(t *testing.T) {
	s := NewJSONSerializer()

	nodeTypes := []*schema.NodeType{
		{Label: "Person", Properties: []schema.Property{{Name: "id", Type: schema.STRING}}},
		{Label: "Company", Properties: []schema.Property{{Name: "name", Type: schema.STRING}}},
	}

	relTypes := []*schema.RelationshipType{
		{Label: "WORKS_FOR", Source: "Person", Target: "Company"},
	}

	result, err := s.SerializeAll(nodeTypes, relTypes)
	if err != nil {
		t.Fatalf("SerializeAll failed: %v", err)
	}

	var parsed SchemaJSON
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}

	if len(parsed.NodeTypes) != 2 {
		t.Errorf("NodeTypes count = %v, want 2", len(parsed.NodeTypes))
	}
	if len(parsed.RelationshipTypes) != 1 {
		t.Errorf("RelationshipTypes count = %v, want 1", len(parsed.RelationshipTypes))
	}

	// Verify order
	if parsed.NodeTypes[0].Label != "Person" {
		t.Errorf("First node type = %v, want Person", parsed.NodeTypes[0].Label)
	}
	if parsed.NodeTypes[1].Label != "Company" {
		t.Errorf("Second node type = %v, want Company", parsed.NodeTypes[1].Label)
	}
}

func TestJSONSerializer_SerializeAll_Empty(t *testing.T) {
	s := NewJSONSerializer()

	result, err := s.SerializeAll(nil, nil)
	if err != nil {
		t.Fatalf("SerializeAll failed: %v", err)
	}

	var parsed SchemaJSON
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}

	if len(parsed.NodeTypes) != 0 {
		t.Errorf("NodeTypes should be empty, got %d", len(parsed.NodeTypes))
	}
	if len(parsed.RelationshipTypes) != 0 {
		t.Errorf("RelationshipTypes should be empty, got %d", len(parsed.RelationshipTypes))
	}
}

func TestJSONSerializer_NodeTypeToMap(t *testing.T) {
	s := NewJSONSerializer()
	node := &schema.NodeType{
		Label:       "Person",
		Description: "A person",
		Properties: []schema.Property{
			{Name: "id", Type: schema.STRING, Required: true, Unique: true},
			{Name: "name", Type: schema.STRING, Description: "Full name"},
		},
	}

	result := s.NodeTypeToMap(node)

	if result["label"] != "Person" {
		t.Errorf("label = %v, want Person", result["label"])
	}
	if result["description"] != "A person" {
		t.Errorf("description = %v, want 'A person'", result["description"])
	}

	props, ok := result["properties"].([]map[string]any)
	if !ok {
		t.Fatal("properties is not []map[string]any")
	}
	if len(props) != 2 {
		t.Errorf("properties count = %v, want 2", len(props))
	}

	// Check first property
	if props[0]["name"] != "id" {
		t.Errorf("first property name = %v, want id", props[0]["name"])
	}
	if props[0]["required"] != true {
		t.Error("first property should be required")
	}
	if props[0]["unique"] != true {
		t.Error("first property should be unique")
	}
}

func TestJSONSerializer_NodeTypeToMap_OmitsEmptyFields(t *testing.T) {
	s := NewJSONSerializer()
	node := &schema.NodeType{
		Label: "Person",
	}

	result := s.NodeTypeToMap(node)

	if _, exists := result["description"]; exists {
		t.Error("description should be omitted when empty")
	}
	if _, exists := result["properties"]; exists {
		t.Error("properties should be omitted when empty")
	}
}

func TestJSONSerializer_RelationshipTypeToMap(t *testing.T) {
	s := NewJSONSerializer()
	rel := &schema.RelationshipType{
		Label:       "WORKS_FOR",
		Source:      "Person",
		Target:      "Company",
		Cardinality: schema.MANY_TO_ONE,
		Description: "Employment",
		Properties: []schema.Property{
			{Name: "since", Type: schema.DATE, Required: true},
		},
	}

	result := s.RelationshipTypeToMap(rel)

	if result["label"] != "WORKS_FOR" {
		t.Errorf("label = %v, want WORKS_FOR", result["label"])
	}
	if result["source"] != "Person" {
		t.Errorf("source = %v, want Person", result["source"])
	}
	if result["target"] != "Company" {
		t.Errorf("target = %v, want Company", result["target"])
	}
	if result["cardinality"] != "MANY_TO_ONE" {
		t.Errorf("cardinality = %v, want MANY_TO_ONE", result["cardinality"])
	}

	props, ok := result["properties"].([]map[string]any)
	if !ok {
		t.Fatal("properties is not []map[string]any")
	}
	if len(props) != 1 {
		t.Errorf("properties count = %v, want 1", len(props))
	}
}

func TestJSONSerializer_VectorIndex(t *testing.T) {
	s := NewJSONSerializer()
	node := &schema.NodeType{
		Label: "Document",
		Indexes: []schema.Index{
			{
				Name:       "embedding_idx",
				Type:       schema.VECTOR,
				Properties: []string{"embedding"},
				Options: map[string]any{
					"dimensions":          384,
					"similarity_function": "cosine",
				},
			},
		},
	}

	result, err := s.SerializeNodeType(node)
	if err != nil {
		t.Fatalf("SerializeNodeType failed: %v", err)
	}

	var parsed NodeTypeJSON
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}

	if len(parsed.Indexes) != 1 {
		t.Fatalf("Indexes count = %v, want 1", len(parsed.Indexes))
	}

	idx := parsed.Indexes[0]
	if idx.Type != "VECTOR" {
		t.Errorf("Index type = %v, want VECTOR", idx.Type)
	}
	if idx.Options["dimensions"] != float64(384) { // JSON numbers are float64
		t.Errorf("Index dimensions = %v, want 384", idx.Options["dimensions"])
	}
	if idx.Options["similarity_function"] != "cosine" {
		t.Errorf("Index similarity_function = %v, want cosine", idx.Options["similarity_function"])
	}
}
