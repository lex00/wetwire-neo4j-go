// Package extraction provides KG pipeline configuration for entity extraction.
package extraction

import (
	"github.com/lex00/wetwire-neo4j-go/internal/kg"
)

// ResearchPaperKG defines the knowledge graph extraction pipeline for academic papers.
// Uses GPT-4 for entity extraction and text-embedding-3-small for vector embeddings.
var ResearchPaperKG = &kg.SimpleKGPipeline{
	BasePipeline: kg.BasePipeline{
		Name: "research-paper-kg",
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
			Description: "A researcher or author",
			Properties: []kg.EntityProperty{
				{Name: "name", Type: "STRING", Required: true},
				{Name: "affiliation", Type: "STRING"},
				{Name: "orcid", Type: "STRING"},
			},
		},
		{
			Name:        "Document",
			Description: "A research paper or publication",
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
			SourceTypes: []string{"Person"},
			TargetTypes: []string{"Document"},
		},
		{
			Name:        "CITES",
			Description: "Citation relationship between papers",
			SourceTypes: []string{"Document"},
			TargetTypes: []string{"Document"},
		},
		{
			Name:        "STUDIES",
			Description: "Research focus on a concept",
			SourceTypes: []string{"Document"},
			TargetTypes: []string{"Concept"},
		},
		{
			Name:        "AFFILIATED_WITH",
			Description: "Researcher affiliation with institution",
			SourceTypes: []string{"Person"},
			TargetTypes: []string{"Institution"},
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
