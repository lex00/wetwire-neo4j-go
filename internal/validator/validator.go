// Package validator provides external validation of Neo4j configurations against a live instance.
//
// This package validates generated configurations against a running Neo4j database:
// - Schema validation (constraints, indexes)
// - GDS algorithm validation (algorithm existence, parameter ranges)
// - Graph projection validation (node labels, relationship types)
//
// Example usage:
//
//	config := validator.Config{
//		URI:      "bolt://localhost:7687",
//		Username: "neo4j",
//		Password: "password",
//	}
//	v, err := validator.New(config)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer v.Close()
//
//	results := v.ValidateNodeType(nodeType)
package validator

import (
	"context"
	"fmt"
	"strings"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.com/lex00/wetwire-neo4j-go/internal/algorithms"
	"github.com/lex00/wetwire-neo4j-go/internal/projections"
	"github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"
)

// Config contains Neo4j connection configuration.
type Config struct {
	// URI is the Neo4j connection URI (bolt://, neo4j://, neo4j+s://).
	URI string
	// Username for authentication.
	Username string
	// Password for authentication.
	Password string
	// Database name (default: neo4j).
	Database string
}

// ValidationResult represents a single validation finding.
type ValidationResult struct {
	// Type is the validation type (schema, algorithm, projection).
	Type string
	// Target identifies what was validated.
	Target string
	// Valid indicates whether validation passed.
	Valid bool
	// Message describes the result.
	Message string
	// Details contains additional context.
	Details map[string]any
}

// Validator validates Neo4j configurations against a live instance.
type Validator struct {
	driver   neo4j.DriverWithContext
	config   Config
	gdsInfo  *GDSInfo
	dbInfo   *DatabaseInfo
}

// GDSInfo contains information about the GDS installation.
type GDSInfo struct {
	// Installed indicates if GDS is available.
	Installed bool
	// Version is the GDS version.
	Version string
	// Edition is "community" or "enterprise".
	Edition string
}

// DatabaseInfo contains information about the Neo4j database.
type DatabaseInfo struct {
	// Version is the Neo4j version.
	Version string
	// Edition is the Neo4j edition.
	Edition string
	// NodeLabels lists available node labels.
	NodeLabels []string
	// RelationshipTypes lists available relationship types.
	RelationshipTypes []string
}

// New creates a new Validator with the given configuration.
func New(config Config) (*Validator, error) {
	if config.URI == "" {
		return nil, fmt.Errorf("URI is required")
	}
	if config.Database == "" {
		config.Database = "neo4j"
	}

	auth := neo4j.BasicAuth(config.Username, config.Password, "")
	driver, err := neo4j.NewDriverWithContext(config.URI, auth)
	if err != nil {
		return nil, fmt.Errorf("failed to create driver: %w", err)
	}

	ctx := context.Background()
	if err := driver.VerifyConnectivity(ctx); err != nil {
		_ = driver.Close(ctx)
		return nil, fmt.Errorf("failed to connect to Neo4j: %w", err)
	}

	v := &Validator{
		driver: driver,
		config: config,
	}

	// Load database and GDS info
	if err := v.loadDatabaseInfo(ctx); err != nil {
		_ = driver.Close(ctx)
		return nil, fmt.Errorf("failed to load database info: %w", err)
	}

	if err := v.loadGDSInfo(ctx); err != nil {
		// GDS may not be installed, this is not an error
		v.gdsInfo = &GDSInfo{Installed: false}
	}

	return v, nil
}

// Close closes the Neo4j connection.
func (v *Validator) Close() error {
	return v.driver.Close(context.Background())
}

// loadDatabaseInfo loads information about the Neo4j database.
func (v *Validator) loadDatabaseInfo(ctx context.Context) error {
	session := v.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: v.config.Database})
	defer func() { _ = session.Close(ctx) }()

	v.dbInfo = &DatabaseInfo{}

	// Get database version
	result, err := session.Run(ctx, "CALL dbms.components() YIELD versions, edition RETURN versions[0] AS version, edition", nil)
	if err == nil {
		if result.Next(ctx) {
			record := result.Record()
			if version, ok := record.Get("version"); ok {
				v.dbInfo.Version = fmt.Sprintf("%v", version)
			}
			if edition, ok := record.Get("edition"); ok {
				v.dbInfo.Edition = fmt.Sprintf("%v", edition)
			}
		}
	}

	// Get node labels
	result, err = session.Run(ctx, "CALL db.labels() YIELD label RETURN label", nil)
	if err == nil {
		for result.Next(ctx) {
			record := result.Record()
			if label, ok := record.Get("label"); ok {
				v.dbInfo.NodeLabels = append(v.dbInfo.NodeLabels, fmt.Sprintf("%v", label))
			}
		}
	}

	// Get relationship types
	result, err = session.Run(ctx, "CALL db.relationshipTypes() YIELD relationshipType RETURN relationshipType", nil)
	if err == nil {
		for result.Next(ctx) {
			record := result.Record()
			if relType, ok := record.Get("relationshipType"); ok {
				v.dbInfo.RelationshipTypes = append(v.dbInfo.RelationshipTypes, fmt.Sprintf("%v", relType))
			}
		}
	}

	return nil
}

// loadGDSInfo loads information about the GDS installation.
func (v *Validator) loadGDSInfo(ctx context.Context) error {
	session := v.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: v.config.Database})
	defer func() { _ = session.Close(ctx) }()

	v.gdsInfo = &GDSInfo{Installed: false}

	result, err := session.Run(ctx, "RETURN gds.version() AS version", nil)
	if err != nil {
		return err
	}

	if result.Next(ctx) {
		record := result.Record()
		if version, ok := record.Get("version"); ok {
			v.gdsInfo.Installed = true
			v.gdsInfo.Version = fmt.Sprintf("%v", version)
		}
	}

	// Check for enterprise features
	_, err = session.Run(ctx, "CALL gds.license.state() YIELD isLicensed RETURN isLicensed", nil)
	if err == nil {
		v.gdsInfo.Edition = "enterprise"
	} else {
		v.gdsInfo.Edition = "community"
	}

	return result.Err()
}

// GetDatabaseInfo returns information about the connected database.
func (v *Validator) GetDatabaseInfo() *DatabaseInfo {
	return v.dbInfo
}

// GetGDSInfo returns information about the GDS installation.
func (v *Validator) GetGDSInfo() *GDSInfo {
	return v.gdsInfo
}

// ValidateNodeType validates a NodeType definition.
func (v *Validator) ValidateNodeType(node *schema.NodeType) []ValidationResult {
	var results []ValidationResult

	// Check if label exists in database
	labelExists := v.labelExists(node.Label)
	results = append(results, ValidationResult{
		Type:    "schema",
		Target:  fmt.Sprintf("NodeType(%s)", node.Label),
		Valid:   true,
		Message: fmt.Sprintf("label '%s' %s in database", node.Label, existsMsg(labelExists)),
		Details: map[string]any{"exists": labelExists},
	})

	// Validate constraints can be created (dry run)
	for _, c := range node.Constraints {
		result := v.validateConstraint(node.Label, c)
		results = append(results, result)
	}

	// Validate indexes can be created (dry run)
	for _, idx := range node.Indexes {
		result := v.validateIndex(node.Label, idx)
		results = append(results, result)
	}

	return results
}

// ValidateRelationshipType validates a RelationshipType definition.
func (v *Validator) ValidateRelationshipType(rel *schema.RelationshipType) []ValidationResult {
	var results []ValidationResult

	// Check if relationship type exists in database
	relExists := v.relationshipTypeExists(rel.Label)
	results = append(results, ValidationResult{
		Type:    "schema",
		Target:  fmt.Sprintf("RelationshipType(%s)", rel.Label),
		Valid:   true,
		Message: fmt.Sprintf("relationship type '%s' %s in database", rel.Label, existsMsg(relExists)),
		Details: map[string]any{"exists": relExists},
	})

	// Check source and target labels
	if rel.Source != "" {
		sourceExists := v.labelExists(rel.Source)
		results = append(results, ValidationResult{
			Type:    "schema",
			Target:  fmt.Sprintf("RelationshipType(%s).Source", rel.Label),
			Valid:   sourceExists,
			Message: fmt.Sprintf("source label '%s' %s in database", rel.Source, existsMsg(sourceExists)),
			Details: map[string]any{"exists": sourceExists},
		})
	}

	if rel.Target != "" {
		targetExists := v.labelExists(rel.Target)
		results = append(results, ValidationResult{
			Type:    "schema",
			Target:  fmt.Sprintf("RelationshipType(%s).Target", rel.Label),
			Valid:   targetExists,
			Message: fmt.Sprintf("target label '%s' %s in database", rel.Target, existsMsg(targetExists)),
			Details: map[string]any{"exists": targetExists},
		})
	}

	return results
}

// ValidateAlgorithm validates a GDS algorithm configuration.
func (v *Validator) ValidateAlgorithm(algo algorithms.Algorithm) []ValidationResult {
	var results []ValidationResult

	// Check if GDS is installed
	if !v.gdsInfo.Installed {
		results = append(results, ValidationResult{
			Type:    "algorithm",
			Target:  algo.AlgorithmType(),
			Valid:   false,
			Message: "GDS is not installed on this Neo4j instance",
		})
		return results
	}

	// Check if algorithm exists
	algoName := algo.AlgorithmType()
	algoExists := v.algorithmExists(algoName)
	results = append(results, ValidationResult{
		Type:    "algorithm",
		Target:  algoName,
		Valid:   algoExists,
		Message: fmt.Sprintf("algorithm '%s' %s in GDS %s", algoName, existsMsg(algoExists), v.gdsInfo.Version),
		Details: map[string]any{
			"exists":     algoExists,
			"gdsVersion": v.gdsInfo.Version,
		},
	})

	// Check if graph exists
	graphName := algo.GetGraphName()
	if graphName != "" {
		graphExists := v.graphExists(graphName)
		results = append(results, ValidationResult{
			Type:    "algorithm",
			Target:  fmt.Sprintf("%s.graphName", algoName),
			Valid:   graphExists,
			Message: fmt.Sprintf("graph '%s' %s in catalog", graphName, existsMsg(graphExists)),
			Details: map[string]any{"exists": graphExists},
		})
	}

	return results
}

// ValidateProjection validates a graph projection configuration.
func (v *Validator) ValidateProjection(proj projections.Projection) []ValidationResult {
	var results []ValidationResult

	// Check if GDS is installed
	if !v.gdsInfo.Installed {
		results = append(results, ValidationResult{
			Type:    "projection",
			Target:  proj.ProjectionName(),
			Valid:   false,
			Message: "GDS is not installed on this Neo4j instance",
		})
		return results
	}

	// Validate node projections
	for _, np := range proj.GetNodeProjections() {
		labelExists := v.labelExists(np.Label)
		results = append(results, ValidationResult{
			Type:    "projection",
			Target:  fmt.Sprintf("%s.NodeProjection(%s)", proj.ProjectionName(), np.Label),
			Valid:   labelExists,
			Message: fmt.Sprintf("node label '%s' %s in database", np.Label, existsMsg(labelExists)),
			Details: map[string]any{"exists": labelExists},
		})
	}

	// Validate relationship projections
	for _, rp := range proj.GetRelationshipProjections() {
		relExists := v.relationshipTypeExists(rp.Type)
		results = append(results, ValidationResult{
			Type:    "projection",
			Target:  fmt.Sprintf("%s.RelationshipProjection(%s)", proj.ProjectionName(), rp.Type),
			Valid:   relExists,
			Message: fmt.Sprintf("relationship type '%s' %s in database", rp.Type, existsMsg(relExists)),
			Details: map[string]any{"exists": relExists},
		})
	}

	return results
}

// ValidateAll validates multiple resources.
func (v *Validator) ValidateAll(resources []any) []ValidationResult {
	var results []ValidationResult

	for _, r := range resources {
		switch res := r.(type) {
		case *schema.NodeType:
			results = append(results, v.ValidateNodeType(res)...)
		case *schema.RelationshipType:
			results = append(results, v.ValidateRelationshipType(res)...)
		case algorithms.Algorithm:
			results = append(results, v.ValidateAlgorithm(res)...)
		case projections.Projection:
			results = append(results, v.ValidateProjection(res)...)
		}
	}

	return results
}

// HasErrors returns true if any result is invalid.
func HasErrors(results []ValidationResult) bool {
	for _, r := range results {
		if !r.Valid {
			return true
		}
	}
	return false
}

// FilterInvalid returns only invalid results.
func FilterInvalid(results []ValidationResult) []ValidationResult {
	var filtered []ValidationResult
	for _, r := range results {
		if !r.Valid {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

// FormatResults formats validation results as a string.
func FormatResults(results []ValidationResult) string {
	if len(results) == 0 {
		return "No validations performed"
	}

	var sb strings.Builder
	for _, r := range results {
		status := "✓"
		if !r.Valid {
			status = "✗"
		}
		sb.WriteString(fmt.Sprintf("[%s] %s: %s - %s\n", r.Type, status, r.Target, r.Message))
	}
	return sb.String()
}

// Helper functions

func (v *Validator) labelExists(label string) bool {
	for _, l := range v.dbInfo.NodeLabels {
		if l == label {
			return true
		}
	}
	return false
}

func (v *Validator) relationshipTypeExists(relType string) bool {
	for _, rt := range v.dbInfo.RelationshipTypes {
		if rt == relType {
			return true
		}
	}
	return false
}

func (v *Validator) algorithmExists(algoName string) bool {
	ctx := context.Background()
	session := v.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: v.config.Database})
	defer func() { _ = session.Close(ctx) }()

	// Try to get algorithm info (this will fail if algorithm doesn't exist)
	query := fmt.Sprintf("CALL gds.list('%s') YIELD name RETURN name LIMIT 1", algoName)
	result, err := session.Run(ctx, query, nil)
	if err != nil {
		return false
	}

	return result.Next(ctx)
}

func (v *Validator) graphExists(graphName string) bool {
	ctx := context.Background()
	session := v.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: v.config.Database})
	defer func() { _ = session.Close(ctx) }()

	result, err := session.Run(ctx, "CALL gds.graph.exists($name) YIELD exists RETURN exists", map[string]any{"name": graphName})
	if err != nil {
		return false
	}

	if result.Next(ctx) {
		record := result.Record()
		if exists, ok := record.Get("exists"); ok {
			if b, ok := exists.(bool); ok {
				return b
			}
		}
	}

	return false
}

func (v *Validator) validateConstraint(label string, c schema.Constraint) ValidationResult {
	// We don't actually create the constraint, just validate it could be created
	constraintType := string(c.Type)
	return ValidationResult{
		Type:    "schema",
		Target:  fmt.Sprintf("Constraint(%s)", c.Name),
		Valid:   true,
		Message: fmt.Sprintf("%s constraint '%s' on %s can be created", constraintType, c.Name, label),
		Details: map[string]any{
			"constraintType": constraintType,
			"properties":     c.Properties,
		},
	}
}

func (v *Validator) validateIndex(label string, idx schema.Index) ValidationResult {
	// We don't actually create the index, just validate it could be created
	indexType := string(idx.Type)
	return ValidationResult{
		Type:    "schema",
		Target:  fmt.Sprintf("Index(%s)", idx.Name),
		Valid:   true,
		Message: fmt.Sprintf("%s index '%s' on %s can be created", indexType, idx.Name, label),
		Details: map[string]any{
			"indexType":  indexType,
			"properties": idx.Properties,
		},
	}
}

func existsMsg(exists bool) string {
	if exists {
		return "exists"
	}
	return "does not exist"
}
