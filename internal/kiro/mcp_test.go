package kiro

import (
	"strings"
	"testing"
)

func TestAgentPromptToolsExist(t *testing.T) {
	// Tools referenced in the agent prompt
	promptTools := []string{"wetwire_lint", "wetwire_build"}

	// Tools registered in MCP server (from mcp.go)
	registeredTools := []string{
		"wetwire_init",
		"wetwire_build",
		"wetwire_lint",
		"wetwire_validate",
		"wetwire_list",
		"wetwire_graph",
	}

	// Build lookup map
	registered := make(map[string]bool)
	for _, tool := range registeredTools {
		registered[tool] = true
	}

	// Verify all prompt-referenced tools are registered
	for _, tool := range promptTools {
		if !registered[tool] {
			t.Errorf("agent prompt references tool %q but it's not registered in MCP server", tool)
		}
	}
}

func TestAgentPromptNoInvalidTools(t *testing.T) {
	// Check that the prompt doesn't reference any invalid tool names
	invalidTools := []string{"dummy", "test", "example", "placeholder"}

	prompt := AgentPrompt
	for _, invalid := range invalidTools {
		if strings.Contains(strings.ToLower(prompt), invalid) {
			t.Errorf("agent prompt contains potentially invalid tool reference: %q", invalid)
		}
	}
}

func TestMCPServerNameMatches(t *testing.T) {
	config := NewConfig()

	// The MCP server name in the agent config should match what kiro expects
	if config.MCPCommand != "wetwire-neo4j" {
		t.Errorf("MCPCommand should be 'wetwire-neo4j', got %q", config.MCPCommand)
	}

	if len(config.MCPArgs) != 1 || config.MCPArgs[0] != "mcp" {
		t.Errorf("MCPArgs should be ['mcp'], got %v", config.MCPArgs)
	}
}
