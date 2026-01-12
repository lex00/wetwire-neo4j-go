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
	"github.com/lex00/wetwire-neo4j-go/internal/kiro"
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
	var provider string

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

Providers:
  - anthropic: Uses Anthropic API directly (default)
  - kiro: Uses Kiro CLI for testing sessions

Example:
    wetwire-neo4j test --persona beginner "Create a social network schema"
    wetwire-neo4j test --provider kiro --persona expert "Set up community detection with Louvain"
    wetwire-neo4j test --all-personas "Create Person and Company nodes"`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			prompt := args[0]

			// Handle provider selection
			if provider == "kiro" {
				return runTestKiro(prompt, outputDir, personaName, allPersonas)
			}

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
	cmd.Flags().StringVar(&provider, "provider", "anthropic", "AI provider (anthropic, kiro)")

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

// runTestKiro executes tests using Kiro CLI.
func runTestKiro(prompt, outputDir, personaName string, allPersonas bool) error {
	ctx := context.Background()

	runner := kiro.NewTestRunner(outputDir)
	if err := runner.EnsureTestEnvironment(); err != nil {
		return fmt.Errorf("preparing test environment: %w", err)
	}

	if allPersonas {
		fmt.Printf("Running test with all personas using Kiro\n\n")
		results, err := runner.RunAllPersonas(ctx, prompt)
		if err != nil {
			return fmt.Errorf("running all personas: %w", err)
		}

		// Print summary
		var failed []string
		for name, result := range results {
			if result.Success {
				fmt.Printf("Persona %s: PASSED\n", name)
			} else {
				fmt.Printf("Persona %s: FAILED\n", name)
				failed = append(failed, name)
			}
		}

		fmt.Printf("\n=== Summary ===\n")
		fmt.Printf("Total: %d\n", len(results))
		fmt.Printf("Passed: %d\n", len(results)-len(failed))
		fmt.Printf("Failed: %d\n", len(failed))

		if len(failed) > 0 {
			return fmt.Errorf("%d personas failed: %v", len(failed), failed)
		}
		return nil
	}

	// Single persona test
	persona, err := personas.Get(personaName)
	if err != nil {
		return fmt.Errorf("invalid persona: %w", err)
	}

	fmt.Printf("Running test with persona '%s' using Kiro\n", personaName)
	fmt.Printf("Prompt: %s\n\n", prompt)

	result, err := runner.RunWithPersona(ctx, prompt, persona)
	if err != nil {
		return fmt.Errorf("running persona test: %w", err)
	}

	// Print summary
	fmt.Println("\n--- Test Summary ---")
	fmt.Printf("Persona: %s\n", personaName)
	fmt.Printf("Duration: %v\n", result.Duration)
	fmt.Printf("Lint passed: %v\n", result.LintPassed)
	fmt.Printf("Build passed: %v\n", result.BuildPassed)
	fmt.Printf("Files created: %d\n", len(result.FilesCreated))
	for _, f := range result.FilesCreated {
		fmt.Printf("  - %s\n", f)
	}
	if len(result.ErrorMessages) > 0 {
		fmt.Printf("Errors: %d\n", len(result.ErrorMessages))
		for _, e := range result.ErrorMessages {
			fmt.Printf("  - %s\n", e)
		}
	}
	fmt.Printf("Success: %v\n", result.Success)

	if !result.Success {
		return fmt.Errorf("test failed")
	}

	return nil
}
