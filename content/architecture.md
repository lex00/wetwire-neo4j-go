---
title: "Architecture"
---

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./wetwire-dark.svg">
  <img src="./wetwire-light.svg" width="100" height="67">
</picture>

This document describes the architecture of wetwire-neo4j-go, a synthesis library for Neo4j that generates Cypher statements and JSON configurations from native Go definitions.

## Overview

### Purpose and Goals

wetwire-neo4j-go follows the "wetwire" pattern: **declare infrastructure in native Go code, then synthesize deployment artifacts**. Instead of writing raw Cypher DDL or JSON configuration files, developers define Neo4j schemas, GDS algorithms, ML pipelines, and GraphRAG retrievers as type-safe Go structs. The toolchain then:

1. **Discovers** definitions via AST scanning
2. **Validates** configurations against lint rules and optionally a live Neo4j instance
3. **Serializes** to Cypher statements or JSON configurations
4. **Outputs** artifacts for deployment or further processing

### Key Benefits

- **Type Safety**: Compile-time validation of schema definitions
- **IDE Support**: Autocomplete, documentation, and refactoring
- **Testability**: Unit test configurations before deployment
- **Reproducibility**: Version-controlled infrastructure as code
- **Integration**: Works with existing Neo4j tooling and workflows

### Relationship to wetwire-core-go

wetwire-neo4j-go depends on `github.com/lex00/wetwire-core-go` for shared infrastructure:

- **MCP Server**: Model Context Protocol server implementation for tool registration
- **CLI Framework**: Command-line interface utilities (build, lint, init commands)
- **Kiro Integration**: Kiro spec generation utilities for AI-assisted development

The core library provides the generic framework; this package implements Neo4j-specific types and serializers.

## Package Structure

```
wetwire-neo4j-go/
├── cmd/wetwire-neo4j/      # CLI entry point
├── internal/               # Internal packages (not importable)
│   ├── algorithms/         # GDS algorithm definitions
│   ├── aura/               # Neo4j Aura cloud integration
│   ├── cli/                # CLI command implementations
│   ├── discovery/          # AST-based resource discovery
│   ├── importer/           # Import from Neo4j/Cypher files
│   ├── kg/                 # Knowledge graph construction pipelines
│   ├── kiro/               # Kiro agent integration
│   ├── lint/               # Lint rules (WN4xxx)
│   ├── pipelines/          # ML pipeline definitions
│   ├── projections/        # Graph projection definitions
│   ├── retrievers/         # GraphRAG retriever definitions
│   ├── serializer/         # Cypher and JSON serializers
│   └── validator/          # Neo4j instance validation
├── pkg/neo4j/schema/       # Public schema types (importable)
└── examples/               # Reference examples
```

### pkg/neo4j/schema/

**Public API** - Types that users import to define their schemas.

| File | Purpose |
|------|---------|
| `types.go` | Core type definitions: `NodeType`, `RelationshipType`, `Property`, `Constraint`, `Index` |
| `validation.go` | Schema-level validation functions |

Key types:
- `NodeType`: Defines a node label with properties, constraints, and indexes
- `RelationshipType`: Defines a relationship type with source/target labels and properties
- `Property`: Defines a property with type, required flag, and uniqueness
- `PropertyType`: Enum of Neo4j data types (STRING, INTEGER, FLOAT, etc.)
- `Constraint`: Defines UNIQUE, EXISTS, NODE_KEY, or REL_KEY constraints
- `Index`: Defines BTREE, TEXT, FULLTEXT, POINT, or VECTOR indexes

### internal/discover/

AST-based resource discovery scans Go source files to find schema definitions.

| Component | Purpose |
|-----------|---------|
| `Scanner` | Parses Go files to find struct literals matching known types |
| `DiscoveredResource` | Metadata about found resources (name, kind, file, line, dependencies) |
| `DependencyGraph` | Builds and sorts resources by dependency order |

The scanner recognizes:
- Schema types: `NodeType`, `RelationshipType`
- Algorithm types: `PageRank`, `Louvain`, `FastRP`, etc.
- Pipeline types: `NodeClassificationPipeline`, `LinkPredictionPipeline`
- Retriever types: `VectorRetriever`, `HybridRetriever`, `Text2CypherRetriever`

### internal/serializer/

Converts schema definitions to output formats.

| File | Purpose |
|------|---------|
| `cypher.go` | Generates Cypher DDL (CREATE CONSTRAINT, CREATE INDEX) |
| `json.go` | Generates JSON configuration objects |

The Cypher serializer uses Go templates to generate:
- UNIQUE, EXISTS, NODE_KEY constraints
- BTREE, TEXT, FULLTEXT, POINT, VECTOR indexes
- Relationship constraints

### internal/algorithms/

Type-safe configurations for Neo4j Graph Data Science algorithms.

| Category | Algorithms |
|----------|------------|
| Centrality | PageRank, ArticleRank, Betweenness, Degree, Closeness |
| Community | Louvain, Leiden, LabelPropagation, WCC, KCore, TriangleCount |
| Similarity | NodeSimilarity, KNN |
| Path Finding | Dijkstra, AStar, BFS, DFS |
| Embeddings | FastRP, Node2Vec, GraphSAGE, HashGNN |

Each algorithm struct embeds `BaseAlgorithm` for common fields (GraphName, Mode, Concurrency).

### internal/pipelines/

ML pipeline configurations for GDS machine learning.

| Pipeline Type | Purpose |
|---------------|---------|
| `NodeClassificationPipeline` | Predict categorical node labels |
| `LinkPredictionPipeline` | Predict future relationships |
| `NodeRegressionPipeline` | Predict numeric node properties |

Supports feature steps (FastRP, PageRank, Degree, Node2Vec) and model types (LogisticRegression, RandomForest, MLP, LinearRegression).

### internal/projections/

Graph projection configurations for creating in-memory GDS graphs.

| Projection Type | Purpose |
|-----------------|---------|
| `NativeProjection` | Project by node labels and relationship types |
| `CypherProjection` | Project using custom Cypher queries |
| `DataFrameProjection` | Project from DataFrames (Aura Analytics) |

### internal/retrievers/

GraphRAG retriever configurations compatible with neo4j-graphrag-python.

| Retriever Type | Purpose |
|----------------|---------|
| `VectorRetriever` | Similarity search using vector indexes |
| `VectorCypherRetriever` | Vector search with custom graph traversal |
| `HybridRetriever` | Combined vector and fulltext search |
| `HybridCypherRetriever` | Hybrid search with custom traversal |
| `Text2CypherRetriever` | LLM-generated Cypher queries |
| `WeaviateRetriever` | External Weaviate integration |
| `PineconeRetriever` | External Pinecone integration |
| `QdrantRetriever` | External Qdrant integration |

### internal/kg/

Knowledge graph construction pipeline configurations.

| Pipeline Type | Purpose |
|---------------|---------|
| `SimpleKGPipeline` | Standard entity and relationship extraction |
| `CustomKGPipeline` | Custom extraction with user-defined prompts |

Includes text splitters (FixedSize, LangChain) and entity resolvers (ExactMatch, FuzzyMatch, SemanticMatch).

### internal/lint/

Lint rules for validating configurations (WN4xxx rule codes).

| Rule Range | Category |
|------------|----------|
| WN4001-WN4029 | GDS Algorithm Rules |
| WN4030-WN4039 | ML Pipeline Rules |
| WN4040-WN4049 | GraphRAG/KG Rules |
| WN4050-WN4059 | Schema Rules |

Key rules:
- **WN4001**: dampingFactor must be in [0, 1)
- **WN4006**: embeddingDimension should be power of 2
- **WN4052**: Node labels should be PascalCase
- **WN4053**: Relationship types should be SCREAMING_SNAKE_CASE

### internal/validator/

Validates configurations against a live Neo4j instance.

| Validation | Description |
|------------|-------------|
| Schema | Check if labels and relationship types exist |
| Algorithm | Check if GDS algorithms are available |
| Projection | Validate graph projection references |
| GDS Info | Query GDS version and edition |

### internal/importer/

Import existing Neo4j schemas and generate Go code.

| Source | Description |
|--------|-------------|
| Cypher files | Parse CREATE CONSTRAINT/INDEX statements |
| Live Neo4j | Query database metadata |

### internal/cli/

CLI command implementations.

| Command | Purpose |
|---------|---------|
| `builder.go` | Build command - generate Cypher/JSON |
| `linter.go` | Lint command - check for issues |
| `lister.go` | List command - show discovered resources |
| `validator.go` | Validate command - check against Neo4j |
| `importer.go` | Import command - generate Go from existing schemas |
| `initializer.go` | Init command - scaffold new projects |
| `graph.go` | Graph command - visualize dependencies |

### cmd/wetwire-neo4j/

Main CLI entry point and command registration.

| File | Purpose |
|------|---------|
| `main.go` | Command registration and execution |
| `domain.go` | Domain-specific command setup |
| `design.go` | AI-assisted schema design command |
| `test.go` | Persona-based testing command |
| `mcp.go` | MCP server for tool integration |

## Data Flow

### Discovery to Serialization to Output

```
┌─────────────────┐      ┌─────────────────┐      ┌─────────────────┐
│   Go Source     │      │    Discovery    │      │   Serializer    │
│   Definitions   │─────>│    Scanner      │─────>│   (Cypher/JSON) │
│                 │      │                 │      │                 │
│ schema.NodeType │      │ DiscoveredRes.  │      │ CREATE CONST... │
│ algorithms.PR   │      │ DependencyGraph │      │ {"nodeTypes":.. │
└─────────────────┘      └─────────────────┘      └─────────────────┘
                                                          │
                                                          v
                                                  ┌─────────────────┐
                                                  │    Output       │
                                                  │   (.cypher/.json│
                                                  │    or stdout)   │
                                                  └─────────────────┘
```

1. **Discovery Phase**: The `Scanner` parses Go source files using `go/parser` to find struct literals that match known resource types. It extracts metadata (name, file, line) and builds a dependency graph.

2. **Validation Phase** (optional): The `Linter` checks configurations against WN4xxx rules. The `Validator` can optionally connect to a live Neo4j instance to verify labels, types, and GDS availability.

3. **Serialization Phase**: The appropriate serializer converts definitions to the target format:
   - `CypherSerializer`: Generates DDL statements using Go templates
   - `JSONSerializer`: Generates configuration objects for GraphRAG/pipelines
   - `AlgorithmSerializer`, `PipelineSerializer`, etc.: Domain-specific serializers

4. **Output Phase**: Results are written to a file or stdout.

### Validation Against Neo4j

```
┌─────────────────┐      ┌─────────────────┐      ┌─────────────────┐
│   Definitions   │─────>│    Validator    │─────>│    Results      │
│                 │      │                 │      │                 │
│ NodeType        │      │ Connect to DB   │      │ ✓ Label exists  │
│ Algorithm       │      │ Query metadata  │      │ ✗ GDS not found │
│ Projection      │      │ Check refs      │      │ ✓ Rel type OK   │
└─────────────────┘      └─────────────────┘      └─────────────────┘
```

The validator connects to Neo4j and queries:
- `db.labels()` - Available node labels
- `db.relationshipTypes()` - Available relationship types
- `gds.version()` - GDS version (if installed)
- `gds.graph.exists()` - Graph projection existence

### Import from Existing Databases

```
┌─────────────────┐      ┌─────────────────┐      ┌─────────────────┐
│   Source        │─────>│    Importer     │─────>│   Go Code       │
│                 │      │                 │      │                 │
│ Cypher DDL file │      │ Parse/Query     │      │ var Person = &  │
│ Live Neo4j DB   │      │ Extract schema  │      │   schema.Node.. │
└─────────────────┘      └─────────────────┘      └─────────────────┘
```

The importer can:
1. Parse Cypher DDL files to extract constraint and index definitions
2. Query a live Neo4j database for schema metadata
3. Generate Go code using the `Generator` that outputs `schema.NodeType` and `schema.RelationshipType` definitions

## Extension Points

### Adding New Schema Types

1. Define the new type in `pkg/neo4j/schema/types.go`:
   ```go
   type NewSchemaType struct {
       Name string
       // ... fields
   }

   func (n *NewSchemaType) ResourceType() string { return "NewSchemaType" }
   func (n *NewSchemaType) ResourceName() string { return n.Name }
   ```

2. Add type alias to `internal/discover/discovery.go`:
   ```go
   s.typeAliases["NewSchemaType"] = KindNewSchema
   ```

3. Add serialization in `internal/serializer/`:
   - Cypher template for DDL generation
   - JSON mapping for configuration output

4. Add lint rules in `internal/lint/lint.go`:
   ```go
   func (l *Linter) LintNewSchemaType(t *schema.NewSchemaType) []LintResult {
       // WN40XX rules
   }
   ```

5. Add tests and update documentation.

### Adding New Algorithms

1. Define the algorithm struct in `internal/algorithms/algorithms.go`:
   ```go
   type NewAlgorithm struct {
       BaseAlgorithm
       // Algorithm-specific parameters
       Param1 float64
       Param2 int
   }

   func (a *NewAlgorithm) AlgorithmType() string       { return "gds.newAlgorithm" }
   func (a *NewAlgorithm) AlgorithmCategory() Category { return Centrality }
   ```

2. Add serialization in `internal/algorithms/serialize.go`:
   - `ToCypher()` method for GDS procedure calls
   - `ToMap()` method for JSON output

3. Add lint rules for parameter validation:
   ```go
   func (l *Linter) lintNewAlgorithm(a *algorithms.NewAlgorithm) []LintResult {
       // Parameter range checks
   }
   ```

4. Register in discovery scanner if using custom type detection.

### Adding New Lint Rules

1. Choose a rule code from the appropriate range:
   - WN4001-WN4029: GDS Algorithms
   - WN4030-WN4039: ML Pipelines
   - WN4040-WN4049: GraphRAG/KG
   - WN4050-WN4059: Schema

2. Add the rule implementation in `internal/lint/lint.go`:
   ```go
   // WN40XX: Description of what the rule checks
   if condition {
       results = append(results, LintResult{
           Rule:     "WN40XX",
           Severity: Warning, // or Error, Info
           Message:  "Descriptive message",
           Location: "Type.Field",
       })
   }
   ```

3. Add test cases in `internal/lint/lint_test.go`.

4. Document the rule in `docs/LINT_RULES.md`.

### Adding New Retrievers

1. Define the retriever struct in `internal/retrievers/retrievers.go`:
   ```go
   type NewRetriever struct {
       BaseRetriever
       // Retriever-specific configuration
   }

   func (r *NewRetriever) RetrieverType() RetrieverType { return NewRetrieverKind }
   ```

2. Add serialization in `internal/retrievers/serialize.go`.

3. Add lint rules if needed.

4. Register in discovery scanner.

## Dependencies

### wetwire-core-go Interfaces

The package depends on interfaces from wetwire-core-go:

| Interface | Usage |
|-----------|-------|
| `cmd.Builder` | Implemented by `cli.Builder` for the build command |
| `cmd.Linter` | Implemented by `cli.Linter` for the lint command |
| `cmd.Initializer` | Implemented by `cli.Initializer` for the init command |
| `mcp.Server` | Used for MCP tool registration |

### Neo4j Go Driver

```go
require github.com/neo4j/neo4j-go-driver/v5 v5.28.4
```

Used by:
- `internal/validator/` - Connect to Neo4j for validation
- `internal/importer/` - Query database schema

### Cobra CLI

```go
require github.com/spf13/cobra v1.10.2
```

Used for command-line interface implementation in `cmd/wetwire-neo4j/`.

### Standard Library

The codebase makes extensive use of:
- `go/ast`, `go/parser`, `go/token` - AST-based discovery
- `text/template` - Cypher template generation
- `encoding/json` - JSON serialization
- `regexp` - Naming convention validation

## Configuration

### Environment Variables

| Variable | Purpose |
|----------|---------|
| `NEO4J_URI` | Default Neo4j connection URI |
| `NEO4J_USERNAME` | Default Neo4j username |
| `NEO4J_PASSWORD` | Default Neo4j password |

### CLI Flags

Common flags across commands:
- `--path, -p`: Path to source definitions (default: `.`)
- `--output, -o`: Output file path (default: stdout)
- `--format, -f`: Output format (cypher, json, table)
- `--verbose, -v`: Enable verbose output

## Testing

### Unit Tests

Each package has corresponding `*_test.go` files:
- Schema validation tests
- Serializer output tests
- Lint rule tests
- Discovery tests

Run all tests:
```bash
go test ./...
```

### Integration Tests

The validator package includes integration tests that require a running Neo4j instance:
```bash
NEO4J_URI=bolt://localhost:7687 go test ./internal/validator/...
```

## Related Documentation

- [CLI Reference](cli/) - Command-line interface documentation
- [Lint Rules](lint-rules/) - Complete lint rule reference
- [Quick Start](quick-start/) - Getting started guide
- [FAQ](faq/) - Frequently asked questions
- [Kiro Integration](NEO4J-KIRO-CLI.md) - AI-assisted development
