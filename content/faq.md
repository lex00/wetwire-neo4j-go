---
title: "FAQ"
---

# Frequently Asked Questions

## General

### What is wetwire-neo4j-go?

wetwire-neo4j-go is a synthesis library for Neo4j that lets you define Neo4j configurations in native Go and compile them to Cypher statements and JSON configurations. It supports schema definitions, GDS algorithms, ML pipelines, graph projections, and GraphRAG configurations.

### How is this different from just writing Cypher directly?

wetwire-neo4j-go provides:
- **Type safety**: Catch errors at compile time, not runtime
- **Validation**: Lint rules check for common mistakes before deployment
- **Discoverability**: AST-based discovery finds all definitions automatically
- **Reusability**: Define once, use across environments
- **IDE support**: Full Go IDE support with autocompletion

### Do I need a running Neo4j instance to use this?

No. wetwire-neo4j-go generates Cypher and JSON output without connecting to Neo4j. However, the `validate` command can optionally check your configurations against a live Neo4j instance.

---

## Schema Definitions

### What constraint types are supported?

- `schema.Unique` - Uniqueness constraint on properties
- `schema.Exists` - Property existence constraint
- `schema.NodeKey` - Node key (unique + exists combined)
- `schema.RelKey` - Relationship key constraint

### What index types are supported?

- `schema.RangeIndex` - Standard B-tree index
- `schema.TextIndex` - Text index for string matching
- `schema.FullTextIndex` - Full-text search index
- `schema.PointIndex` - Spatial index for Point properties
- `schema.VectorIndex` - Vector similarity index

### What property types are available?

- `schema.TypeString`, `schema.TypeInteger`, `schema.TypeFloat`
- `schema.TypeBoolean`, `schema.TypeDate`, `schema.TypeDateTime`
- `schema.TypePoint`, `schema.TypeDuration`
- List types: `schema.TypeListString`, `schema.TypeListInteger`, etc.

---

## GDS Algorithms

### What algorithms are supported?

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

### What execution modes are available?

- `algorithms.Stream` - Returns results as a stream
- `algorithms.Stats` - Returns statistics only
- `algorithms.Mutate` - Writes results to in-memory graph
- `algorithms.Write` - Writes results to database

### Why does lint warn about embedding dimensions?

WN4006 warns when embedding dimensions (FastRP, Node2Vec) are not powers of 2. Powers of 2 enable SIMD optimizations and efficient memory alignment, improving performance.

---

## ML Pipelines

### What pipeline types are available?

- `NodeClassificationPipeline` - Categorical label prediction
- `LinkPredictionPipeline` - Relationship prediction
- `NodeRegressionPipeline` - Numeric property prediction

### What models can I use?

- `LogisticRegression` - For classification
- `RandomForest` - For classification
- `MLP` (Multi-Layer Perceptron) - For classification
- `LinearRegression` - For regression

### How do I add features to a pipeline?

Use feature steps:
- `FastRPStep` - FastRP embeddings
- `PageRankStep` - PageRank scores
- `DegreeStep` - Node degrees
- `Node2VecStep` - Node2Vec embeddings
- `ScalerStep` - Feature scaling

---

## GraphRAG

### What retriever types are available?

- `VectorRetriever` - Vector similarity search
- `VectorCypherRetriever` - Vector search with custom traversal
- `HybridRetriever` - Combined vector + fulltext search
- `HybridCypherRetriever` - Hybrid with custom traversal
- `Text2CypherRetriever` - LLM-generated Cypher queries

### What external integrations are supported?

- Weaviate, Pinecone, Qdrant - External vector databases

### What KG pipeline types are available?

- `SimpleKGPipeline` - Standard entity/relationship extraction
- `CustomKGPipeline` - Custom extraction with user prompts

---

## CLI

### How do I generate Cypher from my definitions?

```bash
wetwire-neo4j build ./schemas/
```

### How do I check for issues?

```bash
wetwire-neo4j lint ./schemas/
```

### How do I validate against a live Neo4j instance?

```bash
wetwire-neo4j validate --uri bolt://localhost:7687 --password secret
```

### How do I import existing schemas?

From a Cypher file:
```bash
wetwire-neo4j import --file schema.cypher -o schema.go
```

From a Neo4j database:
```bash
wetwire-neo4j import --uri bolt://localhost:7687 --password secret -o schema.go
```

---

## Troubleshooting

### Lint says "dampingFactor must be in [0, 1)"

The damping factor for PageRank/ArticleRank must be less than 1.0. Common values are 0.85 or 0.90.

### Lint says "maxIterations must be positive"

The maxIterations parameter must be greater than 0. Use 0 to indicate "use default".

### Validate can't connect to Neo4j

Check:
1. URI is correct (bolt://, neo4j://, neo4j+s://)
2. Username and password are correct
3. Neo4j is running and accessible
4. Firewall allows the connection

### Import generates empty output

Ensure the Cypher file contains CREATE CONSTRAINT or CREATE INDEX statements. The importer only extracts schema definitions, not data.

---

## Integration

### How does this integrate with wetwire-core-go?

wetwire-neo4j-go uses wetwire-core-go for:
- CLI infrastructure and command registration
- MCP server integration for Claude Code
- Common lint rule framework

### Can I use this with Claude Code?

Yes! wetwire-core-go provides MCP server integration. Claude Code can discover and work with your Neo4j configurations through the MCP protocol.
