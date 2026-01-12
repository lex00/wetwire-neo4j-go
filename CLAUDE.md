# wetwire-neo4j-go

Neo4j configuration synthesis for Go - type-safe declarations that compile to Cypher and JSON configurations.

## Package structure

```
wetwire-neo4j-go/
├── cmd/wetwire-neo4j/       # CLI entry point
├── pkg/neo4j/               # Public API
│   └── schema/              # Schema definitions (NodeType, RelationshipType)
├── internal/
│   ├── discovery/           # AST-based resource discovery
│   ├── serializer/          # Cypher and JSON serialization
│   ├── linter/              # Neo4j-specific lint rules (WN4xxx)
│   ├── importer/            # Import from existing Neo4j
│   ├── algorithms/          # GDS algorithm configurations
│   ├── pipelines/           # ML pipeline configurations
│   └── graphrag/            # GraphRAG pipeline support
├── docs/                    # Documentation
└── examples/                # Usage examples
```

## Core components

### Schema definitions (`pkg/neo4j/schema`)

Type-safe schema primitives:
- `NodeType` - Node labels with properties, constraints, indexes
- `RelationshipType` - Relationships with source/target, cardinality
- `Property` - Property definitions with Neo4j types
- `Validator` - Schema validation with naming conventions

### Property types

| Type | Neo4j Type | Go Type |
|------|-----------|---------|
| STRING | String | string |
| INTEGER | Integer | int64 |
| FLOAT | Float | float64 |
| BOOLEAN | Boolean | bool |
| DATE | Date | time.Time |
| DATETIME | DateTime | time.Time |
| POINT | Point | Point |
| LIST_STRING | List<String> | []string |

### Naming conventions

- Node labels: PascalCase (e.g., `Person`, `Company`)
- Relationship types: SCREAMING_SNAKE_CASE (e.g., `WORKS_FOR`, `KNOWS`)
- Property names: camelCase or snake_case

## Dependencies

- `github.com/lex00/wetwire-core-go` - Core infrastructure (MCP server, CLI utilities)

## Building and testing

```bash
# Run tests
go test ./...

# Run with verbose output
go test ./... -v

# Check code
go vet ./...
go fmt ./...
```

## Lint rules (planned)

Neo4j lint rules use prefix WN4:
- WN4001-WN4008: GDS algorithm rules
- WN4030-WN4035: ML pipeline rules
- WN4040-WN4047: GraphRAG rules
- WN4050-WN4056: Schema rules

## References

- [Neo4j GDS Documentation](https://neo4j.com/docs/graph-data-science/current/)
- [neo4j-graphrag-python](https://github.com/neo4j/neo4j-graphrag-python)
- [ROADMAP](https://github.com/lex00/wetwire-neo4j-go/issues/18)
