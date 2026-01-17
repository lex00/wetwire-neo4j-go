# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- Split large lint test file for maintainability
  - `lint_test.go` (830 lines) split into focused files:
    - `lint_test.go` - Basic tests, utilities, configuration (~170 lines)
    - `lint_algorithms_test.go` - Algorithm rule tests (~430 lines)
    - `lint_schema_test.go` - Schema definition tests (~245 lines)
  - No test file over 800 lines

- Renamed `internal/discovery` to `internal/discover` (#104)
  - Updated package name from `discovery` to `discover`
  - Updated all imports across the codebase

- Migrated discovery package to use `wetwire-core-go/ast` utilities
  - Replaced local `getTypeName` with `coreast.ExtractTypeName`
  - Replaced local `isPrimitiveType` with `coreast.IsBuiltinType`
  - Replaced `isPrimitiveType` in `walkExprForDeps` with `coreast.IsBuiltinIdent`
  - Updated wetwire-core-go dependency from v1.15.0 to v1.16.0
  - Reduces code duplication and aligns with shared AST utilities

- **BREAKING**: Migrated MCP server to use `domain.BuildMCPServer()` auto-generation
  - Updated wetwire-core-go dependency from v1.12.0 to v1.13.0
  - Replaced 592-line manual MCP implementation with 24-line auto-generated version
  - MCP tools now auto-registered based on domain interface implementations
  - Removed `wetwire-neo4j import` command (Neo4jDomain doesn't implement ImporterDomain)
  - All core MCP tools (init, build, lint, validate, list, graph) remain functional
  - Fixes domain validator compliance issues (Issue #94)


## [1.5.12] - 2026-01-13

### Added

- `wetwire_read` MCP tool for reading schema source files
  - Allows agents to read full file contents when wetwire_list summary isn't sufficient
  - Security: Only allows safe file types (.go, .mod, .cypher, .json, .yaml, .md, .txt)

## [1.5.11] - 2026-01-13

### Added

- Discovery scanner extracts `AgentContext` from Schema definitions
  - AI agents now see schema-defined instructions in their prompts
  - Supports both regular strings and raw strings (backticks)
  - Displayed prominently as "Agent Instructions" section

## [1.5.10] - 2026-01-13

### Fixed

- Importer prevents variable name collisions between nodes and relationships
  - Adds "Rel" suffix to relationship variables when names would collide
  - Fixes redeclaration and type mismatch errors
  - Neo4j labels preserved correctly, only Go variable names disambiguated

## [1.5.9] - 2026-01-13

### Added

- Property details in `wetwire_list` MCP tool output
  - Shows properties with name, type, and required flag
  - Shows constraints with type and properties
  - Shows indexes with type and properties
  - Shows source/target for relationship types
  - AST extraction methods in `internal/discover`

### Fixed

- Importer generates valid Go code for nodes with multiple constraints/indexes
  - Previously wrote duplicate struct fields causing compile errors
  - Now combines all constraints and indexes into single blocks

## [1.5.8] - 2026-01-13

### Added

- Pre-flight schema discovery for agent prompt injection in `cmd/wetwire-neo4j`
  - Automatically discovers existing schema definitions before running `design` command
  - Injects summarized schema context into agent prompts (both Anthropic and Kiro providers)
  - Helps agents extend existing schemas rather than recreating them
  - `FormatSchemaContext()` function in `internal/discover/context.go`
  - `Neo4jDomainWithContext()` and `NewConfigWithContext()` for context-aware configs

- MCP server integration in `cmd/wetwire-neo4j`
  - Uses wetwire-core-go mcp package for protocol handling
  - `wetwire-neo4j design --mcp-server` starts MCP server on stdio
  - Implements standard wetwire tools: init, build, lint, validate, list, graph
  - JSON schema definitions for all tool inputs
  - Enables Claude Code integration via MCP protocol

- Expanded validator test coverage in `internal/validator`
  - 27 test functions covering helper functions and structures
  - Tests for FormatResults, HasErrors, FilterInvalid, existsMsg
  - Tests for Config, ValidationResult, GDSInfo, DatabaseInfo structures
  - MockValidator tests for labelExists, relationshipTypeExists
  - Full coverage of testable code (Neo4j connection-dependent code requires integration tests)

- CLI command tests in `cmd/wetwire-neo4j`
  - Tests for list, validate, import, graph, version commands
  - Tests for command flags and execution
  - 14 test functions covering CLI layer
  - Updated wetwire-core-go dependency to v1.5.0

- Validator implementing `cmd.Validator` interface in `internal/cli`
  - `Validator` type satisfies wetwire-core-go interface pattern
  - Uses environment variables for configuration (NEO4J_URI, NEO4J_USERNAME, etc.)
  - Returns `cmd.ValidationError` slice for consistent error handling
  - Complements existing `ValidatorCLI` for custom validate command
  - Full test coverage (4 test functions)

- Init command implementing `cmd.Initializer` interface in `internal/cli`
  - `wetwire-neo4j init <project-name>` creates new project scaffold
  - Templates: default, gds, graphrag, full
  - Generates schema, algorithms, pipelines, retrievers, kg directories
  - `--force` flag to overwrite existing directories
  - `--template` flag to select project template
  - Full test coverage (6 test functions)

- Aura Graph Analytics session configuration in `internal/aura`
  - `Session` struct for Aura Analytics serverless GDS sessions
  - `PandasDataSource` for loading data from CSV files via pandas
  - `SnowflakeDataSource` for loading data from Snowflake
  - `BigQueryDataSource` for loading data from Google BigQuery
  - `Serializer` for generating Python graphdatascience client code
  - JSON/Map export for session configurations
  - Session validation (name, TTL, data source requirements)
  - Full test coverage (7 test functions)

- Schema definition system in `pkg/neo4j/schema`
  - `NodeType` struct for defining node labels with properties, constraints, and indexes
  - `RelationshipType` struct for defining relationships with source/target and cardinality
  - `Property` struct with Neo4j types (STRING, INTEGER, FLOAT, BOOLEAN, DATE, DATETIME, POINT, LIST_*)
  - `Constraint` types: UNIQUE, EXISTS, NODE_KEY, REL_KEY
  - `Index` types: BTREE, TEXT, FULLTEXT, POINT, VECTOR
  - `Validator` for schema validation with PascalCase/SCREAMING_SNAKE_CASE enforcement
  - Full test coverage for types and validation

- Discovery system in `internal/discover`
  - `Scanner` for AST-based resource discovery in Go source files
  - Detects NodeType, RelationshipType, Algorithm, Pipeline, and Retriever definitions
  - Supports struct types with embedded resource types and variable declarations
  - `DependencyGraph` with topological sort for dependency ordering
  - Cycle detection for circular dependencies
  - Skips test files and vendor directories
  - Full test coverage (20+ test functions)

- Serializer system in `internal/serializer`
  - `CypherSerializer` for generating Cypher constraint and index statements
  - Supports UNIQUE, EXISTS, NODE_KEY, REL_KEY constraints
  - Supports BTREE, TEXT, FULLTEXT, POINT, VECTOR indexes
  - Template-based Cypher generation with proper escaping
  - `JSONSerializer` for schema export to JSON format
  - `NodeTypeToMap` and `RelationshipTypeToMap` for flexible serialization
  - Full test coverage (26 test functions)

- GDS algorithm configurations in `internal/algorithms`
  - Centrality algorithms: PageRank, ArticleRank, Betweenness, Degree, Closeness
  - Community detection: Louvain, Leiden, LabelPropagation, WCC, TriangleCount, KCore
  - Similarity algorithms: NodeSimilarity, KNN
  - Embeddings: FastRP, Node2Vec, GraphSAGE, HashGNN
  - Path finding: Dijkstra, AStar, BFS, DFS
  - `AlgorithmSerializer` for Cypher CALL statements and JSON output
  - Support for stream, stats, mutate, and write execution modes
  - Full test coverage (30+ test functions)

- GDS ML pipeline configurations in `internal/pipelines`
  - `NodeClassificationPipeline` for categorical label prediction
  - `LinkPredictionPipeline` for relationship prediction
  - `NodeRegressionPipeline` for numeric property prediction
  - Feature steps: FastRP, PageRank, Degree, Node2Vec, Scaler
  - Models: LogisticRegression, RandomForest, MLP, LinearRegression
  - `PipelineSerializer` for Cypher pipeline creation and training commands
  - JSON/Map export for pipeline configurations
  - Support for split config and auto-tuning
  - Full test coverage (18 test functions)

- GDS graph projection configurations in `internal/projections`
  - `NativeProjection` for projecting node labels and relationship types
  - `CypherProjection` for custom Cypher-based projections
  - `DataFrameProjection` for Aura Analytics integration
  - `NodeProjection` and `RelationshipProjection` for detailed configuration
  - Orientation options: NATURAL, REVERSE, UNDIRECTED
  - Aggregation options: NONE, SUM, MIN, MAX, SINGLE, COUNT
  - `ProjectionSerializer` for Cypher generation and JSON export
  - Graph management utilities: DropGraph, GraphExists, ListGraphs
  - Full test coverage (19 test functions)

- GraphRAG retriever configurations in `internal/retrievers`
  - `VectorRetriever` for similarity search using vector indexes
  - `VectorCypherRetriever` for vector search with custom graph traversal
  - `HybridRetriever` combining vector and fulltext search
  - `HybridCypherRetriever` for hybrid search with custom traversal
  - `Text2CypherRetriever` for LLM-generated Cypher queries
  - External integrations: Weaviate, Pinecone, Qdrant
  - `RetrieverSerializer` for JSON configuration export
  - Support for embedder configuration and few-shot examples
  - Full test coverage (20 test functions)

- GraphRAG knowledge graph construction pipelines in `internal/kg`
  - `SimpleKGPipeline` for standard entity and relationship extraction
  - `CustomKGPipeline` for custom extraction with user-defined prompts
  - Entity types with properties and descriptions
  - Relationship types with source/target constraints
  - Text splitters: FixedSizeSplitter, LangChainSplitter
  - Entity resolvers: ExactMatch, FuzzyMatch, SemanticMatch
  - `KGSerializer` for JSON configuration export
  - Schema generation for LLM extraction prompts
  - Full test coverage (22 test functions)

- Lint rules for Neo4j definitions in `internal/lint`
  - WN4001: dampingFactor must be in [0, 1)
  - WN4002: maxIterations must be positive
  - WN4005: tolerance warning if too loose
  - WN4006: embeddingDimension should be power of 2
  - WN4007: topK warning if > 1000
  - WN4032: pipeline requires at least one model
  - WN4040: KG pipeline requires entity types
  - WN4043: entity resolver threshold warning
  - WN4052: node labels should be PascalCase
  - WN4053: relationship types should be SCREAMING_SNAKE_CASE
  - `Linter` with LintAlgorithm, LintPipeline, LintKGPipeline
  - HasErrors, FilterBySeverity, FormatResults utilities
  - Full test coverage (15 test functions)

- CLI implementation in `internal/cli`
  - `Builder` implementing wetwire-core-go cmd.Builder interface
  - `Linter` implementing wetwire-core-go cmd.Linter interface
  - `Lister` for discovering and listing Neo4j definitions
  - Support for Cypher and JSON output formats
  - Table and JSON listing formats
  - Dependency graph visualization
  - Full test coverage (27 test functions)

- CLI entry point in `cmd/wetwire-neo4j/main.go`
  - `wetwire-neo4j build` - Build Cypher/JSON from definitions
  - `wetwire-neo4j lint` - Lint definitions for issues
  - `wetwire-neo4j list` - List discovered definitions
  - `wetwire-neo4j validate` - Validate against live Neo4j instance
  - `wetwire-neo4j import` - Import schemas from Cypher files or Neo4j
  - `wetwire-neo4j graph` - Visualize resource dependencies (DOT, Mermaid)
  - `wetwire-neo4j version` - Show version information

- External validation in `internal/validator`
  - `Validator` for validating configurations against live Neo4j instance
  - Schema validation (node labels, relationship types, constraints, indexes)
  - GDS algorithm validation (algorithm existence, graph catalog checks)
  - Graph projection validation (node labels, relationship types)
  - Database and GDS version detection
  - Full test coverage (14 test functions)

- Schema importer in `internal/importer`
  - `CypherImporter` for parsing Cypher constraint/index statements
  - `Neo4jImporter` for importing from live Neo4j databases
  - `Generator` for generating Go code from imported schemas
  - Supports UNIQUE, NODE_KEY, EXISTS constraints
  - Supports RANGE, FULLTEXT, TEXT, VECTOR indexes
  - Full test coverage (8 test functions)

- Reference examples in `examples/`
  - Schema definitions (Person, Company, WORKS_FOR, KNOWS)
  - GDS algorithms (PageRank, Louvain, FastRP, Node2Vec, KNN, Dijkstra)
  - ML pipelines (NodeClassification, LinkPrediction, NodeRegression)
  - GraphRAG retrievers (Vector, VectorCypher, Hybrid, Text2Cypher)
  - KG pipelines (SimpleKG, CustomKG)
  - Graph projections (Native, Cypher)
  - Integration tests validating all examples

- Documentation
  - `CLAUDE.md` - AI assistant guidelines
  - `docs/CLI.md` - CLI command reference
  - `docs/FAQ.md` - Frequently asked questions
  - `docs/QUICK_START.md` - Getting started guide
  - `docs/LINT_RULES.md` - WN4xxx lint rule documentation
  - Updated `README.md` with architecture and quick start

## [0.1.0] - 2026-01-11

### Added

- Initial project scaffold
- Project structure aligned with wetwire-core-go infrastructure
- CI/CD workflows (test, lint, release)
- Basic README with project overview and examples
