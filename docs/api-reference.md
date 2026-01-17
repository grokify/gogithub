# API Reference

Complete API documentation is available on pkg.go.dev. Each package is documented with function signatures, types, and examples.

## Package Documentation

| Package | Description | Documentation |
|---------|-------------|---------------|
| `gogithub` | Root package with backward-compatible exports | [pkg.go.dev](https://pkg.go.dev/github.com/grokify/gogithub) |
| `auth` | Authentication and client creation | [pkg.go.dev](https://pkg.go.dev/github.com/grokify/gogithub/auth) |
| `config` | Configuration from environment variables | [pkg.go.dev](https://pkg.go.dev/github.com/grokify/gogithub/config) |
| `search` | Search API with query builder | [pkg.go.dev](https://pkg.go.dev/github.com/grokify/gogithub/search) |
| `repo` | Repository operations | [pkg.go.dev](https://pkg.go.dev/github.com/grokify/gogithub/repo) |
| `pr` | Pull request operations | [pkg.go.dev](https://pkg.go.dev/github.com/grokify/gogithub/pr) |
| `release` | Release and asset operations | [pkg.go.dev](https://pkg.go.dev/github.com/grokify/gogithub/release) |
| `graphql` | GraphQL API for contribution stats | [pkg.go.dev](https://pkg.go.dev/github.com/grokify/gogithub/graphql) |
| `errors` | Error types and translation | [pkg.go.dev](https://pkg.go.dev/github.com/grokify/gogithub/errors) |
| `pathutil` | Path validation and normalization | [pkg.go.dev](https://pkg.go.dev/github.com/grokify/gogithub/pathutil) |
| `cliutil` | CLI utilities | [pkg.go.dev](https://pkg.go.dev/github.com/grokify/gogithub/cliutil) |

## Dependencies

GoGitHub builds on these excellent libraries:

| Library | Purpose |
|---------|---------|
| [google/go-github](https://github.com/google/go-github) | GitHub REST API client |
| [shurcooL/githubv4](https://github.com/shurcooL/githubv4) | GitHub GraphQL API client |
| [golang.org/x/oauth2](https://golang.org/x/oauth2) | OAuth2 authentication |

## Version Compatibility

| GoGitHub Version | go-github Version | Min Go Version |
|------------------|-------------------|----------------|
| v0.6.x | v81 | 1.24 |
| v0.5.x | v81 | 1.24 |
| v0.4.x | v81 | 1.22 |
| v0.3.x | v66 | 1.21 |

## Source Code

Browse the source code on GitHub: [github.com/grokify/gogithub](https://github.com/grokify/gogithub)
