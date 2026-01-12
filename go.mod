module github.com/lex00/wetwire-neo4j-go

go 1.23

require (
	github.com/lex00/wetwire-core-go v0.1.0
)

// Core dependency: wetwire-core-go provides shared infrastructure including:
// - mcp: MCP server implementation for tool registration and handling
// - kiro: Kiro spec generation utilities
// Future implementations should use these packages instead of custom solutions.
