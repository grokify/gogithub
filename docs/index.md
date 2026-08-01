# GoGitHub

**GoGitHub** is a high-level Go module for interacting with the GitHub API. It wraps [google/go-github](https://github.com/google/go-github) with convenience functions organized by operation type.

## Version-Isolated Client (Recommended)

The `clientv1` package provides a **stable, version-isolated** wrapper around go-github. Use it to avoid breaking changes when go-github updates its major version (v88 → v89 → v90...).

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
    client, err := clientv1.NewClient(ctx, "your-github-token")
    if err != nil {
        panic(err)
    }

    // All methods return stable gogithub.* types
    user, _ := client.GetAuthenticatedUser(ctx)  // *gogithub.User
    repos, _ := client.ListUserRepos(ctx, user.Login)  // []*gogithub.Repository
    
    fmt.Printf("Found %d repos for %s\n", len(repos), user.Login)
}
```

**Benefits:**

- Types like `gogithub.User` and `gogithub.Repository` won't change
- When go-github updates, only gogithub needs updating—your code stays the same
- Single import pattern: `gogithub` for types, `clientv1` for the client

## Features

- **Version-Isolated Client** - Stable types that don't break when go-github updates
- **Search API** - Query issues, pull requests, code, and commits with a fluent query builder
- **Repository Operations** - Fork, branch, commit, and batch file operations
- **Pull Requests** - Create, list, merge, and manage PRs
- **Releases** - List releases and download assets
- **GraphQL API** - User contribution statistics and detailed commit stats
- **Error Handling** - Typed errors with helper functions for common cases
- **GitHub Enterprise** - Full support for GitHub Enterprise Server

## Legacy Quick Example

The following packages expose go-github types directly and require updates when go-github changes major versions:

```go
package main

import (
    "context"
    "fmt"

    "github.com/grokify/gogithub/auth"
    "github.com/grokify/gogithub/search"
)

func main() {
    ctx := context.Background()
    gh := auth.NewGitHubClient(ctx, "your-github-token")

    client := search.NewClient(gh)
    issues, err := client.SearchIssuesAll(ctx, search.Query{
        search.ParamUser:  "grokify",
        search.ParamState: search.ParamStateValueOpen,
        search.ParamIs:    search.ParamIsValuePR,
    }, nil)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Found %d open PRs\n", len(issues))
}
```

## Package Overview

| Package | Description | Stability |
|---------|-------------|-----------|
| `clientv1` | **Version-isolated client** | **Stable** |
| `gogithub` (root) | Stable types (User, Repository, etc.) | **Stable** |
| [`auth`](guides/auth.md) | Authentication and client creation | Exposes go-github |
| [`config`](guides/auth.md#configuration) | Configuration from environment variables | Exposes go-github |
| [`search`](guides/search.md) | Search API with query builder | Exposes go-github |
| [`repo`](guides/repo.md) | Repository operations (fork, branch, commit, batch) | Exposes go-github |
| [`pr`](guides/pr.md) | Pull request operations | Exposes go-github |
| [`release`](guides/release.md) | Release and asset operations | Exposes go-github |
| [`graphql`](guides/graphql.md) | GraphQL API for contribution statistics | Exposes go-github |
| [`errors`](guides/errors.md) | Error types and translation | Stable |

Packages marked "Exposes go-github" will require consumer updates when go-github changes major versions. Use `clientv1` to avoid this.

## Next Steps

- [Getting Started](getting-started.md) - Installation and first steps
- [Version-Isolated Client](guides/clientv1.md) - Using the stable client API
- [API Reference](api-reference.md) - Links to pkg.go.dev documentation
- [Changelog](https://github.com/grokify/gogithub/blob/main/CHANGELOG.md) - Version history and release notes
