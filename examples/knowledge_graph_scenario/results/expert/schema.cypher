// ============================================================
// Neo4j Schema for Academic Paper Knowledge Graph
// ============================================================

// ------------------------------------------------------------
// Constraints - Unique Identifiers
// ------------------------------------------------------------

// Document constraints
CREATE CONSTRAINT document_id IF NOT EXISTS
FOR (d:Document) REQUIRE d.id IS UNIQUE;

// Person constraints
CREATE CONSTRAINT person_name IF NOT EXISTS
FOR (p:Person) REQUIRE p.name IS UNIQUE;

// Concept constraints
CREATE CONSTRAINT concept_name IF NOT EXISTS
FOR (c:Concept) REQUIRE c.name IS UNIQUE;

// Institution constraints
CREATE CONSTRAINT institution_name IF NOT EXISTS
FOR (i:Institution) REQUIRE i.name IS UNIQUE;

// ------------------------------------------------------------
// Indexes - Performance Optimization
// ------------------------------------------------------------

// Vector index for semantic search on document embeddings
CREATE VECTOR INDEX document_embedding IF NOT EXISTS
FOR (d:Document) ON (d.embedding)
OPTIONS {indexConfig: {
  `vector.dimensions`: 384,
  `vector.similarity_function`: 'cosine'
}};

// Fulltext index for text search on document content and title
CREATE FULLTEXT INDEX document_content IF NOT EXISTS
FOR (d:Document) ON EACH [d.content, d.title];

// Property indexes for common query patterns
CREATE INDEX document_year IF NOT EXISTS
FOR (d:Document) ON (d.year);

CREATE INDEX person_affiliation IF NOT EXISTS
FOR (p:Person) ON (p.affiliation);

CREATE INDEX person_orcid IF NOT EXISTS
FOR (p:Person) ON (p.orcid);

CREATE INDEX concept_field IF NOT EXISTS
FOR (c:Concept) ON (c.field);

CREATE INDEX institution_country IF NOT EXISTS
FOR (i:Institution) ON (i.country);

// ------------------------------------------------------------
// Schema Complete
// ------------------------------------------------------------
