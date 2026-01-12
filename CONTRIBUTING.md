# Contributing to wetwire-neo4j-go

Thank you for your interest in contributing to wetwire-neo4j-go!

## Development Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/lex00/wetwire-neo4j-go.git
   cd wetwire-neo4j-go
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Run tests:
   ```bash
   go test ./...
   ```

4. Run linting:
   ```bash
   golangci-lint run ./...
   ```

## Making Changes

1. Create a feature branch:
   ```bash
   git checkout -b feat/your-feature
   ```

2. Make your changes following TDD:
   - Write failing tests first
   - Implement the feature
   - Verify all tests pass

3. Run tests and lint:
   ```bash
   go test ./...
   golangci-lint run ./...
   ```

4. Commit your changes:
   ```bash
   git commit -m "feat: your feature description"
   ```

5. Push and create a PR:
   ```bash
   git push origin feat/your-feature
   gh pr create
   ```

## Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Pass all `golangci-lint` checks
- Add tests for new functionality
- Update documentation as needed

## Commit Messages

Use conventional commit format:
- `feat:` new features
- `fix:` bug fixes
- `docs:` documentation changes
- `test:` test additions/changes
- `refactor:` code refactoring

## Testing

- Target 80%+ code coverage for new features
- Write table-driven tests where appropriate
- Include both positive and negative test cases
- Test error handling paths

## Pull Requests

- Reference any related issues
- Include a summary of changes
- Update CHANGELOG.md for user-facing changes
- Wait for CI to pass before requesting review

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
