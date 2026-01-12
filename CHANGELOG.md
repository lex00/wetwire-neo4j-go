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

## [0.1.0] - 2026-01-11

### Added

- Initial project scaffold
- Project structure aligned with wetwire-core-go infrastructure
- CI/CD workflows (test, lint, release)
- Basic README with project overview and examples
