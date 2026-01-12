package examples

import (
	"testing"

	"github.com/lex00/wetwire-neo4j-go/internal/algorithms"
	"github.com/lex00/wetwire-neo4j-go/internal/kg"
	"github.com/lex00/wetwire-neo4j-go/internal/lint"
	"github.com/lex00/wetwire-neo4j-go/internal/pipelines"
	"github.com/lex00/wetwire-neo4j-go/internal/projections"
	"github.com/lex00/wetwire-neo4j-go/internal/retrievers"
	"github.com/lex00/wetwire-neo4j-go/internal/serializer"
)

// TestSchemaExamples validates all schema examples.
func TestSchemaExamples(t *testing.T) {
	linter := lint.NewLinter()
	cypherSer := serializer.NewCypherSerializer()
	jsonSer := serializer.NewJSONSerializer()

	for _, node := range AllNodeTypes() {
		t.Run("NodeType_"+node.Label, func(t *testing.T) {
			// Lint the node type
			results := linter.LintNodeType(node)
			if lint.HasErrors(results) {
				t.Errorf("lint errors for %s: %s", node.Label, lint.FormatResults(results))
			}

			// Serialize to Cypher
			cypher, err := cypherSer.SerializeNodeType(node)
			if err != nil {
				t.Errorf("failed to serialize %s to Cypher: %v", node.Label, err)
			}
			if cypher == "" {
				t.Logf("warning: no Cypher generated for %s (may have no constraints/indexes)", node.Label)
			}

			// Serialize to JSON
			m := jsonSer.NodeTypeToMap(node)
			if m == nil {
				t.Errorf("failed to serialize %s to map", node.Label)
			}
		})
	}

	for _, rel := range AllRelationshipTypes() {
		t.Run("RelationshipType_"+rel.Label, func(t *testing.T) {
			// Lint the relationship type
			results := linter.LintRelationshipType(rel)
			if lint.HasErrors(results) {
				t.Errorf("lint errors for %s: %s", rel.Label, lint.FormatResults(results))
			}

			// Serialize to Cypher
			cypher, err := cypherSer.SerializeRelationshipType(rel)
			if err != nil {
				t.Errorf("failed to serialize %s to Cypher: %v", rel.Label, err)
			}
			_ = cypher // May be empty if no constraints

			// Serialize to JSON
			m := jsonSer.RelationshipTypeToMap(rel)
			if m == nil {
				t.Errorf("failed to serialize %s to map", rel.Label)
			}
		})
	}
}

// TestAlgorithmExamples validates all algorithm examples.
func TestAlgorithmExamples(t *testing.T) {
	linter := lint.NewLinter()
	ser := algorithms.NewAlgorithmSerializer()

	for _, algo := range AllAlgorithmExamples() {
		t.Run("Algorithm_"+algo.AlgorithmType(), func(t *testing.T) {
			// Lint the algorithm
			results := linter.LintAlgorithm(algo)
			if lint.HasErrors(results) {
				t.Errorf("lint errors for %s: %s", algo.AlgorithmType(), lint.FormatResults(results))
			}

			// Serialize to Cypher
			cypher, err := ser.ToCypher(algo)
			if err != nil {
				t.Errorf("failed to serialize %s to Cypher: %v", algo.AlgorithmType(), err)
			}
			if cypher == "" {
				t.Errorf("empty Cypher generated for %s", algo.AlgorithmType())
			}

			// Serialize to map
			m := ser.ToMap(algo)
			if m == nil {
				t.Errorf("failed to serialize %s to map", algo.AlgorithmType())
			}
		})
	}
}

// TestPipelineExamples validates all pipeline examples.
func TestPipelineExamples(t *testing.T) {
	linter := lint.NewLinter()
	ser := pipelines.NewPipelineSerializer()

	for _, pipe := range AllPipelineExamples() {
		t.Run("Pipeline_"+pipe.PipelineName(), func(t *testing.T) {
			// Lint the pipeline
			results := linter.LintPipeline(pipe)
			if lint.HasErrors(results) {
				t.Errorf("lint errors for %s: %s", pipe.PipelineName(), lint.FormatResults(results))
			}

			// Serialize to Cypher
			cypher, err := ser.ToCypher(pipe, "test-graph", "test-model")
			if err != nil {
				t.Errorf("failed to serialize %s to Cypher: %v", pipe.PipelineName(), err)
			}
			if cypher == "" {
				t.Errorf("empty Cypher generated for %s", pipe.PipelineName())
			}

			// Serialize to map
			m := ser.ToMap(pipe)
			if m == nil {
				t.Errorf("failed to serialize %s to map", pipe.PipelineName())
			}
		})
	}
}

// TestProjectionExamples validates all projection examples.
func TestProjectionExamples(t *testing.T) {
	ser := projections.NewProjectionSerializer()

	for _, proj := range AllProjectionExamples() {
		t.Run("Projection_"+proj.ProjectionName(), func(t *testing.T) {
			// Serialize to Cypher
			cypher, err := ser.ToCypher(proj)
			if err != nil {
				t.Errorf("failed to serialize %s to Cypher: %v", proj.ProjectionName(), err)
			}
			if cypher == "" {
				t.Errorf("empty Cypher generated for %s", proj.ProjectionName())
			}

			// Serialize to map
			m := ser.ToMap(proj)
			if m == nil {
				t.Errorf("failed to serialize %s to map", proj.ProjectionName())
			}
		})
	}
}

// TestRetrieverExamples validates all retriever examples.
func TestRetrieverExamples(t *testing.T) {
	ser := retrievers.NewRetrieverSerializer()

	for _, ret := range AllRetrieverExamples() {
		t.Run("Retriever_"+string(ret.RetrieverType()), func(t *testing.T) {
			// Serialize to JSON
			json, err := ser.ToJSON(ret)
			if err != nil {
				t.Errorf("failed to serialize %s to JSON: %v", ret.RetrieverType(), err)
			}
			if len(json) == 0 {
				t.Errorf("empty JSON generated for %s", ret.RetrieverType())
			}

			// Serialize to map
			m := ser.ToMap(ret)
			if m == nil {
				t.Errorf("failed to serialize %s to map", ret.RetrieverType())
			}
		})
	}
}

// TestKGPipelineExamples validates all KG pipeline examples.
func TestKGPipelineExamples(t *testing.T) {
	linter := lint.NewLinter()
	ser := kg.NewKGSerializer()

	for _, pipe := range AllKGPipelineExamples() {
		t.Run("KGPipeline_"+pipe.PipelineName(), func(t *testing.T) {
			// Lint the pipeline
			results := linter.LintKGPipeline(pipe)
			if lint.HasErrors(results) {
				t.Errorf("lint errors for %s: %s", pipe.PipelineName(), lint.FormatResults(results))
			}

			// Serialize to JSON
			json, err := ser.ToJSON(pipe)
			if err != nil {
				t.Errorf("failed to serialize %s to JSON: %v", pipe.PipelineName(), err)
			}
			if len(json) == 0 {
				t.Errorf("empty JSON generated for %s", pipe.PipelineName())
			}

			// Serialize to map
			m := ser.ToMap(pipe)
			if m == nil {
				t.Errorf("failed to serialize %s to map", pipe.PipelineName())
			}
		})
	}
}

// TestSchemaExamples_AllCanBeSerialized verifies all examples can be serialized together.
func TestSchemaExamples_AllCanBeSerialized(t *testing.T) {
	cypherSer := serializer.NewCypherSerializer()

	cypher, err := cypherSer.SerializeAll(AllNodeTypes(), AllRelationshipTypes())
	if err != nil {
		t.Fatalf("failed to serialize all schema: %v", err)
	}

	if cypher == "" {
		t.Error("empty Cypher generated for all schemas")
	}

	t.Logf("Generated Cypher length: %d bytes", len(cypher))
}

// TestExamplesCount ensures we have a good number of examples.
func TestExamplesCount(t *testing.T) {
	tests := []struct {
		name     string
		count    int
		minCount int
	}{
		{"NodeTypes", len(AllNodeTypes()), 4},
		{"RelationshipTypes", len(AllRelationshipTypes()), 3},
		{"Algorithms", len(AllAlgorithmExamples()), 10},
		{"Pipelines", len(AllPipelineExamples()), 3},
		{"Projections", len(AllProjectionExamples()), 4},
		{"Retrievers", len(AllRetrieverExamples()), 5},
		{"KGPipelines", len(AllKGPipelineExamples()), 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.count < tt.minCount {
				t.Errorf("expected at least %d %s, got %d", tt.minCount, tt.name, tt.count)
			}
		})
	}
}
