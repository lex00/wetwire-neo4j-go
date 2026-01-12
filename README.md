# wetwire-neo4j-go

[![CI](https://github.com/lex00/wetwire-neo4j-go/actions/workflows/ci.yml/badge.svg)](https://github.com/lex00/wetwire-neo4j-go/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/lex00/wetwire-neo4j-go.svg)](https://pkg.go.dev/github.com/lex00/wetwire-neo4j-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/lex00/wetwire-neo4j-go)](https://goreportcard.com/report/github.com/lex00/wetwire-neo4j-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Type-safe Neo4j schema definitions for AI-assisted Cypher generation.

## Overview

wetwire-neo4j-go lets you define your Neo4j schema in type-safe Go, then use AI agents (Claude, Kiro, etc.) to generate correct Cypher queries. The Go schema acts as a validated contract that agents can read to understand your database structure.

```
Define Schema (Go) → Validate (lint) → Agent Reads Schema → Generates Correct Cypher
```

**Why this approach?**

- **Type-safe schema** - Compiler and linter catch errors before agents see them
- **Rich context** - GoDoc comments, cardinality, constraints inform query generation
- **Consistent results** - Agents generate correct label/relationship names every time
- **Single source of truth** - Schema is version-controlled, reviewable Go code

**Also supports:**

- **GDS Algorithms** - PageRank, Louvain, FastRP, Node2Vec, KNN, Dijkstra
- **ML Pipelines** - Node classification, link prediction, node regression
- **GraphRAG** - Retrievers, knowledge graph construction, entity resolution
- **Graph Projections** - Native and Cypher projections for GDS

## Installation

```bash
go install github.com/lex00/wetwire-neo4j-go/cmd/neo4j@latest
```

## Quick Start

wetwire-neo4j-go lets you define your Neo4j schema in type-safe Go, then use AI agents to generate Cypher queries that are consistent and correct.

### The Workflow

```
┌─────────────────────────────────────────────────────────────┐
│  1. Define your schema in Go (once)                         │
│     schema/nodes.go, schema/relationships.go                │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  2. Validate with lint                                      │
│     wetwire-neo4j lint ./schema/                            │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  3. Ask the agent for queries                               │
│     "Read schema/ and find people at tech companies"        │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  4. Get correct Cypher                                      │
│     MATCH (p:Person)-[:WORKS_FOR]->(c:Company)              │
│     WHERE c.industry = 'tech' RETURN p.name, c.name         │
└─────────────────────────────────────────────────────────────┘
```

### 1. Define Your Schema

Create a `schema/` package with your node and relationship definitions:

**schema/nodes.go**
```go
package schema

import "github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"

// Person represents an individual in the system.
var Person = &schema.NodeType{
    Label: "Person",
    Properties: []schema.Property{
        {Name: "id", Type: schema.STRING, Required: true, Unique: true},
        {Name: "name", Type: schema.STRING, Required: true},
        {Name: "email", Type: schema.STRING, Unique: true},
        {Name: "age", Type: schema.INTEGER},
    },
}

// Company represents an organization.
var Company = &schema.NodeType{
    Label: "Company",
    Properties: []schema.Property{
        {Name: "name", Type: schema.STRING, Required: true, Unique: true},
        {Name: "industry", Type: schema.STRING},
        {Name: "employeeCount", Type: schema.INTEGER},
    },
}
```

**schema/relationships.go**
```go
package schema

import "github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"

// WorksFor connects a Person to their employer.
// Each person can work for one company; companies have many employees.
var WorksFor = &schema.RelationshipType{
    Label:       "WORKS_FOR",
    Source:      "Person",
    Target:      "Company",
    Cardinality: schema.MANY_TO_ONE,
    Properties: []schema.Property{
        {Name: "since", Type: schema.DATE, Required: true},
        {Name: "role", Type: schema.STRING},
    },
}

// Knows represents a social connection between people.
var Knows = &schema.RelationshipType{
    Label:       "KNOWS",
    Source:      "Person",
    Target:      "Person",
    Cardinality: schema.MANY_TO_MANY,
    Properties: []schema.Property{
        {Name: "since", Type: schema.DATE},
        {Name: "strength", Type: schema.FLOAT},
    },
}
```

### 2. Validate Your Schema

```bash
wetwire-neo4j lint ./schema/
```

The linter enforces naming conventions (PascalCase labels, SCREAMING_SNAKE_CASE relationships), required fields, and valid references.

### 3. Generate Queries with AI

Ask your AI agent to read the schema and generate queries:

> "Read schema/ and write a query to find people who work at tech companies"

The agent generates correct Cypher because it knows your exact schema:

```cypher
MATCH (p:Person)-[w:WORKS_FOR]->(c:Company)
WHERE c.industry = 'tech'
RETURN p.name, c.name, w.role
```

### Why Go Schema Instead of JSON?

| Aspect | JSON Schema | Go Schema |
|--------|-------------|-----------|
| Validation | Runtime only | Compile-time + lint |
| Documentation | None | First-class GoDoc comments |
| Structure | Deeply nested | Flat, scannable declarations |
| Naming conventions | Honor system | Enforced by linter |
| Cardinality | Usually missing | Explicit field |
| Typos | Found at query time | Caught by compiler |

The Go schema acts as a **contract** - if it passes `wetwire-neo4j lint`, the agent can trust every label, relationship type, and property name is exactly as defined. This produces more consistent and correct Cypher than having agents work from JSON or documentation.

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

## Kiro CLI Integration

wetwire-neo4j works with [Kiro CLI](https://kiro.dev) for AI-assisted schema design:

```bash
# Auto-configure and start Kiro design session
wetwire-neo4j design --provider kiro "Create a social network schema"
```

The MCP server (`wetwire-neo4j mcp`) exposes three tools to Kiro:
- `wetwire_init` - Initialize a new project
- `wetwire_lint` - Validate schema definitions
- `wetwire_build` - Generate Cypher and JSON configs

See [docs/NEO4J-KIRO-CLI.md](docs/NEO4J-KIRO-CLI.md) for complete integration guide.

## Integration with wetwire-core-go

wetwire-neo4j-go integrates with [wetwire-core-go](https://github.com/lex00/wetwire-core-go) for:

- CLI infrastructure and command registration
- MCP server integration for Claude Code and Kiro CLI
- Common lint rule framework

## References

- [Neo4j GDS Documentation](https://neo4j.com/docs/graph-data-science/current/)
- [neo4j-graphrag-python](https://github.com/neo4j/neo4j-graphrag-python) - Reference implementation
- [wetwire-core-go](https://github.com/lex00/wetwire-core-go) - Core infrastructure

## License

MIT License - see [LICENSE](LICENSE) for details.
