# User Profile Statistics

The `profile` package provides comprehensive GitHub user contribution statistics, aggregating data from multiple API sources into a unified view similar to what GitHub shows on user profile pages.

## Overview

The profile package combines data from:

- **GraphQL API**: Contribution counts, commit history, repositories contributed to
- **REST API**: Contributor stats, release counts

It provides:

| Feature | Description |
|---------|-------------|
| `UserProfile` | Comprehensive stats combining all data sources |
| `ContributionCalendar` | Weekly/daily contribution grid with streak tracking |
| `ActivityTimeline` | Monthly activity feed with GitHub-style summaries |

## Token Requirements

The profile package requires authentication. See [Authentication](auth.md#token-requirements-by-use-case) for detailed requirements.

**For public data** (like viewing any user's public contributions):

| Token Type | Configuration |
|------------|---------------|
| Fine-grained PAT | Repository access: "Public Repositories (read-only)", no permissions needed |
| Classic PAT | No scopes required |

**For private repository data**:

| Token Type | Configuration |
|------------|---------------|
| Fine-grained PAT | Select repositories, Contents: Read-only |
| Classic PAT | `repo` scope |

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/google/go-github/v84/github"
    "github.com/grokify/gogithub/graphql"
    "github.com/grokify/gogithub/profile"
)

func main() {
    ctx := context.Background()
    token := "your-github-token"

    // Create clients
    restClient := github.NewClient(nil).WithAuthToken(token)
    gqlClient := graphql.NewClient(ctx, token)

    // Fetch profile for last year
    from := time.Now().AddDate(-1, 0, 0)
    to := time.Now()

    p, err := profile.GetUserProfile(ctx, restClient, gqlClient, "octocat", from, to, nil)
    if err != nil {
        panic(err)
    }

    fmt.Println(p.Summary())
    // octocat: 150 commits (+10000/-3000) in 12 repos, 25 PRs, 10 issues, 50 reviews
}
```

## UserProfile

The `UserProfile` type contains comprehensive contribution statistics:

```go
type UserProfile struct {
    Username    string
    From        time.Time
    To          time.Time

    // Summary counts
    TotalCommits            int
    TotalIssues             int
    TotalPRs                int
    TotalReviews            int
    TotalReposCreated       int
    RestrictedContributions int  // Private contributions

    // Code changes
    TotalAdditions int
    TotalDeletions int

    // Repository data
    ReposContributedTo int
    RepoStats          []RepoContribution

    // Time-series data
    Calendar *ContributionCalendar
    Activity *ActivityTimeline
}
```

### Options

Configure what data to fetch:

```go
opts := &profile.Options{
    // Filter by repository visibility
    Visibility: graphql.VisibilityAll,      // or VisibilityPublic, VisibilityPrivate

    // Fetch release counts (slower, requires additional API calls)
    IncludeReleases: true,

    // Limit how many repos to fetch releases for (0 = no limit)
    MaxReleaseFetchRepos: 10,
}

p, err := profile.GetUserProfile(ctx, restClient, gqlClient, "octocat", from, to, opts)
```

### Helper Methods

```go
// Text summary
fmt.Println(p.Summary())

// Top repositories by commits
for _, repo := range p.TopReposByCommits(5) {
    fmt.Printf("%s: %d commits\n", repo.FullName, repo.Commits)
}

// Top repositories by lines added
for _, repo := range p.TopReposByAdditions(5) {
    fmt.Printf("%s: +%d lines\n", repo.FullName, repo.Additions)
}

// Filter by visibility
publicRepos := p.PublicRepos()
privateRepos := p.PrivateRepos()
```

## Contribution Calendar

The `ContributionCalendar` represents the contribution grid shown on GitHub profiles:

```go
type ContributionCalendar struct {
    TotalContributions int
    Weeks              []CalendarWeek
}

type CalendarWeek struct {
    StartDate time.Time      // Sunday of this week
    Days      [7]CalendarDay // Sunday through Saturday
}

type CalendarDay struct {
    Date              time.Time
    Weekday           time.Weekday
    ContributionCount int
    Level             ContributionLevel  // 0-4 intensity
}
```

### Contribution Levels

Levels approximate GitHub's visual intensity:

| Level | Count | Description |
|-------|-------|-------------|
| 0 | 0 | No contributions |
| 1 | 1-3 | Low |
| 2 | 4-6 | Medium |
| 3 | 7-9 | High |
| 4 | 10+ | Maximum |

### Calendar Methods

```go
cal := p.Calendar

// Overall stats
fmt.Printf("Total: %d contributions\n", cal.TotalContributions)
fmt.Printf("Active days: %d\n", cal.DaysWithContributions())

// Streaks
fmt.Printf("Longest streak: %d days\n", cal.LongestStreak())
fmt.Printf("Current streak: %d days\n", cal.CurrentStreak())

// Date range
first, last := cal.GetDateRange()
fmt.Printf("From %s to %s\n", first.Format("2006-01-02"), last.Format("2006-01-02"))

// Lookup specific date
day := cal.GetDay(time.Now())
if day != nil {
    fmt.Printf("Today: %d contributions (level %d)\n", day.ContributionCount, day.Level)
}

// Get week containing a date
week := cal.GetWeek(time.Now())
if week != nil {
    fmt.Printf("This week: %d contributions\n", week.TotalForWeek())
}
```

## Activity Timeline

The `ActivityTimeline` provides monthly activity breakdowns similar to GitHub's profile activity feed:

```go
type ActivityTimeline struct {
    Username string
    From     time.Time
    To       time.Time
    Months   []MonthlyActivity
}

type MonthlyActivity struct {
    Year  int
    Month time.Month

    // Contribution counts
    Commits     int
    Issues      int
    PRs         int
    Reviews     int

    // Code changes
    Additions   int
    Deletions   int

    // Repository breakdown
    CommitsByRepo map[string]int  // "owner/repo" -> count
    IssueRepos    []string
    PRRepos       []string
    ReposCreated  []string
}
```

### GitHub-Style Summaries

```go
for _, m := range p.Activity.Months {
    fmt.Printf("\n%s %d:\n", m.MonthName(), m.Year)

    // These return GitHub-style summary strings
    if s := m.CommitSummary(); s != "" {
        fmt.Printf("  - %s\n", s)  // "Created 42 commits in 5 repositories"
    }
    if s := m.PRSummary(); s != "" {
        fmt.Printf("  - %s\n", s)  // "Opened 3 pull requests in 2 repositories"
    }
    if s := m.IssueSummary(); s != "" {
        fmt.Printf("  - %s\n", s)  // "Opened 5 issues in 3 repositories"
    }
    if s := m.ReviewSummary(); s != "" {
        fmt.Printf("  - %s\n", s)  // "Reviewed 10 pull requests"
    }
    if s := m.RepoCreatedSummary(); s != "" {
        fmt.Printf("  - %s\n", s)  // "Created 1 repository: user/newrepo"
    }
}
```

### Timeline Methods

```go
timeline := p.Activity

// Aggregates
fmt.Printf("Total commits: %d\n", timeline.TotalCommits())
fmt.Printf("Total contributions: %d\n", timeline.TotalContributions())
fmt.Printf("Months with activity: %d\n", timeline.MonthsWithActivity())
fmt.Printf("Average per month: %.1f\n", timeline.AverageMonthlyContributions())

// Most active month
if most := timeline.MostActiveMonth(); most != nil {
    fmt.Printf("Most active: %s %d\n", most.MonthName(), most.Year)
}

// Get specific month
if jan := timeline.GetMonth(2024, time.January); jan != nil {
    fmt.Printf("January 2024: %d commits\n", jan.Commits)
}

// Sorting
timeline.SortByDate()     // Oldest first
timeline.SortByDateDesc() // Newest first (default)
```

### Top Repositories

```go
for _, m := range p.Activity.Months {
    fmt.Printf("%s %d top repos:\n", m.MonthName(), m.Year)
    for _, repo := range m.TopCommitRepos(3) {
        fmt.Printf("  %s: %d commits\n", repo.Repo, repo.Commits)
    }
}
```

## Complete Example

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/google/go-github/v84/github"
    "github.com/grokify/gogithub/graphql"
    "github.com/grokify/gogithub/profile"
)

func main() {
    ctx := context.Background()
    token := "your-github-token"

    restClient := github.NewClient(nil).WithAuthToken(token)
    gqlClient := graphql.NewClient(ctx, token)

    // Fetch profile for 2024
    from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
    to := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

    opts := &profile.Options{
        Visibility:      graphql.VisibilityAll,
        IncludeReleases: true,
        MaxReleaseFetchRepos: 5,
    }

    p, err := profile.GetUserProfile(ctx, restClient, gqlClient, "grokify", from, to, opts)
    if err != nil {
        panic(err)
    }

    // === Summary ===
    fmt.Println("=== Profile Summary ===")
    fmt.Println(p.Summary())

    // === Calendar Stats ===
    fmt.Println("\n=== Calendar Stats ===")
    fmt.Printf("Total contributions: %d\n", p.Calendar.TotalContributions)
    fmt.Printf("Days with activity: %d\n", p.Calendar.DaysWithContributions())
    fmt.Printf("Longest streak: %d days\n", p.Calendar.LongestStreak())
    fmt.Printf("Current streak: %d days\n", p.Calendar.CurrentStreak())

    // === Top Repos ===
    fmt.Println("\n=== Top 5 Repositories ===")
    for i, repo := range p.TopReposByCommits(5) {
        fmt.Printf("%d. %s: %d commits (+%d/-%d)\n",
            i+1, repo.FullName, repo.Commits, repo.Additions, repo.Deletions)
        if repo.Releases > 0 {
            fmt.Printf("   Releases: %d\n", repo.Releases)
        }
    }

    // === Monthly Activity ===
    fmt.Println("\n=== Monthly Activity ===")
    for _, m := range p.Activity.Months {
        if m.TotalContributions() == 0 {
            continue
        }
        fmt.Printf("\n%s %d:\n", m.MonthName(), m.Year)
        if s := m.CommitSummary(); s != "" {
            fmt.Printf("  - %s\n", s)
        }
        if s := m.PRSummary(); s != "" {
            fmt.Printf("  - %s\n", s)
        }
        if m.Additions > 0 || m.Deletions > 0 {
            fmt.Printf("  - Code: +%d/-%d lines\n", m.Additions, m.Deletions)
        }
    }
}
```

## Contributor Stats (REST API)

For per-repository contributor statistics (like GitHub's `/graphs/contributors` page), use the `repo` package:

```go
import "github.com/grokify/gogithub/repo"

// Get all contributors for a repository
stats, err := repo.ListContributorStats(ctx, restClient, "owner", "repo")

// Get stats for a specific user
userStats, err := repo.GetContributorStats(ctx, restClient, "owner", "repo", "username")

// Get summarized stats
summary, err := repo.GetContributorSummary(ctx, restClient, "owner", "repo", "username")
if summary != nil {
    fmt.Printf("%s: %d commits (+%d/-%d)\n",
        summary.Username, summary.TotalCommits,
        summary.TotalAdditions, summary.TotalDeletions)
    fmt.Printf("First commit: %s\n", summary.FirstCommit.Format("2006-01-02"))
    fmt.Printf("Last commit: %s\n", summary.LastCommit.Format("2006-01-02"))
}
```

!!! note "202 Accepted Handling"
    GitHub may return `202 Accepted` while computing statistics. The `ListContributorStats` function automatically retries with exponential backoff until stats are ready.

## API Reference

- [pkg.go.dev/github.com/grokify/gogithub/profile](https://pkg.go.dev/github.com/grokify/gogithub/profile)
- [pkg.go.dev/github.com/grokify/gogithub/repo](https://pkg.go.dev/github.com/grokify/gogithub/repo)
