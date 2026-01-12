package aura

import (
	"fmt"
	"strings"

	"github.com/lex00/wetwire-neo4j-go/internal/algorithms"
)

// Serializer serializes Aura sessions to Python code and JSON.
type Serializer struct{}

// NewSerializer creates a new Serializer.
func NewSerializer() *Serializer {
	return &Serializer{}
}

// ToPython generates Python code for the graphdatascience client.
func (s *Serializer) ToPython(session *Session) (string, error) {
	var sb strings.Builder

	// Imports
	sb.WriteString("from graphdatascience import GdsSessions\n")
	sb.WriteString("import pandas as pd\n")
	sb.WriteString("\n")

	// Data source specific imports and setup
	switch ds := session.DataSource.(type) {
	case *PandasDataSource:
		s.writePandasSetup(&sb, ds)
	case *SnowflakeDataSource:
		s.writeSnowflakeSetup(&sb, ds)
	case *BigQueryDataSource:
		s.writeBigQuerySetup(&sb, ds)
	}

	sb.WriteString("\n")

	// Session creation
	sb.WriteString("# Create Aura Analytics session\n")
	sb.WriteString("sessions = GdsSessions()\n")
	sb.WriteString("gds = sessions.create(\n")
	fmt.Fprintf(&sb, "    name=%q,\n", session.Name)
	fmt.Fprintf(&sb, "    ttl_hours=%d,\n", session.TTLHours)
	sb.WriteString(")\n\n")

	// Graph projection
	if session.Projection != nil {
		sb.WriteString("# Project graph from DataFrames\n")
		sb.WriteString("G = gds.graph.construct(\n")
		fmt.Fprintf(&sb, "    graph_name=%q,\n", session.Projection.ProjectionName())
		sb.WriteString("    nodes=nodes_df,\n")
		sb.WriteString("    relationships=rels_df,\n")
		sb.WriteString(")\n\n")
	}

	// Algorithms
	if len(session.Algorithms) > 0 {
		sb.WriteString("# Run algorithms\n")
		for _, algo := range session.Algorithms {
			s.writeAlgorithmCall(&sb, algo)
		}
	}

	return sb.String(), nil
}

func (s *Serializer) writePandasSetup(sb *strings.Builder, ds *PandasDataSource) {
	sb.WriteString("# Load data from CSV files\n")
	if len(ds.NodeFiles) > 0 {
		fmt.Fprintf(sb, "nodes_df = pd.read_csv(%q)\n", ds.NodeFiles[0])
	}
	if len(ds.RelFiles) > 0 {
		fmt.Fprintf(sb, "rels_df = pd.read_csv(%q)\n", ds.RelFiles[0])
	}
}

func (s *Serializer) writeSnowflakeSetup(sb *strings.Builder, ds *SnowflakeDataSource) {
	sb.WriteString("import snowflake.connector\n\n")
	sb.WriteString("# Connect to Snowflake\n")
	sb.WriteString("conn = snowflake.connector.connect(\n")
	fmt.Fprintf(sb, "    account=%q,\n", ds.Account)
	fmt.Fprintf(sb, "    user=%q,\n", ds.User)
	fmt.Fprintf(sb, "    database=%q,\n", ds.Database)
	fmt.Fprintf(sb, "    schema=%q,\n", ds.Schema)
	fmt.Fprintf(sb, "    warehouse=%q,\n", ds.Warehouse)
	sb.WriteString(")\n\n")
	sb.WriteString("# Load data\n")
	fmt.Fprintf(sb, "nodes_df = pd.read_sql(%q, conn)\n", ds.NodeQuery)
	fmt.Fprintf(sb, "rels_df = pd.read_sql(%q, conn)\n", ds.RelQuery)
}

func (s *Serializer) writeBigQuerySetup(sb *strings.Builder, ds *BigQueryDataSource) {
	sb.WriteString("from google.cloud import bigquery\n\n")
	sb.WriteString("# Connect to BigQuery\n")
	fmt.Fprintf(sb, "client = bigquery.Client(project=%q)\n\n", ds.Project)
	sb.WriteString("# Load data\n")
	fmt.Fprintf(sb, "nodes_df = client.query(%q).to_dataframe()\n", ds.NodeQuery)
	fmt.Fprintf(sb, "rels_df = client.query(%q).to_dataframe()\n", ds.RelQuery)
}

func (s *Serializer) writeAlgorithmCall(sb *strings.Builder, algo algorithms.Algorithm) {
	algoType := algo.AlgorithmType()
	pythonName := toSnakeCase(algoType)

	fmt.Fprintf(sb, "result = gds.%s.stream(G)\n", pythonName)
	sb.WriteString("print(result)\n\n")
}

// ToMap converts a session to a map for JSON serialization.
func (s *Serializer) ToMap(session *Session) map[string]any {
	result := map[string]any{
		"name":     session.Name,
		"ttlHours": session.TTLHours,
	}

	if session.DataSource != nil {
		result["dataSourceType"] = session.DataSource.SourceType()
		result["dataSource"] = s.dataSourceToMap(session.DataSource)
	}

	if session.Projection != nil {
		result["projection"] = map[string]any{
			"name": session.Projection.ProjectionName(),
		}
	}

	if len(session.Algorithms) > 0 {
		algos := make([]map[string]any, len(session.Algorithms))
		for i, algo := range session.Algorithms {
			algos[i] = map[string]any{
				"type":      algo.AlgorithmType(),
				"graphName": algo.GetGraphName(),
			}
		}
		result["algorithms"] = algos
	}

	return result
}

func (s *Serializer) dataSourceToMap(ds DataSource) map[string]any {
	switch d := ds.(type) {
	case *PandasDataSource:
		return map[string]any{
			"nodeFiles": d.NodeFiles,
			"relFiles":  d.RelFiles,
		}
	case *SnowflakeDataSource:
		return map[string]any{
			"account":   d.Account,
			"user":      d.User,
			"database":  d.Database,
			"schema":    d.Schema,
			"warehouse": d.Warehouse,
			"nodeQuery": d.NodeQuery,
			"relQuery":  d.RelQuery,
		}
	case *BigQueryDataSource:
		return map[string]any{
			"project":   d.Project,
			"dataset":   d.Dataset,
			"nodeQuery": d.NodeQuery,
			"relQuery":  d.RelQuery,
		}
	default:
		return nil
	}
}

// toSnakeCase converts PascalCase to snake_case.
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}
