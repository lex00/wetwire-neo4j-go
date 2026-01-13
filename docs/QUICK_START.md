# Quick Start Guide

This guide walks you through creating your first Neo4j configuration with wetwire-neo4j-go.

## Prerequisites

- Go 1.21 or later
- Neo4j 5.x (optional, for validation)

## Installation

```bash
go install github.com/lex00/wetwire-neo4j-go/cmd/wetwire-neo4j@latest
```

## Step 1: Create a Project

```bash
mkdir my-neo4j-project
cd my-neo4j-project
go mod init my-neo4j-project
go get github.com/lex00/wetwire-neo4j-go
```

## Step 2: Define Schema

Create `schema/nodes.go`:

```go
package schema

import "github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"

// Person node type with uniqueness constraint
var Person = &schema.NodeType{
    Label: "Person",
    Properties: []schema.Property{
        {Name: "id", Type: schema.TypeString, Required: true},
        {Name: "name", Type: schema.TypeString, Required: true},
        {Name: "email", Type: schema.TypeString},
    },
    Constraints: []schema.Constraint{
        {Name: "person_id_unique", Type: schema.Unique, Properties: []string{"id"}},
    },
    Indexes: []schema.Index{
        {Name: "person_name_idx", Type: schema.RangeIndex, Properties: []string{"name"}},
    },
}

// Company node type
var Company = &schema.NodeType{
    Label: "Company",
    Properties: []schema.Property{
        {Name: "id", Type: schema.TypeString, Required: true},
        {Name: "name", Type: schema.TypeString, Required: true},
    },
    Constraints: []schema.Constraint{
        {Name: "company_id_unique", Type: schema.Unique, Properties: []string{"id"}},
    },
}
```

Create `schema/relationships.go`:

```go
package schema

import "github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"

// WORKS_FOR relationship between Person and Company
var WorksFor = &schema.RelationshipType{
    Label:  "WORKS_FOR",
    Source: "Person",
    Target: "Company",
    Properties: []schema.Property{
        {Name: "since", Type: schema.TypeDate},
        {Name: "role", Type: schema.TypeString},
    },
}
```

## Step 3: Lint Your Definitions

```bash
wetwire-neo4j lint ./schema/
```

Expected output:
```
0 issues found
```

## Step 4: Generate Cypher

```bash
wetwire-neo4j build ./schema/
```

Expected output:
```cypher
CREATE CONSTRAINT person_id_unique IF NOT EXISTS FOR (n:Person) REQUIRE n.id IS UNIQUE;
CREATE INDEX person_name_idx IF NOT EXISTS FOR (n:Person) ON (n.name);
CREATE CONSTRAINT company_id_unique IF NOT EXISTS FOR (n:Company) REQUIRE n.id IS UNIQUE;
```

## Step 5: List Discovered Resources

```bash
wetwire-neo4j list ./schema/
```

Expected output:
```
NAME      KIND              FILE                    LINE
Person    NodeType          schema/nodes.go         7
Company   NodeType          schema/nodes.go         22
WorksFor  RelationshipType  schema/relationships.go 7
```

## Step 6: Add GDS Algorithm (Optional)

Create `algorithms/pagerank.go`:

```go
package algorithms

import "github.com/lex00/wetwire-neo4j-go/internal/algorithms"

// Influence calculates PageRank scores
var Influence = &algorithms.PageRank{
    BaseAlgorithm: algorithms.BaseAlgorithm{
        GraphName: "company-network",
        Mode:      algorithms.Stream,
    },
    DampingFactor: 0.85,
    MaxIterations: 20,
}
```

Build again to generate the algorithm call:

```bash
wetwire-neo4j build ./...
```

## Step 7: Validate Against Neo4j (Optional)

If you have Neo4j running:

```bash
wetwire-neo4j validate ./schema/ --uri bolt://localhost:7687 --password your-password
```

## Next Steps

- See `examples/` for more comprehensive examples
- Read `docs/CLI.md` for full command reference
- Read `docs/LINT_RULES.md` for lint rule documentation
- Explore GDS algorithms in `internal/algorithms/`
- Explore ML pipelines in `internal/pipelines/`
- Explore GraphRAG retrievers in `internal/retrievers/`

## Common Patterns

### Graph Projection for GDS

```go
var SocialGraph = &projections.NativeProjection{
    Name: "company-network",
    NodeProjections: []projections.NodeProjection{
        {Label: "Person"},
        {Label: "Company"},
    },
    RelationshipProjections: []projections.RelationshipProjection{
        {Type: "WORKS_FOR"},
    },
}
```

### ML Pipeline

```go
var FraudDetection = &pipelines.NodeClassificationPipeline{
    BasePipeline: pipelines.BasePipeline{
        Name:       "fraud-detection",
        GraphName:  "transaction-graph",
        TargetLabel: "Transaction",
        TargetProperty: "isFraud",
    },
    Models: []pipelines.Model{
        &pipelines.RandomForest{NumTrees: 100},
    },
}
```

### GraphRAG Retriever

```go
var SemanticSearch = &retrievers.VectorRetriever{
    Name:       "semantic-search",
    IndexName:  "document-embeddings",
    TopK:       10,
    Embedder: retrievers.EmbedderConfig{
        Provider: "openai",
        Model:    "text-embedding-3-small",
    },
}
```

## Troubleshooting

### "cannot find package"

Run `go mod tidy` to download dependencies.

### "no resources discovered"

Ensure your Go files compile and have package-level variable declarations with the correct types.

### "lint errors"

Check `docs/LINT_RULES.md` for explanations and fixes.
