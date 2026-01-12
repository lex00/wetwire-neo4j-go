# Claude Code Guidelines for wetwire-neo4j-go

This document provides AI assistant guidelines for working with wetwire-neo4j-go.

## Overview

wetwire-neo4j-go is a synthesis library for Neo4j that generates Cypher statements and JSON configurations from native Go definitions. It follows the wetwire pattern: declare infrastructure in native code, then synthesize deployment artifacts.

## Package Structure

```
wetwire-neo4j-go/
├── cmd/neo4j/          # CLI entry point
├── internal/
│   ├── algorithms/     # GDS algorithm definitions
│   ├── cli/            # CLI implementation
│   ├── discovery/      # AST-based resource discovery
│   ├── importer/       # Import from Neo4j/Cypher
│   ├── kg/             # Knowledge graph pipelines
│   ├── lint/           # Lint rules (WN4xxx)
│   ├── pipelines/      # ML pipeline definitions
│   ├── projections/    # Graph projection definitions
│   ├── retrievers/     # GraphRAG retrievers
│   ├── serializer/     # Cypher and JSON serializers
│   └── validator/      # Neo4j instance validation
├── pkg/neo4j/schema/   # Public schema types
└── examples/           # Reference examples
```

## Core Patterns

### Schema Definitions

Node types and relationship types are defined as Go struct pointers:

```go
// Node type with label, properties, constraints, and indexes
var Person = &schema.NodeType{
    Label: "Person",
    Properties: []schema.Property{
        {Name: "id", Type: schema.TypeString, Required: true},
        {Name: "name", Type: schema.TypeString, Required: true},
    },
    Constraints: []schema.Constraint{
        {Type: schema.Unique, Properties: []string{"id"}},
    },
}

// Relationship type with source/target labels
var WorksFor = &schema.RelationshipType{
    Label:  "WORKS_FOR",
    Source: "Person",
    Target: "Company",
}
```

### GDS Algorithms

Algorithms are defined with a `BaseAlgorithm` embedded struct:

```go
var Influence = &algorithms.PageRank{
    BaseAlgorithm: algorithms.BaseAlgorithm{
        GraphName: "social-network",
        Mode:      algorithms.Stream,
    },
    DampingFactor: 0.85,
    MaxIterations: 20,
}
```

### Graph Projections

Projections create in-memory graphs for GDS:

```go
var SocialGraph = &projections.NativeProjection{
    Name: "social-network",
    NodeProjections: []projections.NodeProjection{
        {Label: "Person"},
    },
    RelationshipProjections: []projections.RelationshipProjection{
        {Type: "KNOWS"},
    },
}
```

## Naming Conventions

| Element | Convention | Example |
|---------|------------|---------|
| Node Labels | PascalCase | `Person`, `Company`, `MovieRating` |
| Relationship Types | SCREAMING_SNAKE_CASE | `WORKS_FOR`, `HAS_ADDRESS` |
| Property Names | camelCase or snake_case | `firstName`, `created_at` |
| Variable Names | camelCase | `person`, `worksFor` |

## Lint Rules (WN4xxx)

| Range | Category |
|-------|----------|
| WN4001-WN4029 | GDS Algorithm Rules |
| WN4030-WN4039 | ML Pipeline Rules |
| WN4040-WN4049 | GraphRAG Rules |
| WN4050-WN4059 | Schema Rules |
| WN4060-WN4069 | Projection Rules |

Key rules:
- **WN4001**: dampingFactor must be in [0, 1)
- **WN4002**: maxIterations must be positive
- **WN4006**: embeddingDimension should be power of 2
- **WN4052**: node labels should be PascalCase
- **WN4053**: relationship types should be SCREAMING_SNAKE_CASE

## CLI Commands

```bash
wetwire-neo4j build ./schemas/     # Generate Cypher/JSON
wetwire-neo4j lint ./schemas/      # Check for issues
wetwire-neo4j list ./schemas/      # List discovered resources
wetwire-neo4j validate --uri ...   # Validate against Neo4j
wetwire-neo4j import --file ...    # Import existing schemas
```

## Property Types

| Constant | Neo4j Type | Go Type |
|----------|------------|---------|
| `TypeString` | String | string |
| `TypeInteger` | Integer | int64 |
| `TypeFloat` | Float | float64 |
| `TypeBoolean` | Boolean | bool |
| `TypeDate` | Date | time.Time |
| `TypeDateTime` | DateTime | time.Time |
| `TypePoint` | Point | Point |
| `TypeListString` | List<String> | []string |

## Common Tasks

### Adding a New Algorithm

1. Check existing algorithms in `internal/algorithms/`
2. Create a new struct embedding `BaseAlgorithm`
3. Add serialization in `internal/serializer/algorithm.go`
4. Add lint rules in `internal/lint/lint.go`
5. Add tests and examples

### Adding a New Lint Rule

1. Add rule constant in `internal/lint/lint.go`
2. Implement rule logic in appropriate Lint* method
3. Add test cases
4. Document in `docs/LINT_RULES.md`

## Gotchas

1. **Algorithm Mode**: Always set `Mode` in `BaseAlgorithm` - defaults to empty which is invalid
2. **Graph Name**: Algorithms reference a graph by name - ensure a projection with that name exists
3. **Property Types**: Use `schema.Type*` constants, not strings
4. **Constraint Names**: Neo4j requires unique constraint names across the database

## Building and Testing

```bash
go test ./...           # Run all tests
go test ./... -v        # Verbose output
golangci-lint run ./... # Run linter
go build ./...          # Build all packages
```

## Kiro CLI and MCP Integration

wetwire-neo4j integrates with Kiro CLI for AI-assisted schema design:

- **MCP Server**: Available as `wetwire-neo4j mcp` subcommand (NOT a standalone binary)
- **Design Command**: `wetwire-neo4j design --provider kiro` auto-configures Kiro agent
- **MCP Tools**: `wetwire_init`, `wetwire_lint`, `wetwire_build`

See [docs/NEO4J-KIRO-CLI.md](docs/NEO4J-KIRO-CLI.md) for full integration guide.

## Dependencies

- `github.com/lex00/wetwire-core-go` - Core infrastructure (MCP server, CLI utilities)
- `github.com/neo4j/neo4j-go-driver/v5` - Neo4j driver for validation/import

## References

- [Neo4j GDS Documentation](https://neo4j.com/docs/graph-data-science/current/)
- [neo4j-graphrag-python](https://github.com/neo4j/neo4j-graphrag-python)
- [wetwire-core-go](https://github.com/lex00/wetwire-core-go)
- [Kiro CLI Integration](docs/NEO4J-KIRO-CLI.md)
- [ROADMAP](https://github.com/lex00/wetwire-neo4j-go/issues/18)
