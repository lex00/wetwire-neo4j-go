Generate Neo4j files for academic paper knowledge graph:

**schema.cypher:**
- Document: id (UNIQUE), title, content, embedding, year
- Person: name (UNIQUE), affiliation, orcid
- Concept: name (UNIQUE), definition, field
- Institution: name (UNIQUE), country
- Vector index: document_embedding (384d, cosine)
- Fulltext index: document_content on [content, title]

**queries.cypher:**
- Find papers by author (parameterized)
- Find citation network (variable depth)
- Find co-authors with collaboration count
- Semantic search using vector index

**algorithms.json:**
- PageRank on CITES network (dampingFactor=0.85, maxIterations=20)
- Louvain on co-author network (maxLevels=10, tolerance=0.0001)
- Graph projections for both algorithms
