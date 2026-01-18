package pipelines

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNodeClassificationPipeline_Interface(t *testing.T) {
	p := &NodeClassificationPipeline{
		BasePipeline: BasePipeline{
			Name: "fraud_detection",
		},
		TargetProperty: "isFraud",
	}

	if p.PipelineName() != "fraud_detection" {
		t.Errorf("PipelineName() = %v, want fraud_detection", p.PipelineName())
	}
	if p.PipelineType() != NodeClassification {
		t.Errorf("PipelineType() = %v, want NodeClassification", p.PipelineType())
	}
}

func TestLinkPredictionPipeline_Interface(t *testing.T) {
	p := &LinkPredictionPipeline{
		BasePipeline: BasePipeline{
			Name: "link_predictor",
		},
		TargetRelationshipType: "KNOWS",
	}

	if p.PipelineType() != LinkPrediction {
		t.Errorf("PipelineType() = %v, want LinkPrediction", p.PipelineType())
	}
}

func TestNodeRegressionPipeline_Interface(t *testing.T) {
	p := &NodeRegressionPipeline{
		BasePipeline: BasePipeline{
			Name: "price_predictor",
		},
		TargetProperty: "price",
	}

	if p.PipelineType() != NodeRegression {
		t.Errorf("PipelineType() = %v, want NodeRegression", p.PipelineType())
	}
}

func TestBasePipeline_AddFeatureStep(t *testing.T) {
	p := &NodeClassificationPipeline{
		BasePipeline: BasePipeline{Name: "test"},
	}

	step := &FastRPStep{Property: "embedding", EmbeddingDimension: 128}
	p.AddFeatureStep(step)

	if len(p.GetFeatureSteps()) != 1 {
		t.Errorf("GetFeatureSteps() count = %v, want 1", len(p.GetFeatureSteps()))
	}
}

func TestBasePipeline_AddModel(t *testing.T) {
	p := &NodeClassificationPipeline{
		BasePipeline: BasePipeline{Name: "test"},
	}

	model := &LogisticRegression{Penalty: 0.001}
	p.AddModel(model)

	if len(p.GetModels()) != 1 {
		t.Errorf("GetModels() count = %v, want 1", len(p.GetModels()))
	}
}

func TestFeatureSteps(t *testing.T) {
	tests := []struct {
		step     FeatureStep
		stepType string
		property string
	}{
		{&FastRPStep{Property: "emb"}, "fastRP", "emb"},
		{&PageRankStep{Property: "pr"}, "pageRank", "pr"},
		{&DegreeStep{Property: "deg"}, "degree", "deg"},
		{&Node2VecStep{Property: "n2v"}, "node2Vec", "n2v"},
		{&ScalerStep{Property: "scaled"}, "scaler", "scaled"},
	}

	for _, tt := range tests {
		t.Run(tt.stepType, func(t *testing.T) {
			if tt.step.StepType() != tt.stepType {
				t.Errorf("StepType() = %v, want %v", tt.step.StepType(), tt.stepType)
			}
			if tt.step.MutateProperty() != tt.property {
				t.Errorf("MutateProperty() = %v, want %v", tt.step.MutateProperty(), tt.property)
			}
		})
	}
}

func TestModels(t *testing.T) {
	tests := []struct {
		model     Model
		modelType string
		pipelines []PipelineType
	}{
		{
			&LogisticRegression{},
			"LogisticRegression",
			[]PipelineType{NodeClassification, LinkPrediction},
		},
		{
			&RandomForest{},
			"RandomForest",
			[]PipelineType{NodeClassification, LinkPrediction, NodeRegression},
		},
		{
			&MLP{},
			"MLP",
			[]PipelineType{NodeClassification, LinkPrediction},
		},
		{
			&LinearRegression{},
			"LinearRegression",
			[]PipelineType{NodeRegression},
		},
	}

	for _, tt := range tests {
		t.Run(tt.modelType, func(t *testing.T) {
			if tt.model.ModelType() != tt.modelType {
				t.Errorf("ModelType() = %v, want %v", tt.model.ModelType(), tt.modelType)
			}
			supported := tt.model.SupportedPipelines()
			if len(supported) != len(tt.pipelines) {
				t.Errorf("SupportedPipelines() count = %v, want %v", len(supported), len(tt.pipelines))
			}
		})
	}
}

func TestNewPipelineSerializer(t *testing.T) {
	s := NewPipelineSerializer()
	if s == nil {
		t.Fatal("NewPipelineSerializer returned nil")
	}
	if s.templates == nil {
		t.Error("templates is nil")
	}
}

func TestPipelineSerializer_ToCypher_NodeClassification(t *testing.T) {
	s := NewPipelineSerializer()
	p := &NodeClassificationPipeline{
		BasePipeline: BasePipeline{
			Name: "fraud_detection",
			FeatureSteps: []FeatureStep{
				&FastRPStep{Property: "embedding", EmbeddingDimension: 128},
				&PageRankStep{Property: "pagerank", DampingFactor: 0.85},
			},
			Models: []Model{
				&LogisticRegression{Penalty: 0.001, MaxEpochs: 100},
			},
		},
		TargetProperty:   "isFraud",
		TargetNodeLabels: []string{"Transaction"},
		SplitConfig: SplitConfig{
			TestFraction:    0.2,
			ValidationFolds: 5,
		},
	}

	result, err := s.ToCypher(p, "my_graph", "fraud_model")
	if err != nil {
		t.Fatalf("ToCypher failed: %v", err)
	}

	// Check for pipeline creation
	if !strings.Contains(result, "pipeline.create('fraud_detection')") {
		t.Errorf("expected pipeline creation, got: %s", result)
	}

	// Check for feature steps
	if !strings.Contains(result, "fastRP") {
		t.Errorf("expected fastRP step, got: %s", result)
	}
	if !strings.Contains(result, "pageRank") {
		t.Errorf("expected pageRank step, got: %s", result)
	}
	if !strings.Contains(result, "embeddingDimension: 128") {
		t.Errorf("expected embeddingDimension, got: %s", result)
	}

	// Check for model
	if !strings.Contains(result, "LogisticRegression") {
		t.Errorf("expected LogisticRegression model, got: %s", result)
	}

	// Check for split config
	if !strings.Contains(result, "testFraction: 0.2") {
		t.Errorf("expected testFraction, got: %s", result)
	}

	// Check for train command
	if !strings.Contains(result, "pipeline.train") {
		t.Errorf("expected train command, got: %s", result)
	}
	if !strings.Contains(result, "targetProperty: 'isFraud'") {
		t.Errorf("expected targetProperty, got: %s", result)
	}
}

func TestPipelineSerializer_ToCypher_LinkPrediction(t *testing.T) {
	s := NewPipelineSerializer()
	p := &LinkPredictionPipeline{
		BasePipeline: BasePipeline{
			Name: "link_predictor",
			Models: []Model{
				&RandomForest{NumTrees: 100, MaxDepth: 10},
			},
		},
		TargetRelationshipType: "FOLLOWS",
		NegativeSamplingRatio:  1.5,
	}

	result, err := s.ToCypher(p, "social_graph", "follow_model")
	if err != nil {
		t.Fatalf("ToCypher failed: %v", err)
	}

	if !strings.Contains(result, "linkPrediction") {
		t.Errorf("expected linkPrediction pipeline type, got: %s", result)
	}
	if !strings.Contains(result, "RandomForest") {
		t.Errorf("expected RandomForest model, got: %s", result)
	}
	if !strings.Contains(result, "negativeSamplingRatio: 1.5") {
		t.Errorf("expected negativeSamplingRatio, got: %s", result)
	}
}

func TestPipelineSerializer_ToCypher_NodeRegression(t *testing.T) {
	s := NewPipelineSerializer()
	p := &NodeRegressionPipeline{
		BasePipeline: BasePipeline{
			Name: "price_predictor",
			Models: []Model{
				&LinearRegression{Penalty: 0.01},
			},
		},
		TargetProperty: "price",
	}

	result, err := s.ToCypher(p, "product_graph", "price_model")
	if err != nil {
		t.Fatalf("ToCypher failed: %v", err)
	}

	if !strings.Contains(result, "nodeRegression") {
		t.Errorf("expected nodeRegression pipeline type, got: %s", result)
	}
	if !strings.Contains(result, "LinearRegression") {
		t.Errorf("expected LinearRegression model, got: %s", result)
	}
}

func TestPipelineSerializer_ToJSON(t *testing.T) {
	s := NewPipelineSerializer()
	p := &NodeClassificationPipeline{
		BasePipeline: BasePipeline{
			Name: "fraud_detection",
			FeatureSteps: []FeatureStep{
				&FastRPStep{Property: "embedding", EmbeddingDimension: 128},
			},
			Models: []Model{
				&LogisticRegression{Penalty: 0.001},
			},
		},
		TargetProperty:   "isFraud",
		TargetNodeLabels: []string{"Transaction"},
	}

	result, err := s.ToJSON(p)
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}

	if parsed["name"] != "fraud_detection" {
		t.Errorf("name = %v, want fraud_detection", parsed["name"])
	}
	if parsed["pipelineType"] != "NodeClassification" {
		t.Errorf("pipelineType = %v, want NodeClassification", parsed["pipelineType"])
	}
	if parsed["targetProperty"] != "isFraud" {
		t.Errorf("targetProperty = %v, want isFraud", parsed["targetProperty"])
	}

	featureSteps, ok := parsed["featureSteps"].([]any)
	if !ok || len(featureSteps) != 1 {
		t.Error("expected 1 feature step")
	}

	models, ok := parsed["models"].([]any)
	if !ok || len(models) != 1 {
		t.Error("expected 1 model")
	}
}

func TestPipelineSerializer_ToMap(t *testing.T) {
	s := NewPipelineSerializer()
	p := &LinkPredictionPipeline{
		BasePipeline: BasePipeline{
			Name: "link_predictor",
		},
		TargetRelationshipType: "KNOWS",
		SourceNodeLabels:       []string{"Person"},
		TargetNodeLabels:       []string{"Person"},
	}

	result := s.ToMap(p)

	if result["name"] != "link_predictor" {
		t.Errorf("name = %v, want link_predictor", result["name"])
	}
	if result["targetRelationshipType"] != "KNOWS" {
		t.Errorf("targetRelationshipType = %v, want KNOWS", result["targetRelationshipType"])
	}
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  string
	}{
		{"string", "hello", "'hello'"},
		{"int", 42, "42"},
		{"float", 3.14, "3.14"},
		{"bool true", true, "true"},
		{"bool false", false, "false"},
		{"string slice", []string{"a", "b"}, "['a', 'b']"},
		{"int slice", []int{1, 2, 3}, "[1, 2, 3]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatValue(tt.input); got != tt.want {
				t.Errorf("formatValue(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"MaxEpochs", "maxEpochs"},
		{"Penalty", "penalty"},
		{"Name", "name"},
		{"EmbeddingDimension", "embeddingDimension"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := toCamelCase(tt.input); got != tt.want {
				t.Errorf("toCamelCase(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestPipeline_ImplementsInterface(t *testing.T) {
	// Verify all pipelines implement the Pipeline interface
	var _ Pipeline = &NodeClassificationPipeline{}
	var _ Pipeline = &LinkPredictionPipeline{}
	var _ Pipeline = &NodeRegressionPipeline{}
}

func TestFeatureStep_ImplementsInterface(t *testing.T) {
	// Verify all steps implement the FeatureStep interface
	var _ FeatureStep = &FastRPStep{}
	var _ FeatureStep = &PageRankStep{}
	var _ FeatureStep = &DegreeStep{}
	var _ FeatureStep = &Node2VecStep{}
	var _ FeatureStep = &ScalerStep{}
}

func TestModel_ImplementsInterface(t *testing.T) {
	// Verify all models implement the Model interface
	var _ Model = &LogisticRegression{}
	var _ Model = &RandomForest{}
	var _ Model = &MLP{}
	var _ Model = &LinearRegression{}
}
