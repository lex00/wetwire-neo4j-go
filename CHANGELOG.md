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

## [0.1.0] - 2026-01-11

### Added

- Initial project scaffold
- Project structure aligned with wetwire-core-go infrastructure
- CI/CD workflows (test, lint, release)
- Basic README with project overview and examples
