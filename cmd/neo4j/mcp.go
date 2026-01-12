// MCP server implementation for embedded design mode.
//
// When design --mcp-server is called, this runs the MCP protocol over stdio,
// providing wetwire tools for Neo4j schema and algorithm design.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/lex00/wetwire-core-go/cmd"
	"github.com/lex00/wetwire-core-go/mcp"

	"github.com/lex00/wetwire-neo4j-go/internal/cli"
	"github.com/lex00/wetwire-neo4j-go/internal/discovery"
)

// runMCPServer starts the MCP server on stdio transport.
// This is called when design --mcp-server is invoked.
func runMCPServer() error {
	server := mcp.NewServer(mcp.Config{
		Name:    "wetwire-neo4j",
		Version: version,
	})

	// Register standard wetwire tools using core infrastructure
	server.RegisterToolWithSchema("wetwire_init", "Initialize a new wetwire-neo4j project with schema and algorithm directories", handleMCPInit, initSchema)
	server.RegisterToolWithSchema("wetwire_build", "Generate Cypher or JSON output from wetwire-neo4j declarations", handleMCPBuild, buildSchema)
	server.RegisterToolWithSchema("wetwire_lint", "Check code quality and style (WN4xxx lint rules)", handleMCPLint, lintSchema)
	server.RegisterToolWithSchema("wetwire_validate", "Validate configurations against a live Neo4j instance", handleMCPValidate, validateSchema)
	server.RegisterToolWithSchema("wetwire_list", "List discovered Neo4j definitions", handleMCPList, listSchema)
	server.RegisterToolWithSchema("wetwire_graph", "Visualize resource dependencies (DOT/Mermaid)", handleMCPGraph, graphSchema)

	// Run on stdio transport
	return server.Start(context.Background())
}

// Tool input schemas (aligned with wetwire-core-go/mcp standard schemas)
var initSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"name": map[string]any{
			"type":        "string",
			"description": "Project name",
		},
		"path": map[string]any{
			"type":        "string",
			"description": "Output directory (default: current directory)",
		},
		"template": map[string]any{
			"type":        "string",
			"enum":        []string{"default", "gds", "graphrag", "full"},
			"description": "Project template (default: default)",
		},
	},
}

var buildSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"path": map[string]any{
			"type":        "string",
			"description": "Path to discover resources from",
		},
		"format": map[string]any{
			"type":        "string",
			"enum":        []string{"cypher", "json"},
			"description": "Output format (default: cypher)",
		},
	},
}

var lintSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"path": map[string]any{
			"type":        "string",
			"description": "Path to lint",
		},
	},
}

var validateSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"path": map[string]any{
			"type":        "string",
			"description": "Path to validate",
		},
		"dry_run": map[string]any{
			"type":        "boolean",
			"description": "Dry run mode (list resources without validating)",
		},
	},
}

var listSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"path": map[string]any{
			"type":        "string",
			"description": "Path to discover from",
		},
		"format": map[string]any{
			"type":        "string",
			"enum":        []string{"table", "json"},
			"description": "Output format (default: table)",
		},
	},
}

var graphSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"path": map[string]any{
			"type":        "string",
			"description": "Path to analyze",
		},
		"format": map[string]any{
			"type":        "string",
			"enum":        []string{"dot", "mermaid"},
			"description": "Output format (default: mermaid)",
		},
	},
}

// handleMCPInit initializes a new wetwire-neo4j project.
func handleMCPInit(ctx context.Context, args map[string]any) (string, error) {
	name, _ := args["name"].(string)
	path, _ := args["path"].(string)
	template, _ := args["template"].(string)

	result := MCPInitResult{}

	if path == "" {
		path = "."
	}
	if name == "" {
		name = filepath.Base(path)
		if name == "." {
			cwd, _ := os.Getwd()
			name = filepath.Base(cwd)
		}
	}
	if template == "" {
		template = "default"
	}

	// Create project directory path (name is used in path)
	projectPath := filepath.Join(path, name)
	result.Path = projectPath

	// Use the CLI initializer
	initializer := cli.NewInitializer()
	opts := cmd.InitOptions{
		Template: template,
		Force:    false,
	}
	err := initializer.Init(ctx, projectPath, opts)
	if err != nil {
		result.Error = err.Error()
		return toMCPJSON(result)
	}

	result.Success = true
	result.Files = []string{"go.mod", "schema/schema.go"}
	return toMCPJSON(result)
}

// handleMCPBuild generates Cypher or JSON from Go packages.
func handleMCPBuild(ctx context.Context, args map[string]any) (string, error) {
	path, _ := args["path"].(string)
	format, _ := args["format"].(string)

	result := MCPBuildResult{}

	if path == "" {
		path = "."
	}

	if format == "" {
		format = "cypher"
	}
	if format != "cypher" && format != "json" {
		result.Errors = append(result.Errors, fmt.Sprintf("invalid format: %s (use cypher or json)", format))
		return toMCPJSON(result)
	}

	// Capture output
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	builder := cli.NewBuilder()
	opts := cmd.BuildOptions{
		DryRun:  true,
		Verbose: false,
	}
	if format == "json" {
		opts.Output = "output.json"
	}
	err := builder.Build(ctx, path, opts)

	// Restore stdout
	_ = w.Close()
	os.Stdout = oldStdout
	_, _ = io.Copy(&buf, r)

	if err != nil {
		result.Errors = append(result.Errors, err.Error())
		return toMCPJSON(result)
	}

	result.Success = true
	result.Output = buf.String()
	return toMCPJSON(result)
}

// handleMCPLint lints Go packages for wetwire-neo4j style issues.
func handleMCPLint(ctx context.Context, args map[string]any) (string, error) {
	path, _ := args["path"].(string)

	result := MCPLintResult{}

	if path == "" {
		path = "."
	}

	linter := cli.NewLinter()
	opts := cmd.LintOptions{
		Verbose: false,
	}
	issues, err := linter.Lint(ctx, path, opts)
	if err != nil {
		result.Issues = append(result.Issues, MCPLintIssue{
			Severity: "error",
			Message:  err.Error(),
			Rule:     "internal",
		})
	}

	// Convert issues
	for _, issue := range issues {
		result.Issues = append(result.Issues, MCPLintIssue{
			Severity: issue.Severity,
			Message:  issue.Message,
			Rule:     issue.Rule,
			File:     issue.File,
			Line:     issue.Line,
		})
	}

	result.Success = len(result.Issues) == 0
	return toMCPJSON(result)
}

// handleMCPValidate validates configurations.
func handleMCPValidate(_ context.Context, args map[string]any) (string, error) {
	path, _ := args["path"].(string)
	dryRun, _ := args["dry_run"].(bool)

	result := MCPValidateResult{}

	if path == "" {
		path = "."
	}

	// Discover resources
	scanner := discovery.NewScanner()
	resources, err := scanner.ScanDir(path)
	if err != nil {
		result.Error = fmt.Sprintf("discovery failed: %v", err)
		return toMCPJSON(result)
	}

	if dryRun {
		result.Success = true
		result.Message = fmt.Sprintf("Found %d resources to validate", len(resources))
		return toMCPJSON(result)
	}

	// Without Neo4j connection, just report what we found
	result.Success = true
	result.Message = fmt.Sprintf("Discovered %d resources (no Neo4j connection for live validation)", len(resources))
	return toMCPJSON(result)
}

// handleMCPList lists discovered Neo4j definitions.
// Note: format parameter is accepted but output is always JSON for MCP transport.
func handleMCPList(_ context.Context, args map[string]any) (string, error) {
	path, _ := args["path"].(string)

	result := MCPListResult{}

	if path == "" {
		path = "."
	}

	// Discover resources
	scanner := discovery.NewScanner()
	resources, err := scanner.ScanDir(path)
	if err != nil {
		result.Error = fmt.Sprintf("discovery failed: %v", err)
		return toMCPJSON(result)
	}

	for _, res := range resources {
		result.Resources = append(result.Resources, MCPResourceInfo{
			Name: res.Name,
			Kind: string(res.Kind),
			File: res.File,
		})
	}

	result.Success = true
	return toMCPJSON(result)
}

// handleMCPGraph visualizes resource dependencies.
func handleMCPGraph(_ context.Context, args map[string]any) (string, error) {
	path, _ := args["path"].(string)
	format, _ := args["format"].(string)

	result := MCPGraphResult{}

	if path == "" {
		path = "."
	}

	if format == "" {
		format = "mermaid"
	}

	// Use GraphCLI to generate the graph
	graphCLI := cli.NewGraphCLI()
	var buf bytes.Buffer
	err := graphCLI.Generate(path, format, &buf)
	if err != nil {
		result.Error = err.Error()
		return toMCPJSON(result)
	}

	result.Success = true
	result.Graph = buf.String()
	return toMCPJSON(result)
}

// MCP Result types

// MCPInitResult is the result of the wetwire_init tool.
type MCPInitResult struct {
	Success bool     `json:"success"`
	Path    string   `json:"path"`
	Files   []string `json:"files"`
	Error   string   `json:"error,omitempty"`
}

// MCPBuildResult is the result of the wetwire_build tool.
type MCPBuildResult struct {
	Success   bool     `json:"success"`
	Output    string   `json:"output,omitempty"`
	Resources []string `json:"resources,omitempty"`
	Errors    []string `json:"errors,omitempty"`
}

// MCPLintResult is the result of the wetwire_lint tool.
type MCPLintResult struct {
	Success bool           `json:"success"`
	Output  string         `json:"output,omitempty"`
	Issues  []MCPLintIssue `json:"issues"`
}

// MCPLintIssue represents a single lint issue.
type MCPLintIssue struct {
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Rule     string `json:"rule"`
	File     string `json:"file,omitempty"`
	Line     int    `json:"line,omitempty"`
}

// MCPValidateResult is the result of the wetwire_validate tool.
type MCPValidateResult struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// MCPListResult is the result of the wetwire_list tool.
type MCPListResult struct {
	Success   bool              `json:"success"`
	Resources []MCPResourceInfo `json:"resources,omitempty"`
	Error     string            `json:"error,omitempty"`
}

// MCPResourceInfo describes a discovered resource.
type MCPResourceInfo struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
	File string `json:"file"`
}

// MCPGraphResult is the result of the wetwire_graph tool.
type MCPGraphResult struct {
	Success bool   `json:"success"`
	Graph   string `json:"graph,omitempty"`
	Error   string `json:"error,omitempty"`
}

// toMCPJSON converts a value to a JSON string.
func toMCPJSON(v any) (string, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling result: %w", err)
	}
	return string(data), nil
}
