package retrievers

import (
	"encoding/json"
)

// RetrieverSerializer serializes retriever configurations to JSON.
type RetrieverSerializer struct{}

// NewRetrieverSerializer creates a new retriever serializer.
func NewRetrieverSerializer() *RetrieverSerializer {
	return &RetrieverSerializer{}
}

// ToJSON converts a retriever configuration to JSON.
func (s *RetrieverSerializer) ToJSON(retriever Retriever) ([]byte, error) {
	data := s.ToMap(retriever)
	return json.MarshalIndent(data, "", "  ")
}

// ToMap converts a retriever to a map.
func (s *RetrieverSerializer) ToMap(retriever Retriever) map[string]any {
	result := map[string]any{
		"name":          retriever.RetrieverName(),
		"retrieverType": string(retriever.RetrieverType()),
	}

	switch r := retriever.(type) {
	case *VectorRetriever:
		s.addBaseFields(result, &r.BaseRetriever)
		if r.IndexName != "" {
			result["indexName"] = r.IndexName
		}
		if r.EmbedderModel != "" {
			result["embedderModel"] = r.EmbedderModel
		}
		if r.EmbedderConfig != nil {
			result["embedderConfig"] = s.embedderConfigToMap(r.EmbedderConfig)
		}
		if r.TopK > 0 {
			result["topK"] = r.TopK
		}
		if len(r.ReturnProperties) > 0 {
			result["returnProperties"] = r.ReturnProperties
		}
		if r.ScoreThreshold > 0 {
			result["scoreThreshold"] = r.ScoreThreshold
		}

	case *VectorCypherRetriever:
		s.addBaseFields(result, &r.BaseRetriever)
		if r.IndexName != "" {
			result["indexName"] = r.IndexName
		}
		if r.EmbedderModel != "" {
			result["embedderModel"] = r.EmbedderModel
		}
		if r.EmbedderConfig != nil {
			result["embedderConfig"] = s.embedderConfigToMap(r.EmbedderConfig)
		}
		if r.RetrievalQuery != "" {
			result["retrievalQuery"] = r.RetrievalQuery
		}
		if r.TopK > 0 {
			result["topK"] = r.TopK
		}
		if r.ScoreThreshold > 0 {
			result["scoreThreshold"] = r.ScoreThreshold
		}

	case *HybridRetriever:
		s.addBaseFields(result, &r.BaseRetriever)
		if r.VectorIndexName != "" {
			result["vectorIndexName"] = r.VectorIndexName
		}
		if r.FulltextIndexName != "" {
			result["fulltextIndexName"] = r.FulltextIndexName
		}
		if r.EmbedderModel != "" {
			result["embedderModel"] = r.EmbedderModel
		}
		if r.EmbedderConfig != nil {
			result["embedderConfig"] = s.embedderConfigToMap(r.EmbedderConfig)
		}
		if r.TopK > 0 {
			result["topK"] = r.TopK
		}
		if len(r.ReturnProperties) > 0 {
			result["returnProperties"] = r.ReturnProperties
		}
		if r.VectorWeight > 0 {
			result["vectorWeight"] = r.VectorWeight
		}
		if r.FulltextWeight > 0 {
			result["fulltextWeight"] = r.FulltextWeight
		}

	case *HybridCypherRetriever:
		s.addBaseFields(result, &r.BaseRetriever)
		if r.VectorIndexName != "" {
			result["vectorIndexName"] = r.VectorIndexName
		}
		if r.FulltextIndexName != "" {
			result["fulltextIndexName"] = r.FulltextIndexName
		}
		if r.EmbedderModel != "" {
			result["embedderModel"] = r.EmbedderModel
		}
		if r.EmbedderConfig != nil {
			result["embedderConfig"] = s.embedderConfigToMap(r.EmbedderConfig)
		}
		if r.RetrievalQuery != "" {
			result["retrievalQuery"] = r.RetrievalQuery
		}
		if r.TopK > 0 {
			result["topK"] = r.TopK
		}
		if r.VectorWeight > 0 {
			result["vectorWeight"] = r.VectorWeight
		}
		if r.FulltextWeight > 0 {
			result["fulltextWeight"] = r.FulltextWeight
		}

	case *Text2CypherRetriever:
		s.addBaseFields(result, &r.BaseRetriever)
		if r.LLMModel != "" {
			result["llmModel"] = r.LLMModel
		}
		if r.LLMProvider != "" {
			result["llmProvider"] = r.LLMProvider
		}
		if r.LLMAPIKey != "" {
			result["llmApiKey"] = r.LLMAPIKey
		}
		if r.SchemaDescription != "" {
			result["schemaDescription"] = r.SchemaDescription
		}
		if len(r.Examples) > 0 {
			result["examples"] = s.examplesToMaps(r.Examples)
		}
		if r.MaxRetries > 0 {
			result["maxRetries"] = r.MaxRetries
		}

	case *WeaviateRetriever:
		s.addBaseFields(result, &r.BaseRetriever)
		if r.WeaviateURL != "" {
			result["weaviateUrl"] = r.WeaviateURL
		}
		if r.WeaviateAPIKey != "" {
			result["weaviateApiKey"] = r.WeaviateAPIKey
		}
		if r.Collection != "" {
			result["collection"] = r.Collection
		}
		if r.TopK > 0 {
			result["topK"] = r.TopK
		}
		if r.RetrievalQuery != "" {
			result["retrievalQuery"] = r.RetrievalQuery
		}
		if r.IDProperty != "" {
			result["idProperty"] = r.IDProperty
		}

	case *PineconeRetriever:
		s.addBaseFields(result, &r.BaseRetriever)
		if r.PineconeAPIKey != "" {
			result["pineconeApiKey"] = r.PineconeAPIKey
		}
		if r.PineconeHost != "" {
			result["pineconeHost"] = r.PineconeHost
		}
		if r.IndexName != "" {
			result["indexName"] = r.IndexName
		}
		if r.Namespace != "" {
			result["namespace"] = r.Namespace
		}
		if r.TopK > 0 {
			result["topK"] = r.TopK
		}
		if r.RetrievalQuery != "" {
			result["retrievalQuery"] = r.RetrievalQuery
		}
		if r.IDProperty != "" {
			result["idProperty"] = r.IDProperty
		}

	case *QdrantRetriever:
		s.addBaseFields(result, &r.BaseRetriever)
		if r.QdrantURL != "" {
			result["qdrantUrl"] = r.QdrantURL
		}
		if r.QdrantAPIKey != "" {
			result["qdrantApiKey"] = r.QdrantAPIKey
		}
		if r.CollectionName != "" {
			result["collectionName"] = r.CollectionName
		}
		if r.TopK > 0 {
			result["topK"] = r.TopK
		}
		if r.RetrievalQuery != "" {
			result["retrievalQuery"] = r.RetrievalQuery
		}
		if r.IDProperty != "" {
			result["idProperty"] = r.IDProperty
		}
	}

	return result
}

// addBaseFields adds common base retriever fields to the map.
func (s *RetrieverSerializer) addBaseFields(result map[string]any, base *BaseRetriever) {
	if base.Neo4jURI != "" {
		result["neo4jUri"] = base.Neo4jURI
	}
	if base.Neo4jUser != "" {
		result["neo4jUser"] = base.Neo4jUser
	}
	if base.Neo4jPassword != "" {
		result["neo4jPassword"] = base.Neo4jPassword
	}
	if base.Neo4jDatabase != "" {
		result["neo4jDatabase"] = base.Neo4jDatabase
	}
}

// embedderConfigToMap converts an EmbedderConfig to a map.
func (s *RetrieverSerializer) embedderConfigToMap(config *EmbedderConfig) map[string]any {
	result := make(map[string]any)
	if config.Provider != "" {
		result["provider"] = config.Provider
	}
	if config.Model != "" {
		result["model"] = config.Model
	}
	if config.APIKey != "" {
		result["apiKey"] = config.APIKey
	}
	if config.Dimensions > 0 {
		result["dimensions"] = config.Dimensions
	}
	return result
}

// examplesToMaps converts CypherExamples to maps.
func (s *RetrieverSerializer) examplesToMaps(examples []CypherExample) []map[string]string {
	result := make([]map[string]string, len(examples))
	for i, ex := range examples {
		result[i] = map[string]string{
			"question": ex.Question,
			"cypher":   ex.Cypher,
		}
	}
	return result
}

// BatchToJSON converts multiple retrievers to a JSON array.
func (s *RetrieverSerializer) BatchToJSON(retrievers []Retriever) ([]byte, error) {
	configs := make([]map[string]any, len(retrievers))
	for i, r := range retrievers {
		configs[i] = s.ToMap(r)
	}
	return json.MarshalIndent(configs, "", "  ")
}
