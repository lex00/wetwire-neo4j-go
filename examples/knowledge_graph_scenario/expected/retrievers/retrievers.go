// Package retrievers provides retriever configurations for semantic search.
package retrievers

import (
	"github.com/lex00/wetwire-neo4j-go/internal/retrievers"
)

// DocumentRetriever implements hybrid search combining vector and fulltext retrieval.
// Returns top 10 most relevant documents based on weighted combination of similarity scores.
var DocumentRetriever = &retrievers.HybridRetriever{
	VectorIndexName:   "document_embedding_idx",
	FulltextIndexName: "document_content_fulltext",
	TopK:              10,
	VectorWeight:      0.7,
	FulltextWeight:    0.3,
	ReturnProperties:  []string{"id", "title", "content", "year"},
	EmbedderConfig: &retrievers.EmbedderConfig{
		Provider:   "openai",
		Model:      "text-embedding-3-small",
		Dimensions: 384,
	},
}
