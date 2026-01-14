// Package discovery provides AST-based resource discovery for Neo4j definitions.
//
// This package scans Go source files to find type definitions that implement
// the schema.Resource interface, including NodeType, RelationshipType, and
// algorithm configurations.
//
// Example usage:
//
//	scanner := discovery.NewScanner()
//	resources, err := scanner.ScanDir("./schemas")
//	for _, r := range resources {
//	    fmt.Printf("%s: %s at %s:%d\n", r.Type, r.Name, r.File, r.Line)
//	}
package discovery

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
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
)

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
}

// Scanner discovers resources in Go source files.
type Scanner struct {
	// fset is the file set for position tracking.
	fset *token.FileSet
	// typeAliases maps embedded types to resource kinds.
	typeAliases map[string]ResourceKind
}

// NewScanner creates a new resource scanner.
func NewScanner() *Scanner {
	return &Scanner{
		fset: token.NewFileSet(),
		typeAliases: map[string]ResourceKind{
			// Schema types
			"NodeType":         KindNodeType,
			"RelationshipType": KindRelationshipType,
			// Algorithm types (to be added)
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
		},
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

				resources = append(resources, DiscoveredResource{
					Name:         resourceName,
					Kind:         kind,
					File:         filename,
					Line:         pos.Line,
					Package:      pkgName,
					Dependencies: deps,
				})
			}
		}
	}

	return resources, nil
}

// ScanDir scans a directory (recursively) for resource definitions.
func (s *Scanner) ScanDir(dir string) ([]DiscoveredResource, error) {
	var resources []DiscoveredResource

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			// Skip hidden directories and vendor (but not "." or "..")
			name := info.Name()
			if (strings.HasPrefix(name, ".") && name != "." && name != "..") || name == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}

		// Only process .go files (skip test files)
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

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

		typeName := s.getTypeName(field.Type)
		// Try direct match first
		if kind, ok := s.typeAliases[typeName]; ok {
			return kind
		}
		// Try without package prefix (e.g., "schema.NodeType" -> "NodeType")
		if idx := strings.LastIndex(typeName, "."); idx >= 0 {
			baseName := typeName[idx+1:]
			if kind, ok := s.typeAliases[baseName]; ok {
				return kind
			}
		}
	}

	return ""
}

// detectCompositeLitKind detects the resource kind from a composite literal.
func (s *Scanner) detectCompositeLitKind(lit *ast.CompositeLit) ResourceKind {
	typeName := s.getTypeName(lit.Type)

	// Direct type match
	if kind, ok := s.typeAliases[typeName]; ok {
		return kind
	}

	// Check for pointer to known type
	if strings.HasPrefix(typeName, "*") {
		baseName := strings.TrimPrefix(typeName, "*")
		if kind, ok := s.typeAliases[baseName]; ok {
			return kind
		}
	}

	// Try without package prefix (e.g., "schema.NodeType" -> "NodeType")
	if idx := strings.LastIndex(typeName, "."); idx >= 0 {
		baseName := typeName[idx+1:]
		if kind, ok := s.typeAliases[baseName]; ok {
			return kind
		}
	}

	return ""
}

// getTypeName extracts the type name from an expression.
func (s *Scanner) getTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		// pkg.Type
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name + "." + t.Sel.Name
		}
	case *ast.StarExpr:
		// *Type
		return "*" + s.getTypeName(t.X)
	case *ast.UnaryExpr:
		// &Type{}
		return s.getTypeName(t.X)
	}
	return ""
}

// extractDependencies extracts dependencies from struct field types.
func (s *Scanner) extractDependencies(structType *ast.StructType) []string {
	var deps []string
	seen := make(map[string]bool)

	if structType.Fields == nil {
		return deps
	}

	for _, field := range structType.Fields.List {
		// Look for fields with types that could be resource references
		typeName := s.getTypeName(field.Type)
		baseName := strings.TrimPrefix(typeName, "*")
		baseName = strings.TrimPrefix(baseName, "[]")

		// Skip primitive types
		if isPrimitiveType(baseName) {
			continue
		}

		// Skip our own type aliases
		if _, isAlias := s.typeAliases[baseName]; isAlias {
			continue
		}

		// Add as potential dependency if it looks like a user-defined type
		if baseName != "" && !seen[baseName] && isValidIdentifier(baseName) {
			deps = append(deps, baseName)
			seen[baseName] = true
		}
	}

	return deps
}

// extractCompositeLitDependencies extracts dependencies from a composite literal.
func (s *Scanner) extractCompositeLitDependencies(lit *ast.CompositeLit) []string {
	var deps []string
	seen := make(map[string]bool)

	for _, elt := range lit.Elts {
		s.walkExprForDeps(elt, &deps, seen)
	}

	return deps
}

// extractLabelField extracts the Label field value from a composite literal.
// Returns empty string if Label field not found.
func (s *Scanner) extractLabelField(lit *ast.CompositeLit) string {
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}

		// Check if key is "Label"
		keyIdent, ok := kv.Key.(*ast.Ident)
		if !ok || keyIdent.Name != "Label" {
			continue
		}

		// Extract string value
		basicLit, ok := kv.Value.(*ast.BasicLit)
		if !ok || basicLit.Kind != token.STRING {
			continue
		}

		// Remove quotes from string literal
		value := basicLit.Value
		if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
			return value[1 : len(value)-1]
		}
	}
	return ""
}

// walkExprForDeps walks an expression tree looking for identifier references.
func (s *Scanner) walkExprForDeps(expr ast.Expr, deps *[]string, seen map[string]bool) {
	if expr == nil {
		return
	}

	switch e := expr.(type) {
	case *ast.Ident:
		// Check if this could be a resource reference
		name := e.Name
		if !isPrimitiveType(name) && isValidIdentifier(name) && !seen[name] {
			// Heuristic: resource names typically start with uppercase
			if len(name) > 0 && name[0] >= 'A' && name[0] <= 'Z' {
				*deps = append(*deps, name)
				seen[name] = true
			}
		}
	case *ast.KeyValueExpr:
		s.walkExprForDeps(e.Value, deps, seen)
	case *ast.CompositeLit:
		for _, elt := range e.Elts {
			s.walkExprForDeps(elt, deps, seen)
		}
	case *ast.UnaryExpr:
		s.walkExprForDeps(e.X, deps, seen)
	case *ast.SelectorExpr:
		s.walkExprForDeps(e.X, deps, seen)
	case *ast.CallExpr:
		s.walkExprForDeps(e.Fun, deps, seen)
		for _, arg := range e.Args {
			s.walkExprForDeps(arg, deps, seen)
		}
	case *ast.SliceExpr:
		s.walkExprForDeps(e.X, deps, seen)
	case *ast.IndexExpr:
		s.walkExprForDeps(e.X, deps, seen)
	}
}

// isPrimitiveType checks if a type name is a Go primitive.
func isPrimitiveType(name string) bool {
	primitives := map[string]bool{
		"bool": true, "string": true,
		"int": true, "int8": true, "int16": true, "int32": true, "int64": true,
		"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true,
		"float32": true, "float64": true,
		"complex64": true, "complex128": true,
		"byte": true, "rune": true,
		"error": true, "any": true,
	}
	return primitives[name]
}

// isValidIdentifier checks if a string is a valid Go identifier.
func isValidIdentifier(name string) bool {
	if name == "" {
		return false
	}
	for i, c := range name {
		isLower := c >= 'a' && c <= 'z'
		isUpper := c >= 'A' && c <= 'Z'
		isUnderscore := c == '_'
		isDigit := c >= '0' && c <= '9'

		if i == 0 {
			if !isLower && !isUpper && !isUnderscore {
				return false
			}
		} else {
			if !isLower && !isUpper && !isDigit && !isUnderscore {
				return false
			}
		}
	}
	return true
}

// DependencyGraph represents dependencies between resources.
type DependencyGraph struct {
	resources map[string]*DiscoveredResource
	edges     map[string][]string // from -> to
}

// NewDependencyGraph creates a dependency graph from discovered resources.
func NewDependencyGraph(resources []DiscoveredResource) *DependencyGraph {
	g := &DependencyGraph{
		resources: make(map[string]*DiscoveredResource),
		edges:     make(map[string][]string),
	}

	for i := range resources {
		r := &resources[i]
		g.resources[r.Name] = r
		g.edges[r.Name] = r.Dependencies
	}

	return g
}

// TopologicalSort returns resources in dependency order (dependencies first).
// Returns an error if there are circular dependencies.
func (g *DependencyGraph) TopologicalSort() ([]DiscoveredResource, error) {
	// Kahn's algorithm with reversed edges
	// edges[A] = [B, C] means A depends on B and C
	// For topological sort, we need: B must come before A, C must come before A
	// So we build reverse edges: B -> [A], C -> [A] (B and C point to A)

	// Build reverse adjacency list (dependency -> dependents)
	reverseEdges := make(map[string][]string)
	for name := range g.resources {
		reverseEdges[name] = nil
	}
	for name, deps := range g.edges {
		for _, dep := range deps {
			if _, exists := g.resources[dep]; exists {
				reverseEdges[dep] = append(reverseEdges[dep], name)
			}
		}
	}

	// Calculate in-degree (number of dependencies each node has)
	inDegree := make(map[string]int)
	for name := range g.resources {
		inDegree[name] = 0
	}
	for name, deps := range g.edges {
		if _, exists := g.resources[name]; exists {
			for _, dep := range deps {
				if _, exists := g.resources[dep]; exists {
					inDegree[name]++
				}
			}
		}
	}

	// Start with nodes that have no dependencies (in-degree 0)
	var queue []string
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	// Sort queue for deterministic output
	sort.Strings(queue)

	var result []DiscoveredResource
	for len(queue) > 0 {
		// Pop from front
		name := queue[0]
		queue = queue[1:]

		if r, exists := g.resources[name]; exists {
			result = append(result, *r)
		}

		// Reduce in-degree for dependents
		for _, dependent := range reverseEdges[name] {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
				sort.Strings(queue) // Keep queue sorted
			}
		}
	}

	// Check for cycles
	if len(result) != len(g.resources) {
		return nil, fmt.Errorf("circular dependency detected")
	}

	return result, nil
}

// HasCycle checks if the dependency graph has any cycles.
func (g *DependencyGraph) HasCycle() bool {
	_, err := g.TopologicalSort()
	return err != nil
}

// GetDependencies returns all dependencies for a resource (recursive).
func (g *DependencyGraph) GetDependencies(name string) []string {
	visited := make(map[string]bool)
	var result []string

	var visit func(n string)
	visit = func(n string) {
		if visited[n] {
			return
		}
		visited[n] = true

		for _, dep := range g.edges[n] {
			if _, exists := g.resources[dep]; exists {
				result = append(result, dep)
				visit(dep)
			}
		}
	}

	visit(name)
	sort.Strings(result)
	return result
}
