# Release Notes v0.3.0

This is a major release that restructures the `gogithub` package into modular subpackages organized by operation type, significantly expanding functionality while maintaining backward compatibility.

## Highlights

- **Modular architecture**: Reorganized from a flat package into focused subpackages (`auth`, `search`, `repo`, `pr`, `release`, etc.)
- **New functionality**: Added comprehensive support for repository operations, pull requests, releases, and batch commits
- **Error handling**: New `errors` package with standard error types and translation utilities
- **Configuration**: New `config` package with environment variable support and GitHub Enterprise support
- **Updated dependencies**: go-github upgraded from v68 to v81

## New Packages

### `auth/` - Authentication Utilities

- `NewGitHubClient(ctx, token)` - Create authenticated GitHub client
- `NewTokenClient(ctx, token)` - Create authenticated HTTP client
- `GetAuthenticatedUser(ctx, client)` - Get current user's login
- `GetUser(ctx, client, username)` - Get user information
- `AuthError` - Authentication error type

### `search/` - Search API

- `Client` - Search client with pagination support
- `SearchIssues` / `SearchIssuesAll` - Search issues with automatic pagination
- `SearchOpenPullRequests` - Convenience method for PR search
- `Query` - Query builder with parameter constants
- `Issues` / `Issue` - Result types with table generation support

### `repo/` - Repository Operations

**Fork and branch operations (`fork.go`, `branch.go`):**

- `EnsureFork` - Ensure a fork exists, creating if needed
- `GetDefaultBranch` - Get repository's default branch
- `CreateBranch` / `DeleteBranch` / `BranchExists` - Branch management
- `GetBranchSHA` - Get SHA of a branch

**Commit operations (`commit.go`):**

- `CreateCommit` - Create commits using Git tree API
- `ReadLocalFiles` - Read local files for commit
- `FileContent` - File content type for commits

**Listing operations (`list.go`):**

- `ListOrgRepos` / `ListUserRepos` - List repositories with pagination
- `GetRepo` - Get repository details
- `ParseRepoName` - Parse "owner/repo" format

**Batch operations (`batch.go`):**

- `Batch` - Accumulate multiple file operations for atomic commits
- `NewBatch` - Create batch with options
- `Write` / `Delete` - Queue file operations
- `Commit` - Apply all operations in a single commit
- `WithCommitAuthor` - Set commit author option

### `pr/` - Pull Request Operations

- `CreatePR` - Create pull requests
- `GetPR` / `ListPRs` - Retrieve pull requests
- `MergePR` / `ClosePR` - Merge or close PRs
- `AddPRReviewers` - Request reviewers
- `ListPRFiles` / `ListPRComments` - List PR files and comments
- `PRError` - PR operation error type

### `release/` - Release Operations

- `ListReleases` / `ListReleasesSince` - List releases with pagination
- `GetRelease` / `GetLatestRelease` / `GetReleaseByTag` - Get release details
- `ListReleaseAssets` - List release assets

### `errors/` - Error Types and Translation

- Standard errors: `ErrNotFound`, `ErrPermissionDenied`, `ErrRateLimited`, `ErrConflict`, `ErrValidation`, `ErrServerError`
- `APIError` - Wrapper with status code and context
- `Translate` - Convert GitHub API errors to standard errors
- Helper functions: `IsNotFound`, `IsPermissionDenied`, `IsRateLimited`, etc.
- `StatusCode` - Extract HTTP status from errors

### `config/` - Configuration Utilities

- `Config` - Configuration struct with validation
- `FromEnv` / `FromEnvWithConfig` - Load from environment variables
- `FromMap` - Load from string map
- `NewClient` / `MustNewClient` - Create clients from config
- GitHub Enterprise support via `BaseURL` and `UploadURL`
- `EnvConfig` - Customizable environment variable names

### `pathutil/` - Path Utilities

- `Validate` - Validate paths for GitHub API
- `Normalize` - Normalize paths (clean, remove leading slashes)
- `ValidateAndNormalize` - Combined operation
- `Join` / `Split` / `Dir` / `Base` / `Ext` - Path manipulation
- `HasPrefix` - Check path prefix
- Protection against path traversal attacks

### `cliutil/` - CLI Utilities

- `GitStatusShortLines` - Get git status output
- `GitRmDeletedLines` - Generate git rm commands for deleted files
- `GitRmDeletedFile` - Write git rm commands to file

## Backward Compatibility

The root `gogithub` package provides backward-compatible re-exports:

```go
// Old style (still works)
import "github.com/grokify/gogithub"
c := gogithub.NewClient(httpClient)
issues, _ := c.SearchIssuesAll(ctx, gogithub.Query{...}, nil)

// New style (preferred)
import (
    "github.com/grokify/gogithub/auth"
    "github.com/grokify/gogithub/search"
)
gh := auth.NewGitHubClient(ctx, token)
c := search.NewClient(gh)
issues, _ := c.SearchIssuesAll(ctx, search.Query{...}, nil)
```

**Re-exported types and constants:**

- `Query`, `Issues`, `Issue` types
- `ParamUser`, `ParamState`, `ParamStateValueOpen`, `ParamIs`, `ParamIsValuePR`, `ParamPerPageValueMax`
- `UsernameDependabot`, `UserIDDependabot`
- `BaseURLRepoAPI`, `BaseURLRepoHTML`

**Deprecated methods (still available on `Client`):**

- `SearchOpenPullRequests`
- `SearchIssues`
- `SearchIssuesAll`

## Breaking Changes

- **Recommended import path**: While backward compatible, new code should import from subpackages
- **Removed files**:
  - `api_search.go` (moved to `search/`)
  - `constants.go` (moved to `search/`)
  - `issue.go` (moved to `search/`)
  - `pullrequest/` directory (moved to `pr/` and `cmd/`)

## Dependency Updates

- `github.com/google/go-github` upgraded from v68 to v81
- Go version updated to 1.24.0

## Migration Guide

1. **No immediate action required**: Existing code using the root package will continue to work
2. **Gradual migration**: Update imports to use subpackages for new code
3. **Full migration**: Replace root package imports with specific subpackages

Example migration:

```go
// Before
import "github.com/grokify/gogithub"

c := gogithub.NewClient(httpClient)
result, resp, err := c.SearchIssues(ctx, gogithub.Query{
    gogithub.ParamUser: "grokify",
}, nil)

// After
import (
    "github.com/grokify/gogithub/auth"
    "github.com/grokify/gogithub/search"
)

gh := auth.NewGitHubClient(ctx, token)
c := search.NewClient(gh)
result, resp, err := c.SearchIssues(ctx, search.Query{
    search.ParamUser: "grokify",
}, nil)
```
