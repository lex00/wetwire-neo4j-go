package retrievers

import (
	"encoding/json"
	"testing"
)

func TestVectorRetriever_Interface(t *testing.T) {
	r := &VectorRetriever{
		BaseRetriever: BaseRetriever{
			Name: "my_vector_retriever",
		},
		IndexName:        "document_embeddings",
		EmbedderModel:    "text-embedding-3-small",
		TopK:             10,
		ReturnProperties: []string{"text", "title"},
	}

	if r.RetrieverName() != "my_vector_retriever" {
		t.Errorf("RetrieverName() = %v, want my_vector_retriever", r.RetrieverName())
	}
	if r.RetrieverType() != Vector {
		t.Errorf("RetrieverType() = %v, want Vector", r.RetrieverType())
	}
}

func TestVectorCypherRetriever_Interface(t *testing.T) {
	r := &VectorCypherRetriever{
		BaseRetriever: BaseRetriever{
			Name: "cypher_vector",
		},
		IndexName: "embeddings",
		RetrievalQuery: `
			MATCH (node)-[:RELATES_TO]->(related)
			RETURN node.text AS text, collect(related.name) AS related_entities
		`,
	}

	if r.RetrieverType() != VectorCypher {
		t.Errorf("RetrieverType() = %v, want VectorCypher", r.RetrieverType())
	}
	if r.RetrievalQuery == "" {
		t.Error("RetrievalQuery should not be empty")
	}
}

func TestHybridRetriever_Interface(t *testing.T) {
	r := &HybridRetriever{
		BaseRetriever: BaseRetriever{
			Name: "hybrid_search",
		},
		VectorIndexName:   "vector_idx",
		FulltextIndexName: "fulltext_idx",
		TopK:              5,
	}

	if r.RetrieverType() != Hybrid {
		t.Errorf("RetrieverType() = %v, want Hybrid", r.RetrieverType())
	}
}

func TestHybridCypherRetriever_Interface(t *testing.T) {
	r := &HybridCypherRetriever{
		BaseRetriever: BaseRetriever{
			Name: "hybrid_cypher",
		},
		VectorIndexName:   "vec_idx",
		FulltextIndexName: "ft_idx",
		RetrievalQuery:    "MATCH (n) RETURN n",
	}

	if r.RetrieverType() != HybridCypher {
		t.Errorf("RetrieverType() = %v, want HybridCypher", r.RetrieverType())
	}
}

func TestText2CypherRetriever_Interface(t *testing.T) {
	r := &Text2CypherRetriever{
		BaseRetriever: BaseRetriever{
			Name: "text_to_cypher",
		},
		LLMModel: "claude-3-5-sonnet",
		Examples: []CypherExample{
			{
				Question: "Who directed Inception?",
				Cypher:   "MATCH (m:Movie {title: 'Inception'})<-[:DIRECTED]-(d) RETURN d.name",
			},
		},
	}

	if r.RetrieverType() != Text2Cypher {
		t.Errorf("RetrieverType() = %v, want Text2Cypher", r.RetrieverType())
	}
	if len(r.Examples) != 1 {
		t.Errorf("Examples count = %v, want 1", len(r.Examples))
	}
}

func TestWeaviateRetriever_Interface(t *testing.T) {
	r := &WeaviateRetriever{
		BaseRetriever: BaseRetriever{
			Name: "weaviate_retriever",
		},
		WeaviateURL:    "http://localhost:8080",
		WeaviateAPIKey: "api-key",
		Collection:     "Documents",
	}

	if r.RetrieverType() != Weaviate {
		t.Errorf("RetrieverType() = %v, want Weaviate", r.RetrieverType())
	}
}

func TestPineconeRetriever_Interface(t *testing.T) {
	r := &PineconeRetriever{
		BaseRetriever: BaseRetriever{
			Name: "pinecone_retriever",
		},
		PineconeAPIKey: "api-key",
		PineconeHost:   "https://my-index.pinecone.io",
		IndexName:      "my-index",
	}

	if r.RetrieverType() != Pinecone {
		t.Errorf("RetrieverType() = %v, want Pinecone", r.RetrieverType())
	}
}

func TestQdrantRetriever_Interface(t *testing.T) {
	r := &QdrantRetriever{
		BaseRetriever: BaseRetriever{
			Name: "qdrant_retriever",
		},
		QdrantURL:      "http://localhost:6333",
		CollectionName: "documents",
	}

	if r.RetrieverType() != Qdrant {
		t.Errorf("RetrieverType() = %v, want Qdrant", r.RetrieverType())
	}
}

func TestNewRetrieverSerializer(t *testing.T) {
	s := NewRetrieverSerializer()
	if s == nil {
		t.Fatal("NewRetrieverSerializer returned nil")
	}
}

func TestRetrieverSerializer_ToJSON_VectorRetriever(t *testing.T) {
	s := NewRetrieverSerializer()
	r := &VectorRetriever{
		BaseRetriever: BaseRetriever{
			Name: "doc_retriever",
		},
		IndexName:        "document_embeddings",
		EmbedderModel:    "text-embedding-3-small",
		TopK:             10,
		ReturnProperties: []string{"text", "title", "source"},
	}

	result, err := s.ToJSON(r)
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}

	if parsed["name"] != "doc_retriever" {
		t.Errorf("name = %v, want doc_retriever", parsed["name"])
	}
	if parsed["retrieverType"] != "Vector" {
		t.Errorf("retrieverType = %v, want Vector", parsed["retrieverType"])
	}
	if parsed["indexName"] != "document_embeddings" {
		t.Errorf("indexName = %v, want document_embeddings", parsed["indexName"])
	}
	if parsed["topK"] != float64(10) {
		t.Errorf("topK = %v, want 10", parsed["topK"])
	}

	props, ok := parsed["returnProperties"].([]any)
	if !ok || len(props) != 3 {
		t.Error("expected 3 return properties")
	}
}

func TestRetrieverSerializer_ToJSON_VectorCypherRetriever(t *testing.T) {
	s := NewRetrieverSerializer()
	r := &VectorCypherRetriever{
		BaseRetriever: BaseRetriever{
			Name: "cypher_retriever",
		},
		IndexName:      "embeddings",
		RetrievalQuery: "MATCH (n)-[:RELATED_TO]->(m) RETURN n, m",
		TopK:           5,
	}

	result, err := s.ToJSON(r)
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}

	if parsed["retrieverType"] != "VectorCypher" {
		t.Errorf("retrieverType = %v, want VectorCypher", parsed["retrieverType"])
	}
	if parsed["retrievalQuery"] == nil {
		t.Error("expected retrievalQuery to be present")
	}
}

func TestRetrieverSerializer_ToJSON_HybridRetriever(t *testing.T) {
	s := NewRetrieverSerializer()
	r := &HybridRetriever{
		BaseRetriever: BaseRetriever{
			Name: "hybrid",
		},
		VectorIndexName:   "vec_index",
		FulltextIndexName: "ft_index",
		TopK:              10,
	}

	result, err := s.ToJSON(r)
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}

	if parsed["retrieverType"] != "Hybrid" {
		t.Errorf("retrieverType = %v, want Hybrid", parsed["retrieverType"])
	}
	if parsed["vectorIndexName"] != "vec_index" {
		t.Errorf("vectorIndexName = %v, want vec_index", parsed["vectorIndexName"])
	}
	if parsed["fulltextIndexName"] != "ft_index" {
		t.Errorf("fulltextIndexName = %v, want ft_index", parsed["fulltextIndexName"])
	}
}

func TestRetrieverSerializer_ToJSON_Text2CypherRetriever(t *testing.T) {
	s := NewRetrieverSerializer()
	r := &Text2CypherRetriever{
		BaseRetriever: BaseRetriever{
			Name: "nl_to_cypher",
		},
		LLMModel: "claude-3-5-sonnet",
		Examples: []CypherExample{
			{Question: "Find all actors", Cypher: "MATCH (a:Actor) RETURN a"},
		},
		SchemaDescription: "Movie database with Actor, Director, Movie nodes",
	}

	result, err := s.ToJSON(r)
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}

	if parsed["retrieverType"] != "Text2Cypher" {
		t.Errorf("retrieverType = %v, want Text2Cypher", parsed["retrieverType"])
	}
	if parsed["llmModel"] != "claude-3-5-sonnet" {
		t.Errorf("llmModel = %v, want claude-3-5-sonnet", parsed["llmModel"])
	}

	examples, ok := parsed["examples"].([]any)
	if !ok || len(examples) != 1 {
		t.Error("expected 1 example")
	}
}

func TestRetrieverSerializer_ToMap(t *testing.T) {
	s := NewRetrieverSerializer()
	r := &VectorRetriever{
		BaseRetriever: BaseRetriever{
			Name: "test",
		},
		IndexName: "idx",
		TopK:      5,
	}

	result := s.ToMap(r)

	if result["name"] != "test" {
		t.Errorf("name = %v, want test", result["name"])
	}
	if result["indexName"] != "idx" {
		t.Errorf("indexName = %v, want idx", result["indexName"])
	}
}

func TestRetrieverSerializer_BatchToJSON(t *testing.T) {
	s := NewRetrieverSerializer()
	retrievers := []Retriever{
		&VectorRetriever{
			BaseRetriever: BaseRetriever{Name: "vec"},
			IndexName:     "vec_idx",
		},
		&HybridRetriever{
			BaseRetriever:     BaseRetriever{Name: "hyb"},
			VectorIndexName:   "v",
			FulltextIndexName: "f",
		},
	}

	result, err := s.BatchToJSON(retrievers)
	if err != nil {
		t.Fatalf("BatchToJSON failed: %v", err)
	}

	var parsed []map[string]any
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}

	if len(parsed) != 2 {
		t.Errorf("expected 2 retrievers, got %d", len(parsed))
	}
}

func TestRetrieverTypes_Values(t *testing.T) {
	tests := []struct {
		r    Retriever
		want RetrieverType
	}{
		{&VectorRetriever{}, Vector},
		{&VectorCypherRetriever{}, VectorCypher},
		{&HybridRetriever{}, Hybrid},
		{&HybridCypherRetriever{}, HybridCypher},
		{&Text2CypherRetriever{}, Text2Cypher},
		{&WeaviateRetriever{}, Weaviate},
		{&PineconeRetriever{}, Pinecone},
		{&QdrantRetriever{}, Qdrant},
	}

	for _, tt := range tests {
		t.Run(string(tt.want), func(t *testing.T) {
			if tt.r.RetrieverType() != tt.want {
				t.Errorf("RetrieverType() = %v, want %v", tt.r.RetrieverType(), tt.want)
			}
		})
	}
}

func TestRetriever_ImplementsInterface(t *testing.T) {
	// Verify all retrievers implement the Retriever interface
	var _ Retriever = &VectorRetriever{}
	var _ Retriever = &VectorCypherRetriever{}
	var _ Retriever = &HybridRetriever{}
	var _ Retriever = &HybridCypherRetriever{}
	var _ Retriever = &Text2CypherRetriever{}
	var _ Retriever = &WeaviateRetriever{}
	var _ Retriever = &PineconeRetriever{}
	var _ Retriever = &QdrantRetriever{}
}

func TestCypherExample_Structure(t *testing.T) {
	example := CypherExample{
		Question: "How many movies are in the database?",
		Cypher:   "MATCH (m:Movie) RETURN count(m)",
	}

	if example.Question == "" {
		t.Error("Question should not be empty")
	}
	if example.Cypher == "" {
		t.Error("Cypher should not be empty")
	}
}

func TestEmbedderConfig_Structure(t *testing.T) {
	config := EmbedderConfig{
		Provider: "openai",
		Model:    "text-embedding-3-small",
		APIKey:   "sk-xxx",
	}

	if config.Provider == "" {
		t.Error("Provider should not be empty")
	}
	if config.Model == "" {
		t.Error("Model should not be empty")
	}
}
