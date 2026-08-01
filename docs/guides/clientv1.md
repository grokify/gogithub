# Version-Isolated Client

The `clientv1` package provides a stable, version-isolated wrapper around go-github. Use it to avoid breaking changes when go-github updates its major version.

## Why Use clientv1?

The `google/go-github` library increments major versions frequently (v88 → v89 → v90...), requiring import path changes in all consuming code. This creates churn across your dependency tree.

The `clientv1` package solves this by:

## What Does "v1" Mean?

The "1" in `clientv1` is the **API version of the wrapper interface**, not the go-github version it wraps.

- `clientv1` defines a stable Client interface (v1 of our API)
- Internally it wraps whatever go-github version is current (v88, v89, v90...)
- If we ever need breaking changes to the wrapper's interface, we'd create `clientv2`

Consumers import `clientv1` and stay on it indefinitely, while gogithub updates the underlying go-github dependency without affecting them.

## Benefits

The `clientv1` package solves the version churn problem by:

- Defining **stable types** (`gogithub.User`, `gogithub.Repository`, etc.) that don't change
- Providing a **Client interface** that isolates consumers from go-github version details
- Being the **single upgrade point**: update gogithub once, all consumers benefit

## Installation

```bash
go get github.com/grokify/gogithub
```

## Basic Usage

```go
package main

import (
    "context"
    "fmt"

    "github.com/grokify/gogithub"
    "github.com/grokify/gogithub/clientv1"
)

func main() {
    ctx := context.Background()
    
    // Create client with token
    client, err := clientv1.NewClient(ctx, "your-github-token")
    if err != nil {
        panic(err)
    }

    // All methods return stable gogithub.* types
    user, err := client.GetAuthenticatedUser(ctx)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Logged in as: %s\n", user.Login)

    // List repositories
    repos, err := client.ListUserRepos(ctx, user.Login)
    if err != nil {
        panic(err)
    }
    for _, repo := range repos {
        fmt.Printf("- %s (%s)\n", repo.FullName, repo.Language)
    }
}
```

## Available Methods

### Authentication

| Method | Returns | Description |
|--------|---------|-------------|
| `GetAuthenticatedUser(ctx)` | `*gogithub.User` | Get the current user |
| `GetUser(ctx, username)` | `*gogithub.User` | Get a specific user |

### Repositories

| Method | Returns | Description |
|--------|---------|-------------|
| `GetRepository(ctx, owner, repo)` | `*gogithub.Repository` | Get repository details |
| `ListUserRepos(ctx, user)` | `[]*gogithub.Repository` | List user's repositories |
| `ListOrgRepos(ctx, org)` | `[]*gogithub.Repository` | List organization's repositories |

### Content

| Method | Returns | Description |
|--------|---------|-------------|
| `GetFileContent(ctx, owner, repo, path, opts)` | `[]byte` | Get file content |
| `GetFileContentString(ctx, owner, repo, path, opts)` | `string` | Get file content as string |
| `ListDirectory(ctx, owner, repo, path, opts)` | `[]*gogithub.FileContent` | List directory contents |
| `FileExists(ctx, owner, repo, path, opts)` | `bool` | Check if file exists |

### Git References

| Method | Returns | Description |
|--------|---------|-------------|
| `GetRef(ctx, owner, repo, ref)` | `*gogithub.Reference` | Get a git reference |
| `GetBranchSHA(ctx, owner, repo, branch)` | `string` | Get commit SHA for branch |
| `GetTagSHA(ctx, owner, repo, tag)` | `string` | Get commit SHA for tag |
| `ListBranches(ctx, owner, repo)` | `[]*gogithub.Branch` | List all branches |

### Commits

| Method | Returns | Description |
|--------|---------|-------------|
| `GetCommit(ctx, owner, repo, sha)` | `*gogithub.Commit` | Get commit details |
| `ListCommits(ctx, owner, repo, opts)` | `[]*gogithub.Commit` | List commits |

### Pull Requests

| Method | Returns | Description |
|--------|---------|-------------|
| `GetPullRequest(ctx, owner, repo, number)` | `*gogithub.PullRequest` | Get PR details |
| `ListPullRequests(ctx, owner, repo, opts)` | `[]*gogithub.PullRequest` | List pull requests |
| `CreatePullRequest(ctx, owner, repo, input)` | `*gogithub.PullRequest` | Create a new PR |

### Checks

| Method | Returns | Description |
|--------|---------|-------------|
| `ListCheckRuns(ctx, owner, repo, ref)` | `[]*gogithub.CheckRun` | List check runs for a ref |

### Releases

| Method | Returns | Description |
|--------|---------|-------------|
| `GetLatestRelease(ctx, owner, repo)` | `*gogithub.Release` | Get latest release |
| `GetReleaseByTag(ctx, owner, repo, tag)` | `*gogithub.Release` | Get release by tag |
| `ListReleases(ctx, owner, repo)` | `[]*gogithub.Release` | List all releases |

## Stable Types

All types are defined in the root `gogithub` package:

```go
import "github.com/grokify/gogithub"

var user *gogithub.User
var repo *gogithub.Repository
var ref *gogithub.Reference
var commit *gogithub.Commit
var pr *gogithub.PullRequest
var check *gogithub.CheckRun
var release *gogithub.Release
```

## Escape Hatch

For advanced use cases not yet wrapped, use `Raw()` to access the underlying go-github client:

```go
import "github.com/google/go-github/v88/github"

// Get raw client (couples you to go-github version)
raw := client.Raw().(*github.Client)

// Use go-github directly
result, _, err := raw.Actions.ListWorkflowRunsByFileName(ctx, owner, repo, "ci.yml", nil)
```

!!! warning
    Using `Raw()` couples your code to a specific go-github version. Prefer using the wrapped methods when possible.

## Migration from Legacy Packages

### Before (version-coupled)

```go
import (
    "github.com/google/go-github/v88/github"
    "github.com/grokify/gogithub/auth"
    "github.com/grokify/gogithub/repo"
)

gh, _ := auth.NewGitHubClient(ctx, token)
ref, _, _ := gh.Git.GetRef(ctx, owner, repoName, "refs/heads/main")
sha := ref.GetObject().GetSHA()
content, _ := repo.GetFileContent(ctx, gh, owner, repoName, path, nil)
```

### After (version-isolated)

```go
import (
    "github.com/grokify/gogithub"
    "github.com/grokify/gogithub/clientv1"
)

client, _ := clientv1.NewClient(ctx, token)
sha, _ := client.GetBranchSHA(ctx, owner, repoName, "main")
content, _ := client.GetFileContent(ctx, owner, repoName, path, nil)
```

## Best Practices

1. **Use `clientv1` for new code** - Avoid direct go-github imports
2. **Import types from root package** - `gogithub.User`, not `clientv1.User`
3. **Avoid `Raw()` when possible** - Use wrapped methods to stay version-isolated
4. **Check for new methods** - When go-github adds features, we add wrapped methods
