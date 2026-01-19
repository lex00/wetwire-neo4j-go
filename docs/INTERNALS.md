# Internals

This document covers the internal architecture of wetwire-neo4j-go for contributors and maintainers.

**Contents:**
- [AST Discovery](#ast-discovery) - How resource discovery works
- [Cypher Generation](#cypher-generation) - How Cypher statements are built
- [Linter Architecture](#linter-architecture) - How lint rules work
- [Importer](#importer) - Cypher to Go conversion
- [Validator](#validator) - Neo4j instance validation

---

## AST Discovery

wetwire-neo4j uses Go's `go/ast` package to discover schema and algorithm declarations without executing user code.

### How It Works

When you define a schema resource as a package-level variable:

```go
var Person = &schema.NodeType{
    Label: "Person",
    Properties: []schema.Property{
        {Name: "id", Type: schema.TypeString, Required: true},
    },
}
```

The discovery phase:
1. Parses Go source files using `go/parser`
2. Walks the AST looking for `var` declarations
3. Identifies composite literals with schema/algorithm types
4. Extracts metadata: name, type, file, line, properties

### Discovery API

```go
import "github.com/lex00/wetwire-neo4j-go/internal/discover"

resources, err := discover.DiscoverAll("./schemas/...")

// Access discovered resources
for _, node := range resources.NodeTypes {
    fmt.Printf("%s: %s at %s:%d\n", node.Name, node.Label, node.File, node.Line)
}
```

### What Gets Discovered

| Type | Example | Discovered As |
|------|---------|---------------|
| NodeType | `var Person = &schema.NodeType{...}` | NodeType |
| RelationshipType | `var WorksFor = &schema.RelationshipType{...}` | RelationshipType |
| PageRank | `var Influence = &algorithms.PageRank{...}` | Algorithm |
| NativeProjection | `var SocialGraph = &projections.NativeProjection{...}` | Projection |

---

## Cypher Generation

The serializer generates Cypher statements from discovered resources.

### Build Process

```go
import "github.com/lex00/wetwire-neo4j-go/internal/serializer"

// Generate Cypher for a node type
cypher := serializer.NodeTypeToCypher(person)
// Returns: CREATE CONSTRAINT person_id_unique FOR (n:Person) REQUIRE n.id IS UNIQUE
```

### Constraint Generation

```go
// Input
var Person = &schema.NodeType{
    Label: "Person",
    Constraints: []schema.Constraint{
        {Type: schema.Unique, Properties: []string{"id"}},
    },
}

// Output
// CREATE CONSTRAINT person_id_unique FOR (n:Person) REQUIRE n.id IS UNIQUE
```

### Index Generation

```go
// Input
var Person = &schema.NodeType{
    Label: "Person",
    Indexes: []schema.Index{
        {Properties: []string{"name"}},
    },
}

// Output
// CREATE INDEX person_name_idx FOR (n:Person) ON (n.name)
```

### GDS Algorithm Generation

```go
// Input
var Influence = &algorithms.PageRank{
    BaseAlgorithm: algorithms.BaseAlgorithm{
        GraphName: "social",
        Mode:      algorithms.Stream,
    },
    DampingFactor: 0.85,
}

// Output
// CALL gds.pageRank.stream('social', {dampingFactor: 0.85})
// YIELD nodeId, score
// RETURN gds.util.asNode(nodeId).name AS name, score
// ORDER BY score DESC
```

---

## Linter Architecture

The linter checks Go source for style issues and potential problems.

### Rule Structure

Each rule has:
- **ID**: `WN4001` through `WN4069`
- **Severity**: error, warning, or info
- **Check function**: Analyzes discovered resources

```go
type LintResult struct {
    Rule     string
    Severity string
    Message  string
    Location string
    File     string
    Line     int
}
```

### Rule Categories

| Range | Category |
|-------|----------|
| WN4001-WN4029 | GDS Algorithm Rules |
| WN4030-WN4039 | ML Pipeline Rules |
| WN4040-WN4049 | GraphRAG Rules |
| WN4050-WN4059 | Schema Rules |
| WN4060-WN4069 | Projection Rules |

### Key Rules

| ID | Description |
|----|-------------|
| WN4001 | dampingFactor must be in [0, 1) |
| WN4002 | maxIterations must be positive |
| WN4006 | embeddingDimension should be power of 2 |
| WN4052 | Node labels should be PascalCase |
| WN4053 | Relationship types should be SCREAMING_SNAKE_CASE |

### Running the Linter

```go
import "github.com/lex00/wetwire-neo4j-go/internal/lint"

linter := lint.NewLinter()
results := linter.LintNodeType(nodeType)

for _, r := range results {
    fmt.Printf("%s: [%s] %s\n", r.Location, r.Rule, r.Message)
}
```

---

## Importer

The importer converts existing Cypher DDL to Go code.

### Import Process

```go
import "github.com/lex00/wetwire-neo4j-go/internal/importer"

importer := importer.NewCypherImporter("schema.cypher")
result, err := importer.Import(ctx)
```

### Code Generation

```cypher
# Input: Cypher DDL
CREATE CONSTRAINT person_id_unique FOR (p:Person) REQUIRE p.id IS UNIQUE;
CREATE INDEX person_name_idx FOR (p:Person) ON (p.name);
```

```go
// Output: Go code
package schema

import "github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"

var Person = &schema.NodeType{
    Label: "Person",
    Properties: []schema.Property{
        {Name: "id", Type: schema.TypeString, Required: true},
        {Name: "name", Type: schema.TypeString},
    },
    Constraints: []schema.Constraint{
        {Type: schema.Unique, Properties: []string{"id"}},
    },
    Indexes: []schema.Index{
        {Properties: []string{"name"}},
    },
}
```

---

## Validator

The validator checks generated Cypher against a live Neo4j instance.

### Validation Process

```go
import "github.com/lex00/wetwire-neo4j-go/internal/validator"

v := validator.New(validator.Config{
    URI:      "bolt://localhost:7687",
    User:     "neo4j",
    Password: "password",
})

results, err := v.Validate(ctx, resources)
```

### What Gets Validated

1. **Constraint conflicts** - Duplicate constraint names
2. **Index conflicts** - Duplicate index names
3. **Label existence** - Referenced labels exist
4. **Property types** - Property type compatibility

---

## Files Reference

| File | Purpose |
|------|---------|
| `pkg/neo4j/schema/types.go` | NodeType, RelationshipType, Property |
| `internal/algorithms/algorithms.go` | GDS algorithm types |
| `internal/discover/discover.go` | AST-based discovery |
| `internal/serializer/cypher.go` | Cypher generation |
| `internal/lint/lint.go` | Lint rules |
| `internal/importer/cypher.go` | Cypher parser |
| `internal/validator/validator.go` | Neo4j validation |

---

## See Also

- [Developer Guide](DEVELOPERS.md) - Development workflow
- [Lint Rules](LINT_RULES.md) - Complete rule reference
- [CLI Reference](CLI.md) - CLI commands
