# Kiro CLI Integration

Use Kiro CLI with wetwire-neo4j for AI-assisted graph schema design and GDS algorithm configuration.

## Prerequisites

- Go 1.23+ installed
- Kiro CLI installed ([installation guide](https://kiro.dev/docs/cli/installation/))
- AWS Builder ID or GitHub/Google account (for Kiro authentication)

---

## Step 1: Install wetwire-neo4j

### Option A: Using Go (recommended)

```bash
go install github.com/lex00/wetwire-neo4j-go/cmd/neo4j@latest
```

### Option B: Pre-built binaries

Download from [GitHub Releases](https://github.com/lex00/wetwire-neo4j-go/releases):

```bash
# macOS (Apple Silicon)
curl -LO https://github.com/lex00/wetwire-neo4j-go/releases/latest/download/wetwire-neo4j-darwin-arm64
chmod +x wetwire-neo4j-darwin-arm64
sudo mv wetwire-neo4j-darwin-arm64 /usr/local/bin/wetwire-neo4j

# macOS (Intel)
curl -LO https://github.com/lex00/wetwire-neo4j-go/releases/latest/download/wetwire-neo4j-darwin-amd64
chmod +x wetwire-neo4j-darwin-amd64
sudo mv wetwire-neo4j-darwin-amd64 /usr/local/bin/wetwire-neo4j

# Linux (x86-64)
curl -LO https://github.com/lex00/wetwire-neo4j-go/releases/latest/download/wetwire-neo4j-linux-amd64
chmod +x wetwire-neo4j-linux-amd64
sudo mv wetwire-neo4j-linux-amd64 /usr/local/bin/wetwire-neo4j
```

### Verify installation

```bash
wetwire-neo4j version
```

---

## Step 2: Install Kiro CLI

```bash
# Install Kiro CLI
curl -fsSL https://cli.kiro.dev/install | bash

# Verify installation
kiro-cli --version

# Sign in (opens browser)
kiro-cli login
```

---

## Step 3: Configure Kiro for wetwire-neo4j

Run the design command with `--provider kiro` to auto-configure:

```bash
# Create a project directory
mkdir my-graph-schema && cd my-graph-schema

# Initialize Go module
go mod init my-graph-schema

# Run design with Kiro provider (auto-installs configs on first run)
wetwire-neo4j design --provider kiro "Create a social network schema"
```

This automatically installs:

| File | Purpose |
|------|---------|
| `~/.kiro/agents/wetwire-neo4j-runner.json` | Kiro agent configuration |
| `.kiro/mcp.json` | Project MCP server configuration |

### Manual configuration (optional)

The MCP server is provided as a subcommand `wetwire-neo4j mcp`. If you prefer to configure manually:

**~/.kiro/agents/wetwire-neo4j-runner.json:**
```json
{
  "name": "wetwire-neo4j-runner",
  "description": "Graph schema and GDS algorithm design using wetwire-neo4j",
  "prompt": "You are a graph schema design assistant...",
  "model": "claude-sonnet-4",
  "mcpServers": {
    "wetwire": {
      "command": "wetwire-neo4j",
      "args": ["mcp"],
      "cwd": "/path/to/your/project"
    }
  },
  "tools": ["*"]
}
```

**.kiro/mcp.json:**
```json
{
  "mcpServers": {
    "wetwire": {
      "command": "wetwire-neo4j",
      "args": ["mcp"],
      "cwd": "/path/to/your/project"
    }
  }
}
```

> **Note:** The `cwd` field ensures MCP tools resolve paths correctly in your project directory. When using `wetwire-neo4j design --provider kiro`, this is configured automatically.

---

## Step 4: Run Kiro with wetwire design

### Using the wetwire-neo4j CLI

```bash
# Start Kiro design session
wetwire-neo4j design --provider kiro "Create a social network with users and friendships"
```

This launches Kiro CLI with the wetwire-neo4j-runner agent and your prompt.

### Using Kiro CLI directly

```bash
# Start chat with wetwire-neo4j-runner agent
kiro-cli chat --agent wetwire-neo4j-runner

# Or with an initial prompt
kiro-cli chat --agent wetwire-neo4j-runner "Design a recommendation graph schema"
```

---

## Available MCP Tools

The wetwire-neo4j MCP server exposes three tools to Kiro:

| Tool | Description | Example |
|------|-------------|---------|
| `wetwire_init` | Initialize a new project | `wetwire_init(path="./myschema")` |
| `wetwire_lint` | Lint code for issues | `wetwire_lint(path="./schema/...")` |
| `wetwire_build` | Generate Cypher statements and JSON configs | `wetwire_build(path="./schema/...", format="json")` |

---

## Example Session

```
$ wetwire-neo4j design --provider kiro "Create a social network schema with PageRank"

Installed Kiro agent config: ~/.kiro/agents/wetwire-neo4j-runner.json
Installed project MCP config: .kiro/mcp.json
Starting Kiro CLI design session...

> I'll help you create a social network schema with PageRank algorithm.

Let me initialize the project and create the schema definitions.

[Calling wetwire_init...]
[Calling wetwire_lint...]
[Calling wetwire_build...]

I've created the following files:
- schema/nodes.go (User node type)
- schema/relationships.go (FOLLOWS relationship)
- algorithms/pagerank.go (PageRank algorithm configuration)

The schema includes:
- User nodes with id, name, email properties
- FOLLOWS relationships for social connections
- PageRank algorithm configured to compute influence scores

Would you like me to add any additional features?
```

---

## Workflow

The Kiro agent follows this workflow:

1. **Explore** - Understand your graph schema requirements
2. **Plan** - Design the node types, relationships, and algorithms
3. **Implement** - Generate Go code using wetwire-neo4j patterns
4. **Lint** - Run `wetwire_lint` to check for issues
5. **Build** - Run `wetwire_build` to generate Cypher statements and configs

---

## Deploying Generated Schemas

After Kiro generates your schema code:

```bash
# Build Cypher statements
wetwire-neo4j build ./schema > schema.cypher

# Apply to Neo4j
cat schema.cypher | cypher-shell -u neo4j -p password

# Or build as JSON
wetwire-neo4j build ./schema --format json > config.json

# Use with Neo4j GDS pipelines
# Load the configuration into your GDS pipeline runner
```

---

## Troubleshooting

### MCP server not found

```
Mcp error: -32002: No such file or directory
```

**Solution:** Ensure `wetwire-neo4j` is in your PATH:

```bash
which wetwire-neo4j

# If not found, add to PATH or reinstall
go install github.com/lex00/wetwire-neo4j-go/cmd/neo4j@latest
```

### Kiro CLI not found

```
kiro-cli not found in PATH
```

**Solution:** Install Kiro CLI:

```bash
curl -fsSL https://cli.kiro.dev/install | bash
```

### Authentication issues

```
Error: Not authenticated
```

**Solution:** Sign in to Kiro:

```bash
kiro-cli login
```

---

## Known Limitations

### Automated Testing

When using `wetwire-neo4j test --provider kiro`, tests run in non-interactive mode (`--no-interactive`). This means:

- The agent runs autonomously without waiting for user input
- Persona simulation is limited - all personas behave similarly
- The agent won't ask clarifying questions

For true persona simulation with multi-turn conversations, use the Anthropic provider:

```bash
wetwire-neo4j test --provider anthropic --persona expert "Create a social network schema"
```

### Interactive Design Mode

Interactive design mode (`wetwire-neo4j design --provider kiro`) works fully as expected:

- Real-time conversation with the agent
- Agent can ask clarifying questions
- Lint loop executes as specified in the agent prompt

---

## See Also

- [CLI Reference](CLI.md) - Full wetwire-neo4j CLI documentation
- [Quick Start](QUICK_START.md) - Getting started with wetwire-neo4j
- [Kiro CLI Installation](https://kiro.dev/docs/cli/installation/) - Official installation guide
- [Kiro CLI Docs](https://kiro.dev/docs/cli/) - Official Kiro documentation
