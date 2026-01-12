// Package algorithms provides typed configurations for Neo4j Graph Data Science algorithms.
//
// This package implements type-safe configurations for GDS algorithms including:
// - Centrality: PageRank, Betweenness, Degree, Closeness, ArticleRank
// - Community: Louvain, Leiden, Label Propagation, WCC, K-Core, Triangle Count
// - Similarity: Node Similarity, K-Nearest Neighbors
// - Path Finding: Dijkstra, A*, BFS, DFS
// - Embeddings: FastRP, GraphSAGE, Node2Vec, HashGNN
//
// Example usage:
//
//	pr := &algorithms.PageRank{
//		Name:          "influence_score",
//		GraphName:     "social_graph",
//		DampingFactor: 0.85,
//		MaxIterations: 50,
//	}
//	cypher := algorithms.ToCypher(pr)
package algorithms

// Category represents the algorithm category.
type Category string

const (
	Centrality     Category = "Centrality"
	Community      Category = "Community"
	Similarity     Category = "Similarity"
	PathFinding    Category = "PathFinding"
	Embeddings     Category = "Embeddings"
	LinkPrediction Category = "LinkPrediction"
)

// Mode represents the algorithm execution mode.
type Mode string

const (
	// Stream returns results as a stream without persisting.
	Stream Mode = "stream"
	// Stats returns aggregate statistics.
	Stats Mode = "stats"
	// Mutate adds results to the in-memory graph projection.
	Mutate Mode = "mutate"
	// Write writes results back to the database.
	Write Mode = "write"
)

// Algorithm is the interface that all GDS algorithm configurations implement.
type Algorithm interface {
	// AlgorithmName returns the name of this algorithm configuration.
	AlgorithmName() string
	// AlgorithmType returns the GDS algorithm type (e.g., "gds.pageRank").
	AlgorithmType() string
	// AlgorithmCategory returns the category of this algorithm.
	AlgorithmCategory() Category
	// GetGraphName returns the graph projection name.
	GetGraphName() string
	// GetMode returns the execution mode.
	GetMode() Mode
}

// BaseAlgorithm contains common algorithm configuration fields.
type BaseAlgorithm struct {
	// Name is the configuration name.
	Name string
	// GraphName is the name of the graph projection to use.
	GraphName string
	// Mode is the execution mode (stream, stats, mutate, write).
	Mode Mode
	// Concurrency is the number of concurrent threads (default: 4).
	Concurrency int
	// NodeLabels filters which nodes to include.
	NodeLabels []string
	// RelationshipTypes filters which relationships to include.
	RelationshipTypes []string
}

// AlgorithmName returns the configuration name.
func (b *BaseAlgorithm) AlgorithmName() string {
	return b.Name
}

// GetGraphName returns the graph projection name.
func (b *BaseAlgorithm) GetGraphName() string {
	return b.GraphName
}

// GetMode returns the execution mode.
func (b *BaseAlgorithm) GetMode() Mode {
	if b.Mode == "" {
		return Stream
	}
	return b.Mode
}

// PageRank computes the PageRank centrality score.
type PageRank struct {
	BaseAlgorithm
	// DampingFactor is the probability of following an outgoing relationship (default: 0.85).
	DampingFactor float64
	// MaxIterations is the maximum number of iterations (default: 20).
	MaxIterations int
	// Tolerance is the minimum change required for convergence (default: 0.0000001).
	Tolerance float64
	// RelationshipWeightProperty is the property to use for weighted PageRank.
	RelationshipWeightProperty string
	// WriteProperty is the node property to write results to (for write mode).
	WriteProperty string
	// MutateProperty is the node property to mutate (for mutate mode).
	MutateProperty string
}

func (p *PageRank) AlgorithmType() string       { return "gds.pageRank" }
func (p *PageRank) AlgorithmCategory() Category { return Centrality }

// ArticleRank is a variant of PageRank that reduces bias from low-degree nodes.
type ArticleRank struct {
	BaseAlgorithm
	DampingFactor              float64
	MaxIterations              int
	Tolerance                  float64
	RelationshipWeightProperty string
	WriteProperty              string
	MutateProperty             string
}

func (a *ArticleRank) AlgorithmType() string       { return "gds.articleRank" }
func (a *ArticleRank) AlgorithmCategory() Category { return Centrality }

// Betweenness computes the betweenness centrality score.
type Betweenness struct {
	BaseAlgorithm
	// SamplingSize is the number of source nodes to sample (0 = all nodes).
	SamplingSize int
	// SamplingSeed is the random seed for sampling.
	SamplingSeed   int64
	WriteProperty  string
	MutateProperty string
}

func (b *Betweenness) AlgorithmType() string       { return "gds.betweenness" }
func (b *Betweenness) AlgorithmCategory() Category { return Centrality }

// Degree computes the degree centrality (number of relationships).
type Degree struct {
	BaseAlgorithm
	// Orientation is the relationship direction: NATURAL, REVERSE, or UNDIRECTED.
	Orientation                string
	RelationshipWeightProperty string
	WriteProperty              string
	MutateProperty             string
}

func (d *Degree) AlgorithmType() string       { return "gds.degree" }
func (d *Degree) AlgorithmCategory() Category { return Centrality }

// Closeness computes the closeness centrality score.
type Closeness struct {
	BaseAlgorithm
	// UseWassermanFaust enables Wasserman-Faust formula for disconnected graphs.
	UseWassermanFaust bool
	WriteProperty     string
	MutateProperty    string
}

func (c *Closeness) AlgorithmType() string       { return "gds.closeness" }
func (c *Closeness) AlgorithmCategory() Category { return Centrality }

// Louvain detects communities using the Louvain algorithm.
type Louvain struct {
	BaseAlgorithm
	// MaxLevels is the maximum number of hierarchy levels (default: 10).
	MaxLevels int
	// MaxIterations is the max iterations per level (default: 10).
	MaxIterations int
	// Tolerance is the minimum change for convergence (default: 0.0001).
	Tolerance float64
	// IncludeIntermediateCommunities includes all hierarchy levels.
	IncludeIntermediateCommunities bool
	// SeedProperty is the property for initial community assignments.
	SeedProperty               string
	RelationshipWeightProperty string
	WriteProperty              string
	MutateProperty             string
}

func (l *Louvain) AlgorithmType() string       { return "gds.louvain" }
func (l *Louvain) AlgorithmCategory() Category { return Community }

// Leiden is an improved version of Louvain for community detection.
type Leiden struct {
	BaseAlgorithm
	MaxLevels                      int
	Gamma                          float64 // Resolution parameter (default: 1.0)
	Theta                          float64 // Randomness parameter (default: 0.01)
	Tolerance                      float64
	IncludeIntermediateCommunities bool
	RandomSeed                     int64
	RelationshipWeightProperty     string
	WriteProperty                  string
	MutateProperty                 string
}

func (l *Leiden) AlgorithmType() string       { return "gds.leiden" }
func (l *Leiden) AlgorithmCategory() Category { return Community }

// LabelPropagation detects communities using label propagation.
type LabelPropagation struct {
	BaseAlgorithm
	MaxIterations              int
	SeedProperty               string
	RelationshipWeightProperty string
	WriteProperty              string
	MutateProperty             string
}

func (l *LabelPropagation) AlgorithmType() string       { return "gds.labelPropagation" }
func (l *LabelPropagation) AlgorithmCategory() Category { return Community }

// WCC finds weakly connected components.
type WCC struct {
	BaseAlgorithm
	// SeedProperty is the property for initial component assignments.
	SeedProperty               string
	RelationshipWeightProperty string
	// Threshold is the minimum weight for relationships.
	Threshold      float64
	WriteProperty  string
	MutateProperty string
}

func (w *WCC) AlgorithmType() string       { return "gds.wcc" }
func (w *WCC) AlgorithmCategory() Category { return Community }

// TriangleCount counts triangles for each node.
type TriangleCount struct {
	BaseAlgorithm
	// MaxDegree filters out high-degree nodes (0 = no limit).
	MaxDegree      int
	WriteProperty  string
	MutateProperty string
}

func (t *TriangleCount) AlgorithmType() string       { return "gds.triangleCount" }
func (t *TriangleCount) AlgorithmCategory() Category { return Community }

// KCore finds k-core subgraphs.
type KCore struct {
	BaseAlgorithm
	// K is the minimum degree for k-core.
	K              int
	WriteProperty  string
	MutateProperty string
}

func (k *KCore) AlgorithmType() string       { return "gds.kcore" }
func (k *KCore) AlgorithmCategory() Category { return Community }

// NodeSimilarity computes similarity between nodes based on neighbors.
type NodeSimilarity struct {
	BaseAlgorithm
	// SimilarityCutoff is the minimum similarity to report (default: 0.0).
	SimilarityCutoff float64
	// DegreeCutoff is the minimum degree to include (default: 1).
	DegreeCutoff int
	// TopK is the number of similar nodes to return per node (default: 10).
	TopK int
	// TopN is the total number of results to return (0 = unlimited).
	TopN int
	// SimilarityMetric is JACCARD (default), OVERLAP, or COSINE.
	SimilarityMetric string
	// WriteRelationshipType is the relationship type to write.
	WriteRelationshipType string
	// WriteProperty is the property to write similarity to.
	WriteProperty string
}

func (n *NodeSimilarity) AlgorithmType() string       { return "gds.nodeSimilarity" }
func (n *NodeSimilarity) AlgorithmCategory() Category { return Similarity }

// KNN finds k-nearest neighbors based on node properties.
type KNN struct {
	BaseAlgorithm
	// K is the number of neighbors to find (default: 10).
	K int
	// SimilarityCutoff is the minimum similarity to report.
	SimilarityCutoff float64
	// SampleRate is the fraction of nodes to compare (default: 0.5).
	SampleRate float64
	// DeltaThreshold is convergence threshold (default: 0.001).
	DeltaThreshold float64
	// MaxIterations is the maximum number of iterations (default: 100).
	MaxIterations int
	// RandomJoins is the number of random neighbor candidates (default: 10).
	RandomJoins int
	// NodeProperties are the properties to use for similarity.
	NodeProperties []string
	// WriteRelationshipType is the relationship type to write.
	WriteRelationshipType string
	WriteProperty         string
}

func (k *KNN) AlgorithmType() string       { return "gds.knn" }
func (k *KNN) AlgorithmCategory() Category { return Similarity }

// FastRP generates node embeddings using Fast Random Projection.
type FastRP struct {
	BaseAlgorithm
	// EmbeddingDimension is the size of embeddings (default: 128).
	EmbeddingDimension int
	// IterationWeights are weights for each iteration (default: [0.0, 1.0, 1.0]).
	IterationWeights []float64
	// NormalizationStrength adjusts embedding normalization (default: 0.0).
	NormalizationStrength float64
	// PropertyRatio is the ratio of property features (default: 0.0).
	PropertyRatio float64
	// NodeSelfInfluence includes node's own properties (default: 0.0).
	NodeSelfInfluence float64
	// FeatureProperties are node properties to include.
	FeatureProperties          []string
	RelationshipWeightProperty string
	WriteProperty              string
	MutateProperty             string
}

func (f *FastRP) AlgorithmType() string       { return "gds.fastRP" }
func (f *FastRP) AlgorithmCategory() Category { return Embeddings }

// Node2Vec generates embeddings using random walks.
type Node2Vec struct {
	BaseAlgorithm
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
	// WindowSize is the context window size (default: 10).
	WindowSize int
	// NegativeSamplingRate is negative samples per positive (default: 5).
	NegativeSamplingRate int
	// PositiveSamplingFactor is subsampling of frequent nodes (default: 0.001).
	PositiveSamplingFactor float64
	// Iterations is the number of training iterations (default: 1).
	Iterations                 int
	RelationshipWeightProperty string
	WriteProperty              string
	MutateProperty             string
}

func (n *Node2Vec) AlgorithmType() string       { return "gds.node2vec" }
func (n *Node2Vec) AlgorithmCategory() Category { return Embeddings }

// GraphSAGE generates embeddings using GraphSAGE neural network.
type GraphSAGE struct {
	BaseAlgorithm
	// EmbeddingDimension is the size of embeddings (default: 64).
	EmbeddingDimension int
	// AggregatorType is MEAN (default), POOL, or LSTM.
	AggregatorType string
	// Activations are activation functions per layer.
	Activations []string
	// SampleSizes are neighbor sample sizes per layer (default: [25, 10]).
	SampleSizes []int
	// FeatureProperties are node properties to use as features.
	FeatureProperties []string
	// Epochs is the number of training epochs (default: 1).
	Epochs int
	// LearningRate is the optimizer learning rate (default: 0.1).
	LearningRate float64
	// BatchSize is the training batch size (default: 100).
	BatchSize int
	// Tolerance is the convergence threshold (default: 0.0001).
	Tolerance                  float64
	RelationshipWeightProperty string
	ModelName                  string
	WriteProperty              string
	MutateProperty             string
}

func (g *GraphSAGE) AlgorithmType() string       { return "gds.beta.graphSage" }
func (g *GraphSAGE) AlgorithmCategory() Category { return Embeddings }

// HashGNN generates embeddings using HashGNN.
type HashGNN struct {
	BaseAlgorithm
	// EmbeddingDensity is the density of embeddings (default: 64).
	EmbeddingDensity int
	// Iterations is the number of message passing iterations (default: 3).
	Iterations int
	// NeighborInfluence weights neighbor features (default: 1.0).
	NeighborInfluence float64
	// FeatureProperties are node properties to use.
	FeatureProperties          []string
	RelationshipWeightProperty string
	WriteProperty              string
	MutateProperty             string
}

func (h *HashGNN) AlgorithmType() string       { return "gds.hashgnn" }
func (h *HashGNN) AlgorithmCategory() Category { return Embeddings }

// Dijkstra finds shortest paths from a source node.
type Dijkstra struct {
	BaseAlgorithm
	// SourceNode is the starting node ID or property value.
	SourceNode any
	// TargetNode is the optional target node (nil = all nodes).
	TargetNode                 any
	RelationshipWeightProperty string
	WriteProperty              string
}

func (d *Dijkstra) AlgorithmType() string       { return "gds.shortestPath.dijkstra" }
func (d *Dijkstra) AlgorithmCategory() Category { return PathFinding }

// AStar finds shortest path using A* algorithm with heuristic.
type AStar struct {
	BaseAlgorithm
	SourceNode                 any
	TargetNode                 any
	RelationshipWeightProperty string
	// LatitudeProperty is the node property for latitude.
	LatitudeProperty string
	// LongitudeProperty is the node property for longitude.
	LongitudeProperty string
	WriteProperty     string
}

func (a *AStar) AlgorithmType() string       { return "gds.shortestPath.astar" }
func (a *AStar) AlgorithmCategory() Category { return PathFinding }

// BFS performs breadth-first search traversal.
type BFS struct {
	BaseAlgorithm
	SourceNode any
	// TargetNodes are optional target nodes.
	TargetNodes []any
	// MaxDepth is the maximum depth to traverse (0 = unlimited).
	MaxDepth int
}

func (b *BFS) AlgorithmType() string       { return "gds.bfs" }
func (b *BFS) AlgorithmCategory() Category { return PathFinding }

// DFS performs depth-first search traversal.
type DFS struct {
	BaseAlgorithm
	SourceNode  any
	TargetNodes []any
	MaxDepth    int
}

func (d *DFS) AlgorithmType() string       { return "gds.dfs" }
func (d *DFS) AlgorithmCategory() Category { return PathFinding }
