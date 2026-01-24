---
title: "FAQ"
---

# Frequently Asked Questions

## General

<details>
<summary>What is wetwire-neo4j-go?</summary>

wetwire-neo4j-go is a synthesis library for Neo4j that lets you define Neo4j configurations in native Go and compile them to Cypher statements and JSON configurations. It supports schema definitions, GDS algorithms, ML pipelines, graph projections, and GraphRAG configurations.
</details>

<details>
<summary>How is this different from just writing Cypher directly?</summary>

wetwire-neo4j-go provides:
- **Type safety**: Catch errors at compile time, not runtime
- **Validation**: Lint rules check for common mistakes before deployment
- **Discoverability**: AST-based discovery finds all definitions automatically
- **Reusability**: Define once, use across environments
- **IDE support**: Full Go IDE support with autocompletion
</details>

<details>
<summary>Do I need a running Neo4j instance to use this?</summary>

No. wetwire-neo4j-go generates Cypher and JSON output without connecting to Neo4j. However, the `validate` command can optionally check your configurations against a live Neo4j instance.
</details>

---

## Schema Definitions

<details>
<summary>What constraint types are supported?</summary>

- `schema.Unique` - Uniqueness constraint on properties
- `schema.Exists` - Property existence constraint
- `schema.NodeKey` - Node key (unique + exists combined)
- `schema.RelKey` - Relationship key constraint
</details>

<details>
<summary>What index types are supported?</summary>

- `schema.RangeIndex` - Standard B-tree index
- `schema.TextIndex` - Text index for string matching
- `schema.FullTextIndex` - Full-text search index
- `schema.PointIndex` - Spatial index for Point properties
- `schema.VectorIndex` - Vector similarity index
</details>

<details>
<summary>What property types are available?</summary>

- `schema.TypeString`, `schema.TypeInteger`, `schema.TypeFloat`
- `schema.TypeBoolean`, `schema.TypeDate`, `schema.TypeDateTime`
- `schema.TypePoint`, `schema.TypeDuration`
- List types: `schema.TypeListString`, `schema.TypeListInteger`, etc.
</details>

---

## GDS Algorithms

<details>
<summary>What algorithms are supported?</summary>

**Centrality:**
- PageRank, ArticleRank, Betweenness, Degree, Closeness, Eigenvector

**Community Detection:**
- Louvain, Leiden, LabelPropagation, WCC, TriangleCount, KCore

**Similarity:**
- NodeSimilarity, KNN

**Embeddings:**
- FastRP, Node2Vec, GraphSAGE, HashGNN

**Path Finding:**
- Dijkstra, AStar, BFS, DFS
</details>

<details>
<summary>What execution modes are available?</summary>

- `algorithms.Stream` - Returns results as a stream
- `algorithms.Stats` - Returns statistics only
- `algorithms.Mutate` - Writes results to in-memory graph
- `algorithms.Write` - Writes results to database
</details>

<details>
<summary>Why does lint warn about embedding dimensions?</summary>

WN4006 warns when embedding dimensions (FastRP, Node2Vec) are not powers of 2. Powers of 2 enable SIMD optimizations and efficient memory alignment, improving performance.
</details>

---

## ML Pipelines

<details>
<summary>What pipeline types are available?</summary>

- `NodeClassificationPipeline` - Categorical label prediction
- `LinkPredictionPipeline` - Relationship prediction
- `NodeRegressionPipeline` - Numeric property prediction
</details>

<details>
<summary>What models can I use?</summary>

- `LogisticRegression` - For classification
- `RandomForest` - For classification
- `MLP` (Multi-Layer Perceptron) - For classification
- `LinearRegression` - For regression
</details>

<details>
<summary>How do I add features to a pipeline?</summary>

Use feature steps:
- `FastRPStep` - FastRP embeddings
- `PageRankStep` - PageRank scores
- `DegreeStep` - Node degrees
- `Node2VecStep` - Node2Vec embeddings
- `ScalerStep` - Feature scaling
</details>

---

## GraphRAG

<details>
<summary>What retriever types are available?</summary>

- `VectorRetriever` - Vector similarity search
- `VectorCypherRetriever` - Vector search with custom traversal
- `HybridRetriever` - Combined vector + fulltext search
- `HybridCypherRetriever` - Hybrid with custom traversal
- `Text2CypherRetriever` - LLM-generated Cypher queries
</details>

<details>
<summary>What external integrations are supported?</summary>

- Weaviate, Pinecone, Qdrant - External vector databases
</details>

<details>
<summary>What KG pipeline types are available?</summary>

- `SimpleKGPipeline` - Standard entity/relationship extraction
- `CustomKGPipeline` - Custom extraction with user prompts
</details>

---

## CLI

<details>
<summary>How do I generate Cypher from my definitions?</summary>

```bash
wetwire-neo4j build ./schemas/
```
</details>

<details>
<summary>How do I check for issues?</summary>

```bash
wetwire-neo4j lint ./schemas/
```
</details>

<details>
<summary>How do I validate against a live Neo4j instance?</summary>

```bash
wetwire-neo4j validate --uri bolt://localhost:7687 --password secret
```
</details>

<details>
<summary>How do I import existing schemas?</summary>

From a Cypher file:
```bash
wetwire-neo4j import --file schema.cypher -o schema.go
```

From a Neo4j database:
```bash
wetwire-neo4j import --uri bolt://localhost:7687 --password secret -o schema.go
```
</details>

---

## Platform and DevOps

<details>
<summary>How do I manage database connections securely?</summary>

Use environment variables for credentials:
```bash
export NEO4J_URI="bolt://localhost:7687"
export NEO4J_PASSWORD="your-password"
wetwire-neo4j validate --uri $NEO4J_URI --password $NEO4J_PASSWORD
```

For CI/CD pipelines, use your platform's secret management (GitHub Secrets, Vault, etc.) and inject credentials at runtime. Never commit credentials to version control.
</details>

<details>
<summary>Can I import existing Cypher queries?</summary>

Yes. Use the `import` command to convert existing Cypher DDL into Go definitions:

```bash
# From a file
wetwire-neo4j import --file existing-schema.cypher -o schema.go

# From a running database
wetwire-neo4j import --uri bolt://localhost:7687 --password secret -o schema.go
```

The importer extracts schema definitions (constraints, indexes) from Cypher. Data queries are not imported.
</details>

<details>
<summary>How does the linter help catch query errors?</summary>

The linter validates your definitions before deployment:

- **WN4001-WN4010**: Schema validation (missing labels, invalid constraints)
- **WN4011-WN4020**: Algorithm validation (invalid parameters, missing graph references)
- **WN4021-WN4030**: Pipeline validation (missing features, invalid model configs)
- **WN4031-WN4040**: Projection validation (missing node/relationship projections)

Run `wetwire-neo4j lint ./...` in CI to catch issues before they reach production.
</details>

<details>
<summary>What's the recommended project structure?</summary>

```
my-neo4j-project/
├── schema/
│   ├── nodes.go       # Node type definitions
│   ├── rels.go        # Relationship type definitions
│   └── indexes.go     # Index definitions
├── algorithms/
│   ├── centrality.go  # PageRank, Betweenness, etc.
│   └── community.go   # Louvain, WCC, etc.
├── pipelines/
│   └── ml.go          # ML pipeline definitions
├── projections/
│   └── graphs.go      # Graph projection definitions
└── .wetwire.yaml      # Configuration file
```

Group related definitions in separate files. Use the `build` command with `./...` to compile all definitions.
</details>

<details>
<summary>How do I handle graph schema migrations?</summary>

wetwire-neo4j generates idempotent Cypher DDL. For migrations:

1. **Add new definitions** - New constraints/indexes are created with `IF NOT EXISTS`
2. **Modify existing** - Update the Go definition, rebuild, and apply
3. **Remove** - Delete the Go definition; manually drop in Neo4j if needed

For complex migrations, use the `diff` command to compare versions:
```bash
wetwire-neo4j diff old-schema.cypher new-schema.cypher
```

This shows added, removed, and modified resources.
</details>

---

## Troubleshooting

<details>
<summary>Lint says "dampingFactor must be in [0, 1)"</summary>

The damping factor for PageRank/ArticleRank must be less than 1.0. Common values are 0.85 or 0.90.
</details>

<details>
<summary>Lint says "maxIterations must be positive"</summary>

The maxIterations parameter must be greater than 0. Use 0 to indicate "use default".
</details>

<details>
<summary>Validate can't connect to Neo4j</summary>

Check:
1. URI is correct (bolt://, neo4j://, neo4j+s://)
2. Username and password are correct
3. Neo4j is running and accessible
4. Firewall allows the connection
</details>

<details>
<summary>Import generates empty output</summary>

Ensure the Cypher file contains CREATE CONSTRAINT or CREATE INDEX statements. The importer only extracts schema definitions, not data.
</details>
