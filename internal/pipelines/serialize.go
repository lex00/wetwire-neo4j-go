package pipelines

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"text/template"
)

// PipelineSerializer serializes pipeline configurations to Cypher and JSON.
type PipelineSerializer struct {
	templates *template.Template
}

// NewPipelineSerializer creates a new pipeline serializer.
func NewPipelineSerializer() *PipelineSerializer {
	s := &PipelineSerializer{}
	s.templates = s.initTemplates()
	return s
}

func (s *PipelineSerializer) initTemplates() *template.Template {
	tmpl := template.New("pipelines")

	// Create pipeline template
	template.Must(tmpl.New("create_pipeline").Parse(
		`CALL gds.{{.PipelineType}}.pipeline.create('{{.Name}}')`))

	// Add node property step template
	template.Must(tmpl.New("add_node_property").Parse(
		`CALL gds.{{.PipelineType}}.pipeline.addNodeProperty(
  '{{.PipelineName}}',
  '{{.StepType}}',
  {
    mutateProperty: '{{.MutateProperty}}'{{.Config}}
  }
)`))

	// Add model template
	template.Must(tmpl.New("add_model").Parse(
		`CALL gds.{{.PipelineType}}.pipeline.addModel(
  '{{.PipelineName}}',
  '{{.ModelType}}',
  {{ "{" }}{{.Config}}
  }
)`))

	// Configure split template
	template.Must(tmpl.New("configure_split").Parse(
		`CALL gds.{{.PipelineType}}.pipeline.configureSplit(
  '{{.PipelineName}}',
  {
    testFraction: {{.TestFraction}},
    validationFolds: {{.ValidationFolds}}
  }
)`))

	// Train template
	template.Must(tmpl.New("train").Parse(
		`CALL gds.{{.PipelineType}}.pipeline.train(
  '{{.GraphName}}',
  {
    pipeline: '{{.PipelineName}}',
    targetProperty: '{{.TargetProperty}}',
    modelName: '{{.ModelName}}'{{.Config}}
  }
) YIELD modelInfo
RETURN modelInfo`))

	return tmpl
}

// ToCypher generates Cypher statements for creating and configuring a pipeline.
func (s *PipelineSerializer) ToCypher(pipeline Pipeline, graphName, modelName string) (string, error) {
	var statements []string

	pipelineType := s.getPipelineTypeName(pipeline)

	// Create pipeline
	createData := map[string]string{
		"PipelineType": pipelineType,
		"Name":         pipeline.PipelineName(),
	}
	var buf bytes.Buffer
	if err := s.templates.ExecuteTemplate(&buf, "create_pipeline", createData); err != nil {
		return "", err
	}
	statements = append(statements, buf.String())

	// Add feature steps
	for _, step := range pipeline.GetFeatureSteps() {
		stepCypher, err := s.serializeFeatureStep(pipeline, step)
		if err != nil {
			return "", err
		}
		statements = append(statements, stepCypher)
	}

	// Add models
	for _, model := range pipeline.GetModels() {
		modelCypher, err := s.serializeModel(pipeline, model)
		if err != nil {
			return "", err
		}
		statements = append(statements, modelCypher)
	}

	// Configure split
	splitCypher, err := s.serializeSplitConfig(pipeline)
	if err != nil {
		return "", err
	}
	if splitCypher != "" {
		statements = append(statements, splitCypher)
	}

	// Add train command
	trainCypher, err := s.serializeTrainCommand(pipeline, graphName, modelName)
	if err != nil {
		return "", err
	}
	statements = append(statements, trainCypher)

	return strings.Join(statements, ";\n\n") + ";", nil
}

// ToJSON converts a pipeline configuration to JSON.
func (s *PipelineSerializer) ToJSON(pipeline Pipeline) ([]byte, error) {
	data := s.toMap(pipeline)
	return json.MarshalIndent(data, "", "  ")
}

// ToMap converts a pipeline to a map.
func (s *PipelineSerializer) ToMap(pipeline Pipeline) map[string]any {
	return s.toMap(pipeline)
}

// getPipelineTypeName returns the GDS procedure name for the pipeline type.
func (s *PipelineSerializer) getPipelineTypeName(pipeline Pipeline) string {
	switch pipeline.PipelineType() {
	case NodeClassification:
		return "beta.pipeline.nodeClassification"
	case LinkPrediction:
		return "beta.pipeline.linkPrediction"
	case NodeRegression:
		return "alpha.pipeline.nodeRegression"
	default:
		return "pipeline"
	}
}

// serializeFeatureStep generates Cypher for a feature step.
func (s *PipelineSerializer) serializeFeatureStep(pipeline Pipeline, step FeatureStep) (string, error) {
	config := s.buildStepConfig(step)

	data := map[string]string{
		"PipelineType":   s.getPipelineTypeName(pipeline),
		"PipelineName":   pipeline.PipelineName(),
		"StepType":       step.StepType(),
		"MutateProperty": step.MutateProperty(),
		"Config":         config,
	}

	var buf bytes.Buffer
	if err := s.templates.ExecuteTemplate(&buf, "add_node_property", data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// serializeModel generates Cypher for a model candidate.
func (s *PipelineSerializer) serializeModel(pipeline Pipeline, model Model) (string, error) {
	config := s.buildModelConfig(model)

	data := map[string]string{
		"PipelineType": s.getPipelineTypeName(pipeline),
		"PipelineName": pipeline.PipelineName(),
		"ModelType":    model.ModelType(),
		"Config":       config,
	}

	var buf bytes.Buffer
	if err := s.templates.ExecuteTemplate(&buf, "add_model", data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// serializeSplitConfig generates Cypher for split configuration.
func (s *PipelineSerializer) serializeSplitConfig(pipeline Pipeline) (string, error) {
	var splitConfig SplitConfig

	switch p := pipeline.(type) {
	case *NodeClassificationPipeline:
		splitConfig = p.SplitConfig
	case *LinkPredictionPipeline:
		splitConfig = p.SplitConfig
	case *NodeRegressionPipeline:
		splitConfig = p.SplitConfig
	default:
		return "", nil
	}

	// Skip if using defaults
	if splitConfig.TestFraction == 0 && splitConfig.ValidationFolds == 0 {
		return "", nil
	}

	testFraction := splitConfig.TestFraction
	if testFraction == 0 {
		testFraction = 0.2
	}
	validationFolds := splitConfig.ValidationFolds
	if validationFolds == 0 {
		validationFolds = 5
	}

	data := map[string]any{
		"PipelineType":    s.getPipelineTypeName(pipeline),
		"PipelineName":    pipeline.PipelineName(),
		"TestFraction":    testFraction,
		"ValidationFolds": validationFolds,
	}

	var buf bytes.Buffer
	if err := s.templates.ExecuteTemplate(&buf, "configure_split", data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// serializeTrainCommand generates the train command.
func (s *PipelineSerializer) serializeTrainCommand(pipeline Pipeline, graphName, modelName string) (string, error) {
	var targetProperty string
	var configParts []string

	switch p := pipeline.(type) {
	case *NodeClassificationPipeline:
		targetProperty = p.TargetProperty
		if len(p.TargetNodeLabels) > 0 {
			configParts = append(configParts, fmt.Sprintf("targetNodeLabels: %s", formatStringSlice(p.TargetNodeLabels)))
		}
	case *LinkPredictionPipeline:
		targetProperty = p.TargetRelationshipType
		if p.NegativeSamplingRatio > 0 {
			configParts = append(configParts, fmt.Sprintf("negativeSamplingRatio: %v", p.NegativeSamplingRatio))
		}
	case *NodeRegressionPipeline:
		targetProperty = p.TargetProperty
		if len(p.TargetNodeLabels) > 0 {
			configParts = append(configParts, fmt.Sprintf("targetNodeLabels: %s", formatStringSlice(p.TargetNodeLabels)))
		}
	}

	config := ""
	if len(configParts) > 0 {
		config = ",\n    " + strings.Join(configParts, ",\n    ")
	}

	data := map[string]string{
		"PipelineType":   s.getPipelineTypeName(pipeline),
		"GraphName":      graphName,
		"PipelineName":   pipeline.PipelineName(),
		"TargetProperty": targetProperty,
		"ModelName":      modelName,
		"Config":         config,
	}

	var buf bytes.Buffer
	if err := s.templates.ExecuteTemplate(&buf, "train", data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// buildStepConfig builds configuration parameters for a feature step.
func (s *PipelineSerializer) buildStepConfig(step FeatureStep) string {
	params := s.structToParams(step)

	// Remove the property field as it's already in mutateProperty
	delete(params, "property")

	if len(params) == 0 {
		return ""
	}

	var parts []string
	for k, v := range params {
		parts = append(parts, fmt.Sprintf("%s: %s", k, formatValue(v)))
	}

	return ",\n    " + strings.Join(parts, ",\n    ")
}

// buildModelConfig builds configuration parameters for a model.
func (s *PipelineSerializer) buildModelConfig(model Model) string {
	params := s.structToParams(model)

	if len(params) == 0 {
		return ""
	}

	var parts []string
	for k, v := range params {
		parts = append(parts, fmt.Sprintf("%s: %s", k, formatValue(v)))
	}

	return "\n    " + strings.Join(parts, ",\n    ")
}

// structToParams converts a struct to parameter map.
func (s *PipelineSerializer) structToParams(v any) map[string]any {
	result := make(map[string]any)

	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		if !field.IsExported() {
			continue
		}
		if isZeroValue(fieldVal) {
			continue
		}

		name := toCamelCase(field.Name)
		result[name] = fieldVal.Interface()
	}

	return result
}

// toMap converts a pipeline to a map representation.
func (s *PipelineSerializer) toMap(pipeline Pipeline) map[string]any {
	result := map[string]any{
		"name":         pipeline.PipelineName(),
		"pipelineType": string(pipeline.PipelineType()),
	}

	// Add feature steps
	var steps []map[string]any
	for _, step := range pipeline.GetFeatureSteps() {
		stepMap := map[string]any{
			"type":           step.StepType(),
			"mutateProperty": step.MutateProperty(),
		}
		for k, v := range s.structToParams(step) {
			if k != "property" {
				stepMap[k] = v
			}
		}
		steps = append(steps, stepMap)
	}
	if len(steps) > 0 {
		result["featureSteps"] = steps
	}

	// Add models
	var models []map[string]any
	for _, model := range pipeline.GetModels() {
		modelMap := map[string]any{
			"type": model.ModelType(),
		}
		for k, v := range s.structToParams(model) {
			modelMap[k] = v
		}
		models = append(models, modelMap)
	}
	if len(models) > 0 {
		result["models"] = models
	}

	// Add pipeline-specific fields
	switch p := pipeline.(type) {
	case *NodeClassificationPipeline:
		result["targetProperty"] = p.TargetProperty
		if len(p.TargetNodeLabels) > 0 {
			result["targetNodeLabels"] = p.TargetNodeLabels
		}
		if len(p.FeatureProperties) > 0 {
			result["featureProperties"] = p.FeatureProperties
		}
	case *LinkPredictionPipeline:
		result["targetRelationshipType"] = p.TargetRelationshipType
		if len(p.SourceNodeLabels) > 0 {
			result["sourceNodeLabels"] = p.SourceNodeLabels
		}
		if len(p.TargetNodeLabels) > 0 {
			result["targetNodeLabels"] = p.TargetNodeLabels
		}
		if p.NegativeSamplingRatio > 0 {
			result["negativeSamplingRatio"] = p.NegativeSamplingRatio
		}
	case *NodeRegressionPipeline:
		result["targetProperty"] = p.TargetProperty
		if len(p.TargetNodeLabels) > 0 {
			result["targetNodeLabels"] = p.TargetNodeLabels
		}
	}

	return result
}

// formatValue formats a value for Cypher.
func formatValue(v any) string {
	switch val := v.(type) {
	case string:
		return fmt.Sprintf("'%s'", val)
	case []string:
		return formatStringSlice(val)
	case []int:
		strs := make([]string, len(val))
		for i, n := range val {
			strs[i] = fmt.Sprintf("%d", n)
		}
		return "[" + strings.Join(strs, ", ") + "]"
	case []float64:
		strs := make([]string, len(val))
		for i, f := range val {
			strs[i] = fmt.Sprintf("%v", f)
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

// formatStringSlice formats a string slice for Cypher.
func formatStringSlice(ss []string) string {
	quoted := make([]string, len(ss))
	for i, s := range ss {
		quoted[i] = fmt.Sprintf("'%s'", s)
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

// toCamelCase converts a Go field name to camelCase.
func toCamelCase(s string) string {
	if s == "" {
		return s
	}
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
