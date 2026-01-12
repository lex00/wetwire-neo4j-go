// Package pipelines provides typed configurations for Neo4j GDS machine learning pipelines.
//
// This package implements type-safe configurations for ML pipelines including:
// - NodeClassificationPipeline: Predict node labels
// - LinkPredictionPipeline: Predict future relationships
// - NodeRegressionPipeline: Predict numeric properties
//
// Example usage:
//
//	pipeline := &pipelines.NodeClassificationPipeline{
//		Name:           "fraud_detection",
//		TargetProperty: "isFraud",
//		NodeLabels:     []string{"Transaction"},
//	}
//	pipeline.AddFeatureStep(&pipelines.FastRPStep{EmbeddingDimension: 128})
//	pipeline.AddModel(&pipelines.LogisticRegression{Penalty: 0.001})
package pipelines

// PipelineType represents the type of ML pipeline.
type PipelineType string

const (
	// NodeClassification predicts categorical node labels.
	NodeClassification PipelineType = "NodeClassification"
	// LinkPrediction predicts future relationships.
	LinkPrediction PipelineType = "LinkPrediction"
	// NodeRegression predicts numeric node properties.
	NodeRegression PipelineType = "NodeRegression"
)

// Pipeline is the interface that all ML pipeline configurations implement.
type Pipeline interface {
	// PipelineName returns the name of this pipeline.
	PipelineName() string
	// PipelineType returns the type of pipeline.
	PipelineType() PipelineType
	// GetFeatureSteps returns the feature engineering steps.
	GetFeatureSteps() []FeatureStep
	// GetModels returns the model candidates.
	GetModels() []Model
}

// BasePipeline contains common pipeline configuration fields.
type BasePipeline struct {
	// Name is the pipeline name.
	Name string
	// NodeLabels filters which nodes to include in training.
	NodeLabels []string
	// RelationshipTypes filters which relationships to traverse.
	RelationshipTypes []string
	// FeatureSteps are the node property steps for feature engineering.
	FeatureSteps []FeatureStep
	// Models are the model candidates to train and evaluate.
	Models []Model
}

// PipelineName returns the pipeline name.
func (b *BasePipeline) PipelineName() string {
	return b.Name
}

// GetFeatureSteps returns the feature engineering steps.
func (b *BasePipeline) GetFeatureSteps() []FeatureStep {
	return b.FeatureSteps
}

// GetModels returns the model candidates.
func (b *BasePipeline) GetModels() []Model {
	return b.Models
}

// AddFeatureStep adds a feature engineering step.
func (b *BasePipeline) AddFeatureStep(step FeatureStep) {
	b.FeatureSteps = append(b.FeatureSteps, step)
}

// AddModel adds a model candidate.
func (b *BasePipeline) AddModel(model Model) {
	b.Models = append(b.Models, model)
}

// SplitConfig configures train/test splitting.
type SplitConfig struct {
	// TestFraction is the fraction of data for testing (default: 0.2).
	TestFraction float64
	// ValidationFolds is the number of cross-validation folds (default: 5).
	ValidationFolds int
	// RandomSeed for reproducibility.
	RandomSeed int64
}

// AutoTuningConfig configures hyperparameter tuning.
type AutoTuningConfig struct {
	// MaxTrials is the maximum number of trials (default: 10).
	MaxTrials int
	// Metric is the optimization metric (default: depends on pipeline type).
	Metric string
}

// NodeClassificationPipeline predicts categorical node labels.
type NodeClassificationPipeline struct {
	BasePipeline
	// TargetProperty is the property to predict.
	TargetProperty string
	// TargetNodeLabels are the node labels with target property.
	TargetNodeLabels []string
	// FeatureProperties are existing node properties to use as features.
	FeatureProperties []string
	// SplitConfig configures train/test splitting.
	SplitConfig SplitConfig
	// AutoTuning configures hyperparameter tuning.
	AutoTuning AutoTuningConfig
}

func (p *NodeClassificationPipeline) PipelineType() PipelineType { return NodeClassification }

// LinkPredictionPipeline predicts future relationships.
type LinkPredictionPipeline struct {
	BasePipeline
	// TargetRelationshipType is the relationship type to predict.
	TargetRelationshipType string
	// SourceNodeLabels are valid source nodes.
	SourceNodeLabels []string
	// TargetNodeLabels are valid target nodes.
	TargetNodeLabels []string
	// FeatureProperties are existing node properties to use as features.
	FeatureProperties []string
	// NegativeSamplingRatio controls negative edge sampling (default: 1.0).
	NegativeSamplingRatio float64
	// SplitConfig configures train/test splitting.
	SplitConfig SplitConfig
	// AutoTuning configures hyperparameter tuning.
	AutoTuning AutoTuningConfig
}

func (p *LinkPredictionPipeline) PipelineType() PipelineType { return LinkPrediction }

// NodeRegressionPipeline predicts numeric node properties.
type NodeRegressionPipeline struct {
	BasePipeline
	// TargetProperty is the numeric property to predict.
	TargetProperty string
	// TargetNodeLabels are the node labels with target property.
	TargetNodeLabels []string
	// FeatureProperties are existing node properties to use as features.
	FeatureProperties []string
	// SplitConfig configures train/test splitting.
	SplitConfig SplitConfig
	// AutoTuning configures hyperparameter tuning.
	AutoTuning AutoTuningConfig
}

func (p *NodeRegressionPipeline) PipelineType() PipelineType { return NodeRegression }

// FeatureStep is the interface for feature engineering steps.
type FeatureStep interface {
	// StepType returns the step type (e.g., "fastRP", "pageRank").
	StepType() string
	// MutateProperty returns the property name for storing results.
	MutateProperty() string
}

// FastRPStep adds FastRP embeddings as features.
type FastRPStep struct {
	// Property is the name for storing embeddings.
	Property string
	// EmbeddingDimension is the size of embeddings (default: 128).
	EmbeddingDimension int
	// IterationWeights are weights for each iteration.
	IterationWeights []float64
	// NormalizationStrength adjusts embedding normalization.
	NormalizationStrength float64
	// RelationshipWeightProperty for weighted projections.
	RelationshipWeightProperty string
}

func (s *FastRPStep) StepType() string       { return "fastRP" }
func (s *FastRPStep) MutateProperty() string { return s.Property }

// PageRankStep adds PageRank scores as features.
type PageRankStep struct {
	// Property is the name for storing scores.
	Property string
	// DampingFactor is the random walk probability (default: 0.85).
	DampingFactor float64
	// MaxIterations is the maximum iterations (default: 20).
	MaxIterations int
	// Tolerance is the convergence threshold.
	Tolerance float64
}

func (s *PageRankStep) StepType() string       { return "pageRank" }
func (s *PageRankStep) MutateProperty() string { return s.Property }

// DegreeStep adds degree centrality as a feature.
type DegreeStep struct {
	// Property is the name for storing degree.
	Property string
	// Orientation is NATURAL, REVERSE, or UNDIRECTED.
	Orientation string
}

func (s *DegreeStep) StepType() string       { return "degree" }
func (s *DegreeStep) MutateProperty() string { return s.Property }

// Node2VecStep adds Node2Vec embeddings as features.
type Node2VecStep struct {
	// Property is the name for storing embeddings.
	Property string
	// EmbeddingDimension is the size of embeddings (default: 128).
	EmbeddingDimension int
	// WalkLength is the length of random walks (default: 80).
	WalkLength int
	// WalksPerNode is the number of walks per node (default: 10).
	WalksPerNode int
	// InOutFactor (p) controls likelihood of returning (default: 1.0).
	InOutFactor float64
	// ReturnFactor (q) controls exploration vs exploitation (default: 1.0).
	ReturnFactor float64
}

func (s *Node2VecStep) StepType() string       { return "node2Vec" }
func (s *Node2VecStep) MutateProperty() string { return s.Property }

// ScalerStep normalizes feature values.
type ScalerStep struct {
	// Property is the property to scale.
	Property string
	// ScalerType is the type of scaler (e.g., "minmax", "standard").
	ScalerType string
}

func (s *ScalerStep) StepType() string       { return "scaler" }
func (s *ScalerStep) MutateProperty() string { return s.Property }

// Model is the interface for ML model candidates.
type Model interface {
	// ModelType returns the model type (e.g., "LogisticRegression").
	ModelType() string
	// SupportedPipelines returns which pipeline types this model supports.
	SupportedPipelines() []PipelineType
}

// LogisticRegression model for classification.
type LogisticRegression struct {
	// Penalty is the regularization parameter (default: 0.0).
	Penalty float64
	// MaxEpochs is the maximum training epochs (default: 100).
	MaxEpochs int
	// Tolerance is the convergence threshold (default: 0.001).
	Tolerance float64
	// MinEpochs is the minimum epochs before early stopping.
	MinEpochs int
	// Patience is the epochs to wait for improvement.
	Patience int
	// LearningRate is the optimizer learning rate (default: 0.001).
	LearningRate float64
	// BatchSize is the training batch size (default: 100).
	BatchSize int
}

func (m *LogisticRegression) ModelType() string { return "LogisticRegression" }
func (m *LogisticRegression) SupportedPipelines() []PipelineType {
	return []PipelineType{NodeClassification, LinkPrediction}
}

// RandomForest model for classification and regression.
type RandomForest struct {
	// MaxDepth is the maximum tree depth (default: unlimited).
	MaxDepth int
	// NumTrees is the number of trees (default: 100).
	NumTrees int
	// MinSplitSize is the minimum samples to split (default: 2).
	MinSplitSize int
	// MaxFeaturesRatio is the fraction of features per tree (default: 1.0).
	MaxFeaturesRatio float64
	// MinLeafSize is the minimum samples in leaf (default: 1).
	MinLeafSize int
	// SamplingRatio is the fraction of data per tree (default: 1.0).
	SamplingRatio float64
}

func (m *RandomForest) ModelType() string { return "RandomForest" }
func (m *RandomForest) SupportedPipelines() []PipelineType {
	return []PipelineType{NodeClassification, LinkPrediction, NodeRegression}
}

// MLP (Multi-Layer Perceptron) model.
type MLP struct {
	// HiddenLayers specifies hidden layer sizes (default: [100]).
	HiddenLayers []int
	// LearningRate is the optimizer learning rate (default: 0.001).
	LearningRate float64
	// MaxEpochs is the maximum training epochs (default: 100).
	MaxEpochs int
	// BatchSize is the training batch size (default: 100).
	BatchSize int
	// Tolerance is the convergence threshold (default: 0.001).
	Tolerance float64
	// Patience is the epochs to wait for improvement.
	Patience int
	// MinEpochs is the minimum epochs before early stopping.
	MinEpochs int
	// Penalty is the L2 regularization parameter (default: 0.0).
	Penalty float64
}

func (m *MLP) ModelType() string { return "MLP" }
func (m *MLP) SupportedPipelines() []PipelineType {
	return []PipelineType{NodeClassification, LinkPrediction}
}

// LinearRegression model for regression.
type LinearRegression struct {
	// Penalty is the regularization parameter (default: 0.0).
	Penalty float64
	// MaxEpochs is the maximum training epochs (default: 100).
	MaxEpochs int
	// Tolerance is the convergence threshold (default: 0.001).
	Tolerance float64
	// LearningRate is the optimizer learning rate (default: 0.001).
	LearningRate float64
	// BatchSize is the training batch size (default: 100).
	BatchSize int
}

func (m *LinearRegression) ModelType() string { return "LinearRegression" }
func (m *LinearRegression) SupportedPipelines() []PipelineType {
	return []PipelineType{NodeRegression}
}
