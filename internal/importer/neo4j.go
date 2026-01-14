package importer

import (
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// Neo4jImporter imports configurations from a Neo4j database.
type Neo4jImporter struct {
	driver   neo4j.DriverWithContext
	database string
}

// Neo4jConfig holds the configuration for connecting to Neo4j.
type Neo4jConfig struct {
	URI      string
	Username string
	Password string
	Database string
}

// NewNeo4jImporter creates a new Neo4j importer.
func NewNeo4jImporter(ctx context.Context, config Neo4jConfig) (*Neo4jImporter, error) {
	driver, err := neo4j.NewDriverWithContext(config.URI, neo4j.BasicAuth(config.Username, config.Password, ""))
	if err != nil {
		return nil, fmt.Errorf("failed to create Neo4j driver: %w", err)
	}

	// Verify connectivity
	if err := driver.VerifyConnectivity(ctx); err != nil {
		_ = driver.Close(ctx)
		return nil, fmt.Errorf("failed to connect to Neo4j: %w", err)
	}

	database := config.Database
	if database == "" {
		database = "neo4j"
	}

	return &Neo4jImporter{
		driver:   driver,
		database: database,
	}, nil
}

// Close closes the database connection.
func (i *Neo4jImporter) Close(ctx context.Context) error {
	return i.driver.Close(ctx)
}

// Import imports configurations from the Neo4j database.
func (i *Neo4jImporter) Import(ctx context.Context) (*ImportResult, error) {
	result := &ImportResult{
		NodeTypes:         make([]NodeTypeDefinition, 0),
		RelationshipTypes: make([]RelationshipTypeDefinition, 0),
		Constraints:       make([]ConstraintDefinition, 0),
		Indexes:           make([]IndexDefinition, 0),
	}

	// Fetch all node labels
	nodeLabels, err := i.fetchNodeLabels(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch node labels: %w", err)
	}

	// Fetch all relationship types
	relTypes, err := i.fetchRelationshipTypes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch relationship types: %w", err)
	}

	// Fetch constraints
	constraints, err := i.fetchConstraints(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch constraints: %w", err)
	}
	result.Constraints = constraints

	// Fetch indexes
	indexes, err := i.fetchIndexes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch indexes: %w", err)
	}
	result.Indexes = indexes

	// Fetch properties for all node labels
	nodeProperties, err := i.fetchNodeProperties(ctx, nodeLabels)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch node properties: %w", err)
	}

	// Fetch properties for all relationship types
	relProperties, err := i.fetchRelationshipProperties(ctx, relTypes)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch relationship properties: %w", err)
	}

	// Build node and relationship type definitions
	result.NodeTypes, result.RelationshipTypes = i.buildTypeDefinitions(
		nodeLabels, relTypes, constraints, indexes, nodeProperties, relProperties,
	)

	return result, nil
}

func (i *Neo4jImporter) fetchNodeLabels(ctx context.Context) ([]string, error) {
	session := i.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: i.database})
	defer func() { _ = session.Close(ctx) }()

	result, err := session.Run(ctx, "CALL db.labels()", nil)
	if err != nil {
		return nil, err
	}

	var labels []string
	for result.Next(ctx) {
		record := result.Record()
		if label, ok := record.Values[0].(string); ok {
			labels = append(labels, label)
		}
	}

	return labels, result.Err()
}

func (i *Neo4jImporter) fetchRelationshipTypes(ctx context.Context) ([]string, error) {
	session := i.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: i.database})
	defer func() { _ = session.Close(ctx) }()

	result, err := session.Run(ctx, "CALL db.relationshipTypes()", nil)
	if err != nil {
		return nil, err
	}

	var types []string
	for result.Next(ctx) {
		record := result.Record()
		if relType, ok := record.Values[0].(string); ok {
			types = append(types, relType)
		}
	}

	return types, result.Err()
}

func (i *Neo4jImporter) fetchNodeProperties(ctx context.Context, labels []string) (map[string][]PropertyDefinition, error) {
	properties := make(map[string][]PropertyDefinition)

	for _, label := range labels {
		props, err := i.fetchPropertiesForLabel(ctx, label)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch properties for label %s: %w", label, err)
		}
		properties[label] = props
	}

	return properties, nil
}

func (i *Neo4jImporter) fetchPropertiesForLabel(ctx context.Context, label string) ([]PropertyDefinition, error) {
	session := i.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: i.database})
	defer func() { _ = session.Close(ctx) }()

	// Sample a node to discover properties and their types
	query := fmt.Sprintf("MATCH (n:`%s`) RETURN n LIMIT 1", label)
	result, err := session.Run(ctx, query, nil)
	if err != nil {
		return nil, err
	}

	var props []PropertyDefinition
	if result.Next(ctx) {
		record := result.Record()
		if node, ok := record.Values[0].(neo4j.Node); ok {
			for key, value := range node.Props {
				props = append(props, PropertyDefinition{
					Name: key,
					Type: inferNeo4jType(value),
				})
			}
		}
	}

	return props, result.Err()
}

func (i *Neo4jImporter) fetchRelationshipProperties(ctx context.Context, relTypes []string) (map[string][]PropertyDefinition, error) {
	properties := make(map[string][]PropertyDefinition)

	for _, relType := range relTypes {
		props, err := i.fetchPropertiesForRelType(ctx, relType)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch properties for relationship type %s: %w", relType, err)
		}
		properties[relType] = props
	}

	return properties, nil
}

func (i *Neo4jImporter) fetchPropertiesForRelType(ctx context.Context, relType string) ([]PropertyDefinition, error) {
	session := i.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: i.database})
	defer func() { _ = session.Close(ctx) }()

	// Sample a relationship to discover properties and their types
	query := fmt.Sprintf("MATCH ()-[r:`%s`]->() RETURN r LIMIT 1", relType)
	result, err := session.Run(ctx, query, nil)
	if err != nil {
		return nil, err
	}

	var props []PropertyDefinition
	if result.Next(ctx) {
		record := result.Record()
		if rel, ok := record.Values[0].(neo4j.Relationship); ok {
			for key, value := range rel.Props {
				props = append(props, PropertyDefinition{
					Name: key,
					Type: inferNeo4jType(value),
				})
			}
		}
	}

	return props, result.Err()
}

func inferNeo4jType(value any) string {
	switch value.(type) {
	case string:
		return "STRING"
	case int64:
		return "INTEGER"
	case float64:
		return "FLOAT"
	case bool:
		return "BOOLEAN"
	case []any:
		return "LIST"
	default:
		return "STRING"
	}
}

func (i *Neo4jImporter) fetchConstraints(ctx context.Context) ([]ConstraintDefinition, error) {
	session := i.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: i.database})
	defer func() { _ = session.Close(ctx) }()

	result, err := session.Run(ctx, "SHOW CONSTRAINTS", nil)
	if err != nil {
		return nil, err
	}

	var constraints []ConstraintDefinition
	for result.Next(ctx) {
		record := result.Record()

		name, _ := record.Get("name")
		constraintType, _ := record.Get("type")
		entityType, _ := record.Get("entityType")
		labelsOrTypes, _ := record.Get("labelsOrTypes")
		properties, _ := record.Get("properties")

		constraint := ConstraintDefinition{
			Name:       toString(name),
			Type:       toString(constraintType),
			EntityType: toString(entityType),
		}

		// Parse labels/types
		if labels, ok := labelsOrTypes.([]any); ok && len(labels) > 0 {
			constraint.Label = toString(labels[0])
		}

		// Parse properties
		if props, ok := properties.([]any); ok {
			for _, p := range props {
				constraint.Properties = append(constraint.Properties, toString(p))
			}
		}

		constraints = append(constraints, constraint)
	}

	return constraints, result.Err()
}

func (i *Neo4jImporter) fetchIndexes(ctx context.Context) ([]IndexDefinition, error) {
	session := i.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: i.database})
	defer func() { _ = session.Close(ctx) }()

	result, err := session.Run(ctx, "SHOW INDEXES", nil)
	if err != nil {
		return nil, err
	}

	var indexes []IndexDefinition
	for result.Next(ctx) {
		record := result.Record()

		name, _ := record.Get("name")
		indexType, _ := record.Get("type")
		entityType, _ := record.Get("entityType")
		labelsOrTypes, _ := record.Get("labelsOrTypes")
		properties, _ := record.Get("properties")

		// Skip constraints indexes (they're covered by constraints)
		if owningConstraint, _ := record.Get("owningConstraint"); owningConstraint != nil {
			continue
		}

		index := IndexDefinition{
			Name:       toString(name),
			Type:       toString(indexType),
			EntityType: toString(entityType),
			Options:    make(map[string]any),
		}

		// Parse labels/types
		if labels, ok := labelsOrTypes.([]any); ok && len(labels) > 0 {
			index.Label = toString(labels[0])
		}

		// Parse properties
		if props, ok := properties.([]any); ok {
			for _, p := range props {
				index.Properties = append(index.Properties, toString(p))
			}
		}

		indexes = append(indexes, index)
	}

	return indexes, result.Err()
}

func (i *Neo4jImporter) buildTypeDefinitions(
	nodeLabels []string,
	relationshipTypes []string,
	constraints []ConstraintDefinition,
	indexes []IndexDefinition,
	nodeProperties map[string][]PropertyDefinition,
	relProperties map[string][]PropertyDefinition,
) ([]NodeTypeDefinition, []RelationshipTypeDefinition) {
	nodeTypes := make(map[string]*NodeTypeDefinition)
	relTypes := make(map[string]*RelationshipTypeDefinition)

	// Initialize all node labels from db.labels()
	for _, label := range nodeLabels {
		nodeTypes[label] = &NodeTypeDefinition{
			Label:       label,
			Properties:  make([]PropertyDefinition, 0),
			Constraints: make([]ConstraintDefinition, 0),
			Indexes:     make([]IndexDefinition, 0),
		}
		// Add properties discovered from sampling
		if props, ok := nodeProperties[label]; ok {
			nodeTypes[label].Properties = append(nodeTypes[label].Properties, props...)
		}
	}

	// Initialize all relationship types from db.relationshipTypes()
	for _, relType := range relationshipTypes {
		relTypes[relType] = &RelationshipTypeDefinition{
			Type:        relType,
			Properties:  make([]PropertyDefinition, 0),
			Constraints: make([]ConstraintDefinition, 0),
			Indexes:     make([]IndexDefinition, 0),
		}
		// Add properties discovered from sampling
		if props, ok := relProperties[relType]; ok {
			relTypes[relType].Properties = append(relTypes[relType].Properties, props...)
		}
	}

	// Process constraints
	for _, c := range constraints {
		switch c.EntityType {
		case "NODE":
			if node, exists := nodeTypes[c.Label]; exists {
				node.Constraints = append(node.Constraints, c)

				// Mark properties as required if they have existence constraints
				for _, prop := range c.Properties {
					if c.Type == "NODE_PROPERTY_EXISTENCE" || c.Type == "NODE_KEY" {
						for i := range node.Properties {
							if node.Properties[i].Name == prop {
								node.Properties[i].Required = true
							}
						}
					}
				}
			}
		case "RELATIONSHIP":
			if rel, exists := relTypes[c.Label]; exists {
				rel.Constraints = append(rel.Constraints, c)
			}
		}
	}

	// Process indexes
	for _, idx := range indexes {
		switch idx.EntityType {
		case "NODE":
			if node, exists := nodeTypes[idx.Label]; exists {
				node.Indexes = append(node.Indexes, idx)
			}
		case "RELATIONSHIP":
			if rel, exists := relTypes[idx.Label]; exists {
				rel.Indexes = append(rel.Indexes, idx)
			}
		}
	}

	// Convert maps to slices
	nodes := make([]NodeTypeDefinition, 0, len(nodeTypes))
	for _, n := range nodeTypes {
		nodes = append(nodes, *n)
	}

	rels := make([]RelationshipTypeDefinition, 0, len(relTypes))
	for _, r := range relTypes {
		rels = append(rels, *r)
	}

	return nodes, rels
}

func toString(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}
