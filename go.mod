module github.com/lex00/wetwire-neo4j-go

go 1.23.0

require (
	github.com/fsnotify/fsnotify v1.9.0
	github.com/lex00/wetwire-core-go v1.17.1
	github.com/neo4j/neo4j-go-driver/v5 v5.28.4
	github.com/spf13/cobra v1.10.2
)

require (
	github.com/anthropics/anthropic-sdk-go v1.19.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
	github.com/tidwall/gjson v1.18.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/sjson v1.2.5 // indirect
	golang.org/x/sys v0.34.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// Core dependency: wetwire-core-go provides shared infrastructure including:
// - mcp: MCP server implementation for tool registration and handling
// - kiro: Kiro spec generation utilities
// Future implementations should use these packages instead of custom solutions.
