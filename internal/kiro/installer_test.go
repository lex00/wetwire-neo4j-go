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
