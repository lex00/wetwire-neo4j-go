// Package retrievers provides typed configurations for neo4j-graphrag retriever types.
//
// This package implements type-safe configurations for GraphRAG retrievers including:
// - VectorRetriever: Similarity search using vector indexes
// - VectorCypherRetriever: Vector search with custom graph traversal
// - HybridRetriever: Combined vector and fulltext search
// - HybridCypherRetriever: Hybrid search with custom traversal
// - Text2CypherRetriever: LLM-generated Cypher queries
// - External retrievers: Weaviate, Pinecone, Qdrant integration
//
// Example usage:
//
//	retriever := &retrievers.VectorRetriever{
//		Name:             "doc_search",
//		IndexName:        "document_embeddings",
//		EmbedderModel:    "text-embedding-3-small",
//		TopK:             10,
//	}
//	config, err := serializer.ToJSON(retriever)
package retrievers

// RetrieverType represents the type of GraphRAG retriever.
type RetrieverType string

const (
	// Vector retriever for similarity search.
	Vector RetrieverType = "Vector"
	// VectorCypher retriever for vector search with graph traversal.
	VectorCypher RetrieverType = "VectorCypher"
	// Hybrid retriever combining vector and fulltext search.
	Hybrid RetrieverType = "Hybrid"
	// HybridCypher retriever combining hybrid search with traversal.
	HybridCypher RetrieverType = "HybridCypher"
	// Text2Cypher retriever using LLM to generate Cypher.
	Text2Cypher RetrieverType = "Text2Cypher"
	// Weaviate retriever for external Weaviate integration.
	Weaviate RetrieverType = "Weaviate"
	// Pinecone retriever for external Pinecone integration.
	Pinecone RetrieverType = "Pinecone"
	// Qdrant retriever for external Qdrant integration.
	Qdrant RetrieverType = "Qdrant"
)

// Retriever is the interface that all GraphRAG retriever configurations implement.
type Retriever interface {
	// RetrieverName returns the name of this retriever.
	RetrieverName() string
	// RetrieverType returns the type of retriever.
	RetrieverType() RetrieverType
}

// BaseRetriever contains common retriever configuration fields.
type BaseRetriever struct {
	// Name is the retriever name for identification.
	Name string
	// Neo4jURI is the connection URI for Neo4j.
	Neo4jURI string
	// Neo4jUser is the Neo4j username.
	Neo4jUser string
	// Neo4jPassword is the Neo4j password.
	Neo4jPassword string
	// Neo4jDatabase is the database name (default: neo4j).
	Neo4jDatabase string
}

// RetrieverName returns the retriever name.
func (b *BaseRetriever) RetrieverName() string {
	return b.Name
}

// EmbedderConfig configures the text embedder.
type EmbedderConfig struct {
	// Provider is the embedder provider (openai, anthropic, etc.).
	Provider string
	// Model is the embedding model name.
	Model string
	// APIKey is the API key for the provider.
	APIKey string
	// Dimensions is the embedding dimension (optional).
	Dimensions int
}

// CypherExample provides a question-to-Cypher example for Text2Cypher.
type CypherExample struct {
	// Question is a natural language question.
	Question string
	// Cypher is the corresponding Cypher query.
	Cypher string
}

// VectorRetriever performs similarity search using Neo4j vector indexes.
type VectorRetriever struct {
	BaseRetriever
	// IndexName is the name of the vector index to search.
	IndexName string
	// EmbedderModel is the embedding model name for query embedding.
	EmbedderModel string
	// EmbedderConfig provides detailed embedder configuration.
	EmbedderConfig *EmbedderConfig
	// TopK is the number of results to return (default: 5).
	TopK int
	// ReturnProperties are node properties to return.
	ReturnProperties []string
	// ScoreThreshold filters results below this similarity score.
	ScoreThreshold float64
}

func (r *VectorRetriever) RetrieverType() RetrieverType { return Vector }

// VectorCypherRetriever combines vector search with custom graph traversal.
type VectorCypherRetriever struct {
	BaseRetriever
	// IndexName is the name of the vector index to search.
	IndexName string
	// EmbedderModel is the embedding model name.
	EmbedderModel string
	// EmbedderConfig provides detailed embedder configuration.
	EmbedderConfig *EmbedderConfig
	// RetrievalQuery is the Cypher query for post-search traversal.
	// Use $node to reference the matched node.
	RetrievalQuery string
	// TopK is the number of results to return (default: 5).
	TopK int
	// ScoreThreshold filters results below this similarity score.
	ScoreThreshold float64
}

func (r *VectorCypherRetriever) RetrieverType() RetrieverType { return VectorCypher }

// HybridRetriever combines vector and fulltext search.
type HybridRetriever struct {
	BaseRetriever
	// VectorIndexName is the name of the vector index.
	VectorIndexName string
	// FulltextIndexName is the name of the fulltext index.
	FulltextIndexName string
	// EmbedderModel is the embedding model name.
	EmbedderModel string
	// EmbedderConfig provides detailed embedder configuration.
	EmbedderConfig *EmbedderConfig
	// TopK is the number of results to return (default: 5).
	TopK int
	// ReturnProperties are node properties to return.
	ReturnProperties []string
	// VectorWeight is the weight for vector search (0-1).
	VectorWeight float64
	// FulltextWeight is the weight for fulltext search (0-1).
	FulltextWeight float64
}

func (r *HybridRetriever) RetrieverType() RetrieverType { return Hybrid }

// HybridCypherRetriever combines hybrid search with custom traversal.
type HybridCypherRetriever struct {
	BaseRetriever
	// VectorIndexName is the name of the vector index.
	VectorIndexName string
	// FulltextIndexName is the name of the fulltext index.
	FulltextIndexName string
	// EmbedderModel is the embedding model name.
	EmbedderModel string
	// EmbedderConfig provides detailed embedder configuration.
	EmbedderConfig *EmbedderConfig
	// RetrievalQuery is the Cypher query for post-search traversal.
	RetrievalQuery string
	// TopK is the number of results to return (default: 5).
	TopK int
	// VectorWeight is the weight for vector search (0-1).
	VectorWeight float64
	// FulltextWeight is the weight for fulltext search (0-1).
	FulltextWeight float64
}

func (r *HybridCypherRetriever) RetrieverType() RetrieverType { return HybridCypher }

// Text2CypherRetriever uses an LLM to generate Cypher from natural language.
type Text2CypherRetriever struct {
	BaseRetriever
	// LLMModel is the LLM model name for Cypher generation.
	LLMModel string
	// LLMProvider is the LLM provider (openai, anthropic, etc.).
	LLMProvider string
	// LLMAPIKey is the API key for the LLM provider.
	LLMAPIKey string
	// SchemaDescription describes the graph schema for the LLM.
	SchemaDescription string
	// Examples are question-to-Cypher examples for few-shot learning.
	Examples []CypherExample
	// MaxRetries is the maximum retries on Cypher generation failure.
	MaxRetries int
}

func (r *Text2CypherRetriever) RetrieverType() RetrieverType { return Text2Cypher }

// WeaviateRetriever integrates with external Weaviate vector database.
type WeaviateRetriever struct {
	BaseRetriever
	// WeaviateURL is the Weaviate server URL.
	WeaviateURL string
	// WeaviateAPIKey is the Weaviate API key.
	WeaviateAPIKey string
	// Collection is the Weaviate collection name.
	Collection string
	// TopK is the number of results to return.
	TopK int
	// RetrievalQuery is optional Cypher for Neo4j traversal.
	RetrievalQuery string
	// IDProperty is the Neo4j property containing Weaviate IDs.
	IDProperty string
}

func (r *WeaviateRetriever) RetrieverType() RetrieverType { return Weaviate }

// PineconeRetriever integrates with external Pinecone vector database.
type PineconeRetriever struct {
	BaseRetriever
	// PineconeAPIKey is the Pinecone API key.
	PineconeAPIKey string
	// PineconeHost is the Pinecone index host URL.
	PineconeHost string
	// IndexName is the Pinecone index name.
	IndexName string
	// Namespace is the optional Pinecone namespace.
	Namespace string
	// TopK is the number of results to return.
	TopK int
	// RetrievalQuery is optional Cypher for Neo4j traversal.
	RetrievalQuery string
	// IDProperty is the Neo4j property containing Pinecone IDs.
	IDProperty string
}

func (r *PineconeRetriever) RetrieverType() RetrieverType { return Pinecone }

// QdrantRetriever integrates with external Qdrant vector database.
type QdrantRetriever struct {
	BaseRetriever
	// QdrantURL is the Qdrant server URL.
	QdrantURL string
	// QdrantAPIKey is the optional Qdrant API key.
	QdrantAPIKey string
	// CollectionName is the Qdrant collection name.
	CollectionName string
	// TopK is the number of results to return.
	TopK int
	// RetrievalQuery is optional Cypher for Neo4j traversal.
	RetrievalQuery string
	// IDProperty is the Neo4j property containing Qdrant IDs.
	IDProperty string
}

func (r *QdrantRetriever) RetrieverType() RetrieverType { return Qdrant }
