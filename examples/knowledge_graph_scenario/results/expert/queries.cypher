// ============================================================
// Common Queries for Academic Paper Knowledge Graph
// ============================================================

// ------------------------------------------------------------
// Query 1: Find papers by author
// ------------------------------------------------------------
// Parameters: $authorName (string)
// Returns: Paper titles and publication years for a specific author
// Usage: CALL { ... } IN TRANSACTIONS

MATCH (p:Person)-[:AUTHORED]->(d:Document)
WHERE p.name = $authorName
RETURN d.title AS title,
       d.year AS year,
       d.id AS documentId
ORDER BY d.year DESC;

// ------------------------------------------------------------
// Query 2: Find citation network (variable depth)
// ------------------------------------------------------------
// Parameters: $documentId (string), $maxDepth (integer, default: 3)
// Returns: Citation paths from a specific document
// Note: Use LIMIT to prevent excessive results

MATCH path = (d1:Document)-[:CITES*1..3]->(d2:Document)
WHERE d1.id = $documentId
RETURN path,
       length(path) AS citationDepth,
       d2.title AS citedPaper,
       d2.year AS citedYear
ORDER BY citationDepth, d2.year DESC;

// ------------------------------------------------------------
// Query 3: Find co-authors with collaboration count
// ------------------------------------------------------------
// Parameters: $personId (string) OR $personName (string)
// Returns: Co-authors and number of collaborative papers

// Version 1: Using person name
MATCH (p1:Person)-[:AUTHORED]->(d:Document)<-[:AUTHORED]-(p2:Person)
WHERE p1.name = $personName AND p1 <> p2
RETURN DISTINCT p2.name AS coAuthor,
                p2.affiliation AS affiliation,
                p2.orcid AS orcid,
                count(d) AS collaborations
ORDER BY collaborations DESC, p2.name;

// Version 2: Including shared papers
MATCH (p1:Person)-[:AUTHORED]->(d:Document)<-[:AUTHORED]-(p2:Person)
WHERE p1.name = $personName AND p1 <> p2
WITH p2, count(d) AS collaborations, collect(d.title) AS sharedPapers
RETURN p2.name AS coAuthor,
       p2.affiliation AS affiliation,
       collaborations,
       sharedPapers
ORDER BY collaborations DESC;

// ------------------------------------------------------------
// Query 4: Semantic search using vector index
// ------------------------------------------------------------
// Parameters: $queryEmbedding (list of floats), $topK (integer, default: 10)
// Returns: Most semantically similar documents

CALL db.index.vector.queryNodes('document_embedding', $topK, $queryEmbedding)
YIELD node AS document, score
RETURN document.id AS documentId,
       document.title AS title,
       document.year AS year,
       document.content AS content,
       score AS similarityScore
ORDER BY score DESC;

// ------------------------------------------------------------
// Query 5: Find papers studying specific concepts
// ------------------------------------------------------------
// Parameters: $conceptName (string)
// Returns: Papers related to a specific research concept

MATCH (d:Document)-[:STUDIES]->(c:Concept)
WHERE c.name = $conceptName
RETURN d.title AS title,
       d.year AS year,
       c.definition AS conceptDefinition,
       c.field AS field
ORDER BY d.year DESC;

// ------------------------------------------------------------
// Query 6: Find author's institutional affiliations
// ------------------------------------------------------------
// Parameters: $authorName (string)
// Returns: Institutions associated with an author

MATCH (p:Person)-[:AFFILIATED_WITH]->(i:Institution)
WHERE p.name = $authorName
RETURN p.name AS author,
       i.name AS institution,
       i.country AS country;

// ------------------------------------------------------------
// Query 7: Citation impact analysis
// ------------------------------------------------------------
// Parameters: $documentId (string)
// Returns: Incoming and outgoing citation counts

MATCH (d:Document)
WHERE d.id = $documentId
OPTIONAL MATCH (d)-[:CITES]->(cited:Document)
OPTIONAL MATCH (citing:Document)-[:CITES]->(d)
RETURN d.title AS paper,
       d.year AS year,
       count(DISTINCT cited) AS citationsGiven,
       count(DISTINCT citing) AS citationsReceived;

// ------------------------------------------------------------
// Query 8: Research concept network
// ------------------------------------------------------------
// Parameters: $conceptName (string), $depth (integer, default: 2)
// Returns: Related concepts through shared papers

MATCH path = (c1:Concept)<-[:STUDIES]-(d:Document)-[:STUDIES]->(c2:Concept)
WHERE c1.name = $conceptName AND c1 <> c2
RETURN DISTINCT c2.name AS relatedConcept,
                c2.field AS field,
                count(DISTINCT d) AS sharedPapers
ORDER BY sharedPapers DESC
LIMIT 20;

// ------------------------------------------------------------
// Query 9: Fulltext search on document content
// ------------------------------------------------------------
// Parameters: $searchQuery (string)
// Returns: Documents matching text search

CALL db.index.fulltext.queryNodes('document_content', $searchQuery)
YIELD node AS document, score
RETURN document.id AS documentId,
       document.title AS title,
       document.year AS year,
       score AS relevanceScore
ORDER BY score DESC
LIMIT 10;

// ------------------------------------------------------------
// Query 10: Author productivity by year
// ------------------------------------------------------------
// Parameters: $authorName (string)
// Returns: Publication count per year for an author

MATCH (p:Person)-[:AUTHORED]->(d:Document)
WHERE p.name = $authorName
RETURN d.year AS year,
       count(d) AS publications,
       collect(d.title) AS papers
ORDER BY year DESC;
