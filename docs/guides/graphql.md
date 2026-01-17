# GraphQL API

The `graphql` package provides access to GitHub's GraphQL API for user contribution statistics. This data powers the contribution graph shown on GitHub profile pages.

## Overview

The package provides two approaches:

| Approach | Function | Speed | Additions/Deletions | Public/Private Split |
|----------|----------|-------|---------------------|---------------------|
| Quick Stats | `GetContributionStats` | Fast | No | Partial |
| Detailed Stats | `GetCommitStats` | Slower | Yes | Yes |

## Authentication

The GraphQL API **always requires authentication**, even for public data:

```go
import "github.com/grokify/gogithub/graphql"

ctx := context.Background()
client := graphql.NewClient(ctx, "your-github-token")
```

!!! note "Minimal Token Required"
    For public data only, a token with **no scopes** is sufficient. Create a "Fine-grained token" with no repository or account permissions.

### GitHub Enterprise

```go
client := graphql.NewEnterpriseClient(ctx, "your-token", "https://github.mycompany.com/api/graphql")
```

## Quick Stats (ContributionsCollection)

This approach queries the same data that powers the GitHub profile contribution graph. It's fast (one query per year) but doesn't provide line additions/deletions.

### Basic Usage

```go
from := time.Now().AddDate(-1, 0, 0)  // 1 year ago
to := time.Now()

stats, err := graphql.GetContributionStats(ctx, client, "octocat", from, to)
if err != nil {
    return err
}

fmt.Printf("Total commits: %d\n", stats.TotalCommitContributions)
fmt.Printf("Total PRs: %d\n", stats.TotalPRContributions)
fmt.Printf("Total issues: %d\n", stats.TotalIssueContributions)
fmt.Printf("Private contributions: %d\n", stats.RestrictedContributions)
```

### Multi-Year Queries

The GitHub API limits queries to 1 year. Use `GetContributionStatsMultiYear` for longer periods:

```go
from := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
to := time.Now()

stats, err := graphql.GetContributionStatsMultiYear(ctx, client, "octocat", from, to)
```

### ContributionStats Fields

```go
type ContributionStats struct {
    Username                     string
    From                         time.Time
    To                           time.Time
    TotalCommitContributions     int
    TotalIssueContributions      int
    TotalPRContributions         int
    TotalPRReviewContributions   int
    TotalRepositoryContributions int
    RestrictedContributions      int  // All private contributions
    ContributionsByMonth         []MonthlyContribution
}
```

### Monthly Breakdown

```go
for _, month := range stats.ContributionsByMonth {
    fmt.Printf("%s: %d contributions\n", month.YearMonth(), month.Count)
}
```

!!! warning "RestrictedContributions Limitation"
    `RestrictedContributions` includes **all** private contribution types (commits, issues, PRs, reviews), not just commits. For commit-specific private counts, use the detailed stats approach.

## Detailed Stats (Commit History)

This approach iterates through repositories to get actual commit data, including additions and deletions. It supports precise public/private filtering.

### Basic Usage

```go
from := time.Now().AddDate(-1, 0, 0)
to := time.Now()

stats, err := graphql.GetCommitStats(ctx, client, "octocat", from, to, graphql.VisibilityAll)
if err != nil {
    return err
}

fmt.Printf("Total commits: %d\n", stats.TotalCommits)
fmt.Printf("Additions: %d\n", stats.Additions)
fmt.Printf("Deletions: %d\n", stats.Deletions)
```

### Visibility Filtering

```go
// All repositories
all, err := graphql.GetCommitStats(ctx, client, "octocat", from, to, graphql.VisibilityAll)

// Public only
public, err := graphql.GetCommitStats(ctx, client, "octocat", from, to, graphql.VisibilityPublic)

// Private only
private, err := graphql.GetCommitStats(ctx, client, "octocat", from, to, graphql.VisibilityPrivate)
```

### Get All Three at Once

```go
all, public, private, err := graphql.GetCommitStatsByVisibility(ctx, client, "octocat", from, to)
if err != nil {
    return err
}

fmt.Printf("Public commits: %d (additions: %d, deletions: %d)\n",
    public.TotalCommits, public.Additions, public.Deletions)
fmt.Printf("Private commits: %d (additions: %d, deletions: %d)\n",
    private.TotalCommits, private.Additions, private.Deletions)
```

### CommitStats Fields

```go
type CommitStats struct {
    Username     string
    From         time.Time
    To           time.Time
    Visibility   Visibility
    TotalCommits int
    Additions    int
    Deletions    int
    ByMonth      []MonthlyCommitStats
    ByRepo       []RepoCommitStats
}
```

### Monthly Breakdown with Lines

```go
for _, month := range stats.ByMonth {
    fmt.Printf("%s: %d commits (+%d/-%d)\n",
        month.YearMonth(),
        month.Commits,
        month.Additions,
        month.Deletions,
    )
}
```

### Per-Repository Breakdown

```go
for _, repo := range stats.ByRepo {
    visibility := "public"
    if repo.IsPrivate {
        visibility = "private"
    }
    fmt.Printf("%s/%s (%s): %d commits (+%d/-%d)\n",
        repo.Owner, repo.Name, visibility,
        repo.Commits, repo.Additions, repo.Deletions,
    )
}
```

## Complete Example

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/grokify/gogithub/graphql"
)

func main() {
    ctx := context.Background()
    client := graphql.NewClient(ctx, "your-token")

    username := "octocat"
    from := time.Now().AddDate(-1, 0, 0)
    to := time.Now()

    // Quick stats
    fmt.Println("=== Quick Stats (Profile View) ===")
    contribStats, err := graphql.GetContributionStats(ctx, client, username, from, to)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Commits: %d\n", contribStats.TotalCommitContributions)
    fmt.Printf("PRs: %d\n", contribStats.TotalPRContributions)
    fmt.Printf("Issues: %d\n", contribStats.TotalIssueContributions)

    // Detailed stats
    fmt.Println("\n=== Detailed Stats ===")
    all, public, private, err := graphql.GetCommitStatsByVisibility(ctx, client, username, from, to)
    if err != nil {
        panic(err)
    }

    fmt.Printf("\nAll repositories:\n")
    fmt.Printf("  Commits: %d, Additions: %d, Deletions: %d\n",
        all.TotalCommits, all.Additions, all.Deletions)

    fmt.Printf("\nPublic repositories:\n")
    fmt.Printf("  Commits: %d, Additions: %d, Deletions: %d\n",
        public.TotalCommits, public.Additions, public.Deletions)

    fmt.Printf("\nPrivate repositories:\n")
    fmt.Printf("  Commits: %d, Additions: %d, Deletions: %d\n",
        private.TotalCommits, private.Additions, private.Deletions)

    // Monthly breakdown
    fmt.Println("\n=== Monthly Breakdown (Public) ===")
    for _, m := range public.ByMonth {
        fmt.Printf("%s: %d commits (+%d/-%d)\n",
            m.YearMonth(), m.Commits, m.Additions, m.Deletions)
    }
}
```

## Rate Limits

The GraphQL API has its own rate limiting system based on "points":

- **5,000 points per hour** for authenticated requests
- Complex queries cost more points
- `GetContributionStats`: ~1 point per call
- `GetCommitStats`: Variable, depends on number of repos and commits

## API Reference

See [pkg.go.dev/github.com/grokify/gogithub/graphql](https://pkg.go.dev/github.com/grokify/gogithub/graphql) for complete API documentation.
