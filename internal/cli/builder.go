// Package cli provides CLI implementations for the wetwire-neo4j command.
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lex00/wetwire-neo4j-go/internal/algorithms"
	"github.com/lex00/wetwire-neo4j-go/internal/discovery"
	"github.com/lex00/wetwire-neo4j-go/internal/kg"
	"github.com/lex00/wetwire-neo4j-go/internal/pipelines"
	"github.com/lex00/wetwire-neo4j-go/internal/projections"
	"github.com/lex00/wetwire-neo4j-go/internal/retrievers"
	"github.com/lex00/wetwire-neo4j-go/internal/serializer"
	"github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"
)

// Builder implements the Builder interface for Neo4j definitions.
type Builder struct {
	scanner          *discovery.Scanner
	cypherSerializer *serializer.CypherSerializer
	jsonSerializer   *serializer.JSONSerializer
	algoSerializer   *algorithms.AlgorithmSerializer
	pipeSerializer   *pipelines.PipelineSerializer
	projSerializer   *projections.ProjectionSerializer
	retSerializer    *retrievers.RetrieverSerializer
	kgSerializer     *kg.KGSerializer
}

// NewBuilder creates a new Builder.
func NewBuilder() *Builder {
	return &Builder{
		scanner:          discovery.NewScanner(),
		cypherSerializer: serializer.NewCypherSerializer(),
		jsonSerializer:   serializer.NewJSONSerializer(),
		algoSerializer:   algorithms.NewAlgorithmSerializer(),
		pipeSerializer:   pipelines.NewPipelineSerializer(),
		projSerializer:   projections.NewProjectionSerializer(),
		retSerializer:    retrievers.NewRetrieverSerializer(),
		kgSerializer:     kg.NewKGSerializer(),
	}
}

// Build implements Builder.Build.
func (b *Builder) Build(ctx context.Context, path string, opts BuildOptions) error {
	// Discover resources
	resources, err := b.scanner.ScanDir(path)
	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}

	if len(resources) == 0 {
		if opts.Verbose {
			fmt.Println("No resources found")
		}
		return nil
	}

	// Sort resources by dependency order
	graph := discovery.NewDependencyGraph(resources)
	sortedResources, err := graph.TopologicalSort()
	if err != nil {
		return fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	// Generate output based on format (determined by output file extension)
	format := b.detectFormat(opts.Output)
	output, err := b.generateOutput(sortedResources, format, opts.Verbose)
	if err != nil {
		return fmt.Errorf("failed to generate output: %w", err)
	}

	if opts.DryRun {
		fmt.Println(output)
		return nil
	}

	// Write output
	if opts.Output == "" {
		fmt.Println(output)
	} else {
		if err := os.WriteFile(opts.Output, []byte(output), 0644); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
		if opts.Verbose {
			fmt.Printf("Wrote output to %s\n", opts.Output)
		}
	}

	return nil
}

// detectFormat determines the output format from the filename extension.
func (b *Builder) detectFormat(output string) string {
	if output == "" {
		return "cypher"
	}
	ext := strings.ToLower(filepath.Ext(output))
	switch ext {
	case ".json":
		return "json"
	case ".cypher", ".cql":
		return "cypher"
	default:
		return "cypher"
	}
}

// generateOutput generates the output in the specified format.
func (b *Builder) generateOutput(resources []discovery.DiscoveredResource, format string, verbose bool) (string, error) {
	switch format {
	case "json":
		return b.generateJSON(resources, verbose)
	case "cypher":
		return b.generateCypher(resources, verbose)
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}

// generateCypher generates Cypher output for schema resources.
func (b *Builder) generateCypher(resources []discovery.DiscoveredResource, verbose bool) (string, error) {
	var statements []string

	// Count resources by type for summary
	counts := make(map[discovery.ResourceKind]int)
	for _, r := range resources {
		counts[r.Kind]++
	}

	if verbose {
		fmt.Printf("Found %d resources:\n", len(resources))
		for kind, count := range counts {
			fmt.Printf("  %s: %d\n", kind, count)
		}
	}

	// Generate Cypher for schema resources
	// Note: This generates template Cypher based on discovered type names
	// In a full implementation, we'd need to load and parse the actual definitions
	for _, r := range resources {
		switch r.Kind {
		case discovery.KindNodeType:
			statements = append(statements, fmt.Sprintf("// NodeType: %s (from %s:%d)", r.Name, r.File, r.Line))
		case discovery.KindRelationshipType:
			statements = append(statements, fmt.Sprintf("// RelationshipType: %s (from %s:%d)", r.Name, r.File, r.Line))
		case discovery.KindAlgorithm:
			statements = append(statements, fmt.Sprintf("// Algorithm: %s (from %s:%d)", r.Name, r.File, r.Line))
		case discovery.KindPipeline:
			statements = append(statements, fmt.Sprintf("// Pipeline: %s (from %s:%d)", r.Name, r.File, r.Line))
		case discovery.KindRetriever:
			statements = append(statements, fmt.Sprintf("// Retriever: %s (from %s:%d)", r.Name, r.File, r.Line))
		}
	}

	return strings.Join(statements, "\n"), nil
}

// generateJSON generates JSON output for all resources.
func (b *Builder) generateJSON(resources []discovery.DiscoveredResource, verbose bool) (string, error) {
	output := make(map[string][]map[string]any)

	for _, r := range resources {
		entry := map[string]any{
			"name":    r.Name,
			"file":    r.File,
			"line":    r.Line,
			"package": r.Package,
		}
		if len(r.Dependencies) > 0 {
			entry["dependencies"] = r.Dependencies
		}

		key := string(r.Kind)
		output[key] = append(output[key], entry)
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(data), nil
}

// BuildFromResources builds output from actual loaded resources.
// This is used when resources have already been loaded and parsed.
func (b *Builder) BuildFromResources(
	nodeTypes []*schema.NodeType,
	relTypes []*schema.RelationshipType,
	algos []algorithms.Algorithm,
	pipes []pipelines.Pipeline,
	projs []projections.Projection,
	rets []retrievers.Retriever,
	kgPipes []kg.KGPipeline,
	format string,
) (string, error) {
	switch format {
	case "cypher":
		return b.buildCypherFromResources(nodeTypes, relTypes, algos, pipes, projs)
	case "json":
		return b.buildJSONFromResources(nodeTypes, relTypes, algos, pipes, projs, rets, kgPipes)
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}

// buildCypherFromResources generates Cypher from loaded resources.
func (b *Builder) buildCypherFromResources(
	nodeTypes []*schema.NodeType,
	relTypes []*schema.RelationshipType,
	algos []algorithms.Algorithm,
	pipes []pipelines.Pipeline,
	projs []projections.Projection,
) (string, error) {
	var sections []string

	// Schema Cypher
	if len(nodeTypes) > 0 || len(relTypes) > 0 {
		cypher, err := b.cypherSerializer.SerializeAll(nodeTypes, relTypes)
		if err != nil {
			return "", fmt.Errorf("failed to serialize schema: %w", err)
		}
		if cypher != "" {
			sections = append(sections, "// Schema Constraints and Indexes\n"+cypher)
		}
	}

	// Algorithm Cypher
	for _, algo := range algos {
		cypher, err := b.algoSerializer.ToCypher(algo)
		if err != nil {
			return "", fmt.Errorf("failed to serialize algorithm: %w", err)
		}
		sections = append(sections, cypher)
	}

	// Pipeline Cypher
	for _, pipe := range pipes {
		cypher, err := b.pipeSerializer.ToCypher(pipe, "graph", "model")
		if err != nil {
			return "", fmt.Errorf("failed to serialize pipeline: %w", err)
		}
		sections = append(sections, cypher)
	}

	// Projection Cypher
	for _, proj := range projs {
		cypher, err := b.projSerializer.ToCypher(proj)
		if err != nil {
			return "", fmt.Errorf("failed to serialize projection: %w", err)
		}
		sections = append(sections, cypher)
	}

	return strings.Join(sections, "\n\n"), nil
}

// buildJSONFromResources generates JSON from loaded resources.
func (b *Builder) buildJSONFromResources(
	nodeTypes []*schema.NodeType,
	relTypes []*schema.RelationshipType,
	algos []algorithms.Algorithm,
	pipes []pipelines.Pipeline,
	projs []projections.Projection,
	rets []retrievers.Retriever,
	kgPipes []kg.KGPipeline,
) (string, error) {
	output := make(map[string]any)

	// Schema
	if len(nodeTypes) > 0 {
		nodes := make([]map[string]any, len(nodeTypes))
		for i, n := range nodeTypes {
			nodes[i] = b.jsonSerializer.NodeTypeToMap(n)
		}
		output["nodeTypes"] = nodes
	}

	if len(relTypes) > 0 {
		rels := make([]map[string]any, len(relTypes))
		for i, r := range relTypes {
			rels[i] = b.jsonSerializer.RelationshipTypeToMap(r)
		}
		output["relationshipTypes"] = rels
	}

	// Algorithms
	if len(algos) > 0 {
		algoMaps := make([]map[string]any, len(algos))
		for i, a := range algos {
			algoMaps[i] = b.algoSerializer.ToMap(a)
		}
		output["algorithms"] = algoMaps
	}

	// Pipelines
	if len(pipes) > 0 {
		pipeMaps := make([]map[string]any, len(pipes))
		for i, p := range pipes {
			pipeMaps[i] = b.pipeSerializer.ToMap(p)
		}
		output["pipelines"] = pipeMaps
	}

	// Projections
	if len(projs) > 0 {
		projMaps := make([]map[string]any, len(projs))
		for i, p := range projs {
			projMaps[i] = b.projSerializer.ToMap(p)
		}
		output["projections"] = projMaps
	}

	// Retrievers
	if len(rets) > 0 {
		retMaps := make([]map[string]any, len(rets))
		for i, r := range rets {
			retMaps[i] = b.retSerializer.ToMap(r)
		}
		output["retrievers"] = retMaps
	}

	// KG Pipelines
	if len(kgPipes) > 0 {
		kgMaps := make([]map[string]any, len(kgPipes))
		for i, p := range kgPipes {
			kgMaps[i] = b.kgSerializer.ToMap(p)
		}
		output["kgPipelines"] = kgMaps
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(data), nil
}
