# CLAUDE.md

Project-specific guidelines for gogithub.

## Project Overview

gogithub is a high-level Go module for interacting with the GitHub API. It wraps `google/go-github` with convenience functions for common operations.

## IMPORTANT: Version-Isolated Client (clientv1)

**Always use `clientv1` for new code.** The `clientv1` package provides a stable, version-isolated wrapper around go-github that protects consumers from breaking changes.

### Why This Matters

The `google/go-github` library increments major versions frequently (v88 → v89 → v90...), requiring import path changes in all consuming code.

### What Does "v1" Mean?

The "1" in `clientv1` is the **API version of the wrapper interface**, not the go-github version it wraps. Consumers stay on `clientv1` indefinitely while gogithub internally updates go-github. If we ever need breaking changes to the wrapper's interface, we'd create `clientv2`.

### Benefits

The `clientv1` package:

- Defines **stable types** (`gogithub.User`, `gogithub.Repository`, etc.) that don't change
- Provides a **Client interface** that isolates consumers from go-github version churn
- Is the **single upgrade point**: update gogithub once, all consumers benefit

### Usage Pattern

```go
// RECOMMENDED: Use clientv1 for version isolation
import (
    "github.com/grokify/gogithub"
    "github.com/grokify/gogithub/clientv1"
)

client, err := clientv1.NewClient(ctx, token)
user, err := client.GetAuthenticatedUser(ctx)      // Returns *gogithub.User
repos, err := client.ListUserRepos(ctx, "user")    // Returns []*gogithub.Repository
sha, err := client.GetBranchSHA(ctx, "owner", "repo", "main")

// AVOID: Direct go-github imports couple you to a specific version
// import "github.com/google/go-github/v89/github"  // DON'T DO THIS
```

### Type Organization

Stable types are defined in the root `gogithub` package:
- `gogithub.User`, `gogithub.Repository`, `gogithub.Reference`, etc.

The `clientv1` package provides the `Client` interface that returns these types.

### When Updating go-github

1. Update import paths in gogithub internal files only
2. Update converters in `clientv1/convert.go` if types changed
3. Run tests to ensure `clientv1` API still works
4. Consumers of `clientv1` require NO CHANGES

## Key Packages

| Package | Purpose | Stability |
|---------|---------|-----------|
| `clientv1` | **Version-isolated client wrapper** | **STABLE - use this** |
| `auth` | Authentication (OAuth2 tokens, GitHub App) | Exposes go-github types |
| `checks` | CI/CD check runs and suites | Exposes go-github types |
| `config` | Configuration loading from env/files | Exposes go-github types |
| `errors` | Error types and translation utilities | Stable |
| `graphql` | GraphQL API client wrapper | Exposes githubv4 types |
| `pathutil` | Path validation and normalization | Stable |
| `profile` | User contribution statistics | Exposes go-github types |
| `pr` | Pull request operations | Exposes go-github types |
| `release` | Release management | Exposes go-github types |
| `repo` | Repository operations (commits, branches, batch) | Exposes go-github types |
| `search` | Search API with query builder | Exposes go-github types |
| `tag` | Git tag operations | Exposes go-github types |
| `sarif` | SARIF upload for code scanning | Exposes go-github types |

**Note:** Packages marked "Exposes go-github types" will require consumer updates when go-github changes. Prefer `clientv1` when possible.

## Development Commands

```bash
# Build
go build ./...

# Test
go test ./...

# Lint (required before commits)
golangci-lint run

# Generate changelog (requires schangelog from github.com/grokify/schangelog)
schangelog generate CHANGELOG.json -o CHANGELOG.md
```

## Commit Conventions

Follow [Conventional Commits](https://www.conventionalcommits.org/):

- `feat`: New feature
- `fix`: Bug fix
- `refactor`: Code change that neither fixes a bug nor adds a feature
- `docs`: Documentation only
- `test`: Adding or updating tests
- `build`: Build system or dependencies
- `chore`: Other changes

## Release Workflow

1. Update `CHANGELOG.json` with new version entry and commit hashes
2. Run `schangelog generate CHANGELOG.json -o CHANGELOG.md`
3. Create `docs/releases/vX.Y.Z.md` release notes
4. Update `mkdocs.yml` navigation
5. Commit: `docs: add vX.Y.Z release notes and update changelog`
6. Push and wait for CI to pass
7. Tag: `git tag vX.Y.Z && git push origin vX.Y.Z`

### Versioning (SemVer)

- **PATCH** (0.0.x): Bug fixes, refactoring, docs, tests
- **MINOR** (0.x.0): New features (backwards compatible)
- **MAJOR** (x.0.0): Breaking changes

## Key Files

| File | Purpose |
|------|---------|
| `CHANGELOG.json` | Structured changelog (source of truth) |
| `CHANGELOG.md` | Generated from JSON using `schangelog` binary (do not edit manually) |
| `TASKS.md` | Refactoring tasks and future ideas |
| `mkdocs.yml` | Documentation site configuration |
| `docs/releases/` | Release notes by version |

## Dependencies

- `github.com/google/go-github/v88` - GitHub REST API client (internal use)
- `github.com/shurcooL/githubv4` - GitHub GraphQL client
- `golang.org/x/oauth2` - OAuth2 authentication
- `github.com/golang-jwt/jwt/v5` - JWT for GitHub App auth
- `github.com/grokify/mogo` - Utility functions

### Updating go-github

When updating go-github (e.g., v88 to v89):

1. Update `go.mod`: `go get github.com/google/go-github/v89`
2. Update import paths in all internal gogithub files
3. Update `clientv1/convert.go` if any type structures changed
4. Run `go test ./...` to verify
5. **Consumers using `clientv1` require NO changes**

## Code Patterns

### Pagination

Use go-github's built-in iterators (Go 1.23+):

```go
for item, err := range client.Repositories.ListReleasesIter(ctx, owner, repo, nil) {
    if err != nil {
        return nil, err
    }
    // process item
}
```

### Error Types

Each package defines domain-specific errors with `Unwrap()` for error chains:

```go
type SomeError struct {
    Context string
    Err     error
}

func (e *SomeError) Error() string { return "message: " + e.Err.Error() }
func (e *SomeError) Unwrap() error { return e.Err }
```

### Constants

Magic strings should be extracted to constants (see `repo/constants.go`, `tag/constants.go`).

## Testing

- Unit tests: `*_test.go` files in each package
- Integration tests: `*_integration_test.go` (require `GITHUB_TOKEN`)
- API-calling functions need HTTP mocking for unit tests

## Documentation Site

Built with MkDocs Material. Source in `docs/` directory.

```bash
# Local preview (requires mkdocs-material)
mkdocs serve

# Build
mkdocs build
```
