package aura

import (
	"strings"
	"testing"

	"github.com/lex00/wetwire-neo4j-go/internal/algorithms"
	"github.com/lex00/wetwire-neo4j-go/internal/projections"
)

func TestAuraSession_Basic(t *testing.T) {
	session := &Session{
		Name:     "analytics-session",
		TTLHours: 24,
		DataSource: &PandasDataSource{
			NodeFiles: []string{"nodes.csv"},
			RelFiles:  []string{"rels.csv"},
		},
		Projection: &projections.DataFrameProjection{
			BaseProjection: projections.BaseProjection{Name: "my-graph"},
			NodeDataFrames: []projections.NodeDataFrame{
				{Label: "Person", Properties: []string{"name", "age"}},
			},
		},
		Algorithms: []algorithms.Algorithm{
			&algorithms.PageRank{
				BaseAlgorithm: algorithms.BaseAlgorithm{
					GraphName: "my-graph",
					Mode:      algorithms.Stream,
				},
			},
		},
	}

	if session.Name != "analytics-session" {
		t.Errorf("expected name 'analytics-session', got %s", session.Name)
	}
	if session.TTLHours != 24 {
		t.Errorf("expected TTL 24, got %d", session.TTLHours)
	}
}

func TestPandasDataSource(t *testing.T) {
	ds := &PandasDataSource{
		NodeFiles: []string{"nodes.csv", "products.csv"},
		RelFiles:  []string{"purchases.csv"},
	}

	if ds.SourceType() != "pandas" {
		t.Errorf("expected source type 'pandas', got %s", ds.SourceType())
	}
}

func TestSnowflakeDataSource(t *testing.T) {
	ds := &SnowflakeDataSource{
		Account:   "myaccount",
		User:      "myuser",
		Database:  "mydb",
		Schema:    "public",
		Warehouse: "compute_wh",
		NodeQuery: "SELECT * FROM nodes",
		RelQuery:  "SELECT * FROM rels",
	}

	if ds.SourceType() != "snowflake" {
		t.Errorf("expected source type 'snowflake', got %s", ds.SourceType())
	}
	if ds.Account != "myaccount" {
		t.Errorf("expected account 'myaccount', got %s", ds.Account)
	}
}

func TestBigQueryDataSource(t *testing.T) {
	ds := &BigQueryDataSource{
		Project:   "my-project",
		Dataset:   "my_dataset",
		NodeQuery: "SELECT * FROM nodes",
		RelQuery:  "SELECT * FROM rels",
	}

	if ds.SourceType() != "bigquery" {
		t.Errorf("expected source type 'bigquery', got %s", ds.SourceType())
	}
	if ds.Project != "my-project" {
		t.Errorf("expected project 'my-project', got %s", ds.Project)
	}
}

func TestSessionSerializer_ToPython(t *testing.T) {
	session := &Session{
		Name:     "test-session",
		TTLHours: 12,
		DataSource: &PandasDataSource{
			NodeFiles: []string{"nodes.csv"},
			RelFiles:  []string{"rels.csv"},
		},
		Projection: &projections.DataFrameProjection{
			BaseProjection: projections.BaseProjection{Name: "test-graph"},
			NodeDataFrames: []projections.NodeDataFrame{
				{Label: "Person", Properties: []string{"name"}},
			},
		},
		Algorithms: []algorithms.Algorithm{
			&algorithms.PageRank{
				BaseAlgorithm: algorithms.BaseAlgorithm{
					GraphName: "test-graph",
					Mode:      algorithms.Stream,
				},
				DampingFactor: 0.85,
			},
		},
	}

	serializer := NewSerializer()
	python, err := serializer.ToPython(session)
	if err != nil {
		t.Fatalf("ToPython failed: %v", err)
	}

	// Check for expected Python code elements
	if !strings.Contains(python, "from graphdatascience") {
		t.Error("expected Python code to import graphdatascience")
	}
	if !strings.Contains(python, "GdsSessions") {
		t.Error("expected Python code to use GdsSessions")
	}
	if !strings.Contains(python, "page_rank") {
		t.Errorf("expected Python code to include page_rank algorithm, got: %s", python)
	}
}

func TestSessionSerializer_ToJSON(t *testing.T) {
	session := &Session{
		Name:     "test-session",
		TTLHours: 6,
		DataSource: &PandasDataSource{
			NodeFiles: []string{"nodes.csv"},
		},
		Projection: &projections.DataFrameProjection{
			BaseProjection: projections.BaseProjection{Name: "test-graph"},
		},
	}

	serializer := NewSerializer()
	jsonMap := serializer.ToMap(session)

	if jsonMap["name"] != "test-session" {
		t.Errorf("expected name 'test-session', got %v", jsonMap["name"])
	}
	if jsonMap["ttlHours"] != 6 {
		t.Errorf("expected ttlHours 6, got %v", jsonMap["ttlHours"])
	}
	if jsonMap["dataSourceType"] != "pandas" {
		t.Errorf("expected dataSourceType 'pandas', got %v", jsonMap["dataSourceType"])
	}
}

func TestSession_Validate(t *testing.T) {
	tests := []struct {
		name    string
		session *Session
		wantErr bool
	}{
		{
			name: "valid session",
			session: &Session{
				Name:     "valid",
				TTLHours: 1,
				DataSource: &PandasDataSource{
					NodeFiles: []string{"nodes.csv"},
				},
				Projection: &projections.DataFrameProjection{
					BaseProjection: projections.BaseProjection{Name: "graph"},
				},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			session: &Session{
				TTLHours: 1,
				DataSource: &PandasDataSource{
					NodeFiles: []string{"nodes.csv"},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid TTL",
			session: &Session{
				Name:     "test",
				TTLHours: 0,
				DataSource: &PandasDataSource{
					NodeFiles: []string{"nodes.csv"},
				},
			},
			wantErr: true,
		},
		{
			name: "missing data source",
			session: &Session{
				Name:     "test",
				TTLHours: 1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.session.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
