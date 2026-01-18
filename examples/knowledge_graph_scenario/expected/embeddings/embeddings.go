// Package embeddings provides embedding configuration for the knowledge graph.
package embeddings

import (
	"github.com/lex00/wetwire-neo4j-go/internal/retrievers"
)

// DocumentEmbedder defines the embedder configuration for document vectorization.
// Uses OpenAI's text-embedding-3-small model with 384 dimensions.
var DocumentEmbedder = &retrievers.EmbedderConfig{
	Provider:   "openai",
	Model:      "text-embedding-3-small",
	Dimensions: 384,
}
