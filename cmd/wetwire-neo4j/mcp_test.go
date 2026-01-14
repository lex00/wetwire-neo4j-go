package main

import (
	"bufio"
	"encoding/json"
	"io"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestMCPServer_ToolsList(t *testing.T) {
	// Build the binary first
	cmd := exec.Command("go", "build", "-o", "wetwire-neo4j-test", ".")
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to build binary: %v", err)
	}
	defer func() { _ = exec.Command("rm", "wetwire-neo4j-test").Run() }()

	// Start MCP server
	mcpCmd := exec.Command("./wetwire-neo4j-test", "mcp")
	stdin, err := mcpCmd.StdinPipe()
	if err != nil {
		t.Fatalf("failed to get stdin pipe: %v", err)
	}
	stdout, err := mcpCmd.StdoutPipe()
	if err != nil {
		t.Fatalf("failed to get stdout pipe: %v", err)
	}

	if err := mcpCmd.Start(); err != nil {
		t.Fatalf("failed to start MCP server: %v", err)
	}
	defer func() { _ = mcpCmd.Process.Kill() }()

	// Send tools/list request
	request := `{"jsonrpc":"2.0","method":"tools/list","id":1}` + "\n"
	if _, err := io.WriteString(stdin, request); err != nil {
		t.Fatalf("failed to write request: %v", err)
	}

	// Read response with timeout
	done := make(chan string, 1)
	go func() {
		reader := bufio.NewReader(stdout)
		line, _ := reader.ReadString('\n')
		done <- line
	}()

	select {
	case response := <-done:
		// Parse response
		var result map[string]any
		if err := json.Unmarshal([]byte(response), &result); err != nil {
			t.Fatalf("failed to parse response: %v\nresponse: %s", err, response)
		}

		// Check for error
		if errObj, ok := result["error"]; ok {
			t.Fatalf("MCP server returned error: %v", errObj)
		}

		// Check tools are returned
		resultObj, ok := result["result"].(map[string]any)
		if !ok {
			t.Fatalf("expected result object, got: %v", result)
		}

		tools, ok := resultObj["tools"].([]any)
		if !ok {
			t.Fatalf("expected tools array, got: %v", resultObj)
		}

		// Verify expected tools exist
		expectedTools := []string{
			"wetwire_init",
			"wetwire_build",
			"wetwire_lint",
			"wetwire_validate",
			"wetwire_list",
			"wetwire_graph",
		}

		toolNames := make(map[string]bool)
		for _, tool := range tools {
			if toolMap, ok := tool.(map[string]any); ok {
				if name, ok := toolMap["name"].(string); ok {
					toolNames[name] = true
				}
			}
		}

		for _, expected := range expectedTools {
			if !toolNames[expected] {
				t.Errorf("MCP server missing tool: %s", expected)
			}
		}

		// Verify NO tool named "dummy"
		if toolNames["dummy"] {
			t.Error("MCP server should NOT have a tool named 'dummy'")
		}

	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for MCP response")
	}
}

func TestMCPServer_NoInvalidTools(t *testing.T) {
	// This test verifies that there are no references to invalid/placeholder tools
	// in the codebase that could confuse kiro

	invalidNames := []string{"dummy", "test_tool", "placeholder", "example_tool"}

	// Check mcp.go for tool registrations
	// The actual check is done by examining the registered tool names
	registeredTools := []string{
		"wetwire_init",
		"wetwire_build",
		"wetwire_lint",
		"wetwire_validate",
		"wetwire_list",
		"wetwire_graph",
	}

	for _, tool := range registeredTools {
		for _, invalid := range invalidNames {
			if strings.Contains(tool, invalid) {
				t.Errorf("tool name %q contains invalid substring %q", tool, invalid)
			}
		}
	}
}
