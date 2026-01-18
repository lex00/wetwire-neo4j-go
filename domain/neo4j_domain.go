// Package domain provides the Neo4jDomain implementation for wetwire-core-go.
package domain

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	coredomain "github.com/lex00/wetwire-core-go/domain"
	"github.com/lex00/wetwire-neo4j-go/internal/discover"
	"github.com/lex00/wetwire-neo4j-go/internal/lint"
	"github.com/spf13/cobra"
)

// Version is set at build time
var Version = "dev"

// Re-export core types for convenience
type (
	Context      = coredomain.Context
	BuildOpts    = coredomain.BuildOpts
	LintOpts     = coredomain.LintOpts
	InitOpts     = coredomain.InitOpts
	ValidateOpts = coredomain.ValidateOpts
	ListOpts     = coredomain.ListOpts
	GraphOpts    = coredomain.GraphOpts
	Result       = coredomain.Result
	Error        = coredomain.Error
)

var (
	NewResult              = coredomain.NewResult
	NewResultWithData      = coredomain.NewResultWithData
	NewErrorResult         = coredomain.NewErrorResult
	NewErrorResultMultiple = coredomain.NewErrorResultMultiple
)

// Neo4jDomain implements the Domain interface for Neo4j GDS.
type Neo4jDomain struct{}

// Compile-time checks
var (
	_ coredomain.Domain        = (*Neo4jDomain)(nil)
	_ coredomain.ListerDomain  = (*Neo4jDomain)(nil)
	_ coredomain.GrapherDomain = (*Neo4jDomain)(nil)
)

// Name returns "neo4j"
func (d *Neo4jDomain) Name() string {
	return "neo4j"
}

// Version returns the current version
func (d *Neo4jDomain) Version() string {
	return Version
}

// Builder returns the Neo4j builder implementation
func (d *Neo4jDomain) Builder() coredomain.Builder {
	return &neo4jBuilder{}
}

// Linter returns the Neo4j linter implementation
func (d *Neo4jDomain) Linter() coredomain.Linter {
	return &neo4jLinter{}
}

// Initializer returns the Neo4j initializer implementation
func (d *Neo4jDomain) Initializer() coredomain.Initializer {
	return &neo4jInitializer{}
}

// Validator returns the Neo4j validator implementation
func (d *Neo4jDomain) Validator() coredomain.Validator {
	return &neo4jValidator{}
}

// Lister returns the Neo4j lister implementation
func (d *Neo4jDomain) Lister() coredomain.Lister {
	return &neo4jLister{}
}

// Grapher returns the Neo4j grapher implementation
func (d *Neo4jDomain) Grapher() coredomain.Grapher {
	return &neo4jGrapher{}
}

// CreateRootCommand creates the root command using the domain interface.
func CreateRootCommand(d coredomain.Domain) *cobra.Command {
	return coredomain.Run(d)
}

// neo4jBuilder implements domain.Builder
type neo4jBuilder struct{}

func (b *neo4jBuilder) Build(ctx *Context, path string, opts BuildOpts) (*Result, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve path: %w", err)
	}

	// Discover all resources
	scanner := discover.NewScanner()
	resources, err := scanner.ScanDir(absPath)
	if err != nil {
		return nil, fmt.Errorf("discovery failed: %w", err)
	}

	if len(resources) == 0 {
		return NewErrorResult("no resources found", Error{
			Path:    absPath,
			Message: "no Neo4j definitions found",
		}), nil
	}

	// Sort resources by dependency order
	graph := discover.NewDependencyGraph(resources)
	sortedResources, err := graph.TopologicalSort()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	// Determine output format
	format := opts.Format
	if format == "" {
		format = detectFormatFromOutput(opts.Output)
	}

	// Generate output
	var output string
	switch format {
	case "json", "pretty":
		output, err = b.buildJSON(sortedResources, format == "pretty")
	case "cypher":
		output, err = b.buildCypher(sortedResources)
	default:
		output, err = b.buildJSON(sortedResources, true)
	}

	if err != nil {
		return nil, fmt.Errorf("output generation failed: %w", err)
	}

	// Handle output file
	if !opts.DryRun && opts.Output != "" {
		if err := os.WriteFile(opts.Output, []byte(output), 0644); err != nil {
			return nil, fmt.Errorf("write output: %w", err)
		}
		return NewResult(fmt.Sprintf("Wrote %s", opts.Output)), nil
	}

	return NewResultWithData("Build completed", output), nil
}

func (b *neo4jBuilder) buildJSON(resources []discover.DiscoveredResource, pretty bool) (string, error) {
	output := make(map[string]any)

	// Group by resource kind
	nodeTypes := []map[string]any{}
	relTypes := []map[string]any{}
	algorithms := []map[string]any{}
	pipelines := []map[string]any{}
	retrievers := []map[string]any{}

	for _, r := range resources {
		switch r.Kind {
		case discover.KindNodeType:
			nodeTypes = append(nodeTypes, resourceToMap(r))
		case discover.KindRelationshipType:
			relTypes = append(relTypes, resourceToMap(r))
		case discover.KindAlgorithm:
			algorithms = append(algorithms, resourceToMap(r))
		case discover.KindPipeline:
			pipelines = append(pipelines, resourceToMap(r))
		case discover.KindRetriever:
			retrievers = append(retrievers, resourceToMap(r))
		}
	}

	if len(nodeTypes) > 0 {
		output["nodeTypes"] = nodeTypes
	}
	if len(relTypes) > 0 {
		output["relationshipTypes"] = relTypes
	}
	if len(algorithms) > 0 {
		output["algorithms"] = algorithms
	}
	if len(pipelines) > 0 {
		output["pipelines"] = pipelines
	}
	if len(retrievers) > 0 {
		output["retrievers"] = retrievers
	}

	var data []byte
	var err error
	if pretty {
		data, err = json.MarshalIndent(output, "", "  ")
	} else {
		data, err = json.Marshal(output)
	}

	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (b *neo4jBuilder) buildCypher(resources []discover.DiscoveredResource) (string, error) {
	// TODO: Implement Cypher generation
	// For now, return a placeholder
	return "-- Cypher output not yet implemented\n", nil
}

func detectFormatFromOutput(output string) string {
	if output == "" {
		return "json"
	}
	ext := filepath.Ext(output)
	switch ext {
	case ".cypher", ".cql":
		return "cypher"
	case ".json":
		return "json"
	default:
		return "json"
	}
}

func resourceToMap(r discover.DiscoveredResource) map[string]any {
	m := map[string]any{
		"name": r.Name,
		"kind": string(r.Kind),
		"file": r.File,
		"line": r.Line,
	}

	if len(r.Properties) > 0 {
		m["properties"] = r.Properties
	}
	if len(r.Constraints) > 0 {
		m["constraints"] = r.Constraints
	}
	if len(r.Indexes) > 0 {
		m["indexes"] = r.Indexes
	}
	if r.Source != "" {
		m["source"] = r.Source
	}
	if r.Target != "" {
		m["target"] = r.Target
	}
	if r.AgentContext != "" {
		m["agentContext"] = r.AgentContext
	}

	return m
}

// neo4jLinter implements domain.Linter
type neo4jLinter struct{}

func (l *neo4jLinter) Lint(ctx *Context, path string, opts LintOpts) (*Result, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve path: %w", err)
	}

	// Discover all resources
	scanner := discover.NewScanner()
	resources, err := scanner.ScanDir(absPath)
	if err != nil {
		return nil, fmt.Errorf("discovery failed: %w", err)
	}

	// Build lint options from LintOpts
	lintOpts := lint.LintOptions{
		DisabledRules: opts.Disable,
		Fix:           opts.Fix,
	}

	// Run lint on all resources
	linter := lint.NewLinter()
	var allResults []lint.LintResult

	// Convert discovered resources to lintable objects
	for _, r := range resources {
		// For schema types, we need to lint the discovered resources
		// Create synthetic types based on discovered metadata
		switch r.Kind {
		case discover.KindNodeType:
			node := &discover.LintableNodeType{
				Label:      r.Name,
				Properties: r.Properties,
			}
			// Lint using the discovered node type
			nodeResults := linter.LintNodeType(node.ToSchemaNodeType())
			allResults = append(allResults, nodeResults...)
		case discover.KindRelationshipType:
			rel := &discover.LintableRelationshipType{
				Label:      r.Name,
				Source:     r.Source,
				Target:     r.Target,
				Properties: r.Properties,
			}
			// Lint using the discovered relationship type
			relResults := linter.LintRelationshipType(rel.ToSchemaRelationshipType())
			allResults = append(allResults, relResults...)
		}
	}

	// Filter out disabled rules
	if len(lintOpts.DisabledRules) > 0 {
		disabled := make(map[string]bool)
		for _, rule := range lintOpts.DisabledRules {
			disabled[rule] = true
		}

		var filtered []lint.LintResult
		for _, r := range allResults {
			if !disabled[r.Rule] {
				filtered = append(filtered, r)
			}
		}
		allResults = filtered
	}

	if len(allResults) == 0 {
		return NewResult("No lint issues found"), nil
	}

	// Convert to domain errors
	errs := make([]Error, 0, len(allResults))
	for _, r := range allResults {
		errs = append(errs, Error{
			Path:     r.Location,
			Message:  r.Message,
			Severity: string(r.Severity),
			Code:     r.Rule,
		})
	}

	// If Fix mode is enabled, add a note about auto-fixing
	if opts.Fix {
		return NewErrorResultMultiple("lint issues found (auto-fix not yet implemented for these issues)", errs), nil
	}
	return NewErrorResultMultiple("lint issues found", errs), nil
}

// neo4jInitializer implements domain.Initializer
type neo4jInitializer struct{}

func (i *neo4jInitializer) Init(ctx *Context, path string, opts InitOpts) (*Result, error) {
	// Use opts.Path if provided, otherwise fall back to path argument
	targetPath := opts.Path
	if targetPath == "" || targetPath == "." {
		targetPath = path
	}

	// Handle scenario initialization
	if opts.Scenario {
		return i.initScenario(ctx, targetPath, opts)
	}

	// Basic project initialization
	return i.initProject(ctx, targetPath, opts)
}

// initScenario creates a full scenario structure with prompts and expected outputs
func (i *neo4jInitializer) initScenario(ctx *Context, path string, opts InitOpts) (*Result, error) {
	name := opts.Name
	if name == "" {
		name = filepath.Base(path)
	}

	description := opts.Description
	if description == "" {
		description = "Neo4j GDS scenario"
	}

	// Use core's scenario scaffolding
	scenario := coredomain.ScaffoldScenario(name, description, "neo4j")
	created, err := coredomain.WriteScenario(path, scenario)
	if err != nil {
		return nil, fmt.Errorf("write scenario: %w", err)
	}

	// Create neo4j-specific expected directories
	expectedDirs := []string{
		filepath.Join(path, "expected", "schema"),
		filepath.Join(path, "expected", "algorithms"),
		filepath.Join(path, "expected", "pipelines"),
	}
	for _, dir := range expectedDirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("create directory %s: %w", dir, err)
		}
	}

	// Create example schema in expected/schema/
	exampleSchema := `package schema

import "github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"

// Person represents a person node in the graph
var Person = schema.NodeType{
	Label:       "Person",
	Description: "A person in the social network",
	Properties: []schema.Property{
		{
			Name:     "name",
			Type:     schema.STRING,
			Required: true,
		},
		{
			Name:     "email",
			Type:     schema.STRING,
			Unique:   true,
		},
	},
	Constraints: []schema.Constraint{
		{
			Name:       "person_email_unique",
			Type:       schema.UniqueConstraint,
			Properties: []string{"email"},
		},
	},
}

// Knows represents a relationship between two people
var Knows = schema.RelationshipType{
	Label:       "KNOWS",
	Source:      "Person",
	Target:      "Person",
	Cardinality: schema.ManyToMany,
	Properties: []schema.Property{
		{
			Name: "since",
			Type: schema.DATE,
		},
	},
}
`
	schemaPath := filepath.Join(path, "expected", "schema", "schema.go")
	if err := os.WriteFile(schemaPath, []byte(exampleSchema), 0644); err != nil {
		return nil, fmt.Errorf("write example schema: %w", err)
	}
	created = append(created, "expected/schema/schema.go")

	return NewResultWithData(
		fmt.Sprintf("Created scenario %s with %d files", name, len(created)),
		created,
	), nil
}

// initProject creates a basic project with example schema
func (i *neo4jInitializer) initProject(ctx *Context, path string, opts InitOpts) (*Result, error) {
	// Create directory
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("create directory: %w", err)
	}

	// Create example schema file
	exampleContent := `package schema

import "github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"

// Person represents a person node in the graph
var Person = schema.NodeType{
	Label:       "Person",
	Description: "A person in the social network",
	Properties: []schema.Property{
		{
			Name:     "name",
			Type:     schema.STRING,
			Required: true,
		},
		{
			Name:     "email",
			Type:     schema.STRING,
			Unique:   true,
		},
		{
			Name: "age",
			Type: schema.INTEGER,
		},
	},
	Constraints: []schema.Constraint{
		{
			Name:       "person_name_unique",
			Type:       schema.UniqueConstraint,
			Properties: []string{"name"},
		},
	},
	Indexes: []schema.Index{
		{
			Name:       "person_email_index",
			Type:       schema.RangeIndex,
			Properties: []string{"email"},
		},
	},
}

// Knows represents a relationship between two people
var Knows = schema.RelationshipType{
	Label:       "KNOWS",
	Source:      "Person",
	Target:      "Person",
	Cardinality: schema.ManyToMany,
	Properties: []schema.Property{
		{
			Name: "since",
			Type: schema.DATE,
		},
	},
}
`
	examplePath := filepath.Join(path, "schema.go")
	if err := os.WriteFile(examplePath, []byte(exampleContent), 0644); err != nil {
		return nil, fmt.Errorf("write example: %w", err)
	}

	return NewResult(fmt.Sprintf("Created %s with example Neo4j schema", examplePath)), nil
}

// neo4jValidator implements domain.Validator
type neo4jValidator struct{}

func (v *neo4jValidator) Validate(ctx *Context, path string, opts ValidateOpts) (*Result, error) {
	// For now, validation is the same as lint
	linter := &neo4jLinter{}
	return linter.Lint(ctx, path, LintOpts{})
}

// neo4jLister implements domain.Lister
type neo4jLister struct{}

func (l *neo4jLister) List(ctx *Context, path string, opts ListOpts) (*Result, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve path: %w", err)
	}

	// Discover all resources
	scanner := discover.NewScanner()
	resources, err := scanner.ScanDir(absPath)
	if err != nil {
		return nil, fmt.Errorf("discovery failed: %w", err)
	}

	// Filter by type if specified
	if opts.Type != "" {
		filtered := []discover.DiscoveredResource{}
		for _, r := range resources {
			if string(r.Kind) == opts.Type || matchesTypeAlias(string(r.Kind), opts.Type) {
				filtered = append(filtered, r)
			}
		}
		resources = filtered
	}

	// Sort by name for consistent output
	sort.Slice(resources, func(i, j int) bool {
		return resources[i].Name < resources[j].Name
	})

	// Build list
	list := make([]map[string]string, 0, len(resources))
	for _, r := range resources {
		list = append(list, map[string]string{
			"name": r.Name,
			"type": string(r.Kind),
			"file": r.File,
		})
	}

	return NewResultWithData(fmt.Sprintf("Discovered %d resources", len(list)), list), nil
}

func matchesTypeAlias(kind, requestedType string) bool {
	// Handle plural forms
	kindLower := string(kind)
	typeLower := requestedType

	plurals := map[string]string{
		"NodeType":           "nodetypes",
		"RelationshipType":   "relationshiptypes",
		"Algorithm":          "algorithms",
		"Pipeline":           "pipelines",
		"Retriever":          "retrievers",
		"Schema":             "schemas",
	}

	if plural, ok := plurals[kindLower]; ok && plural == typeLower {
		return true
	}

	return false
}

// neo4jGrapher implements domain.Grapher
type neo4jGrapher struct{}

func (g *neo4jGrapher) Graph(ctx *Context, path string, opts GraphOpts) (*Result, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve path: %w", err)
	}

	// Discover all resources
	scanner := discover.NewScanner()
	resources, err := scanner.ScanDir(absPath)
	if err != nil {
		return nil, fmt.Errorf("discovery failed: %w", err)
	}

	// Build dependency graph
	depGraph := discover.NewDependencyGraph(resources)

	// Generate output based on format
	var graphOutput string
	switch opts.Format {
	case "dot", "":
		graphOutput = generateDOT(resources, depGraph)
	case "mermaid":
		graphOutput = generateMermaid(resources, depGraph)
	default:
		return nil, fmt.Errorf("unknown format: %s", opts.Format)
	}

	return NewResultWithData("Graph generated", graphOutput), nil
}

func generateDOT(resources []discover.DiscoveredResource, graph *discover.DependencyGraph) string {
	output := "digraph G {\n"
	output += "  rankdir=LR;\n"
	output += "  node [shape=box];\n\n"

	// Add nodes
	for _, r := range resources {
		shape := "box"
		color := ""

		switch r.Kind {
		case discover.KindNodeType:
			shape = "ellipse"
			color = "lightblue"
		case discover.KindRelationshipType:
			shape = "diamond"
			color = "lightgreen"
		case discover.KindAlgorithm:
			shape = "hexagon"
			color = "lightyellow"
		case discover.KindPipeline:
			shape = "folder"
			color = "lightpink"
		case discover.KindRetriever:
			shape = "component"
			color = "lightgray"
		}

		attrs := fmt.Sprintf("shape=%s", shape)
		if color != "" {
			attrs += fmt.Sprintf(", fillcolor=%s, style=filled", color)
		}

		output += fmt.Sprintf("  \"%s\" [%s];\n", r.Name, attrs)
	}

	output += "\n"

	// Add edges
	for _, r := range resources {
		for _, dep := range r.Dependencies {
			output += fmt.Sprintf("  \"%s\" -> \"%s\";\n", r.Name, dep)
		}
	}

	output += "}\n"
	return output
}

func generateMermaid(resources []discover.DiscoveredResource, graph *discover.DependencyGraph) string {
	output := "graph LR\n"

	// Add nodes
	for _, r := range resources {
		nodeType := ""

		switch r.Kind {
		case discover.KindNodeType:
			nodeType = "(%s)"
		case discover.KindRelationshipType:
			nodeType = "{%s}"
		case discover.KindAlgorithm:
			nodeType = "[[%s]]"
		case discover.KindPipeline:
			nodeType = "[/%s/]"
		case discover.KindRetriever:
			nodeType = "{{%s}}"
		default:
			nodeType = "[%s]"
		}

		output += fmt.Sprintf("  %s"+nodeType+"\n", r.Name, r.Name)
	}

	output += "\n"

	// Add edges
	for _, r := range resources {
		for _, dep := range r.Dependencies {
			output += fmt.Sprintf("  %s --> %s\n", r.Name, dep)
		}
	}

	return output
}
