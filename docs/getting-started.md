# Getting Started

## Installation

```bash
go get github.com/grokify/gogithub
```

## Prerequisites

- Go 1.21 or later
- A GitHub personal access token (for most operations)

### Creating a GitHub Token

1. Go to [GitHub Settings > Developer settings > Personal access tokens](https://github.com/settings/tokens)
2. Click "Generate new token (classic)" or "Fine-grained tokens"
3. Select the scopes you need:

    | Scope | Required For |
    |-------|--------------|
    | `repo` | Private repository access |
    | `public_repo` | Public repository write access |
    | (no scopes) | Public read-only access |

4. Copy the token and store it securely

## Basic Usage

### Creating a Client (Recommended)

Use the `clientv1` package for a version-isolated client that won't break when go-github updates:

```go
import (
    "github.com/grokify/gogithub"
    "github.com/grokify/gogithub/clientv1"
)

ctx := context.Background()
client, err := clientv1.NewClient(ctx, "your-github-token")
if err != nil {
    panic(err)
}

// Returns stable gogithub.* types
user, _ := client.GetAuthenticatedUser(ctx)  // *gogithub.User
repos, _ := client.ListUserRepos(ctx, user.Login)  // []*gogithub.Repository
```

See the [Version-Isolated Client Guide](guides/clientv1.md) for full documentation.

### Creating a Client from Environment or GitHub Enterprise

The `config` package can build a `clientv1.Client` too, via `NewClientV1`/`MustNewClientV1`:

=== "From Environment"

    ```go
    import "github.com/grokify/gogithub/config"

    cfg, err := config.FromEnv()
    if err != nil {
        panic(err)
    }
    client, err := cfg.NewClientV1(ctx)
    ```

=== "GitHub Enterprise"

    ```go
    import "github.com/grokify/gogithub/config"

    cfg := &config.Config{
        Token:     "your-token",
        BaseURL:   "https://github.mycompany.com/api/v3",
        UploadURL: "https://github.mycompany.com/api/uploads",
    }
    client, err := cfg.NewClientV1(ctx)
    ```

`config.Config.NewClient()`/`MustNewClient()` (returning `*github.Client` directly) are deprecated;
use `NewClientV1()`/`MustNewClientV1()` instead.

### Escape Hatch: Raw go-github Client

For advanced operations not yet wrapped by `clientv1`, `auth.NewGitHubClient()` still returns a raw
`*github.Client`:

```go
import "github.com/grokify/gogithub/auth"

ctx := context.Background()
gh, err := auth.NewGitHubClient(ctx, "your-github-token")
if err != nil {
    panic(err)
}
```

This couples your code to go-github's current major version — prefer `clientv1.Client` wherever a
package accepts it.

### Environment Variables

The `config` package reads from these environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `GITHUB_TOKEN` | Personal access token | (required) |
| `GITHUB_OWNER` | Default repository owner | - |
| `GITHUB_REPO` | Default repository name | - |
| `GITHUB_BRANCH` | Default branch | `main` |
| `GITHUB_BASE_URL` | API base URL | `https://api.github.com/` |
| `GITHUB_UPLOAD_URL` | Upload URL | `https://uploads.github.com/` |

## Common Operations

These operations all take the `clientv1.Client` created in [Creating a Client (Recommended)](#creating-a-client-recommended) above.

### Search for Issues/PRs

```go
import "github.com/grokify/gogithub/search"

c := search.NewClient(client)

// Using Query map
issues, err := c.SearchIssuesAll(ctx, search.Query{
    search.ParamUser:  "octocat",
    search.ParamIs:    search.ParamIsValueIssue,
    search.ParamState: search.ParamStateValueOpen,
}, nil)

// Using QueryBuilder (type-safe)
qb := search.NewQueryBuilder().
    User("octocat").
    Is(search.ParamIsValueIssue).
    State(search.ParamStateValueOpen)

issues, err = c.SearchIssuesAll(ctx, qb.Build(), nil)
```

### Create a Branch and Commit

```go
import "github.com/grokify/gogithub/repo"

// Get the SHA of the base branch
sha, err := repo.GetBranchSHA(ctx, client, "owner", "repo", "main")

// Create a new branch
err = repo.CreateBranch(ctx, client, "owner", "repo", "feature-branch", sha)

// Commit files
files := []repo.FileContent{
    {Path: "hello.txt", Content: []byte("Hello, World!")},
}
commitSHA, err := repo.CreateCommit(ctx, client, "owner", "repo", "feature-branch", "Add hello.txt", files)
```

### Create a Pull Request

```go
import "github.com/grokify/gogithub/pr"

pullRequest, err := pr.CreatePR(ctx, client,
    "upstream-owner", "upstream-repo",  // base repo
    "fork-owner", "feature-branch",     // head
    "main",                             // base branch
    "My PR Title",
    "Description of changes",
)
fmt.Printf("PR URL: %s\n", pullRequest.HTMLURL)
```

### Get User Contribution Stats (GraphQL)

```go
import "github.com/grokify/gogithub/graphql"

client := graphql.NewClient(ctx, "your-github-token")

// Quick stats (like profile page)
from := time.Now().AddDate(-1, 0, 0)
to := time.Now()
stats, err := graphql.GetContributionStats(ctx, client, "octocat", from, to)
fmt.Printf("Total commits: %d\n", stats.TotalCommitContributions)

// Detailed stats with additions/deletions
commitStats, err := graphql.GetCommitStats(ctx, client, "octocat", from, to, graphql.VisibilityPublic)
fmt.Printf("Additions: %d, Deletions: %d\n", commitStats.Additions, commitStats.Deletions)
```

## Next Steps

- [Authentication Guide](guides/auth.md) - Detailed authentication options
- [Search API Guide](guides/search.md) - Query syntax and examples
- [GraphQL Guide](guides/graphql.md) - Contribution statistics
- [Testing Guide](guides/testing.md) - Running unit and integration tests
