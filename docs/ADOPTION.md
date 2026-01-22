<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./wetwire-dark.svg">
  <img src="./wetwire-light.svg" width="100" height="67">
</picture>

Practical guidance for teams adopting wetwire-neo4j alongside existing Neo4j infrastructure.

---

## Migration Strategies

### Side-by-Side Adoption

You don't need to migrate everything at once. wetwire-neo4j generates standard Cypher DDL that runs with the same tools you already use.

**Coexistence patterns:**

| Existing Approach | Integration |
|-------------------|-------------|
| Raw Cypher DDL | Keep existing scripts; add new schemas in Go |
| Neo4j Browser | Continue using Browser; wetwire generates compatible Cypher |
| cypher-shell | wetwire output works directly with cypher-shell |
| Neo4j Migrations | Use wetwire alongside migration tools |

### Incremental Migration Path

**Week 1: Proof of concept**
- Pick a small, isolated schema (new node type, new index)
- Write it in wetwire-neo4j
- Verify the generated Cypher output
- Apply to a test environment

**Week 2-4: Build confidence**
- Convert 2-3 more schemas
- Establish team patterns (file organization, naming conventions)
- Set up CI/CD for schema validation

**Ongoing: New schemas in Go**
- All new schema changes use wetwire-neo4j
- Migrate legacy schemas opportunistically

### What NOT to Migrate

Some schemas are better left alone:
- **Stable production schemas** that never change
- **Third-party integrations** with their own schema management
- **Schemas managed by Neo4j tools** (Bloom, etc.)

Migration should reduce maintenance burden, not create it.

---

## Escape Hatches

When you hit an edge case the library doesn't handle cleanly.

### Raw Cypher Passthrough

For constraints or indexes not yet typed:

```go
var CustomConstraint = &schema.RawCypher{
    Statement: "CREATE CONSTRAINT custom_constraint FOR (n:Custom) REQUIRE n.prop IS TYPED STRING",
}
```

### Custom Algorithm Configurations

For GDS procedures with unsupported parameters:

```go
var CustomAlgo = &algorithms.Custom{
    Procedure: "gds.myCustomAlgo.stream",
    Config: map[string]any{
        "customParam": "value",
        "anotherParam": 42,
    },
}
```

### When to File an Issue

If you're using escape hatches for:
- Common constraint types
- Standard GDS algorithms
- Patterns other teams would need

...file an issue. The library should handle it.

---

## Team Onboarding

A playbook for getting your team productive in the first week.

### Day 1: Environment Setup

```bash
# Clone your schema repo
git clone <repo>
cd <repo>

# Install wetwire-neo4j CLI
go install github.com/lex00/wetwire-neo4j-go/cmd/wetwire-neo4j@latest

# Verify it works
wetwire-neo4j list ./schemas/... && echo "OK"
```

**What to check:**
- Go 1.21+ installed
- wetwire-neo4j CLI available in PATH
- Neo4j credentials configured (for validation)

### Day 1-2: Read the Code

Start with a schema file:

```go
package schemas

import "github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"

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

That's the pattern. Every schema file looks like this.

### Day 2-3: Make a Small Change

Find something low-risk:
- Add a property to an existing node type
- Add an index
- Create a new relationship type

```go
// Before
var Person = &schema.NodeType{
    Label: "Person",
    Properties: []schema.Property{
        {Name: "id", Type: schema.TypeString, Required: true},
    },
}

// After
var Person = &schema.NodeType{
    Label: "Person",
    Properties: []schema.Property{
        {Name: "id", Type: schema.TypeString, Required: true},
        {Name: "email", Type: schema.TypeString},
    },
    Indexes: []schema.Index{
        {Properties: []string{"email"}},
    },
}
```

Run it, diff the output, apply to dev.

### Day 3-4: Add a New Schema

Create a new file in the package:

```go
// company.go
package schemas

import "github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"

var Company = &schema.NodeType{
    Label: "Company",
    Properties: []schema.Property{
        {Name: "name", Type: schema.TypeString, Required: true},
    },
}

var WorksFor = &schema.RelationshipType{
    Label:  "WORKS_FOR",
    Source: "Person",
    Target: "Company",
}
```

Schemas auto-register when discovered via AST parsing.

### Day 5: Review the Patterns

By now you've seen:
- Pointer struct literals for types (`&schema.NodeType{...}`)
- PascalCase for node labels
- SCREAMING_SNAKE_CASE for relationship types
- Package-level var declarations

That's 90% of what you need.

### Common Gotchas

| Problem | Solution |
|---------|----------|
| "undefined: schema" | Add import for the schema package |
| "Resource not in output" | Ensure it's a package-level var declaration |
| Lint warning WN4052 | Use PascalCase for node labels |
| Lint warning WN4053 | Use SCREAMING_SNAKE_CASE for relationship types |

### Team Conventions to Establish

Decide these early:
- **File organization**: By domain (social.go, commerce.go) or by type (nodes.go, relationships.go)?
- **Naming**: Match Go variable names to labels?
- **Constraint naming**: Convention for constraint names

Document in your repo's README.

---

## Resources

- [Quick Start](QUICK_START.md) - 5-minute intro
- [CLI Reference](CLI.md) - Build and validate commands
- [Internals](INTERNALS.md) - How AST discovery works

---

## See Also

- [FAQ](FAQ.md) - Common questions and answers
