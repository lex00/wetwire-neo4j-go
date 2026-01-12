package examples

import (
	"github.com/lex00/wetwire-neo4j-go/internal/kg"
)

// SimpleKGPipelineExample demonstrates basic knowledge graph construction.
// Reference: https://github.com/neo4j/neo4j-graphrag-python
var SimpleKGPipelineExample = &kg.SimpleKGPipeline{
	BasePipeline: kg.BasePipeline{
		Name: "news-article-kg",
		LLMConfig: &kg.LLMConfig{
			Provider:    "openai",
			Model:       "gpt-4",
			Temperature: 0.0,
			MaxTokens:   2000,
		},
		EmbedderConfig: &kg.EmbedderConfig{
			Provider:   "openai",
			Model:      "text-embedding-3-small",
			Dimensions: 384,
		},
	},
	EntityTypes: []kg.EntityType{
		{
			Name:        "Person",
			Description: "A human individual mentioned in the text",
			Properties: []kg.EntityProperty{
				{Name: "name", Type: "STRING", Required: true},
				{Name: "title", Type: "STRING"},
				{Name: "organization", Type: "STRING"},
			},
		},
		{
			Name:        "Organization",
			Description: "A company, government body, or other organization",
			Properties: []kg.EntityProperty{
				{Name: "name", Type: "STRING", Required: true},
				{Name: "type", Type: "STRING"},
				{Name: "location", Type: "STRING"},
			},
		},
		{
			Name:        "Location",
			Description: "A geographic location (city, country, region)",
			Properties: []kg.EntityProperty{
				{Name: "name", Type: "STRING", Required: true},
				{Name: "type", Type: "STRING"},
			},
		},
		{
			Name:        "Event",
			Description: "A notable event or occurrence",
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
			Description: "Employment relationship",
			SourceTypes: []string{"Person"},
			TargetTypes: []string{"Organization"},
		},
		{
			Name:        "LOCATED_IN",
			Description: "Geographic location of an entity",
			SourceTypes: []string{"Person", "Organization", "Event"},
			TargetTypes: []string{"Location"},
		},
		{
			Name:        "PARTICIPATED_IN",
			Description: "Participation in an event",
			SourceTypes: []string{"Person", "Organization"},
			TargetTypes: []string{"Event"},
		},
		{
			Name:        "RELATED_TO",
			Description: "General relationship between entities",
			SourceTypes: []string{"Person", "Organization"},
			TargetTypes: []string{"Person", "Organization"},
		},
	},
	TextSplitter: &kg.FixedSizeSplitter{
		ChunkSize:    500,
		ChunkOverlap: 50,
	},
	EntityResolver: &kg.FuzzyMatchResolver{
		Threshold:       0.85,
		ResolveProperty: "name",
	},
	OnError: "IGNORE",
}

// ScientificKGPipelineExample demonstrates KG construction for scientific papers.
var ScientificKGPipelineExample = &kg.SimpleKGPipeline{
	BasePipeline: kg.BasePipeline{
		Name: "scientific-paper-kg",
		LLMConfig: &kg.LLMConfig{
			Provider:    "anthropic",
			Model:       "claude-3-sonnet-20240229",
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
			Name:        "Researcher",
			Description: "A scientist or researcher who authored or is cited in papers",
			Properties: []kg.EntityProperty{
				{Name: "name", Type: "STRING", Required: true},
				{Name: "affiliation", Type: "STRING"},
				{Name: "orcid", Type: "STRING"},
			},
		},
		{
			Name:        "Paper",
			Description: "A scientific publication",
			Properties: []kg.EntityProperty{
				{Name: "title", Type: "STRING", Required: true},
				{Name: "year", Type: "INTEGER"},
				{Name: "doi", Type: "STRING"},
				{Name: "abstract", Type: "STRING"},
			},
		},
		{
			Name:        "Concept",
			Description: "A scientific concept, theory, or methodology",
			Properties: []kg.EntityProperty{
				{Name: "name", Type: "STRING", Required: true},
				{Name: "definition", Type: "STRING"},
				{Name: "field", Type: "STRING"},
			},
		},
		{
			Name:        "Institution",
			Description: "A research institution or university",
			Properties: []kg.EntityProperty{
				{Name: "name", Type: "STRING", Required: true},
				{Name: "country", Type: "STRING"},
			},
		},
	},
	RelationTypes: []kg.RelationType{
		{
			Name:        "AUTHORED",
			Description: "Authorship of a paper",
			SourceTypes: []string{"Researcher"},
			TargetTypes: []string{"Paper"},
		},
		{
			Name:        "CITES",
			Description: "Citation relationship between papers",
			SourceTypes: []string{"Paper"},
			TargetTypes: []string{"Paper"},
		},
		{
			Name:        "STUDIES",
			Description: "Research focus on a concept",
			SourceTypes: []string{"Paper", "Researcher"},
			TargetTypes: []string{"Concept"},
		},
		{
			Name:        "AFFILIATED_WITH",
			Description: "Researcher affiliation with institution",
			SourceTypes: []string{"Researcher"},
			TargetTypes: []string{"Institution"},
		},
		{
			Name:        "BUILDS_ON",
			Description: "Conceptual dependency between concepts",
			SourceTypes: []string{"Concept"},
			TargetTypes: []string{"Concept"},
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
}

// CustomKGPipelineExample demonstrates custom extraction prompts.
var CustomKGPipelineExample = &kg.CustomKGPipeline{
	BasePipeline: kg.BasePipeline{
		Name: "custom-extraction-kg",
		LLMConfig: &kg.LLMConfig{
			Provider:    "openai",
			Model:       "gpt-4-turbo",
			Temperature: 0.0,
			MaxTokens:   4000,
		},
	},
	ExtractionPrompt: `
You are an expert at extracting structured information from text.
Given the following text, extract entities and relationships according to the schema.

Text:
{text}

Schema:
{schema}

Output the extracted information as JSON with the following structure:
{
  "entities": [{"type": "...", "properties": {...}}],
  "relationships": [{"type": "...", "source": "...", "target": "...", "properties": {...}}]
}
	`,
	SchemaPrompt: `
Entity Types:
- Product: name (required), category, price, description
- Feature: name (required), description
- Review: rating (required), text, author

Relationship Types:
- HAS_FEATURE: Product -> Feature
- HAS_REVIEW: Product -> Review
- SIMILAR_TO: Product -> Product
	`,
	TextSplitter: &kg.FixedSizeSplitter{
		ChunkSize:    800,
		ChunkOverlap: 100,
	},
	OnError: "RAISE",
}

// PDFKGPipelineExample demonstrates KG construction from PDF documents.
var PDFKGPipelineExample = &kg.SimpleKGPipeline{
	BasePipeline: kg.BasePipeline{
		Name: "pdf-document-kg",
		LLMConfig: &kg.LLMConfig{
			Provider:    "openai",
			Model:       "gpt-4-vision-preview",
			Temperature: 0.0,
			MaxTokens:   4000,
		},
	},
	EntityTypes: []kg.EntityType{
		{
			Name:        "Section",
			Description: "A section or chapter in the document",
			Properties: []kg.EntityProperty{
				{Name: "title", Type: "STRING", Required: true},
				{Name: "pageNumber", Type: "INTEGER"},
			},
		},
		{
			Name:        "Figure",
			Description: "A figure, chart, or diagram",
			Properties: []kg.EntityProperty{
				{Name: "caption", Type: "STRING"},
				{Name: "pageNumber", Type: "INTEGER"},
			},
		},
		{
			Name:        "Table",
			Description: "A data table",
			Properties: []kg.EntityProperty{
				{Name: "caption", Type: "STRING"},
				{Name: "headers", Type: "LIST_STRING"},
			},
		},
	},
	RelationTypes: []kg.RelationType{
		{
			Name:        "CONTAINS",
			Description: "Section contains subsection or figure",
			SourceTypes: []string{"Section"},
			TargetTypes: []string{"Section", "Figure", "Table"},
		},
		{
			Name:        "REFERENCES",
			Description: "Cross-reference within document",
			SourceTypes: []string{"Section"},
			TargetTypes: []string{"Figure", "Table", "Section"},
		},
	},
	FromPDF: true,
	OnError: "IGNORE",
}

// AllKGPipelineExamples returns all example KG pipeline configurations.
func AllKGPipelineExamples() []kg.KGPipeline {
	return []kg.KGPipeline{
		SimpleKGPipelineExample,
		ScientificKGPipelineExample,
		CustomKGPipelineExample,
		PDFKGPipelineExample,
	}
}
