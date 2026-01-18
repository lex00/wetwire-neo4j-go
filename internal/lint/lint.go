// Package lint provides lint rules for validating Neo4j GDS configurations.
//
// This package implements WN4xxx lint rules for:
// - GDS algorithm configurations (WN4001-WN4008)
// - Style enforcement (WN4010-WN4013)
// - ML pipeline configurations (WN4030-WN4035)
// - GraphRAG configurations (WN4040-WN4047)
// - Schema definitions (WN4050-WN4056)
//
// Example usage:
//
//	linter := lint.NewLinter()
//	results := linter.LintAlgorithm(pageRankConfig)
//	if lint.HasErrors(results) {
//		// Handle errors
//	}
package lint

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/lex00/wetwire-neo4j-go/internal/algorithms"
	"github.com/lex00/wetwire-neo4j-go/internal/kg"
	"github.com/lex00/wetwire-neo4j-go/internal/pipelines"
	"github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"
)

// Severity represents the severity level of a lint result.
type Severity string

const (
	// Error indicates a configuration error that must be fixed.
	Error Severity = "error"
	// Warning indicates a potential issue that should be reviewed.
	Warning Severity = "warning"
	// Info provides informational guidance.
	Info Severity = "info"
)

// LintResult represents a single lint finding.
type LintResult struct {
	// Rule is the rule identifier (e.g., "WN4001").
	Rule string
	// Severity indicates the severity level.
	Severity Severity
	// Message describes the issue.
	Message string
	// Location indicates where the issue was found.
	Location string
}

// LintOptions configures the linter behavior.
type LintOptions struct {
	// DisabledRules specifies rule IDs to skip (e.g., "WN4001", "WN4052").
	DisabledRules []string
	// Fix automatically fixes fixable issues (reserved for future use).
	Fix bool
}

// Linter validates Neo4j GDS configurations.
type Linter struct {
	pascalCaseRegex       *regexp.Regexp
	screamingRegex        *regexp.Regexp
	maxInlineProperties   int
	maxNestingDepth       int
}

// NewLinter creates a new linter with default settings.
func NewLinter() *Linter {
	return &Linter{
		// PascalCase: starts with uppercase, followed by letters/digits, must have at least one lowercase
		pascalCaseRegex:     regexp.MustCompile(`^[A-Z][a-zA-Z0-9]*[a-z][a-zA-Z0-9]*$|^[A-Z][a-z][a-zA-Z0-9]*$`),
		screamingRegex:      regexp.MustCompile(`^[A-Z][A-Z0-9]*(_[A-Z0-9]+)*$`),
		maxInlineProperties: 5,  // WN4011: warn if more than 5 inline properties
		maxNestingDepth:     3,  // WN4012: max nesting depth for schema definitions
	}
}

// WithMaxInlineProperties sets the threshold for WN4011 inline property warnings.
func (l *Linter) WithMaxInlineProperties(max int) *Linter {
	l.maxInlineProperties = max
	return l
}

// WithMaxNestingDepth sets the threshold for WN4012 nesting depth warnings.
func (l *Linter) WithMaxNestingDepth(max int) *Linter {
	l.maxNestingDepth = max
	return l
}

// LintAlgorithm validates a GDS algorithm configuration.
func (l *Linter) LintAlgorithm(algo algorithms.Algorithm) []LintResult {
	var results []LintResult

	switch a := algo.(type) {
	case *algorithms.PageRank:
		results = append(results, l.lintPageRank(a)...)
	case *algorithms.ArticleRank:
		results = append(results, l.lintArticleRank(a)...)
	case *algorithms.FastRP:
		results = append(results, l.lintFastRP(a)...)
	case *algorithms.Node2Vec:
		results = append(results, l.lintNode2Vec(a)...)
	case *algorithms.KNN:
		results = append(results, l.lintKNN(a)...)
	case *algorithms.NodeSimilarity:
		results = append(results, l.lintNodeSimilarity(a)...)
	}

	return results
}

func (l *Linter) lintPageRank(pr *algorithms.PageRank) []LintResult {
	var results []LintResult

	// WN4001: damping_factor must be in [0, 1)
	if pr.DampingFactor < 0 || pr.DampingFactor >= 1 {
		results = append(results, LintResult{
			Rule:     "WN4001",
			Severity: Error,
			Message:  fmt.Sprintf("dampingFactor must be in [0, 1), got %v", pr.DampingFactor),
			Location: "PageRank.DampingFactor",
		})
	}

	// WN4002: max_iterations must be positive
	if pr.MaxIterations < 0 {
		results = append(results, LintResult{
			Rule:     "WN4002",
			Severity: Error,
			Message:  fmt.Sprintf("maxIterations must be positive, got %d", pr.MaxIterations),
			Location: "PageRank.MaxIterations",
		})
	}

	// WN4005: warn if tolerance is too loose
	if pr.Tolerance > 1e-5 && pr.Tolerance != 0 {
		results = append(results, LintResult{
			Rule:     "WN4005",
			Severity: Warning,
			Message:  fmt.Sprintf("tolerance %v may be too loose for convergence", pr.Tolerance),
			Location: "PageRank.Tolerance",
		})
	}

	return results
}

func (l *Linter) lintArticleRank(ar *algorithms.ArticleRank) []LintResult {
	var results []LintResult

	// WN4001: damping_factor must be in [0, 1)
	if ar.DampingFactor < 0 || ar.DampingFactor >= 1 {
		results = append(results, LintResult{
			Rule:     "WN4001",
			Severity: Error,
			Message:  fmt.Sprintf("dampingFactor must be in [0, 1), got %v", ar.DampingFactor),
			Location: "ArticleRank.DampingFactor",
		})
	}

	// WN4002: max_iterations must be positive
	if ar.MaxIterations < 0 {
		results = append(results, LintResult{
			Rule:     "WN4002",
			Severity: Error,
			Message:  fmt.Sprintf("maxIterations must be positive, got %d", ar.MaxIterations),
			Location: "ArticleRank.MaxIterations",
		})
	}

	return results
}

func (l *Linter) lintFastRP(frp *algorithms.FastRP) []LintResult {
	var results []LintResult

	// WN4006: embedding_dimension should be power of 2
	if frp.EmbeddingDimension > 0 && !isPowerOfTwo(frp.EmbeddingDimension) {
		results = append(results, LintResult{
			Rule:     "WN4006",
			Severity: Warning,
			Message:  fmt.Sprintf("embeddingDimension %d is not a power of 2", frp.EmbeddingDimension),
			Location: "FastRP.EmbeddingDimension",
		})
	}

	return results
}

func (l *Linter) lintNode2Vec(n2v *algorithms.Node2Vec) []LintResult {
	var results []LintResult

	// WN4006: embedding_dimension should be power of 2
	if n2v.EmbeddingDimension > 0 && !isPowerOfTwo(n2v.EmbeddingDimension) {
		results = append(results, LintResult{
			Rule:     "WN4006",
			Severity: Warning,
			Message:  fmt.Sprintf("embeddingDimension %d is not a power of 2", n2v.EmbeddingDimension),
			Location: "Node2Vec.EmbeddingDimension",
		})
	}

	return results
}

func (l *Linter) lintKNN(knn *algorithms.KNN) []LintResult {
	var results []LintResult

	// WN4007: K warn if > 1000
	if knn.K > 1000 {
		results = append(results, LintResult{
			Rule:     "WN4007",
			Severity: Warning,
			Message:  fmt.Sprintf("K %d may cause performance issues", knn.K),
			Location: "KNN.K",
		})
	}

	return results
}

func (l *Linter) lintNodeSimilarity(ns *algorithms.NodeSimilarity) []LintResult {
	var results []LintResult

	// WN4007: topK warn if > 1000
	if ns.TopK > 1000 {
		results = append(results, LintResult{
			Rule:     "WN4007",
			Severity: Warning,
			Message:  fmt.Sprintf("topK %d may cause performance issues", ns.TopK),
			Location: "NodeSimilarity.TopK",
		})
	}

	return results
}

// LintPipeline validates an ML pipeline configuration.
func (l *Linter) LintPipeline(pipeline pipelines.Pipeline) []LintResult {
	var results []LintResult

	// WN4032: at least one model candidate required
	if len(pipeline.GetModels()) == 0 {
		results = append(results, LintResult{
			Rule:     "WN4032",
			Severity: Error,
			Message:  "pipeline must have at least one model candidate",
			Location: fmt.Sprintf("%s.Models", pipeline.PipelineName()),
		})
	}

	// Check split config for specific pipeline types
	switch p := pipeline.(type) {
	case *pipelines.NodeClassificationPipeline:
		results = append(results, l.lintSplitConfig(p.SplitConfig, p.Name)...)
	case *pipelines.LinkPredictionPipeline:
		results = append(results, l.lintSplitConfig(p.SplitConfig, p.Name)...)
	case *pipelines.NodeRegressionPipeline:
		results = append(results, l.lintSplitConfig(p.SplitConfig, p.Name)...)
	}

	return results
}

func (l *Linter) lintSplitConfig(config pipelines.SplitConfig, name string) []LintResult {
	var results []LintResult

	// WN4031: test_fraction must be < 1.0
	if config.TestFraction >= 1.0 {
		results = append(results, LintResult{
			Rule:     "WN4031",
			Severity: Error,
			Message:  fmt.Sprintf("testFraction must be < 1.0, got %v", config.TestFraction),
			Location: fmt.Sprintf("%s.SplitConfig.TestFraction", name),
		})
	}

	return results
}

// LintKGPipeline validates a KG construction pipeline.
func (l *Linter) LintKGPipeline(pipeline kg.KGPipeline) []LintResult {
	var results []LintResult

	switch p := pipeline.(type) {
	case *kg.SimpleKGPipeline:
		// WN4040: Schema must have at least one EntityType
		if len(p.EntityTypes) == 0 {
			results = append(results, LintResult{
				Rule:     "WN4040",
				Severity: Error,
				Message:  "pipeline must have at least one entity type",
				Location: fmt.Sprintf("%s.EntityTypes", p.Name),
			})
		}

		// WN4043: Entity resolver threshold should be >= 0.8
		if p.EntityResolver != nil {
			switch r := p.EntityResolver.(type) {
			case *kg.FuzzyMatchResolver:
				if r.Threshold > 0 && r.Threshold < 0.8 {
					results = append(results, LintResult{
						Rule:     "WN4043",
						Severity: Warning,
						Message:  fmt.Sprintf("fuzzy match threshold %v may be too low", r.Threshold),
						Location: fmt.Sprintf("%s.EntityResolver.Threshold", p.Name),
					})
				}
			case *kg.SemanticMatchResolver:
				if r.Threshold > 0 && r.Threshold < 0.8 {
					results = append(results, LintResult{
						Rule:     "WN4043",
						Severity: Warning,
						Message:  fmt.Sprintf("semantic match threshold %v may be too low", r.Threshold),
						Location: fmt.Sprintf("%s.EntityResolver.Threshold", p.Name),
					})
				}
			}
		}
	}

	return results
}

// LintNodeType validates a node type definition.
func (l *Linter) LintNodeType(node *schema.NodeType) []LintResult {
	return l.lintNodeTypeWithDepth(node, 0)
}

func (l *Linter) lintNodeTypeWithDepth(node *schema.NodeType, depth int) []LintResult {
	var results []LintResult

	// WN4010: Use typed NodeType, not raw structs (empty Label indicates raw struct)
	if node.Label == "" {
		results = append(results, LintResult{
			Rule:     "WN4010",
			Severity: Error,
			Message:  "NodeType must have a Label; use typed constructor instead of raw struct",
			Location: "NodeType.Label",
		})
	}

	// WN4011: Extract inline Property definitions to named vars
	if len(node.Properties) > l.maxInlineProperties {
		results = append(results, LintResult{
			Rule:     "WN4011",
			Severity: Warning,
			Message:  fmt.Sprintf("NodeType '%s' has %d inline properties; consider extracting to named variables for readability (max %d)", node.Label, len(node.Properties), l.maxInlineProperties),
			Location: fmt.Sprintf("NodeType(%s).Properties", node.Label),
		})
	}

	// WN4012: Prevent deeply nested schema definitions
	if depth > l.maxNestingDepth {
		results = append(results, LintResult{
			Rule:     "WN4012",
			Severity: Warning,
			Message:  fmt.Sprintf("NodeType '%s' nested at depth %d exceeds max depth %d", node.Label, depth, l.maxNestingDepth),
			Location: fmt.Sprintf("NodeType(%s)", node.Label),
		})
	}

	// WN4052: Node labels should be PascalCase
	if node.Label != "" && !l.pascalCaseRegex.MatchString(node.Label) {
		results = append(results, LintResult{
			Rule:     "WN4052",
			Severity: Warning,
			Message:  fmt.Sprintf("node label '%s' should be PascalCase", node.Label),
			Location: fmt.Sprintf("NodeType(%s).Label", node.Label),
		})
	}

	return results
}

// LintRelationshipType validates a relationship type definition.
func (l *Linter) LintRelationshipType(rel *schema.RelationshipType) []LintResult {
	return l.lintRelationshipTypeWithDepth(rel, 0)
}

func (l *Linter) lintRelationshipTypeWithDepth(rel *schema.RelationshipType, depth int) []LintResult {
	var results []LintResult

	// WN4010: Use typed RelationshipType, not raw structs (empty Label indicates raw struct)
	if rel.Label == "" {
		results = append(results, LintResult{
			Rule:     "WN4010",
			Severity: Error,
			Message:  "RelationshipType must have a Label; use typed constructor instead of raw struct",
			Location: "RelationshipType.Label",
		})
	}

	// WN4011: Extract inline Property definitions to named vars
	if len(rel.Properties) > l.maxInlineProperties {
		results = append(results, LintResult{
			Rule:     "WN4011",
			Severity: Warning,
			Message:  fmt.Sprintf("RelationshipType '%s' has %d inline properties; consider extracting to named variables for readability (max %d)", rel.Label, len(rel.Properties), l.maxInlineProperties),
			Location: fmt.Sprintf("RelationshipType(%s).Properties", rel.Label),
		})
	}

	// WN4012: Prevent deeply nested schema definitions
	if depth > l.maxNestingDepth {
		results = append(results, LintResult{
			Rule:     "WN4012",
			Severity: Warning,
			Message:  fmt.Sprintf("RelationshipType '%s' nested at depth %d exceeds max depth %d", rel.Label, depth, l.maxNestingDepth),
			Location: fmt.Sprintf("RelationshipType(%s)", rel.Label),
		})
	}

	// WN4013: Use direct references for relationship Source/Target
	if rel.Source == "" {
		results = append(results, LintResult{
			Rule:     "WN4013",
			Severity: Error,
			Message:  fmt.Sprintf("RelationshipType '%s' must have a Source node type reference", rel.Label),
			Location: fmt.Sprintf("RelationshipType(%s).Source", rel.Label),
		})
	}
	if rel.Target == "" {
		results = append(results, LintResult{
			Rule:     "WN4013",
			Severity: Error,
			Message:  fmt.Sprintf("RelationshipType '%s' must have a Target node type reference", rel.Label),
			Location: fmt.Sprintf("RelationshipType(%s).Target", rel.Label),
		})
	}

	// WN4053: Relationship types should be SCREAMING_SNAKE_CASE
	if rel.Label != "" && !l.screamingRegex.MatchString(rel.Label) {
		results = append(results, LintResult{
			Rule:     "WN4053",
			Severity: Warning,
			Message:  fmt.Sprintf("relationship type '%s' should be SCREAMING_SNAKE_CASE", rel.Label),
			Location: fmt.Sprintf("RelationshipType(%s).Label", rel.Label),
		})
	}

	return results
}

// LintAll validates multiple resources and returns all results.
func (l *Linter) LintAll(resources []any) []LintResult {
	return l.LintAllWithOptions(resources, LintOptions{})
}

// LintAllWithOptions validates multiple resources with configurable options.
func (l *Linter) LintAllWithOptions(resources []any, opts LintOptions) []LintResult {
	var results []LintResult

	for _, r := range resources {
		switch v := r.(type) {
		case algorithms.Algorithm:
			results = append(results, l.LintAlgorithm(v)...)
		case pipelines.Pipeline:
			results = append(results, l.LintPipeline(v)...)
		case kg.KGPipeline:
			results = append(results, l.LintKGPipeline(v)...)
		case *schema.NodeType:
			results = append(results, l.LintNodeType(v)...)
		case *schema.RelationshipType:
			results = append(results, l.LintRelationshipType(v)...)
		}
	}

	// Filter out disabled rules
	if len(opts.DisabledRules) > 0 {
		disabled := make(map[string]bool)
		for _, rule := range opts.DisabledRules {
			disabled[rule] = true
		}

		var filtered []LintResult
		for _, r := range results {
			if !disabled[r.Rule] {
				filtered = append(filtered, r)
			}
		}
		return filtered
	}

	return results
}

// HasErrors returns true if any result has Error severity.
func HasErrors(results []LintResult) bool {
	for _, r := range results {
		if r.Severity == Error {
			return true
		}
	}
	return false
}

// FilterBySeverity returns results matching the given severity.
func FilterBySeverity(results []LintResult, severity Severity) []LintResult {
	var filtered []LintResult
	for _, r := range results {
		if r.Severity == severity {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

// FormatResults formats lint results as a string.
func FormatResults(results []LintResult) string {
	if len(results) == 0 {
		return "No issues found"
	}

	var sb strings.Builder
	for _, r := range results {
		sb.WriteString(fmt.Sprintf("[%s] %s: %s (%s)\n", r.Rule, r.Severity, r.Message, r.Location))
	}
	return sb.String()
}

// isPowerOfTwo checks if n is a power of 2.
func isPowerOfTwo(n int) bool {
	return n > 0 && (n&(n-1)) == 0
}
