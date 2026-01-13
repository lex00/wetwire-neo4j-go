package kiro

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	corekiro "github.com/lex00/wetwire-core-go/kiro"
)

// kiroAgentJSON matches the kiro CLI's expected agent config format.
type kiroAgentJSON struct {
	Schema       string               `json:"$schema,omitempty"`
	Name         string               `json:"name"`
	Prompt       string               `json:"prompt"`
	MCPServers   map[string]mcpServer `json:"mcpServers,omitempty"`
	AllowedTools []string             `json:"allowedTools,omitempty"`
}

type mcpServer struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

// Available MCP tools - must match what's registered in mcp.go
var mcpTools = []string{
	"wetwire_init",
	"wetwire_build",
	"wetwire_lint",
	"wetwire_validate",
	"wetwire_list",
	"wetwire_graph",
}

// EnsureInstalled installs the Kiro agent configuration if not already present.
func EnsureInstalled() error {
	return EnsureInstalledWithForce(false)
}

// EnsureInstalledWithForce installs the Kiro agent configuration.
// If force is true, overwrites any existing configuration.
func EnsureInstalledWithForce(force bool) error {
	config := NewConfig()
	return InstallConfig(config)
}

// InstallConfig installs a specific Kiro configuration.
// This allows installing configurations with custom schema context.
func InstallConfig(config corekiro.Config) error {
	// Get kiro agents directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home dir: %w", err)
	}
	agentsDir := filepath.Join(homeDir, ".kiro", "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		return fmt.Errorf("create agents dir: %w", err)
	}

	// Build agent config with correct field names
	agent := kiroAgentJSON{
		Name:   config.AgentName,
		Prompt: config.AgentPrompt,
		MCPServers: map[string]mcpServer{
			"wetwire-neo4j": {
				Command: config.MCPCommand,
				Args:    config.MCPArgs,
			},
		},
		AllowedTools: mcpTools,
	}

	// Write agent config
	agentPath := filepath.Join(agentsDir, config.AgentName+".json")
	data, err := json.MarshalIndent(agent, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal agent config: %w", err)
	}
	if err := os.WriteFile(agentPath, data, 0644); err != nil {
		return fmt.Errorf("write agent config: %w", err)
	}

	return nil
}
