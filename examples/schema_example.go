// Package examples provides reference examples from Neo4j documentation.
//
// These examples demonstrate idiomatic usage of wetwire-neo4j-go for defining
// Neo4j schemas, GDS algorithms, and GraphRAG configurations.
package examples

import (
	"github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"
)

// PersonNode demonstrates a basic node type definition.
// Reference: https://neo4j.com/docs/cypher-manual/current/constraints/
var PersonNode = &schema.NodeType{
	Label:       "Person",
	Description: "A person entity with name and age properties",
	Properties: []schema.Property{
		{Name: "id", Type: schema.STRING, Required: true, Unique: true},
		{Name: "name", Type: schema.STRING, Required: true},
		{Name: "age", Type: schema.INTEGER},
		{Name: "email", Type: schema.STRING, Unique: true},
	},
	Constraints: []schema.Constraint{
		{Name: "person_id_unique", Type: schema.UNIQUE, Properties: []string{"id"}},
	},
	Indexes: []schema.Index{
		{Name: "person_name_idx", Type: schema.BTREE, Properties: []string{"name"}},
		{Name: "person_email_text_idx", Type: schema.TEXT, Properties: []string{"email"}},
	},
}

// CompanyNode demonstrates a node with multiple indexes.
var CompanyNode = &schema.NodeType{
	Label:       "Company",
	Description: "A company entity",
	Properties: []schema.Property{
		{Name: "name", Type: schema.STRING, Required: true, Unique: true},
		{Name: "industry", Type: schema.STRING},
		{Name: "founded", Type: schema.DATE},
		{Name: "description", Type: schema.STRING},
	},
	Constraints: []schema.Constraint{
		{Name: "company_name_unique", Type: schema.UNIQUE, Properties: []string{"name"}},
	},
	Indexes: []schema.Index{
		{Name: "company_industry_idx", Type: schema.BTREE, Properties: []string{"industry"}},
		{Name: "company_fulltext_idx", Type: schema.FULLTEXT, Properties: []string{"name", "description"}},
	},
}

// DocumentNode demonstrates a node with vector embedding for semantic search.
// Reference: https://neo4j.com/docs/cypher-manual/current/indexes/semantic-indexes/vector-indexes/
var DocumentNode = &schema.NodeType{
	Label:       "Document",
	Description: "A document with text content and vector embedding",
	Properties: []schema.Property{
		{Name: "id", Type: schema.STRING, Required: true, Unique: true},
		{Name: "content", Type: schema.STRING, Required: true},
		{Name: "embedding", Type: schema.LIST_FLOAT},
	},
	Indexes: []schema.Index{
		{
			Name:       "document_embedding_vector_idx",
			Type:       schema.VECTOR,
			Properties: []string{"embedding"},
			Options: map[string]any{
				"dimensions":          384,
				"similarity_function": "cosine",
			},
		},
		{Name: "document_content_fulltext_idx", Type: schema.FULLTEXT, Properties: []string{"content"}},
	},
}

// LocationNode demonstrates a node with spatial point index.
// Reference: https://neo4j.com/docs/cypher-manual/current/indexes/semantic-indexes/point-indexes/
var LocationNode = &schema.NodeType{
	Label:       "Location",
	Description: "A geographic location",
	Properties: []schema.Property{
		{Name: "name", Type: schema.STRING, Required: true},
		{Name: "coordinates", Type: schema.POINT, Required: true},
		{Name: "address", Type: schema.STRING},
	},
	Indexes: []schema.Index{
		{Name: "location_coords_point_idx", Type: schema.POINT_INDEX, Properties: []string{"coordinates"}},
	},
}

// WorksForRelationship demonstrates a relationship with properties.
// Reference: https://neo4j.com/docs/cypher-manual/current/constraints/
var WorksForRelationship = &schema.RelationshipType{
	Label:       "WORKS_FOR",
	Description: "Employment relationship between Person and Company",
	Source:      "Person",
	Target:      "Company",
	Cardinality: schema.MANY_TO_ONE,
	Properties: []schema.Property{
		{Name: "since", Type: schema.DATE, Required: true},
		{Name: "role", Type: schema.STRING},
		{Name: "salary", Type: schema.FLOAT},
	},
	Constraints: []schema.Constraint{
		{Name: "works_for_since_exists", Type: schema.EXISTS, Properties: []string{"since"}},
	},
}

// KnowsRelationship demonstrates a simple relationship.
var KnowsRelationship = &schema.RelationshipType{
	Label:       "KNOWS",
	Description: "Social connection between people",
	Source:      "Person",
	Target:      "Person",
	Cardinality: schema.MANY_TO_MANY,
	Properties: []schema.Property{
		{Name: "since", Type: schema.DATE},
		{Name: "weight", Type: schema.FLOAT},
	},
}

// LocatedInRelationship demonstrates location association.
var LocatedInRelationship = &schema.RelationshipType{
	Label:       "LOCATED_IN",
	Description: "Geographic location of an entity",
	Source:      "Company",
	Target:      "Location",
	Cardinality: schema.MANY_TO_ONE,
}

// AllNodeTypes returns all example node types.
func AllNodeTypes() []*schema.NodeType {
	return []*schema.NodeType{
		PersonNode,
		CompanyNode,
		DocumentNode,
		LocationNode,
	}
}

// AllRelationshipTypes returns all example relationship types.
func AllRelationshipTypes() []*schema.RelationshipType {
	return []*schema.RelationshipType{
		WorksForRelationship,
		KnowsRelationship,
		LocatedInRelationship,
	}
}
