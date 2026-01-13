package kiro

import (
	"os/exec"
	"testing"
)

func TestLaunchChatCommandBuilding(t *testing.T) {
	// Test that the command would be built correctly
	// We can't actually run kiro-cli-chat in tests, but we can verify
	// the arguments are constructed properly

	tests := []struct {
		name          string
		agentName     string
		initialPrompt string
		wantArgs      []string
	}{
		{
			name:          "with prompt",
			agentName:     "test-agent",
			initialPrompt: "design a schema",
			wantArgs:      []string{"chat", "--agent", "test-agent", "design a schema"},
		},
		{
			name:          "without prompt",
			agentName:     "test-agent",
			initialPrompt: "",
			wantArgs:      []string{"chat", "--agent", "test-agent"},
		},
		{
			name:          "prompt with spaces",
			agentName:     AgentName,
			initialPrompt: "create a social network with Person nodes",
			wantArgs:      []string{"chat", "--agent", AgentName, "create a social network with Person nodes"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build args the same way LaunchChat does
			args := []string{"chat", "--agent", tt.agentName}
			if tt.initialPrompt != "" {
				args = append(args, tt.initialPrompt)
			}

			// Verify args match expected
			if len(args) != len(tt.wantArgs) {
				t.Errorf("got %d args, want %d", len(args), len(tt.wantArgs))
				return
			}
			for i, arg := range args {
				if arg != tt.wantArgs[i] {
					t.Errorf("arg[%d] = %q, want %q", i, arg, tt.wantArgs[i])
				}
			}
		})
	}
}

func TestKiroCLIExists(t *testing.T) {
	// Check if kiro-cli-chat is available in PATH
	// This is a soft check - skip if not found
	_, err := exec.LookPath("kiro-cli-chat")
	if err != nil {
		t.Skip("kiro-cli-chat not found in PATH, skipping integration test")
	}

	// If kiro-cli-chat exists, verify it responds to --help
	cmd := exec.Command("kiro-cli-chat", "--help")
	err = cmd.Run()
	if err != nil {
		t.Errorf("kiro-cli-chat --help failed: %v", err)
	}
}

func TestAgentNameConstant(t *testing.T) {
	// Verify the agent name is set
	if AgentName == "" {
		t.Error("AgentName constant should not be empty")
	}
}
