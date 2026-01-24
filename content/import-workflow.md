---
title: "Import Workflow"
---


This document explains the Cypher import workflow used by wetwire-neo4j-go to convert existing Neo4j schemas and configurations into Go code.

## Overview

The import workflow enables converting existing Neo4j artifacts into wetwire-neo4j Go declarations:

1. **Cypher files** - Schema definitions (constraints, indexes)
2. **Live Neo4j instances** - Introspect running database schemas
3. **JSON configurations** - GDS algorithm configurations

## Import Sources

### Cypher Files

Import schema definitions from `.cypher` files:

```bash
wetwire-neo4j import --file schema.cypher -o schema.go
```

Supported Cypher statements:
- `CREATE CONSTRAINT ... FOR (n:Label) REQUIRE ...`
- `CREATE INDEX ... FOR (n:Label) ON ...`
- Node and relationship type patterns

### Live Neo4j Instance

Introspect schema from a running Neo4j database:

```bash
wetwire-neo4j import --uri bolt://localhost:7687 \
  --user neo4j --password secret \
  -o schema.go
```

This queries:
- `SHOW CONSTRAINTS` - Existing constraints
- `SHOW INDEXES` - Existing indexes
- Node labels and relationship types

### JSON Configurations

Import GDS algorithm configurations from JSON:

```bash
wetwire-neo4j import --file algorithms.json -o algorithms.go
```

## Workflow Steps

```
┌─────────────────────────────────────────────────────────────┐
│                    Import Workflow                           │
│                                                              │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────────┐  │
│  │    PARSE    │ ─▶ │   CONVERT   │ ─▶ │    GENERATE     │  │
│  └─────────────┘    └─────────────┘    └─────────────────┘  │
│                                                              │
│  Parse Cypher or   Convert to IR      Generate Go code      │
│  introspect DB     (intermediate)     with wetwire types    │
└─────────────────────────────────────────────────────────────┘
```

### Stage 1: Parse

The parser reads the input source:
- **Cypher files**: Regex-based statement parsing
- **Neo4j instance**: Driver queries for metadata
- **JSON**: Standard JSON unmarshaling

### Stage 2: Convert

Converts parsed data to intermediate representation:

```go
type SchemaIR struct {
    NodeTypes         []NodeTypeIR
    RelationshipTypes []RelationshipTypeIR
    Constraints       []ConstraintIR
    Indexes           []IndexIR
}
```

### Stage 3: Generate

Generates idiomatic Go code using wetwire patterns:

```go
// Input: CREATE CONSTRAINT person_id_unique FOR (p:Person) REQUIRE p.id IS UNIQUE

// Output:
var Person = &schema.NodeType{
    Label: "Person",
    Properties: []schema.Property{
        {Name: "id", Type: schema.TypeString, Required: true},
    },
    Constraints: []schema.Constraint{
        {Type: schema.Unique, Properties: []string{"id"}},
    },
}
```

## Usage Examples

### Import schema from Cypher file

```bash
# Basic import
wetwire-neo4j import --file schema.cypher -o schema.go

# With custom package name
wetwire-neo4j import --file schema.cypher --package mydb -o schema.go
```

### Import from running Neo4j

```bash
# With authentication
wetwire-neo4j import \
  --uri bolt://localhost:7687 \
  --user neo4j \
  --password password \
  -o schema.go

# With environment variables
export NEO4J_URI="bolt://localhost:7687"
export NEO4J_USER="neo4j"
export NEO4J_PASSWORD="password"
wetwire-neo4j import -o schema.go
```

### Import GDS configurations

```bash
wetwire-neo4j import --file algorithms.json --type gds -o algorithms.go
```

## Output Structure

The importer generates organized Go files:

```go
// schema.go
package neo4jschema

import "github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"

// Node types
var Person = &schema.NodeType{
    Label: "Person",
    Properties: []schema.Property{
        {Name: "id", Type: schema.TypeString, Required: true},
        {Name: "name", Type: schema.TypeString},
    },
    Constraints: []schema.Constraint{
        {Type: schema.Unique, Properties: []string{"id"}},
    },
}

var Company = &schema.NodeType{
    Label: "Company",
    Properties: []schema.Property{
        {Name: "name", Type: schema.TypeString, Required: true},
    },
}

// Relationship types
var WorksFor = &schema.RelationshipType{
    Label:  "WORKS_FOR",
    Source: "Person",
    Target: "Company",
}
```

## Validation

After import, verify the generated code:

```bash
# Check syntax
go build ./...

# Lint for issues
wetwire-neo4j lint ./...

# Build Cypher output
wetwire-neo4j build ./...
```

## Limitations

1. **Property type inference** - Types are inferred from constraint definitions; may need manual adjustment
2. **Complex Cypher** - Only schema DDL statements are supported, not data queries
3. **GDS procedures** - Only stream/write mode configurations are importable

## See Also

- [Developer Guide](developers/) - Development workflow
- [Internals](internals/) - Architecture details
- [CLI Reference](cli/) - Import command options
