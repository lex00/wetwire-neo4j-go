# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Schema definition system in `pkg/neo4j/schema`
  - `NodeType` struct for defining node labels with properties, constraints, and indexes
  - `RelationshipType` struct for defining relationships with source/target and cardinality
  - `Property` struct with Neo4j types (STRING, INTEGER, FLOAT, BOOLEAN, DATE, DATETIME, POINT, LIST_*)
  - `Constraint` types: UNIQUE, EXISTS, NODE_KEY, REL_KEY
  - `Index` types: BTREE, TEXT, FULLTEXT, POINT, VECTOR
  - `Validator` for schema validation with PascalCase/SCREAMING_SNAKE_CASE enforcement
  - Full test coverage for types and validation

- Discovery system in `internal/discovery`
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

## [0.1.0] - 2026-01-11

### Added

- Initial project scaffold
- Project structure aligned with wetwire-core-go infrastructure
- CI/CD workflows (test, lint, release)
- Basic README with project overview and examples
