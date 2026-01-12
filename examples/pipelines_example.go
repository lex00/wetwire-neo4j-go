package examples

import (
	"github.com/lex00/wetwire-neo4j-go/internal/pipelines"
)

// NodeClassificationExample demonstrates a node classification ML pipeline.
// Reference: https://neo4j.com/docs/graph-data-science/current/machine-learning/node-property-prediction/nodeclassification-pipelines/
var NodeClassificationExample = &pipelines.NodeClassificationPipeline{
	BasePipeline: pipelines.BasePipeline{
		Name: "fraud-detection-pipeline",
		FeatureSteps: []pipelines.FeatureStep{
			&pipelines.FastRPStep{
				Property:           "embedding",
				EmbeddingDimension: 128,
			},
			&pipelines.PageRankStep{
				Property:      "pageRank",
				DampingFactor: 0.85,
			},
			&pipelines.DegreeStep{
				Property: "degree",
			},
		},
		Models: []pipelines.Model{
			&pipelines.LogisticRegression{
				Penalty:      0.1,
				MaxEpochs:    100,
				LearningRate: 0.001,
				Tolerance:    0.0001,
			},
			&pipelines.RandomForest{
				NumTrees:     100,
				MaxDepth:     10,
				MinSplitSize: 2,
			},
		},
	},
	SplitConfig: pipelines.SplitConfig{
		TestFraction:    0.2,
		ValidationFolds: 5,
	},
	TargetProperty:   "isFraud",
	TargetNodeLabels: []string{"Account"},
}

// LinkPredictionExample demonstrates a link prediction ML pipeline.
// Reference: https://neo4j.com/docs/graph-data-science/current/machine-learning/linkprediction-pipelines/
var LinkPredictionExample = &pipelines.LinkPredictionPipeline{
	BasePipeline: pipelines.BasePipeline{
		Name: "friendship-prediction-pipeline",
		FeatureSteps: []pipelines.FeatureStep{
			&pipelines.Node2VecStep{
				Property:           "node2vec",
				EmbeddingDimension: 64,
			},
			&pipelines.DegreeStep{
				Property: "degree",
			},
		},
		Models: []pipelines.Model{
			&pipelines.LogisticRegression{
				Penalty:   0.001,
				MaxEpochs: 500,
			},
			&pipelines.MLP{
				HiddenLayers: []int{64, 32},
				LearningRate: 0.001,
				MaxEpochs:    500,
				Patience:     10,
			},
		},
	},
	SplitConfig: pipelines.SplitConfig{
		TestFraction:    0.1,
		ValidationFolds: 3,
	},
	TargetRelationshipType: "FRIENDS_WITH",
}

// NodeRegressionExample demonstrates a node regression ML pipeline.
// Reference: https://neo4j.com/docs/graph-data-science/current/machine-learning/node-property-prediction/noderegression-pipelines/
var NodeRegressionExample = &pipelines.NodeRegressionPipeline{
	BasePipeline: pipelines.BasePipeline{
		Name: "house-price-prediction-pipeline",
		FeatureSteps: []pipelines.FeatureStep{
			&pipelines.FastRPStep{
				Property:           "locationEmbedding",
				EmbeddingDimension: 256,
			},
			&pipelines.ScalerStep{
				Property:   "sqft",
				ScalerType: "MinMax",
			},
		},
		Models: []pipelines.Model{
			&pipelines.LinearRegression{},
			&pipelines.RandomForest{
				NumTrees: 50,
				MaxDepth: 15,
			},
		},
	},
	SplitConfig: pipelines.SplitConfig{
		TestFraction:    0.2,
		ValidationFolds: 5,
	},
	TargetProperty:   "price",
	TargetNodeLabels: []string{"House"},
}

// MLPClassifierExample demonstrates MLP usage for complex classification.
var MLPClassifierExample = &pipelines.NodeClassificationPipeline{
	BasePipeline: pipelines.BasePipeline{
		Name: "document-classification-pipeline",
		FeatureSteps: []pipelines.FeatureStep{
			&pipelines.FastRPStep{
				Property:           "docEmbedding",
				EmbeddingDimension: 512,
			},
		},
		Models: []pipelines.Model{
			&pipelines.MLP{
				HiddenLayers: []int{256, 128, 64},
				LearningRate: 0.0005,
				MaxEpochs:    1000,
				Patience:     20,
				MinEpochs:    50,
				Tolerance:    0.0001,
			},
		},
	},
	SplitConfig: pipelines.SplitConfig{
		TestFraction:    0.15,
		ValidationFolds: 3,
	},
	TargetProperty:   "category",
	TargetNodeLabels: []string{"Document"},
}

// AllPipelineExamples returns all example pipeline configurations.
func AllPipelineExamples() []pipelines.Pipeline {
	return []pipelines.Pipeline{
		NodeClassificationExample,
		LinkPredictionExample,
		NodeRegressionExample,
		MLPClassifierExample,
	}
}
