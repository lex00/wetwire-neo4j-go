Build a knowledge graph system for academic papers with:

**Schema:**
- Node types: Document, Person, Concept, Institution
- Relationships: AUTHORED, CITES, STUDIES, AFFILIATED_WITH
- Vector indexes for semantic search (384d embeddings)
- Fulltext indexes on content

**KG Pipeline:**
- Entity extraction with OpenAI GPT-4
- Text embeddings with text-embedding-3-small
- Text splitting (500 chars, 50 overlap)
- Fuzzy entity resolution (0.85 threshold)

**Retriever:**
- Hybrid search (vector + fulltext)
- Return top 10 results

**Algorithms:**
- PageRank for paper influence
- Louvain for research communities
