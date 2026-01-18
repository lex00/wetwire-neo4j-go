package examples

import (
	"github.com/lex00/wetwire-neo4j-go/internal/algorithms"
)

// PageRankExample demonstrates the PageRank centrality algorithm.
// Reference: https://neo4j.com/docs/graph-data-science/current/algorithms/page-rank/
var PageRankExample = &algorithms.PageRank{
	BaseAlgorithm: algorithms.BaseAlgorithm{
		GraphName: "social-network",
		Mode:      algorithms.Stream,
	},
	DampingFactor: 0.85,
	MaxIterations: 20,
	Tolerance:     0.0000001,
}

// ArticleRankExample demonstrates the ArticleRank algorithm.
// Reference: https://neo4j.com/docs/graph-data-science/current/algorithms/article-rank/
var ArticleRankExample = &algorithms.ArticleRank{
	BaseAlgorithm: algorithms.BaseAlgorithm{
		GraphName: "citation-network",
		Mode:      algorithms.Mutate,
	},
	DampingFactor:  0.85,
	MaxIterations:  20,
	MutateProperty: "articleRank",
}

// BetweennessExample demonstrates the Betweenness Centrality algorithm.
// Reference: https://neo4j.com/docs/graph-data-science/current/algorithms/betweenness-centrality/
var BetweennessExample = &algorithms.Betweenness{
	BaseAlgorithm: algorithms.BaseAlgorithm{
		GraphName: "transport-network",
		Mode:      algorithms.Stream,
	},
	SamplingSize: 1000,
}

// LouvainExample demonstrates community detection with Louvain.
// Reference: https://neo4j.com/docs/graph-data-science/current/algorithms/louvain/
var LouvainExample = &algorithms.Louvain{
	BaseAlgorithm: algorithms.BaseAlgorithm{
		GraphName: "social-network",
		Mode:      algorithms.Mutate,
	},
	MaxLevels:                      10,
	MaxIterations:                  10,
	Tolerance:                      0.0001,
	IncludeIntermediateCommunities: true,
	MutateProperty:                 "communityId",
}

// LeidenExample demonstrates the Leiden community detection algorithm.
// Reference: https://neo4j.com/docs/graph-data-science/current/algorithms/leiden/
var LeidenExample = &algorithms.Leiden{
	BaseAlgorithm: algorithms.BaseAlgorithm{
		GraphName: "social-network",
		Mode:      algorithms.Stream,
	},
	MaxLevels: 10,
	Gamma:     1.0,
	Theta:     0.01,
}

// WCCExample demonstrates Weakly Connected Components.
// Reference: https://neo4j.com/docs/graph-data-science/current/algorithms/wcc/
var WCCExample = &algorithms.WCC{
	BaseAlgorithm: algorithms.BaseAlgorithm{
		GraphName: "graph",
		Mode:      algorithms.Mutate,
	},
	MutateProperty: "componentId",
}

// FastRPExample demonstrates the FastRP node embedding algorithm.
// Reference: https://neo4j.com/docs/graph-data-science/current/machine-learning/node-embeddings/fastrp/
var FastRPExample = &algorithms.FastRP{
	BaseAlgorithm: algorithms.BaseAlgorithm{
		GraphName: "graph",
		Mode:      algorithms.Mutate,
	},
	EmbeddingDimension:    128,
	IterationWeights:      []float64{0.0, 1.0, 1.0, 0.8, 0.5},
	NormalizationStrength: 0.5,
	MutateProperty:        "embedding",
}

// Node2VecExample demonstrates the Node2Vec embedding algorithm.
// Reference: https://neo4j.com/docs/graph-data-science/current/machine-learning/node-embeddings/node2vec/
var Node2VecExample = &algorithms.Node2Vec{
	BaseAlgorithm: algorithms.BaseAlgorithm{
		GraphName: "graph",
		Mode:      algorithms.Mutate,
	},
	EmbeddingDimension: 64,
	WalkLength:         80,
	WalksPerNode:       10,
	WindowSize:         10,
	InOutFactor:        1.0,
	ReturnFactor:       1.0,
	MutateProperty:     "node2vec",
}

// KNNExample demonstrates the K-Nearest Neighbors algorithm.
// Reference: https://neo4j.com/docs/graph-data-science/current/algorithms/knn/
var KNNExample = &algorithms.KNN{
	BaseAlgorithm: algorithms.BaseAlgorithm{
		GraphName: "similarity-graph",
		Mode:      algorithms.Mutate,
	},
	NodeProperties:        []string{"embedding"},
	K:                     10,
	SimilarityCutoff:      0.5,
	WriteRelationshipType: "SIMILAR",
	WriteProperty:         "score",
}

// NodeSimilarityExample demonstrates Node Similarity algorithm.
// Reference: https://neo4j.com/docs/graph-data-science/current/algorithms/node-similarity/
var NodeSimilarityExample = &algorithms.NodeSimilarity{
	BaseAlgorithm: algorithms.BaseAlgorithm{
		GraphName: "bipartite-graph",
		Mode:      algorithms.Stream,
	},
	TopK:             10,
	SimilarityCutoff: 0.1,
	DegreeCutoff:     1,
}

// DijkstraExample demonstrates shortest path with Dijkstra.
// Reference: https://neo4j.com/docs/graph-data-science/current/algorithms/dijkstra-source-target/
var DijkstraExample = &algorithms.Dijkstra{
	BaseAlgorithm: algorithms.BaseAlgorithm{
		GraphName: "road-network",
		Mode:      algorithms.Stream,
	},
	SourceNode:                 0,
	TargetNode:                 100,
	RelationshipWeightProperty: "distance",
}

// AllAlgorithmExamples returns all example algorithm configurations.
func AllAlgorithmExamples() []algorithms.Algorithm {
	return []algorithms.Algorithm{
		PageRankExample,
		ArticleRankExample,
		BetweennessExample,
		LouvainExample,
		LeidenExample,
		WCCExample,
		FastRPExample,
		Node2VecExample,
		KNNExample,
		NodeSimilarityExample,
		DijkstraExample,
	}
}
