# Developer Guide

This guide covers development setup, testing patterns, and contribution workflows for wetwire-neo4j-go.

## Development Setup

### Prerequisites

1. **Go 1.21+**
   ```bash
   # macOS
   brew install go
   
   # Or download from https://golang.org/dl/
   ```

2. **Neo4j (for integration tests)**
   ```bash
   # Docker (recommended for development)
   docker run -d \
     --name neo4j-dev \
     -p 7474:7474 -p 7687:7687 \
     -e NEO4J_AUTH=neo4j/password \
     neo4j:5
   
   # Or download from https://neo4j.com/download/
   ```

3. **GDS Plugin (optional, for GDS algorithm testing)**
   ```bash
   # With Docker, use the GDS-enabled image
   docker run -d \
     --name neo4j-gds \
     -p 7474:7474 -p 7687:7687 \
     -e NEO4J_AUTH=neo4j/password \
     -e NEO4J_PLUGINS='["graph-data-science"]' \
     neo4j:5
   ```

4. **Development Tools**
   ```bash
   # golangci-lint for linting
   brew install golangci-lint
   
   # Or install via go
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   ```

### Clone and Install

```bash
# Clone the repository
git clone https://github.com/lex00/wetwire-neo4j-go.git
cd wetwire-neo4j-go

# Download dependencies
go mod download

# Verify installation
go build ./...
go test ./...
```

### IDE Setup

#### VSCode

Install the [Go extension](https://marketplace.visualstudio.com/items?itemName=golang.Go) and add to `.vscode/settings.json`:

```json
{
  "go.lintTool": "golangci-lint",
  "go.lintFlags": ["--fast"],
  "go.testFlags": ["-v", "-race"],
  "go.coverOnSave": true,
  "go.coverageDecorator": {
    "type": "gutter"
  },
  "[go]": {
    "editor.formatOnSave": true,
    "editor.codeActionsOnSave": {
      "source.organizeImports": "explicit"
    }
  }
}
```

#### GoLand

1. Enable `gofmt` on save: **Preferences > Tools > File Watchers > + > go fmt**
2. Configure golangci-lint: **Preferences > Tools > External Tools > + > golangci-lint**
3. Set test flags: **Run > Edit Configurations > Go Test > Add `-race -v`**

---

## Project Structure

```
wetwire-neo4j-go/
├── cmd/neo4j/              # CLI entry point
│   ├── main.go             # Main CLI setup with Cobra
│   ├── domain.go           # Domain-specific subcommands
│   ├── mcp.go              # MCP server subcommand
│   └── test.go             # Test runner subcommand
├── internal/
│   ├── algorithms/         # GDS algorithm type definitions
│   │   ├── algorithms.go   # Algorithm interfaces and types
│   │   └── serialize.go    # Cypher serialization
│   ├── cli/                # CLI command implementations
│   │   ├── builder.go      # Build command
│   │   ├── linter.go       # Lint command
│   │   └── validator.go    # Validate command
│   ├── discovery/          # AST-based resource discovery
│   ├── importer/           # Import from Neo4j/Cypher files
│   ├── kg/                 # Knowledge graph pipeline definitions
│   ├── lint/               # Lint rules (WN4xxx)
│   ├── pipelines/          # ML pipeline definitions
│   ├── projections/        # Graph projection definitions
│   ├── retrievers/         # GraphRAG retriever definitions
│   ├── serializer/         # Cypher and JSON serializers
│   └── validator/          # Neo4j instance validation
├── pkg/neo4j/schema/       # Public schema types (exported API)
│   ├── types.go            # NodeType, RelationshipType, Property
│   └── validation.go       # Schema validation utilities
├── examples/               # Reference examples
└── docs/                   # Documentation
```

### Key Interfaces

| Interface | Package | Purpose |
|-----------|---------|---------|
| `Algorithm` | `internal/algorithms` | GDS algorithm configuration |
| `Retriever` | `internal/retrievers` | GraphRAG retriever configuration |
| `Pipeline` | `internal/pipelines` | ML pipeline configuration |
| `Projection` | `internal/projections` | Graph projection configuration |
| `Resource` | `pkg/neo4j/schema` | Base interface for all schema resources |

### File Naming Conventions

| Pattern | Purpose | Example |
|---------|---------|---------|
| `<name>.go` | Main implementation | `algorithms.go` |
| `<name>_test.go` | Unit tests | `algorithms_test.go` |
| `serialize.go` | Serialization logic | `internal/algorithms/serialize.go` |
| `<name>_example.go` | Example code | `examples/schema_example.go` |

---

## Testing

### Unit Tests

Run all unit tests:

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run with race detection
go test -race ./...

# Run specific package
go test -v ./internal/lint/...

# Run specific test
go test -v -run TestLinter_WN4001 ./internal/lint/
```

### Integration Tests

Integration tests require a running Neo4j instance:

```bash
# Set environment variables
export NEO4J_URI="bolt://localhost:7687"
export NEO4J_USER="neo4j"
export NEO4J_PASSWORD="password"

# Run integration tests (tagged with //go:build integration)
go test -tags=integration ./...
```

### Test Patterns

#### Table-Driven Tests

Use table-driven tests for comprehensive coverage:

```go
func TestLinter_WN4001_DampingFactor(t *testing.T) {
    l := NewLinter()

    tests := []struct {
        name          string
        dampingFactor float64
        expectError   bool
    }{
        {"valid 0.85", 0.85, false},
        {"valid 0", 0, false},
        {"invalid 1.0", 1.0, true},
        {"invalid negative", -0.1, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            algo := &algorithms.PageRank{
                DampingFactor: tt.dampingFactor,
            }
            results := l.LintAlgorithm(algo)
            hasError := containsRule(results, "WN4001")
            if hasError != tt.expectError {
                t.Errorf("WN4001 error = %v, want %v", hasError, tt.expectError)
            }
        })
    }
}
```

#### Mock Patterns for Neo4j-Dependent Code

For code that requires Neo4j, use interface-based mocking:

```go
// Define an interface for the operations you need
type SchemaReader interface {
    FetchConstraints(ctx context.Context) ([]ConstraintDefinition, error)
    FetchIndexes(ctx context.Context) ([]IndexDefinition, error)
}

// In tests, create a mock implementation
type mockSchemaReader struct {
    constraints []ConstraintDefinition
    indexes     []IndexDefinition
    err         error
}

func (m *mockSchemaReader) FetchConstraints(ctx context.Context) ([]ConstraintDefinition, error) {
    return m.constraints, m.err
}

func (m *mockSchemaReader) FetchIndexes(ctx context.Context) ([]IndexDefinition, error) {
    return m.indexes, m.err
}
```

For testing Cypher output without a Neo4j instance, use file-based tests:

```go
func TestCypherImporter_Import(t *testing.T) {
    content := `
CREATE CONSTRAINT person_id_unique FOR (p:Person) REQUIRE p.id IS UNIQUE;
CREATE INDEX person_name_idx FOR (p:Person) ON (p.name);
`
    tmpDir := t.TempDir()
    tmpFile := filepath.Join(tmpDir, "schema.cypher")
    if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
        t.Fatalf("failed to write temp file: %v", err)
    }

    importer := NewCypherImporter(tmpFile)
    result, err := importer.Import(context.Background())
    if err != nil {
        t.Fatalf("Import failed: %v", err)
    }

    // Assert on result...
}
```

### Coverage Goals

- Target **80%+ coverage** for new features
- Critical paths (serialization, validation) should have **90%+ coverage**
- Run coverage report:

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage in browser
go tool cover -html=coverage.out

# View coverage summary
go tool cover -func=coverage.out
```

---

## Adding Features

### New Algorithm Types

1. **Add the type definition** in `internal/algorithms/algorithms.go`:

```go
// MyAlgorithm computes something useful.
type MyAlgorithm struct {
    BaseAlgorithm
    // Add algorithm-specific parameters
    MyParam int
}

func (m *MyAlgorithm) AlgorithmType() string       { return "gds.myAlgorithm" }
func (m *MyAlgorithm) AlgorithmCategory() Category { return Centrality }
```

2. **Add serialization** in `internal/algorithms/serialize.go` (or `internal/serializer/`):

```go
func serializeMyAlgorithm(algo *MyAlgorithm) string {
    // Generate Cypher statement
}
```

3. **Add lint rules** in `internal/lint/lint.go`:

```go
func (l *Linter) lintMyAlgorithm(algo *algorithms.MyAlgorithm) []LintResult {
    var results []LintResult
    
    if algo.MyParam < 0 {
        results = append(results, LintResult{
            Rule:     "WN4008",
            Severity: Error,
            Message:  "myParam must be non-negative",
            Location: "MyAlgorithm.MyParam",
        })
    }
    
    return results
}
```

4. **Add to the switch statement** in `LintAlgorithm()`:

```go
case *algorithms.MyAlgorithm:
    results = append(results, l.lintMyAlgorithm(a)...)
```

5. **Write tests** in `internal/algorithms/algorithms_test.go` and `internal/lint/lint_test.go`

6. **Add an example** in `examples/algorithms_example.go`

### New Retriever Types

1. **Add the type definition** in `internal/retrievers/retrievers.go`:

```go
// MyRetriever does something useful.
type MyRetriever struct {
    BaseRetriever
    // Add retriever-specific fields
    CustomField string
}

func (r *MyRetriever) RetrieverType() RetrieverType { return "MyRetriever" }
```

2. **Add serialization** in `internal/retrievers/serialize.go`

3. **Add tests** in `internal/retrievers/retrievers_test.go`

4. **Add an example** in `examples/retrievers_example.go`

### New Lint Rules (WN4xxx)

Lint rules follow the `WN4xxx` numbering scheme:

| Range | Category |
|-------|----------|
| WN4001-WN4029 | GDS Algorithm Rules |
| WN4030-WN4039 | ML Pipeline Rules |
| WN4040-WN4049 | GraphRAG Rules |
| WN4050-WN4059 | Schema Rules |
| WN4060-WN4069 | Projection Rules |

To add a new rule:

1. **Choose the next available number** in the appropriate range

2. **Add the rule logic** in `internal/lint/lint.go`:

```go
// WN4008: myParam must be non-negative
if algo.MyParam < 0 {
    results = append(results, LintResult{
        Rule:     "WN4008",
        Severity: Error,
        Message:  fmt.Sprintf("myParam must be non-negative, got %d", algo.MyParam),
        Location: "MyAlgorithm.MyParam",
    })
}
```

3. **Add tests** in `internal/lint/lint_test.go`:

```go
func TestLinter_WN4008_MyParam(t *testing.T) {
    l := NewLinter()

    tests := []struct {
        name        string
        myParam     int
        expectError bool
    }{
        {"valid positive", 10, false},
        {"valid zero", 0, false},
        {"invalid negative", -1, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            algo := &algorithms.MyAlgorithm{MyParam: tt.myParam}
            results := l.LintAlgorithm(algo)
            hasError := containsRule(results, "WN4008")
            if hasError != tt.expectError {
                t.Errorf("WN4008 error = %v, want %v", hasError, tt.expectError)
            }
        })
    }
}
```

4. **Document the rule** in `docs/LINT_RULES.md`

---

## Code Style

### Go Formatting

- Use `gofmt` for all Go code (enforced by CI)
- Use `goimports` to organize imports
- Follow [Effective Go](https://golang.org/doc/effective_go) guidelines

```bash
# Format all files
gofmt -w .

# Check formatting
gofmt -d .
```

### Error Handling

- Always wrap errors with context using `fmt.Errorf`:

```go
if err != nil {
    return nil, fmt.Errorf("failed to connect to Neo4j: %w", err)
}
```

- Use sentinel errors for expected conditions:

```go
var ErrNotFound = errors.New("resource not found")
```

- Check errors immediately after function calls:

```go
result, err := doSomething()
if err != nil {
    return err
}
// Use result...
```

### Documentation

- All exported types, functions, and methods must have doc comments
- Package comments go in a `doc.go` file or the main file of the package
- Use complete sentences starting with the name of the element:

```go
// PageRank computes the PageRank centrality score for nodes in a graph.
// It uses the iterative power method with configurable damping factor
// and maximum iterations.
type PageRank struct {
    // DampingFactor is the probability of following an outgoing relationship.
    // Must be in the range [0, 1). Default: 0.85.
    DampingFactor float64
}
```

---

## PR Process

### Branch Naming

Use descriptive branch names with prefixes:

| Prefix | Purpose | Example |
|--------|---------|---------|
| `feat/` | New features | `feat/graphsage-algorithm` |
| `fix/` | Bug fixes | `fix/pagerank-validation` |
| `docs/` | Documentation | `docs/developers-guide` |
| `refactor/` | Code refactoring | `refactor/lint-structure` |
| `test/` | Test additions | `test/integration-coverage` |

### Commit Messages

Use [Conventional Commits](https://www.conventionalcommits.org/) format:

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Test additions/changes
- `refactor`: Code refactoring
- `chore`: Maintenance tasks

Examples:
```
feat(algorithms): add GraphSAGE embedding algorithm

fix(lint): correct WN4001 damping factor validation

docs(readme): add installation instructions

test(serializer): add Cypher output tests for PageRank
```

### CI Requirements

All PRs must pass the following CI checks:

1. **Tests**: `go test -race -coverprofile=coverage.out ./...`
2. **Lint**: `golangci-lint run ./...`
3. **Build**: `go build -v ./...`

CI is configured in `.github/workflows/ci.yml`.

### PR Checklist

Before submitting a PR:

- [ ] All tests pass locally (`go test ./...`)
- [ ] No lint errors (`golangci-lint run ./...`)
- [ ] Code is formatted (`gofmt -d .` shows no output)
- [ ] New features have tests
- [ ] Documentation is updated
- [ ] CHANGELOG.md is updated for user-facing changes
- [ ] Commit messages follow conventional format
- [ ] PR description summarizes changes and links related issues

---

## Quick Reference

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out

# Run linter
golangci-lint run ./...

# Build CLI
go build -o neo4j ./cmd/neo4j/

# Format code
gofmt -w .

# Check for issues before committing
go test ./... && golangci-lint run ./... && gofmt -d .
```

---

## Getting Help

- **Issues**: [GitHub Issues](https://github.com/lex00/wetwire-neo4j-go/issues)
- **Documentation**: See `docs/` directory and `CLAUDE.md`
- **Examples**: See `examples/` directory
- **Core Library**: [wetwire-core-go](https://github.com/lex00/wetwire-core-go)
