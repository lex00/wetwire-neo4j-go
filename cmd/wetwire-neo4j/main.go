// Package main is the entry point for the wetwire-neo4j CLI.
//
// Usage:
//
//	wetwire-neo4j build        - Build Cypher/JSON from definitions
//	wetwire-neo4j lint         - Lint definitions for issues
//	wetwire-neo4j init         - Initialize a new project
//	wetwire-neo4j list         - List discovered definitions
//	wetwire-neo4j validate     - Validate against live Neo4j instance
//	wetwire-neo4j import       - Import schemas from Neo4j or Cypher files
//	wetwire-neo4j graph        - Visualize resource dependencies
//	wetwire-neo4j design       - AI-assisted schema and algorithm design
//	wetwire-neo4j test         - Run persona-based testing
//	wetwire-neo4j version      - Show version information
package main

import (
	"fmt"
	"os"
	"runtime/debug"

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
	initializer := cli.NewInitializer()

	// Add commands
	rootCmd.AddCommand(cmd.NewBuildCommand(builder))
	rootCmd.AddCommand(cmd.NewLintCommand(linter))
	rootCmd.AddCommand(cmd.NewInitCommand(initializer))
	rootCmd.AddCommand(newListCommand())
	rootCmd.AddCommand(newValidateCommand())
	rootCmd.AddCommand(newImportCommand())
	rootCmd.AddCommand(newGraphCommand())
	rootCmd.AddCommand(newDesignCmd())
	rootCmd.AddCommand(newTestCmd())
	rootCmd.AddCommand(newMCPCommand())
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

func newValidateCommand() *cobra.Command {
	var path string
	var uri string
	var username string
	var password string
	var database string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate configurations against a live Neo4j instance",
		Long: `Validate discovers Neo4j configurations and validates them against a live Neo4j instance.

This checks that:
- Node labels and relationship types exist in the database
- GDS algorithms are available
- Graph projections reference valid labels and types`,
		RunE: func(cmd *cobra.Command, args []string) error {
			validator := cli.NewValidatorCLI()

			if dryRun {
				return validator.ValidateDryRun(path, cmd.OutOrStdout())
			}

			config := validator.ParseConfig(uri, username, password, database)
			return validator.ValidateWithConfig(path, config, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVarP(&path, "path", "p", ".", "Path to source definitions")
	cmd.Flags().StringVar(&uri, "uri", "", "Neo4j connection URI (or $NEO4J_URI)")
	cmd.Flags().StringVar(&username, "username", "neo4j", "Neo4j username (or $NEO4J_USERNAME)")
	cmd.Flags().StringVar(&password, "password", "", "Neo4j password (or $NEO4J_PASSWORD)")
	cmd.Flags().StringVar(&database, "database", "neo4j", "Database name")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "List discovered resources without validating")

	return cmd
}

func newImportCommand() *cobra.Command {
	var file string
	var uri string
	var username string
	var password string
	var database string
	var packageName string
	var output string

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import schemas from Neo4j or Cypher files",
		Long: `Import generates Go code from existing Neo4j schemas.

Sources:
- Cypher file containing CREATE CONSTRAINT/INDEX statements
- Live Neo4j database

The generated Go code uses wetwire-neo4j-go schema types.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			importer := cli.NewImporterCLI()

			if output != "" {
				return importer.ImportToFile(file, uri, username, password, database, packageName, output)
			}

			if file != "" {
				return importer.ImportFromCypher(file, packageName, cmd.OutOrStdout())
			}

			return importer.ImportFromNeo4j(uri, username, password, database, packageName, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "Cypher file to import from")
	cmd.Flags().StringVar(&uri, "uri", "", "Neo4j connection URI (or $NEO4J_URI)")
	cmd.Flags().StringVar(&username, "username", "neo4j", "Neo4j username")
	cmd.Flags().StringVar(&password, "password", "", "Neo4j password")
	cmd.Flags().StringVar(&database, "database", "neo4j", "Database name")
	cmd.Flags().StringVar(&packageName, "package", "schema", "Go package name for generated code")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file path (default: stdout)")

	return cmd
}

func newGraphCommand() *cobra.Command {
	var path string
	var format string

	cmd := &cobra.Command{
		Use:   "graph",
		Short: "Visualize resource dependencies",
		Long: `Graph generates a dependency graph visualization of discovered resources.

Supported formats:
- dot: Graphviz DOT format (default)
- mermaid: Mermaid diagram format

The graph shows:
- Resources as nodes, colored by kind
- Dependencies as directed edges`,
		RunE: func(cmd *cobra.Command, args []string) error {
			graph := cli.NewGraphCLI()
			return graph.Generate(path, format, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVarP(&path, "path", "p", ".", "Path to source definitions")
	cmd.Flags().StringVarP(&format, "format", "f", "dot", "Output format (dot, mermaid)")

	return cmd
}

func newMCPCommand() *cobra.Command {
	return &cobra.Command{
		Use:    "mcp",
		Short:  "Run MCP server on stdio",
		Long:   `Run the Model Context Protocol server on stdio transport.`,
		Hidden: true, // Hidden from help, used by Kiro
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMCPServer()
		},
	}
}

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			v := version
			c := commit

			// Try to get version from Go module info (set by go install module@version)
			if info, ok := debug.ReadBuildInfo(); ok && v == "dev" {
				if info.Main.Version != "" && info.Main.Version != "(devel)" {
					v = info.Main.Version
				}
				// Get commit from build settings
				for _, setting := range info.Settings {
					if setting.Key == "vcs.revision" && c == "none" {
						c = setting.Value
						if len(c) > 7 {
							c = c[:7]
						}
					}
				}
			}

			fmt.Printf("wetwire-neo4j %s (commit: %s)\n", v, c)
		},
	}
}
