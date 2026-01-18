// Package algorithms provides graph algorithm configurations for knowledge graph analytics.
package algorithms

import (
	"github.com/lex00/wetwire-neo4j-go/internal/algorithms"
)

// PaperInfluence uses PageRank to identify influential papers based on citation network.
// Higher scores indicate papers that are frequently cited by other important papers.
var PaperInfluence = &algorithms.PageRank{
	BaseAlgorithm: algorithms.BaseAlgorithm{
		GraphName: "citation-network",
		Mode:      algorithms.Stream,
	},
	DampingFactor: 0.85,
	MaxIterations: 20,
	Tolerance:     0.0000001,
}

// ResearchCommunities uses Louvain to detect clusters of related researchers.
// Identifies research groups based on coauthorship and citation patterns.
var ResearchCommunities = &algorithms.Louvain{
	BaseAlgorithm: algorithms.BaseAlgorithm{
		GraphName: "coauthor-network",
		Mode:      algorithms.Mutate,
	},
	MaxLevels:                      10,
	MaxIterations:                  10,
	Tolerance:                      0.0001,
	IncludeIntermediateCommunities: true,
	MutateProperty:                 "communityId",
}
