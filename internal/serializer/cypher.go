// Package serializer provides serialization of Neo4j schema definitions to various output formats.
//
// Supported formats:
// - Cypher: Constraint and index creation statements
// - JSON: GDS parameters and GraphRAG configurations
// - YAML: Human-readable configuration files
//
// Example usage:
//
//	s := serializer.NewCypherSerializer()
//	cypher, err := s.SerializeNodeType(nodeType)
//	fmt.Println(cypher)
package serializer

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"
)

// CypherSerializer serializes schema definitions to Cypher statements.
type CypherSerializer struct {
	// templates holds compiled Cypher templates.
	templates *template.Template
}

// NewCypherSerializer creates a new Cypher serializer.
func NewCypherSerializer() *CypherSerializer {
	s := &CypherSerializer{}
	s.templates = s.initTemplates()
	return s
}

// initTemplates initializes the Cypher templates.
func (s *CypherSerializer) initTemplates() *template.Template {
	tmpl := template.New("cypher").Funcs(template.FuncMap{
		"join":  strings.Join,
		"quote": func(s string) string { return fmt.Sprintf("`%s`", s) },
	})

	// Constraint templates
	template.Must(tmpl.New("unique_constraint").Parse(
		`CREATE CONSTRAINT {{.Name}} IF NOT EXISTS FOR (n:{{.Label}}) REQUIRE ({{range $i, $p := .Properties}}{{if $i}}, {{end}}n.{{$p}}{{end}}) IS UNIQUE`))

	template.Must(tmpl.New("exists_constraint").Parse(
		`CREATE CONSTRAINT {{.Name}} IF NOT EXISTS FOR (n:{{.Label}}) REQUIRE n.{{index .Properties 0}} IS NOT NULL`))

	template.Must(tmpl.New("node_key_constraint").Parse(
		`CREATE CONSTRAINT {{.Name}} IF NOT EXISTS FOR (n:{{.Label}}) REQUIRE ({{range $i, $p := .Properties}}{{if $i}}, {{end}}n.{{$p}}{{end}}) IS NODE KEY`))

	template.Must(tmpl.New("rel_exists_constraint").Parse(
		`CREATE CONSTRAINT {{.Name}} IF NOT EXISTS FOR ()-[r:{{.Label}}]-() REQUIRE r.{{index .Properties 0}} IS NOT NULL`))

	template.Must(tmpl.New("rel_key_constraint").Parse(
		`CREATE CONSTRAINT {{.Name}} IF NOT EXISTS FOR ()-[r:{{.Label}}]-() REQUIRE ({{range $i, $p := .Properties}}{{if $i}}, {{end}}r.{{$p}}{{end}}) IS RELATIONSHIP KEY`))

	// Index templates
	template.Must(tmpl.New("btree_index").Parse(
		`CREATE INDEX {{.Name}} IF NOT EXISTS FOR (n:{{.Label}}) ON ({{range $i, $p := .Properties}}{{if $i}}, {{end}}n.{{$p}}{{end}})`))

	template.Must(tmpl.New("text_index").Parse(
		`CREATE TEXT INDEX {{.Name}} IF NOT EXISTS FOR (n:{{.Label}}) ON (n.{{index .Properties 0}})`))

	template.Must(tmpl.New("fulltext_index").Parse(
		`CREATE FULLTEXT INDEX {{.Name}} IF NOT EXISTS FOR (n:{{.Label}}) ON EACH [{{range $i, $p := .Properties}}{{if $i}}, {{end}}n.{{$p}}{{end}}]`))

	template.Must(tmpl.New("point_index").Parse(
		`CREATE POINT INDEX {{.Name}} IF NOT EXISTS FOR (n:{{.Label}}) ON (n.{{index .Properties 0}})`))

	template.Must(tmpl.New("vector_index").Parse(
		`CREATE VECTOR INDEX {{.Name}} IF NOT EXISTS FOR (n:{{.Label}}) ON (n.{{index .Properties 0}}) OPTIONS {indexConfig: {` +
			"`vector.dimensions`" + `: {{.Dimensions}}, ` +
			"`vector.similarity_function`" + `: '{{.SimilarityFunction}}'}}`,
	))

	return tmpl
}

// constraintData holds data for constraint templates.
type constraintData struct {
	Name       string
	Label      string
	Properties []string
}

// indexData holds data for index templates.
type indexData struct {
	Name               string
	Label              string
	Properties         []string
	Dimensions         int
	SimilarityFunction string
}

// SerializeNodeType serializes a NodeType to Cypher statements.
func (s *CypherSerializer) SerializeNodeType(n *schema.NodeType) (string, error) {
	var statements []string

	// Generate constraint statements
	for _, c := range n.Constraints {
		stmt, err := s.serializeConstraint(n.Label, c)
		if err != nil {
			return "", fmt.Errorf("failed to serialize constraint %s: %w", c.Name, err)
		}
		statements = append(statements, stmt)
	}

	// Generate implicit constraints from properties
	for _, p := range n.Properties {
		if p.Required {
			name := fmt.Sprintf("%s_%s_not_null", strings.ToLower(n.Label), p.Name)
			c := schema.Constraint{Name: name, Type: schema.EXISTS, Properties: []string{p.Name}}
			stmt, err := s.serializeConstraint(n.Label, c)
			if err != nil {
				return "", fmt.Errorf("failed to serialize required constraint for %s: %w", p.Name, err)
			}
			statements = append(statements, stmt)
		}
		if p.Unique {
			name := fmt.Sprintf("%s_%s_unique", strings.ToLower(n.Label), p.Name)
			c := schema.Constraint{Name: name, Type: schema.UNIQUE, Properties: []string{p.Name}}
			stmt, err := s.serializeConstraint(n.Label, c)
			if err != nil {
				return "", fmt.Errorf("failed to serialize unique constraint for %s: %w", p.Name, err)
			}
			statements = append(statements, stmt)
		}
	}

	// Generate index statements
	for _, idx := range n.Indexes {
		stmt, err := s.serializeIndex(n.Label, idx)
		if err != nil {
			return "", fmt.Errorf("failed to serialize index %s: %w", idx.Name, err)
		}
		statements = append(statements, stmt)
	}

	return strings.Join(statements, ";\n") + ";", nil
}

// SerializeRelationshipType serializes a RelationshipType to Cypher statements.
func (s *CypherSerializer) SerializeRelationshipType(r *schema.RelationshipType) (string, error) {
	var statements []string

	// Generate constraint statements
	for _, c := range r.Constraints {
		stmt, err := s.serializeRelConstraint(r.Label, c)
		if err != nil {
			return "", fmt.Errorf("failed to serialize constraint %s: %w", c.Name, err)
		}
		statements = append(statements, stmt)
	}

	// Generate implicit constraints from properties
	for _, p := range r.Properties {
		if p.Required {
			name := fmt.Sprintf("%s_%s_not_null", strings.ToLower(r.Label), p.Name)
			c := schema.Constraint{Name: name, Type: schema.EXISTS, Properties: []string{p.Name}}
			stmt, err := s.serializeRelConstraint(r.Label, c)
			if err != nil {
				return "", fmt.Errorf("failed to serialize required constraint for %s: %w", p.Name, err)
			}
			statements = append(statements, stmt)
		}
	}

	if len(statements) == 0 {
		return "", nil
	}

	return strings.Join(statements, ";\n") + ";", nil
}

// serializeConstraint serializes a single node constraint.
func (s *CypherSerializer) serializeConstraint(label string, c schema.Constraint) (string, error) {
	data := constraintData{
		Name:       c.Name,
		Label:      label,
		Properties: c.Properties,
	}

	var tmplName string
	switch c.Type {
	case schema.UNIQUE:
		tmplName = "unique_constraint"
	case schema.EXISTS:
		tmplName = "exists_constraint"
	case schema.NODE_KEY:
		tmplName = "node_key_constraint"
	default:
		return "", fmt.Errorf("unsupported constraint type: %s", c.Type)
	}

	var buf bytes.Buffer
	if err := s.templates.ExecuteTemplate(&buf, tmplName, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// serializeRelConstraint serializes a single relationship constraint.
func (s *CypherSerializer) serializeRelConstraint(label string, c schema.Constraint) (string, error) {
	data := constraintData{
		Name:       c.Name,
		Label:      label,
		Properties: c.Properties,
	}

	var tmplName string
	switch c.Type {
	case schema.EXISTS:
		tmplName = "rel_exists_constraint"
	case schema.REL_KEY:
		tmplName = "rel_key_constraint"
	default:
		return "", fmt.Errorf("unsupported relationship constraint type: %s", c.Type)
	}

	var buf bytes.Buffer
	if err := s.templates.ExecuteTemplate(&buf, tmplName, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// serializeIndex serializes a single index.
func (s *CypherSerializer) serializeIndex(label string, idx schema.Index) (string, error) {
	data := indexData{
		Name:       idx.Name,
		Label:      label,
		Properties: idx.Properties,
	}

	// Extract vector-specific options
	if idx.Type == schema.VECTOR {
		if dims, ok := idx.Options["dimensions"].(int); ok {
			data.Dimensions = dims
		} else {
			data.Dimensions = 384 // default
		}
		if sim, ok := idx.Options["similarity_function"].(string); ok {
			data.SimilarityFunction = sim
		} else {
			data.SimilarityFunction = "cosine" // default
		}
	}

	var tmplName string
	switch idx.Type {
	case schema.BTREE:
		tmplName = "btree_index"
	case schema.TEXT:
		tmplName = "text_index"
	case schema.FULLTEXT:
		tmplName = "fulltext_index"
	case schema.POINT_INDEX:
		tmplName = "point_index"
	case schema.VECTOR:
		tmplName = "vector_index"
	default:
		return "", fmt.Errorf("unsupported index type: %s", idx.Type)
	}

	var buf bytes.Buffer
	if err := s.templates.ExecuteTemplate(&buf, tmplName, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// SerializeAll serializes multiple resources to Cypher statements in dependency order.
func (s *CypherSerializer) SerializeAll(nodeTypes []*schema.NodeType, relTypes []*schema.RelationshipType) (string, error) {
	var allStatements []string

	// Serialize node types first (they are dependencies for relationships)
	for _, n := range nodeTypes {
		stmt, err := s.SerializeNodeType(n)
		if err != nil {
			return "", fmt.Errorf("failed to serialize node type %s: %w", n.Label, err)
		}
		if stmt != "" {
			allStatements = append(allStatements, "// "+n.Label+" constraints and indexes")
			allStatements = append(allStatements, stmt)
		}
	}

	// Serialize relationship types
	for _, r := range relTypes {
		stmt, err := s.SerializeRelationshipType(r)
		if err != nil {
			return "", fmt.Errorf("failed to serialize relationship type %s: %w", r.Label, err)
		}
		if stmt != "" {
			allStatements = append(allStatements, "// "+r.Label+" constraints")
			allStatements = append(allStatements, stmt)
		}
	}

	return strings.Join(allStatements, "\n\n"), nil
}
