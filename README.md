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

### 4. Extend Schema with AI (Optional)

Use the `design` command for AI-assisted schema modifications:

```bash
wetwire-neo4j design "Add a Project node and ASSIGNED_TO relationship"
```

The agent automatically discovers your existing schema and extends it correctly - no need to re-explain what exists.

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

## Import from Existing Database

Already have a Neo4j database? Import your existing constraints and indexes to bootstrap your schema:

### From a Live Database

```bash
wetwire-neo4j import --uri bolt://localhost:7687 --password secret -o schema/generated.go
```

### From a Cypher File

```bash
wetwire-neo4j import --file ./constraints.cypher -o schema/generated.go
```

### Example

Given these existing constraints:

```cypher
CREATE CONSTRAINT person_id FOR (p:Person) REQUIRE p.id IS UNIQUE;
CREATE INDEX person_name FOR (p:Person) ON (p.name);
CREATE CONSTRAINT company_name FOR (c:Company) REQUIRE c.name IS UNIQUE;
```

The import command generates:

```go
package schema

import "github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"

var Person = &schema.NodeType{
    Label: "Person",
    Properties: []schema.Property{
        {Name: "id", Type: schema.STRING, Required: true, Unique: true},
        {Name: "name", Type: schema.STRING},
    },
}

var Company = &schema.NodeType{
    Label: "Company",
    Properties: []schema.Property{
        {Name: "name", Type: schema.STRING, Required: true, Unique: true},
    },
}
```

After importing, you can add relationship definitions, cardinality, and GoDoc comments to complete your schema.

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

## AI-Assisted Schema Design

The `design` command provides AI-assisted schema generation with automatic schema awareness:

```bash
# Start an AI design session (Anthropic API)
wetwire-neo4j design "Create a social network schema"

# Or use Kiro CLI
wetwire-neo4j design --provider kiro "Add audit logging to my schema"
```

### Pre-flight Schema Discovery

When you run `design`, the CLI automatically scans your project for existing schema definitions and injects them into the agent's context. This means:

- **Agents know what exists** - The AI sees your current nodes, relationships, and algorithms
- **Extend, don't recreate** - Agents add to your schema rather than starting from scratch
- **Works with imports** - Schemas imported from Neo4j are automatically discovered

```
┌─────────────────────────────────────────────────────────────┐
│  1. You have existing schema (manual or imported)           │
│     schema/person.go, schema/company.go                     │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  2. Run design command                                      │
│     wetwire-neo4j design "Add audit logging"                │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  3. Agent sees existing schema in its prompt                │
│     "Existing: Person, Company, WORKS_FOR..."               │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  4. Agent extends your schema correctly                     │
│     Adds AuditLog node, AUDITED_BY relationship             │
└─────────────────────────────────────────────────────────────┘
```

The schema context is summarized (name + file location) to keep prompts lean. Agents can use `wetwire_list` or read source files for full property details.

### MCP Tools

The MCP server (`wetwire-neo4j mcp`) exposes tools for AI agents:
- `wetwire_init` - Initialize a new project
- `wetwire_lint` - Validate schema definitions
- `wetwire_build` - Generate Cypher and JSON configs
- `wetwire_list` - List discovered resources with details

See [docs/NEO4J-KIRO-CLI.md](docs/NEO4J-KIRO-CLI.md) for complete Kiro integration guide.

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
