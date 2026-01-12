package kg

import (
	"encoding/json"
	"testing"
)

func TestSimpleKGPipeline_Interface(t *testing.T) {
	p := &SimpleKGPipeline{
		BasePipeline: BasePipeline{
			Name: "document_kg",
		},
		EntityTypes: []EntityType{
			{Name: "Person", Description: "A human being"},
			{Name: "Organization", Description: "A company or institution"},
		},
		RelationTypes: []RelationType{
			{Name: "WORKS_FOR", Description: "Employment relationship"},
		},
	}

	if p.PipelineName() != "document_kg" {
		t.Errorf("PipelineName() = %v, want document_kg", p.PipelineName())
	}
	if p.PipelineType() != SimpleKG {
		t.Errorf("PipelineType() = %v, want SimpleKG", p.PipelineType())
	}
}

func TestCustomKGPipeline_Interface(t *testing.T) {
	p := &CustomKGPipeline{
		BasePipeline: BasePipeline{
			Name: "custom_kg",
		},
		ExtractionPrompt: "Extract entities from the following text...",
	}

	if p.PipelineType() != CustomKG {
		t.Errorf("PipelineType() = %v, want CustomKG", p.PipelineType())
	}
}

func TestEntityType_Structure(t *testing.T) {
	e := EntityType{
		Name:        "Person",
		Description: "A human being mentioned in the text",
		Properties: []EntityProperty{
			{Name: "name", Type: "STRING", Description: "Full name"},
			{Name: "age", Type: "INTEGER", Description: "Age in years"},
		},
	}

	if e.Name != "Person" {
		t.Errorf("Name = %v, want Person", e.Name)
	}
	if len(e.Properties) != 2 {
		t.Errorf("Properties count = %v, want 2", len(e.Properties))
	}
}

func TestRelationType_Structure(t *testing.T) {
	r := RelationType{
		Name:        "WORKS_FOR",
		Description: "Employment relationship",
		SourceTypes: []string{"Person"},
		TargetTypes: []string{"Organization"},
	}

	if r.Name != "WORKS_FOR" {
		t.Errorf("Name = %v, want WORKS_FOR", r.Name)
	}
	if len(r.SourceTypes) != 1 {
		t.Errorf("SourceTypes count = %v, want 1", len(r.SourceTypes))
	}
}

func TestTextSplitter_FixedSize(t *testing.T) {
	s := &FixedSizeSplitter{
		ChunkSize:    1000,
		ChunkOverlap: 200,
	}

	if s.SplitterType() != "fixed_size" {
		t.Errorf("SplitterType() = %v, want fixed_size", s.SplitterType())
	}
}

func TestTextSplitter_LangChain(t *testing.T) {
	s := &LangChainSplitter{
		SplitterClass: "RecursiveCharacterTextSplitter",
		ChunkSize:     1500,
		ChunkOverlap:  100,
	}

	if s.SplitterType() != "langchain" {
		t.Errorf("SplitterType() = %v, want langchain", s.SplitterType())
	}
}

func TestEntityResolver_ExactMatch(t *testing.T) {
	r := &ExactMatchResolver{
		ResolveProperty: "name",
	}

	if r.ResolverType() != "exact_match" {
		t.Errorf("ResolverType() = %v, want exact_match", r.ResolverType())
	}
}

func TestEntityResolver_FuzzyMatch(t *testing.T) {
	r := &FuzzyMatchResolver{
		ResolveProperty: "name",
		Threshold:       0.85,
	}

	if r.ResolverType() != "fuzzy_match" {
		t.Errorf("ResolverType() = %v, want fuzzy_match", r.ResolverType())
	}
}

func TestEntityResolver_SemanticMatch(t *testing.T) {
	r := &SemanticMatchResolver{
		ResolveProperty: "name",
		Threshold:       0.8,
		Model:           "en_core_web_md",
	}

	if r.ResolverType() != "semantic_match" {
		t.Errorf("ResolverType() = %v, want semantic_match", r.ResolverType())
	}
}

func TestLLMConfig_Structure(t *testing.T) {
	c := LLMConfig{
		Provider:    "openai",
		Model:       "gpt-4",
		APIKey:      "sk-xxx",
		Temperature: 0.0,
		MaxTokens:   4096,
	}

	if c.Provider != "openai" {
		t.Error("Provider should be openai")
	}
	if c.Model != "gpt-4" {
		t.Error("Model should be gpt-4")
	}
}

func TestEmbedderConfig_Structure(t *testing.T) {
	c := EmbedderConfig{
		Provider:   "openai",
		Model:      "text-embedding-3-small",
		APIKey:     "sk-xxx",
		Dimensions: 1536,
	}

	if c.Provider != "openai" {
		t.Error("Provider should be openai")
	}
}

func TestNewKGSerializer(t *testing.T) {
	s := NewKGSerializer()
	if s == nil {
		t.Fatal("NewKGSerializer returned nil")
	}
}

func TestKGSerializer_ToJSON_SimpleKGPipeline(t *testing.T) {
	s := NewKGSerializer()
	p := &SimpleKGPipeline{
		BasePipeline: BasePipeline{
			Name: "document_kg",
			LLMConfig: &LLMConfig{
				Provider: "openai",
				Model:    "gpt-4",
			},
		},
		EntityTypes: []EntityType{
			{Name: "Person", Description: "A human being"},
			{Name: "Organization", Description: "A company"},
		},
		RelationTypes: []RelationType{
			{Name: "WORKS_FOR", Description: "Employment"},
		},
	}

	result, err := s.ToJSON(p)
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}

	if parsed["name"] != "document_kg" {
		t.Errorf("name = %v, want document_kg", parsed["name"])
	}
	if parsed["pipelineType"] != "SimpleKG" {
		t.Errorf("pipelineType = %v, want SimpleKG", parsed["pipelineType"])
	}

	entityTypes, ok := parsed["entityTypes"].([]any)
	if !ok || len(entityTypes) != 2 {
		t.Error("expected 2 entity types")
	}

	relationTypes, ok := parsed["relationTypes"].([]any)
	if !ok || len(relationTypes) != 1 {
		t.Error("expected 1 relation type")
	}

	llmConfig, ok := parsed["llmConfig"].(map[string]any)
	if !ok {
		t.Error("expected llmConfig")
	}
	if llmConfig["provider"] != "openai" {
		t.Errorf("llmConfig.provider = %v, want openai", llmConfig["provider"])
	}
}

func TestKGSerializer_ToJSON_WithTextSplitter(t *testing.T) {
	s := NewKGSerializer()
	p := &SimpleKGPipeline{
		BasePipeline: BasePipeline{
			Name: "chunked_kg",
		},
		TextSplitter: &FixedSizeSplitter{
			ChunkSize:    1000,
			ChunkOverlap: 200,
		},
		EntityTypes: []EntityType{
			{Name: "Entity"},
		},
	}

	result, err := s.ToJSON(p)
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}

	splitter, ok := parsed["textSplitter"].(map[string]any)
	if !ok {
		t.Error("expected textSplitter")
	}
	if splitter["type"] != "fixed_size" {
		t.Errorf("textSplitter.type = %v, want fixed_size", splitter["type"])
	}
	if splitter["chunkSize"] != float64(1000) {
		t.Errorf("textSplitter.chunkSize = %v, want 1000", splitter["chunkSize"])
	}
}

func TestKGSerializer_ToJSON_WithEntityResolver(t *testing.T) {
	s := NewKGSerializer()
	p := &SimpleKGPipeline{
		BasePipeline: BasePipeline{
			Name: "resolved_kg",
		},
		EntityResolver: &FuzzyMatchResolver{
			ResolveProperty: "name",
			Threshold:       0.85,
		},
		EntityTypes: []EntityType{
			{Name: "Entity"},
		},
	}

	result, err := s.ToJSON(p)
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}

	resolver, ok := parsed["entityResolver"].(map[string]any)
	if !ok {
		t.Error("expected entityResolver")
	}
	if resolver["type"] != "fuzzy_match" {
		t.Errorf("entityResolver.type = %v, want fuzzy_match", resolver["type"])
	}
	if resolver["threshold"] != 0.85 {
		t.Errorf("entityResolver.threshold = %v, want 0.85", resolver["threshold"])
	}
}

func TestKGSerializer_ToJSON_CustomKGPipeline(t *testing.T) {
	s := NewKGSerializer()
	p := &CustomKGPipeline{
		BasePipeline: BasePipeline{
			Name: "custom_extraction",
		},
		ExtractionPrompt: "Extract entities and relationships from: {text}",
		SchemaPrompt:     "The graph has Person and Organization nodes",
	}

	result, err := s.ToJSON(p)
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}

	if parsed["pipelineType"] != "CustomKG" {
		t.Errorf("pipelineType = %v, want CustomKG", parsed["pipelineType"])
	}
	if parsed["extractionPrompt"] == nil {
		t.Error("expected extractionPrompt")
	}
}

func TestKGSerializer_ToMap(t *testing.T) {
	s := NewKGSerializer()
	p := &SimpleKGPipeline{
		BasePipeline: BasePipeline{
			Name: "test",
		},
		EntityTypes: []EntityType{
			{Name: "Test"},
		},
	}

	result := s.ToMap(p)

	if result["name"] != "test" {
		t.Errorf("name = %v, want test", result["name"])
	}
}

func TestKGSerializer_GenerateSchema(t *testing.T) {
	s := NewKGSerializer()
	p := &SimpleKGPipeline{
		BasePipeline: BasePipeline{
			Name: "schema_test",
		},
		EntityTypes: []EntityType{
			{
				Name:        "Person",
				Description: "A human being",
				Properties: []EntityProperty{
					{Name: "name", Type: "STRING"},
				},
			},
		},
		RelationTypes: []RelationType{
			{
				Name:        "KNOWS",
				Description: "Social connection",
				SourceTypes: []string{"Person"},
				TargetTypes: []string{"Person"},
			},
		},
	}

	schema := s.GenerateSchema(p)

	if schema == "" {
		t.Error("schema should not be empty")
	}
	if len(schema) < 50 {
		t.Error("schema should contain entity and relationship descriptions")
	}
}

func TestPipelineType_Values(t *testing.T) {
	tests := []struct {
		p    KGPipeline
		want PipelineType
	}{
		{&SimpleKGPipeline{}, SimpleKG},
		{&CustomKGPipeline{}, CustomKG},
	}

	for _, tt := range tests {
		t.Run(string(tt.want), func(t *testing.T) {
			if tt.p.PipelineType() != tt.want {
				t.Errorf("PipelineType() = %v, want %v", tt.p.PipelineType(), tt.want)
			}
		})
	}
}

func TestKGPipeline_ImplementsInterface(t *testing.T) {
	// Verify all pipelines implement the KGPipeline interface
	var _ KGPipeline = &SimpleKGPipeline{}
	var _ KGPipeline = &CustomKGPipeline{}
}

func TestTextSplitter_ImplementsInterface(t *testing.T) {
	// Verify all splitters implement the TextSplitter interface
	var _ TextSplitter = &FixedSizeSplitter{}
	var _ TextSplitter = &LangChainSplitter{}
}

func TestEntityResolver_ImplementsInterface(t *testing.T) {
	// Verify all resolvers implement the EntityResolver interface
	var _ EntityResolver = &ExactMatchResolver{}
	var _ EntityResolver = &FuzzyMatchResolver{}
	var _ EntityResolver = &SemanticMatchResolver{}
}
