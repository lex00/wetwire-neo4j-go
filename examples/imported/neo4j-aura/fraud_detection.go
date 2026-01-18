package neo4j_aura

import (
	"github.com/lex00/wetwire-neo4j-go/internal/algorithms"
	"github.com/lex00/wetwire-neo4j-go/internal/projections"
	"github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"
)

// FraudDetectionSchema demonstrates a fraud detection graph pattern from Neo4j Aura.
// This schema is based on common fraud detection use cases including:
// - Transaction fraud rings
// - First-party fraud (synthetic identities)
// - Account takeover detection
//
// References:
// - https://neo4j.com/blog/developer/exploring-fraud-detection-neo4j-graph-data-science-part-1/
// - https://neo4j.com/developer/demos/fraud-demo/
// - https://neo4j.com/blog/developer/mastering-fraud-detection-temporal-graph/
var FraudDetectionSchema = &schema.Schema{
	Name:        "fraud-detection",
	Description: "Fraud detection pattern for financial transactions and identity networks",
	AgentContext: "This schema tracks financial entities and their relationships to detect fraud patterns. " +
		"Focus on relationship patterns like shared identifiers (email, phone, IP) and transaction velocity.",
	Nodes: []*schema.NodeType{
		CustomerNode,
		AccountNode,
		TransactionNode,
		EmailNode,
		PhoneNode,
		IPAddressNode,
		DeviceNode,
		MerchantNode,
	},
	Relationships: []*schema.RelationshipType{
		HasAccountRel,
		PerformedTransactionRel,
		HasEmailRel,
		HasPhoneRel,
		UsedIPRel,
		UsedDeviceRel,
		TransactionToMerchantRel,
		SimilarToRel,
	},
}

// CustomerNode represents a customer in the fraud detection system.
var CustomerNode = &schema.NodeType{
	Label:       "Customer",
	Description: "A customer who owns accounts and performs transactions",
	Properties: []schema.Property{
		{Name: "customerId", Type: schema.STRING, Required: true, Unique: true},
		{Name: "name", Type: schema.STRING, Required: true},
		{Name: "ssn", Type: schema.STRING, Description: "Social Security Number (encrypted)"},
		{Name: "dateOfBirth", Type: schema.DATE},
		{Name: "registrationDate", Type: schema.DATETIME, Required: true},
		{Name: "kycStatus", Type: schema.STRING, Description: "Know Your Customer verification status"},
		{Name: "riskScore", Type: schema.FLOAT, Description: "Calculated fraud risk score"},
		{Name: "isFlagged", Type: schema.BOOLEAN, Description: "Flagged for suspicious activity"},
	},
	Constraints: []schema.Constraint{
		{Type: schema.UNIQUE, Properties: []string{"customerId"}},
	},
	Indexes: []schema.Index{
		{Type: schema.BTREE, Properties: []string{"riskScore"}},
		{Type: schema.BTREE, Properties: []string{"isFlagged"}},
	},
	AgentHint: "Query by customerId for unique identification. Use riskScore for fraud analysis.",
}

// AccountNode represents a financial account.
var AccountNode = &schema.NodeType{
	Label:       "Account",
	Description: "A financial account (checking, savings, credit card, etc.)",
	Properties: []schema.Property{
		{Name: "accountId", Type: schema.STRING, Required: true, Unique: true},
		{Name: "accountType", Type: schema.STRING, Required: true, Description: "checking, savings, credit, etc."},
		{Name: "balance", Type: schema.FLOAT},
		{Name: "openDate", Type: schema.DATETIME, Required: true},
		{Name: "status", Type: schema.STRING, Description: "active, suspended, closed"},
		{Name: "dailyTransactionLimit", Type: schema.FLOAT},
		{Name: "isBlocked", Type: schema.BOOLEAN},
	},
	Constraints: []schema.Constraint{
		{Type: schema.UNIQUE, Properties: []string{"accountId"}},
	},
	Indexes: []schema.Index{
		{Type: schema.BTREE, Properties: []string{"status"}},
		{Type: schema.BTREE, Properties: []string{"isBlocked"}},
	},
	AgentHint: "Query by accountId for unique identification. Filter by status and isBlocked for fraud monitoring.",
}

// TransactionNode represents a financial transaction.
var TransactionNode = &schema.NodeType{
	Label:       "Transaction",
	Description: "A financial transaction between accounts or to merchants",
	Properties: []schema.Property{
		{Name: "transactionId", Type: schema.STRING, Required: true, Unique: true},
		{Name: "amount", Type: schema.FLOAT, Required: true},
		{Name: "currency", Type: schema.STRING, Required: true},
		{Name: "timestamp", Type: schema.DATETIME, Required: true},
		{Name: "type", Type: schema.STRING, Description: "purchase, transfer, withdrawal, deposit"},
		{Name: "status", Type: schema.STRING, Description: "pending, completed, failed, declined"},
		{Name: "fraudScore", Type: schema.FLOAT, Description: "ML-based fraud score"},
		{Name: "isDeclined", Type: schema.BOOLEAN},
		{Name: "declineReason", Type: schema.STRING},
		{Name: "location", Type: schema.POINT, Description: "Transaction location"},
	},
	Constraints: []schema.Constraint{
		{Type: schema.UNIQUE, Properties: []string{"transactionId"}},
	},
	Indexes: []schema.Index{
		{Type: schema.BTREE, Properties: []string{"timestamp"}},
		{Type: schema.BTREE, Properties: []string{"fraudScore"}},
		{Type: schema.BTREE, Properties: []string{"isDeclined"}},
		{Type: schema.POINT_INDEX, Properties: []string{"location"}},
	},
	AgentHint: "Use timestamp for temporal analysis. Sort by fraudScore for risk prioritization.",
}

// EmailNode represents an email address shared across customers.
var EmailNode = &schema.NodeType{
	Label:       "Email",
	Description: "Email address that may be shared across multiple customers (fraud indicator)",
	Properties: []schema.Property{
		{Name: "address", Type: schema.STRING, Required: true, Unique: true},
		{Name: "domain", Type: schema.STRING, Description: "Email domain for pattern analysis"},
		{Name: "isDisposable", Type: schema.BOOLEAN, Description: "Temporary/disposable email service"},
		{Name: "riskLevel", Type: schema.STRING, Description: "low, medium, high"},
	},
	Constraints: []schema.Constraint{
		{Type: schema.UNIQUE, Properties: []string{"address"}},
	},
	Indexes: []schema.Index{
		{Type: schema.BTREE, Properties: []string{"domain"}},
		{Type: schema.BTREE, Properties: []string{"isDisposable"}},
	},
	AgentHint: "Multiple customers sharing an email is a fraud indicator. Check isDisposable flag.",
}

// PhoneNode represents a phone number shared across customers.
var PhoneNode = &schema.NodeType{
	Label:       "Phone",
	Description: "Phone number that may be shared across multiple customers (fraud indicator)",
	Properties: []schema.Property{
		{Name: "number", Type: schema.STRING, Required: true, Unique: true},
		{Name: "countryCode", Type: schema.STRING},
		{Name: "type", Type: schema.STRING, Description: "mobile, landline, voip"},
		{Name: "isVirtual", Type: schema.BOOLEAN, Description: "Virtual/VOIP number"},
		{Name: "riskLevel", Type: schema.STRING, Description: "low, medium, high"},
	},
	Constraints: []schema.Constraint{
		{Type: schema.UNIQUE, Properties: []string{"number"}},
	},
	Indexes: []schema.Index{
		{Type: schema.BTREE, Properties: []string{"countryCode"}},
		{Type: schema.BTREE, Properties: []string{"isVirtual"}},
	},
	AgentHint: "Multiple customers sharing a phone number is a fraud indicator. Check isVirtual flag.",
}

// IPAddressNode represents an IP address used for transactions.
var IPAddressNode = &schema.NodeType{
	Label:       "IPAddress",
	Description: "IP address used for account access or transactions",
	Properties: []schema.Property{
		{Name: "address", Type: schema.STRING, Required: true, Unique: true},
		{Name: "country", Type: schema.STRING},
		{Name: "city", Type: schema.STRING},
		{Name: "isProxy", Type: schema.BOOLEAN, Description: "VPN/proxy detected"},
		{Name: "isTor", Type: schema.BOOLEAN, Description: "Tor exit node"},
		{Name: "riskLevel", Type: schema.STRING, Description: "low, medium, high"},
		{Name: "location", Type: schema.POINT, Description: "Geolocation of IP"},
	},
	Constraints: []schema.Constraint{
		{Type: schema.UNIQUE, Properties: []string{"address"}},
	},
	Indexes: []schema.Index{
		{Type: schema.BTREE, Properties: []string{"country"}},
		{Type: schema.BTREE, Properties: []string{"isProxy"}},
		{Type: schema.BTREE, Properties: []string{"isTor"}},
		{Type: schema.POINT_INDEX, Properties: []string{"location"}},
	},
	AgentHint: "Shared IPs and proxy/Tor usage are fraud indicators. Use location for geographic analysis.",
}

// DeviceNode represents a device used for transactions.
var DeviceNode = &schema.NodeType{
	Label:       "Device",
	Description: "Device fingerprint for transaction tracking",
	Properties: []schema.Property{
		{Name: "deviceId", Type: schema.STRING, Required: true, Unique: true},
		{Name: "type", Type: schema.STRING, Description: "mobile, desktop, tablet"},
		{Name: "os", Type: schema.STRING, Description: "Operating system"},
		{Name: "browser", Type: schema.STRING},
		{Name: "fingerprint", Type: schema.STRING, Description: "Device fingerprint hash"},
		{Name: "isEmulator", Type: schema.BOOLEAN},
		{Name: "riskLevel", Type: schema.STRING, Description: "low, medium, high"},
	},
	Constraints: []schema.Constraint{
		{Type: schema.UNIQUE, Properties: []string{"deviceId"}},
	},
	Indexes: []schema.Index{
		{Type: schema.BTREE, Properties: []string{"type"}},
		{Type: schema.BTREE, Properties: []string{"isEmulator"}},
	},
	AgentHint: "Multiple accounts using same device is a fraud indicator. Check isEmulator flag.",
}

// MerchantNode represents a merchant receiving transactions.
var MerchantNode = &schema.NodeType{
	Label:       "Merchant",
	Description: "Merchant or vendor receiving transactions",
	Properties: []schema.Property{
		{Name: "merchantId", Type: schema.STRING, Required: true, Unique: true},
		{Name: "name", Type: schema.STRING, Required: true},
		{Name: "category", Type: schema.STRING, Description: "Merchant category code"},
		{Name: "country", Type: schema.STRING},
		{Name: "riskLevel", Type: schema.STRING, Description: "low, medium, high"},
		{Name: "isFlagged", Type: schema.BOOLEAN, Description: "Flagged for suspicious activity"},
	},
	Constraints: []schema.Constraint{
		{Type: schema.UNIQUE, Properties: []string{"merchantId"}},
	},
	Indexes: []schema.Index{
		{Type: schema.BTREE, Properties: []string{"category"}},
		{Type: schema.BTREE, Properties: []string{"isFlagged"}},
	},
	AgentHint: "Query by merchantId for unique identification. Use category for pattern analysis.",
}

// Relationships

var HasAccountRel = &schema.RelationshipType{
	Label:       "HAS_ACCOUNT",
	Source:      "Customer",
	Target:      "Account",
	Cardinality: schema.ONE_TO_MANY,
	Description: "Customer owns an account",
	Properties: []schema.Property{
		{Name: "since", Type: schema.DATETIME, Description: "When account was opened"},
		{Name: "isPrimary", Type: schema.BOOLEAN, Description: "Primary account for customer"},
	},
}

var PerformedTransactionRel = &schema.RelationshipType{
	Label:       "PERFORMED",
	Source:      "Account",
	Target:      "Transaction",
	Cardinality: schema.ONE_TO_MANY,
	Description: "Account performed a transaction",
	Properties: []schema.Property{
		{Name: "timestamp", Type: schema.DATETIME, Required: true},
	},
}

var HasEmailRel = &schema.RelationshipType{
	Label:       "HAS_EMAIL",
	Source:      "Customer",
	Target:      "Email",
	Cardinality: schema.MANY_TO_ONE,
	Description: "Customer has an email address (shared emails indicate fraud)",
	Properties: []schema.Property{
		{Name: "verifiedAt", Type: schema.DATETIME},
		{Name: "isPrimary", Type: schema.BOOLEAN},
	},
	AgentHint: "Multiple customers with same email is a strong fraud indicator",
}

var HasPhoneRel = &schema.RelationshipType{
	Label:       "HAS_PHONE",
	Source:      "Customer",
	Target:      "Phone",
	Cardinality: schema.MANY_TO_ONE,
	Description: "Customer has a phone number (shared phones indicate fraud)",
	Properties: []schema.Property{
		{Name: "verifiedAt", Type: schema.DATETIME},
		{Name: "isPrimary", Type: schema.BOOLEAN},
	},
	AgentHint: "Multiple customers with same phone is a strong fraud indicator",
}

var UsedIPRel = &schema.RelationshipType{
	Label:       "USED_IP",
	Source:      "Transaction",
	Target:      "IPAddress",
	Cardinality: schema.MANY_TO_ONE,
	Description: "Transaction was performed from an IP address",
	Properties: []schema.Property{
		{Name: "timestamp", Type: schema.DATETIME, Required: true},
	},
}

var UsedDeviceRel = &schema.RelationshipType{
	Label:       "USED_DEVICE",
	Source:      "Transaction",
	Target:      "Device",
	Cardinality: schema.MANY_TO_ONE,
	Description: "Transaction was performed on a device",
	Properties: []schema.Property{
		{Name: "timestamp", Type: schema.DATETIME, Required: true},
	},
}

var TransactionToMerchantRel = &schema.RelationshipType{
	Label:       "TO_MERCHANT",
	Source:      "Transaction",
	Target:      "Merchant",
	Cardinality: schema.MANY_TO_ONE,
	Description: "Transaction was made to a merchant",
	Properties: []schema.Property{
		{Name: "timestamp", Type: schema.DATETIME, Required: true},
	},
}

var SimilarToRel = &schema.RelationshipType{
	Label:       "SIMILAR_TO",
	Source:      "Customer",
	Target:      "Customer",
	Cardinality: schema.MANY_TO_MANY,
	Description: "Customers with similar behavior patterns (computed by GDS)",
	Properties: []schema.Property{
		{Name: "similarity", Type: schema.FLOAT, Description: "Similarity score 0-1"},
		{Name: "computedAt", Type: schema.DATETIME},
	},
	AgentHint: "Generated by GDS algorithms for fraud ring detection",
}

// GDS Graph Projection for fraud detection analysis
var FraudDetectionProjection = &projections.NativeProjection{
	BaseProjection: projections.BaseProjection{
		GraphName: "fraud-network",
	},
	NodeProjections: []projections.NodeProjection{
		{Label: "Customer", Properties: []string{"riskScore", "isFlagged"}},
		{Label: "Account", Properties: []string{"balance", "isBlocked"}},
		{Label: "Transaction", Properties: []string{"amount", "fraudScore", "timestamp"}},
		{Label: "Email"},
		{Label: "Phone"},
		{Label: "IPAddress", Properties: []string{"isProxy", "isTor"}},
		{Label: "Device", Properties: []string{"isEmulator"}},
		{Label: "Merchant", Properties: []string{"isFlagged"}},
	},
	RelationshipProjections: []projections.RelationshipProjection{
		{Type: "HAS_ACCOUNT"},
		{Type: "PERFORMED", Properties: []string{"timestamp"}},
		{Type: "HAS_EMAIL"},
		{Type: "HAS_PHONE"},
		{Type: "USED_IP"},
		{Type: "USED_DEVICE"},
		{Type: "TO_MERCHANT"},
	},
}

// FraudRingDetectionWCC uses Weakly Connected Components to find fraud rings.
// Customers connected through shared identifiers form potential fraud rings.
// Reference: https://neo4j.com/blog/developer/exploring-fraud-detection-neo4j-graph-data-science-part-1/
var FraudRingDetectionWCC = &algorithms.WCC{
	BaseAlgorithm: algorithms.BaseAlgorithm{
		GraphName: "fraud-network",
		Mode:      algorithms.Mutate,
	},
	MutateProperty: "fraudRingId",
}

// FraudCustomerSimilarity uses Node Similarity to find similar customers.
// Similarity based on shared identifiers and transaction patterns.
var FraudCustomerSimilarity = &algorithms.NodeSimilarity{
	BaseAlgorithm: algorithms.BaseAlgorithm{
		GraphName: "fraud-network",
		Mode:      algorithms.Mutate,
	},
	TopK:                  10,
	SimilarityCutoff:      0.3,
	DegreeCutoff:          1,
	WriteRelationshipType: "SIMILAR_TO",
	WriteProperty:         "similarity",
}

// FraudPageRank identifies influential nodes in the fraud network.
// High PageRank customers may be fraud ring leaders.
var FraudPageRank = &algorithms.PageRank{
	BaseAlgorithm: algorithms.BaseAlgorithm{
		GraphName: "fraud-network",
		Mode:      algorithms.Mutate,
	},
	DampingFactor:  0.85,
	MaxIterations:  20,
	MutateProperty: "influence",
}

// FraudCommunityDetection uses Louvain to find communities.
// Communities of customers may indicate fraud rings or legitimate customer groups.
var FraudCommunityDetection = &algorithms.Louvain{
	BaseAlgorithm: algorithms.BaseAlgorithm{
		GraphName: "fraud-network",
		Mode:      algorithms.Mutate,
	},
	MaxLevels:      10,
	MaxIterations:  10,
	Tolerance:      0.0001,
	MutateProperty: "communityId",
}
