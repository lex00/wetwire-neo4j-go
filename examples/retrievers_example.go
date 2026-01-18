package examples

import (
	"github.com/lex00/wetwire-neo4j-go/internal/retrievers"
)

// VectorRetrieverExample demonstrates basic vector similarity search.
// Reference: https://github.com/neo4j/neo4j-graphrag-python
var VectorRetrieverExample = &retrievers.VectorRetriever{
	IndexName:        "document_embedding_vector_idx",
	ReturnProperties: []string{"id", "content", "title"},
	TopK:             10,
	EmbedderConfig: &retrievers.EmbedderConfig{
		Provider:   "openai",
		Model:      "text-embedding-3-small",
		Dimensions: 384,
	},
}

// VectorCypherRetrieverExample demonstrates vector search with graph traversal.
var VectorCypherRetrieverExample = &retrievers.VectorCypherRetriever{
	IndexName: "document_embedding_vector_idx",
	TopK:      5,
	RetrievalQuery: `
		MATCH (doc:Document)
		WHERE doc = node
		OPTIONAL MATCH (doc)-[:MENTIONS]->(entity:Entity)
		OPTIONAL MATCH (doc)-[:AUTHORED_BY]->(author:Person)
		RETURN doc.content AS content,
		       doc.title AS title,
		       collect(DISTINCT entity.name) AS entities,
		       author.name AS author
	`,
	EmbedderConfig: &retrievers.EmbedderConfig{
		Provider:   "openai",
		Model:      "text-embedding-3-small",
		Dimensions: 384,
	},
}

// HybridRetrieverExample demonstrates combining vector and fulltext search.
var HybridRetrieverExample = &retrievers.HybridRetriever{
	VectorIndexName:   "document_embedding_vector_idx",
	FulltextIndexName: "document_content_fulltext_idx",
	TopK:              10,
	VectorWeight:      0.7,
	FulltextWeight:    0.3,
	ReturnProperties:  []string{"id", "content", "title"},
	EmbedderConfig: &retrievers.EmbedderConfig{
		Provider:   "cohere",
		Model:      "embed-english-v3.0",
		Dimensions: 1024,
	},
}

// HybridCypherRetrieverExample demonstrates hybrid search with graph context.
var HybridCypherRetrieverExample = &retrievers.HybridCypherRetriever{
	VectorIndexName:   "chunk_embedding_idx",
	FulltextIndexName: "chunk_text_idx",
	TopK:              5,
	RetrievalQuery: `
		MATCH (chunk:Chunk)
		WHERE chunk = node
		MATCH (chunk)-[:PART_OF]->(doc:Document)
		OPTIONAL MATCH (doc)-[:HAS_TOPIC]->(topic:Topic)
		RETURN chunk.text AS text,
		       doc.title AS documentTitle,
		       collect(DISTINCT topic.name) AS topics
	`,
	EmbedderConfig: &retrievers.EmbedderConfig{
		Provider:   "openai",
		Model:      "text-embedding-3-large",
		Dimensions: 1536,
	},
}

// Text2CypherRetrieverExample demonstrates LLM-generated Cypher queries.
var Text2CypherRetrieverExample = &retrievers.Text2CypherRetriever{
	LLMProvider: "openai",
	LLMModel:    "gpt-4",
	SchemaDescription: `
		Node labels: Person, Company, Location, Document
		Relationship types:
		  - (Person)-[:WORKS_FOR]->(Company)
		  - (Person)-[:KNOWS]->(Person)
		  - (Company)-[:LOCATED_IN]->(Location)
		  - (Document)-[:MENTIONS]->(Person|Company|Location)
	`,
	Examples: []retrievers.CypherExample{
		{
			Question: "Who works at Acme Corp?",
			Cypher:   "MATCH (p:Person)-[:WORKS_FOR]->(c:Company {name: 'Acme Corp'}) RETURN p.name",
		},
		{
			Question: "Find documents mentioning John Smith",
			Cypher:   "MATCH (d:Document)-[:MENTIONS]->(p:Person {name: 'John Smith'}) RETURN d.title, d.content",
		},
	},
}

// WeaviateRetrieverExample demonstrates integration with Weaviate.
var WeaviateRetrieverExample = &retrievers.WeaviateRetriever{
	WeaviateURL: "http://localhost:8080",
	Collection:  "Document",
	TopK:        5,
	RetrievalQuery: `
		MATCH (doc:Document {id: $id})
		OPTIONAL MATCH (doc)-[:RELATED_TO]->(related:Document)
		RETURN doc, collect(related) AS relatedDocs
	`,
	IDProperty: "id",
}

// PineconeRetrieverExample demonstrates integration with Pinecone.
var PineconeRetrieverExample = &retrievers.PineconeRetriever{
	PineconeAPIKey: "${PINECONE_API_KEY}",
	IndexName:      "documents",
	Namespace:      "production",
	TopK:           10,
	RetrievalQuery: `
		MATCH (doc:Document {id: $id})
		MATCH (doc)-[:HAS_CHUNK]->(chunk:Chunk)
		RETURN doc.title, collect(chunk.text) AS chunks
	`,
	IDProperty: "id",
}

// QdrantRetrieverExample demonstrates integration with Qdrant.
var QdrantRetrieverExample = &retrievers.QdrantRetriever{
	QdrantURL:      "http://localhost:6333",
	CollectionName: "knowledge-base",
	TopK:           5,
	RetrievalQuery: `
		MATCH (item {id: $id})
		OPTIONAL MATCH path = (item)-[*1..2]-(related)
		RETURN item, nodes(path) AS context
	`,
	IDProperty: "item_id",
}

// AllRetrieverExamples returns all example retriever configurations.
func AllRetrieverExamples() []retrievers.Retriever {
	return []retrievers.Retriever{
		VectorRetrieverExample,
		VectorCypherRetrieverExample,
		HybridRetrieverExample,
		HybridCypherRetrieverExample,
		Text2CypherRetrieverExample,
		WeaviateRetrieverExample,
		PineconeRetrieverExample,
		QdrantRetrieverExample,
	}
}
