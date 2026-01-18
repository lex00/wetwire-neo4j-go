package neo4j_gds

import (
	"github.com/lex00/wetwire-neo4j-go/internal/algorithms"
	"github.com/lex00/wetwire-neo4j-go/internal/pipelines"
	"github.com/lex00/wetwire-neo4j-go/internal/projections"
	"github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"
)

// RecommendationEngineSchema demonstrates a recommendation system pattern.
// This schema supports collaborative filtering and content-based recommendations using:
// - User-item interactions (ratings, purchases, views)
// - Item similarity and user similarity
// - Graph-based recommendations with GDS algorithms
//
// References:
// - https://neo4j.com/blog/developer/recommendation-engine-hands-on-1/
// - https://towardsdatascience.com/exploring-practical-recommendation-engines-in-neo4j-ff09fe767782
// - https://www.515tech.com/post/explore-the-power-of-neo4j-building-a-recommendation-system-powered-by-graph-data-science
var RecommendationEngineSchema = &schema.Schema{
	Name:        "recommendation-engine",
	Description: "Product recommendation system with collaborative filtering",
	AgentContext: "This schema tracks user interactions with products to generate personalized recommendations. " +
		"Use rating strength, view counts, and purchase history for recommendation quality.",
	Nodes: []*schema.NodeType{
		UserNode,
		ProductNode,
		CategoryNode,
		TagNode,
		BrandNode,
	},
	Relationships: []*schema.RelationshipType{
		RatedRel,
		PurchasedRel,
		ViewedRel,
		InCategoryRel,
		HasTagRel,
		ManufacturedByRel,
		SimilarProductRel,
	},
}

// UserNode represents a user in the recommendation system.
var UserNode = &schema.NodeType{
	Label:       "User",
	Description: "A user who interacts with products",
	Properties: []schema.Property{
		{Name: "userId", Type: schema.STRING, Required: true, Unique: true},
		{Name: "username", Type: schema.STRING, Required: true},
		{Name: "email", Type: schema.STRING, Unique: true},
		{Name: "joinedAt", Type: schema.DATETIME, Required: true},
		{Name: "preferences", Type: schema.LIST_STRING, Description: "User preference tags"},
		{Name: "location", Type: schema.STRING},
		{Name: "ageGroup", Type: schema.STRING, Description: "18-24, 25-34, 35-44, 45-54, 55+"},
	},
	Constraints: []schema.Constraint{
		{Type: schema.UNIQUE, Properties: []string{"userId"}},
		{Type: schema.UNIQUE, Properties: []string{"email"}},
	},
	Indexes: []schema.Index{
		{Type: schema.BTREE, Properties: []string{"location"}},
		{Type: schema.BTREE, Properties: []string{"ageGroup"}},
	},
	AgentHint: "Query by userId for unique identification. Use preferences and location for personalization.",
}

// ProductNode represents a product that can be recommended.
var ProductNode = &schema.NodeType{
	Label:       "Product",
	Description: "A product available for purchase or rating",
	Properties: []schema.Property{
		{Name: "productId", Type: schema.STRING, Required: true, Unique: true},
		{Name: "name", Type: schema.STRING, Required: true},
		{Name: "description", Type: schema.STRING},
		{Name: "price", Type: schema.FLOAT, Required: true},
		{Name: "stock", Type: schema.INTEGER},
		{Name: "releaseDate", Type: schema.DATE},
		{Name: "averageRating", Type: schema.FLOAT, Description: "Average user rating"},
		{Name: "ratingCount", Type: schema.INTEGER, Description: "Number of ratings"},
		{Name: "viewCount", Type: schema.INTEGER, Description: "Total views"},
		{Name: "purchaseCount", Type: schema.INTEGER, Description: "Total purchases"},
		{Name: "embedding", Type: schema.LIST_FLOAT, Description: "Product embedding vector"},
	},
	Constraints: []schema.Constraint{
		{Type: schema.UNIQUE, Properties: []string{"productId"}},
	},
	Indexes: []schema.Index{
		{Type: schema.BTREE, Properties: []string{"averageRating"}},
		{Type: schema.BTREE, Properties: []string{"price"}},
		{Type: schema.TEXT, Properties: []string{"name", "description"}},
	},
	AgentHint: "Query by productId for unique identification. Sort by averageRating and purchaseCount for popularity.",
}

// CategoryNode represents a product category.
var CategoryNode = &schema.NodeType{
	Label:       "Category",
	Description: "Product category for content-based filtering",
	Properties: []schema.Property{
		{Name: "categoryId", Type: schema.STRING, Required: true, Unique: true},
		{Name: "name", Type: schema.STRING, Required: true},
		{Name: "description", Type: schema.STRING},
		{Name: "parentCategory", Type: schema.STRING, Description: "Parent category ID for hierarchy"},
	},
	Constraints: []schema.Constraint{
		{Type: schema.UNIQUE, Properties: []string{"categoryId"}},
	},
	AgentHint: "Use for content-based recommendations within same category.",
}

// TagNode represents a product tag or attribute.
var TagNode = &schema.NodeType{
	Label:       "Tag",
	Description: "Product tag for content-based filtering",
	Properties: []schema.Property{
		{Name: "tagId", Type: schema.STRING, Required: true, Unique: true},
		{Name: "name", Type: schema.STRING, Required: true},
		{Name: "type", Type: schema.STRING, Description: "feature, style, use-case, etc."},
	},
	Constraints: []schema.Constraint{
		{Type: schema.UNIQUE, Properties: []string{"tagId"}},
	},
	Indexes: []schema.Index{
		{Type: schema.BTREE, Properties: []string{"type"}},
	},
	AgentHint: "Use for finding similar products with shared tags.",
}

// BrandNode represents a product brand/manufacturer.
var BrandNode = &schema.NodeType{
	Label:       "Brand",
	Description: "Product brand or manufacturer",
	Properties: []schema.Property{
		{Name: "brandId", Type: schema.STRING, Required: true, Unique: true},
		{Name: "name", Type: schema.STRING, Required: true},
		{Name: "country", Type: schema.STRING},
		{Name: "reputation", Type: schema.FLOAT, Description: "Brand reputation score"},
	},
	Constraints: []schema.Constraint{
		{Type: schema.UNIQUE, Properties: []string{"brandId"}},
	},
	AgentHint: "Use for brand-based recommendations and filtering.",
}

// Relationships

var RatedRel = &schema.RelationshipType{
	Label:       "RATED",
	Source:      "User",
	Target:      "Product",
	Cardinality: schema.MANY_TO_MANY,
	Description: "User rated a product (explicit feedback)",
	Properties: []schema.Property{
		{Name: "rating", Type: schema.FLOAT, Required: true, Description: "Rating value (1-5)"},
		{Name: "timestamp", Type: schema.DATETIME, Required: true},
		{Name: "review", Type: schema.STRING, Description: "Optional review text"},
	},
	AgentHint: "Primary signal for collaborative filtering. Higher ratings indicate stronger preference.",
}

var PurchasedRel = &schema.RelationshipType{
	Label:       "PURCHASED",
	Source:      "User",
	Target:      "Product",
	Cardinality: schema.MANY_TO_MANY,
	Description: "User purchased a product (implicit feedback)",
	Properties: []schema.Property{
		{Name: "timestamp", Type: schema.DATETIME, Required: true},
		{Name: "quantity", Type: schema.INTEGER, Required: true},
		{Name: "price", Type: schema.FLOAT, Description: "Purchase price"},
	},
	AgentHint: "Strong implicit signal for recommendation. Weight purchases higher than views.",
}

var ViewedRel = &schema.RelationshipType{
	Label:       "VIEWED",
	Source:      "User",
	Target:      "Product",
	Cardinality: schema.MANY_TO_MANY,
	Description: "User viewed a product (weak implicit feedback)",
	Properties: []schema.Property{
		{Name: "timestamp", Type: schema.DATETIME, Required: true},
		{Name: "durationSeconds", Type: schema.INTEGER, Description: "View duration"},
	},
	AgentHint: "Weak implicit signal. Consider view duration for signal strength.",
}

var InCategoryRel = &schema.RelationshipType{
	Label:       "IN_CATEGORY",
	Source:      "Product",
	Target:      "Category",
	Cardinality: schema.MANY_TO_ONE,
	Description: "Product belongs to a category",
}

var HasTagRel = &schema.RelationshipType{
	Label:       "HAS_TAG",
	Source:      "Product",
	Target:      "Tag",
	Cardinality: schema.MANY_TO_MANY,
	Description: "Product has a tag attribute",
	Properties: []schema.Property{
		{Name: "relevance", Type: schema.FLOAT, Description: "Tag relevance score"},
	},
}

var ManufacturedByRel = &schema.RelationshipType{
	Label:       "MANUFACTURED_BY",
	Source:      "Product",
	Target:      "Brand",
	Cardinality: schema.MANY_TO_ONE,
	Description: "Product is manufactured by a brand",
}

var SimilarProductRel = &schema.RelationshipType{
	Label:       "SIMILAR_PRODUCT",
	Source:      "Product",
	Target:      "Product",
	Cardinality: schema.MANY_TO_MANY,
	Description: "Products are similar (computed by GDS)",
	Properties: []schema.Property{
		{Name: "similarity", Type: schema.FLOAT, Required: true, Description: "Similarity score 0-1"},
		{Name: "computedAt", Type: schema.DATETIME},
		{Name: "method", Type: schema.STRING, Description: "Algorithm used: node-similarity, knn, etc."},
	},
	AgentHint: "Generated by GDS algorithms for product recommendations.",
}

// GDS Graph Projection for recommendation analysis
var RecommendationProjection = &projections.NativeProjection{
	BaseProjection: projections.BaseProjection{
		GraphName: "recommendation-graph",
	},
	NodeProjections: []projections.NodeProjection{
		{Label: "User", Properties: []string{"preferences", "location"}},
		{Label: "Product", Properties: []string{"price", "averageRating", "embedding"}},
		{Label: "Category"},
		{Label: "Tag"},
		{Label: "Brand", Properties: []string{"reputation"}},
	},
	RelationshipProjections: []projections.RelationshipProjection{
		{Type: "RATED", Properties: []string{"rating", "timestamp"}},
		{Type: "PURCHASED", Properties: []string{"quantity"}},
		{Type: "VIEWED"},
		{Type: "IN_CATEGORY"},
		{Type: "HAS_TAG", Properties: []string{"relevance"}},
		{Type: "MANUFACTURED_BY"},
	},
}

// Product Similarity using Node Similarity algorithm
// Finds similar products based on shared user interactions.
// Reference: https://neo4j.com/docs/graph-data-science/current/algorithms/node-similarity/
var ProductSimilarityAlgorithm = &algorithms.NodeSimilarity{
	BaseAlgorithm: algorithms.BaseAlgorithm{
		GraphName: "recommendation-graph",
		Mode:      algorithms.Mutate,
	},
	TopK:                  20,
	SimilarityCutoff:      0.1,
	DegreeCutoff:          3,
	WriteRelationshipType: "SIMILAR_PRODUCT",
	WriteProperty:         "similarity",
}

// Product Embeddings using FastRP
// Creates vector embeddings for products based on graph structure.
// Reference: https://neo4j.com/docs/graph-data-science/current/machine-learning/node-embeddings/fastrp/
var ProductEmbeddingsFastRP = &algorithms.FastRP{
	BaseAlgorithm: algorithms.BaseAlgorithm{
		GraphName: "recommendation-graph",
		Mode:      algorithms.Mutate,
	},
	EmbeddingDimension:    128,
	IterationWeights:      []float64{0.0, 1.0, 1.0, 0.5},
	NormalizationStrength: 0.5,
	MutateProperty:        "embedding",
}

// K-Nearest Neighbors for product recommendations
// Finds similar products using embedding vectors.
// Reference: https://neo4j.com/docs/graph-data-science/current/algorithms/knn/
var ProductKNNRecommendations = &algorithms.KNN{
	BaseAlgorithm: algorithms.BaseAlgorithm{
		GraphName: "recommendation-graph",
		Mode:      algorithms.Mutate,
	},
	NodeProperties:        []string{"embedding"},
	K:                     10,
	SimilarityCutoff:      0.5,
	WriteRelationshipType: "SIMILAR_PRODUCT",
	WriteProperty:         "similarity",
}

// Community Detection with Louvain
// Identifies product communities for diverse recommendations.
var ProductCommunityDetection = &algorithms.Louvain{
	BaseAlgorithm: algorithms.BaseAlgorithm{
		GraphName: "recommendation-graph",
		Mode:      algorithms.Mutate,
	},
	MaxLevels:      10,
	MaxIterations:  10,
	Tolerance:      0.0001,
	MutateProperty: "communityId",
}

// PageRank for identifying popular/influential products
// High PageRank products are frequently co-purchased.
var ProductPopularityPageRank = &algorithms.PageRank{
	BaseAlgorithm: algorithms.BaseAlgorithm{
		GraphName: "recommendation-graph",
		Mode:      algorithms.Mutate,
	},
	DampingFactor:  0.85,
	MaxIterations:  20,
	MutateProperty: "popularity",
}

// Link Prediction Pipeline for user-product recommendations
// Predicts which products a user is likely to rate or purchase.
// Reference: https://neo4j.com/docs/graph-data-science/current/machine-learning/linkprediction-pipelines/
var UserProductLinkPrediction = &pipelines.LinkPredictionPipeline{
	BasePipeline: pipelines.BasePipeline{
		Name:       "user-product-recommendations",
		NodeLabels: []string{"User", "Product"},
		FeatureSteps: []pipelines.FeatureStep{
			&pipelines.PageRankStep{
				Property:      "popularity",
				DampingFactor: 0.85,
				MaxIterations: 20,
			},
			&pipelines.DegreeStep{
				Property: "activityScore",
			},
			&pipelines.FastRPStep{
				Property:           "embedding",
				EmbeddingDimension: 64,
				IterationWeights:   []float64{0.0, 1.0, 1.0},
			},
		},
		Models: []pipelines.Model{
			&pipelines.LogisticRegression{
				Penalty:   0.001,
				MaxEpochs: 100,
			},
		},
	},
	SplitConfig: pipelines.SplitConfig{
		ValidationFolds: 5,
		RandomSeed:      42,
	},
	TargetRelationshipType: "RATED",
}

// Node Classification Pipeline for predicting user preferences
// Classifies users into preference segments.
var UserPreferenceClassification = &pipelines.NodeClassificationPipeline{
	BasePipeline: pipelines.BasePipeline{
		Name:       "user-preference-classification",
		NodeLabels: []string{"User"},
		FeatureSteps: []pipelines.FeatureStep{
			&pipelines.FastRPStep{
				Property:           "embedding",
				EmbeddingDimension: 64,
				IterationWeights:   []float64{0.0, 1.0, 1.0},
			},
			&pipelines.DegreeStep{
				Property: "purchaseCount",
			},
		},
		Models: []pipelines.Model{
			&pipelines.RandomForest{
				NumTrees: 100,
				MaxDepth: 10,
			},
		},
	},
	SplitConfig: pipelines.SplitConfig{
		ValidationFolds: 3,
		RandomSeed:      42,
	},
	TargetProperty:   "preferenceSegment",
	TargetNodeLabels: []string{"User"},
}
