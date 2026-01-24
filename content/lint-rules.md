---
title: "Lint Rules"
---


wetwire-neo4j-go includes lint rules to validate Neo4j configurations before deployment. All rules use the `WN4xxx` prefix.

## Rule Categories

| Range | Category |
|-------|----------|
| WN4001-WN4029 | GDS Algorithm Rules |
| WN4030-WN4039 | ML Pipeline Rules |
| WN4040-WN4049 | GraphRAG Rules |
| WN4050-WN4059 | Schema Rules |
| WN4060-WN4069 | Projection Rules |

---

## GDS Algorithm Rules

### WN4001: Damping Factor Range

**Severity:** Error

The `dampingFactor` parameter for PageRank and ArticleRank must be in the range [0, 1).

```go
// Error: damping factor must be < 1.0
algo := &algorithms.PageRank{
    DampingFactor: 1.0, // WN4001: invalid value
}

// Valid
algo := &algorithms.PageRank{
    DampingFactor: 0.85,
}
```

**Reference:** [Neo4j PageRank Documentation](https://neo4j.com/docs/graph-data-science/current/algorithms/page-rank/)

---

### WN4002: Max Iterations Positive

**Severity:** Error

The `maxIterations` parameter must be a positive integer.

```go
// Error: negative iterations
algo := &algorithms.PageRank{
    MaxIterations: -1, // WN4002: invalid value
}

// Valid (0 means use default)
algo := &algorithms.PageRank{
    MaxIterations: 20,
}
```

---

### WN4006: Embedding Dimension Power of Two

**Severity:** Warning

Embedding dimensions for FastRP and Node2Vec should be powers of 2 for optimal performance.

```go
// Warning: 100 is not a power of 2
algo := &algorithms.FastRP{
    EmbeddingDimension: 100, // WN4006: consider using 64 or 128
}

// Recommended
algo := &algorithms.FastRP{
    EmbeddingDimension: 128,
}
```

**Reason:** Powers of 2 enable SIMD optimizations and efficient memory alignment.

---

### WN4007: High TopK Warning

**Severity:** Warning

The `topK` parameter greater than 1000 may impact performance.

```go
// Warning: high topK value
algo := &algorithms.NodeSimilarity{
    TopK: 5000, // WN4007: may cause performance issues
}

// Recommended
algo := &algorithms.NodeSimilarity{
    TopK: 100,
}
```

---

## ML Pipeline Rules

### WN4031: Test Fraction Range

**Severity:** Error

The `testFraction` in SplitConfig must be less than 1.0.

```go
// Error: invalid test fraction
pipeline := &pipelines.NodeClassificationPipeline{
    SplitConfig: pipelines.SplitConfig{
        TestFraction: 1.0, // WN4031: must be < 1.0
    },
}

// Valid
pipeline := &pipelines.NodeClassificationPipeline{
    SplitConfig: pipelines.SplitConfig{
        TestFraction: 0.2,
    },
}
```

---

### WN4032: Model Candidates Required

**Severity:** Error

ML pipelines must have at least one model candidate.

```go
// Error: no models defined
pipeline := &pipelines.NodeClassificationPipeline{
    BasePipeline: pipelines.BasePipeline{
        Models: []pipelines.Model{}, // WN4032: at least one model required
    },
}

// Valid
pipeline := &pipelines.NodeClassificationPipeline{
    BasePipeline: pipelines.BasePipeline{
        Models: []pipelines.Model{
            &pipelines.LogisticRegression{},
        },
    },
}
```

---

## GraphRAG Rules

### WN4040: Entity Types Required

**Severity:** Error

KG pipelines must define at least one entity type.

```go
// Error: no entity types
pipeline := &kg.SimpleKGPipeline{
    EntityTypes: []kg.EntityType{}, // WN4040: at least one required
}

// Valid
pipeline := &kg.SimpleKGPipeline{
    EntityTypes: []kg.EntityType{
        {Name: "Person"},
    },
}
```

---

### WN4043: Entity Resolver Threshold

**Severity:** Warning

Entity resolver threshold should be >= 0.8 to avoid false positive matches.

```go
// Warning: low threshold may cause incorrect merges
pipeline := &kg.SimpleKGPipeline{
    EntityResolver: &kg.FuzzyMatchResolver{
        Threshold: 0.5, // WN4043: consider >= 0.8
    },
}

// Recommended
pipeline := &kg.SimpleKGPipeline{
    EntityResolver: &kg.FuzzyMatchResolver{
        Threshold: 0.85,
    },
}
```

---

## Schema Rules

### WN4052: Node Label Case

**Severity:** Warning

Node labels should use PascalCase.

```go
// Warning: should be PascalCase
node := &schema.NodeType{
    Label: "person", // WN4052: use "Person"
}

// Valid
node := &schema.NodeType{
    Label: "Person",
}
```

**Convention:** Neo4j node labels conventionally use PascalCase (e.g., `Person`, `MovieRating`, `UserAccount`).

---

### WN4053: Relationship Type Case

**Severity:** Warning

Relationship types should use SCREAMING_SNAKE_CASE.

```go
// Warning: should be SCREAMING_SNAKE_CASE
rel := &schema.RelationshipType{
    Label: "worksFor", // WN4053: use "WORKS_FOR"
}

// Valid
rel := &schema.RelationshipType{
    Label: "WORKS_FOR",
}
```

**Convention:** Neo4j relationship types conventionally use SCREAMING_SNAKE_CASE (e.g., `WORKS_FOR`, `HAS_ADDRESS`, `ACTED_IN`).

---

## Suppressing Rules

### Inline Suppression

Add a comment to suppress a specific rule:

```go
//nolint:WN4006 - dimension matches external model
algo := &algorithms.FastRP{
    EmbeddingDimension: 384,
}
```

### Configuration File

Suppress rules globally in `.wetwire.yaml`:

```yaml
lint:
  ignore:
    - WN4006  # Embedding dimension warnings
    - WN4007  # High topK warnings
```

### Command Line

Skip specific rules when running lint:

```bash
neo4j lint ./schemas/ --skip WN4006,WN4007
```

---

## Severity Levels

| Level | Description |
|-------|-------------|
| **Error** | Configuration will not work correctly; must be fixed |
| **Warning** | May cause issues or suboptimal behavior; should be reviewed |
| **Info** | Best practice suggestions; optional to address |

---

## Custom Rules

wetwire-neo4j-go supports custom lint rules through the plugin system. See [wetwire-core-go documentation](https://github.com/lex00/wetwire-core-go) for details on creating custom lint rules.
