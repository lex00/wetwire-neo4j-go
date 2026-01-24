---
title: "Wetwire Neo4j"
---

[![Go Reference](https://pkg.go.dev/badge/github.com/lex00/wetwire-neo4j-go.svg)](https://pkg.go.dev/github.com/lex00/wetwire-neo4j-go)
[![CI](https://github.com/lex00/wetwire-neo4j-go/actions/workflows/ci.yml/badge.svg)](https://github.com/lex00/wetwire-neo4j-go/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Generate Neo4j Cypher queries and graph pipelines from Go structs with AI-assisted design.

## Philosophy

Wetwire uses typed constraints to reduce the model capability required for accurate code generation.

**Core hypothesis:** Typed input + smaller model ≈ Semantic input + larger model

The type system and lint rules act as a force multiplier — cheaper models produce quality output when guided by schema-generated types and iterative lint feedback.

## Documentation

| Document | Description |
|----------|-------------|
| [CLI Reference]({{< relref "/cli" >}}) | Command-line interface |
| [Quick Start]({{< relref "/quick-start" >}}) | Get started in 5 minutes |
| [Examples]({{< relref "/examples" >}}) | Sample graph projects |
| [FAQ]({{< relref "/faq" >}}) | Frequently asked questions |

## Installation

```bash
go install github.com/lex00/wetwire-neo4j-go@latest
```

## Quick Example

```go
var FindUsers = cypher.Query{
    Match:  "(u:User)-[:FOLLOWS]->(f:User)",
    Where:  "u.active = true",
    Return: "u.name, count(f) as followers",
}
```
