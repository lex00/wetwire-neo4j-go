Dataset: academic papers. Create these files:

**expected/schema.go:**
- Document: id, title, content, embedding(LIST_FLOAT), year. Unique constraint on id. Vector index (384d, cosine) on embedding. Fulltext on content.
- Person: name, affiliation, orcid. Unique on name.
- Concept: name, definition, field. Unique on name.
- Institution: name, country. Unique on name.
- AUTHORED: Person->Document, MANY_TO_MANY, properties: order(INTEGER), corresponding(BOOLEAN)
- CITES: Document->Document, MANY_TO_MANY
- STUDIES: Document->Concept, MANY_TO_MANY
- AFFILIATED_WITH: Person->Institution, MANY_TO_ONE

**expected/extraction.go:**
- SimpleKGPipeline: "research-paper-kg", gpt-4, text-embedding-3-small(384d)
- Entity types: Person(name, affiliation, orcid), Document(title, year, doi, abstract), Concept(name, definition, field), Institution(name, country)
- Relation types: AUTHORED(Person->Document), CITES(Document->Document), STUDIES(Document->Concept), AFFILIATED_WITH(Person->Institution)
- FixedSizeSplitter(500, 50), FuzzyMatchResolver(0.85, "name")

**expected/embeddings.go:**
- EmbedderConfig for text-embedding-3-small, 384 dimensions

**expected/retrievers.go:**
- HybridRetriever: document_embedding_idx, document_content_fulltext, topk=10, weights(0.7, 0.3)

**expected/algorithms.go:**
- PageRank: citation-network, dampingFactor=0.85, maxIter=20
- Louvain: coauthor-network, maxLevels=10, tolerance=0.0001
