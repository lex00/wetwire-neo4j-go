package kiro

import (
	"encoding/json"
	"os"
	"path/filepath"
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

	// Must have allowedTools listing MCP tools
	allowedTools, ok := agent["allowedTools"].([]any)
	if !ok {
		t.Fatal("agent config must have 'allowedTools' array")
	}
	if len(allowedTools) == 0 {
		t.Error("allowedTools should not be empty")
	}
	// Verify expected tools are listed
	toolSet := make(map[string]bool)
	for _, tool := range allowedTools {
		if s, ok := tool.(string); ok {
			toolSet[s] = true
		}
	}
	expectedTools := []string{"wetwire-neo4j:wetwire_init", "wetwire-neo4j:wetwire_build", "wetwire-neo4j:wetwire_lint"}
	for _, expected := range expectedTools {
		if !toolSet[expected] {
			t.Errorf("allowedTools should contain %q", expected)
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
