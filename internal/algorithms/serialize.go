package algorithms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"text/template"
)

// AlgorithmSerializer serializes algorithm configurations to Cypher and JSON.
type AlgorithmSerializer struct {
	templates *template.Template
}

// NewAlgorithmSerializer creates a new algorithm serializer.
func NewAlgorithmSerializer() *AlgorithmSerializer {
	s := &AlgorithmSerializer{}
	s.templates = s.initTemplates()
	return s
}

func (s *AlgorithmSerializer) initTemplates() *template.Template {
	tmpl := template.New("gds")

	// Generic algorithm call template
	template.Must(tmpl.New("algorithm_call").Parse(
		`CALL {{.Procedure}}(
  '{{.GraphName}}'{{if .Config}},
  {
{{.Config}}
  }{{end}}
)
YIELD {{.YieldFields}}`))

	return tmpl
}

// ToCypher converts an algorithm configuration to a Cypher CALL statement.
func (s *AlgorithmSerializer) ToCypher(algo Algorithm) (string, error) {
	procedure := algo.AlgorithmType() + "." + string(algo.GetMode())
	config := s.buildConfig(algo)
	yieldFields := s.getYieldFields(algo)

	data := map[string]string{
		"Procedure":   procedure,
		"GraphName":   algo.GetGraphName(),
		"Config":      config,
		"YieldFields": yieldFields,
	}

	var buf bytes.Buffer
	if err := s.templates.ExecuteTemplate(&buf, "algorithm_call", data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// ToJSON converts an algorithm configuration to JSON.
func (s *AlgorithmSerializer) ToJSON(algo Algorithm) ([]byte, error) {
	params := s.toMap(algo)
	return json.MarshalIndent(params, "", "  ")
}

// ToMap converts an algorithm configuration to a map.
func (s *AlgorithmSerializer) ToMap(algo Algorithm) map[string]any {
	return s.toMap(algo)
}

// buildConfig builds the configuration parameters string for Cypher.
func (s *AlgorithmSerializer) buildConfig(algo Algorithm) string {
	params := s.toMap(algo)

	// Remove meta fields
	delete(params, "name")
	delete(params, "graphName")
	delete(params, "mode")
	delete(params, "algorithmType")
	delete(params, "category")

	if len(params) == 0 {
		return ""
	}

	var lines []string
	for k, v := range params {
		line := fmt.Sprintf("    %s: %s", k, formatValue(v))
		lines = append(lines, line)
	}

	return strings.Join(lines, ",\n")
}

// toMap converts an algorithm to a map using reflection.
func (s *AlgorithmSerializer) toMap(algo Algorithm) map[string]any {
	result := make(map[string]any)

	val := reflect.ValueOf(algo)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	typ := val.Type()

	// Process embedded BaseAlgorithm first
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		if field.Anonymous && field.Type == reflect.TypeOf(BaseAlgorithm{}) {
			base := fieldVal.Interface().(BaseAlgorithm)
			if base.Name != "" {
				result["name"] = base.Name
			}
			if base.GraphName != "" {
				result["graphName"] = base.GraphName
			}
			if base.Mode != "" {
				result["mode"] = string(base.Mode)
			}
			if base.Concurrency > 0 {
				result["concurrency"] = base.Concurrency
			}
			if len(base.NodeLabels) > 0 {
				result["nodeLabels"] = base.NodeLabels
			}
			if len(base.RelationshipTypes) > 0 {
				result["relationshipTypes"] = base.RelationshipTypes
			}
			continue
		}

		// Skip unexported fields and zero values
		if !field.IsExported() {
			continue
		}
		if isZeroValue(fieldVal) {
			continue
		}

		// Convert field name to camelCase
		name := toCamelCase(field.Name)
		result[name] = fieldVal.Interface()
	}

	// Add algorithm metadata
	result["algorithmType"] = algo.AlgorithmType()
	result["category"] = string(algo.AlgorithmCategory())

	return result
}

// getYieldFields returns the YIELD fields based on algorithm category and mode.
func (s *AlgorithmSerializer) getYieldFields(algo Algorithm) string {
	mode := algo.GetMode()
	category := algo.AlgorithmCategory()

	switch mode {
	case Stats:
		return "nodeCount, relationshipCount, computeMillis"
	case Write:
		return "nodePropertiesWritten, computeMillis"
	case Mutate:
		return "nodePropertiesWritten, computeMillis"
	case Stream:
		switch category {
		case Centrality:
			return "nodeId, score"
		case Community:
			return "nodeId, communityId"
		case Similarity:
			return "node1, node2, similarity"
		case Embeddings:
			return "nodeId, embedding"
		case PathFinding:
			return "sourceNode, targetNode, path, totalCost"
		default:
			return "*"
		}
	}

	return "*"
}

// formatValue formats a value for Cypher.
func formatValue(v any) string {
	switch val := v.(type) {
	case string:
		return fmt.Sprintf("'%s'", val)
	case []string:
		quoted := make([]string, len(val))
		for i, s := range val {
			quoted[i] = fmt.Sprintf("'%s'", s)
		}
		return "[" + strings.Join(quoted, ", ") + "]"
	case []float64:
		strs := make([]string, len(val))
		for i, f := range val {
			strs[i] = fmt.Sprintf("%v", f)
		}
		return "[" + strings.Join(strs, ", ") + "]"
	case []int:
		strs := make([]string, len(val))
		for i, n := range val {
			strs[i] = fmt.Sprintf("%d", n)
		}
		return "[" + strings.Join(strs, ", ") + "]"
	case bool:
		if val {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", v)
	}
}

// toCamelCase converts a Go field name to camelCase.
func toCamelCase(s string) string {
	if s == "" {
		return s
	}
	// Handle acronyms
	s = strings.ReplaceAll(s, "ID", "Id")
	s = strings.ReplaceAll(s, "URL", "Url")

	// Convert first letter to lowercase
	runes := []rune(s)
	runes[0] = rune(strings.ToLower(string(runes[0]))[0])
	return string(runes)
}

// isZeroValue checks if a reflect.Value is the zero value.
func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Slice, reflect.Map, reflect.Array:
		return v.Len() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

// BatchToCypher converts multiple algorithms to a batch of Cypher statements.
func (s *AlgorithmSerializer) BatchToCypher(algorithms []Algorithm) (string, error) {
	var statements []string

	for _, algo := range algorithms {
		stmt, err := s.ToCypher(algo)
		if err != nil {
			return "", fmt.Errorf("failed to serialize %s: %w", algo.AlgorithmName(), err)
		}
		comment := fmt.Sprintf("// %s - %s", algo.AlgorithmName(), algo.AlgorithmType())
		statements = append(statements, comment+"\n"+stmt)
	}

	return strings.Join(statements, "\n\n"), nil
}

// BatchToJSON converts multiple algorithms to a JSON array.
func (s *AlgorithmSerializer) BatchToJSON(algorithms []Algorithm) ([]byte, error) {
	var configs []map[string]any

	for _, algo := range algorithms {
		configs = append(configs, s.ToMap(algo))
	}

	return json.MarshalIndent(configs, "", "  ")
}
