// Package kg provides typed configurations for GraphRAG knowledge graph construction pipelines.
//
// This package implements type-safe configurations for KG construction including:
// - SimpleKGPipeline: Standard entity and relationship extraction
// - CustomKGPipeline: Custom extraction with user-defined prompts
// - Text splitters: FixedSizeSplitter, LangChainSplitter
// - Entity resolvers: ExactMatch, FuzzyMatch, SemanticMatch
//
// Example usage:
//
//	pipeline := &kg.SimpleKGPipeline{
//		Name: "document_kg",
//		EntityTypes: []kg.EntityType{
//			{Name: "Person", Description: "A human being"},
//		},
//		RelationTypes: []kg.RelationType{
//			{Name: "KNOWS", Description: "Social connection"},
//		},
//	}
//	config, err := serializer.ToJSON(pipeline)
package kg

// PipelineType represents the type of KG construction pipeline.
type PipelineType string

const (
	// SimpleKG is the standard SimpleKGPipeline from neo4j-graphrag.
	SimpleKG PipelineType = "SimpleKG"
	// CustomKG is a custom KG pipeline with user-defined prompts.
	CustomKG PipelineType = "CustomKG"
)

// KGPipeline is the interface that all KG construction pipelines implement.
type KGPipeline interface {
	// PipelineName returns the name of this pipeline.
	PipelineName() string
	// PipelineType returns the type of pipeline.
	PipelineType() PipelineType
}

// BasePipeline contains common pipeline configuration fields.
type BasePipeline struct {
	// Name is the pipeline name for identification.
	Name string
	// Neo4jURI is the connection URI for Neo4j.
	Neo4jURI string
	// Neo4jUser is the Neo4j username.
	Neo4jUser string
	// Neo4jPassword is the Neo4j password.
	Neo4jPassword string
	// Neo4jDatabase is the database name (default: neo4j).
	Neo4jDatabase string
	// LLMConfig configures the language model for extraction.
	LLMConfig *LLMConfig
	// EmbedderConfig configures the embedder for similarity.
	EmbedderConfig *EmbedderConfig
}

// PipelineName returns the pipeline name.
func (b *BasePipeline) PipelineName() string {
	return b.Name
}

// LLMConfig configures the language model.
type LLMConfig struct {
	// Provider is the LLM provider (openai, anthropic, etc.).
	Provider string
	// Model is the model name.
	Model string
	// APIKey is the API key for the provider.
	APIKey string
	// Temperature controls randomness (0-1).
	Temperature float64
	// MaxTokens is the maximum tokens to generate.
	MaxTokens int
	// TopP is nucleus sampling parameter.
	TopP float64
}

// EmbedderConfig configures the text embedder.
type EmbedderConfig struct {
	// Provider is the embedder provider (openai, etc.).
	Provider string
	// Model is the embedding model name.
	Model string
	// APIKey is the API key for the provider.
	APIKey string
	// Dimensions is the embedding dimension (optional).
	Dimensions int
}

// EntityType defines an entity type for extraction.
type EntityType struct {
	// Name is the entity type name (will become a node label).
	Name string
	// Description helps the LLM understand what to extract.
	Description string
	// Properties are the properties to extract for this entity.
	Properties []EntityProperty
}

// EntityProperty defines a property for an entity type.
type EntityProperty struct {
	// Name is the property name.
	Name string
	// Type is the property type (STRING, INTEGER, FLOAT, etc.).
	Type string
	// Description helps the LLM understand what to extract.
	Description string
	// Required indicates if this property must be present.
	Required bool
}

// RelationType defines a relationship type for extraction.
type RelationType struct {
	// Name is the relationship type name.
	Name string
	// Description helps the LLM understand when to create this relationship.
	Description string
	// SourceTypes are valid source entity types.
	SourceTypes []string
	// TargetTypes are valid target entity types.
	TargetTypes []string
	// Properties are the properties to extract for this relationship.
	Properties []RelationProperty
}

// RelationProperty defines a property for a relationship type.
type RelationProperty struct {
	// Name is the property name.
	Name string
	// Type is the property type.
	Type string
	// Description helps the LLM understand what to extract.
	Description string
}

// TextSplitter is the interface for text chunking strategies.
type TextSplitter interface {
	// SplitterType returns the type of splitter.
	SplitterType() string
}

// FixedSizeSplitter splits text into fixed-size chunks.
type FixedSizeSplitter struct {
	// ChunkSize is the size of each chunk in characters.
	ChunkSize int
	// ChunkOverlap is the overlap between chunks.
	ChunkOverlap int
}

func (s *FixedSizeSplitter) SplitterType() string { return "fixed_size" }

// LangChainSplitter uses a LangChain text splitter.
type LangChainSplitter struct {
	// SplitterClass is the LangChain splitter class name.
	SplitterClass string
	// ChunkSize is the target chunk size.
	ChunkSize int
	// ChunkOverlap is the overlap between chunks.
	ChunkOverlap int
	// Separators are the separators for recursive splitting.
	Separators []string
}

func (s *LangChainSplitter) SplitterType() string { return "langchain" }

// EntityResolver is the interface for entity resolution strategies.
type EntityResolver interface {
	// ResolverType returns the type of resolver.
	ResolverType() string
}

// ExactMatchResolver resolves entities by exact property match.
type ExactMatchResolver struct {
	// ResolveProperty is the property to match on.
	ResolveProperty string
}

func (r *ExactMatchResolver) ResolverType() string { return "exact_match" }

// FuzzyMatchResolver resolves entities using fuzzy string matching.
type FuzzyMatchResolver struct {
	// ResolveProperty is the property to match on.
	ResolveProperty string
	// Threshold is the similarity threshold (0-1).
	Threshold float64
}

func (r *FuzzyMatchResolver) ResolverType() string { return "fuzzy_match" }

// SemanticMatchResolver resolves entities using embedding similarity.
type SemanticMatchResolver struct {
	// ResolveProperty is the property to match on.
	ResolveProperty string
	// Threshold is the cosine similarity threshold (0-1).
	Threshold float64
	// Model is the spaCy or embedding model name.
	Model string
}

func (r *SemanticMatchResolver) ResolverType() string { return "semantic_match" }

// SimpleKGPipeline is the standard KG construction pipeline.
type SimpleKGPipeline struct {
	BasePipeline
	// EntityTypes are the entity types to extract.
	EntityTypes []EntityType
	// RelationTypes are the relationship types to extract.
	RelationTypes []RelationType
	// TextSplitter configures text chunking.
	TextSplitter TextSplitter
	// EntityResolver configures entity resolution.
	EntityResolver EntityResolver
	// PerformEntityResolution enables entity resolution (default: true).
	PerformEntityResolution *bool
	// FromPDF indicates input is PDF files.
	FromPDF bool
	// OnError specifies error handling ("RAISE", "IGNORE", "WARN").
	OnError string
}

func (p *SimpleKGPipeline) PipelineType() PipelineType { return SimpleKG }

// CustomKGPipeline allows custom extraction with user prompts.
type CustomKGPipeline struct {
	BasePipeline
	// ExtractionPrompt is the prompt template for entity extraction.
	// Use {text} placeholder for the input text.
	ExtractionPrompt string
	// SchemaPrompt describes the graph schema for the LLM.
	SchemaPrompt string
	// TextSplitter configures text chunking.
	TextSplitter TextSplitter
	// EntityResolver configures entity resolution.
	EntityResolver EntityResolver
	// OnError specifies error handling.
	OnError string
}

func (p *CustomKGPipeline) PipelineType() PipelineType { return CustomKG }
