<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./wetwire-dark.svg">
  <img src="./wetwire-light.svg" width="100" height="67">
</picture>

wetwire-neo4j-go provides a command-line interface for building, linting, and managing Neo4j configuration definitions.

## Installation

See [README.md](../README.md#installation) for installation instructions.

## Commands

### build

Build Cypher statements and JSON configurations from Go definitions.

```bash
neo4j build [path] [flags]
```

**Arguments:**
- `path` - Directory or file to scan for definitions (default: current directory)

**Flags:**
- `-o, --output` - Output directory for generated files (default: stdout)
- `-f, --format` - Output format: `cypher`, `json`, or `both` (default: cypher)
- `--dry-run` - Show what would be generated without writing files

**Example:**
```bash
# Generate Cypher to stdout
neo4j build ./schemas/

# Generate JSON to a directory
neo4j build ./schemas/ -f json -o ./output/

# Generate both Cypher and JSON
neo4j build ./schemas/ -f both -o ./output/
```

**Output:**

For schema definitions:
- `CREATE CONSTRAINT` statements for uniqueness constraints
- `CREATE INDEX` statements for indexes
- Node label and relationship type configurations

For GDS algorithms:
- `CALL gds.*.stream(...)` or `CALL gds.*.mutate(...)` procedure calls

For ML pipelines:
- Pipeline creation and training Cypher statements

---

### lint

Validate definitions against wetwire lint rules (WN4xxx).

```bash
neo4j lint [path] [flags]
```

**Arguments:**
- `path` - Directory or file to lint (default: current directory)

**Flags:**
- `--fix` - Automatically fix certain issues (where possible)
- `--format` - Output format: `text`, `json` (default: text)
- `--severity` - Minimum severity to report: `error`, `warning`, `info`

**Example:**
```bash
# Lint current directory
neo4j lint

# Lint specific directory with JSON output
neo4j lint ./schemas/ --format json

# Only show errors
neo4j lint ./schemas/ --severity error
```

**Exit Codes:**
- `0` - No errors found
- `1` - Errors found
- `2` - Invalid input or configuration

See [LINT_RULES.md](LINT_RULES.md) for a complete list of lint rules.

---

### init

Initialize a new Neo4j/GDS project with scaffolded definitions.

```bash
neo4j init <project-name> [flags]
```

**Arguments:**
- `project-name` - Name of the project directory to create

**Flags:**
- `--template` - Project template: `default`, `gds`, `graphrag`, `full` (default: default)
- `--force` - Overwrite existing directory

**Templates:**
- `default` - Basic schema definitions (NodeType, RelationshipType)
- `gds` - GDS-focused project (schema + algorithms + pipelines)
- `graphrag` - GraphRAG project (schema + retrievers + kg pipelines)
- `full` - Complete project with all definition types

**Example:**
```bash
# Create a basic project
neo4j init my-graph-project

# Create a GDS analytics project
neo4j init analytics-project --template gds

# Create a GraphRAG project
neo4j init rag-project --template graphrag

# Overwrite existing directory
neo4j init existing-project --force
```

**Generated Structure:**
```
my-graph-project/
├── main.go
├── schema/
│   └── schema.go
├── algorithms/
│   └── algorithms.go  (gds, full templates)
├── pipelines/
│   └── pipelines.go   (gds, full templates)
├── retrievers/
│   └── retrievers.go  (graphrag, full templates)
└── kg/
    └── kg.go          (graphrag, full templates)
```

---

### list

List all discovered definitions in a directory.

```bash
neo4j list [path] [flags]
```

**Arguments:**
- `path` - Directory to scan (default: current directory)

**Flags:**
- `--format` - Output format: `table`, `json` (default: table)
- `--kind` - Filter by resource kind: `NodeType`, `RelationshipType`, `Algorithm`, `Pipeline`, `Retriever`

**Example:**
```bash
# List all definitions
neo4j list ./schemas/

# List only algorithms as JSON
neo4j list ./schemas/ --kind Algorithm --format json
```

**Output:**
```
NAME                KIND              FILE                  LINE
Person              NodeType          schemas/nodes.go      15
Company             NodeType          schemas/nodes.go      28
WORKS_FOR           RelationshipType  schemas/rels.go       10
InfluenceScore      Algorithm         schemas/algo.go       5
FraudDetection      Pipeline          schemas/ml.go         12
```

---

### validate

Validate configurations against a live Neo4j instance.

```bash
neo4j validate [flags]
```

**Flags:**
- `--uri` - Neo4j connection URI (default: `$NEO4J_URI` or `bolt://localhost:7687`)
- `--username` - Neo4j username (default: `$NEO4J_USERNAME` or `neo4j`)
- `--password` - Neo4j password (default: `$NEO4J_PASSWORD`)
- `--database` - Database name (default: `neo4j`)

**Example:**
```bash
# Validate against local Neo4j
neo4j validate ./schemas/

# Validate against specific instance
neo4j validate ./schemas/ --uri bolt://myhost:7687 --username admin
```

**Checks performed:**
- Node labels exist in the database
- Relationship types exist in the database
- Constraints can be applied
- Indexes can be created
- GDS algorithms are available

---

### import

Import existing Neo4j constraints and indexes to generate wetwire definitions.

```bash
neo4j import [flags]
```

**Flags:**
- `--uri` - Neo4j connection URI
- `--username` - Neo4j username
- `--password` - Neo4j password
- `--database` - Database name
- `--file` - Import from Cypher file instead of database
- `-o, --output` - Output file path (default: stdout)
- `--package` - Go package name for generated code (default: `schema`)

**Example:**
```bash
# Import from live database
neo4j import --uri bolt://localhost:7687 --password secret -o schema.go

# Import from Cypher file
neo4j import --file ./existing/schema.cypher -o schema.go
```

**Generated output:**
```go
package schema

import (
    "github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"
)

// person represents the Person node type.
var person = &schema.NodeType{
    Label: "Person",
    Properties: []schema.Property{
        {Name: "id", Type: schema.TypeString, Required: true},
    },
    Constraints: []schema.Constraint{
        {Type: schema.Unique, Properties: []string{"id"}},
    },
}
```

---

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `NEO4J_URI` | Neo4j connection URI | `bolt://localhost:7687` |
| `NEO4J_USERNAME` | Neo4j username | `neo4j` |
| `NEO4J_PASSWORD` | Neo4j password | (none) |
| `NEO4J_DATABASE` | Database name | `neo4j` |

---

## Configuration File

wetwire-neo4j-go looks for a `.wetwire.yaml` configuration file in the current directory or any parent directory.

```yaml
# .wetwire.yaml
neo4j:
  uri: bolt://localhost:7687
  username: neo4j
  database: neo4j

build:
  output: ./generated
  format: both

lint:
  severity: warning
  ignore:
    - WN4006  # Ignore embedding dimension warning
```

---

## Integration with wetwire-core-go

wetwire-neo4j-go integrates with the wetwire-core-go infrastructure:

```bash
# Using wetwire CLI
wetwire neo4j build ./schemas/
wetwire neo4j lint ./schemas/
```

The `neo4j` subcommand is automatically registered when wetwire-neo4j-go is installed.
