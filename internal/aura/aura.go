// Package aura provides Aura Graph Analytics session configuration.
//
// Aura Graph Analytics is Neo4j's serverless GDS offering that allows
// running graph algorithms on data without a persistent Neo4j database.
//
// Example usage:
//
//	session := &aura.Session{
//		Name:     "analytics-session",
//		TTLHours: 24,
//		DataSource: &aura.PandasDataSource{
//			NodeFiles: []string{"nodes.csv"},
//			RelFiles:  []string{"rels.csv"},
//		},
//		Projection: &projections.DataFrameProjection{...},
//		Algorithms: []algorithms.Algorithm{...},
//	}
package aura

import (
	"fmt"

	"github.com/lex00/wetwire-neo4j-go/internal/algorithms"
	"github.com/lex00/wetwire-neo4j-go/internal/projections"
)

// Session represents an Aura Graph Analytics session.
type Session struct {
	// Name is the session identifier.
	Name string
	// TTLHours is the session time-to-live in hours.
	TTLHours int
	// DataSource specifies where to load data from.
	DataSource DataSource
	// Projection defines how to project the graph from DataFrames.
	Projection *projections.DataFrameProjection
	// Algorithms are the GDS algorithms to execute.
	Algorithms []algorithms.Algorithm
}

// Validate checks the session configuration for errors.
func (s *Session) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("session name is required")
	}
	if s.TTLHours <= 0 {
		return fmt.Errorf("TTLHours must be positive")
	}
	if s.DataSource == nil {
		return fmt.Errorf("data source is required")
	}
	return nil
}

// DataSource is the interface for data source configurations.
type DataSource interface {
	// SourceType returns the data source type identifier.
	SourceType() string
}

// PandasDataSource configures loading data from Pandas DataFrames.
type PandasDataSource struct {
	// NodeFiles are CSV files containing node data.
	NodeFiles []string
	// RelFiles are CSV files containing relationship data.
	RelFiles []string
}

// SourceType returns "pandas".
func (p *PandasDataSource) SourceType() string {
	return "pandas"
}

// SnowflakeDataSource configures loading data from Snowflake.
type SnowflakeDataSource struct {
	// Account is the Snowflake account identifier.
	Account string
	// User is the Snowflake username.
	User string
	// Database is the Snowflake database name.
	Database string
	// Schema is the Snowflake schema name.
	Schema string
	// Warehouse is the Snowflake warehouse to use.
	Warehouse string
	// NodeQuery is the SQL query to fetch node data.
	NodeQuery string
	// RelQuery is the SQL query to fetch relationship data.
	RelQuery string
}

// SourceType returns "snowflake".
func (s *SnowflakeDataSource) SourceType() string {
	return "snowflake"
}

// BigQueryDataSource configures loading data from Google BigQuery.
type BigQueryDataSource struct {
	// Project is the GCP project ID.
	Project string
	// Dataset is the BigQuery dataset name.
	Dataset string
	// NodeQuery is the SQL query to fetch node data.
	NodeQuery string
	// RelQuery is the SQL query to fetch relationship data.
	RelQuery string
}

// SourceType returns "bigquery".
func (b *BigQueryDataSource) SourceType() string {
	return "bigquery"
}
