// Command test runs automated persona-based testing for Neo4j code generation.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/lex00/wetwire-core-go/agent/agents"
	"github.com/lex00/wetwire-core-go/agent/orchestrator"
	"github.com/lex00/wetwire-core-go/agent/personas"
	"github.com/lex00/wetwire-core-go/agent/results"
	"github.com/spf13/cobra"
)

// newTestCmd creates the "test" subcommand for automated persona-based testing.
// It runs AI agents with different personas to evaluate code generation quality.
func newTestCmd() *cobra.Command {
	var outputDir string
	var personaName string
	var scenario string
	var maxLintCycles int
	var stream bool
	var allPersonas bool

	cmd := &cobra.Command{
		Use:   "test [prompt]",
		Short: "Run automated persona-based testing",
		Long: `Run automated testing with AI personas to evaluate code generation quality.

Available personas:
  - beginner: New to Neo4j/graph databases, asks many clarifying questions
  - intermediate: Familiar with Neo4j basics, asks targeted questions
  - expert: Deep Neo4j/GDS knowledge, asks advanced questions
  - terse: Gives minimal responses
  - verbose: Provides detailed context

Example:
    wetwire-neo4j test --persona beginner "Create a social network schema"
    wetwire-neo4j test --persona expert "Set up community detection with Louvain"
    wetwire-neo4j test --all-personas "Create Person and Company nodes"`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			prompt := args[0]
			if allPersonas {
				return runTestAllPersonas(prompt, outputDir, scenario, maxLintCycles, stream)
			}
			return runTest(prompt, outputDir, personaName, scenario, maxLintCycles, stream)
		},
	}

	cmd.Flags().StringVarP(&outputDir, "output", "o", ".", "Output directory for generated files")
	cmd.Flags().StringVarP(&personaName, "persona", "p", "intermediate", "Persona to use (beginner, intermediate, expert, terse, verbose)")
	cmd.Flags().StringVarP(&scenario, "scenario", "S", "default", "Scenario name for tracking")
	cmd.Flags().IntVarP(&maxLintCycles, "max-lint-cycles", "l", 3, "Maximum lint/fix cycles")
	cmd.Flags().BoolVarP(&stream, "stream", "s", false, "Stream AI responses")
	cmd.Flags().BoolVar(&allPersonas, "all-personas", false, "Run test with all personas")

	return cmd
}

// runTest executes a single persona test using the Anthropic API.
func runTest(prompt, outputDir, personaName, scenario string, maxLintCycles int, stream bool) error {
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

	// Get persona
	persona, err := personas.Get(personaName)
	if err != nil {
		return fmt.Errorf("invalid persona: %w", err)
	}

	// Create session for tracking
	session := results.NewSession(personaName, scenario)

	// Create AI developer with persona
	responder := agents.CreateDeveloperResponder("")
	developer := orchestrator.NewAIDeveloper(persona, responder)

	// Create stream handler if streaming enabled
	var streamHandler agents.StreamHandler
	if stream {
		streamHandler = func(text string) {
			fmt.Print(text)
		}
	}

	// Create runner agent with Neo4j domain
	runner, err := agents.NewRunnerAgent(agents.RunnerConfig{
		WorkDir:       outputDir,
		MaxLintCycles: maxLintCycles,
		Session:       session,
		Developer:     developer,
		StreamHandler: streamHandler,
		Domain:        Neo4jDomain(),
	})
	if err != nil {
		return fmt.Errorf("creating runner: %w", err)
	}

	fmt.Printf("Running test with persona '%s' and scenario '%s'\n", personaName, scenario)
	fmt.Printf("Prompt: %s\n\n", prompt)

	// Run the agent
	if err := runner.Run(ctx, prompt); err != nil {
		return fmt.Errorf("test failed: %w", err)
	}

	// Complete session
	session.Complete()

	// Write results
	writer := results.NewWriter(outputDir)
	if err := writer.Write(session); err != nil {
		fmt.Printf("Warning: failed to write results: %v\n", err)
	} else {
		fmt.Printf("\nResults written to: %s\n", outputDir)
	}

	// Print summary
	fmt.Println("\n--- Test Summary ---")
	fmt.Printf("Persona: %s\n", personaName)
	fmt.Printf("Scenario: %s\n", scenario)
	fmt.Printf("Generated files: %d\n", len(runner.GetGeneratedFiles()))
	for _, f := range runner.GetGeneratedFiles() {
		fmt.Printf("  - %s\n", f)
	}
	fmt.Printf("Lint cycles: %d\n", runner.GetLintCycles())
	fmt.Printf("Lint passed: %v\n", runner.LintPassed())
	fmt.Printf("Questions asked: %d\n", len(session.Questions))

	return nil
}

// runTestAllPersonas runs the test with all available personas sequentially.
// It aggregates results and reports which personas passed or failed.
func runTestAllPersonas(prompt, outputDir, scenario string, maxLintCycles int, stream bool) error {
	personaNames := personas.Names()
	var failed []string

	fmt.Printf("Running tests with all %d personas\n\n", len(personaNames))

	for _, personaName := range personaNames {
		// Create persona-specific output directory
		personaOutputDir := fmt.Sprintf("%s/%s", outputDir, personaName)

		fmt.Printf("=== Running persona: %s ===\n", personaName)

		err := runTest(prompt, personaOutputDir, personaName, scenario, maxLintCycles, stream)
		if err != nil {
			fmt.Printf("Persona %s: FAILED - %v\n\n", personaName, err)
			failed = append(failed, personaName)
		} else {
			fmt.Printf("Persona %s: PASSED\n\n", personaName)
		}
	}

	// Print summary
	fmt.Println("\n=== All Personas Summary ===")
	fmt.Printf("Total: %d\n", len(personaNames))
	fmt.Printf("Passed: %d\n", len(personaNames)-len(failed))
	fmt.Printf("Failed: %d\n", len(failed))
	if len(failed) > 0 {
		fmt.Printf("Failed personas: %v\n", failed)
		return fmt.Errorf("%d personas failed", len(failed))
	}

	return nil
}
