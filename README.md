# wetwire-neo4j-go

[![CI](https://github.com/lex00/wetwire-neo4j-go/actions/workflows/ci.yml/badge.svg)](https://github.com/lex00/wetwire-neo4j-go/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/lex00/wetwire-neo4j-go.svg)](https://pkg.go.dev/github.com/lex00/wetwire-neo4j-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/lex00/wetwire-neo4j-go)](https://goreportcard.com/report/github.com/lex00/wetwire-neo4j-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Neo4j configuration synthesis for Go - type-safe declarations that compile to Cypher and JSON configurations.

## Overview

wetwire-neo4j-go is a **synthesis library** for Neo4j that goes beyond Cypher queries. Define your Neo4j configurations in native Go and compile them to Cypher statements and JSON configurations.

**Supported configurations:**

- **Schema Definitions** - Node types, relationships, constraints, indexes
- **GDS Algorithms** - PageRank, Louvain, FastRP, Node2Vec, KNN, Dijkstra, and more
- **ML Pipelines** - Node classification, link prediction, node regression
- **GraphRAG** - Retrievers, knowledge graph construction, entity resolution
- **Graph Projections** - Native and Cypher projections for GDS

```
Go Structs → wetwire-neo4j build → Cypher + JSON → Neo4j / GDS / GraphRAG APIs
```

## Installation

```bash
go install github.com/lex00/wetwire-neo4j-go/cmd/neo4j@latest
```

## Quick Start

### 1. Define Schema

```go
package schema

import "github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"

// PersonNode defines a Person node type with constraints.
var PersonNode = &schema.NodeType{
    Label: "Person",
    Properties: []schema.Property{
        {Name: "id", Type: schema.STRING, Required: true},
        {Name: "name", Type: schema.STRING, Required: true},
        {Name: "email", Type: schema.STRING},
    },
    Constraints: []schema.Constraint{
        {Name: "person_id_unique", Type: schema.UNIQUE, Properties: []string{"id"}},
    },
    Indexes: []schema.Index{
        {Name: "person_name_idx", Type: schema.BTREE, Properties: []string{"name"}},
    },
}

// WorksForRelationship defines the WORKS_FOR relationship.
var WorksForRelationship = &schema.RelationshipType{
    Label:  "WORKS_FOR",
    Source: "Person",
    Target: "Company",
    Properties: []schema.Property{
        {Name: "since", Type: schema.DATE},
    },
}
```

### 2. Define Algorithms

```go
package algorithms

import "github.com/lex00/wetwire-neo4j-go/internal/algorithms"

// PageRank configuration
var InfluenceScore = &algorithms.PageRank{
    BaseAlgorithm: algorithms.BaseAlgorithm{
        GraphName: "social-network",
        Mode:      algorithms.Stream,
    },
    DampingFactor: 0.85,
    MaxIterations: 20,
}

// Community detection with Louvain
var Communities = &algorithms.Louvain{
    BaseAlgorithm: algorithms.BaseAlgorithm{
        GraphName: "social-network",
        Mode:      algorithms.Mutate,
    },
    MutateProperty: "communityId",
}
```

### 3. Build and Validate

```bash
# Generate Cypher
neo4j build ./schema/

# Lint configurations
neo4j lint ./schema/

# Validate against live Neo4j
neo4j validate --uri bolt://localhost:7687
```

## Commands

| Command | Description |
|---------|-------------|
| `build` | Generate Cypher statements and JSON configurations |
| `lint` | Validate against wetwire lint rules (WN4xxx) |
| `list` | List all discovered definitions |
| `validate` | Check against a live Neo4j instance |
| `import` | Import existing Neo4j constraints and indexes |

See [docs/CLI.md](docs/CLI.md) for detailed command reference.

## Lint Rules

wetwire-neo4j-go includes lint rules to catch common configuration issues:

| Rule | Description |
|------|-------------|
| WN4001 | Damping factor must be in [0, 1) |
| WN4002 | Max iterations must be positive |
| WN4006 | Embedding dimension should be power of 2 |
| WN4052 | Node labels should use PascalCase |
| WN4053 | Relationship types should use SCREAMING_SNAKE_CASE |

See [docs/LINT_RULES.md](docs/LINT_RULES.md) for complete rule documentation.

## Examples

The `examples/` directory contains comprehensive examples:

- **Schema definitions** - Node types, relationship types with constraints
- **GDS algorithms** - PageRank, Louvain, FastRP, Node2Vec, KNN, Dijkstra
- **ML pipelines** - Node classification, link prediction, regression
- **GraphRAG retrievers** - Vector, Hybrid, Text2Cypher, external integrations
- **KG pipelines** - Entity extraction, knowledge graph construction

## Architecture

```
wetwire-neo4j-go/
├── cmd/neo4j/          # CLI entry point
├── internal/
│   ├── algorithms/     # GDS algorithm definitions
│   ├── cli/            # CLI implementation (build, lint, list)
│   ├── discovery/      # AST-based resource discovery
│   ├── importer/       # Import from Neo4j/Cypher files
│   ├── kg/             # Knowledge graph pipeline definitions
│   ├── lint/           # Lint rules (WN4xxx)
│   ├── pipelines/      # ML pipeline definitions
│   ├── projections/    # Graph projection definitions
│   ├── retrievers/     # GraphRAG retriever definitions
│   ├── serializer/     # Cypher and JSON serializers
│   └── validator/      # Neo4j instance validation
├── pkg/neo4j/schema/   # Public schema types
└── examples/           # Reference examples
```

## Integration with wetwire-core-go

wetwire-neo4j-go integrates with [wetwire-core-go](https://github.com/lex00/wetwire-core-go) for:

- CLI infrastructure and command registration
- MCP server integration for Claude Code
- Common lint rule framework

## References

- [Neo4j GDS Documentation](https://neo4j.com/docs/graph-data-science/current/)
- [neo4j-graphrag-python](https://github.com/neo4j/neo4j-graphrag-python) - Reference implementation
- [wetwire-core-go](https://github.com/lex00/wetwire-core-go) - Core infrastructure

## License

MIT License - see [LICENSE](LICENSE) for details.
