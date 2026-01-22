// ============================================================================
// Research Paper Knowledge Graph - Common Queries
// ============================================================================
// This file contains frequently used Cypher queries for analyzing research
// papers, authors, citations, and research communities.

// ----------------------------------------------------------------------------
// AUTHOR QUERIES
// ----------------------------------------------------------------------------

// Find all papers by a specific author (ordered by year)
// Parameters: $authorName (string)
MATCH (p:Person)-[:AUTHORED]->(d:Document)
WHERE p.name = $authorName
RETURN d.id, d.title, d.year, d.venue, d.citationCount
ORDER BY d.year DESC;

// Find an author's most cited papers
// Parameters: $authorName (string), $limit (integer, default 10)
MATCH (p:Person)-[:AUTHORED]->(d:Document)
WHERE p.name = $authorName
RETURN d.title, d.year, d.citationCount
ORDER BY d.citationCount DESC
LIMIT $limit;

// Find co-authors and collaboration frequency
// Parameters: $authorName (string)
MATCH (p1:Person)-[:AUTHORED]->(d:Document)<-[:AUTHORED]-(p2:Person)
WHERE p1.name = $authorName AND p1 <> p2
RETURN p2.name as coauthor,
       p2.id as coauthorId,
       count(d) as collaborations,
       collect(d.title) as sharedPapers
ORDER BY collaborations DESC;

// Find author's research institutions over time
// Parameters: $authorName (string)
MATCH (p:Person)-[:AFFILIATED_WITH]->(i:Institution)
WHERE p.name = $authorName
RETURN i.name, i.country, i.type
ORDER BY i.name;

// ----------------------------------------------------------------------------
// CITATION NETWORK QUERIES
// ----------------------------------------------------------------------------

// Find direct citations from a paper
// Parameters: $documentId (string)
MATCH (d1:Document)-[:CITES]->(d2:Document)
WHERE d1.id = $documentId
RETURN d2.id, d2.title, d2.year, d2.citationCount
ORDER BY d2.year DESC;

// Find papers that cite a specific paper (incoming citations)
// Parameters: $documentId (string)
MATCH (d1:Document)-[:CITES]->(d2:Document)
WHERE d2.id = $documentId
RETURN d1.id, d1.title, d1.year, d1.citationCount
ORDER BY d1.citationCount DESC;

// Find citation chain (papers citing papers up to 3 levels deep)
// Parameters: $documentId (string)
MATCH path = (d1:Document)-[:CITES*1..3]->(d2:Document)
WHERE d1.id = $documentId
RETURN path
LIMIT 100;

// Find common citations between two papers (bibliographic coupling)
// Parameters: $documentId1 (string), $documentId2 (string)
MATCH (d1:Document)-[:CITES]->(cited:Document)<-[:CITES]-(d2:Document)
WHERE d1.id = $documentId1 AND d2.id = $documentId2
RETURN cited.id, cited.title, cited.year, cited.citationCount
ORDER BY cited.citationCount DESC;

// Find papers cited by both papers (co-citation analysis)
// Parameters: $documentId1 (string), $documentId2 (string)
MATCH (citing:Document)-[:CITES]->(d1:Document),
      (citing)-[:CITES]->(d2:Document)
WHERE d1.id = $documentId1 AND d2.id = $documentId2
RETURN citing.id, citing.title, citing.year
ORDER BY citing.year DESC;

// ----------------------------------------------------------------------------
// TOPIC/CONCEPT QUERIES
// ----------------------------------------------------------------------------

// Find papers on a specific topic/concept
// Parameters: $conceptName (string)
MATCH (c:Concept)<-[:STUDIES]-(d:Document)
WHERE c.name = $conceptName
RETURN d.id, d.title, d.year, d.citationCount
ORDER BY d.citationCount DESC;

// Find related concepts based on shared papers
// Parameters: $conceptName (string), $limit (integer, default 10)
MATCH (c1:Concept)<-[:STUDIES]-(d:Document)-[:STUDIES]->(c2:Concept)
WHERE c1.name = $conceptName AND c1 <> c2
RETURN c2.name as relatedConcept,
       count(d) as sharedPapers,
       collect(d.title)[0..5] as samplePapers
ORDER BY sharedPapers DESC
LIMIT $limit;

// Find trending concepts by year
// Parameters: $year (integer)
MATCH (d:Document)-[:STUDIES]->(c:Concept)
WHERE d.year = $year
RETURN c.name, count(d) as paperCount
ORDER BY paperCount DESC
LIMIT 20;

// Find author's research topics
// Parameters: $authorName (string)
MATCH (p:Person)-[:AUTHORED]->(d:Document)-[:STUDIES]->(c:Concept)
WHERE p.name = $authorName
RETURN c.name as topic,
       count(d) as paperCount,
       collect(d.title)[0..3] as samplePapers
ORDER BY paperCount DESC;

// ----------------------------------------------------------------------------
// SEMANTIC SEARCH QUERIES
// ----------------------------------------------------------------------------

// Semantic search for similar papers using vector embeddings
// Parameters: $queryEmbedding (list of floats), $limit (integer, default 10)
CALL db.index.vector.queryNodes('document_embedding', $limit, $queryEmbedding)
YIELD node as d, score
RETURN d.id, d.title, d.abstract, d.year, score
ORDER BY score DESC;

// Hybrid search: combine fulltext and vector search
// Parameters: $queryText (string), $queryEmbedding (list of floats), $limit (integer)
CALL {
  // Fulltext search
  CALL db.index.fulltext.queryNodes('document_fulltext', $queryText)
  YIELD node, score
  RETURN node as d, score as textScore, 0.0 as vectorScore
  LIMIT $limit
  UNION
  // Vector search
  CALL db.index.vector.queryNodes('document_embedding', $limit, $queryEmbedding)
  YIELD node, score
  RETURN node as d, 0.0 as textScore, score as vectorScore
}
WITH d, max(textScore) as maxTextScore, max(vectorScore) as maxVectorScore
RETURN d.id, d.title, d.abstract,
       (maxTextScore * 0.4 + maxVectorScore * 0.6) as combinedScore
ORDER BY combinedScore DESC
LIMIT $limit;

// Find semantically similar concepts
// Parameters: $conceptEmbedding (list of floats), $limit (integer, default 10)
CALL db.index.vector.queryNodes('concept_embedding', $limit, $conceptEmbedding)
YIELD node as c, score
RETURN c.name, c.description, score
ORDER BY score DESC;

// ----------------------------------------------------------------------------
// RESEARCH COMMUNITY QUERIES
// ----------------------------------------------------------------------------

// Find research communities (authors frequently collaborating)
// Returns clusters of co-authors
MATCH (p1:Person)-[:AUTHORED]->(:Document)<-[:AUTHORED]-(p2:Person)
WHERE id(p1) < id(p2)
WITH p1, p2, count(*) as collaborations
WHERE collaborations >= 3
RETURN p1.name, p2.name, collaborations
ORDER BY collaborations DESC;

// Find institutions with most papers in a concept area
// Parameters: $conceptName (string), $limit (integer, default 10)
MATCH (i:Institution)<-[:AFFILIATED_WITH]-(p:Person)-[:AUTHORED]->(d:Document)-[:STUDIES]->(c:Concept)
WHERE c.name = $conceptName
RETURN i.name, i.country, count(DISTINCT d) as paperCount, count(DISTINCT p) as authorCount
ORDER BY paperCount DESC
LIMIT $limit;

// Find cross-institutional collaborations
// Parameters: $minPapers (integer, default 3)
MATCH (i1:Institution)<-[:AFFILIATED_WITH]-(p1:Person)-[:AUTHORED]->(d:Document)<-[:AUTHORED]-(p2:Person)-[:AFFILIATED_WITH]->(i2:Institution)
WHERE id(i1) < id(i2)
WITH i1, i2, count(DISTINCT d) as sharedPapers
WHERE sharedPapers >= $minPapers
RETURN i1.name, i2.name, sharedPapers
ORDER BY sharedPapers DESC;

// ----------------------------------------------------------------------------
// TEMPORAL QUERIES
// ----------------------------------------------------------------------------

// Find papers published in a specific year range
// Parameters: $startYear (integer), $endYear (integer)
MATCH (d:Document)
WHERE d.year >= $startYear AND d.year <= $endYear
RETURN d.id, d.title, d.year, d.venue
ORDER BY d.year DESC, d.citationCount DESC;

// Track concept evolution over time
// Parameters: $conceptName (string)
MATCH (c:Concept)<-[:STUDIES]-(d:Document)
WHERE c.name = $conceptName
WITH d.year as year, count(d) as paperCount
RETURN year, paperCount
ORDER BY year ASC;

// Find author's publication timeline
// Parameters: $authorName (string)
MATCH (p:Person)-[:AUTHORED]->(d:Document)
WHERE p.name = $authorName
WITH d.year as year, count(d) as papers, sum(d.citationCount) as totalCitations
RETURN year, papers, totalCitations
ORDER BY year ASC;

// ----------------------------------------------------------------------------
// INFLUENCE QUERIES
// ----------------------------------------------------------------------------

// Find most influential papers (by citation count)
// Parameters: $limit (integer, default 20)
MATCH (d:Document)
RETURN d.id, d.title, d.year, d.citationCount, d.venue
ORDER BY d.citationCount DESC
LIMIT $limit;

// Find most prolific authors
// Parameters: $limit (integer, default 20)
MATCH (p:Person)-[:AUTHORED]->(d:Document)
WITH p, count(d) as paperCount, sum(d.citationCount) as totalCitations
RETURN p.name, p.id, paperCount, totalCitations,
       toInteger(totalCitations * 1.0 / paperCount) as avgCitations
ORDER BY paperCount DESC
LIMIT $limit;

// Find rising stars (authors with recent high-impact papers)
// Parameters: $recentYears (integer, default 3), $limit (integer, default 10)
MATCH (p:Person)-[:AUTHORED]->(d:Document)
WHERE d.year >= date().year - $recentYears
WITH p, count(d) as recentPapers, sum(d.citationCount) as recentCitations
WHERE recentPapers >= 2
RETURN p.name, p.id, recentPapers, recentCitations,
       toInteger(recentCitations * 1.0 / recentPapers) as avgCitations
ORDER BY avgCitations DESC
LIMIT $limit;

// Find seminal papers (older papers still being cited)
// Parameters: $oldYear (integer), $minCitations (integer)
MATCH (d:Document)
WHERE d.year <= $oldYear AND d.citationCount >= $minCitations
RETURN d.id, d.title, d.year, d.citationCount, d.venue
ORDER BY d.citationCount DESC;
