package kiro

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallConfig_WritesCorrectFormat(t *testing.T) {
	// Use temp dir as home
	tmpHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	// Install config
	config := NewConfig()
	err := InstallConfig(config)
	if err != nil {
		t.Fatalf("InstallConfig failed: %v", err)
	}

	// Read the installed config
	agentPath := filepath.Join(tmpHome, ".kiro", "agents", AgentName+".json")
	data, err := os.ReadFile(agentPath)
	if err != nil {
		t.Fatalf("failed to read agent config: %v", err)
	}

	// Parse and verify format
	var agent map[string]any
	if err := json.Unmarshal(data, &agent); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Must have "prompt" not "systemPrompt"
	if _, ok := agent["systemPrompt"]; ok {
		t.Error("agent config should use 'prompt' not 'systemPrompt'")
	}
	if _, ok := agent["prompt"]; !ok {
		t.Error("agent config must have 'prompt' field")
	}

	// Must have name
	if agent["name"] != AgentName {
		t.Errorf("expected name %q, got %v", AgentName, agent["name"])
	}

	// Must have mcpServers
	mcpServers, ok := agent["mcpServers"].(map[string]any)
	if !ok {
		t.Fatal("agent config must have 'mcpServers' map")
	}
	if _, ok := mcpServers["wetwire-neo4j"]; !ok {
		t.Error("mcpServers must contain 'wetwire-neo4j'")
	}

	// allowedTools is optional - kiro auto-discovers from MCP server
	// Just verify it's either absent or a valid array if present
	if allowedTools, ok := agent["allowedTools"]; ok {
		if _, isArray := allowedTools.([]any); !isArray {
			t.Error("allowedTools should be an array if present")
		}
	}
}

func TestInstallConfig_AgentFileExists(t *testing.T) {
	tmpHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	config := NewConfig()
	err := InstallConfig(config)
	if err != nil {
		t.Fatalf("InstallConfig failed: %v", err)
	}

	agentPath := filepath.Join(tmpHome, ".kiro", "agents", AgentName+".json")
	if _, err := os.Stat(agentPath); os.IsNotExist(err) {
		t.Errorf("agent config file not created at %s", agentPath)
	}
}

func TestInstallConfig_UsesFullPath(t *testing.T) {
	tmpHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	// Create a fake binary in ~/go/bin
	goBin := filepath.Join(tmpHome, "go", "bin")
	if err := os.MkdirAll(goBin, 0755); err != nil {
		t.Fatalf("failed to create go/bin: %v", err)
	}
	fakeBinary := filepath.Join(goBin, "wetwire-neo4j")
	if err := os.WriteFile(fakeBinary, []byte("#!/bin/sh\necho test"), 0755); err != nil {
		t.Fatalf("failed to create fake binary: %v", err)
	}

	config := NewConfig()
	err := InstallConfig(config)
	if err != nil {
		t.Fatalf("InstallConfig failed: %v", err)
	}

	// Read and verify the config uses full path
	agentPath := filepath.Join(tmpHome, ".kiro", "agents", AgentName+".json")
	data, err := os.ReadFile(agentPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	var agent map[string]any
	if err := json.Unmarshal(data, &agent); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	mcpServers := agent["mcpServers"].(map[string]any)
	server := mcpServers["wetwire-neo4j"].(map[string]any)
	command := server["command"].(string)

	// Should be full path, not just "wetwire-neo4j"
	if command == "wetwire-neo4j" {
		t.Errorf("expected full path to binary, got bare name: %s", command)
	}
	if command != fakeBinary {
		t.Errorf("expected command %q, got %q", fakeBinary, command)
	}
}

func TestInstallConfig_MCPServerFormat(t *testing.T) {
	// Verify the MCP server config matches kiro's expected format
	tmpHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	// Create a fake binary
	goBin := filepath.Join(tmpHome, "go", "bin")
	if err := os.MkdirAll(goBin, 0755); err != nil {
		t.Fatalf("failed to create go/bin: %v", err)
	}
	fakeBinary := filepath.Join(goBin, "wetwire-neo4j")
	if err := os.WriteFile(fakeBinary, []byte("#!/bin/sh\necho test"), 0755); err != nil {
		t.Fatalf("failed to create fake binary: %v", err)
	}

	config := NewConfig()
	if err := InstallConfig(config); err != nil {
		t.Fatalf("InstallConfig failed: %v", err)
	}

	agentPath := filepath.Join(tmpHome, ".kiro", "agents", AgentName+".json")
	data, err := os.ReadFile(agentPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	var agent map[string]any
	if err := json.Unmarshal(data, &agent); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Verify mcpServers structure
	mcpServers, ok := agent["mcpServers"].(map[string]any)
	if !ok {
		t.Fatal("mcpServers must be an object")
	}

	server, ok := mcpServers["wetwire-neo4j"].(map[string]any)
	if !ok {
		t.Fatal("mcpServers must contain 'wetwire-neo4j' object")
	}

	// Must have command
	if _, ok := server["command"].(string); !ok {
		t.Error("MCP server must have 'command' as string")
	}

	// Must have args array
	args, ok := server["args"].([]any)
	if !ok {
		t.Fatal("MCP server must have 'args' as array")
	}

	// Args must contain "mcp"
	hasMCP := false
	for _, arg := range args {
		if arg == "mcp" {
			hasMCP = true
			break
		}
	}
	if !hasMCP {
		t.Errorf("MCP server args must contain 'mcp', got: %v", args)
	}
}

func TestInstallConfig_NoInvalidToolReferences(t *testing.T) {
	// Verify the config doesn't reference any invalid/placeholder tools
	tmpHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	config := NewConfig()
	if err := InstallConfig(config); err != nil {
		t.Fatalf("InstallConfig failed: %v", err)
	}

	agentPath := filepath.Join(tmpHome, ".kiro", "agents", AgentName+".json")
	data, err := os.ReadFile(agentPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	// Check raw JSON for invalid tool names
	invalidTools := []string{"dummy", "test_tool", "placeholder", "example_tool", "fake"}
	configStr := strings.ToLower(string(data))
	for _, invalid := range invalidTools {
		if strings.Contains(configStr, strings.ToLower(invalid)) {
			t.Errorf("config contains potentially invalid tool reference: %q", invalid)
		}
	}
}

func TestInstallConfig_MinimalValidStructure(t *testing.T) {
	// Test that the generated config has the minimal valid structure for kiro
	tmpHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	config := NewConfig()
	if err := InstallConfig(config); err != nil {
		t.Fatalf("InstallConfig failed: %v", err)
	}

	agentPath := filepath.Join(tmpHome, ".kiro", "agents", AgentName+".json")
	data, err := os.ReadFile(agentPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	var agent map[string]any
	if err := json.Unmarshal(data, &agent); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Required fields per kiro agent schema
	requiredFields := []string{"name", "prompt", "mcpServers"}
	for _, field := range requiredFields {
		if _, ok := agent[field]; !ok {
			t.Errorf("missing required field: %s", field)
		}
	}

	// name must be non-empty string
	name, ok := agent["name"].(string)
	if !ok || name == "" {
		t.Error("name must be a non-empty string")
	}

	// prompt must be non-empty string
	prompt, ok := agent["prompt"].(string)
	if !ok || prompt == "" {
		t.Error("prompt must be a non-empty string")
	}

	// mcpServers must be an object (not an array)
	if _, ok := agent["mcpServers"].([]any); ok {
		t.Error("mcpServers must be an object, not an array")
	}
	if _, ok := agent["mcpServers"].(map[string]any); !ok {
		t.Error("mcpServers must be an object")
	}
}

func TestInstallConfig_HasToolsArray(t *testing.T) {
	// Test that tools array is present and references the MCP server
	// This is required for kiro to enable tool usage (see GitHub issue #2640)
	tmpHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	config := NewConfig()
	if err := InstallConfig(config); err != nil {
		t.Fatalf("InstallConfig failed: %v", err)
	}

	agentPath := filepath.Join(tmpHome, ".kiro", "agents", AgentName+".json")
	data, err := os.ReadFile(agentPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	var agent map[string]any
	if err := json.Unmarshal(data, &agent); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Must have tools array
	tools, ok := agent["tools"].([]any)
	if !ok {
		t.Fatal("config must have 'tools' array")
	}

	// Must contain @wetwire-neo4j to include all MCP tools
	hasServerRef := false
	for _, tool := range tools {
		if toolStr, ok := tool.(string); ok {
			if toolStr == "@wetwire-neo4j" {
				hasServerRef = true
				break
			}
		}
	}
	if !hasServerRef {
		t.Errorf("tools array must contain '@wetwire-neo4j', got: %v", tools)
	}
}
