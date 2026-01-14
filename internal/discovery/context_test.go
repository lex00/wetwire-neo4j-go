package discovery

import (
	"strings"
	"testing"
)

func TestFormatSchemaContext(t *testing.T) {
	tests := []struct {
		name      string
		resources []DiscoveredResource
		want      []string // substrings that should appear
		wantEmpty bool
	}{
		{
			name:      "empty resources returns empty string",
			resources: []DiscoveredResource{},
			wantEmpty: true,
		},
		{
			name: "single node type",
			resources: []DiscoveredResource{
				{Name: "Person", Kind: KindNodeType, File: "schemas/person.go", Line: 12},
			},
			want: []string{
				"## Existing Schema",
				"### Nodes",
				"Person",
				"schemas/person.go:12",
			},
		},
		{
			name: "single relationship type",
			resources: []DiscoveredResource{
				{Name: "WorksFor", Kind: KindRelationshipType, File: "schemas/rels.go", Line: 25},
			},
			want: []string{
				"## Existing Schema",
				"### Relationships",
				"WorksFor",
				"schemas/rels.go:25",
			},
		},
		{
			name: "mixed resources",
			resources: []DiscoveredResource{
				{Name: "Person", Kind: KindNodeType, File: "schemas/person.go", Line: 12},
				{Name: "Company", Kind: KindNodeType, File: "schemas/company.go", Line: 8},
				{Name: "WorksFor", Kind: KindRelationshipType, File: "schemas/rels.go", Line: 25},
				{Name: "PageRankConfig", Kind: KindAlgorithm, File: "schemas/algos.go", Line: 10},
			},
			want: []string{
				"## Existing Schema",
				"### Nodes",
				"Person",
				"Company",
				"### Relationships",
				"WorksFor",
				"### Algorithms",
				"PageRankConfig",
			},
		},
		{
			name: "algorithms only",
			resources: []DiscoveredResource{
				{Name: "PageRankConfig", Kind: KindAlgorithm, File: "algos.go", Line: 10},
				{Name: "LouvainConfig", Kind: KindAlgorithm, File: "algos.go", Line: 30},
			},
			want: []string{
				"### Algorithms",
				"PageRankConfig",
				"LouvainConfig",
			},
		},
		{
			name: "pipelines",
			resources: []DiscoveredResource{
				{Name: "ClassifierPipeline", Kind: KindPipeline, File: "pipelines.go", Line: 15},
			},
			want: []string{
				"### Pipelines",
				"ClassifierPipeline",
			},
		},
		{
			name: "retrievers",
			resources: []DiscoveredResource{
				{Name: "VectorSearch", Kind: KindRetriever, File: "retrievers.go", Line: 20},
			},
			want: []string{
				"### Retrievers",
				"VectorSearch",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatSchemaContext(tt.resources)

			if tt.wantEmpty {
				if got != "" {
					t.Errorf("FormatSchemaContext() = %q, want empty string", got)
				}
				return
			}

			for _, substr := range tt.want {
				if !strings.Contains(got, substr) {
					t.Errorf("FormatSchemaContext() missing %q\nGot:\n%s", substr, got)
				}
			}
		})
	}
}

func TestFormatSchemaContext_DoNotRecreateMessage(t *testing.T) {
	resources := []DiscoveredResource{
		{Name: "Person", Kind: KindNodeType, File: "person.go", Line: 1},
	}

	got := FormatSchemaContext(resources)

	// Should contain instruction to not recreate
	if !strings.Contains(got, "Do not recreate") {
		t.Errorf("FormatSchemaContext() should contain 'Do not recreate' message\nGot:\n%s", got)
	}
}

func TestFormatSchemaContext_ToolReference(t *testing.T) {
	resources := []DiscoveredResource{
		{Name: "Person", Kind: KindNodeType, File: "person.go", Line: 1},
	}

	got := FormatSchemaContext(resources)

	// Should reference wetwire_list tool for details
	if !strings.Contains(got, "wetwire_list") {
		t.Errorf("FormatSchemaContext() should reference wetwire_list tool\nGot:\n%s", got)
	}
}

func TestFormatSchemaContext_AgentContext(t *testing.T) {
	resources := []DiscoveredResource{
		{
			Name:         "MySchema",
			Kind:         KindSchema,
			File:         "schema.go",
			Line:         1,
			AgentContext: "Multi-tenant database - always filter by tenantId.\nIgnore nodes prefixed with _ (internal).",
		},
		{Name: "Person", Kind: KindNodeType, File: "schema.go", Line: 10},
		{Name: "Company", Kind: KindNodeType, File: "schema.go", Line: 20},
	}

	got := FormatSchemaContext(resources)

	// Should include Agent Instructions section
	if !strings.Contains(got, "### Agent Instructions") {
		t.Errorf("FormatSchemaContext() should include Agent Instructions section\nGot:\n%s", got)
	}

	// Should include the actual instructions
	if !strings.Contains(got, "Multi-tenant database") {
		t.Errorf("FormatSchemaContext() should include AgentContext content\nGot:\n%s", got)
	}

	if !strings.Contains(got, "tenantId") {
		t.Errorf("FormatSchemaContext() should include tenantId instruction\nGot:\n%s", got)
	}

	// Should still include nodes
	if !strings.Contains(got, "Person") {
		t.Errorf("FormatSchemaContext() should still include nodes\nGot:\n%s", got)
	}
}

func TestFormatSchemaContext_NoAgentContext(t *testing.T) {
	resources := []DiscoveredResource{
		{Name: "Person", Kind: KindNodeType, File: "schema.go", Line: 10},
	}

	got := FormatSchemaContext(resources)

	// Should NOT include Agent Instructions section when no AgentContext
	if strings.Contains(got, "Agent Instructions") {
		t.Errorf("FormatSchemaContext() should not include Agent Instructions when empty\nGot:\n%s", got)
	}
}
