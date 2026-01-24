---
title: "Quick Start"
---


This guide walks you through creating your first Neo4j configuration with wetwire-neo4j-go.

## Prerequisites

- Go 1.21 or later
- Neo4j 5.x (optional, for validation)

## Installation

See [README.md](../README.md#installation) for installation instructions.

## Step 1: Create a Project

```bash
# Initialize a new project with wetwire-neo4j
wetwire-neo4j init my-neo4j-project
cd my-neo4j-project

# Initialize Go module and fetch dependencies
go mod init my-neo4j-project
go mod tidy
```

This creates a project scaffold with starter schema, and directories for algorithms, pipelines, retrievers, and knowledge graphs.

**Templates available:**
- `--template default` - Basic schema only (default)
- `--template gds` - Schema + algorithms + pipelines
- `--template graphrag` - Schema + retrievers + KG pipelines
- `--template full` - Everything

```bash
# Example: Create a GDS-focused project
wetwire-neo4j init my-gds-project --template gds
```

## Step 2: Define Schema (or Import from Neo4j)

You can define your schema from scratch, or import from an existing Neo4j database.

### Option A: Import from Existing Database

If you already have a Neo4j database, import the schema directly:

```bash
# Import from a live Neo4j instance
wetwire-neo4j import \
  --uri bolt://localhost:7687 \
  --username neo4j \
  --password your-password \
  --database neo4j \
  --package schema \
  --output schema/generated.go
```

This discovers all node labels, relationship types, properties, constraints, and indexes from your database.

You can also import from a Cypher file:

```bash
# Import from a Cypher DDL file
wetwire-neo4j import --file ./existing-schema.cypher --output schema/generated.go
```

After importing, add `AgentContext` and `AgentHint` fields (see Step 3) to guide AI agents.

### Option B: Define from Scratch

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

## Step 3: Add Agent Context (Recommended)

The key advantage of wetwire schemas is guiding AI agents. Add context to help agents generate better queries.

### Schema-Level Context

Create `schema/schema.go` to define global instructions:

```go
package schema

import "github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"

// MySchema defines the complete schema with agent instructions.
var MySchema = &schema.Schema{
    Name:        "hr-database",
    Description: "Human resources and organizational data",
    Nodes:       []*schema.NodeType{Person, Company},
    Relationships: []*schema.RelationshipType{WorksFor},

    // Global instructions for AI agents
    AgentContext: `
        Multi-tenant database - always filter by tenantId property.
        Ignore nodes with label prefix "_" (internal system nodes).
        Prefer WORKS_FOR over legacy EMPLOYED_BY relationship.
    `,
}
```

### Resource-Level Hints

Add hints to individual nodes and relationships:

```go
var Person = &schema.NodeType{
    Label: "Person",
    Properties: []schema.Property{
        {Name: "id", Type: schema.STRING, Required: true},
        {Name: "email", Type: schema.STRING},
    },
    // Agent-specific guidance for this node
    AgentHint: "Query by email for uniqueness. Name field may have duplicates.",
}

var WorksFor = &schema.RelationshipType{
    Label:  "WORKS_FOR",
    Source: "Person",
    Target: "Company",
    // Agent-specific guidance for this relationship
    AgentHint: "Canonical employment relationship. Prefer over EMPLOYED_BY (deprecated).",
}
```

When agents read your schema, they'll use this context to:
- Apply correct filters automatically
- Choose the right relationships
- Avoid internal/system nodes
- Follow your data modeling conventions

## Step 4: Lint Your Definitions

```bash
wetwire-neo4j lint ./schema/
```

Expected output:
```
0 issues found
```

## Step 5: Generate Cypher

```bash
wetwire-neo4j build ./schema/
```

Expected output:
```cypher
CREATE CONSTRAINT person_id_unique IF NOT EXISTS FOR (n:Person) REQUIRE n.id IS UNIQUE;
CREATE INDEX person_name_idx IF NOT EXISTS FOR (n:Person) ON (n.name);
CREATE CONSTRAINT company_id_unique IF NOT EXISTS FOR (n:Company) REQUIRE n.id IS UNIQUE;
```

## Step 6: List Discovered Resources

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

## Step 7: Add GDS Algorithm (Optional)

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

## Step 8: Validate Against Neo4j (Optional)

If you have Neo4j running:

```bash
wetwire-neo4j validate ./schema/ --uri bolt://localhost:7687 --password your-password
```

## AI-Assisted Design

Let AI help create your Neo4j schema:

```bash
# No API key required - uses Claude CLI
wetwire-neo4j design "Create a schema for a social network with users, posts, and comments"
```

The design command automatically discovers your existing schema and extends it with new definitions.

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
