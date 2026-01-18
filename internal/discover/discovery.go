// Package discover provides AST-based resource discovery for Neo4j definitions.
//
// This package scans Go source files to find type definitions that implement
// the schema.Resource interface, including NodeType, RelationshipType, and
// algorithm configurations.
//
// Example usage:
//
//	scanner := discover.NewScanner()
//	resources, err := scanner.ScanDir("./schemas")
//	for _, r := range resources {
//	    fmt.Printf("%s: %s at %s:%d\n", r.Type, r.Name, r.File, r.Line)
//	}
package discover

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"

	coreast "github.com/lex00/wetwire-core-go/ast"
	corediscover "github.com/lex00/wetwire-core-go/discover"
)

// ResourceKind represents the type of discovered resource.
type ResourceKind string

const (
	// KindNodeType represents a NodeType definition.
	KindNodeType ResourceKind = "NodeType"
	// KindRelationshipType represents a RelationshipType definition.
	KindRelationshipType ResourceKind = "RelationshipType"
	// KindAlgorithm represents a GDS algorithm configuration.
	KindAlgorithm ResourceKind = "Algorithm"
	// KindPipeline represents a ML pipeline configuration.
	KindPipeline ResourceKind = "Pipeline"
	// KindRetriever represents a GraphRAG retriever configuration.
	KindRetriever ResourceKind = "Retriever"
	// KindSchema represents a Schema definition with AgentContext.
	KindSchema ResourceKind = "Schema"
)

// PropertyInfo describes a property on a node or relationship type.
type PropertyInfo struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required,omitempty"`
}

// ConstraintInfo describes a constraint on a node type.
type ConstraintInfo struct {
	Type       string   `json:"type"`
	Properties []string `json:"properties"`
}

// IndexInfo describes an index on a node type.
type IndexInfo struct {
	Type       string   `json:"type"`
	Properties []string `json:"properties"`
}

// DiscoveredResource represents a resource found in source code.
type DiscoveredResource struct {
	// Name is the resource name (struct name or variable name).
	Name string
	// Kind is the type of resource (NodeType, RelationshipType, etc.).
	Kind ResourceKind
	// File is the source file path.
	File string
	// Line is the line number where the resource is defined.
	Line int
	// Package is the Go package name.
	Package string
	// Dependencies are the names of other resources this one references.
	Dependencies []string

	// Properties contains property definitions for NodeType and RelationshipType.
	Properties []PropertyInfo `json:"properties,omitempty"`
	// Constraints contains constraint definitions for NodeType.
	Constraints []ConstraintInfo `json:"constraints,omitempty"`
	// Indexes contains index definitions for NodeType.
	Indexes []IndexInfo `json:"indexes,omitempty"`
	// Source is the source node label for RelationshipType.
	Source string `json:"source,omitempty"`
	// Target is the target node label for RelationshipType.
	Target string `json:"target,omitempty"`
	// AgentContext contains instructions for AI agents (from Schema.AgentContext).
	AgentContext string `json:"agentContext,omitempty"`
}

// Scanner discovers resources in Go source files.
type Scanner struct {
	// fset is the file set for position tracking.
	fset *token.FileSet
	// typeAliases maps embedded types to resource kinds.
	typeAliases map[string]ResourceKind
}

// neo4jTypeAliases maps type names to their resource kinds.
// This is used by both the Scanner and the TypeMatcher.
var neo4jTypeAliases = map[string]ResourceKind{
	// Schema types
	"Schema":           KindSchema,
	"NodeType":         KindNodeType,
	"RelationshipType": KindRelationshipType,
	// Algorithm types
	"PageRank":         KindAlgorithm,
	"Louvain":          KindAlgorithm,
	"Leiden":           KindAlgorithm,
	"LabelPropagation": KindAlgorithm,
	"WCC":              KindAlgorithm,
	"Betweenness":      KindAlgorithm,
	"Closeness":        KindAlgorithm,
	"Degree":           KindAlgorithm,
	"ArticleRank":      KindAlgorithm,
	"KCore":            KindAlgorithm,
	"TriangleCount":    KindAlgorithm,
	"NodeSimilarity":   KindAlgorithm,
	"KNN":              KindAlgorithm,
	"Dijkstra":         KindAlgorithm,
	"AStar":            KindAlgorithm,
	"BellmanFord":      KindAlgorithm,
	"BFS":              KindAlgorithm,
	"DFS":              KindAlgorithm,
	"FastRP":           KindAlgorithm,
	"GraphSAGE":        KindAlgorithm,
	"Node2Vec":         KindAlgorithm,
	"HashGNN":          KindAlgorithm,
	// Pipeline types
	"NodeClassificationPipeline": KindPipeline,
	"LinkPredictionPipeline":     KindPipeline,
	"NodeRegressionPipeline":     KindPipeline,
	// Retriever types
	"VectorRetriever":        KindRetriever,
	"VectorCypherRetriever":  KindRetriever,
	"HybridRetriever":        KindRetriever,
	"HybridCypherRetriever":  KindRetriever,
	"Text2CypherRetriever":   KindRetriever,
	"WeaviateNeo4jRetriever": KindRetriever,
	"PineconeNeo4jRetriever": KindRetriever,
	"QdrantNeo4jRetriever":   KindRetriever,
}

// Neo4jTypeMatcher returns a corediscover.TypeMatcher for Neo4j resource types.
// This can be used with the core discover infrastructure for basic discovery.
// For rich Neo4j metadata (properties, constraints, etc.), use the Scanner instead.
func Neo4jTypeMatcher() corediscover.TypeMatcher {
	return func(pkgName, typeName string, imports map[string]string) (string, bool) {
		// Check if the type is a known Neo4j resource type
		if kind, ok := neo4jTypeAliases[typeName]; ok {
			// Verify it comes from a Neo4j-related package if package is specified
			if pkgName != "" {
				importPath := imports[pkgName]
				// Accept types from any Neo4j-related package or local definitions
				if !strings.Contains(importPath, "neo4j") &&
					!strings.Contains(importPath, "schema") &&
					!strings.Contains(importPath, "algorithms") &&
					!strings.Contains(importPath, "pipelines") &&
					!strings.Contains(importPath, "retrievers") {
					return "", false
				}
			}
			return string(kind), true
		}
		return "", false
	}
}

// NewScanner creates a new resource scanner.
func NewScanner() *Scanner {
	return &Scanner{
		fset:        token.NewFileSet(),
		typeAliases: neo4jTypeAliases,
	}
}

// ScanFile scans a single Go file for resource definitions.
func (s *Scanner) ScanFile(filename string) ([]DiscoveredResource, error) {
	src, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	f, err := parser.ParseFile(s.fset, filename, src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	var resources []DiscoveredResource
	pkgName := f.Name.Name

	// Scan for struct type declarations
	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			// Check for embedded types that indicate resource kind
			kind := s.detectResourceKind(structType)
			if kind == "" {
				continue
			}

			pos := s.fset.Position(typeSpec.Pos())
			deps := s.extractDependencies(structType)

			resources = append(resources, DiscoveredResource{
				Name:         typeSpec.Name.Name,
				Kind:         kind,
				File:         filename,
				Line:         pos.Line,
				Package:      pkgName,
				Dependencies: deps,
			})
		}
	}

	// Scan for top-level variable declarations with struct literals
	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.VAR {
			continue
		}

		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}

			for i, name := range valueSpec.Names {
				if i >= len(valueSpec.Values) {
					continue
				}

				value := valueSpec.Values[i]

				// Check for composite literals with known types
				var compLit *ast.CompositeLit
				switch v := value.(type) {
				case *ast.CompositeLit:
					compLit = v
				case *ast.UnaryExpr:
					// Handle &NodeType{} case
					if v.Op == token.AND {
						if cl, ok := v.X.(*ast.CompositeLit); ok {
							compLit = cl
						}
					}
				}

				if compLit == nil {
					continue
				}

				kind := s.detectCompositeLitKind(compLit)
				if kind == "" {
					continue
				}

				pos := s.fset.Position(name.Pos())
				deps := s.extractCompositeLitDependencies(compLit)

				// Extract Label field value for the resource name (actual Neo4j label)
				// Falls back to variable name if Label not found
				resourceName := s.extractLabelField(compLit)
				if resourceName == "" {
					resourceName = name.Name
				}

				// Extract additional details based on resource kind
				res := DiscoveredResource{
					Name:         resourceName,
					Kind:         kind,
					File:         filename,
					Line:         pos.Line,
					Package:      pkgName,
					Dependencies: deps,
				}

				// Extract properties, constraints, indexes for NodeType and RelationshipType
				if kind == KindNodeType || kind == KindRelationshipType {
					res.Properties = s.extractProperties(compLit)
				}
				if kind == KindNodeType {
					res.Constraints = s.extractConstraints(compLit)
					res.Indexes = s.extractIndexes(compLit)
				}
				if kind == KindRelationshipType {
					res.Source, res.Target = s.extractSourceTarget(compLit)
				}
				// Extract AgentContext for Schema
				if kind == KindSchema {
					res.AgentContext = s.extractAgentContext(compLit)
				}

				resources = append(resources, res)
			}
		}
	}

	return resources, nil
}

// ScanDir scans a directory (recursively) for resource definitions.
// Uses corediscover.WalkDir for directory traversal with standard skip patterns.
func (s *Scanner) ScanDir(dir string) ([]DiscoveredResource, error) {
	var resources []DiscoveredResource

	walkOpts := corediscover.WalkOptions{
		SkipTests:    true,
		SkipVendor:   true,
		SkipHidden:   true,
		SkipTestdata: true,
	}

	err := corediscover.WalkDir(dir, walkOpts, func(path string) error {
		fileResources, err := s.ScanFile(path)
		if err != nil {
			// Print parse errors to stderr so users can debug
			fmt.Fprintf(os.Stderr, "warning: failed to parse %s: %v\n", path, err)
			return nil
		}

		resources = append(resources, fileResources...)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return resources, nil
}

// detectResourceKind checks if a struct type embeds a known resource type.
func (s *Scanner) detectResourceKind(structType *ast.StructType) ResourceKind {
	if structType.Fields == nil {
		return ""
	}

	for _, field := range structType.Fields.List {
		// Check for embedded fields (no name)
		if len(field.Names) > 0 {
			continue
		}

		// Use coreast.ExtractTypeName which returns (typeName, pkgName).
		// It unwraps pointers and selectors, returning just the base type name.
		typeName, _ := coreast.ExtractTypeName(field.Type)
		if kind, ok := s.typeAliases[typeName]; ok {
			return kind
		}
	}

	return ""
}

// detectCompositeLitKind detects the resource kind from a composite literal.
func (s *Scanner) detectCompositeLitKind(lit *ast.CompositeLit) ResourceKind {
	// Use coreast.ExtractTypeName which returns (typeName, pkgName).
	// It unwraps pointers and selectors, returning just the base type name.
	typeName, _ := coreast.ExtractTypeName(lit.Type)
	if kind, ok := s.typeAliases[typeName]; ok {
		return kind
	}
	return ""
}
