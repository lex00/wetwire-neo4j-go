# Research Paper Knowledge Graph

This directory contains Neo4j schema, queries, algorithms, and pipeline configurations for building a knowledge graph from academic research papers.

## Files

- **schema.cypher** - Database schema with constraints, indexes, and vector indexes
- **queries.cypher** - Common graph queries for analysis and search
- **algorithms.json** - Graph Data Science (GDS) algorithm configurations
- **pipeline.json** - Complete ML pipeline for entity extraction and relationship detection

## Quick Start

### 1. Create the Schema

Run the schema file to set up constraints and indexes:

```bash
cat schema.cypher | cypher-shell -u neo4j -p your-password
```

Or load it through Neo4j Browser:
```cypher
:source schema.cypher
```

### 2. Load Your Data

The schema expects four main node types:

**Document (Research Paper)**
```cypher
CREATE (d:Document {
  id: 'paper-001',
  title: 'Deep Learning for Natural Language Processing',
  abstract: 'This paper explores...',
  content: 'Full paper text...',
  year: 2023,
  venue: 'AAAI',
  citationCount: 0,
  embedding: [0.123, 0.456, ...] // 384-dimensional vector
})
```

**Person (Author)**
```cypher
CREATE (p:Person {
  id: 'author-001',
  name: 'Jane Smith',
  email: 'jane@university.edu',
  orcid: '0000-0001-2345-6789'
})
```

**Concept (Topic/Keyword)**
```cypher
CREATE (c:Concept {
  id: 'concept-nlp',
  name: 'Natural Language Processing',
  description: 'Field of AI focused on language understanding',
  category: 'AI',
  embedding: [0.789, 0.012, ...] // 384-dimensional vector
})
```

**Institution**
```cypher
CREATE (i:Institution {
  id: 'inst-001',
  name: 'Stanford University',
  country: 'USA',
  type: 'University'
})
```

### 3. Create Relationships

```cypher
// Author wrote paper
MATCH (p:Person {id: 'author-001'}), (d:Document {id: 'paper-001'})
CREATE (p)-[:AUTHORED {position: 1, isCorresponding: true}]->(d)

// Paper cites another paper
MATCH (d1:Document {id: 'paper-001'}), (d2:Document {id: 'paper-002'})
CREATE (d1)-[:CITES {context: 'methodology', section: 'introduction'}]->(d2)

// Paper studies concept
MATCH (d:Document {id: 'paper-001'}), (c:Concept {id: 'concept-nlp'})
CREATE (d)-[:STUDIES {relevance: 0.95, frequency: 15}]->(c)

// Author affiliated with institution
MATCH (p:Person {id: 'author-001'}), (i:Institution {id: 'inst-001'})
CREATE (p)-[:AFFILIATED_WITH {position: 'Professor', startDate: date('2020-01-01')}]->(i)
```

### 4. Run Sample Queries

Find papers by author:
```cypher
MATCH (p:Person)-[:AUTHORED]->(d:Document)
WHERE p.name = 'Jane Smith'
RETURN d.title, d.year, d.citationCount
ORDER BY d.year DESC;
```

Semantic search (requires embeddings):
```cypher
// Generate query embedding using your embedding model, then:
CALL db.index.vector.queryNodes('document_embedding', 10, $queryEmbedding)
YIELD node as d, score
RETURN d.title, d.abstract, score
ORDER BY score DESC;
```

Find co-authors:
```cypher
MATCH (p1:Person)-[:AUTHORED]->(:Document)<-[:AUTHORED]-(p2:Person)
WHERE p1.name = 'Jane Smith' AND p1 <> p2
RETURN p2.name, count(*) as collaborations
ORDER BY collaborations DESC;
```

### 5. Run Graph Algorithms

The `algorithms.json` file contains 10 pre-configured graph algorithms. To run them using Neo4j GDS:

**Citation PageRank (Find Influential Papers)**
```cypher
// Create graph projection
CALL gds.graph.project(
  'citation-graph',
  'Document',
  'CITES'
)

// Run PageRank
CALL gds.pageRank.write('citation-graph', {
  writeProperty: 'pagerank',
  dampingFactor: 0.85,
  maxIterations: 20
})

// Query results
MATCH (d:Document)
RETURN d.title, d.pagerank
ORDER BY d.pagerank DESC
LIMIT 20
```

**Author Communities (Louvain)**
```cypher
// First create collaboration relationships
MATCH (p1:Person)-[:AUTHORED]->(:Document)<-[:AUTHORED]-(p2:Person)
WHERE id(p1) < id(p2)
MERGE (p1)-[r:COLLABORATED_WITH]-(p2)
ON CREATE SET r.count = 1
ON MATCH SET r.count = r.count + 1

// Project graph
CALL gds.graph.project(
  'author-collaboration',
  'Person',
  {COLLABORATED_WITH: {orientation: 'UNDIRECTED'}}
)

// Run Louvain
CALL gds.louvain.write('author-collaboration', {
  writeProperty: 'community',
  maxLevels: 10
})

// Find community members
MATCH (p:Person)
WHERE p.community = 1
RETURN p.name
```

## Key Features

### Semantic Search
- Vector embeddings (384 dimensions) for documents and concepts
- Cosine similarity search using `db.index.vector.queryNodes()`
- Hybrid search combining fulltext and vector search

### Citation Analysis
- Citation network queries (direct, chain, co-citation)
- PageRank for influence scoring
- Betweenness centrality for bridge papers

### Research Communities
- Louvain algorithm for community detection
- Triangle counting for collaboration cohesion
- Institution collaboration networks

### Topic Analysis
- Concept co-occurrence relationships
- Topic similarity using Node Similarity
- Trend analysis over time

## Pipeline

The `pipeline.json` file defines a complete ML pipeline with 7 steps:

1. **Entity Extraction** - Extract persons, institutions, and concepts from text
2. **Embedding Generation** - Create vector embeddings for semantic search
3. **Relationship Extraction** - Detect AUTHORED, CITES, STUDIES, AFFILIATED_WITH
4. **Co-occurrence Detection** - Create implicit relationships (COLLABORATED_WITH, etc.)
5. **Citation Enrichment** - Calculate citation metrics
6. **Graph Algorithms** - Run PageRank, Louvain, etc.
7. **Quality Validation** - Validate data completeness

## Query Categories

The `queries.cypher` file includes 40+ queries organized by category:

- **Author Queries** - Papers by author, co-authors, affiliations
- **Citation Network** - Citation chains, bibliographic coupling, co-citation
- **Topic/Concept** - Papers by topic, related concepts, trending topics
- **Semantic Search** - Vector search, hybrid search
- **Research Communities** - Collaboration networks, institutional partnerships
- **Temporal** - Publication trends, concept evolution
- **Influence** - Most cited papers, prolific authors, rising stars

## Graph Algorithms

10 pre-configured algorithms in `algorithms.json`:

1. **citation-pagerank** - Influential papers
2. **author-communities** - Research communities (Louvain)
3. **concept-similarity** - Similar topics (Node Similarity)
4. **citation-betweenness** - Bridge papers (Betweenness Centrality)
5. **author-degree-centrality** - Most connected authors
6. **institution-collaboration-strength** - Institutional partnerships
7. **weakly-connected-components** - Disconnected subgraphs
8. **label-propagation-topics** - Topic clustering
9. **triangle-count-collaboration** - Collaboration triangles
10. **local-clustering-coefficient** - Author clustering measure

## Best Practices

1. **Always generate embeddings** for semantic search capabilities
2. **Run PageRank** to identify influential papers
3. **Create derived relationships** (COLLABORATED_WITH, CO_OCCURS) for richer analysis
4. **Use parameterized queries** to prevent injection and improve caching
5. **Monitor query performance** using `PROFILE` and `EXPLAIN`
6. **Regularly update citation counts** as new papers are added

## Example Use Cases

### Find Research Gaps
```cypher
// Find concepts with few recent papers
MATCH (c:Concept)<-[:STUDIES]-(d:Document)
WHERE d.year >= 2020
WITH c, count(d) as recentPapers
WHERE recentPapers < 5
RETURN c.name, recentPapers
ORDER BY recentPapers ASC
```

### Track Research Impact
```cypher
// Author's h-index approximation
MATCH (p:Person)-[:AUTHORED]->(d:Document)
WHERE p.name = $authorName
WITH d.citationCount as citations
ORDER BY citations DESC
WITH collect(citations) as citationList
UNWIND range(0, size(citationList)-1) as idx
WITH idx+1 as paperNum, citationList[idx] as citations
WHERE citations >= paperNum
RETURN max(paperNum) as hIndex
```

### Identify Emerging Topics
```cypher
// Concepts with accelerating publication rate
MATCH (c:Concept)<-[:STUDIES]-(d:Document)
WHERE d.year >= 2020
WITH c, d.year as year, count(d) as papers
ORDER BY c, year
WITH c, collect(papers) as yearlyPapers
WHERE size(yearlyPapers) >= 3
WITH c, yearlyPapers[-1] > yearlyPapers[-2] * 1.5 as isAccelerating
WHERE isAccelerating
RETURN c.name
```

## Resources

- [Neo4j Cypher Manual](https://neo4j.com/docs/cypher-manual/current/)
- [Neo4j Graph Data Science Library](https://neo4j.com/docs/graph-data-science/current/)
- [Vector Search in Neo4j](https://neo4j.com/docs/cypher-manual/current/indexes-for-vector-search/)
- [GraphRAG Documentation](https://neo4j.com/labs/genai-ecosystem/graphrag/)
