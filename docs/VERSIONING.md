<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./wetwire-dark.svg">
  <img src="./wetwire-light.svg" width="100" height="67">
</picture>

This document explains the versioning system for wetwire-neo4j-go.

## Semantic Versioning

The project follows [Semantic Versioning](https://semver.org/):

- **MAJOR** (X): Breaking API changes
- **MINOR** (Y): New features, backwards compatible
- **PATCH** (Z): Bug fixes, backwards compatible

**Location:** Git tags (e.g., `v1.0.0`)

**Detection:** Runtime via `debug.ReadBuildInfo()` or CLI `version` command

---

## Versioned Components

| Component | Tracked By |
|-----------|------------|
| **Package Version** | Git tags (vX.Y.Z) |
| **Neo4j Compatibility** | Documented in README |
| **GDS Compatibility** | Documented in README |
| **Lint Rules** | Rule codes (WN4xxx) |

### Package Version

The main version for releases. Updated for:
- New schema types or algorithm support
- Bug fixes
- Breaking changes to public API

### Neo4j Compatibility

The Neo4j versions supported are documented in README:

```markdown
| wetwire-neo4j-go | Neo4j Version | GDS Version |
|------------------|---------------|-------------|
| 1.x | 5.x | 2.x |
```

### Lint Rules

Lint rules are identified by code (WN4001, WN4052, etc.). New rules are added in minor versions. Rule behavior changes are documented in CHANGELOG.

---

## Version Resolution

The CLI determines its version using this priority:

1. **ldflags**: If built with `-ldflags "-X main.version=v1.0.0"`
2. **Build info**: If installed via `go install @version`
3. **Default**: Returns `"dev"` for local development builds

```go
func getVersion() string {
    if version != "" {
        return version
    }
    if info, ok := debug.ReadBuildInfo(); ok {
        if info.Main.Version != "" && info.Main.Version != "(devel)" {
            return info.Main.Version
        }
    }
    return "dev"
}
```

---

## Viewing Current Version

### From CLI

```bash
wetwire-neo4j version
# or
wetwire-neo4j --version
```

### From Go Code

```go
import "runtime/debug"

if info, ok := debug.ReadBuildInfo(); ok {
    fmt.Println(info.Main.Version)
}
```

---

## Bumping the Version

When releasing a new version:

1. Update CHANGELOG.md with changes

2. Run tests:
   ```bash
   go test ./...
   golangci-lint run ./...
   ```

3. Commit and tag:
   ```bash
   git commit -am "chore: release v1.1.0"
   git tag v1.1.0
   git push && git push --tags
   ```

The tag triggers the release workflow in GitHub Actions.

---

## Release Process

1. **Update CHANGELOG.md**
   - Move items from `[Unreleased]` to new version section
   - Add release date

2. **Create release commit**
   ```bash
   git add CHANGELOG.md
   git commit -m "chore: release vX.Y.Z"
   ```

3. **Tag the release**
   ```bash
   git tag vX.Y.Z
   git push origin main --tags
   ```

4. **GitHub Actions** automatically:
   - Builds binaries for multiple platforms
   - Creates GitHub release
   - Users can install via `go install @vX.Y.Z`

---

## Compatibility Matrix

| wetwire-neo4j-go | Go Version | Neo4j | GDS | wetwire-core-go |
|------------------|------------|-------|-----|-----------------|
| 1.x | 1.21+ | 5.x | 2.x | 1.x |

---

## See Also

- [Developer Guide](DEVELOPERS.md) - Full development guide
- [CHANGELOG.md](../CHANGELOG.md) - Release history
