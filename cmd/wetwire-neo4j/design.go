// Command design provides AI-assisted Neo4j schema and algorithm design.
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/lex00/wetwire-core-go/agent/agents"
	"github.com/lex00/wetwire-core-go/agent/orchestrator"
	"github.com/lex00/wetwire-core-go/agent/results"
	"github.com/lex00/wetwire-neo4j-go/internal/discovery"
	"github.com/lex00/wetwire-neo4j-go/internal/kiro"
	"github.com/spf13/cobra"
)

// newDesignCmd creates the "design" subcommand for AI-assisted Neo4j schema design.
// It supports Anthropic API and Kiro for interactive code generation.
func newDesignCmd() *cobra.Command {
	var outputDir string
	var maxLintCycles int
	var stream bool
	var mcpServer bool
	var provider string
	var promptFlag string

	cmd := &cobra.Command{
		Use:   "design [prompt]",
		Short: "AI-assisted Neo4j schema and algorithm design",
		Long: `Start an interactive AI-assisted session to design and generate Neo4j code.

The AI agent will:
1. Ask clarifying questions about your requirements
2. Generate Go code using wetwire-neo4j patterns
3. Run the linter and fix any issues
4. Build the Cypher queries

Providers:
  - anthropic: Uses Anthropic API directly (default)
  - kiro: Uses Kiro CLI for interactive sessions

Example:
    wetwire-neo4j design "Create a social network schema with Person nodes and KNOWS relationships"
    wetwire-neo4j design --prompt "Set up PageRank on my user graph"
    wetwire-neo4j design --provider kiro "Create a document store with vector embeddings for RAG"`,
		Args: cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			// MCP server mode - run as stdio MCP server
			if mcpServer {
				return runMCPServer()
			}

			// Get prompt from flag or positional args
			var prompt string
			if promptFlag != "" {
				prompt = promptFlag
			} else if len(args) > 0 {
				prompt = strings.Join(args, " ")
			}

			// Prompt is required for design mode
			if prompt == "" {
				return fmt.Errorf("prompt is required\n\nUsage:\n  wetwire-neo4j design \"Your prompt here\"\n  wetwire-neo4j design --prompt \"Your prompt here\"")
			}

			// Pre-flight discovery to find existing schema
			schemaContext := discoverSchemaContext(outputDir)

			// Handle provider selection
			if provider == "kiro" {
				return launchKiroWithContext(schemaContext, prompt)
			}

			return runDesign(prompt, outputDir, maxLintCycles, stream, schemaContext)
		},
	}

	cmd.Flags().StringVarP(&outputDir, "output", "o", ".", "Output directory for generated files")
	cmd.Flags().IntVarP(&maxLintCycles, "max-lint-cycles", "l", 3, "Maximum lint/fix cycles")
	cmd.Flags().BoolVarP(&stream, "stream", "s", true, "Stream AI responses")
	cmd.Flags().StringVar(&provider, "provider", "anthropic", "AI provider (anthropic, kiro)")
	cmd.Flags().StringVarP(&promptFlag, "prompt", "p", "", "Design prompt (alternative to positional argument)")
	cmd.Flags().BoolVar(&mcpServer, "mcp-server", false, "Run as MCP server on stdio")
	_ = cmd.Flags().MarkHidden("mcp-server")

	return cmd
}

// discoverSchemaContext scans the output directory for existing schema definitions
// and formats them as context for the agent prompt.
func discoverSchemaContext(dir string) string {
	scanner := discovery.NewScanner()
	resources, err := scanner.ScanDir(dir)
	if err != nil {
		// Log warning but continue without context
		fmt.Fprintf(os.Stderr, "Warning: could not scan for existing schema: %v\n", err)
		return ""
	}
	return discovery.FormatSchemaContext(resources)
}

// launchKiroWithContext launches Kiro chat with schema context injected into the config.
func launchKiroWithContext(schemaContext, prompt string) error {
	config := kiro.NewConfigWithContext(schemaContext)
	if err := kiro.InstallConfig(config); err != nil {
		return fmt.Errorf("installing Kiro config: %w", err)
	}
	return kiro.LaunchChat(kiro.AgentName, prompt)
}

// runDesign runs an interactive design session using the Anthropic API.
// It creates a runner agent that generates code, runs the linter, and fixes issues.
func runDesign(prompt, outputDir string, maxLintCycles int, stream bool, schemaContext string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nInterrupted, cleaning up...")
		cancel()
	}()

	// Create session for tracking
	session := results.NewSession("human", "design")

	// Create human developer (reads from stdin)
	reader := bufio.NewReader(os.Stdin)
	developer := orchestrator.NewHumanDeveloper(func() (string, error) {
		return reader.ReadString('\n')
	})

	// Create stream handler if streaming enabled
	var streamHandler agents.StreamHandler
	if stream {
		streamHandler = func(text string) {
			fmt.Print(text)
		}
	}

	// Create runner agent with Neo4j domain (including any discovered schema context)
	runner, err := agents.NewRunnerAgent(agents.RunnerConfig{
		WorkDir:       outputDir,
		MaxLintCycles: maxLintCycles,
		Session:       session,
		Developer:     developer,
		StreamHandler: streamHandler,
		Domain:        Neo4jDomainWithContext(schemaContext),
	})
	if err != nil {
		return fmt.Errorf("creating runner: %w", err)
	}

	fmt.Println("Starting AI-assisted Neo4j design session...")
	fmt.Println("The AI will ask questions and generate schema/algorithm code.")
	fmt.Println("Press Ctrl+C to stop.")
	fmt.Println()

	// Run the agent
	if err := runner.Run(ctx, prompt); err != nil {
		return fmt.Errorf("design session failed: %w", err)
	}

	// Print summary
	fmt.Println("\n--- Session Summary ---")
	fmt.Printf("Generated files: %d\n", len(runner.GetGeneratedFiles()))
	for _, f := range runner.GetGeneratedFiles() {
		fmt.Printf("  - %s\n", f)
	}
	fmt.Printf("Lint cycles: %d\n", runner.GetLintCycles())
	fmt.Printf("Lint passed: %v\n", runner.LintPassed())

	return nil
}
