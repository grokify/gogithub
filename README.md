# GoGitHub

[![Build Status][build-status-svg]][build-status-url]
[![Lint Status][lint-status-svg]][lint-status-url]
[![Go Report Card][goreport-svg]][goreport-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![License][license-svg]][license-url]

**[Documentation](https://grokify.github.io/gogithub/)** | **[API Reference](https://pkg.go.dev/github.com/grokify/gogithub)**

 [build-status-svg]: https://github.com/grokify/gogithub/actions/workflows/ci.yaml/badge.svg?branch=main
 [build-status-url]: https://github.com/grokify/gogithub/actions/workflows/ci.yaml
 [lint-status-svg]: https://github.com/grokify/gogithub/actions/workflows/lint.yaml/badge.svg?branch=main
 [lint-status-url]: https://github.com/grokify/gogithub/actions/workflows/lint.yaml
 [goreport-svg]: https://goreportcard.com/badge/github.com/grokify/gogithub
 [goreport-url]: https://goreportcard.com/report/github.com/grokify/gogithub
 [codeclimate-status-svg]: https://codeclimate.com/github/grokify/gogithub/badges/gpa.svg
 [codeclimate-status-url]: https://codeclimate.com/github/grokify/gogithub
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/grokify/gogithub
 [docs-godoc-url]: https://pkg.go.dev/github.com/grokify/gogithub
 [loc-svg]: https://tokei.rs/b1/github/grokify/gogithub
 [loc-url]: https://github.com/grokify/gogithub
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/grokify/gogithub/blob/master/LICENSE

`gogithub` is a high-level Go module for interacting with the GitHub API. It wraps [go-github](https://github.com/google/go-github) with convenience functions organized by operation type.

## Installation

```bash
go get github.com/grokify/gogithub
```

## Directory Structure

The package is organized into subdirectories by operation type for scalability:

```
gogithub/
├── gogithub.go        # Client factory, backward-compatible re-exports
├── auth/              # Authentication utilities
│   └── auth.go        # NewGitHubClient, GetAuthenticatedUser
├── config/            # Configuration utilities
│   └── config.go      # Config struct, FromEnv, GitHub Enterprise support
├── errors/            # Error types and translation
│   └── errors.go      # APIError, Translate, IsNotFound, IsRateLimited
├── graphql/           # GraphQL API for contribution statistics
│   ├── client.go      # NewClient, NewEnterpriseClient
│   ├── contributions.go # GetContributionStats, GetContributionStatsMultiYear
│   └── commitstats.go # GetCommitStats, GetCommitStatsByVisibility
├── profile/           # User profile aggregation
│   ├── profile.go     # UserProfile, GetUserProfile
│   ├── calendar.go    # ContributionCalendar, streaks
│   └── activity.go    # MonthlyActivity, ActivityTimeline
├── pathutil/          # Path validation and normalization
│   └── pathutil.go    # Validate, Normalize, Join, Split
├── search/            # Search API operations
│   ├── search.go      # SearchIssues, SearchIssuesAll
│   ├── query.go       # Query builder, parameter constants
│   └── issues.go      # Issues type, table generation
├── repo/              # Repository operations
│   ├── fork.go        # EnsureFork, GetDefaultBranch
│   ├── branch.go      # CreateBranch, GetBranchSHA, DeleteBranch
│   ├── commit.go      # CreateCommit (Git tree API), ReadLocalFiles
│   ├── list.go        # ListOrgRepos, ListUserRepos, GetRepo
│   ├── contributors.go # ListContributorStats, GetContributorSummary
│   └── batch.go       # Batch for atomic multi-file commits
├── pr/                # Pull request operations
│   └── pullrequest.go # CreatePR, GetPR, ListPRs, MergePR, ApprovePR, IsMergeable
├── release/           # Release operations
│   └── release.go     # ListReleases, GetLatestRelease, CreateRelease, DeleteRelease
├── checks/            # Check runs operations
│   └── checks.go      # ListCheckRuns, WaitForChecks, AllChecksPassed
├── tag/               # Git tag operations
│   └── tag.go         # ListTags, CreateTag, GetTagSHA, TagExists
├── cliutil/           # CLI utilities
│   └── status.go      # Git status helpers
├── cmd/               # CLI tools
│   ├── gogithub/      # Main CLI (profile, search-prs commands)
│   └── searchuserpr/  # Search user PRs example
└── web/               # Profile Viewer web application
    └── src/           # TypeScript source (Vite, Chart.js)
```

## Usage

### Basic Example

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

    // Create authenticated client
    gh := auth.NewGitHubClient(ctx, "your-github-token")

    // Search for open pull requests
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

### Creating a Pull Request

```go
package main

import (
    "context"
    "fmt"

    "github.com/grokify/gogithub/auth"
    "github.com/grokify/gogithub/pr"
    "github.com/grokify/gogithub/repo"
)

func main() {
    ctx := context.Background()
    gh := auth.NewGitHubClient(ctx, "your-github-token")

    // Get branch SHA
    sha, err := repo.GetBranchSHA(ctx, gh, "owner", "repo", "main")
    if err != nil {
        panic(err)
    }

    // Create a new branch
    err = repo.CreateBranch(ctx, gh, "owner", "repo", "feature-branch", sha)
    if err != nil {
        panic(err)
    }

    // Create files and commit
    files := []repo.FileContent{
        {Path: "README.md", Content: []byte("# Hello")},
    }
    _, err = repo.CreateCommit(ctx, gh, "owner", "repo", "feature-branch", "Add README", files)
    if err != nil {
        panic(err)
    }

    // Create pull request
    pullRequest, err := pr.CreatePR(ctx, gh, "upstream-owner", "upstream-repo",
        "fork-owner", "feature-branch", "main", "My PR Title", "PR description")
    if err != nil {
        panic(err)
    }

    fmt.Printf("PR created: %s\n", pullRequest.GetHTMLURL())
}
```

### Waiting for CI Checks

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/grokify/gogithub/auth"
    "github.com/grokify/gogithub/checks"
)

func main() {
    ctx := context.Background()
    gh := auth.NewGitHubClient(ctx, "your-github-token")

    // Wait for all checks to complete (with 10 minute timeout)
    checkRuns, allPassed, err := checks.WaitForChecks(ctx, gh, "owner", "repo", "commit-sha",
        10*time.Minute, 30*time.Second)
    if err != nil {
        panic(err)
    }

    // Get aggregate status
    status := checks.GetChecksStatus(checkRuns)
    fmt.Printf("Checks: %d passed, %d failed, %d pending\n",
        status.Passed, status.Failed, status.Pending)

    if allPassed {
        fmt.Println("All checks passed!")
    }
}
```

### Creating Tags and Releases

```go
package main

import (
    "context"
    "fmt"

    "github.com/grokify/gogithub/auth"
    "github.com/grokify/gogithub/release"
    "github.com/grokify/gogithub/tag"
)

func main() {
    ctx := context.Background()
    gh := auth.NewGitHubClient(ctx, "your-github-token")

    // Create an annotated tag
    err := tag.CreateTag(ctx, gh, "owner", "repo", "v1.0.0", "commit-sha", "Release v1.0.0")
    if err != nil {
        panic(err)
    }

    // Create a release
    rel, err := release.CreateReleaseSimple(ctx, gh, "owner", "repo",
        "v1.0.0",           // tag name
        "Version 1.0.0",    // release name
        "Release notes...", // body
        false,              // draft
        false,              // prerelease
        true,               // generate notes
    )
    if err != nil {
        panic(err)
    }

    fmt.Printf("Release created: %s\n", rel.GetHTMLURL())
}
```

### User Profile Statistics

Get comprehensive contribution statistics similar to GitHub profile pages. See the [full documentation](https://grokify.github.io/gogithub/guides/profile/) for details.

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/google/go-github/v82/github"
    "github.com/grokify/gogithub/graphql"
    "github.com/grokify/gogithub/profile"
)

func main() {
    ctx := context.Background()
    token := "your-github-token"

    restClient := github.NewClient(nil).WithAuthToken(token)
    gqlClient := graphql.NewClient(ctx, token)

    // Fetch profile for last year
    from := time.Now().AddDate(-1, 0, 0)
    to := time.Now()

    p, err := profile.GetUserProfile(ctx, restClient, gqlClient, "grokify", from, to, nil)
    if err != nil {
        panic(err)
    }

    // Summary
    fmt.Println(p.Summary())
    // grokify: 150 commits (+10000/-3000) in 12 repos, 25 PRs, 10 issues, 50 reviews

    // Calendar stats
    fmt.Printf("Longest streak: %d days\n", p.Calendar.LongestStreak())

    // Monthly activity
    for _, m := range p.Activity.Months {
        if s := m.CommitSummary(); s != "" {
            fmt.Printf("%s %d: %s\n", m.MonthName(), m.Year, s)
        }
    }
}
```

## Adding New Functionality

When adding new GitHub API functionality, follow this structure:

1. **Identify the operation category** - Determine which subdirectory the functionality belongs to:
   - `auth/` - Authentication, user identity
   - `config/` - Configuration, environment variables, GitHub Enterprise
   - `errors/` - Error types and translation utilities
   - `pathutil/` - Path validation and normalization
   - `search/` - Search API (issues, PRs, code, commits, etc.)
   - `repo/` - Repository operations (forks, branches, commits, batch operations)
   - `pr/` - Pull request operations
   - `release/` - Release and asset operations
   - Create new directories for distinct API areas (e.g., `issues/`, `actions/`, `gists/`)

2. **Create focused files** - Within each subdirectory, organize by specific functionality:
   - One file per logical grouping (e.g., `fork.go`, `branch.go`, `commit.go`)
   - Keep files focused and cohesive

3. **Use consistent patterns**:
   - Functions take `context.Context` and `*github.Client` as first parameters
   - Return appropriate error types with context
   - Provide both low-level functions and convenience wrappers

4. **Define custom error types** when needed:

   ```go
   type ForkError struct {
       Owner string
       Repo  string
       Err   error
   }

   func (e *ForkError) Error() string {
       return "failed to fork " + e.Owner + "/" + e.Repo + ": " + e.Err.Error()
   }

   func (e *ForkError) Unwrap() error {
       return e.Err
   }
   ```

5. **Add tests** in corresponding `*_test.go` files

### Example: Adding Gist Support

```
gogithub/
└── gist/
    ├── gist.go       # Create, Get, List, Update, Delete
    └── gist_test.go
```

```go
// gist/gist.go
package gist

import (
    "context"
    "github.com/google/go-github/v82/github"
)

func Create(ctx context.Context, gh *github.Client, description string, public bool, files map[string]string) (*github.Gist, error) {
    // Implementation
}

func Get(ctx context.Context, gh *github.Client, id string) (*github.Gist, error) {
    // Implementation
}
```

## Backward Compatibility

The root `gogithub` package provides backward-compatible re-exports for existing code:

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

## Dependencies

- [google/go-github](https://github.com/google/go-github) v82 - GitHub API client
- [golang.org/x/oauth2](https://golang.org/x/oauth2) - OAuth2 authentication

## License

MIT License
