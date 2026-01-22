# Academic Papers Knowledge Graph System

This directory contains the complete Neo4j configuration for an academic papers knowledge graph system with semantic search, entity extraction, and graph analytics.

## Files Overview

### 1. `schema.cypher`
Defines the graph schema including:
- **Node constraints** for unique identifiers (Document, Person, Concept, Institution)
- **Vector indexes** for semantic search (384-dimensional embeddings)
- **Fulltext indexes** for hybrid search on content
- **Property indexes** for efficient filtering and querying

### 2. `queries.cypher`
Common Cypher queries organized by category:
- **Basic retrieval**: Find papers by author, institution, concept
- **Citation network**: Direct citations, citation chains, common citations
- **Collaboration**: Co-authors, collaboration networks
- **Concept analysis**: Related concepts, trending topics, author expertise
- **Hybrid search**: Vector similarity, fulltext, and combined searches
- **Analytics**: Most cited papers, prolific authors, institution productivity
- **Recommendations**: Paper/author recommendations based on research patterns

### 3. `pipeline.json`
Knowledge graph extraction pipeline configuration:
- **Entity Extractor**: OpenAI GPT-4 for extracting entities and relationships
- **Text Splitter**: 500 character chunks with 50 character overlap
- **Text Embedder**: OpenAI text-embedding-3-small (384 dimensions)
- **Entity Resolver**: Fuzzy matching with 0.85 threshold for deduplication

### 4. `retriever.json`
Hybrid search retriever configuration:
- **Vector search**: 60% weight on semantic similarity
- **Fulltext search**: 40% weight on keyword matching
- **Context enrichment**: Include citations, authors, and concepts
- **Filtering**: Support for year, type, and concept filters
- Returns top 10 results by default

### 5. `algorithms.json`
Graph Data Science (GDS) algorithm configurations:
- **PageRank**: Compute paper influence from citation network
- **Louvain**: Detect research communities
- **Node Similarity**: Find similar authors by collaboration patterns
- **Betweenness Centrality**: Identify bridging concepts
- **Link Prediction**: Predict potential future citations

## Entity Schema

### Nodes

#### Document
- **Properties**: id, title, abstract, content, year, type, doi, embedding
- **Description**: Academic papers and publications

#### Person
- **Properties**: id, name, email, orcid
- **Description**: Authors and researchers

#### Concept
- **Properties**: id, name, description, category
- **Description**: Research topics, keywords, methodologies

#### Institution
- **Properties**: id, name, location, type
- **Description**: Universities, labs, organizations

### Relationships

#### AUTHORED
- **Source**: Person → Document
- **Properties**: position (author order)
- **Description**: Authorship relationship

#### CITES
- **Source**: Document → Document
- **Properties**: citedYear
- **Description**: Citation relationship

#### STUDIES
- **Source**: Document → Concept
- **Properties**: relevance
- **Description**: Topic/concept relationship

#### AFFILIATED_WITH
- **Source**: Person → Institution
- **Properties**: startDate, endDate, role
- **Description**: Institutional affiliation

## Usage

### 1. Create Schema
```bash
cat schema.cypher | cypher-shell -u neo4j -p password
```

### 2. Run Knowledge Graph Pipeline
```python
from neo4j_graphrag import KnowledgeGraphPipeline
import json

with open('pipeline.json') as f:
    config = json.load(f)

pipeline = KnowledgeGraphPipeline.from_config(config)
pipeline.run(documents)
```

### 3. Execute Hybrid Search
```python
from neo4j_graphrag import HybridRetriever
import json

with open('retriever.json') as f:
    config = json.load(f)

retriever = HybridRetriever.from_config(config)
results = retriever.search(query="machine learning in healthcare")
```

### 4. Run Graph Algorithms
```cypher
// Create graph projection
CALL gds.graph.project(
  'citation-network',
  'Document',
  'CITES'
);

// Run PageRank
CALL gds.pageRank.write('citation-network', {
  dampingFactor: 0.85,
  maxIterations: 20,
  writeProperty: 'influenceScore'
});

// Run Louvain
CALL gds.graph.project(
  'research-network',
  ['Document', 'Person', 'Concept'],
  {
    AUTHORED: {orientation: 'UNDIRECTED'},
    CITES: {orientation: 'UNDIRECTED'},
    STUDIES: {orientation: 'UNDIRECTED'}
  }
);

CALL gds.louvain.write('research-network', {
  writeProperty: 'communityId'
});
```

### 5. Query Examples

**Find influential papers in a research area:**
```cypher
MATCH (d:Document)-[:STUDIES]->(c:Concept)
WHERE c.name = 'Deep Learning'
RETURN d.title, d.year, d.influenceScore
ORDER BY d.influenceScore DESC
LIMIT 10;
```

**Find research communities:**
```cypher
MATCH (d:Document)
WHERE d.communityId IS NOT NULL
WITH d.communityId as community, collect(d) as papers
RETURN community, size(papers) as paperCount,
       [p IN papers | p.title][..5] as samplePapers
ORDER BY paperCount DESC;
```

**Hybrid search with filters:**
```cypher
CALL {
  CALL db.index.vector.queryNodes('document_embedding', 20, $queryEmbedding)
  YIELD node, score
  RETURN node, score * 0.6 as finalScore
  UNION
  CALL db.index.fulltext.queryNodes('document_content', $searchQuery)
  YIELD node, score
  RETURN node, score * 0.4 as finalScore
}
WITH node, sum(finalScore) as combinedScore
WHERE node.year >= 2020
RETURN node.title, node.year, combinedScore
ORDER BY combinedScore DESC
LIMIT 10;
```

## Pipeline Execution Order

1. **Text Splitting**: Break documents into chunks
2. **Entity Extraction**: Extract entities and relationships using GPT-4
3. **Entity Resolution**: Deduplicate entities using fuzzy matching
4. **Text Embedding**: Generate embeddings for semantic search

## Retriever Configuration

The hybrid retriever combines:
- **Vector search** (60%): Semantic similarity using cosine distance
- **Fulltext search** (40%): Keyword matching with fuzzy support

Results include enriched context:
- Up to 10 authors
- Up to 5 related concepts
- Up to 5 citations

## Graph Algorithms

### Paper Influence (PageRank)
- **Input**: Citation network (Document→CITES→Document)
- **Output**: `influenceScore` property on Document nodes
- **Parameters**: Damping factor 0.85, max 20 iterations

### Research Communities (Louvain)
- **Input**: Multi-node network (Document, Person, Concept with all relationships)
- **Output**: `communityId` property on all nodes
- **Purpose**: Identify research clusters and collaboration groups

### Author Similarity (Node Similarity)
- **Input**: Co-authorship network
- **Output**: SIMILAR_TO relationships between Person nodes
- **Parameters**: Top 10 similar authors per person

### Concept Centrality (Betweenness)
- **Input**: Concept co-occurrence network
- **Output**: `betweennessCentrality` property on Concept nodes
- **Purpose**: Find bridging topics between research areas

### Citation Prediction (Link Prediction)
- **Input**: Citation network
- **Output**: Stream of predicted citation pairs
- **Purpose**: Recommend papers that might be relevant to cite

## Integration Notes

This configuration is designed to work with:
- **Neo4j 5.x** with Graph Data Science (GDS) plugin
- **neo4j-graphrag-python** library for pipeline execution
- **OpenAI API** for embeddings and entity extraction

## Next Steps

1. Load schema into Neo4j instance
2. Configure API keys for OpenAI integration
3. Run pipeline on academic paper corpus
4. Execute graph algorithms for analytics
5. Use retriever for semantic search applications
