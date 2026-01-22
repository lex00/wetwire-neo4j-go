// ============================================================================
// Research Paper Knowledge Graph Schema
// ============================================================================
// This schema defines nodes, relationships, constraints, and indexes for
// managing academic papers, authors, concepts, and institutions.

// ----------------------------------------------------------------------------
// CONSTRAINTS - Ensure data integrity with unique identifiers
// ----------------------------------------------------------------------------

// Document constraints
CREATE CONSTRAINT document_id IF NOT EXISTS
FOR (d:Document) REQUIRE d.id IS UNIQUE;

// Person (Author) constraints
CREATE CONSTRAINT person_id IF NOT EXISTS
FOR (p:Person) REQUIRE p.id IS UNIQUE;

// Concept (Topic/Keyword) constraints
CREATE CONSTRAINT concept_id IF NOT EXISTS
FOR (c:Concept) REQUIRE c.id IS UNIQUE;

// Institution constraints
CREATE CONSTRAINT institution_id IF NOT EXISTS
FOR (i:Institution) REQUIRE i.id IS UNIQUE;

// ----------------------------------------------------------------------------
// PROPERTY INDEXES - Optimize query performance
// ----------------------------------------------------------------------------

// Index on document year for temporal queries
CREATE INDEX document_year IF NOT EXISTS
FOR (d:Document) ON (d.year);

// Index on document publication venue
CREATE INDEX document_venue IF NOT EXISTS
FOR (d:Document) ON (d.venue);

// Index on person name for author lookups
CREATE INDEX person_name IF NOT EXISTS
FOR (p:Person) ON (p.name);

// Index on concept name for topic searches
CREATE INDEX concept_name IF NOT EXISTS
FOR (c:Concept) ON (c.name);

// Index on institution name
CREATE INDEX institution_name IF NOT EXISTS
FOR (i:Institution) ON (i.name);

// ----------------------------------------------------------------------------
// VECTOR INDEXES - Enable semantic search with embeddings
// ----------------------------------------------------------------------------

// Vector index for document content embeddings (semantic search)
CREATE VECTOR INDEX document_embedding IF NOT EXISTS
FOR (d:Document) ON (d.embedding)
OPTIONS {indexConfig: {
  `vector.dimensions`: 384,
  `vector.similarity_function`: 'cosine'
}};

// Vector index for concept embeddings (topic similarity)
CREATE VECTOR INDEX concept_embedding IF NOT EXISTS
FOR (c:Concept) ON (c.embedding)
OPTIONS {indexConfig: {
  `vector.dimensions`: 384,
  `vector.similarity_function`: 'cosine'
}};

// ----------------------------------------------------------------------------
// FULLTEXT INDEXES - Enable text search
// ----------------------------------------------------------------------------

// Fulltext search on document content and metadata
CREATE FULLTEXT INDEX document_fulltext IF NOT EXISTS
FOR (d:Document) ON EACH [d.title, d.abstract, d.content];

// Fulltext search on concepts
CREATE FULLTEXT INDEX concept_fulltext IF NOT EXISTS
FOR (c:Concept) ON EACH [c.name, c.description];

// Fulltext search on person names
CREATE FULLTEXT INDEX person_fulltext IF NOT EXISTS
FOR (p:Person) ON EACH [p.name, p.orcid];

// ----------------------------------------------------------------------------
// RELATIONSHIP INDEXES - Optimize relationship queries
// ----------------------------------------------------------------------------

// Index on AUTHORED relationship for author-paper lookups
CREATE INDEX authored_index IF NOT EXISTS
FOR ()-[r:AUTHORED]-() ON (r.position);

// Index on CITES relationship for citation analysis
CREATE INDEX cites_index IF NOT EXISTS
FOR ()-[r:CITES]-() ON (r.context);

// Index on STUDIES relationship for concept relevance
CREATE INDEX studies_index IF NOT EXISTS
FOR ()-[r:STUDIES]-() ON (r.relevance);
