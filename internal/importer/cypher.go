package importer

import (
	"bufio"
	"context"
	"os"
	"regexp"
	"strings"
)

// CypherImporter imports configurations from Cypher schema files.
type CypherImporter struct {
	filePath string
}

// NewCypherImporter creates a new Cypher file importer.
func NewCypherImporter(filePath string) *CypherImporter {
	return &CypherImporter{filePath: filePath}
}

// Import parses a Cypher file and extracts schema definitions.
func (i *CypherImporter) Import(ctx context.Context) (*ImportResult, error) {
	file, err := os.Open(i.filePath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	result := &ImportResult{
		NodeTypes:         make([]NodeTypeDefinition, 0),
		RelationshipTypes: make([]RelationshipTypeDefinition, 0),
		Constraints:       make([]ConstraintDefinition, 0),
		Indexes:           make([]IndexDefinition, 0),
	}

	scanner := bufio.NewScanner(file)
	var currentStatement strings.Builder

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "--") {
			continue
		}

		currentStatement.WriteString(line)
		currentStatement.WriteString(" ")

		// Check if statement is complete (ends with semicolon)
		if strings.HasSuffix(line, ";") {
			stmt := strings.TrimSpace(currentStatement.String())
			i.parseStatement(stmt, result)
			currentStatement.Reset()
		}
	}

	// Handle any remaining statement
	if currentStatement.Len() > 0 {
		stmt := strings.TrimSpace(currentStatement.String())
		i.parseStatement(stmt, result)
	}

	// Build type definitions from constraints and indexes
	nodes, rels := buildTypeDefinitionsFromResults(result.Constraints, result.Indexes)
	result.NodeTypes = nodes
	result.RelationshipTypes = rels

	return result, scanner.Err()
}

func (i *CypherImporter) parseStatement(stmt string, result *ImportResult) {
	stmtUpper := strings.ToUpper(stmt)

	if strings.HasPrefix(stmtUpper, "CREATE CONSTRAINT") {
		if c := parseConstraintStatement(stmt); c != nil {
			result.Constraints = append(result.Constraints, *c)
		}
	} else if strings.HasPrefix(stmtUpper, "CREATE INDEX") {
		if idx := parseIndexStatement(stmt); idx != nil {
			result.Indexes = append(result.Indexes, *idx)
		}
	}
}

// Regex patterns for parsing Cypher statements
var (
	// CREATE CONSTRAINT constraint_name FOR (n:Label) REQUIRE n.property IS UNIQUE
	uniqueConstraintRe = regexp.MustCompile(`(?i)CREATE\s+CONSTRAINT\s+(\w+)?\s*(?:IF\s+NOT\s+EXISTS\s+)?FOR\s+\((\w+):(\w+)\)\s+REQUIRE\s+\(?([^)]+)\)?\s+IS\s+UNIQUE`)

	// CREATE CONSTRAINT constraint_name FOR (n:Label) REQUIRE (n.prop1, n.prop2) IS NODE KEY
	nodeKeyConstraintRe = regexp.MustCompile(`(?i)CREATE\s+CONSTRAINT\s+(\w+)?\s*(?:IF\s+NOT\s+EXISTS\s+)?FOR\s+\((\w+):(\w+)\)\s+REQUIRE\s+\(([^)]+)\)\s+IS\s+NODE\s+KEY`)

	// CREATE CONSTRAINT constraint_name FOR (n:Label) REQUIRE n.property IS NOT NULL
	existenceConstraintRe = regexp.MustCompile(`(?i)CREATE\s+CONSTRAINT\s+(\w+)?\s*(?:IF\s+NOT\s+EXISTS\s+)?FOR\s+\((\w+):(\w+)\)\s+REQUIRE\s+(\w+\.\w+)\s+IS\s+NOT\s+NULL`)

	// CREATE INDEX index_name FOR (n:Label) ON (n.property)
	indexRe = regexp.MustCompile(`(?i)CREATE\s+(?:(RANGE|FULLTEXT|TEXT|VECTOR)\s+)?INDEX\s+(\w+)?\s*(?:IF\s+NOT\s+EXISTS\s+)?FOR\s+\((\w+):(\w+)\)\s+ON\s+\(([^)]+)\)`)
)

func parseConstraintStatement(stmt string) *ConstraintDefinition {
	// Try unique constraint
	if matches := uniqueConstraintRe.FindStringSubmatch(stmt); matches != nil {
		return &ConstraintDefinition{
			Name:       matches[1],
			Type:       "UNIQUENESS",
			EntityType: "NODE",
			Label:      matches[3],
			Properties: parsePropertyList(matches[4], matches[2]),
		}
	}

	// Try node key constraint
	if matches := nodeKeyConstraintRe.FindStringSubmatch(stmt); matches != nil {
		return &ConstraintDefinition{
			Name:       matches[1],
			Type:       "NODE_KEY",
			EntityType: "NODE",
			Label:      matches[3],
			Properties: parsePropertyList(matches[4], matches[2]),
		}
	}

	// Try existence constraint
	if matches := existenceConstraintRe.FindStringSubmatch(stmt); matches != nil {
		return &ConstraintDefinition{
			Name:       matches[1],
			Type:       "NODE_PROPERTY_EXISTENCE",
			EntityType: "NODE",
			Label:      matches[3],
			Properties: parsePropertyList(matches[4], matches[2]),
		}
	}

	return nil
}

func parseIndexStatement(stmt string) *IndexDefinition {
	if matches := indexRe.FindStringSubmatch(stmt); matches != nil {
		indexType := "RANGE"
		if matches[1] != "" {
			indexType = strings.ToUpper(matches[1])
		}

		return &IndexDefinition{
			Name:       matches[2],
			Type:       indexType,
			EntityType: "NODE",
			Label:      matches[4],
			Properties: parsePropertyList(matches[5], matches[3]),
			Options:    make(map[string]any),
		}
	}

	return nil
}

func parsePropertyList(propStr, varName string) []string {
	// Handle both "n.prop1, n.prop2" and "prop1, prop2" formats
	propStr = strings.TrimSpace(propStr)

	// Split by comma
	parts := strings.Split(propStr, ",")
	props := make([]string, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)

		// Remove variable prefix if present (e.g., "n.property" -> "property")
		if strings.Contains(part, ".") {
			parts := strings.SplitN(part, ".", 2)
			if len(parts) == 2 {
				part = parts[1]
			}
		} else if varName != "" && strings.HasPrefix(part, varName+".") {
			part = strings.TrimPrefix(part, varName+".")
		}

		if part != "" {
			props = append(props, part)
		}
	}

	return props
}

func buildTypeDefinitionsFromResults(constraints []ConstraintDefinition, indexes []IndexDefinition) ([]NodeTypeDefinition, []RelationshipTypeDefinition) {
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

			// Add properties
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
						Type:     "STRING",
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

			// Add properties
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
		}
	}

	// Convert to slices
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
