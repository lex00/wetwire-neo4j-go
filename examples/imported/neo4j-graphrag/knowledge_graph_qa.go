package neo4j_graphrag

import (
	"github.com/lex00/wetwire-neo4j-go/internal/kg"
	"github.com/lex00/wetwire-neo4j-go/internal/retrievers"
	"github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"
)

// KnowledgeGraphQASchema demonstrates a GraphRAG knowledge graph for Q&A.
// This schema supports question-answering over documents using:
// - Document chunking and entity extraction
// - Entity linking and relationship extraction
// - Vector embeddings for semantic search
// - Hybrid retrieval (vector + graph traversal)
//
// References:
// - https://neo4j.com/labs/genai-ecosystem/llm-graph-builder/
// - https://neo4j.com/labs/genai-ecosystem/graphrag-python/
// - https://graphacademy.neo4j.com/knowledge-graph-rag/
var KnowledgeGraphQASchema = &schema.Schema{
	Name:        "knowledge-graph-qa",
	Description: "Knowledge graph for document Q&A with GraphRAG",
	AgentContext: "This schema stores documents, chunks, entities, and their relationships for RAG. " +
		"Use vector similarity for initial retrieval, then traverse graph for context enrichment.",
	Nodes: []*schema.NodeType{
		DocumentNode,
		ChunkNode,
		EntityNode,
		ConceptNode,
		TopicNode,
	},
	Relationships: []*schema.RelationshipType{
		HasChunkRel,
		MentionsEntityRel,
		EntityRelationRel,
		AboutConceptRel,
		SubconceptOfRel,
		InTopicRel,
		RelatedToDocRel,
	},
}

// DocumentNode represents a source document.
var DocumentNode = &schema.NodeType{
	Label:       "Document",
	Description: "A source document (article, paper, webpage, etc.)",
	Properties: []schema.Property{
		{Name: "documentId", Type: schema.STRING, Required: true, Unique: true},
		{Name: "title", Type: schema.STRING, Required: true},
		{Name: "source", Type: schema.STRING, Description: "URL, file path, or source identifier"},
		{Name: "documentType", Type: schema.STRING, Description: "pdf, html, text, markdown"},
		{Name: "author", Type: schema.STRING},
		{Name: "publishedDate", Type: schema.DATE},
		{Name: "summary", Type: schema.STRING, Description: "Document summary"},
		{Name: "embedding", Type: schema.LIST_FLOAT, Description: "Document-level embedding"},
		{Name: "language", Type: schema.STRING},
		{Name: "metadata", Type: schema.STRING, Description: "JSON metadata"},
	},
	Constraints: []schema.Constraint{
		{Type: schema.UNIQUE, Properties: []string{"documentId"}},
	},
	Indexes: []schema.Index{
		{Type: schema.TEXT, Properties: []string{"title", "summary"}},
		{Type: schema.BTREE, Properties: []string{"publishedDate"}},
		{Type: schema.VECTOR, Properties: []string{"embedding"}, Options: map[string]any{
			"dimensions": 1536,
			"similarity": "cosine",
		}},
	},
	AgentHint: "Query by documentId for unique identification. Use vector search on embedding for similarity.",
}

// ChunkNode represents a text chunk from a document.
var ChunkNode = &schema.NodeType{
	Label:       "Chunk",
	Description: "A text chunk extracted from a document for RAG",
	Properties: []schema.Property{
		{Name: "chunkId", Type: schema.STRING, Required: true, Unique: true},
		{Name: "text", Type: schema.STRING, Required: true},
		{Name: "position", Type: schema.INTEGER, Required: true, Description: "Chunk position in document"},
		{Name: "embedding", Type: schema.LIST_FLOAT, Required: true, Description: "Vector embedding for retrieval"},
		{Name: "tokenCount", Type: schema.INTEGER},
		{Name: "pageNumber", Type: schema.INTEGER, Description: "Page number (for PDFs)"},
		{Name: "section", Type: schema.STRING, Description: "Section or heading"},
	},
	Constraints: []schema.Constraint{
		{Type: schema.UNIQUE, Properties: []string{"chunkId"}},
	},
	Indexes: []schema.Index{
		{Type: schema.TEXT, Properties: []string{"text"}},
		{Type: schema.VECTOR, Properties: []string{"embedding"}, Options: map[string]any{
			"dimensions": 1536,
			"similarity": "cosine",
		}},
	},
	AgentHint: "Primary retrieval target. Use vector search on embedding, then traverse to entities.",
}

// EntityNode represents a named entity extracted from text.
var EntityNode = &schema.NodeType{
	Label:       "Entity",
	Description: "A named entity (person, organization, location, etc.)",
	Properties: []schema.Property{
		{Name: "entityId", Type: schema.STRING, Required: true, Unique: true},
		{Name: "name", Type: schema.STRING, Required: true},
		{Name: "type", Type: schema.STRING, Required: true, Description: "Person, Organization, Location, etc."},
		{Name: "description", Type: schema.STRING, Description: "Entity description"},
		{Name: "aliases", Type: schema.LIST_STRING, Description: "Alternative names"},
		{Name: "embedding", Type: schema.LIST_FLOAT, Description: "Entity embedding"},
		{Name: "properties", Type: schema.STRING, Description: "JSON additional properties"},
	},
	Constraints: []schema.Constraint{
		{Type: schema.UNIQUE, Properties: []string{"entityId"}},
	},
	Indexes: []schema.Index{
		{Type: schema.BTREE, Properties: []string{"type"}},
		{Type: schema.TEXT, Properties: []string{"name", "description"}},
		{Type: schema.VECTOR, Properties: []string{"embedding"}, Options: map[string]any{
			"dimensions": 1536,
			"similarity": "cosine",
		}},
	},
	AgentHint: "Query by type for entity filtering. Use graph traversal from entities to chunks.",
}

// ConceptNode represents an abstract concept or topic.
var ConceptNode = &schema.NodeType{
	Label:       "Concept",
	Description: "An abstract concept, theory, or idea",
	Properties: []schema.Property{
		{Name: "conceptId", Type: schema.STRING, Required: true, Unique: true},
		{Name: "name", Type: schema.STRING, Required: true},
		{Name: "definition", Type: schema.STRING},
		{Name: "domain", Type: schema.STRING, Description: "Domain or field (science, business, etc.)"},
		{Name: "embedding", Type: schema.LIST_FLOAT},
	},
	Constraints: []schema.Constraint{
		{Type: schema.UNIQUE, Properties: []string{"conceptId"}},
	},
	Indexes: []schema.Index{
		{Type: schema.TEXT, Properties: []string{"name", "definition"}},
		{Type: schema.BTREE, Properties: []string{"domain"}},
	},
	AgentHint: "Use for conceptual understanding and topic-based retrieval.",
}

// TopicNode represents a document topic or category.
var TopicNode = &schema.NodeType{
	Label:       "Topic",
	Description: "A topic or category for document classification",
	Properties: []schema.Property{
		{Name: "topicId", Type: schema.STRING, Required: true, Unique: true},
		{Name: "name", Type: schema.STRING, Required: true},
		{Name: "description", Type: schema.STRING},
		{Name: "parentTopic", Type: schema.STRING, Description: "Parent topic ID for hierarchy"},
	},
	Constraints: []schema.Constraint{
		{Type: schema.UNIQUE, Properties: []string{"topicId"}},
	},
	AgentHint: "Use for topic-based filtering and document organization.",
}

// Relationships

var HasChunkRel = &schema.RelationshipType{
	Label:       "HAS_CHUNK",
	Source:      "Document",
	Target:      "Chunk",
	Cardinality: schema.ONE_TO_MANY,
	Description: "Document contains a text chunk",
	Properties: []schema.Property{
		{Name: "position", Type: schema.INTEGER, Required: true},
	},
	AgentHint: "Use to find all chunks from a document for context.",
}

var MentionsEntityRel = &schema.RelationshipType{
	Label:       "MENTIONS",
	Source:      "Chunk",
	Target:      "Entity",
	Cardinality: schema.MANY_TO_MANY,
	Description: "Chunk mentions an entity",
	Properties: []schema.Property{
		{Name: "count", Type: schema.INTEGER, Description: "Number of mentions"},
		{Name: "salience", Type: schema.FLOAT, Description: "Entity salience score 0-1"},
		{Name: "sentiment", Type: schema.STRING, Description: "positive, negative, neutral"},
	},
	AgentHint: "Use salience to prioritize important entity mentions.",
}

var EntityRelationRel = &schema.RelationshipType{
	Label:       "RELATES_TO",
	Source:      "Entity",
	Target:      "Entity",
	Cardinality: schema.MANY_TO_MANY,
	Description: "Relationship between entities extracted from text",
	Properties: []schema.Property{
		{Name: "relationType", Type: schema.STRING, Required: true, Description: "Type of relationship"},
		{Name: "description", Type: schema.STRING, Description: "Relationship description"},
		{Name: "confidence", Type: schema.FLOAT, Description: "Extraction confidence 0-1"},
		{Name: "sourceChunkId", Type: schema.STRING, Description: "Chunk where relationship was found"},
	},
	AgentHint: "Use relationType for specific relationship queries. Check confidence for reliability.",
}

var AboutConceptRel = &schema.RelationshipType{
	Label:       "ABOUT",
	Source:      "Document",
	Target:      "Concept",
	Cardinality: schema.MANY_TO_MANY,
	Description: "Document is about a concept",
	Properties: []schema.Property{
		{Name: "relevance", Type: schema.FLOAT, Description: "Concept relevance score 0-1"},
	},
}

var SubconceptOfRel = &schema.RelationshipType{
	Label:       "SUBCONCEPT_OF",
	Source:      "Concept",
	Target:      "Concept",
	Cardinality: schema.MANY_TO_ONE,
	Description: "Concept is a subconcept of another (hierarchy)",
}

var InTopicRel = &schema.RelationshipType{
	Label:       "IN_TOPIC",
	Source:      "Document",
	Target:      "Topic",
	Cardinality: schema.MANY_TO_MANY,
	Description: "Document belongs to a topic",
	Properties: []schema.Property{
		{Name: "confidence", Type: schema.FLOAT, Description: "Classification confidence 0-1"},
	},
}

var RelatedToDocRel = &schema.RelationshipType{
	Label:       "RELATED_TO",
	Source:      "Document",
	Target:      "Document",
	Cardinality: schema.MANY_TO_MANY,
	Description: "Documents are related (computed by similarity)",
	Properties: []schema.Property{
		{Name: "similarity", Type: schema.FLOAT, Required: true, Description: "Similarity score 0-1"},
		{Name: "method", Type: schema.STRING, Description: "Similarity computation method"},
	},
	AgentHint: "Use for finding related documents for context expansion.",
}

// Knowledge Graph Pipeline for document processing
// Extracts entities and relationships from documents using LLMs.
// Reference: https://neo4j.com/labs/genai-ecosystem/llm-graph-builder/
var DocumentKGPipeline = &kg.SimpleKGPipeline{
	BasePipeline: kg.BasePipeline{
		Name: "document-knowledge-graph",
		LLMConfig: &kg.LLMConfig{
			Provider:    "openai",
			Model:       "gpt-4-turbo",
			Temperature: 0.0,
			MaxTokens:   4000,
		},
		EmbedderConfig: &kg.EmbedderConfig{
			Provider:   "openai",
			Model:      "text-embedding-3-large",
			Dimensions: 1536,
		},
	},
	EntityTypes: []kg.EntityType{
		{
			Name:        "Person",
			Description: "A person mentioned in the document",
			Properties: []kg.EntityProperty{
				{Name: "name", Type: "STRING", Required: true},
				{Name: "role", Type: "STRING"},
				{Name: "affiliation", Type: "STRING"},
			},
		},
		{
			Name:        "Organization",
			Description: "An organization, company, or institution",
			Properties: []kg.EntityProperty{
				{Name: "name", Type: "STRING", Required: true},
				{Name: "type", Type: "STRING"},
				{Name: "industry", Type: "STRING"},
			},
		},
		{
			Name:        "Location",
			Description: "A geographic location",
			Properties: []kg.EntityProperty{
				{Name: "name", Type: "STRING", Required: true},
				{Name: "type", Type: "STRING"},
				{Name: "country", Type: "STRING"},
			},
		},
		{
			Name:        "Technology",
			Description: "A technology, tool, or system",
			Properties: []kg.EntityProperty{
				{Name: "name", Type: "STRING", Required: true},
				{Name: "type", Type: "STRING"},
				{Name: "version", Type: "STRING"},
			},
		},
		{
			Name:        "Concept",
			Description: "An abstract concept, theory, or methodology",
			Properties: []kg.EntityProperty{
				{Name: "name", Type: "STRING", Required: true},
				{Name: "definition", Type: "STRING"},
				{Name: "domain", Type: "STRING"},
			},
		},
		{
			Name:        "Event",
			Description: "An event or occurrence",
			Properties: []kg.EntityProperty{
				{Name: "name", Type: "STRING", Required: true},
				{Name: "date", Type: "DATE"},
				{Name: "description", Type: "STRING"},
			},
		},
	},
	RelationTypes: []kg.RelationType{
		{
			Name:        "WORKS_FOR",
			Description: "Person works for an organization",
			SourceTypes: []string{"Person"},
			TargetTypes: []string{"Organization"},
		},
		{
			Name:        "LOCATED_IN",
			Description: "Entity is located in a location",
			SourceTypes: []string{"Person", "Organization", "Event"},
			TargetTypes: []string{"Location"},
		},
		{
			Name:        "USES",
			Description: "Entity uses a technology",
			SourceTypes: []string{"Person", "Organization"},
			TargetTypes: []string{"Technology"},
		},
		{
			Name:        "IMPLEMENTS",
			Description: "Technology implements a concept",
			SourceTypes: []string{"Technology"},
			TargetTypes: []string{"Concept"},
		},
		{
			Name:        "RELATED_TO",
			Description: "General relationship between entities",
			SourceTypes: []string{"Person", "Organization", "Technology", "Concept"},
			TargetTypes: []string{"Person", "Organization", "Technology", "Concept"},
		},
		{
			Name:        "PART_OF",
			Description: "Entity is part of another entity",
			SourceTypes: []string{"Organization", "Location", "Concept"},
			TargetTypes: []string{"Organization", "Location", "Concept"},
		},
		{
			Name:        "OCCURRED_AT",
			Description: "Event occurred at a location",
			SourceTypes: []string{"Event"},
			TargetTypes: []string{"Location"},
		},
		{
			Name:        "PARTICIPATED_IN",
			Description: "Person or organization participated in an event",
			SourceTypes: []string{"Person", "Organization"},
			TargetTypes: []string{"Event"},
		},
	},
	TextSplitter: &kg.LangChainSplitter{
		SplitterClass: "RecursiveCharacterTextSplitter",
		ChunkSize:     1000,
		ChunkOverlap:  200,
		Separators:    []string{"\n\n", "\n", ". ", " "},
	},
	EntityResolver: &kg.SemanticMatchResolver{
		Threshold:       0.90,
		ResolveProperty: "name",
		Model:           "text-embedding-3-small",
	},
	OnError: "IGNORE",
}

// Vector Retriever for semantic search over chunks
// Reference: https://neo4j.com/labs/genai-ecosystem/graphrag-python/
var ChunkVectorRetriever = &retrievers.VectorRetriever{
	IndexName: "chunk-embeddings",
	EmbedderConfig: &retrievers.EmbedderConfig{
		Provider:   "openai",
		Model:      "text-embedding-3-large",
		Dimensions: 1536,
	},
	TopK:             5,
	ScoreThreshold:   0.7,
	ReturnProperties: []string{"text", "position", "pageNumber", "section"},
}

// Hybrid Retriever combining vector and fulltext search
var HybridChunkRetriever = &retrievers.HybridRetriever{
	VectorIndexName:   "chunk-embeddings",
	FulltextIndexName: "chunk-fulltext",
	EmbedderConfig: &retrievers.EmbedderConfig{
		Provider:   "openai",
		Model:      "text-embedding-3-large",
		Dimensions: 1536,
	},
	TopK:             5,
	VectorWeight:     0.7,
	FulltextWeight:   0.3,
	ReturnProperties: []string{"text", "position"},
}

// Text2Cypher Retriever for natural language queries
// Converts questions to Cypher queries for structured retrieval.
var Text2CypherRetriever = &retrievers.Text2CypherRetriever{
	LLMProvider: "openai",
	LLMModel:    "gpt-4",
	SchemaDescription: `
	The knowledge graph contains Documents, Chunks, Entities, Concepts, and Topics.
	Documents have chunks (HAS_CHUNK), chunks mention entities (MENTIONS),
	entities relate to each other (RELATES_TO), and documents are about concepts (ABOUT).
	Use vector search on chunk embeddings for semantic retrieval.
	`,
	Examples: []retrievers.CypherExample{
		{
			Question: "What are the main concepts discussed in documents about AI?",
			Cypher: `
				MATCH (d:Document)-[:ABOUT]->(c:Concept)
				WHERE c.domain = 'AI'
				RETURN d.title, c.name, c.definition
			`,
		},
		{
			Question: "Which organizations are mentioned in the same context as 'machine learning'?",
			Cypher: `
				MATCH (c:Chunk)-[:MENTIONS]->(e:Entity {name: 'machine learning'})
				MATCH (c)-[:MENTIONS]->(org:Entity)
				WHERE org.type = 'Organization'
				RETURN DISTINCT org.name, org.description
			`,
		},
	},
	MaxRetries: 3,
}

// GraphRAG Retriever with multi-hop traversal
// Retrieves chunks, then expands context via graph traversal.
var GraphRAGRetriever = &retrievers.VectorCypherRetriever{
	IndexName: "chunk-embeddings",
	EmbedderConfig: &retrievers.EmbedderConfig{
		Provider:   "openai",
		Model:      "text-embedding-3-large",
		Dimensions: 1536,
	},
	TopK:           3,
	ScoreThreshold: 0.7,
	// Retrieval query expands context by traversing to entities and related chunks
	RetrievalQuery: `
		MATCH (node:Chunk)
		WHERE node = chunk
		MATCH (node)-[:MENTIONS]->(e:Entity)<-[:MENTIONS]-(relatedChunk:Chunk)
		WHERE node.chunkId <> relatedChunk.chunkId
		RETURN node.text AS text,
		       node.position AS position,
		       collect({text: relatedChunk.text, entity: e.name}) AS relatedContext
		LIMIT 5
	`,
}
