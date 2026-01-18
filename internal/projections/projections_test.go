package projections

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNativeProjection_Interface(t *testing.T) {
	p := &NativeProjection{
		BaseProjection: BaseProjection{
			Name:      "social_graph",
			GraphName: "social_graph",
		},
		NodeLabels:        []string{"Person", "Company"},
		RelationshipTypes: []string{"KNOWS", "WORKS_AT"},
	}

	if p.ProjectionName() != "social_graph" {
		t.Errorf("ProjectionName() = %v, want social_graph", p.ProjectionName())
	}
	if p.ProjectionType() != Native {
		t.Errorf("ProjectionType() = %v, want Native", p.ProjectionType())
	}
}

func TestCypherProjection_Interface(t *testing.T) {
	p := &CypherProjection{
		BaseProjection: BaseProjection{
			Name:      "custom_graph",
			GraphName: "custom_graph",
		},
		NodeQuery:         "MATCH (n:Person) RETURN id(n) AS id, labels(n) AS labels",
		RelationshipQuery: "MATCH (a:Person)-[r:KNOWS]->(b:Person) RETURN id(a) AS source, id(b) AS target, type(r) AS type",
	}

	if p.ProjectionType() != Cypher {
		t.Errorf("ProjectionType() = %v, want Cypher", p.ProjectionType())
	}
}

func TestDataFrameProjection_Interface(t *testing.T) {
	p := &DataFrameProjection{
		BaseProjection: BaseProjection{
			Name:      "aura_graph",
			GraphName: "aura_graph",
		},
		NodeDataFrames: []NodeDataFrame{
			{Label: "Person", Properties: []string{"name", "age"}},
		},
	}

	if p.ProjectionType() != DataFrame {
		t.Errorf("ProjectionType() = %v, want DataFrame", p.ProjectionType())
	}
}

func TestNativeProjection_NodeProjections(t *testing.T) {
	p := &NativeProjection{
		BaseProjection: BaseProjection{Name: "test"},
		NodeLabels:     []string{"Person", "Company"},
	}

	projections := p.GetNodeProjections()
	if len(projections) != 2 {
		t.Errorf("GetNodeProjections() count = %v, want 2", len(projections))
	}
}

func TestNativeProjection_RelationshipProjections(t *testing.T) {
	p := &NativeProjection{
		BaseProjection:    BaseProjection{Name: "test"},
		RelationshipTypes: []string{"KNOWS", "WORKS_AT"},
	}

	projections := p.GetRelationshipProjections()
	if len(projections) != 2 {
		t.Errorf("GetRelationshipProjections() count = %v, want 2", len(projections))
	}
}

func TestNodeProjection_Configuration(t *testing.T) {
	np := NodeProjection{
		Label:      "Person",
		Properties: []string{"name", "age"},
	}

	if np.Label != "Person" {
		t.Errorf("Label = %v, want Person", np.Label)
	}
	if len(np.Properties) != 2 {
		t.Errorf("Properties count = %v, want 2", len(np.Properties))
	}
}

func TestRelationshipProjection_Configuration(t *testing.T) {
	rp := RelationshipProjection{
		Type:        "KNOWS",
		Orientation: Undirected,
		Aggregation: Sum,
		Properties:  []string{"weight", "since"},
	}

	if rp.Type != "KNOWS" {
		t.Errorf("Type = %v, want KNOWS", rp.Type)
	}
	if rp.Orientation != Undirected {
		t.Errorf("Orientation = %v, want Undirected", rp.Orientation)
	}
}

func TestNewProjectionSerializer(t *testing.T) {
	s := NewProjectionSerializer()
	if s == nil {
		t.Fatal("NewProjectionSerializer returned nil")
	}
	if s.templates == nil {
		t.Error("templates is nil")
	}
}

func TestProjectionSerializer_ToCypher_NativeSimple(t *testing.T) {
	s := NewProjectionSerializer()
	p := &NativeProjection{
		BaseProjection: BaseProjection{
			Name:      "social_graph",
			GraphName: "social_graph",
		},
		NodeLabels:        []string{"Person"},
		RelationshipTypes: []string{"KNOWS"},
	}

	result, err := s.ToCypher(p)
	if err != nil {
		t.Fatalf("ToCypher failed: %v", err)
	}

	if !strings.Contains(result, "gds.graph.project") {
		t.Errorf("expected gds.graph.project, got: %s", result)
	}
	if !strings.Contains(result, "'social_graph'") {
		t.Errorf("expected graph name, got: %s", result)
	}
	if !strings.Contains(result, "'Person'") {
		t.Errorf("expected Person node label, got: %s", result)
	}
	if !strings.Contains(result, "'KNOWS'") {
		t.Errorf("expected KNOWS relationship, got: %s", result)
	}
}

func TestProjectionSerializer_ToCypher_NativeWithProperties(t *testing.T) {
	s := NewProjectionSerializer()
	p := &NativeProjection{
		BaseProjection: BaseProjection{
			Name:      "weighted_graph",
			GraphName: "weighted_graph",
		},
		NodeProjections: []NodeProjection{
			{Label: "Person", Properties: []string{"age", "score"}},
		},
		RelationshipProjections: []RelationshipProjection{
			{Type: "KNOWS", Properties: []string{"weight"}, Orientation: Undirected},
		},
	}

	result, err := s.ToCypher(p)
	if err != nil {
		t.Fatalf("ToCypher failed: %v", err)
	}

	if !strings.Contains(result, "age") {
		t.Errorf("expected age property, got: %s", result)
	}
	if !strings.Contains(result, "weight") {
		t.Errorf("expected weight property, got: %s", result)
	}
	if !strings.Contains(result, "UNDIRECTED") {
		t.Errorf("expected UNDIRECTED orientation, got: %s", result)
	}
}

func TestProjectionSerializer_ToCypher_Cypher(t *testing.T) {
	s := NewProjectionSerializer()
	p := &CypherProjection{
		BaseProjection: BaseProjection{
			Name:      "custom_graph",
			GraphName: "custom_graph",
		},
		NodeQuery:         "MATCH (n:Person) RETURN id(n) AS id",
		RelationshipQuery: "MATCH (a)-[r:KNOWS]->(b) RETURN id(a) AS source, id(b) AS target",
	}

	result, err := s.ToCypher(p)
	if err != nil {
		t.Fatalf("ToCypher failed: %v", err)
	}

	if !strings.Contains(result, "gds.graph.project.cypher") {
		t.Errorf("expected gds.graph.project.cypher, got: %s", result)
	}
	if !strings.Contains(result, "MATCH (n:Person)") {
		t.Errorf("expected node query, got: %s", result)
	}
	if !strings.Contains(result, "MATCH (a)-[r:KNOWS]->(b)") {
		t.Errorf("expected relationship query, got: %s", result)
	}
}

func TestProjectionSerializer_ToJSON_Native(t *testing.T) {
	s := NewProjectionSerializer()
	p := &NativeProjection{
		BaseProjection: BaseProjection{
			Name:      "test_graph",
			GraphName: "test_graph",
		},
		NodeLabels:        []string{"Person", "Company"},
		RelationshipTypes: []string{"KNOWS", "WORKS_AT"},
	}

	result, err := s.ToJSON(p)
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}

	if parsed["name"] != "test_graph" {
		t.Errorf("name = %v, want test_graph", parsed["name"])
	}
	if parsed["projectionType"] != "Native" {
		t.Errorf("projectionType = %v, want Native", parsed["projectionType"])
	}

	nodeLabels, ok := parsed["nodeLabels"].([]any)
	if !ok || len(nodeLabels) != 2 {
		t.Error("expected 2 node labels")
	}
}

func TestProjectionSerializer_ToMap(t *testing.T) {
	s := NewProjectionSerializer()
	p := &CypherProjection{
		BaseProjection: BaseProjection{
			Name:      "custom",
			GraphName: "custom",
		},
		NodeQuery: "MATCH (n) RETURN id(n) AS id",
	}

	result := s.ToMap(p)

	if result["name"] != "custom" {
		t.Errorf("name = %v, want custom", result["name"])
	}
	if result["nodeQuery"] != "MATCH (n) RETURN id(n) AS id" {
		t.Errorf("nodeQuery = %v, want MATCH query", result["nodeQuery"])
	}
}

func TestOrientation_Values(t *testing.T) {
	tests := []struct {
		o    Orientation
		want string
	}{
		{Natural, "NATURAL"},
		{Reverse, "REVERSE"},
		{Undirected, "UNDIRECTED"},
	}

	for _, tt := range tests {
		if string(tt.o) != tt.want {
			t.Errorf("Orientation = %v, want %v", string(tt.o), tt.want)
		}
	}
}

func TestAggregation_Values(t *testing.T) {
	tests := []struct {
		a    Aggregation
		want string
	}{
		{None, "NONE"},
		{Sum, "SUM"},
		{Min, "MIN"},
		{Max, "MAX"},
		{Single, "SINGLE"},
		{Count, "COUNT"},
	}

	for _, tt := range tests {
		if string(tt.a) != tt.want {
			t.Errorf("Aggregation = %v, want %v", string(tt.a), tt.want)
		}
	}
}

func TestProjectionSerializer_Drop(t *testing.T) {
	s := NewProjectionSerializer()
	result := s.DropGraph("my_graph")

	if !strings.Contains(result, "gds.graph.drop") {
		t.Errorf("expected gds.graph.drop, got: %s", result)
	}
	if !strings.Contains(result, "'my_graph'") {
		t.Errorf("expected graph name, got: %s", result)
	}
}

func TestProjectionSerializer_Exists(t *testing.T) {
	s := NewProjectionSerializer()
	result := s.GraphExists("my_graph")

	if !strings.Contains(result, "gds.graph.exists") {
		t.Errorf("expected gds.graph.exists, got: %s", result)
	}
}

func TestProjectionSerializer_List(t *testing.T) {
	s := NewProjectionSerializer()
	result := s.ListGraphs()

	if !strings.Contains(result, "gds.graph.list") {
		t.Errorf("expected gds.graph.list, got: %s", result)
	}
}

func TestProjection_ImplementsInterface(t *testing.T) {
	// Verify all projections implement the Projection interface
	var _ Projection = &NativeProjection{}
	var _ Projection = &CypherProjection{}
	var _ Projection = &DataFrameProjection{}
}
