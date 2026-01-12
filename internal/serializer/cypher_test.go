package serializer

import (
	"strings"
	"testing"

	"github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"
)

func TestNewCypherSerializer(t *testing.T) {
	s := NewCypherSerializer()
	if s == nil {
		t.Fatal("NewCypherSerializer returned nil")
	}
	if s.templates == nil {
		t.Error("templates is nil")
	}
}

func TestCypherSerializer_SerializeNodeType_UniqueConstraint(t *testing.T) {
	s := NewCypherSerializer()
	node := &schema.NodeType{
		Label: "Person",
		Properties: []schema.Property{
			{Name: "email", Type: schema.STRING, Unique: true},
		},
	}

	result, err := s.SerializeNodeType(node)
	if err != nil {
		t.Fatalf("SerializeNodeType failed: %v", err)
	}

	if !strings.Contains(result, "CREATE CONSTRAINT person_email_unique IF NOT EXISTS") {
		t.Errorf("expected unique constraint, got: %s", result)
	}
	if !strings.Contains(result, "IS UNIQUE") {
		t.Errorf("expected IS UNIQUE clause, got: %s", result)
	}
}

func TestCypherSerializer_SerializeNodeType_RequiredConstraint(t *testing.T) {
	s := NewCypherSerializer()
	node := &schema.NodeType{
		Label: "Person",
		Properties: []schema.Property{
			{Name: "name", Type: schema.STRING, Required: true},
		},
	}

	result, err := s.SerializeNodeType(node)
	if err != nil {
		t.Fatalf("SerializeNodeType failed: %v", err)
	}

	if !strings.Contains(result, "CREATE CONSTRAINT person_name_not_null IF NOT EXISTS") {
		t.Errorf("expected not null constraint, got: %s", result)
	}
	if !strings.Contains(result, "IS NOT NULL") {
		t.Errorf("expected IS NOT NULL clause, got: %s", result)
	}
}

func TestCypherSerializer_SerializeNodeType_ExplicitConstraints(t *testing.T) {
	s := NewCypherSerializer()
	node := &schema.NodeType{
		Label: "Person",
		Constraints: []schema.Constraint{
			{Name: "person_key", Type: schema.NODE_KEY, Properties: []string{"id", "email"}},
		},
	}

	result, err := s.SerializeNodeType(node)
	if err != nil {
		t.Fatalf("SerializeNodeType failed: %v", err)
	}

	if !strings.Contains(result, "CREATE CONSTRAINT person_key IF NOT EXISTS") {
		t.Errorf("expected constraint name, got: %s", result)
	}
	if !strings.Contains(result, "IS NODE KEY") {
		t.Errorf("expected IS NODE KEY clause, got: %s", result)
	}
	if !strings.Contains(result, "n.id") && !strings.Contains(result, "n.email") {
		t.Errorf("expected property references, got: %s", result)
	}
}

func TestCypherSerializer_SerializeNodeType_BTreeIndex(t *testing.T) {
	s := NewCypherSerializer()
	node := &schema.NodeType{
		Label: "Person",
		Indexes: []schema.Index{
			{Name: "person_name_idx", Type: schema.BTREE, Properties: []string{"name"}},
		},
	}

	result, err := s.SerializeNodeType(node)
	if err != nil {
		t.Fatalf("SerializeNodeType failed: %v", err)
	}

	if !strings.Contains(result, "CREATE INDEX person_name_idx IF NOT EXISTS") {
		t.Errorf("expected index creation, got: %s", result)
	}
	if !strings.Contains(result, "ON (n.name)") {
		t.Errorf("expected ON clause, got: %s", result)
	}
}

func TestCypherSerializer_SerializeNodeType_TextIndex(t *testing.T) {
	s := NewCypherSerializer()
	node := &schema.NodeType{
		Label: "Article",
		Indexes: []schema.Index{
			{Name: "article_title_text", Type: schema.TEXT, Properties: []string{"title"}},
		},
	}

	result, err := s.SerializeNodeType(node)
	if err != nil {
		t.Fatalf("SerializeNodeType failed: %v", err)
	}

	if !strings.Contains(result, "CREATE TEXT INDEX article_title_text IF NOT EXISTS") {
		t.Errorf("expected TEXT index, got: %s", result)
	}
}

func TestCypherSerializer_SerializeNodeType_FulltextIndex(t *testing.T) {
	s := NewCypherSerializer()
	node := &schema.NodeType{
		Label: "Article",
		Indexes: []schema.Index{
			{Name: "article_fulltext", Type: schema.FULLTEXT, Properties: []string{"title", "content"}},
		},
	}

	result, err := s.SerializeNodeType(node)
	if err != nil {
		t.Fatalf("SerializeNodeType failed: %v", err)
	}

	if !strings.Contains(result, "CREATE FULLTEXT INDEX article_fulltext IF NOT EXISTS") {
		t.Errorf("expected FULLTEXT index, got: %s", result)
	}
	if !strings.Contains(result, "ON EACH [n.title, n.content]") {
		t.Errorf("expected ON EACH clause with properties, got: %s", result)
	}
}

func TestCypherSerializer_SerializeNodeType_PointIndex(t *testing.T) {
	s := NewCypherSerializer()
	node := &schema.NodeType{
		Label: "Location",
		Indexes: []schema.Index{
			{Name: "location_point_idx", Type: schema.POINT_INDEX, Properties: []string{"coordinates"}},
		},
	}

	result, err := s.SerializeNodeType(node)
	if err != nil {
		t.Fatalf("SerializeNodeType failed: %v", err)
	}

	if !strings.Contains(result, "CREATE POINT INDEX location_point_idx IF NOT EXISTS") {
		t.Errorf("expected POINT index, got: %s", result)
	}
}

func TestCypherSerializer_SerializeNodeType_VectorIndex(t *testing.T) {
	s := NewCypherSerializer()
	node := &schema.NodeType{
		Label: "Document",
		Indexes: []schema.Index{
			{
				Name:       "document_embedding_idx",
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

	if !strings.Contains(result, "CREATE VECTOR INDEX document_embedding_idx IF NOT EXISTS") {
		t.Errorf("expected VECTOR index, got: %s", result)
	}
	if !strings.Contains(result, "384") {
		t.Errorf("expected dimensions 384, got: %s", result)
	}
	if !strings.Contains(result, "cosine") {
		t.Errorf("expected cosine similarity, got: %s", result)
	}
}

func TestCypherSerializer_SerializeRelationshipType_RequiredProperty(t *testing.T) {
	s := NewCypherSerializer()
	rel := &schema.RelationshipType{
		Label:  "WORKS_FOR",
		Source: "Person",
		Target: "Company",
		Properties: []schema.Property{
			{Name: "since", Type: schema.DATE, Required: true},
		},
	}

	result, err := s.SerializeRelationshipType(rel)
	if err != nil {
		t.Fatalf("SerializeRelationshipType failed: %v", err)
	}

	if !strings.Contains(result, "FOR ()-[r:WORKS_FOR]-()") {
		t.Errorf("expected relationship pattern, got: %s", result)
	}
	if !strings.Contains(result, "IS NOT NULL") {
		t.Errorf("expected IS NOT NULL, got: %s", result)
	}
}

func TestCypherSerializer_SerializeRelationshipType_RelKey(t *testing.T) {
	s := NewCypherSerializer()
	rel := &schema.RelationshipType{
		Label:  "WORKS_FOR",
		Source: "Person",
		Target: "Company",
		Constraints: []schema.Constraint{
			{Name: "works_for_key", Type: schema.REL_KEY, Properties: []string{"since", "role"}},
		},
	}

	result, err := s.SerializeRelationshipType(rel)
	if err != nil {
		t.Fatalf("SerializeRelationshipType failed: %v", err)
	}

	if !strings.Contains(result, "IS RELATIONSHIP KEY") {
		t.Errorf("expected RELATIONSHIP KEY, got: %s", result)
	}
}

func TestCypherSerializer_SerializeRelationshipType_Empty(t *testing.T) {
	s := NewCypherSerializer()
	rel := &schema.RelationshipType{
		Label:  "KNOWS",
		Source: "Person",
		Target: "Person",
	}

	result, err := s.SerializeRelationshipType(rel)
	if err != nil {
		t.Fatalf("SerializeRelationshipType failed: %v", err)
	}

	if result != "" {
		t.Errorf("expected empty result for relationship without constraints, got: %s", result)
	}
}

func TestCypherSerializer_SerializeAll(t *testing.T) {
	s := NewCypherSerializer()

	nodeTypes := []*schema.NodeType{
		{
			Label: "Person",
			Properties: []schema.Property{
				{Name: "id", Type: schema.STRING, Unique: true},
			},
		},
		{
			Label: "Company",
			Properties: []schema.Property{
				{Name: "name", Type: schema.STRING, Required: true},
			},
		},
	}

	relTypes := []*schema.RelationshipType{
		{
			Label:  "WORKS_FOR",
			Source: "Person",
			Target: "Company",
			Properties: []schema.Property{
				{Name: "since", Type: schema.DATE, Required: true},
			},
		},
	}

	result, err := s.SerializeAll(nodeTypes, relTypes)
	if err != nil {
		t.Fatalf("SerializeAll failed: %v", err)
	}

	// Check for comments
	if !strings.Contains(result, "// Person") {
		t.Errorf("expected Person comment, got: %s", result)
	}
	if !strings.Contains(result, "// Company") {
		t.Errorf("expected Company comment, got: %s", result)
	}
	if !strings.Contains(result, "// WORKS_FOR") {
		t.Errorf("expected WORKS_FOR comment, got: %s", result)
	}

	// Check for constraints
	if !strings.Contains(result, "person_id_unique") {
		t.Errorf("expected person_id_unique constraint, got: %s", result)
	}
	if !strings.Contains(result, "company_name_not_null") {
		t.Errorf("expected company_name_not_null constraint, got: %s", result)
	}
}

func TestCypherSerializer_UnsupportedConstraintType(t *testing.T) {
	s := NewCypherSerializer()
	node := &schema.NodeType{
		Label: "Person",
		Constraints: []schema.Constraint{
			{Name: "invalid", Type: schema.ConstraintType("INVALID"), Properties: []string{"id"}},
		},
	}

	_, err := s.SerializeNodeType(node)
	if err == nil {
		t.Error("expected error for unsupported constraint type")
	}
}

func TestCypherSerializer_UnsupportedIndexType(t *testing.T) {
	s := NewCypherSerializer()
	node := &schema.NodeType{
		Label: "Person",
		Indexes: []schema.Index{
			{Name: "invalid", Type: schema.IndexType("INVALID"), Properties: []string{"id"}},
		},
	}

	_, err := s.SerializeNodeType(node)
	if err == nil {
		t.Error("expected error for unsupported index type")
	}
}

func TestCypherSerializer_UnsupportedRelConstraintType(t *testing.T) {
	s := NewCypherSerializer()
	rel := &schema.RelationshipType{
		Label:  "WORKS_FOR",
		Source: "Person",
		Target: "Company",
		Constraints: []schema.Constraint{
			{Name: "invalid", Type: schema.UNIQUE, Properties: []string{"id"}}, // UNIQUE not supported for rels
		},
	}

	_, err := s.SerializeRelationshipType(rel)
	if err == nil {
		t.Error("expected error for unsupported relationship constraint type")
	}
}
