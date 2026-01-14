package kiro

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	corekiro "github.com/lex00/wetwire-core-go/kiro"
)

// kiroAgentJSON matches the kiro CLI's expected agent config format.
type kiroAgentJSON struct {
	Schema       string               `json:"$schema,omitempty"`
	Name         string               `json:"name"`
	Prompt       string               `json:"prompt"`
	MCPServers   map[string]mcpServer `json:"mcpServers,omitempty"`
	Tools        []string             `json:"tools,omitempty"`
	AllowedTools []string             `json:"allowedTools,omitempty"`
}

type mcpServer struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
	Cwd     string   `json:"cwd,omitempty"`
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

	// Find the full path to the MCP command
	mcpCommand := config.MCPCommand
	if fullPath, err := exec.LookPath(config.MCPCommand); err == nil {
		mcpCommand = fullPath
	} else {
		// Try common Go binary locations
		goPath := filepath.Join(homeDir, "go", "bin", config.MCPCommand)
		if _, err := os.Stat(goPath); err == nil {
			mcpCommand = goPath
		}
	}

	// Build agent config with correct field names
	// Tools array uses @server_name format to include all tools from that MCP server
	// Cwd is required so the MCP server runs in the project directory
	agent := kiroAgentJSON{
		Name:   config.AgentName,
		Prompt: config.AgentPrompt,
		MCPServers: map[string]mcpServer{
			"wetwire-neo4j": {
				Command: mcpCommand,
				Args:    config.MCPArgs,
				Cwd:     config.WorkDir,
			},
		},
		Tools: []string{"@wetwire-neo4j"},
	}

	// Write agent config to ~/.kiro/agents/
	agentPath := filepath.Join(agentsDir, config.AgentName+".json")
	data, err := json.MarshalIndent(agent, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal agent config: %w", err)
	}
	if err := os.WriteFile(agentPath, data, 0644); err != nil {
		return fmt.Errorf("write agent config: %w", err)
	}

	// Also write project-level .kiro/mcp.json
	// Kiro reads MCP server config from the project directory, not from agent config
	if err := installProjectMCPConfig(config.WorkDir, mcpCommand, config.MCPArgs); err != nil {
		return fmt.Errorf("install project MCP config: %w", err)
	}

	return nil
}

// installProjectMCPConfig writes the .kiro/mcp.json file to the project directory.
// This is where kiro actually reads the MCP server configuration from.
func installProjectMCPConfig(projectDir, mcpCommand string, mcpArgs []string) error {
	kiroDir := filepath.Join(projectDir, ".kiro")
	if err := os.MkdirAll(kiroDir, 0755); err != nil {
		return fmt.Errorf("create .kiro dir: %w", err)
	}

	// MCP config uses mcpServers map format
	mcpConfig := map[string]any{
		"mcpServers": map[string]mcpServer{
			"wetwire-neo4j": {
				Command: mcpCommand,
				Args:    mcpArgs,
				// No cwd needed - server runs from project directory
			},
		},
	}

	data, err := json.MarshalIndent(mcpConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal mcp config: %w", err)
	}

	mcpPath := filepath.Join(kiroDir, "mcp.json")
	if err := os.WriteFile(mcpPath, data, 0644); err != nil {
		return fmt.Errorf("write mcp config: %w", err)
	}

	return nil
}
