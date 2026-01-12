package examples

import (
	"github.com/lex00/wetwire-neo4j-go/internal/projections"
)

// SocialNetworkProjection demonstrates a simple native projection.
// Reference: https://neo4j.com/docs/graph-data-science/current/management-ops/graph-catalog-ops/
var SocialNetworkProjection = &projections.NativeProjection{
	BaseProjection: projections.BaseProjection{
		Name:      "social-network",
		GraphName: "social-network",
	},
	NodeLabels:        []string{"Person"},
	RelationshipTypes: []string{"KNOWS", "FOLLOWS"},
}

// BipartiteGraphProjection demonstrates a bipartite graph projection.
var BipartiteGraphProjection = &projections.NativeProjection{
	BaseProjection: projections.BaseProjection{
		Name:      "bipartite-graph",
		GraphName: "user-product-graph",
	},
	NodeLabels:        []string{"User", "Product"},
	RelationshipTypes: []string{"PURCHASED", "VIEWED", "RATED"},
}

// WeightedProjection demonstrates projection with relationship properties.
var WeightedProjection = &projections.NativeProjection{
	BaseProjection: projections.BaseProjection{
		Name:            "weighted-network",
		GraphName:       "weighted-network",
		ReadConcurrency: 4,
	},
	NodeProjections: []projections.NodeProjection{
		{
			Label:      "Person",
			Properties: []string{"age", "income"},
		},
	},
	RelationshipProjections: []projections.RelationshipProjection{
		{
			Type:        "KNOWS",
			Orientation: projections.Undirected,
			Properties:  []string{"weight"},
		},
		{
			Type:        "WORKS_WITH",
			Orientation: projections.Undirected,
			Aggregation: projections.Sum,
			Properties:  []string{"collaborations"},
		},
	},
}

// CypherProjectionExample demonstrates a Cypher-based projection.
// Reference: https://neo4j.com/docs/graph-data-science/current/management-ops/graph-catalog-ops/#catalog-graph-create-cypher
var CypherProjectionExample = &projections.CypherProjection{
	BaseProjection: projections.BaseProjection{
		Name:      "custom-projection",
		GraphName: "custom-graph",
	},
	NodeQuery: `
		MATCH (n:Person)
		WHERE n.active = true
		RETURN id(n) AS id, labels(n) AS labels, n.age AS age
	`,
	RelationshipQuery: `
		MATCH (a:Person)-[r:KNOWS]->(b:Person)
		WHERE a.active = true AND b.active = true
		RETURN id(a) AS source, id(b) AS target, type(r) AS type, r.weight AS weight
	`,
}

// MultiLabelProjection demonstrates projection with multiple node labels.
var MultiLabelProjection = &projections.NativeProjection{
	BaseProjection: projections.BaseProjection{
		Name:      "knowledge-graph",
		GraphName: "knowledge-graph",
	},
	NodeProjections: []projections.NodeProjection{
		{Label: "Person", Properties: []string{"name", "embedding"}},
		{Label: "Company", Properties: []string{"name", "embedding"}},
		{Label: "Location", Properties: []string{"name", "coordinates"}},
		{Label: "Document", Properties: []string{"title", "embedding"}},
	},
	RelationshipProjections: []projections.RelationshipProjection{
		{Type: "WORKS_FOR", Orientation: projections.Natural},
		{Type: "LOCATED_IN", Orientation: projections.Natural},
		{Type: "MENTIONS", Orientation: projections.Natural},
		{Type: "RELATED_TO", Orientation: projections.Undirected},
	},
}

// AllProjectionExamples returns all example projection configurations.
func AllProjectionExamples() []projections.Projection {
	return []projections.Projection{
		SocialNetworkProjection,
		BipartiteGraphProjection,
		WeightedProjection,
		CypherProjectionExample,
		MultiLabelProjection,
	}
}
