package projections

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"
)

// ProjectionSerializer serializes projection configurations to Cypher and JSON.
type ProjectionSerializer struct {
	templates *template.Template
}

// NewProjectionSerializer creates a new projection serializer.
func NewProjectionSerializer() *ProjectionSerializer {
	s := &ProjectionSerializer{}
	s.templates = s.initTemplates()
	return s
}

func (s *ProjectionSerializer) initTemplates() *template.Template {
	tmpl := template.New("projections")

	// Native projection template (simple form)
	template.Must(tmpl.New("native_simple").Parse(
		`CALL gds.graph.project(
  '{{.GraphName}}',
  {{.NodeLabels}},
  {{.RelationshipTypes}}{{if .Config}},
  {{.Config}}{{end}}
)
YIELD graphName, nodeCount, relationshipCount`))

	// Native projection template (extended form with properties)
	template.Must(tmpl.New("native_extended").Parse(
		`CALL gds.graph.project(
  '{{.GraphName}}',
  {{.NodeProjections}},
  {{.RelationshipProjections}}{{if .Config}},
  {{.Config}}{{end}}
)
YIELD graphName, nodeCount, relationshipCount`))

	// Cypher projection template
	template.Must(tmpl.New("cypher").Parse(
		`CALL gds.graph.project.cypher(
  '{{.GraphName}}',
  '{{.NodeQuery}}',
  '{{.RelationshipQuery}}'{{if .Config}},
  {{.Config}}{{end}}
)
YIELD graphName, nodeCount, relationshipCount`))

	// Drop graph template
	template.Must(tmpl.New("drop").Parse(
		`CALL gds.graph.drop('{{.GraphName}}') YIELD graphName`))

	// Check graph exists template
	template.Must(tmpl.New("exists").Parse(
		`RETURN gds.graph.exists('{{.GraphName}}') AS exists`))

	// List graphs template
	template.Must(tmpl.New("list").Parse(
		`CALL gds.graph.list() YIELD graphName, nodeCount, relationshipCount`))

	return tmpl
}

// ToCypher converts a projection configuration to a Cypher CALL statement.
func (s *ProjectionSerializer) ToCypher(projection Projection) (string, error) {
	switch p := projection.(type) {
	case *NativeProjection:
		return s.nativeToCypher(p)
	case *CypherProjection:
		return s.cypherToCypher(p)
	case *DataFrameProjection:
		return s.dataframeToCypher(p)
	default:
		return "", fmt.Errorf("unknown projection type: %T", projection)
	}
}

func (s *ProjectionSerializer) nativeToCypher(p *NativeProjection) (string, error) {
	// Use simple form if only labels/types are specified
	if len(p.NodeProjections) == 0 && len(p.RelationshipProjections) == 0 {
		return s.nativeSimpleToCypher(p)
	}
	return s.nativeExtendedToCypher(p)
}

func (s *ProjectionSerializer) nativeSimpleToCypher(p *NativeProjection) (string, error) {
	nodeLabels := formatLabels(p.NodeLabels)
	relTypes := formatLabels(p.RelationshipTypes)

	config := s.buildNativeConfig(p)

	data := map[string]string{
		"GraphName":         p.GraphName,
		"NodeLabels":        nodeLabels,
		"RelationshipTypes": relTypes,
		"Config":            config,
	}

	var buf bytes.Buffer
	if err := s.templates.ExecuteTemplate(&buf, "native_simple", data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (s *ProjectionSerializer) nativeExtendedToCypher(p *NativeProjection) (string, error) {
	nodeProjections := s.formatNodeProjections(p.GetNodeProjections())
	relProjections := s.formatRelationshipProjections(p.GetRelationshipProjections())

	config := s.buildNativeConfig(p)

	data := map[string]string{
		"GraphName":               p.GraphName,
		"NodeProjections":         nodeProjections,
		"RelationshipProjections": relProjections,
		"Config":                  config,
	}

	var buf bytes.Buffer
	if err := s.templates.ExecuteTemplate(&buf, "native_extended", data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (s *ProjectionSerializer) cypherToCypher(p *CypherProjection) (string, error) {
	config := s.buildCypherConfig(p)

	data := map[string]string{
		"GraphName":         p.GraphName,
		"NodeQuery":         escapeString(p.NodeQuery),
		"RelationshipQuery": escapeString(p.RelationshipQuery),
		"Config":            config,
	}

	var buf bytes.Buffer
	if err := s.templates.ExecuteTemplate(&buf, "cypher", data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (s *ProjectionSerializer) dataframeToCypher(p *DataFrameProjection) (string, error) {
	// DataFrame projections are primarily used with Python client
	// Generate a comment explaining this
	return fmt.Sprintf("// DataFrame projection '%s' - use with GDS Python client\n"+
		"// gds.graph.construct(\n"+
		"//   '%s',\n"+
		"//   nodes_df,\n"+
		"//   relationships_df\n"+
		"// )", p.Name, p.GraphName), nil
}

func (s *ProjectionSerializer) buildNativeConfig(p *NativeProjection) string {
	var parts []string

	if p.ReadConcurrency > 0 {
		parts = append(parts, fmt.Sprintf("readConcurrency: %d", p.ReadConcurrency))
	}

	if len(parts) == 0 {
		return ""
	}

	return "{\n    " + strings.Join(parts, ",\n    ") + "\n  }"
}

func (s *ProjectionSerializer) buildCypherConfig(p *CypherProjection) string {
	var parts []string

	if p.ReadConcurrency > 0 {
		parts = append(parts, fmt.Sprintf("readConcurrency: %d", p.ReadConcurrency))
	}
	if p.ValidateRelationships {
		parts = append(parts, "validateRelationships: true")
	}
	if len(p.Parameters) > 0 {
		params := make([]string, 0, len(p.Parameters))
		for k, v := range p.Parameters {
			params = append(params, fmt.Sprintf("%s: %s", k, formatValue(v)))
		}
		parts = append(parts, "parameters: {\n      "+strings.Join(params, ",\n      ")+"\n    }")
	}

	if len(parts) == 0 {
		return ""
	}

	return "{\n    " + strings.Join(parts, ",\n    ") + "\n  }"
}

func (s *ProjectionSerializer) formatNodeProjections(projections []NodeProjection) string {
	if len(projections) == 0 {
		return "'*'"
	}
	if len(projections) == 1 && len(projections[0].Properties) == 0 {
		return fmt.Sprintf("'%s'", projections[0].Label)
	}

	parts := make([]string, len(projections))
	for i, np := range projections {
		if len(np.Properties) == 0 {
			parts[i] = fmt.Sprintf("%s: {label: '%s'}", np.Label, np.Label)
		} else {
			props := formatLabels(np.Properties)
			parts[i] = fmt.Sprintf("%s: {\n      label: '%s',\n      properties: %s\n    }",
				np.Label, np.Label, props)
		}
	}

	return "{\n    " + strings.Join(parts, ",\n    ") + "\n  }"
}

func (s *ProjectionSerializer) formatRelationshipProjections(projections []RelationshipProjection) string {
	if len(projections) == 0 {
		return "'*'"
	}
	if len(projections) == 1 && len(projections[0].Properties) == 0 && projections[0].Orientation == "" {
		return fmt.Sprintf("'%s'", projections[0].Type)
	}

	parts := make([]string, len(projections))
	for i, rp := range projections {
		var config []string
		config = append(config, fmt.Sprintf("type: '%s'", rp.Type))

		if rp.Orientation != "" {
			config = append(config, fmt.Sprintf("orientation: '%s'", rp.Orientation))
		}
		if rp.Aggregation != "" {
			config = append(config, fmt.Sprintf("aggregation: '%s'", rp.Aggregation))
		}
		if len(rp.Properties) > 0 {
			config = append(config, fmt.Sprintf("properties: %s", formatLabels(rp.Properties)))
		}

		parts[i] = fmt.Sprintf("%s: {\n      %s\n    }", rp.Type, strings.Join(config, ",\n      "))
	}

	return "{\n    " + strings.Join(parts, ",\n    ") + "\n  }"
}

// ToJSON converts a projection configuration to JSON.
func (s *ProjectionSerializer) ToJSON(projection Projection) ([]byte, error) {
	data := s.ToMap(projection)
	return json.MarshalIndent(data, "", "  ")
}

// ToMap converts a projection to a map.
func (s *ProjectionSerializer) ToMap(projection Projection) map[string]any {
	result := map[string]any{
		"name":           projection.ProjectionName(),
		"projectionType": string(projection.ProjectionType()),
	}

	switch p := projection.(type) {
	case *NativeProjection:
		result["graphName"] = p.GraphName
		if len(p.NodeLabels) > 0 {
			result["nodeLabels"] = p.NodeLabels
		}
		if len(p.RelationshipTypes) > 0 {
			result["relationshipTypes"] = p.RelationshipTypes
		}
		if len(p.NodeProjections) > 0 {
			result["nodeProjections"] = s.nodeProjectionsToMaps(p.NodeProjections)
		}
		if len(p.RelationshipProjections) > 0 {
			result["relationshipProjections"] = s.relProjectionsToMaps(p.RelationshipProjections)
		}
		if p.ReadConcurrency > 0 {
			result["readConcurrency"] = p.ReadConcurrency
		}

	case *CypherProjection:
		result["graphName"] = p.GraphName
		if p.NodeQuery != "" {
			result["nodeQuery"] = p.NodeQuery
		}
		if p.RelationshipQuery != "" {
			result["relationshipQuery"] = p.RelationshipQuery
		}
		if len(p.Parameters) > 0 {
			result["parameters"] = p.Parameters
		}
		if p.ValidateRelationships {
			result["validateRelationships"] = p.ValidateRelationships
		}
		if p.ReadConcurrency > 0 {
			result["readConcurrency"] = p.ReadConcurrency
		}

	case *DataFrameProjection:
		result["graphName"] = p.GraphName
		if len(p.NodeDataFrames) > 0 {
			result["nodeDataFrames"] = s.nodeDataFramesToMaps(p.NodeDataFrames)
		}
		if len(p.RelationshipDataFrames) > 0 {
			result["relationshipDataFrames"] = s.relDataFramesToMaps(p.RelationshipDataFrames)
		}
	}

	return result
}

func (s *ProjectionSerializer) nodeProjectionsToMaps(projections []NodeProjection) []map[string]any {
	result := make([]map[string]any, len(projections))
	for i, np := range projections {
		m := map[string]any{"label": np.Label}
		if len(np.Properties) > 0 {
			m["properties"] = np.Properties
		}
		if np.DefaultValue != nil {
			m["defaultValue"] = np.DefaultValue
		}
		result[i] = m
	}
	return result
}

func (s *ProjectionSerializer) relProjectionsToMaps(projections []RelationshipProjection) []map[string]any {
	result := make([]map[string]any, len(projections))
	for i, rp := range projections {
		m := map[string]any{"type": rp.Type}
		if rp.Orientation != "" {
			m["orientation"] = string(rp.Orientation)
		}
		if rp.Aggregation != "" {
			m["aggregation"] = string(rp.Aggregation)
		}
		if len(rp.Properties) > 0 {
			m["properties"] = rp.Properties
		}
		if rp.DefaultValue != nil {
			m["defaultValue"] = rp.DefaultValue
		}
		result[i] = m
	}
	return result
}

func (s *ProjectionSerializer) nodeDataFramesToMaps(dataframes []NodeDataFrame) []map[string]any {
	result := make([]map[string]any, len(dataframes))
	for i, df := range dataframes {
		m := map[string]any{"label": df.Label}
		if len(df.Properties) > 0 {
			m["properties"] = df.Properties
		}
		if df.IDColumn != "" {
			m["idColumn"] = df.IDColumn
		}
		result[i] = m
	}
	return result
}

func (s *ProjectionSerializer) relDataFramesToMaps(dataframes []RelationshipDataFrame) []map[string]any {
	result := make([]map[string]any, len(dataframes))
	for i, df := range dataframes {
		m := map[string]any{"type": df.Type}
		if df.SourceColumn != "" {
			m["sourceColumn"] = df.SourceColumn
		}
		if df.TargetColumn != "" {
			m["targetColumn"] = df.TargetColumn
		}
		if len(df.Properties) > 0 {
			m["properties"] = df.Properties
		}
		result[i] = m
	}
	return result
}

// DropGraph generates a Cypher statement to drop a projected graph.
func (s *ProjectionSerializer) DropGraph(graphName string) string {
	var buf bytes.Buffer
	_ = s.templates.ExecuteTemplate(&buf, "drop", map[string]string{"GraphName": graphName})
	return buf.String()
}

// GraphExists generates a Cypher statement to check if a graph exists.
func (s *ProjectionSerializer) GraphExists(graphName string) string {
	var buf bytes.Buffer
	_ = s.templates.ExecuteTemplate(&buf, "exists", map[string]string{"GraphName": graphName})
	return buf.String()
}

// ListGraphs generates a Cypher statement to list all projected graphs.
func (s *ProjectionSerializer) ListGraphs() string {
	var buf bytes.Buffer
	_ = s.templates.ExecuteTemplate(&buf, "list", nil)
	return buf.String()
}

// formatLabels formats a string slice for Cypher.
func formatLabels(labels []string) string {
	if len(labels) == 0 {
		return "'*'"
	}
	if len(labels) == 1 {
		return fmt.Sprintf("'%s'", labels[0])
	}
	quoted := make([]string, len(labels))
	for i, l := range labels {
		quoted[i] = fmt.Sprintf("'%s'", l)
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

// formatValue formats a value for Cypher.
func formatValue(v any) string {
	switch val := v.(type) {
	case string:
		return fmt.Sprintf("'%s'", val)
	case []string:
		return formatLabels(val)
	case bool:
		if val {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", v)
	}
}

// escapeString escapes single quotes in a string for Cypher.
func escapeString(s string) string {
	return strings.ReplaceAll(s, "'", "\\'")
}
