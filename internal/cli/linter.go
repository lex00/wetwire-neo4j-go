package cli

import (
	"context"
	"fmt"

	"github.com/lex00/wetwire-core-go/cmd"
	"github.com/lex00/wetwire-neo4j-go/internal/algorithms"
	"github.com/lex00/wetwire-neo4j-go/internal/discovery"
	"github.com/lex00/wetwire-neo4j-go/internal/kg"
	"github.com/lex00/wetwire-neo4j-go/internal/lint"
	"github.com/lex00/wetwire-neo4j-go/internal/pipelines"
	"github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"
)

// Linter implements the cmd.Linter interface for Neo4j definitions.
type Linter struct {
	scanner *discovery.Scanner
	linter  *lint.Linter
}

// NewLinter creates a new Linter.
func NewLinter() *Linter {
	return &Linter{
		scanner: discovery.NewScanner(),
		linter:  lint.NewLinter(),
	}
}

// Lint implements cmd.Linter.Lint.
func (l *Linter) Lint(ctx context.Context, path string, opts cmd.LintOptions) ([]cmd.Issue, error) {
	// Discover resources
	resources, err := l.scanner.ScanDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to scan directory: %w", err)
	}

	if len(resources) == 0 {
		return nil, nil
	}

	var issues []cmd.Issue

	// For each discovered resource, generate structural lint issues
	for _, r := range resources {
		resourceIssues := l.lintResource(r)
		issues = append(issues, resourceIssues...)
	}

	return issues, nil
}

// lintResource generates lint issues for a discovered resource.
func (l *Linter) lintResource(r discovery.DiscoveredResource) []cmd.Issue {
	var issues []cmd.Issue

	// Add basic structural checks based on resource type
	switch r.Kind {
	case discovery.KindNodeType:
		// Check naming convention
		if !isPascalCase(r.Name) {
			issues = append(issues, cmd.Issue{
				File:     r.File,
				Line:     r.Line,
				Column:   1,
				Severity: "warning",
				Message:  fmt.Sprintf("node type '%s' should use PascalCase naming", r.Name),
				Rule:     "WN4052",
			})
		}
	case discovery.KindRelationshipType:
		// Check naming convention (should be SCREAMING_SNAKE_CASE)
		if !isScreamingSnakeCase(r.Name) {
			issues = append(issues, cmd.Issue{
				File:     r.File,
				Line:     r.Line,
				Column:   1,
				Severity: "warning",
				Message:  fmt.Sprintf("relationship type '%s' should use SCREAMING_SNAKE_CASE naming", r.Name),
				Rule:     "WN4053",
			})
		}
	}

	return issues
}

// LintAlgorithm lints an algorithm configuration.
func (l *Linter) LintAlgorithm(algo algorithms.Algorithm, file string, line int) []cmd.Issue {
	results := l.linter.LintAlgorithm(algo)
	return l.convertResults(results, file, line)
}

// LintPipeline lints a pipeline configuration.
func (l *Linter) LintPipeline(pipe pipelines.Pipeline, file string, line int) []cmd.Issue {
	results := l.linter.LintPipeline(pipe)
	return l.convertResults(results, file, line)
}

// LintKGPipeline lints a KG pipeline configuration.
func (l *Linter) LintKGPipeline(pipe kg.KGPipeline, file string, line int) []cmd.Issue {
	results := l.linter.LintKGPipeline(pipe)
	return l.convertResults(results, file, line)
}

// LintNodeType lints a node type definition.
func (l *Linter) LintNodeType(node *schema.NodeType, file string, line int) []cmd.Issue {
	results := l.linter.LintNodeType(node)
	return l.convertResults(results, file, line)
}

// LintRelationshipType lints a relationship type definition.
func (l *Linter) LintRelationshipType(rel *schema.RelationshipType, file string, line int) []cmd.Issue {
	results := l.linter.LintRelationshipType(rel)
	return l.convertResults(results, file, line)
}

// LintAll lints multiple resources.
func (l *Linter) LintAll(resources []any, file string, line int) []cmd.Issue {
	results := l.linter.LintAll(resources)
	return l.convertResults(results, file, line)
}

// convertResults converts lint.LintResult to cmd.Issue.
func (l *Linter) convertResults(results []lint.LintResult, file string, line int) []cmd.Issue {
	issues := make([]cmd.Issue, len(results))
	for i, r := range results {
		issues[i] = cmd.Issue{
			File:     file,
			Line:     line,
			Column:   1,
			Severity: string(r.Severity),
			Message:  r.Message,
			Rule:     r.Rule,
		}
	}
	return issues
}

// isPascalCase checks if a string is PascalCase.
func isPascalCase(s string) bool {
	if len(s) == 0 {
		return false
	}
	// Must start with uppercase
	if s[0] < 'A' || s[0] > 'Z' {
		return false
	}
	// Must contain at least one lowercase letter
	hasLower := false
	for _, c := range s {
		if c >= 'a' && c <= 'z' {
			hasLower = true
			break
		}
	}
	return hasLower
}

// isScreamingSnakeCase checks if a string is SCREAMING_SNAKE_CASE.
func isScreamingSnakeCase(s string) bool {
	if len(s) == 0 {
		return false
	}
	// First character must be uppercase letter
	if s[0] < 'A' || s[0] > 'Z' {
		return false
	}
	for _, c := range s {
		// Must be uppercase letter, digit, or underscore
		isUppercase := c >= 'A' && c <= 'Z'
		isDigit := c >= '0' && c <= '9'
		isUnderscore := c == '_'
		if !isUppercase && !isDigit && !isUnderscore {
			return false
		}
	}
	return true
}
