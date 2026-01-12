// Package schema provides type-safe schema definition primitives for Neo4j.
//
// This package implements the core schema definition system including:
// - NodeType: Define node labels with properties and constraints
// - RelationshipType: Define relationship types with source/target and cardinality
// - Property: Define properties with types, constraints, and validation
// - Constraint: Define node and relationship constraints
//
// Example usage:
//
//	person := &schema.NodeType{
//		Label: "Person",
//		Properties: []schema.Property{
//			{Name: "name", Type: schema.STRING, Required: true},
//			{Name: "age", Type: schema.INTEGER},
//		},
//	}
package schema

import "time"

// PropertyType represents the data type of a property in Neo4j.
type PropertyType string

const (
	// STRING represents a Neo4j String type.
	STRING PropertyType = "STRING"
	// INTEGER represents a Neo4j Integer type.
	INTEGER PropertyType = "INTEGER"
	// FLOAT represents a Neo4j Float type.
	FLOAT PropertyType = "FLOAT"
	// BOOLEAN represents a Neo4j Boolean type.
	BOOLEAN PropertyType = "BOOLEAN"
	// DATE represents a Neo4j Date type.
	DATE PropertyType = "DATE"
	// DATETIME represents a Neo4j DateTime type.
	DATETIME PropertyType = "DATETIME"
	// POINT represents a Neo4j Point type.
	POINT PropertyType = "POINT"
	// LIST_STRING represents a Neo4j List<String> type.
	LIST_STRING PropertyType = "LIST_STRING"
	// LIST_INTEGER represents a Neo4j List<Integer> type.
	LIST_INTEGER PropertyType = "LIST_INTEGER"
	// LIST_FLOAT represents a Neo4j List<Float> type.
	LIST_FLOAT PropertyType = "LIST_FLOAT"
)

// Cardinality defines the relationship cardinality between nodes.
type Cardinality string

const (
	// ONE_TO_ONE represents a 1:1 relationship cardinality.
	ONE_TO_ONE Cardinality = "ONE_TO_ONE"
	// ONE_TO_MANY represents a 1:N relationship cardinality.
	ONE_TO_MANY Cardinality = "ONE_TO_MANY"
	// MANY_TO_ONE represents a N:1 relationship cardinality.
	MANY_TO_ONE Cardinality = "MANY_TO_ONE"
	// MANY_TO_MANY represents a N:M relationship cardinality.
	MANY_TO_MANY Cardinality = "MANY_TO_MANY"
)

// ConstraintType defines the type of constraint.
type ConstraintType string

const (
	// UNIQUE enforces that a property value is unique across all nodes with that label.
	UNIQUE ConstraintType = "UNIQUE"
	// EXISTS enforces that a property must have a value (NOT NULL).
	EXISTS ConstraintType = "EXISTS"
	// NODE_KEY creates a composite key constraint (unique + exists on multiple properties).
	NODE_KEY ConstraintType = "NODE_KEY"
	// REL_KEY creates a relationship key constraint.
	REL_KEY ConstraintType = "REL_KEY"
)

// IndexType defines the type of index.
type IndexType string

const (
	// BTREE is the default index type for range queries.
	BTREE IndexType = "BTREE"
	// TEXT is optimized for text search operations.
	TEXT IndexType = "TEXT"
	// FULLTEXT enables full-text search across multiple properties.
	FULLTEXT IndexType = "FULLTEXT"
	// POINT is optimized for spatial queries.
	POINT_INDEX IndexType = "POINT"
	// VECTOR is for vector similarity search (Neo4j 5.13+).
	VECTOR IndexType = "VECTOR"
)

// Property represents a property definition with type and constraints.
type Property struct {
	// Name is the property key name.
	Name string
	// Type is the Neo4j data type.
	Type PropertyType
	// Required indicates if the property must have a value (enforced via EXISTS constraint).
	Required bool
	// Unique indicates if the property must be unique across nodes with this label.
	Unique bool
	// Description is optional documentation for the property.
	Description string
	// DefaultValue is the default value for the property (type depends on Type).
	DefaultValue any
}

// Constraint represents a constraint definition on a node or relationship.
type Constraint struct {
	// Name is the constraint name in Neo4j.
	Name string
	// Type is the constraint type (UNIQUE, EXISTS, NODE_KEY, REL_KEY).
	Type ConstraintType
	// Properties are the property names involved in the constraint.
	Properties []string
}

// Index represents an index definition on a node or relationship.
type Index struct {
	// Name is the index name in Neo4j.
	Name string
	// Type is the index type (BTREE, TEXT, FULLTEXT, POINT, VECTOR).
	Type IndexType
	// Properties are the property names to index.
	Properties []string
	// Options contains index-specific options (e.g., vector dimensions).
	Options map[string]any
}

// NodeType represents a node label definition with properties and constraints.
type NodeType struct {
	// Label is the Neo4j node label (should be PascalCase).
	Label string
	// Properties defines the properties for nodes with this label.
	Properties []Property
	// Constraints defines constraints on this node type.
	Constraints []Constraint
	// Indexes defines indexes on this node type.
	Indexes []Index
	// Description is optional documentation for the node type.
	Description string
}

// RelationshipType represents a relationship type definition.
type RelationshipType struct {
	// Label is the Neo4j relationship type (should be SCREAMING_SNAKE_CASE).
	Label string
	// Source is the source node type label.
	Source string
	// Target is the target node type label.
	Target string
	// Cardinality defines the relationship cardinality.
	Cardinality Cardinality
	// Properties defines the properties for this relationship type.
	Properties []Property
	// Constraints defines constraints on this relationship type.
	Constraints []Constraint
	// Description is optional documentation for the relationship type.
	Description string
}

// Point represents a Neo4j spatial point with coordinates.
type Point struct {
	// SRID is the Spatial Reference System Identifier (4326 for WGS84, 7203 for Cartesian).
	SRID int
	// X is the longitude (WGS84) or x coordinate (Cartesian).
	X float64
	// Y is the latitude (WGS84) or y coordinate (Cartesian).
	Y float64
	// Z is the optional height/z coordinate.
	Z *float64
}

// Resource is an interface that all Neo4j resource definitions implement.
type Resource interface {
	// ResourceType returns the type of resource ("NodeType", "RelationshipType", etc.).
	ResourceType() string
	// ResourceName returns the name/label of the resource.
	ResourceName() string
}

// Ensure NodeType implements Resource.
func (n *NodeType) ResourceType() string { return "NodeType" }
func (n *NodeType) ResourceName() string { return n.Label }

// Ensure RelationshipType implements Resource.
func (r *RelationshipType) ResourceType() string { return "RelationshipType" }
func (r *RelationshipType) ResourceName() string { return r.Label }

// PropertyValue represents a typed property value for validation.
type PropertyValue struct {
	String      *string
	Integer     *int64
	Float       *float64
	Boolean     *bool
	Date        *time.Time
	DateTime    *time.Time
	Point       *Point
	ListString  []string
	ListInteger []int64
	ListFloat   []float64
}

// NewStringValue creates a PropertyValue containing a string.
func NewStringValue(s string) PropertyValue {
	return PropertyValue{String: &s}
}

// NewIntegerValue creates a PropertyValue containing an integer.
func NewIntegerValue(i int64) PropertyValue {
	return PropertyValue{Integer: &i}
}

// NewFloatValue creates a PropertyValue containing a float.
func NewFloatValue(f float64) PropertyValue {
	return PropertyValue{Float: &f}
}

// NewBooleanValue creates a PropertyValue containing a boolean.
func NewBooleanValue(b bool) PropertyValue {
	return PropertyValue{Boolean: &b}
}

// NewDateValue creates a PropertyValue containing a date.
func NewDateValue(t time.Time) PropertyValue {
	return PropertyValue{Date: &t}
}

// NewDateTimeValue creates a PropertyValue containing a datetime.
func NewDateTimeValue(t time.Time) PropertyValue {
	return PropertyValue{DateTime: &t}
}

// NewPointValue creates a PropertyValue containing a point.
func NewPointValue(p Point) PropertyValue {
	return PropertyValue{Point: &p}
}

// NewListStringValue creates a PropertyValue containing a list of strings.
func NewListStringValue(s []string) PropertyValue {
	return PropertyValue{ListString: s}
}

// NewListIntegerValue creates a PropertyValue containing a list of integers.
func NewListIntegerValue(i []int64) PropertyValue {
	return PropertyValue{ListInteger: i}
}

// NewListFloatValue creates a PropertyValue containing a list of floats.
func NewListFloatValue(f []float64) PropertyValue {
	return PropertyValue{ListFloat: f}
}
