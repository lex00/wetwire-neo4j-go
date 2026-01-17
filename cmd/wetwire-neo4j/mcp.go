// Package main provides MCP server implementation for wetwire-neo4j.
package main

import (
	"context"

	"github.com/lex00/wetwire-core-go/domain"
	domainpkg "github.com/lex00/wetwire-neo4j-go/domain"
)

// runMCPServer starts the MCP server using auto-generated configuration.
// This uses domain.BuildMCPServer() to automatically register all tools
// based on the domain interfaces implemented by Neo4jDomain.
func runMCPServer() error {
	// Create domain instance
	d := &domainpkg.Neo4jDomain{}

	// Build MCP server with auto-generated tools
	server := domain.BuildMCPServer(d)

	// Run on stdio transport
	return server.Start(context.Background())
}
