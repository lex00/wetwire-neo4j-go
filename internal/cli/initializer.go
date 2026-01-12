package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lex00/wetwire-core-go/cmd"
)

// Initializer creates new Neo4j/GDS project scaffolding.
type Initializer struct{}

// NewInitializer creates a new Initializer.
func NewInitializer() *Initializer {
	return &Initializer{}
}

// Init creates a new project with the specified name.
func (i *Initializer) Init(ctx context.Context, path string, opts cmd.InitOptions) error {
	if path == "" {
		return fmt.Errorf("project path is required")
	}

	// Check if directory exists
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		if !opts.Force {
			return fmt.Errorf("directory %q already exists, use --force to overwrite", path)
		}
	}

	// Create directory structure
	dirs := []string{
		"",
		"schema",
		"algorithms",
		"pipelines",
		"retrievers",
		"kg",
	}

	for _, dir := range dirs {
		dirPath := filepath.Join(path, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dirPath, err)
		}
	}

	// Get the project name from the path
	projectName := filepath.Base(path)

	// Generate files based on template
	template := opts.Template
	if template == "" {
		template = "default"
	}

	if err := i.generateMainGo(path, projectName); err != nil {
		return err
	}

	switch template {
	case "default":
		if err := i.generateSchema(path); err != nil {
			return err
		}
	case "gds":
		if err := i.generateSchema(path); err != nil {
			return err
		}
		if err := i.generateAlgorithms(path); err != nil {
			return err
		}
		if err := i.generatePipelines(path); err != nil {
			return err
		}
	case "graphrag":
		if err := i.generateSchema(path); err != nil {
			return err
		}
		if err := i.generateRetrievers(path); err != nil {
			return err
		}
		if err := i.generateKG(path); err != nil {
			return err
		}
	case "full":
		if err := i.generateSchema(path); err != nil {
			return err
		}
		if err := i.generateAlgorithms(path); err != nil {
			return err
		}
		if err := i.generatePipelines(path); err != nil {
			return err
		}
		if err := i.generateRetrievers(path); err != nil {
			return err
		}
		if err := i.generateKG(path); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown template %q, valid templates: default, gds, graphrag, full", template)
	}

	return nil
}

func (i *Initializer) generateMainGo(path, projectName string) error {
	content := fmt.Sprintf(`package main

import (
	"fmt"
	"os"

	"github.com/lex00/wetwire-core-go/cmd"
	"github.com/lex00/wetwire-neo4j-go/internal/cli"
)

func main() {
	root := cmd.NewRootCommand("wetwire-neo4j", "%s - Neo4j/GDS infrastructure")
	root.AddCommand(cmd.NewBuildCommand(cli.NewBuilder()))
	root.AddCommand(cmd.NewLintCommand(cli.NewLinter()))
	root.AddCommand(cmd.NewInitCommand(cli.NewInitializer()))

	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %%v\n", err)
		os.Exit(1)
	}
}
`, projectName)

	return os.WriteFile(filepath.Join(path, "main.go"), []byte(content), 0644)
}

func (i *Initializer) generateSchema(path string) error {
	content := `package schema

import "github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"

// Person represents a person node in the graph.
var Person = &schema.NodeType{
	Label: "Person",
	Properties: []schema.Property{
		{Name: "name", Type: schema.STRING, Required: true},
		{Name: "age", Type: schema.INTEGER},
	},
	Constraints: []schema.Constraint{
		{Type: schema.UNIQUE, Properties: []string{"name"}},
	},
}

// Company represents a company node in the graph.
var Company = &schema.NodeType{
	Label: "Company",
	Properties: []schema.Property{
		{Name: "name", Type: schema.STRING, Required: true},
		{Name: "founded", Type: schema.INTEGER},
	},
}

// WorksFor represents an employment relationship.
var WorksFor = &schema.RelationshipType{
	Label:       "WORKS_FOR",
	Source:      "Person",
	Target:      "Company",
	Cardinality: schema.ManyToOne,
	Properties: []schema.Property{
		{Name: "since", Type: schema.DATE},
	},
}
`
	return os.WriteFile(filepath.Join(path, "schema", "schema.go"), []byte(content), 0644)
}

func (i *Initializer) generateAlgorithms(path string) error {
	content := `package algorithms

import "github.com/lex00/wetwire-neo4j-go/internal/algorithms"

// PageRankConfig configures PageRank for the social graph.
var PageRankConfig = &algorithms.PageRank{
	BaseAlgorithm: algorithms.BaseAlgorithm{
		GraphName: "social-graph",
		Mode:      algorithms.Stream,
	},
	DampingFactor: 0.85,
	MaxIterations: 20,
}

// LouvainConfig configures Louvain community detection.
var LouvainConfig = &algorithms.Louvain{
	BaseAlgorithm: algorithms.BaseAlgorithm{
		GraphName: "social-graph",
		Mode:      algorithms.Mutate,
	},
	MutateProperty: "community",
}
`
	return os.WriteFile(filepath.Join(path, "algorithms", "algorithms.go"), []byte(content), 0644)
}

func (i *Initializer) generatePipelines(path string) error {
	content := `package pipelines

import "github.com/lex00/wetwire-neo4j-go/internal/pipelines"

// NodeClassifier is a pipeline for predicting node labels.
var NodeClassifier = &pipelines.NodeClassificationPipeline{
	BasePipeline: pipelines.BasePipeline{
		Name:      "node-classifier",
		GraphName: "social-graph",
	},
	TargetProperty:   "label",
	TargetNodeLabels: []string{"Person"},
	FeatureSteps: []pipelines.FeatureStep{
		{Type: "degree", Config: map[string]any{"mutateProperty": "degree"}},
		{Type: "pageRank", Config: map[string]any{"mutateProperty": "pr"}},
	},
	Models: []pipelines.Model{
		{Type: "logisticRegression", Config: map[string]any{"penalty": 0.1}},
	},
}
`
	return os.WriteFile(filepath.Join(path, "pipelines", "pipelines.go"), []byte(content), 0644)
}

func (i *Initializer) generateRetrievers(path string) error {
	content := `package retrievers

import "github.com/lex00/wetwire-neo4j-go/internal/retrievers"

// VectorSearch is a vector similarity retriever.
var VectorSearch = &retrievers.VectorRetriever{
	BaseRetriever: retrievers.BaseRetriever{
		Name: "document-search",
	},
	IndexName:    "document-embeddings",
	TopK:         10,
	NodeLabel:    "Document",
	TextProperty: "content",
	Embedder: &retrievers.EmbedderConfig{
		Provider: "openai",
		Model:    "text-embedding-3-small",
	},
}

// HybridSearch combines vector and fulltext search.
var HybridSearch = &retrievers.HybridRetriever{
	BaseRetriever: retrievers.BaseRetriever{
		Name: "hybrid-search",
	},
	VectorIndexName:   "document-embeddings",
	FulltextIndexName: "document-fulltext",
	TopK:              10,
	NodeLabel:         "Document",
	TextProperty:      "content",
}
`
	return os.WriteFile(filepath.Join(path, "retrievers", "retrievers.go"), []byte(content), 0644)
}

func (i *Initializer) generateKG(path string) error {
	content := `package kg

import "github.com/lex00/wetwire-neo4j-go/internal/kg"

// SimpleKG is a simple knowledge graph extraction pipeline.
var SimpleKG = &kg.SimpleKGPipeline{
	BasePipeline: kg.BasePipeline{
		Name: "simple-kg",
	},
	EntityTypes: []kg.EntityType{
		{Name: "Person", Description: "A human being"},
		{Name: "Organization", Description: "A company or institution"},
		{Name: "Location", Description: "A geographical location"},
	},
	RelationTypes: []kg.RelationType{
		{Name: "WORKS_AT", Source: "Person", Target: "Organization"},
		{Name: "LOCATED_IN", Source: "Organization", Target: "Location"},
	},
	TextSplitter: &kg.FixedSizeSplitter{
		ChunkSize:    1000,
		ChunkOverlap: 200,
	},
}
`
	return os.WriteFile(filepath.Join(path, "kg", "kg.go"), []byte(content), 0644)
}
