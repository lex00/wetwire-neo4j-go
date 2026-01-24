---
title: "Examples"
---

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./wetwire-dark.svg">
  <img src="./wetwire-light.svg" width="100" height="67">
</picture>

The `examples/` directory contains reference implementations demonstrating wetwire-neo4j patterns. These serve as:

1. **Learning resources** - See how to define schemas and algorithms
2. **Test artifacts** - Validate the build pipeline
3. **Starting points** - Copy and modify for your needs

## Directory Structure

```
examples/
├── social/                 # Social network schema
│   ├── nodes.go           # Person, Company, etc.
│   ├── relationships.go   # KNOWS, WORKS_FOR, etc.
│   └── algorithms.go      # PageRank, community detection
│
├── e-commerce/            # E-commerce schema
│   ├── schema.go          # Product, Order, Customer
│   └── recommendations.go # Recommendation algorithms
│
└── knowledge-graph/       # Knowledge graph example
    ├── schema.go          # Entity, Concept types
    ├── retrievers.go      # GraphRAG retrievers
    └── pipelines.go       # ML pipeline definitions
```

## Social Network Example

A social network schema demonstrating core patterns.

### nodes.go

```go
package social

import "github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"

var Person = &schema.NodeType{
    Label: "Person",
    Properties: []schema.Property{
        {Name: "id", Type: schema.TypeString, Required: true},
        {Name: "name", Type: schema.TypeString, Required: true},
        {Name: "email", Type: schema.TypeString},
        {Name: "joinedAt", Type: schema.TypeDateTime},
    },
    Constraints: []schema.Constraint{
        {Type: schema.Unique, Properties: []string{"id"}},
    },
    Indexes: []schema.Index{
        {Properties: []string{"name"}},
        {Properties: []string{"email"}},
    },
}

var Company = &schema.NodeType{
    Label: "Company",
    Properties: []schema.Property{
        {Name: "name", Type: schema.TypeString, Required: true},
        {Name: "industry", Type: schema.TypeString},
    },
}
```

### relationships.go

```go
package social

import "github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"

var Knows = &schema.RelationshipType{
    Label:  "KNOWS",
    Source: "Person",
    Target: "Person",
    Properties: []schema.Property{
        {Name: "since", Type: schema.TypeDate},
    },
}

var WorksFor = &schema.RelationshipType{
    Label:  "WORKS_FOR",
    Source: "Person",
    Target: "Company",
    Properties: []schema.Property{
        {Name: "role", Type: schema.TypeString},
        {Name: "startDate", Type: schema.TypeDate},
    },
}
```

### algorithms.go

```go
package social

import "github.com/lex00/wetwire-neo4j-go/internal/algorithms"

var SocialGraph = &projections.NativeProjection{
    Name: "social-network",
    NodeProjections: []projections.NodeProjection{
        {Label: "Person"},
    },
    RelationshipProjections: []projections.RelationshipProjection{
        {Type: "KNOWS"},
    },
}

var Influence = &algorithms.PageRank{
    BaseAlgorithm: algorithms.BaseAlgorithm{
        GraphName: "social-network",
        Mode:      algorithms.Stream,
    },
    DampingFactor: 0.85,
    MaxIterations: 20,
}

var Communities = &algorithms.Louvain{
    BaseAlgorithm: algorithms.BaseAlgorithm{
        GraphName: "social-network",
        Mode:      algorithms.Write,
    },
    WriteProperty: "community",
}
```

## Using Examples

### View the generated Cypher

```bash
cd examples/social
wetwire-neo4j build .
```

### List resources

```bash
cd examples/social
wetwire-neo4j list .
```

### Copy as starting point

```bash
cp -r examples/social ./my-schema
cd my-schema
# Edit files
wetwire-neo4j build .
```

## Notable Examples

| Example | Description |
|---------|-------------|
| `social/` | Social network with PageRank and community detection |
| `e-commerce/` | Product catalog with recommendation algorithms |
| `knowledge-graph/` | Entity extraction with GraphRAG retrievers |

## Building Examples

To build and verify all examples:

```bash
# Build all
for dir in examples/*/; do
    echo "Building $dir..."
    wetwire-neo4j build "$dir" || exit 1
done

# Lint all
wetwire-neo4j lint ./examples/...
```

## Notes

- Examples demonstrate patterns, not production-ready schemas
- Some examples use advanced GDS features requiring the GDS plugin
- Run `wetwire-neo4j lint ./examples/...` to check for issues

## See Also

- [Quick Start](quick-start/) - Getting started guide
- [Developer Guide](developers/) - Development workflow
- [Internals](internals/) - How discovery works
