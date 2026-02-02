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

### Creating a Client

=== "With Token"

    ```go
    import "github.com/grokify/gogithub/auth"

    ctx := context.Background()
    gh := auth.NewGitHubClient(ctx, "your-github-token")
    ```

=== "From Environment"

    ```go
    import "github.com/grokify/gogithub/config"

    cfg, err := config.FromEnv()
    if err != nil {
        panic(err)
    }
    gh, err := cfg.NewClient(ctx)
    ```

=== "GitHub Enterprise"

    ```go
    import "github.com/grokify/gogithub/config"

    cfg := &config.Config{
        Token:     "your-token",
        BaseURL:   "https://github.mycompany.com/api/v3",
        UploadURL: "https://github.mycompany.com/api/uploads",
    }
    gh, err := cfg.NewClient(ctx)
    ```

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

### Search for Issues/PRs

```go
import "github.com/grokify/gogithub/search"

client := search.NewClient(gh)

// Using Query map
issues, err := client.SearchIssuesAll(ctx, search.Query{
    search.ParamUser:  "octocat",
    search.ParamIs:    search.ParamIsValueIssue,
    search.ParamState: search.ParamStateValueOpen,
}, nil)

// Using QueryBuilder (type-safe)
qb := search.NewQueryBuilder().
    User("octocat").
    Is(search.ParamIsValueIssue).
    State(search.ParamStateValueOpen)

issues, err := client.SearchIssuesAll(ctx, qb.Build(), nil)
```

### Create a Branch and Commit

```go
import "github.com/grokify/gogithub/repo"

// Get the SHA of the base branch
sha, err := repo.GetBranchSHA(ctx, gh, "owner", "repo", "main")

// Create a new branch
err = repo.CreateBranch(ctx, gh, "owner", "repo", "feature-branch", sha)

// Commit files
files := []repo.FileContent{
    {Path: "hello.txt", Content: []byte("Hello, World!")},
}
commit, err := repo.CreateCommit(ctx, gh, "owner", "repo", "feature-branch", "Add hello.txt", files)
```

### Create a Pull Request

```go
import "github.com/grokify/gogithub/pr"

pullRequest, err := pr.CreatePR(ctx, gh,
    "upstream-owner", "upstream-repo",  // base repo
    "fork-owner", "feature-branch",     // head
    "main",                             // base branch
    "My PR Title",
    "Description of changes",
)
fmt.Printf("PR URL: %s\n", pullRequest.GetHTMLURL())
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
