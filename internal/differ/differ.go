// Package differ provides semantic comparison of Neo4j graph schemas.
package differ

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	coredomain "github.com/lex00/wetwire-core-go/domain"
	"github.com/lex00/wetwire-neo4j-go/internal/discover"
)

// SchemaDiffer implements coredomain.Differ for Neo4j graph schemas.
type SchemaDiffer struct{}

// Compile-time check that SchemaDiffer implements Differ.
var _ coredomain.Differ = (*SchemaDiffer)(nil)

// New creates a new schema differ.
func New() *SchemaDiffer {
	return &SchemaDiffer{}
}

// Diff compares two Neo4j schema directories/files and returns differences.
// It detects breaking changes that would require data migration.
func (d *SchemaDiffer) Diff(ctx *coredomain.Context, file1, file2 string, opts coredomain.DiffOpts) (*coredomain.DiffResult, error) {
	// Check if paths are files or directories
	info1, err := os.Stat(file1)
	if err != nil {
		return nil, fmt.Errorf("path1: %w", err)
	}
	info2, err := os.Stat(file2)
	if err != nil {
		return nil, fmt.Errorf("path2: %w", err)
	}

	var res1, res2 []discover.DiscoveredResource

	if info1.IsDir() && info2.IsDir() {
		// Scan directories
		scanner := discover.NewScanner()
		res1, err = scanner.ScanDir(file1)
		if err != nil {
			return nil, fmt.Errorf("scan dir1: %w", err)
		}
		res2, err = scanner.ScanDir(file2)
		if err != nil {
			return nil, fmt.Errorf("scan dir2: %w", err)
		}
	} else if !info1.IsDir() && !info2.IsDir() {
		// Try to parse as JSON schema files first
		ext1 := strings.ToLower(filepath.Ext(file1))
		ext2 := strings.ToLower(filepath.Ext(file2))

		if ext1 == ".json" && ext2 == ".json" {
			return d.diffJSONSchemas(file1, file2, opts)
		}

		// Otherwise try to scan as Go files
		scanner := discover.NewScanner()
		res1, err = scanner.ScanFile(file1)
		if err != nil {
			return nil, fmt.Errorf("scan file1: %w", err)
		}
		res2, err = scanner.ScanFile(file2)
		if err != nil {
			return nil, fmt.Errorf("scan file2: %w", err)
		}
	} else {
		return nil, fmt.Errorf("cannot compare directory with file")
	}

	return compareSchemas(res1, res2, opts)
}

// diffJSONSchemas compares two JSON schema files.
func (d *SchemaDiffer) diffJSONSchemas(file1, file2 string, opts coredomain.DiffOpts) (*coredomain.DiffResult, error) {
	data1, err := os.ReadFile(file1)
	if err != nil {
		return nil, fmt.Errorf("read file1: %w", err)
	}
	data2, err := os.ReadFile(file2)
	if err != nil {
		return nil, fmt.Errorf("read file2: %w", err)
	}

	var schema1, schema2 map[string]interface{}
	if err := json.Unmarshal(data1, &schema1); err != nil {
		return nil, fmt.Errorf("parse file1: %w", err)
	}
	if err := json.Unmarshal(data2, &schema2); err != nil {
		return nil, fmt.Errorf("parse file2: %w", err)
	}

	return compareJSONSchemas(schema1, schema2, opts)
}

// compareSchemas compares two sets of discovered resources.
func compareSchemas(res1, res2 []discover.DiscoveredResource, opts coredomain.DiffOpts) (*coredomain.DiffResult, error) {
	result := &coredomain.DiffResult{}

	// Build maps by name and kind
	map1 := buildResourceMap(res1)
	map2 := buildResourceMap(res2)

	// Find added resources
	for key, r := range map2 {
		if _, exists := map1[key]; !exists {
			changes := describeResource(r)
			// Check if adding new resource is a breaking change
			if hasBreakingAddition(r) {
				changes = append(changes, "[BREAKING: requires data population]")
			}
			result.Entries = append(result.Entries, coredomain.DiffEntry{
				Resource: r.Name,
				Type:     string(r.Kind),
				Action:   "added",
				Changes:  changes,
			})
		}
	}

	// Find removed resources
	for key, r := range map1 {
		if _, exists := map2[key]; !exists {
			changes := []string{"[BREAKING: existing data will be orphaned]"}
			result.Entries = append(result.Entries, coredomain.DiffEntry{
				Resource: r.Name,
				Type:     string(r.Kind),
				Action:   "removed",
				Changes:  changes,
			})
		}
	}

	// Find modified resources
	for key, r1 := range map1 {
		if r2, exists := map2[key]; exists {
			changes := compareResources(r1, r2, opts)
			if len(changes) > 0 {
				result.Entries = append(result.Entries, coredomain.DiffEntry{
					Resource: r1.Name,
					Type:     string(r1.Kind),
					Action:   "modified",
					Changes:  changes,
				})
			}
		}
	}

	// Sort entries for consistent output
	sort.Slice(result.Entries, func(i, j int) bool {
		if result.Entries[i].Action != result.Entries[j].Action {
			order := map[string]int{"added": 0, "modified": 1, "removed": 2}
			return order[result.Entries[i].Action] < order[result.Entries[j].Action]
		}
		return result.Entries[i].Resource < result.Entries[j].Resource
	})

	// Calculate summary
	for _, e := range result.Entries {
		switch e.Action {
		case "added":
			result.Summary.Added++
		case "removed":
			result.Summary.Removed++
		case "modified":
			result.Summary.Modified++
		}
	}
	result.Summary.Total = result.Summary.Added + result.Summary.Removed + result.Summary.Modified

	return result, nil
}

// buildResourceMap creates a map of resources by name+kind.
func buildResourceMap(resources []discover.DiscoveredResource) map[string]discover.DiscoveredResource {
	m := make(map[string]discover.DiscoveredResource)
	for _, r := range resources {
		key := fmt.Sprintf("%s:%s", r.Kind, r.Name)
		m[key] = r
	}
	return m
}

// describeResource returns a description of a resource.
func describeResource(r discover.DiscoveredResource) []string {
	var desc []string
	if len(r.Properties) > 0 {
		desc = append(desc, fmt.Sprintf("%d properties", len(r.Properties)))
	}
	if len(r.Constraints) > 0 {
		desc = append(desc, fmt.Sprintf("%d constraints", len(r.Constraints)))
	}
	if len(r.Indexes) > 0 {
		desc = append(desc, fmt.Sprintf("%d indexes", len(r.Indexes)))
	}
	if r.Source != "" && r.Target != "" {
		desc = append(desc, fmt.Sprintf("(%s)-[]->(%s)", r.Source, r.Target))
	}
	return desc
}

// hasBreakingAddition checks if adding a resource is a breaking change.
func hasBreakingAddition(r discover.DiscoveredResource) bool {
	// Adding a new node type or relationship type with required properties
	// is a breaking change because existing data won't have those properties
	for _, prop := range r.Properties {
		if prop.Required {
			return true
		}
	}
	return false
}

// compareResources compares two resources and returns changes.
func compareResources(r1, r2 discover.DiscoveredResource, opts coredomain.DiffOpts) []string {
	var changes []string

	// Compare properties
	propChanges := compareProperties(r1.Properties, r2.Properties)
	changes = append(changes, propChanges...)

	// Compare constraints (only for NodeType)
	if r1.Kind == discover.KindNodeType {
		constraintChanges := compareConstraints(r1.Constraints, r2.Constraints)
		changes = append(changes, constraintChanges...)

		// Compare indexes
		indexChanges := compareIndexes(r1.Indexes, r2.Indexes)
		changes = append(changes, indexChanges...)
	}

	// Compare source/target for relationships
	if r1.Kind == discover.KindRelationshipType {
		if r1.Source != r2.Source {
			changes = append(changes, fmt.Sprintf("source changed: %s → %s [BREAKING]", r1.Source, r2.Source))
		}
		if r1.Target != r2.Target {
			changes = append(changes, fmt.Sprintf("target changed: %s → %s [BREAKING]", r1.Target, r2.Target))
		}
	}

	// Compare AgentContext for Schema
	if r1.Kind == discover.KindSchema {
		if r1.AgentContext != r2.AgentContext {
			changes = append(changes, "agentContext changed")
		}
	}

	sort.Strings(changes)
	return changes
}

// compareProperties compares two property lists.
func compareProperties(props1, props2 []discover.PropertyInfo) []string {
	var changes []string

	// Build maps
	map1 := make(map[string]discover.PropertyInfo)
	for _, p := range props1 {
		map1[p.Name] = p
	}
	map2 := make(map[string]discover.PropertyInfo)
	for _, p := range props2 {
		map2[p.Name] = p
	}

	// Find added properties
	for name, p := range map2 {
		if _, exists := map1[name]; !exists {
			if p.Required {
				changes = append(changes, fmt.Sprintf("property %q added (required) [BREAKING: requires data migration]", name))
			} else {
				changes = append(changes, fmt.Sprintf("property %q added", name))
			}
		}
	}

	// Find removed properties
	for name := range map1 {
		if _, exists := map2[name]; !exists {
			changes = append(changes, fmt.Sprintf("property %q removed [BREAKING: data loss]", name))
		}
	}

	// Find modified properties
	for name, p1 := range map1 {
		if p2, exists := map2[name]; exists {
			if p1.Type != p2.Type {
				changes = append(changes, fmt.Sprintf("property %q type changed: %s → %s [BREAKING]", name, p1.Type, p2.Type))
			}
			if p1.Required != p2.Required {
				if p2.Required {
					changes = append(changes, fmt.Sprintf("property %q now required [BREAKING: existing null values invalid]", name))
				} else {
					changes = append(changes, fmt.Sprintf("property %q now optional", name))
				}
			}
		}
	}

	return changes
}

// compareConstraints compares two constraint lists.
func compareConstraints(cons1, cons2 []discover.ConstraintInfo) []string {
	var changes []string

	// Build maps by constraint key (type + properties)
	makeKey := func(c discover.ConstraintInfo) string {
		return fmt.Sprintf("%s:%v", c.Type, c.Properties)
	}

	map1 := make(map[string]discover.ConstraintInfo)
	for _, c := range cons1 {
		map1[makeKey(c)] = c
	}
	map2 := make(map[string]discover.ConstraintInfo)
	for _, c := range cons2 {
		map2[makeKey(c)] = c
	}

	// Find added constraints
	for key, c := range map2 {
		if _, exists := map1[key]; !exists {
			changes = append(changes, fmt.Sprintf("constraint %s(%v) added [may fail if existing data violates]", c.Type, c.Properties))
		}
	}

	// Find removed constraints
	for key, c := range map1 {
		if _, exists := map2[key]; !exists {
			changes = append(changes, fmt.Sprintf("constraint %s(%v) removed [BREAKING: requires migration]", c.Type, c.Properties))
		}
	}

	return changes
}

// compareIndexes compares two index lists.
func compareIndexes(idx1, idx2 []discover.IndexInfo) []string {
	var changes []string

	// Build maps by index key (type + properties)
	makeKey := func(i discover.IndexInfo) string {
		return fmt.Sprintf("%s:%v", i.Type, i.Properties)
	}

	map1 := make(map[string]discover.IndexInfo)
	for _, i := range idx1 {
		map1[makeKey(i)] = i
	}
	map2 := make(map[string]discover.IndexInfo)
	for _, i := range idx2 {
		map2[makeKey(i)] = i
	}

	// Find added indexes
	for key, i := range map2 {
		if _, exists := map1[key]; !exists {
			changes = append(changes, fmt.Sprintf("index %s(%v) added", i.Type, i.Properties))
		}
	}

	// Find removed indexes
	for key, i := range map1 {
		if _, exists := map2[key]; !exists {
			changes = append(changes, fmt.Sprintf("index %s(%v) removed [performance impact]", i.Type, i.Properties))
		}
	}

	return changes
}

// compareJSONSchemas compares two JSON schema objects.
func compareJSONSchemas(schema1, schema2 map[string]interface{}, opts coredomain.DiffOpts) (*coredomain.DiffResult, error) {
	result := &coredomain.DiffResult{}

	// Compare nodeTypes
	if nodes1, ok := schema1["nodeTypes"].([]interface{}); ok {
		nodes2, _ := schema2["nodeTypes"].([]interface{})
		nodeChanges := compareJSONResourceList("NodeType", nodes1, nodes2, opts)
		result.Entries = append(result.Entries, nodeChanges...)
	}

	// Compare relationshipTypes
	if rels1, ok := schema1["relationshipTypes"].([]interface{}); ok {
		rels2, _ := schema2["relationshipTypes"].([]interface{})
		relChanges := compareJSONResourceList("RelationshipType", rels1, rels2, opts)
		result.Entries = append(result.Entries, relChanges...)
	}

	// Compare algorithms
	if algos1, ok := schema1["algorithms"].([]interface{}); ok {
		algos2, _ := schema2["algorithms"].([]interface{})
		algoChanges := compareJSONResourceList("Algorithm", algos1, algos2, opts)
		result.Entries = append(result.Entries, algoChanges...)
	}

	// Compare pipelines
	if pipes1, ok := schema1["pipelines"].([]interface{}); ok {
		pipes2, _ := schema2["pipelines"].([]interface{})
		pipeChanges := compareJSONResourceList("Pipeline", pipes1, pipes2, opts)
		result.Entries = append(result.Entries, pipeChanges...)
	}

	// Compare retrievers
	if rets1, ok := schema1["retrievers"].([]interface{}); ok {
		rets2, _ := schema2["retrievers"].([]interface{})
		retChanges := compareJSONResourceList("Retriever", rets1, rets2, opts)
		result.Entries = append(result.Entries, retChanges...)
	}

	// Sort entries
	sort.Slice(result.Entries, func(i, j int) bool {
		if result.Entries[i].Action != result.Entries[j].Action {
			order := map[string]int{"added": 0, "modified": 1, "removed": 2}
			return order[result.Entries[i].Action] < order[result.Entries[j].Action]
		}
		return result.Entries[i].Resource < result.Entries[j].Resource
	})

	// Calculate summary
	for _, e := range result.Entries {
		switch e.Action {
		case "added":
			result.Summary.Added++
		case "removed":
			result.Summary.Removed++
		case "modified":
			result.Summary.Modified++
		}
	}
	result.Summary.Total = result.Summary.Added + result.Summary.Removed + result.Summary.Modified

	return result, nil
}

// compareJSONResourceList compares two lists of JSON resources.
func compareJSONResourceList(resourceType string, list1, list2 []interface{}, opts coredomain.DiffOpts) []coredomain.DiffEntry {
	var entries []coredomain.DiffEntry

	// Build maps by name
	map1 := make(map[string]map[string]interface{})
	for _, item := range list1 {
		if m, ok := item.(map[string]interface{}); ok {
			if name, ok := m["name"].(string); ok {
				map1[name] = m
			}
		}
	}
	map2 := make(map[string]map[string]interface{})
	for _, item := range list2 {
		if m, ok := item.(map[string]interface{}); ok {
			if name, ok := m["name"].(string); ok {
				map2[name] = m
			}
		}
	}

	// Find added
	for name := range map2 {
		if _, exists := map1[name]; !exists {
			entries = append(entries, coredomain.DiffEntry{
				Resource: name,
				Type:     resourceType,
				Action:   "added",
			})
		}
	}

	// Find removed
	for name := range map1 {
		if _, exists := map2[name]; !exists {
			entries = append(entries, coredomain.DiffEntry{
				Resource: name,
				Type:     resourceType,
				Action:   "removed",
				Changes:  []string{"[BREAKING]"},
			})
		}
	}

	// Find modified
	for name, m1 := range map1 {
		if m2, exists := map2[name]; exists {
			if !reflect.DeepEqual(m1, m2) {
				entries = append(entries, coredomain.DiffEntry{
					Resource: name,
					Type:     resourceType,
					Action:   "modified",
					Changes:  findMapChanges(m1, m2),
				})
			}
		}
	}

	return entries
}

// findMapChanges finds differences between two maps.
func findMapChanges(m1, m2 map[string]interface{}) []string {
	var changes []string

	// Find added/modified keys
	for k, v2 := range m2 {
		if v1, exists := m1[k]; exists {
			if !reflect.DeepEqual(v1, v2) {
				changes = append(changes, fmt.Sprintf("%s changed", k))
			}
		} else {
			changes = append(changes, fmt.Sprintf("%s added", k))
		}
	}

	// Find removed keys
	for k := range m1 {
		if _, exists := m2[k]; !exists {
			changes = append(changes, fmt.Sprintf("%s removed", k))
		}
	}

	sort.Strings(changes)
	return changes
}
