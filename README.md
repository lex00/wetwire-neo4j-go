# wetwire-neo4j-go

Neo4j configuration synthesis for Go - type-safe declarations that compile to Cypher and JSON configurations.

## Overview

wetwire-neo4j is a **synthesis library** for Neo4j beyond Cypher queries. It targets:

1. **GDS Algorithm Configurations** - PageRank, Louvain, FastRP with typed parameters
2. **ML Pipelines** - Node classification, link prediction, node regression
3. **GraphRAG Pipelines** - Retrievers, knowledge graph construction, entity resolution
4. **Schema Definitions** - Node types, relationships, constraints, taxonomy structures

```
Go Structs → wetwire-neo4j build → Cypher + JSON → Neo4j GDS/GraphRAG APIs
```

## Installation

```bash
go install github.com/lex00/wetwire-neo4j-go/cmd/wetwire-neo4j@latest
```

## Quick Example

```go
package main

import (
    "github.com/lex00/wetwire-neo4j-go/pkg/neo4j"
)

// Schema definition
type Person struct {
    neo4j.NodeType
    Label string `neo4j:"Person"`
}

func (p Person) Properties() []neo4j.Property {
    return []neo4j.Property{
        {Name: "id", Type: neo4j.String, Required: true, Unique: true},
        {Name: "name", Type: neo4j.String, Required: true},
        {Name: "email", Type: neo4j.String, Unique: true},
    }
}

// Algorithm configuration
type InfluenceScore struct {
    neo4j.PageRank
    DampingFactor float64 `neo4j:"dampingFactor" default:"0.85"`
    MaxIterations int     `neo4j:"maxIterations" default:"50"`
    Tolerance     float64 `neo4j:"tolerance" default:"1e-7"`
}
```

## Commands

```bash
wetwire-neo4j build ./schemas/    # Generate Cypher constraints + JSON configs
wetwire-neo4j lint ./schemas/     # Validate schemas, algorithms, pipelines
wetwire-neo4j list                # List all definitions
wetwire-neo4j validate            # Check against live Neo4j instance
wetwire-neo4j import              # Import existing constraints from Neo4j
wetwire-neo4j graph               # Visualize schema as Mermaid/DOT
```

## Status

Under development - see [ROADMAP](https://github.com/lex00/wetwire-neo4j-go/issues/18)
