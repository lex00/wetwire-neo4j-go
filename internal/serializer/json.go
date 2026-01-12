package serializer

import (
	"encoding/json"

	"github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"
)

// JSONSerializer serializes schema definitions to JSON format.
type JSONSerializer struct{}

// NewJSONSerializer creates a new JSON serializer.
func NewJSONSerializer() *JSONSerializer {
	return &JSONSerializer{}
}

// NodeTypeJSON represents a NodeType in JSON format.
type NodeTypeJSON struct {
	Label       string           `json:"label"`
	Description string           `json:"description,omitempty"`
	Properties  []PropertyJSON   `json:"properties,omitempty"`
	Constraints []ConstraintJSON `json:"constraints,omitempty"`
	Indexes     []IndexJSON      `json:"indexes,omitempty"`
}

// PropertyJSON represents a Property in JSON format.
type PropertyJSON struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Required     bool   `json:"required,omitempty"`
	Unique       bool   `json:"unique,omitempty"`
	Description  string `json:"description,omitempty"`
	DefaultValue any    `json:"defaultValue,omitempty"`
}

// ConstraintJSON represents a Constraint in JSON format.
type ConstraintJSON struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	Properties []string `json:"properties"`
}

// IndexJSON represents an Index in JSON format.
type IndexJSON struct {
	Name       string         `json:"name"`
	Type       string         `json:"type"`
	Properties []string       `json:"properties"`
	Options    map[string]any `json:"options,omitempty"`
}

// RelationshipTypeJSON represents a RelationshipType in JSON format.
type RelationshipTypeJSON struct {
	Label       string           `json:"label"`
	Source      string           `json:"source"`
	Target      string           `json:"target"`
	Cardinality string           `json:"cardinality,omitempty"`
	Description string           `json:"description,omitempty"`
	Properties  []PropertyJSON   `json:"properties,omitempty"`
	Constraints []ConstraintJSON `json:"constraints,omitempty"`
}

// SchemaJSON represents the full schema in JSON format.
type SchemaJSON struct {
	NodeTypes         []NodeTypeJSON         `json:"nodeTypes,omitempty"`
	RelationshipTypes []RelationshipTypeJSON `json:"relationshipTypes,omitempty"`
}

// SerializeNodeType serializes a NodeType to JSON.
func (s *JSONSerializer) SerializeNodeType(n *schema.NodeType) ([]byte, error) {
	jsonNode := s.convertNodeType(n)
	return json.MarshalIndent(jsonNode, "", "  ")
}

// SerializeRelationshipType serializes a RelationshipType to JSON.
func (s *JSONSerializer) SerializeRelationshipType(r *schema.RelationshipType) ([]byte, error) {
	jsonRel := s.convertRelationshipType(r)
	return json.MarshalIndent(jsonRel, "", "  ")
}

// SerializeAll serializes all node types and relationship types to JSON.
func (s *JSONSerializer) SerializeAll(nodeTypes []*schema.NodeType, relTypes []*schema.RelationshipType) ([]byte, error) {
	schemaJSON := SchemaJSON{
		NodeTypes:         make([]NodeTypeJSON, 0, len(nodeTypes)),
		RelationshipTypes: make([]RelationshipTypeJSON, 0, len(relTypes)),
	}

	for _, n := range nodeTypes {
		schemaJSON.NodeTypes = append(schemaJSON.NodeTypes, s.convertNodeType(n))
	}

	for _, r := range relTypes {
		schemaJSON.RelationshipTypes = append(schemaJSON.RelationshipTypes, s.convertRelationshipType(r))
	}

	return json.MarshalIndent(schemaJSON, "", "  ")
}

// convertNodeType converts a schema.NodeType to NodeTypeJSON.
func (s *JSONSerializer) convertNodeType(n *schema.NodeType) NodeTypeJSON {
	jsonNode := NodeTypeJSON{
		Label:       n.Label,
		Description: n.Description,
		Properties:  make([]PropertyJSON, 0, len(n.Properties)),
		Constraints: make([]ConstraintJSON, 0, len(n.Constraints)),
		Indexes:     make([]IndexJSON, 0, len(n.Indexes)),
	}

	for _, p := range n.Properties {
		jsonNode.Properties = append(jsonNode.Properties, PropertyJSON{
			Name:         p.Name,
			Type:         string(p.Type),
			Required:     p.Required,
			Unique:       p.Unique,
			Description:  p.Description,
			DefaultValue: p.DefaultValue,
		})
	}

	for _, c := range n.Constraints {
		jsonNode.Constraints = append(jsonNode.Constraints, ConstraintJSON{
			Name:       c.Name,
			Type:       string(c.Type),
			Properties: c.Properties,
		})
	}

	for _, idx := range n.Indexes {
		jsonNode.Indexes = append(jsonNode.Indexes, IndexJSON{
			Name:       idx.Name,
			Type:       string(idx.Type),
			Properties: idx.Properties,
			Options:    idx.Options,
		})
	}

	return jsonNode
}

// convertRelationshipType converts a schema.RelationshipType to RelationshipTypeJSON.
func (s *JSONSerializer) convertRelationshipType(r *schema.RelationshipType) RelationshipTypeJSON {
	jsonRel := RelationshipTypeJSON{
		Label:       r.Label,
		Source:      r.Source,
		Target:      r.Target,
		Cardinality: string(r.Cardinality),
		Description: r.Description,
		Properties:  make([]PropertyJSON, 0, len(r.Properties)),
		Constraints: make([]ConstraintJSON, 0, len(r.Constraints)),
	}

	for _, p := range r.Properties {
		jsonRel.Properties = append(jsonRel.Properties, PropertyJSON{
			Name:         p.Name,
			Type:         string(p.Type),
			Required:     p.Required,
			Unique:       p.Unique,
			Description:  p.Description,
			DefaultValue: p.DefaultValue,
		})
	}

	for _, c := range r.Constraints {
		jsonRel.Constraints = append(jsonRel.Constraints, ConstraintJSON{
			Name:       c.Name,
			Type:       string(c.Type),
			Properties: c.Properties,
		})
	}

	return jsonRel
}

// ToMap converts a NodeType to a map for flexible serialization.
func (s *JSONSerializer) NodeTypeToMap(n *schema.NodeType) map[string]any {
	result := map[string]any{
		"label": n.Label,
	}

	if n.Description != "" {
		result["description"] = n.Description
	}

	if len(n.Properties) > 0 {
		props := make([]map[string]any, 0, len(n.Properties))
		for _, p := range n.Properties {
			prop := map[string]any{
				"name": p.Name,
				"type": string(p.Type),
			}
			if p.Required {
				prop["required"] = true
			}
			if p.Unique {
				prop["unique"] = true
			}
			if p.Description != "" {
				prop["description"] = p.Description
			}
			if p.DefaultValue != nil {
				prop["defaultValue"] = p.DefaultValue
			}
			props = append(props, prop)
		}
		result["properties"] = props
	}

	return result
}

// ToMap converts a RelationshipType to a map for flexible serialization.
func (s *JSONSerializer) RelationshipTypeToMap(r *schema.RelationshipType) map[string]any {
	result := map[string]any{
		"label":  r.Label,
		"source": r.Source,
		"target": r.Target,
	}

	if r.Cardinality != "" {
		result["cardinality"] = string(r.Cardinality)
	}

	if r.Description != "" {
		result["description"] = r.Description
	}

	if len(r.Properties) > 0 {
		props := make([]map[string]any, 0, len(r.Properties))
		for _, p := range r.Properties {
			prop := map[string]any{
				"name": p.Name,
				"type": string(p.Type),
			}
			if p.Required {
				prop["required"] = true
			}
			if p.Description != "" {
				prop["description"] = p.Description
			}
			props = append(props, prop)
		}
		result["properties"] = props
	}

	return result
}
