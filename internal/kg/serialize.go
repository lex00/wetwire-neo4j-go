package kg

import (
	"encoding/json"
	"fmt"
	"strings"
)

// KGSerializer serializes KG pipeline configurations to JSON.
type KGSerializer struct{}

// NewKGSerializer creates a new KG serializer.
func NewKGSerializer() *KGSerializer {
	return &KGSerializer{}
}

// ToJSON converts a KG pipeline configuration to JSON.
func (s *KGSerializer) ToJSON(pipeline KGPipeline) ([]byte, error) {
	data := s.ToMap(pipeline)
	return json.MarshalIndent(data, "", "  ")
}

// ToMap converts a KG pipeline to a map.
func (s *KGSerializer) ToMap(pipeline KGPipeline) map[string]any {
	result := map[string]any{
		"name":         pipeline.PipelineName(),
		"pipelineType": string(pipeline.PipelineType()),
	}

	switch p := pipeline.(type) {
	case *SimpleKGPipeline:
		s.addBaseFields(result, &p.BasePipeline)
		if len(p.EntityTypes) > 0 {
			result["entityTypes"] = s.entityTypesToMaps(p.EntityTypes)
		}
		if len(p.RelationTypes) > 0 {
			result["relationTypes"] = s.relationTypesToMaps(p.RelationTypes)
		}
		if p.TextSplitter != nil {
			result["textSplitter"] = s.textSplitterToMap(p.TextSplitter)
		}
		if p.EntityResolver != nil {
			result["entityResolver"] = s.entityResolverToMap(p.EntityResolver)
		}
		if p.PerformEntityResolution != nil {
			result["performEntityResolution"] = *p.PerformEntityResolution
		}
		if p.FromPDF {
			result["fromPDF"] = p.FromPDF
		}
		if p.OnError != "" {
			result["onError"] = p.OnError
		}

	case *CustomKGPipeline:
		s.addBaseFields(result, &p.BasePipeline)
		if p.ExtractionPrompt != "" {
			result["extractionPrompt"] = p.ExtractionPrompt
		}
		if p.SchemaPrompt != "" {
			result["schemaPrompt"] = p.SchemaPrompt
		}
		if p.TextSplitter != nil {
			result["textSplitter"] = s.textSplitterToMap(p.TextSplitter)
		}
		if p.EntityResolver != nil {
			result["entityResolver"] = s.entityResolverToMap(p.EntityResolver)
		}
		if p.OnError != "" {
			result["onError"] = p.OnError
		}
	}

	return result
}

// addBaseFields adds common base pipeline fields to the map.
func (s *KGSerializer) addBaseFields(result map[string]any, base *BasePipeline) {
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
	if base.LLMConfig != nil {
		result["llmConfig"] = s.llmConfigToMap(base.LLMConfig)
	}
	if base.EmbedderConfig != nil {
		result["embedderConfig"] = s.embedderConfigToMap(base.EmbedderConfig)
	}
}

// llmConfigToMap converts an LLMConfig to a map.
func (s *KGSerializer) llmConfigToMap(config *LLMConfig) map[string]any {
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
	if config.Temperature > 0 {
		result["temperature"] = config.Temperature
	}
	if config.MaxTokens > 0 {
		result["maxTokens"] = config.MaxTokens
	}
	if config.TopP > 0 {
		result["topP"] = config.TopP
	}
	return result
}

// embedderConfigToMap converts an EmbedderConfig to a map.
func (s *KGSerializer) embedderConfigToMap(config *EmbedderConfig) map[string]any {
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

// entityTypesToMaps converts EntityTypes to maps.
func (s *KGSerializer) entityTypesToMaps(types []EntityType) []map[string]any {
	result := make([]map[string]any, len(types))
	for i, et := range types {
		m := map[string]any{"name": et.Name}
		if et.Description != "" {
			m["description"] = et.Description
		}
		if len(et.Properties) > 0 {
			m["properties"] = s.entityPropertiesToMaps(et.Properties)
		}
		result[i] = m
	}
	return result
}

// entityPropertiesToMaps converts EntityProperties to maps.
func (s *KGSerializer) entityPropertiesToMaps(props []EntityProperty) []map[string]any {
	result := make([]map[string]any, len(props))
	for i, prop := range props {
		m := map[string]any{"name": prop.Name}
		if prop.Type != "" {
			m["type"] = prop.Type
		}
		if prop.Description != "" {
			m["description"] = prop.Description
		}
		if prop.Required {
			m["required"] = prop.Required
		}
		result[i] = m
	}
	return result
}

// relationTypesToMaps converts RelationTypes to maps.
func (s *KGSerializer) relationTypesToMaps(types []RelationType) []map[string]any {
	result := make([]map[string]any, len(types))
	for i, rt := range types {
		m := map[string]any{"name": rt.Name}
		if rt.Description != "" {
			m["description"] = rt.Description
		}
		if len(rt.SourceTypes) > 0 {
			m["sourceTypes"] = rt.SourceTypes
		}
		if len(rt.TargetTypes) > 0 {
			m["targetTypes"] = rt.TargetTypes
		}
		if len(rt.Properties) > 0 {
			m["properties"] = s.relationPropertiesToMaps(rt.Properties)
		}
		result[i] = m
	}
	return result
}

// relationPropertiesToMaps converts RelationProperties to maps.
func (s *KGSerializer) relationPropertiesToMaps(props []RelationProperty) []map[string]any {
	result := make([]map[string]any, len(props))
	for i, prop := range props {
		m := map[string]any{"name": prop.Name}
		if prop.Type != "" {
			m["type"] = prop.Type
		}
		if prop.Description != "" {
			m["description"] = prop.Description
		}
		result[i] = m
	}
	return result
}

// textSplitterToMap converts a TextSplitter to a map.
func (s *KGSerializer) textSplitterToMap(splitter TextSplitter) map[string]any {
	result := map[string]any{
		"type": splitter.SplitterType(),
	}

	switch sp := splitter.(type) {
	case *FixedSizeSplitter:
		if sp.ChunkSize > 0 {
			result["chunkSize"] = sp.ChunkSize
		}
		if sp.ChunkOverlap > 0 {
			result["chunkOverlap"] = sp.ChunkOverlap
		}
	case *LangChainSplitter:
		if sp.SplitterClass != "" {
			result["splitterClass"] = sp.SplitterClass
		}
		if sp.ChunkSize > 0 {
			result["chunkSize"] = sp.ChunkSize
		}
		if sp.ChunkOverlap > 0 {
			result["chunkOverlap"] = sp.ChunkOverlap
		}
		if len(sp.Separators) > 0 {
			result["separators"] = sp.Separators
		}
	}

	return result
}

// entityResolverToMap converts an EntityResolver to a map.
func (s *KGSerializer) entityResolverToMap(resolver EntityResolver) map[string]any {
	result := map[string]any{
		"type": resolver.ResolverType(),
	}

	switch r := resolver.(type) {
	case *ExactMatchResolver:
		if r.ResolveProperty != "" {
			result["resolveProperty"] = r.ResolveProperty
		}
	case *FuzzyMatchResolver:
		if r.ResolveProperty != "" {
			result["resolveProperty"] = r.ResolveProperty
		}
		if r.Threshold > 0 {
			result["threshold"] = r.Threshold
		}
	case *SemanticMatchResolver:
		if r.ResolveProperty != "" {
			result["resolveProperty"] = r.ResolveProperty
		}
		if r.Threshold > 0 {
			result["threshold"] = r.Threshold
		}
		if r.Model != "" {
			result["model"] = r.Model
		}
	}

	return result
}

// GenerateSchema generates a schema description for LLM extraction.
func (s *KGSerializer) GenerateSchema(pipeline *SimpleKGPipeline) string {
	var sb strings.Builder

	sb.WriteString("Graph Schema:\n\n")

	if len(pipeline.EntityTypes) > 0 {
		sb.WriteString("Entity Types:\n")
		for _, et := range pipeline.EntityTypes {
			sb.WriteString(fmt.Sprintf("- %s", et.Name))
			if et.Description != "" {
				sb.WriteString(fmt.Sprintf(": %s", et.Description))
			}
			sb.WriteString("\n")
			for _, prop := range et.Properties {
				sb.WriteString(fmt.Sprintf("  - %s (%s)", prop.Name, prop.Type))
				if prop.Description != "" {
					sb.WriteString(fmt.Sprintf(": %s", prop.Description))
				}
				sb.WriteString("\n")
			}
		}
		sb.WriteString("\n")
	}

	if len(pipeline.RelationTypes) > 0 {
		sb.WriteString("Relationship Types:\n")
		for _, rt := range pipeline.RelationTypes {
			sb.WriteString(fmt.Sprintf("- %s", rt.Name))
			if rt.Description != "" {
				sb.WriteString(fmt.Sprintf(": %s", rt.Description))
			}
			sb.WriteString("\n")
			if len(rt.SourceTypes) > 0 && len(rt.TargetTypes) > 0 {
				sb.WriteString(fmt.Sprintf("  Source: %s, Target: %s\n",
					strings.Join(rt.SourceTypes, ", "),
					strings.Join(rt.TargetTypes, ", ")))
			}
			for _, prop := range rt.Properties {
				sb.WriteString(fmt.Sprintf("  - %s (%s)", prop.Name, prop.Type))
				if prop.Description != "" {
					sb.WriteString(fmt.Sprintf(": %s", prop.Description))
				}
				sb.WriteString("\n")
			}
		}
	}

	return sb.String()
}

// BatchToJSON converts multiple KG pipelines to a JSON array.
func (s *KGSerializer) BatchToJSON(pipelines []KGPipeline) ([]byte, error) {
	configs := make([]map[string]any, len(pipelines))
	for i, p := range pipelines {
		configs[i] = s.ToMap(p)
	}
	return json.MarshalIndent(configs, "", "  ")
}
