// Package schema provides Neo4j schema definitions for knowledge graph scenario.
package schema

import "github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"

// Document represents a research paper or document in the knowledge graph.
// Includes vector embeddings for semantic search and fulltext indexing.
var Document = &schema.NodeType{
	Label:       "Document",
	Description: "A research paper or academic document",
	Properties: []schema.Property{
		{Name: "id", Type: schema.STRING, Required: true, Unique: true},
		{Name: "title", Type: schema.STRING, Required: true},
		{Name: "content", Type: schema.STRING, Required: true},
		{Name: "embedding", Type: schema.LIST_FLOAT},
		{Name: "year", Type: schema.INTEGER},
		{Name: "doi", Type: schema.STRING},
		{Name: "abstract", Type: schema.STRING},
	},
	Constraints: []schema.Constraint{
		{Name: "document_id_unique", Type: schema.UNIQUE, Properties: []string{"id"}},
	},
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
		{Name: "document_content_fulltext", Type: schema.FULLTEXT, Properties: []string{"content"}},
	},
}

// Person represents a researcher or author in the knowledge graph.
var Person = &schema.NodeType{
	Label:       "Person",
	Description: "A researcher or author",
	Properties: []schema.Property{
		{Name: "name", Type: schema.STRING, Required: true, Unique: true},
		{Name: "affiliation", Type: schema.STRING},
		{Name: "orcid", Type: schema.STRING},
	},
	Constraints: []schema.Constraint{
		{Name: "person_name_unique", Type: schema.UNIQUE, Properties: []string{"name"}},
	},
}

// Concept represents a scientific concept, theory, or methodology.
var Concept = &schema.NodeType{
	Label:       "Concept",
	Description: "A scientific concept, theory, or methodology",
	Properties: []schema.Property{
		{Name: "name", Type: schema.STRING, Required: true, Unique: true},
		{Name: "definition", Type: schema.STRING},
		{Name: "field", Type: schema.STRING},
	},
	Constraints: []schema.Constraint{
		{Name: "concept_name_unique", Type: schema.UNIQUE, Properties: []string{"name"}},
	},
}

// Institution represents a research institution or university.
var Institution = &schema.NodeType{
	Label:       "Institution",
	Description: "A research institution or university",
	Properties: []schema.Property{
		{Name: "name", Type: schema.STRING, Required: true, Unique: true},
		{Name: "country", Type: schema.STRING},
	},
	Constraints: []schema.Constraint{
		{Name: "institution_name_unique", Type: schema.UNIQUE, Properties: []string{"name"}},
	},
}

// Authored represents authorship of a paper.
var Authored = &schema.RelationshipType{
	Label:       "AUTHORED",
	Description: "Authorship relationship between person and document",
	Source:      "Person",
	Target:      "Document",
	Cardinality: schema.MANY_TO_MANY,
	Properties: []schema.Property{
		{Name: "order", Type: schema.INTEGER},
		{Name: "corresponding", Type: schema.BOOLEAN},
	},
}

// Cites represents citation relationships between papers.
var Cites = &schema.RelationshipType{
	Label:       "CITES",
	Description: "Citation relationship between documents",
	Source:      "Document",
	Target:      "Document",
	Cardinality: schema.MANY_TO_MANY,
}

// Studies represents research focus on a concept.
var Studies = &schema.RelationshipType{
	Label:       "STUDIES",
	Description: "Research focus on a concept",
	Source:      "Document",
	Target:      "Concept",
	Cardinality: schema.MANY_TO_MANY,
}

// AffiliatedWith represents researcher affiliation with an institution.
var AffiliatedWith = &schema.RelationshipType{
	Label:       "AFFILIATED_WITH",
	Description: "Researcher affiliation with institution",
	Source:      "Person",
	Target:      "Institution",
	Cardinality: schema.MANY_TO_ONE,
}
