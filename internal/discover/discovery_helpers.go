// Package discover provides AST-based resource discovery for Neo4j definitions.
//
// This file contains helper functions and utility types extracted from discovery.go
// for better code organization and maintainability.
package discover

import (
	"fmt"
	"go/ast"
	"go/token"
	"sort"
	"strings"

	coreast "github.com/lex00/wetwire-core-go/ast"
	"github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"
)

// extractDependencies extracts dependencies from struct field types.
func (s *Scanner) extractDependencies(structType *ast.StructType) []string {
	var deps []string
	seen := make(map[string]bool)

	if structType.Fields == nil {
		return deps
	}

	for _, field := range structType.Fields.List {
		// Use coreast.ExtractTypeName which unwraps pointers, slices, and selectors
		baseName, _ := coreast.ExtractTypeName(field.Type)

		// Skip builtin types (use coreast.IsBuiltinType for Go builtins)
		if coreast.IsBuiltinType(baseName) {
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
	return s.extractStringField(lit, "Label")
}

// extractStringField extracts a string field value from a composite literal.
func (s *Scanner) extractStringField(lit *ast.CompositeLit, fieldName string) string {
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}

		keyIdent, ok := kv.Key.(*ast.Ident)
		if !ok || keyIdent.Name != fieldName {
			continue
		}

		basicLit, ok := kv.Value.(*ast.BasicLit)
		if !ok || basicLit.Kind != token.STRING {
			continue
		}

		value := basicLit.Value
		if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
			return value[1 : len(value)-1]
		}
	}
	return ""
}

// extractProperties extracts property definitions from a composite literal.
func (s *Scanner) extractProperties(lit *ast.CompositeLit) []PropertyInfo {
	var props []PropertyInfo

	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}

		keyIdent, ok := kv.Key.(*ast.Ident)
		if !ok || keyIdent.Name != "Properties" {
			continue
		}

		// Properties should be a composite literal (slice)
		propsLit, ok := kv.Value.(*ast.CompositeLit)
		if !ok {
			continue
		}

		// Each element is a Property struct
		for _, propElt := range propsLit.Elts {
			propLit, ok := propElt.(*ast.CompositeLit)
			if !ok {
				continue
			}

			prop := PropertyInfo{}
			for _, propField := range propLit.Elts {
				propKV, ok := propField.(*ast.KeyValueExpr)
				if !ok {
					continue
				}

				propKey, ok := propKV.Key.(*ast.Ident)
				if !ok {
					continue
				}

				switch propKey.Name {
				case "Name":
					if bl, ok := propKV.Value.(*ast.BasicLit); ok && bl.Kind == token.STRING {
						prop.Name = strings.Trim(bl.Value, `"`)
					}
				case "Type":
					// Type can be schema.STRING or just an identifier
					prop.Type = s.extractTypeConstant(propKV.Value)
				case "Required":
					if ident, ok := propKV.Value.(*ast.Ident); ok {
						prop.Required = ident.Name == "true"
					}
				}
			}

			if prop.Name != "" {
				props = append(props, prop)
			}
		}
	}

	return props
}

// extractTypeConstant extracts a type constant value (e.g., schema.STRING -> "STRING").
func (s *Scanner) extractTypeConstant(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		// schema.STRING -> STRING
		return e.Sel.Name
	}
	return ""
}

// extractConstraints extracts constraint definitions from a composite literal.
func (s *Scanner) extractConstraints(lit *ast.CompositeLit) []ConstraintInfo {
	var constraints []ConstraintInfo

	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}

		keyIdent, ok := kv.Key.(*ast.Ident)
		if !ok || keyIdent.Name != "Constraints" {
			continue
		}

		constraintsLit, ok := kv.Value.(*ast.CompositeLit)
		if !ok {
			continue
		}

		for _, cElt := range constraintsLit.Elts {
			cLit, ok := cElt.(*ast.CompositeLit)
			if !ok {
				continue
			}

			c := ConstraintInfo{}
			for _, cField := range cLit.Elts {
				cKV, ok := cField.(*ast.KeyValueExpr)
				if !ok {
					continue
				}

				cKey, ok := cKV.Key.(*ast.Ident)
				if !ok {
					continue
				}

				switch cKey.Name {
				case "Type":
					c.Type = s.extractTypeConstant(cKV.Value)
				case "Properties":
					c.Properties = s.extractStringSlice(cKV.Value)
				}
			}

			if c.Type != "" {
				constraints = append(constraints, c)
			}
		}
	}

	return constraints
}

// extractIndexes extracts index definitions from a composite literal.
func (s *Scanner) extractIndexes(lit *ast.CompositeLit) []IndexInfo {
	var indexes []IndexInfo

	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}

		keyIdent, ok := kv.Key.(*ast.Ident)
		if !ok || keyIdent.Name != "Indexes" {
			continue
		}

		indexesLit, ok := kv.Value.(*ast.CompositeLit)
		if !ok {
			continue
		}

		for _, iElt := range indexesLit.Elts {
			iLit, ok := iElt.(*ast.CompositeLit)
			if !ok {
				continue
			}

			idx := IndexInfo{}
			for _, iField := range iLit.Elts {
				iKV, ok := iField.(*ast.KeyValueExpr)
				if !ok {
					continue
				}

				iKey, ok := iKV.Key.(*ast.Ident)
				if !ok {
					continue
				}

				switch iKey.Name {
				case "Type":
					idx.Type = s.extractTypeConstant(iKV.Value)
				case "Properties":
					idx.Properties = s.extractStringSlice(iKV.Value)
				}
			}

			if idx.Type != "" {
				indexes = append(indexes, idx)
			}
		}
	}

	return indexes
}

// extractStringSlice extracts a []string value from an expression.
func (s *Scanner) extractStringSlice(expr ast.Expr) []string {
	var result []string

	lit, ok := expr.(*ast.CompositeLit)
	if !ok {
		return result
	}

	for _, elt := range lit.Elts {
		if bl, ok := elt.(*ast.BasicLit); ok && bl.Kind == token.STRING {
			result = append(result, strings.Trim(bl.Value, `"`))
		}
	}

	return result
}

// extractSourceTarget extracts Source and Target fields from a relationship composite literal.
func (s *Scanner) extractSourceTarget(lit *ast.CompositeLit) (source, target string) {
	source = s.extractStringField(lit, "Source")
	target = s.extractStringField(lit, "Target")
	return
}

// extractAgentContext extracts the AgentContext field from a Schema composite literal.
// Handles both regular strings ("...") and raw strings (`...`).
func (s *Scanner) extractAgentContext(lit *ast.CompositeLit) string {
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}

		keyIdent, ok := kv.Key.(*ast.Ident)
		if !ok || keyIdent.Name != "AgentContext" {
			continue
		}

		basicLit, ok := kv.Value.(*ast.BasicLit)
		if !ok || basicLit.Kind != token.STRING {
			continue
		}

		value := basicLit.Value
		// Handle regular strings "..."
		if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
			return value[1 : len(value)-1]
		}
		// Handle raw strings `...`
		if len(value) >= 2 && value[0] == '`' && value[len(value)-1] == '`' {
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
		// Use coreast.IsBuiltinIdent to skip Go builtins (types, funcs, consts)
		if !coreast.IsBuiltinIdent(name) && isValidIdentifier(name) && !seen[name] {
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

// LintableNodeType is a wrapper that converts DiscoveredResource to schema.NodeType for linting.
type LintableNodeType struct {
	Label      string
	Properties []PropertyInfo
}

// ToSchemaNodeType converts to a schema.NodeType suitable for linting.
func (l *LintableNodeType) ToSchemaNodeType() *schema.NodeType {
	// Import schema package - we need to add this import
	props := make([]schema.Property, len(l.Properties))
	for i, p := range l.Properties {
		props[i] = schema.Property{
			Name:     p.Name,
			Type:     stringToPropertyType(p.Type),
			Required: p.Required,
		}
	}
	return &schema.NodeType{
		Label:      l.Label,
		Properties: props,
	}
}

// LintableRelationshipType is a wrapper that converts DiscoveredResource to schema.RelationshipType for linting.
type LintableRelationshipType struct {
	Label      string
	Source     string
	Target     string
	Properties []PropertyInfo
}

// ToSchemaRelationshipType converts to a schema.RelationshipType suitable for linting.
func (l *LintableRelationshipType) ToSchemaRelationshipType() *schema.RelationshipType {
	props := make([]schema.Property, len(l.Properties))
	for i, p := range l.Properties {
		props[i] = schema.Property{
			Name:     p.Name,
			Type:     stringToPropertyType(p.Type),
			Required: p.Required,
		}
	}
	return &schema.RelationshipType{
		Label:      l.Label,
		Source:     l.Source,
		Target:     l.Target,
		Properties: props,
	}
}

// stringToPropertyType converts a string type name to a PropertyType constant.
func stringToPropertyType(s string) schema.PropertyType {
	switch s {
	case "StringType", "STRING":
		return schema.STRING
	case "IntType", "INTEGER", "INT":
		return schema.INTEGER
	case "FloatType", "FLOAT":
		return schema.FLOAT
	case "BoolType", "BOOLEAN", "BOOL":
		return schema.BOOLEAN
	case "DateType", "DATE":
		return schema.DATE
	case "DateTimeType", "DATETIME":
		return schema.DATETIME
	case "PointType", "POINT":
		return schema.POINT
	case "ListStringType", "LIST_STRING":
		return schema.LIST_STRING
	case "ListIntType", "LIST_INT", "LIST_INTEGER":
		return schema.LIST_INTEGER
	case "ListFloatType", "LIST_FLOAT":
		return schema.LIST_FLOAT
	default:
		return schema.STRING // Default to string
	}
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
