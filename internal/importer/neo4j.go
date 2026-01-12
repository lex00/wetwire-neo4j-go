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

	// Build node and relationship type definitions from constraints and indexes
	result.NodeTypes, result.RelationshipTypes = i.buildTypeDefinitions(constraints, indexes)

	return result, nil
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

func (i *Neo4jImporter) buildTypeDefinitions(constraints []ConstraintDefinition, indexes []IndexDefinition) ([]NodeTypeDefinition, []RelationshipTypeDefinition) {
	nodeTypes := make(map[string]*NodeTypeDefinition)
	relTypes := make(map[string]*RelationshipTypeDefinition)

	// Process constraints
	for _, c := range constraints {
		if c.EntityType == "NODE" {
			if _, exists := nodeTypes[c.Label]; !exists {
				nodeTypes[c.Label] = &NodeTypeDefinition{
					Label:       c.Label,
					Properties:  make([]PropertyDefinition, 0),
					Constraints: make([]ConstraintDefinition, 0),
					Indexes:     make([]IndexDefinition, 0),
				}
			}
			nodeTypes[c.Label].Constraints = append(nodeTypes[c.Label].Constraints, c)

			// Add properties from constraints
			for _, prop := range c.Properties {
				found := false
				for _, existing := range nodeTypes[c.Label].Properties {
					if existing.Name == prop {
						found = true
						break
					}
				}
				if !found {
					nodeTypes[c.Label].Properties = append(nodeTypes[c.Label].Properties, PropertyDefinition{
						Name:     prop,
						Type:     "STRING", // Default type
						Required: c.Type == "NODE_PROPERTY_EXISTENCE" || c.Type == "NODE_KEY",
					})
				}
			}
		} else if c.EntityType == "RELATIONSHIP" {
			if _, exists := relTypes[c.Label]; !exists {
				relTypes[c.Label] = &RelationshipTypeDefinition{
					Type:        c.Label,
					Properties:  make([]PropertyDefinition, 0),
					Constraints: make([]ConstraintDefinition, 0),
					Indexes:     make([]IndexDefinition, 0),
				}
			}
			relTypes[c.Label].Constraints = append(relTypes[c.Label].Constraints, c)
		}
	}

	// Process indexes
	for _, idx := range indexes {
		if idx.EntityType == "NODE" {
			if _, exists := nodeTypes[idx.Label]; !exists {
				nodeTypes[idx.Label] = &NodeTypeDefinition{
					Label:       idx.Label,
					Properties:  make([]PropertyDefinition, 0),
					Constraints: make([]ConstraintDefinition, 0),
					Indexes:     make([]IndexDefinition, 0),
				}
			}
			nodeTypes[idx.Label].Indexes = append(nodeTypes[idx.Label].Indexes, idx)

			// Add properties from indexes
			for _, prop := range idx.Properties {
				found := false
				for _, existing := range nodeTypes[idx.Label].Properties {
					if existing.Name == prop {
						found = true
						break
					}
				}
				if !found {
					nodeTypes[idx.Label].Properties = append(nodeTypes[idx.Label].Properties, PropertyDefinition{
						Name: prop,
						Type: "STRING",
					})
				}
			}
		} else if idx.EntityType == "RELATIONSHIP" {
			if _, exists := relTypes[idx.Label]; !exists {
				relTypes[idx.Label] = &RelationshipTypeDefinition{
					Type:        idx.Label,
					Properties:  make([]PropertyDefinition, 0),
					Constraints: make([]ConstraintDefinition, 0),
					Indexes:     make([]IndexDefinition, 0),
				}
			}
			relTypes[idx.Label].Indexes = append(relTypes[idx.Label].Indexes, idx)
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
