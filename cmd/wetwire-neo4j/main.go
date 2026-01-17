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
//	wetwire-neo4j diff         - Compare two Neo4j configurations
//	wetwire-neo4j watch        - Watch for file changes and auto-rebuild
//	wetwire-neo4j version      - Show version information
package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/lex00/wetwire-neo4j-go/domain"
	"github.com/lex00/wetwire-neo4j-go/internal/cli"
	"github.com/spf13/cobra"
)

// Version information set by goreleaser.
var (
	version = "dev"
	commit  = "none"
)

func main() {
	// Set version in domain
	domain.Version = version

	// Create domain instance
	d := &domain.Neo4jDomain{}

	// Get root command from domain
	rootCmd := domain.CreateRootCommand(d)

	// Add custom commands that aren't part of the core domain interface
	rootCmd.AddCommand(newValidateCommand())
	rootCmd.AddCommand(newDesignCmd())
	rootCmd.AddCommand(newTestCmd())
	rootCmd.AddCommand(newDiffCmd())
	rootCmd.AddCommand(newWatchCmd())
	rootCmd.AddCommand(newMCPCommand())
	rootCmd.AddCommand(newVersionCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
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
