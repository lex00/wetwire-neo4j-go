// Package main is the entry point for the wetwire-neo4j CLI.
//
// Usage:
//
//	wetwire-neo4j build        - Build Cypher/JSON from definitions
//	wetwire-neo4j lint         - Lint definitions for issues
//	wetwire-neo4j list         - List discovered definitions
//	wetwire-neo4j version      - Show version information
package main

import (
	"fmt"
	"os"

	"github.com/lex00/wetwire-core-go/cmd"
	"github.com/lex00/wetwire-neo4j-go/internal/cli"
	"github.com/spf13/cobra"
)

// Version information set by goreleaser.
var (
	version = "dev"
	commit  = "none"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	rootCmd := cmd.NewRootCommand(
		"wetwire-neo4j",
		"Wetwire CLI for Neo4j GDS schema and algorithm definitions",
	)

	// Create implementations
	builder := cli.NewBuilder()
	linter := cli.NewLinter()

	// Add commands
	rootCmd.AddCommand(cmd.NewBuildCommand(builder))
	rootCmd.AddCommand(cmd.NewLintCommand(linter))
	rootCmd.AddCommand(newListCommand())
	rootCmd.AddCommand(newVersionCommand())

	return rootCmd.Execute()
}

func newListCommand() *cobra.Command {
	var path string
	var format string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List discovered Neo4j definitions",
		Long: `List discovers and displays all Neo4j schema and algorithm definitions
in the specified path.

Supported definition types:
- NodeType: Node label definitions
- RelationshipType: Relationship type definitions
- Algorithm: GDS algorithm configurations
- Pipeline: ML pipeline configurations
- Retriever: GraphRAG retriever configurations
- KGPipeline: Knowledge graph construction pipelines`,
		RunE: func(cmd *cobra.Command, args []string) error {
			lister := cli.NewLister()
			return lister.List(path, format)
		},
	}

	cmd.Flags().StringVarP(&path, "path", "p", ".", "Path to source definitions")
	cmd.Flags().StringVarP(&format, "format", "f", "table", "Output format (table, json)")

	return cmd
}

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("wetwire-neo4j %s (commit: %s)\n", version, commit)
		},
	}
}
