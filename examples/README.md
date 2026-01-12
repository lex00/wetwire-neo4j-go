# Examples

This directory contains reference examples demonstrating wetwire-neo4j-go features.

## Contents

### schema.go
Node and relationship type definitions with constraints and indexes.
- `Person` - Node type with properties, unique constraint, and age index
- `Company` - Node type with fulltext index
- `WorksFor` - Relationship type with source/target constraints

### algorithms.go
GDS algorithm configurations for graph analytics.
- Centrality: PageRank, ArticleRank, Betweenness, Closeness, Degree
- Community: Louvain, Leiden, LabelPropagation, WCC, TriangleCount, KCore
- Similarity: NodeSimilarity, KNN
- Embeddings: FastRP, Node2Vec, GraphSAGE, HashGNN
- Path Finding: Dijkstra, AStar, BFS, DFS

### pipelines.go
GDS ML pipeline configurations for machine learning on graphs.
- Node Classification pipeline with feature steps and models
- Link Prediction pipeline for relationship prediction
- Node Regression pipeline for numeric property prediction

### projections.go
GDS graph projection configurations.
- Native projection with node labels and relationship types
- Cypher projection with custom queries
- DataFrame projection for Aura Analytics

### retrievers.go
GraphRAG retriever configurations for RAG applications.
- Vector retriever for similarity search
- Hybrid retriever combining vector and fulltext
- Text2Cypher retriever for natural language queries

### kg.go
Knowledge graph construction pipeline configurations.
- Simple KG pipeline with entity and relationship extraction
- Custom KG pipeline with user-defined prompts

## Running Examples

Tests validate all examples compile correctly:

```bash
go test ./examples/... -v
```

## Attribution

Examples are based on:
- [Neo4j GDS Documentation](https://neo4j.com/docs/graph-data-science/current/)
- [neo4j-graphrag-python](https://github.com/neo4j/neo4j-graphrag-python)
