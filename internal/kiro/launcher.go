package kiro

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	corekiro "github.com/lex00/wetwire-core-go/kiro"
)

// LaunchChat launches an interactive Kiro CLI chat session with the wetwire-neo4j agent.
// It connects stdin/stdout directly to the terminal for interactive use.
// Note: Caller should install config first via InstallConfig if custom context is needed.
func LaunchChat(agentName, initialPrompt string) error {
	// Build kiro-cli-chat command with prompt as positional argument
	// Usage: kiro-cli-chat chat --agent <AGENT> [INPUT]
	args := []string{"chat", "--agent", agentName}
	if initialPrompt != "" {
		args = append(args, initialPrompt)
	}

	cmd := exec.Command("kiro-cli-chat", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// RunTest runs the Kiro agent in non-interactive test mode.
// Returns the output from the test session.
func RunTest(ctx context.Context, prompt string) (string, error) {
	// Ensure config is installed
	if err := EnsureInstalledWithForce(true); err != nil {
		return "", fmt.Errorf("installing kiro config: %w", err)
	}

	config := NewConfig()
	result, err := corekiro.RunTest(ctx, config, prompt)
	if err != nil {
		return "", err
	}
	return result.Output, nil
}
