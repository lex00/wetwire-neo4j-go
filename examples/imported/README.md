# Imported Examples

This directory contains examples imported and adapted from Neo4j's official documentation, templates, and reference implementations. These examples demonstrate real-world graph patterns and use cases.

## Structure

```
imported/
├── neo4j-aura/        # Examples from Neo4j Aura templates
├── neo4j-gds/         # Examples from Graph Data Science documentation
└── neo4j-graphrag/    # Examples from Neo4j Labs GraphRAG projects
```

## Neo4j Aura Examples

### Fraud Detection (`neo4j-aura/fraud_detection.go`)

A comprehensive fraud detection pattern for financial transactions and identity networks.

**Use Cases:**
- Transaction fraud ring detection
- First-party fraud (synthetic identities)
- Account takeover detection
- Shared identifier analysis (email, phone, IP, device)

**Key Components:**
- **Nodes:** Customer, Account, Transaction, Email, Phone, IPAddress, Device, Merchant
- **Algorithms:** WCC for fraud rings, Node Similarity for customer patterns, PageRank for influential nodes, Louvain for community detection
- **Pattern:** Detects fraud by identifying customers who share contact information, devices, or IP addresses

**Sources:**
- [Exploring Fraud Detection With Neo4j & Graph Data Science – Part 1](https://neo4j.com/blog/developer/exploring-fraud-detection-neo4j-graph-data-science-part-1/)
- [Neo4j Fraud Demo](https://neo4j.com/developer/demos/fraud-demo/)
- [Mastering Fraud Detection With Temporal Graph Modeling](https://neo4j.com/blog/developer/mastering-fraud-detection-temporal-graph/)
- [Neo4j Aura Graph Analytics Demo: Fraud Detection in P2P Networks](https://neo4j.com/videos/neo4j-aura-graph-analytics-demo-fraud-detection-in-p2p-networks-using/)

## Neo4j Graph Data Science Examples

### Recommendation Engine (`neo4j-gds/recommendation_engine.go`)

A product recommendation system using collaborative filtering and content-based approaches.

**Use Cases:**
- E-commerce product recommendations
- Content recommendations (articles, videos)
- Personalized search results
- Cross-sell and upsell suggestions

**Key Components:**
- **Nodes:** User, Product, Category, Tag, Brand
- **Relationships:** RATED (explicit feedback), PURCHASED, VIEWED (implicit feedback)
- **Algorithms:**
  - Node Similarity for collaborative filtering
  - FastRP for product embeddings
  - KNN for similar product recommendations
  - Louvain for product communities
  - PageRank for popularity
- **ML Pipelines:** Link prediction for user-product recommendations, node classification for user segmentation

**Sources:**
- [Building a Recommendation Engine Using Neo4j Hands-On — Part 1](https://neo4j.com/blog/developer/recommendation-engine-hands-on-1/)
- [Exploring Practical Recommendation Engines In Neo4j](https://towardsdatascience.com/exploring-practical-recommendation-engines-in-neo4j-ff09fe767782/)
- [Explore the Power of Neo4j: Building a Recommendation System Powered by Graph Data Science](https://www.515tech.com/post/explore-the-power-of-neo4j-building-a-recommendation-system-powered-by-graph-data-science)
- [Tutorial: Build a Cypher Recommendation Engine](https://neo4j.com/docs/getting-started/appendix/tutorials/guide-build-a-recommendation-engine/)

## Neo4j Labs GraphRAG Examples

### Knowledge Graph Q&A (`neo4j-graphrag/knowledge_graph_qa.go`)

A GraphRAG implementation for question-answering over documents using knowledge graphs.

**Use Cases:**
- Document Q&A systems
- Research paper analysis
- Technical documentation search
- Enterprise knowledge management

**Key Components:**
- **Nodes:** Document, Chunk, Entity, Concept, Topic
- **Schema:** Supports document chunking, entity extraction, relationship mapping
- **KG Pipeline:** Automated entity and relationship extraction using LLMs
- **Retrievers:**
  - Vector retriever for semantic search
  - Hybrid retriever combining vector + fulltext
  - Text2Cypher for natural language queries
  - GraphRAG retriever with multi-hop context expansion

**Pattern:** Documents are chunked, entities are extracted, and embeddings are created. Retrieval combines vector similarity with graph traversal to enrich context.

**Sources:**
- [Neo4j LLM Knowledge Graph Builder](https://neo4j.com/labs/genai-ecosystem/llm-graph-builder/)
- [Neo4j GraphRAG Python Package](https://neo4j.com/labs/genai-ecosystem/graphrag-python/)
- [Knowledge Graph and GraphRAG courses on GraphAcademy](https://graphacademy.neo4j.com/knowledge-graph-rag/)
- [The GraphRAG Manifesto: Adding Knowledge to GenAI](https://neo4j.com/blog/genai/graphrag-manifesto/)
- [GitHub: neo4j/neo4j-graphrag-python](https://github.com/neo4j/neo4j-graphrag-python)
- [GitHub: neo4j-labs/llm-graph-builder](https://github.com/neo4j-labs/llm-graph-builder)

## Running Examples

All examples are Go package definitions that demonstrate wetwire-neo4j-go patterns. They are validated through tests:

```bash
# Run all tests including imported examples
go test ./examples/... -v

# Build examples to verify compilation
go build ./examples/imported/...
```

## Using These Examples

These examples can be used as:

1. **Reference implementations** - Study the patterns and adapt them to your use case
2. **Starting templates** - Copy and modify for your specific domain
3. **Learning resources** - Understand how to structure graph schemas and GDS workflows

## Adapting Examples

To adapt these examples for your use case:

1. **Schema:** Modify node and relationship types to match your domain
2. **Properties:** Add or remove properties based on your data
3. **Algorithms:** Choose GDS algorithms that fit your analytical needs
4. **Constraints:** Add appropriate uniqueness and existence constraints
5. **Indexes:** Create indexes on properties you'll query frequently

## Additional Resources

- [Neo4j Graph Data Science Documentation](https://neo4j.com/docs/graph-data-science/current/)
- [Neo4j Aura Documentation](https://neo4j.com/docs/aura/)
- [Neo4j GraphRAG Developer Guide](https://neo4j.com/developer/genai-ecosystem/)
- [Neo4j Use Cases](https://neo4j.com/use-cases/)
- [Neo4j Graph Academy](https://graphacademy.neo4j.com/)

## Contributing

When adding new imported examples:

1. Create a new file in the appropriate subdirectory
2. Include comprehensive documentation with references
3. Add the example to this README with description and sources
4. Ensure examples compile and follow Go conventions
5. Add appropriate `AgentHint` fields for AI-assisted development

## License

These examples are adapted from Neo4j's public documentation and examples. See individual source links for original licenses. The wetwire-neo4j-go implementations are provided as educational resources.
