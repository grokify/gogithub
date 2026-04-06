# CLAUDE.md

Project-specific guidelines for gogithub.

## Project Overview

gogithub is a high-level Go module for interacting with the GitHub API. It wraps `google/go-github` with convenience functions for common operations.

## Key Packages

| Package | Purpose |
|---------|---------|
| `auth` | Authentication (OAuth2 tokens, GitHub App) |
| `checks` | CI/CD check runs and suites |
| `config` | Configuration loading from env/files |
| `errors` | Error types and translation utilities |
| `graphql` | GraphQL API client wrapper |
| `profile` | User contribution statistics |
| `pr` | Pull request operations |
| `release` | Release management |
| `repo` | Repository operations (commits, branches, batch) |
| `search` | Search API with query builder |
| `tag` | Git tag operations |
| `sarif` | SARIF upload for code scanning |

## Development Commands

```bash
# Build
go build ./...

# Test
go test ./...

# Lint (required before commits)
golangci-lint run

# Generate changelog
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
| `CHANGELOG.md` | Generated markdown changelog |
| `TASKS.md` | Refactoring tasks and future ideas |
| `mkdocs.yml` | Documentation site configuration |
| `docs/releases/` | Release notes by version |

## Dependencies

- `github.com/google/go-github/v84` - GitHub REST API client
- `github.com/shurcooL/githubv4` - GitHub GraphQL client
- `golang.org/x/oauth2` - OAuth2 authentication
- `github.com/golang-jwt/jwt/v5` - JWT for GitHub App auth
- `github.com/grokify/mogo` - Utility functions

### Updating go-github

When updating go-github (e.g., v84 to v85):
1. Update import paths in all files
2. Check for API changes in go-github release notes
3. Update `go.mod` and run `go mod tidy`

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
