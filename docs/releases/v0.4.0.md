# Release Notes v0.4.0

This is a breaking change release that removes the backward compatibility layer introduced in v0.3.0 and relocates constants to more appropriate packages.

## Breaking Changes

### Removed from Root Package

The following deprecated items have been removed from the root `gogithub` package:

| Removed | Replacement |
|---------|-------------|
| `Client` type | Use `auth.NewGitHubClient()` + subpackage clients |
| `NewClient(httpClient)` | Use `github.NewClient(httpClient)` directly |
| `NewClientWithToken(ctx, token)` | Use `auth.NewGitHubClient(ctx, token)` |
| `Query` type alias | Use `search.Query` |
| `Issues` type alias | Use `search.Issues` |
| `Issue` type alias | Use `search.Issue` |
| `ParamUser` constant | Use `search.ParamUser` |
| `ParamState` constant | Use `search.ParamState` |
| `ParamStateValueOpen` constant | Use `search.ParamStateValueOpen` |
| `ParamIs` constant | Use `search.ParamIs` |
| `ParamIsValuePR` constant | Use `search.ParamIsValuePR` |
| `ParamPerPageValueMax` constant | Use `search.ParamPerPageValueMax` |
| `SearchOpenPullRequests()` method | Use `search.Client.SearchOpenPullRequests()` |
| `SearchIssues()` method | Use `search.Client.SearchIssues()` |
| `SearchIssuesAll()` method | Use `search.Client.SearchIssuesAll()` |

### Relocated Constants

| Constant | Old Location | New Location |
|----------|--------------|--------------|
| `UsernameDependabot` | `search.UsernameDependabot` | `auth.UsernameDependabot` |
| `UserIDDependabot` | `search.UserIDDependabot` | `auth.UserIDDependabot` |
| `BaseURLRepoAPI` | `search.BaseURLRepoAPI` | `gogithub.BaseURLRepoAPI` |
| `BaseURLRepoHTML` | `search.BaseURLRepoHTML` | `gogithub.BaseURLRepoHTML` |

## New Features

### AuthError Now Supports Error Chains

`AuthError` now implements `Unwrap()` for Go 1.13+ error chain compatibility:

```go
// You can now use errors.Is() and errors.As() with AuthError
if errors.Is(err, someSpecificError) {
    // Handle specific underlying error
}
```

## Bug Fixes

### Query.Encode() Now Deterministic

`Query.Encode()` now sorts keys before encoding, ensuring consistent output. This fixes issues with:
- Flaky tests due to non-deterministic map iteration order
- Cache key mismatches for equivalent queries

## Migration Guide

### Before (v0.3.x)

```go
import "github.com/grokify/gogithub"

// Create client using deprecated wrapper
client := gogithub.NewClientWithToken(ctx, token)

// Use deprecated type aliases and constants
query := gogithub.Query{
    gogithub.ParamUser:  "grokify",
    gogithub.ParamState: gogithub.ParamStateValueOpen,
}

// Use deprecated wrapper method
issues, err := client.SearchIssuesAll(ctx, query, nil)

// Use bot constants from search package
if username == search.UsernameDependabot {
    // ...
}
```

### After (v0.4.0)

```go
import (
    "github.com/grokify/gogithub"
    "github.com/grokify/gogithub/auth"
    "github.com/grokify/gogithub/search"
)

// Create client using auth package
gh := auth.NewGitHubClient(ctx, token)
client := search.NewClient(gh)

// Use types and constants from search package directly
query := search.Query{
    search.ParamUser:  "grokify",
    search.ParamState: search.ParamStateValueOpen,
}

// Use search client method directly
issues, err := client.SearchIssuesAll(ctx, query, nil)

// Use bot constants from auth package
if username == auth.UsernameDependabot {
    // ...
}

// Use URL constants from root package
apiURL := gogithub.BaseURLRepoAPI + "/owner/repo"
```

## What's in the Root Package Now

The root `gogithub` package now contains only:

- Package documentation with usage examples
- `BaseURLRepoAPI` constant - GitHub API base URL for repositories
- `BaseURLRepoHTML` constant - GitHub web base URL for repositories

All functionality is in subpackages:

| Package | Purpose |
|---------|---------|
| `auth` | Authentication, bot user constants |
| `config` | Configuration, environment variables |
| `errors` | Error types and translation |
| `pathutil` | Path validation and normalization |
| `search` | Search API operations |
| `repo` | Repository operations |
| `pr` | Pull request operations |
| `release` | Release operations |
| `cliutil` | CLI utilities |
