// Package importer provides functionality to import existing Neo4j configurations
// and generate wetwire Go code.
package importer

import (
	"context"
	"fmt"
	"strings"
)

// ImportResult represents the result of an import operation.
type ImportResult struct {
	// NodeTypes contains the discovered node type definitions.
	NodeTypes []NodeTypeDefinition
	// RelationshipTypes contains the discovered relationship type definitions.
	RelationshipTypes []RelationshipTypeDefinition
	// Constraints contains the raw constraint definitions.
	Constraints []ConstraintDefinition
	// Indexes contains the raw index definitions.
	Indexes []IndexDefinition
}

// NodeTypeDefinition represents a discovered node type.
type NodeTypeDefinition struct {
	Label       string
	Properties  []PropertyDefinition
	Constraints []ConstraintDefinition
	Indexes     []IndexDefinition
}

// RelationshipTypeDefinition represents a discovered relationship type.
type RelationshipTypeDefinition struct {
	Type        string
	Source      string // Source node label
	Target      string // Target node label
	Properties  []PropertyDefinition
	Constraints []ConstraintDefinition
	Indexes     []IndexDefinition
}

// PropertyDefinition represents a property on a node or relationship.
type PropertyDefinition struct {
	Name     string
	Type     string // Neo4j type (STRING, INTEGER, etc.)
	Required bool
}

// ConstraintDefinition represents a Neo4j constraint.
type ConstraintDefinition struct {
	Name       string
	Type       string // UNIQUENESS, NODE_KEY, EXISTENCE, etc.
	EntityType string // NODE or RELATIONSHIP
	Label      string // Node label or relationship type
	Properties []string
}

// IndexDefinition represents a Neo4j index.
type IndexDefinition struct {
	Name       string
	Type       string // RANGE, FULLTEXT, VECTOR, etc.
	EntityType string // NODE or RELATIONSHIP
	Label      string
	Properties []string
	Options    map[string]any
}

// Importer defines the interface for importing Neo4j configurations.
type Importer interface {
	// Import imports configurations from the source.
	Import(ctx context.Context) (*ImportResult, error)
}

// Generator generates Go code from import results.
type Generator struct {
	PackageName string
}

// NewGenerator creates a new code generator.
func NewGenerator(packageName string) *Generator {
	return &Generator{PackageName: packageName}
}

// Generate generates Go code from the import result.
func (g *Generator) Generate(result *ImportResult) (string, error) {
	var sb strings.Builder

	// Package declaration
	sb.WriteString(fmt.Sprintf("package %s\n\n", g.PackageName))

	// Imports
	sb.WriteString("import (\n")
	sb.WriteString("\t\"github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema\"\n")
	sb.WriteString(")\n\n")

	// Generate node types
	for _, node := range result.NodeTypes {
		code := g.generateNodeType(node)
		sb.WriteString(code)
		sb.WriteString("\n")
	}

	// Generate relationship types
	for _, rel := range result.RelationshipTypes {
		code := g.generateRelationshipType(rel)
		sb.WriteString(code)
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

func (g *Generator) generateNodeType(node NodeTypeDefinition) string {
	var sb strings.Builder

	// Variable declaration
	varName := toCamelCase(node.Label)
	sb.WriteString(fmt.Sprintf("// %s represents the %s node type.\n", varName, node.Label))
	sb.WriteString(fmt.Sprintf("var %s = &schema.NodeType{\n", varName))
	sb.WriteString(fmt.Sprintf("\tLabel: %q,\n", node.Label))

	// Properties
	if len(node.Properties) > 0 {
		sb.WriteString("\tProperties: []schema.Property{\n")
		for _, prop := range node.Properties {
			sb.WriteString(fmt.Sprintf("\t\t{Name: %q, Type: schema.%s", prop.Name, mapNeo4jType(prop.Type)))
			if prop.Required {
				sb.WriteString(", Required: true")
			}
			sb.WriteString("},\n")
		}
		sb.WriteString("\t},\n")
	}

	// Constraints
	for _, c := range node.Constraints {
		switch c.Type {
		case "UNIQUENESS":
			sb.WriteString("\tConstraints: []schema.Constraint{\n")
			sb.WriteString(fmt.Sprintf("\t\t{Type: schema.Unique, Properties: %s},\n", formatStringSlice(c.Properties)))
			sb.WriteString("\t},\n")
		case "NODE_KEY":
			sb.WriteString("\tConstraints: []schema.Constraint{\n")
			sb.WriteString(fmt.Sprintf("\t\t{Type: schema.NodeKey, Properties: %s},\n", formatStringSlice(c.Properties)))
			sb.WriteString("\t},\n")
		}
	}

	// Indexes
	for _, idx := range node.Indexes {
		sb.WriteString("\tIndexes: []schema.Index{\n")
		sb.WriteString(fmt.Sprintf("\t\t{Type: schema.%s, Properties: %s},\n", mapIndexType(idx.Type), formatStringSlice(idx.Properties)))
		sb.WriteString("\t},\n")
	}

	sb.WriteString("}\n")
	return sb.String()
}

func (g *Generator) generateRelationshipType(rel RelationshipTypeDefinition) string {
	var sb strings.Builder

	// Variable declaration
	varName := toCamelCase(rel.Type)
	sb.WriteString(fmt.Sprintf("// %s represents the %s relationship type.\n", varName, rel.Type))
	sb.WriteString(fmt.Sprintf("var %s = &schema.RelationshipType{\n", varName))
	sb.WriteString(fmt.Sprintf("\tLabel: %q,\n", rel.Type))

	if rel.Source != "" {
		sb.WriteString(fmt.Sprintf("\tSource: %q,\n", rel.Source))
	}
	if rel.Target != "" {
		sb.WriteString(fmt.Sprintf("\tTarget: %q,\n", rel.Target))
	}

	// Properties
	if len(rel.Properties) > 0 {
		sb.WriteString("\tProperties: []schema.Property{\n")
		for _, prop := range rel.Properties {
			sb.WriteString(fmt.Sprintf("\t\t{Name: %q, Type: schema.%s", prop.Name, mapNeo4jType(prop.Type)))
			if prop.Required {
				sb.WriteString(", Required: true")
			}
			sb.WriteString("},\n")
		}
		sb.WriteString("\t},\n")
	}

	sb.WriteString("}\n")
	return sb.String()
}

// Helper functions

func toCamelCase(s string) string {
	// Convert SCREAMING_SNAKE_CASE or PascalCase to camelCase
	s = strings.ReplaceAll(s, "_", " ")
	words := strings.Fields(s)
	if len(words) == 0 {
		return s
	}

	result := strings.ToLower(words[0])
	for i := 1; i < len(words); i++ {
		if len(words[i]) > 0 {
			result += strings.ToUpper(words[i][:1]) + strings.ToLower(words[i][1:])
		}
	}
	return result
}

func mapNeo4jType(t string) string {
	switch strings.ToUpper(t) {
	case "STRING":
		return "TypeString"
	case "INTEGER", "INT", "LONG":
		return "TypeInteger"
	case "FLOAT", "DOUBLE":
		return "TypeFloat"
	case "BOOLEAN", "BOOL":
		return "TypeBoolean"
	case "DATE":
		return "TypeDate"
	case "DATETIME", "ZONED DATETIME":
		return "TypeDateTime"
	case "POINT":
		return "TypePoint"
	case "LIST":
		return "TypeListString" // Default to string list
	default:
		return "TypeString"
	}
}

func mapIndexType(t string) string {
	switch strings.ToUpper(t) {
	case "RANGE":
		return "RangeIndex"
	case "FULLTEXT":
		return "FullTextIndex"
	case "VECTOR":
		return "VectorIndex"
	case "TEXT":
		return "TextIndex"
	default:
		return "RangeIndex"
	}
}

func formatStringSlice(ss []string) string {
	if len(ss) == 0 {
		return "[]string{}"
	}
	quoted := make([]string, len(ss))
	for i, s := range ss {
		quoted[i] = fmt.Sprintf("%q", s)
	}
	return "[]string{" + strings.Join(quoted, ", ") + "}"
}
