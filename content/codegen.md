---
title: "Codegen"
---

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./wetwire-dark.svg">
  <img src="./wetwire-light.svg" width="100" height="67">
</picture>

This document describes the code generation approach in wetwire-neo4j-go.

---

## Overview

Unlike AWS CloudFormation (which generates types from a spec), wetwire-neo4j uses hand-crafted Go types that model Neo4j concepts. The "code generation" in this context refers to generating Cypher DDL and GDS procedure calls from Go declarations.

---

## Type Definitions

### Schema Types

Schema types are defined in `pkg/neo4j/schema/`:

```go
// NodeType represents a Neo4j node label with properties and constraints
type NodeType struct {
    Label       string
    Properties  []Property
    Constraints []Constraint
    Indexes     []Index
}

// RelationshipType represents a Neo4j relationship type
type RelationshipType struct {
    Label      string
    Source     string
    Target     string
    Properties []Property
}

// Property represents a node or relationship property
type Property struct {
    Name     string
    Type     PropertyType
    Required bool
}
```

### Algorithm Types

GDS algorithm types are defined in `internal/algorithms/`:

```go
// BaseAlgorithm contains common algorithm configuration
type BaseAlgorithm struct {
    GraphName string
    Mode      Mode // Stream, Write, Mutate, Stats
}

// PageRank computes PageRank centrality scores
type PageRank struct {
    BaseAlgorithm
    DampingFactor float64
    MaxIterations int
    Tolerance     float64
}
```

---

## Cypher Generation Pipeline

```
┌─────────────────────────────────────────────────────────────┐
│                   Cypher Generation Pipeline                 │
│                                                              │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────────┐  │
│  │   DISCOVER  │ ─▶ │  SERIALIZE  │ ─▶ │     OUTPUT      │  │
│  └─────────────┘    └─────────────┘    └─────────────────┘  │
│                                                              │
│  AST parsing        Convert to         Write .cypher files  │
│  finds types        Cypher DDL         or execute directly  │
└─────────────────────────────────────────────────────────────┘
```

### Stage 1: Discover

The discovery phase uses Go's AST package:

```go
import "github.com/lex00/wetwire-neo4j-go/internal/discover"

resources, err := discover.DiscoverAll("./schemas/...")
```

### Stage 2: Serialize

The serializer converts Go structs to Cypher:

```go
import "github.com/lex00/wetwire-neo4j-go/internal/serializer"

// Generate constraint Cypher
cypher := serializer.ConstraintToCypher(constraint, "Person")
// CREATE CONSTRAINT person_id_unique FOR (n:Person) REQUIRE n.id IS UNIQUE

// Generate index Cypher
cypher := serializer.IndexToCypher(index, "Person")
// CREATE INDEX person_name_idx FOR (n:Person) ON (n.name)
```

### Stage 3: Output

Write Cypher to files or execute directly:

```bash
# Write to file
wetwire-neo4j build ./schemas/ -o schema.cypher

# Execute against Neo4j
wetwire-neo4j build ./schemas/ --execute \
  --uri bolt://localhost:7687 \
  --user neo4j --password password
```

---

## Generated Cypher Examples

### Constraints

```go
// Input
var Person = &schema.NodeType{
    Label: "Person",
    Constraints: []schema.Constraint{
        {Type: schema.Unique, Properties: []string{"id"}},
        {Type: schema.Exists, Properties: []string{"name"}},
    },
}

// Output
// CREATE CONSTRAINT person_id_unique FOR (n:Person) REQUIRE n.id IS UNIQUE;
// CREATE CONSTRAINT person_name_exists FOR (n:Person) REQUIRE n.name IS NOT NULL;
```

### Indexes

```go
// Input
var Person = &schema.NodeType{
    Label: "Person",
    Indexes: []schema.Index{
        {Properties: []string{"name"}},
        {Properties: []string{"email"}, Type: schema.FullText},
    },
}

// Output
// CREATE INDEX person_name_idx FOR (n:Person) ON (n.name);
// CREATE FULLTEXT INDEX person_email_ft FOR (n:Person) ON EACH [n.email];
```

### GDS Algorithms

```go
// Input
var Influence = &algorithms.PageRank{
    BaseAlgorithm: algorithms.BaseAlgorithm{
        GraphName: "social",
        Mode:      algorithms.Stream,
    },
    DampingFactor: 0.85,
    MaxIterations: 20,
}

// Output
// CALL gds.pageRank.stream('social', {
//   dampingFactor: 0.85,
//   maxIterations: 20
// })
// YIELD nodeId, score
// RETURN gds.util.asNode(nodeId) AS node, score
// ORDER BY score DESC
```

---

## Property Type Mapping

| Go Constant | Neo4j Type | Cypher |
|-------------|------------|--------|
| `TypeString` | String | `String` |
| `TypeInteger` | Integer | `Integer` |
| `TypeFloat` | Float | `Float` |
| `TypeBoolean` | Boolean | `Boolean` |
| `TypeDate` | Date | `Date` |
| `TypeDateTime` | DateTime | `DateTime` |
| `TypePoint` | Point | `Point` |
| `TypeListString` | List<String> | `List<String>` |

---

## Validation

After generation, verify the output:

```bash
# Check syntax
wetwire-neo4j lint ./schemas/...

# Build and review
wetwire-neo4j build ./schemas/...

# Validate against Neo4j
wetwire-neo4j validate ./schemas/ --uri bolt://localhost:7687
```

---

## See Also

- [Developer Guide](developers/) - Development workflow
- [Internals](internals/) - Architecture details
- [CLI Reference](cli/) - Build command options
