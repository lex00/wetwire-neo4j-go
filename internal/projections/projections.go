// Package projections provides typed configurations for Neo4j GDS graph projections.
//
// This package implements type-safe configurations for graph projections including:
// - NativeProjection: Project node labels and relationship types directly
// - CypherProjection: Use Cypher queries to define projections
// - DataFrameProjection: Used with Aura Analytics
//
// Example usage:
//
//	projection := &projections.NativeProjection{
//		Name:              "social_graph",
//		NodeLabels:        []string{"Person", "Company"},
//		RelationshipTypes: []string{"KNOWS", "WORKS_AT"},
//	}
//	cypher, err := serializer.ToCypher(projection)
package projections

// ProjectionType represents the type of graph projection.
type ProjectionType string

const (
	// Native projection using node labels and relationship types.
	Native ProjectionType = "Native"
	// Cypher projection using custom Cypher queries.
	Cypher ProjectionType = "Cypher"
	// DataFrame projection for Aura Analytics.
	DataFrame ProjectionType = "DataFrame"
)

// Orientation defines the direction of relationships.
type Orientation string

const (
	// Natural preserves the original direction.
	Natural Orientation = "NATURAL"
	// Reverse inverts the relationship direction.
	Reverse Orientation = "REVERSE"
	// Undirected treats relationships as bidirectional.
	Undirected Orientation = "UNDIRECTED"
)

// Aggregation defines how to aggregate parallel relationships.
type Aggregation string

const (
	// None keeps all parallel relationships.
	None Aggregation = "NONE"
	// Sum adds property values for parallel relationships.
	Sum Aggregation = "SUM"
	// Min keeps the minimum property value.
	Min Aggregation = "MIN"
	// Max keeps the maximum property value.
	Max Aggregation = "MAX"
	// Single keeps a single relationship arbitrarily.
	Single Aggregation = "SINGLE"
	// Count replaces property with count of relationships.
	Count Aggregation = "COUNT"
)

// Projection is the interface that all graph projection configurations implement.
type Projection interface {
	// ProjectionName returns the name of this projection.
	ProjectionName() string
	// ProjectionType returns the type of projection.
	ProjectionType() ProjectionType
	// GetNodeProjections returns node projection configurations.
	GetNodeProjections() []NodeProjection
	// GetRelationshipProjections returns relationship projection configurations.
	GetRelationshipProjections() []RelationshipProjection
}

// BaseProjection contains common projection configuration fields.
type BaseProjection struct {
	// Name is the projection name used for identification.
	Name string
	// GraphName is the name to use for the projected graph in GDS catalog.
	GraphName string
	// ReadConcurrency for parallel graph loading (default: 4).
	ReadConcurrency int
}

// ProjectionName returns the projection name.
func (b *BaseProjection) ProjectionName() string {
	return b.Name
}

// NodeProjection defines how to project nodes.
type NodeProjection struct {
	// Label is the node label to project.
	Label string
	// Properties are node properties to include.
	Properties []string
	// DefaultValue is the default for missing properties.
	DefaultValue any
}

// RelationshipProjection defines how to project relationships.
type RelationshipProjection struct {
	// Type is the relationship type to project.
	Type string
	// Orientation defines direction handling.
	Orientation Orientation
	// Aggregation defines parallel relationship handling.
	Aggregation Aggregation
	// Properties are relationship properties to include.
	Properties []string
	// DefaultValue is the default for missing properties.
	DefaultValue any
}

// NativeProjection projects the graph using node labels and relationship types.
type NativeProjection struct {
	BaseProjection
	// NodeLabels are the node labels to include (simple form).
	NodeLabels []string
	// RelationshipTypes are the relationship types to include (simple form).
	RelationshipTypes []string
	// NodeProjections are detailed node projection configurations.
	NodeProjections []NodeProjection
	// RelationshipProjections are detailed relationship projection configurations.
	RelationshipProjections []RelationshipProjection
}

func (p *NativeProjection) ProjectionType() ProjectionType { return Native }

// GetNodeProjections returns node projections, converting simple form if needed.
func (p *NativeProjection) GetNodeProjections() []NodeProjection {
	if len(p.NodeProjections) > 0 {
		return p.NodeProjections
	}
	// Convert simple form to projection form
	projections := make([]NodeProjection, len(p.NodeLabels))
	for i, label := range p.NodeLabels {
		projections[i] = NodeProjection{Label: label}
	}
	return projections
}

// GetRelationshipProjections returns relationship projections, converting simple form if needed.
func (p *NativeProjection) GetRelationshipProjections() []RelationshipProjection {
	if len(p.RelationshipProjections) > 0 {
		return p.RelationshipProjections
	}
	// Convert simple form to projection form
	projections := make([]RelationshipProjection, len(p.RelationshipTypes))
	for i, relType := range p.RelationshipTypes {
		projections[i] = RelationshipProjection{Type: relType}
	}
	return projections
}

// CypherProjection projects the graph using custom Cypher queries.
type CypherProjection struct {
	BaseProjection
	// NodeQuery is the Cypher query for nodes.
	// Must return id, and optionally labels and properties.
	NodeQuery string
	// RelationshipQuery is the Cypher query for relationships.
	// Must return source, target, and optionally type and properties.
	RelationshipQuery string
	// Parameters are query parameters.
	Parameters map[string]any
	// ValidateRelationships validates that source/target nodes exist.
	ValidateRelationships bool
}

func (p *CypherProjection) ProjectionType() ProjectionType { return Cypher }

// GetNodeProjections returns empty for Cypher projections (uses query instead).
func (p *CypherProjection) GetNodeProjections() []NodeProjection {
	return nil
}

// GetRelationshipProjections returns empty for Cypher projections (uses query instead).
func (p *CypherProjection) GetRelationshipProjections() []RelationshipProjection {
	return nil
}

// NodeDataFrame describes node data for DataFrame projection.
type NodeDataFrame struct {
	// Label is the node label.
	Label string
	// Properties are the properties to include.
	Properties []string
	// IDColumn is the column to use as node ID.
	IDColumn string
}

// RelationshipDataFrame describes relationship data for DataFrame projection.
type RelationshipDataFrame struct {
	// Type is the relationship type.
	Type string
	// SourceColumn is the column for source node ID.
	SourceColumn string
	// TargetColumn is the column for target node ID.
	TargetColumn string
	// Properties are the properties to include.
	Properties []string
}

// DataFrameProjection projects the graph from DataFrames (Aura Analytics).
type DataFrameProjection struct {
	BaseProjection
	// NodeDataFrames define node projections from DataFrames.
	NodeDataFrames []NodeDataFrame
	// RelationshipDataFrames define relationship projections from DataFrames.
	RelationshipDataFrames []RelationshipDataFrame
}

func (p *DataFrameProjection) ProjectionType() ProjectionType { return DataFrame }

// GetNodeProjections converts DataFrame node configs to standard form.
func (p *DataFrameProjection) GetNodeProjections() []NodeProjection {
	projections := make([]NodeProjection, len(p.NodeDataFrames))
	for i, df := range p.NodeDataFrames {
		projections[i] = NodeProjection{
			Label:      df.Label,
			Properties: df.Properties,
		}
	}
	return projections
}

// GetRelationshipProjections converts DataFrame relationship configs to standard form.
func (p *DataFrameProjection) GetRelationshipProjections() []RelationshipProjection {
	projections := make([]RelationshipProjection, len(p.RelationshipDataFrames))
	for i, df := range p.RelationshipDataFrames {
		projections[i] = RelationshipProjection{
			Type:       df.Type,
			Properties: df.Properties,
		}
	}
	return projections
}
