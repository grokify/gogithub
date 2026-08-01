# GoGitHub

[![Go CI][go-ci-svg]][go-ci-url]
[![Go Lint][go-lint-svg]][go-lint-url]
[![Go SAST][go-sast-svg]][go-sast-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![Docs][docs-mkdoc-svg]][docs-mkdoc-url]
[![Visualization][viz-svg]][viz-url]
[![License][license-svg]][license-url]

 [go-ci-svg]: https://github.com/grokify/gogithub/actions/workflows/go-ci.yaml/badge.svg?branch=main
 [go-ci-url]: https://github.com/grokify/gogithub/actions/workflows/go-ci.yaml
 [go-lint-svg]: https://github.com/grokify/gogithub/actions/workflows/go-lint.yaml/badge.svg?branch=main
 [go-lint-url]: https://github.com/grokify/gogithub/actions/workflows/go-lint.yaml
 [go-sast-svg]: https://github.com/grokify/gogithub/actions/workflows/go-sast-codeql.yaml/badge.svg?branch=main
 [go-sast-url]: https://github.com/grokify/gogithub/actions/workflows/go-sast-codeql.yaml
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/grokify/gogithub
 [docs-godoc-url]: https://pkg.go.dev/github.com/grokify/gogithub
 [docs-mkdoc-svg]: https://img.shields.io/badge/Go-dev%20guide-blue.svg
 [docs-mkdoc-url]: https://grokify.github.io/gogithub
 [viz-svg]: https://img.shields.io/badge/visualizaton-Go-blue.svg
 [viz-url]: https://mango-dune-07a8b7110.1.azurestaticapps.net/?repo=grokify%2Fgogithub
 [loc-svg]: https://tokei.rs/b1/github/grokify/gogithub
 [repo-url]: https://github.com/grokify/gogithub
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/grokify/gogithub/blob/main/LICENSE

**[Documentation](https://grokify.github.io/gogithub/)** | **[API Reference](https://pkg.go.dev/github.com/grokify/gogithub)**

`gogithub` is a high-level Go module for interacting with the GitHub API. It wraps [go-github](https://github.com/google/go-github) with convenience functions organized by operation type.

## Installation

```bash
go get github.com/grokify/gogithub
```

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
    sha, _ := client.GetBranchSHA(ctx, "owner", "repo", "main")  // string
    content, _ := client.GetFileContent(ctx, "owner", "repo", "README.md", nil)  // []byte
    
    fmt.Printf("Found %d repos for %s\n", len(repos), user.Login)
}
```

**Benefits:**

- Types like `gogithub.User` and `gogithub.Repository` won't change
- When go-github updates, only gogithub needs updating—your code stays the same  
- Single import pattern: `gogithub` for types, `clientv1` for the client

**What does "v1" mean?** The "1" is the API version of the wrapper interface, not the go-github version. Consumers stay on `clientv1` indefinitely while gogithub internally updates go-github (v88 → v89 → v90...). If we ever need breaking changes to the wrapper's interface, we'd create `clientv2`.

See the [Version-Isolated Client Guide](https://grokify.github.io/gogithub/guides/clientv1/) for full documentation.

## Directory Structure

The package is organized into subdirectories by operation type for scalability:

```
gogithub/
├── types.go              # Stable types (User, Repository, etc.)
├── gogithub.go           # Client factory, backward-compatible re-exports
├── clientv1/             # Version-isolated client wrapper (RECOMMENDED)
│   ├── client.go         # Client interface
│   ├── client_impl.go    # Implementation wrapping go-github
│   ├── convert.go        # Type converters
│   └── doc.go            # Package documentation
├── auth/                 # Authentication utilities
│   └── auth.go           # NewGitHubClient, GetAuthenticatedUser
├── config/               # Configuration utilities
│   └── config.go         # Config struct, FromEnv, GitHub Enterprise support
├── errors/               # Error types and translation
│   └── errors.go         # APIError, Translate, IsNotFound, IsRateLimited
├── graphql/              # GraphQL API for contribution statistics
│   ├── client.go         # NewClient, NewEnterpriseClient
│   ├── contributions.go  # GetContributionStats, GetContributionStatsMultiYear
│   └── commitstats.go    # GetCommitStats, GetCommitStatsByVisibility
├── profile/              # User profile aggregation
│   ├── profile.go        # UserProfile, GetUserProfile
│   ├── calendar.go       # ContributionCalendar, streaks
│   ├── activity.go       # MonthlyActivity, ActivityTimeline, MonthlyStats
│   ├── monthly_output.go # WriteMonthlyFile, WriteMonthlyFiles
│   ├── stats_report.go   # StatsReport, BuildStatsReport, LoadMonthlyFiles
│   ├── stats_render.go   # RenderToMarkdown, RenderToHTML, RenderToText
│   ├── readme/           # README.md generation
│   │   ├── readme.go     # Generate, DefaultConfig
│   │   ├── heatmap.go    # RenderHeatmap (Unicode contribution calendar)
│   │   └── template.go   # Template helpers
│   └── svg/              # SVG visualization generation
│       ├── card.go       # GenerateStatsCard
│       ├── stats.go      # Stats rendering
│       ├── theme.go      # Theme definitions (dark, dracula, nord, etc.)
│       ├── icons.go      # Metric icons
│       └── chart/        # Chart primitives
│           ├── bar.go    # Bar chart rendering
│           └── types.go  # Chart data types
├── pathutil/             # Path validation and normalization
│   └── pathutil.go       # Validate, Normalize, Join, Split
├── search/               # Search API operations
│   ├── search.go         # SearchIssues, SearchIssuesAll
│   ├── query.go          # Query builder, parameter constants
│   └── issues.go         # Issues type, table generation
├── repo/                 # Repository operations
│   ├── fork.go           # EnsureFork, GetDefaultBranch
│   ├── branch.go         # CreateBranch, GetBranchSHA, DeleteBranch
│   ├── commit.go         # CreateCommit (Git tree API), ReadLocalFiles
│   ├── list.go           # ListOrgRepos, ListUserRepos, GetRepo
│   ├── contributors.go   # ListContributorStats, GetContributorSummary
│   └── batch.go          # Batch for atomic multi-file commits
├── pr/                   # Pull request operations
│   └── pullrequest.go    # CreatePR, GetPR, ListPRs, MergePR, ApprovePR, IsMergeable
├── release/              # Release operations
│   └── release.go        # ListReleases, GetLatestRelease, CreateRelease, DeleteRelease
├── checks/               # Check runs operations
│   └── checks.go         # ListCheckRuns, WaitForChecks, AllChecksPassed
├── sarif/                # SARIF upload for GitHub Code Scanning
│   └── sarif.go          # Upload, UploadFile, GetUploadStatus, WaitForProcessing
├── tag/                  # Git tag operations
│   └── tag.go            # ListTags, CreateTag, GetTagSHA, TagExists
├── cliutil/              # CLI utilities
│   └── status.go         # Git status helpers
├── cmd/                  # CLI tools
│   ├── gogithub/         # Main CLI (profile, search-prs, stats-report commands)
│   └── searchuserpr/     # Search user PRs example
└── web/                  # Profile Viewer web application
    └── src/              # TypeScript source (Vite, Chart.js)
```

## Usage

### Version-Isolated Client (Recommended)

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
    user, _ := client.GetAuthenticatedUser(ctx)
    repos, _ := client.ListUserRepos(ctx, user.Login)
    
    for _, repo := range repos {
        fmt.Printf("- %s (%s)\n", repo.FullName, repo.Language)
    }
}
```

### Operation Packages (search, repo, pr, checks, tag, release, sarif)

These packages take a `clientv1.Client`, so they stay version-isolated just like the client itself.

```go
package main

import (
    "context"
    "fmt"

    "github.com/grokify/gogithub/clientv1"
    "github.com/grokify/gogithub/search"
)

func main() {
    ctx := context.Background()

    client, err := clientv1.NewClient(ctx, "your-github-token")
    if err != nil {
        panic(err)
    }

    // Search for open pull requests
    c := search.NewClient(client)
    issues, err := c.SearchIssuesAll(ctx, search.Query{
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

    "github.com/grokify/gogithub/clientv1"
    "github.com/grokify/gogithub/pr"
    "github.com/grokify/gogithub/repo"
)

func main() {
    ctx := context.Background()
    client, err := clientv1.NewClient(ctx, "your-github-token")
    if err != nil {
        panic(err)
    }

    // Get branch SHA
    sha, err := repo.GetBranchSHA(ctx, client, "owner", "repo", "main")
    if err != nil {
        panic(err)
    }

    // Create a new branch
    err = repo.CreateBranch(ctx, client, "owner", "repo", "feature-branch", sha)
    if err != nil {
        panic(err)
    }

    // Create files and commit
    files := []repo.FileContent{
        {Path: "README.md", Content: []byte("# Hello")},
    }
    _, err = repo.CreateCommit(ctx, client, "owner", "repo", "feature-branch", "Add README", files)
    if err != nil {
        panic(err)
    }

    // Create pull request
    pullRequest, err := pr.CreatePR(ctx, client, "upstream-owner", "upstream-repo",
        "fork-owner", "feature-branch", "main", "My PR Title", "PR description")
    if err != nil {
        panic(err)
    }

    fmt.Printf("PR created: %s\n", pullRequest.HTMLURL)
}
```

### Waiting for CI Checks

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/grokify/gogithub/checks"
    "github.com/grokify/gogithub/clientv1"
)

func main() {
    ctx := context.Background()
    client, err := clientv1.NewClient(ctx, "your-github-token")
    if err != nil {
        panic(err)
    }

    // Wait for all checks to complete (with 10 minute timeout)
    checkRuns, allPassed, err := checks.WaitForChecks(ctx, client, "owner", "repo", "commit-sha",
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

    "github.com/grokify/gogithub/clientv1"
    "github.com/grokify/gogithub/release"
    "github.com/grokify/gogithub/tag"
)

func main() {
    ctx := context.Background()
    client, err := clientv1.NewClient(ctx, "your-github-token")
    if err != nil {
        panic(err)
    }

    // Create an annotated tag
    err = tag.CreateTag(ctx, client, "owner", "repo", "v1.0.0", "commit-sha", "Release v1.0.0")
    if err != nil {
        panic(err)
    }

    // Create a release
    rel, err := release.CreateReleaseSimple(ctx, client, "owner", "repo",
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

    fmt.Printf("Release created: %s\n", rel.HTMLURL)
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

    "github.com/grokify/gogithub/clientv1"
    "github.com/grokify/gogithub/graphql"
    "github.com/grokify/gogithub/profile"
)

func main() {
    ctx := context.Background()
    token := "your-github-token"

    restClient, err := clientv1.NewClient(ctx, token)
    if err != nil {
        panic(err)
    }
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

### Profile Output Formats

Generate profile visualizations in multiple formats:

```bash
# Generate all outputs
gogithub profile --user grokify --from 2024-01-01 --to 2024-12-31 \
    --output-readme README.md \
    --output-svg stats.svg --svg-theme dracula \
    --output-chart chart.svg \
    --output-chart-json chart.json
```

#### README with Contribution Heatmap

Generate a GitHub profile README with a Unicode contribution calendar:

```go
import "github.com/grokify/gogithub/profile/readme"

config := readme.DefaultConfig()
config.ShowHeatmap = true

output, err := readme.Generate(profile, config)
```

#### SVG Stats Card

Generate embeddable stats cards with theme support:

```go
import "github.com/grokify/gogithub/profile/svg"

// Available themes: default, dark, dracula, nord, gruvbox, solarized
card, err := svg.GenerateStatsCard(profile, svg.ThemeDracula, "My GitHub Stats")
```

#### Monthly Activity Charts

Generate charts as SVG or JSON intermediate representation:

```go
import "github.com/grokify/gogithub/profile/svg"

// SVG chart
chartSVG, err := svg.GenerateMonthlyChart(profile.Timeline, svg.ChartOptions{
    Width:  800,
    Height: 400,
})

// JSON IR for custom rendering
chartJSON, err := svg.GenerateChartJSON(profile.Timeline)
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
   - Functions take `context.Context` and `clientv1.Client` as first parameters (use `client.Raw()` internally only for operations `clientv1.Client` doesn't yet wrap)
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

    "github.com/grokify/gogithub"
    "github.com/grokify/gogithub/clientv1"
)

func Create(ctx context.Context, client clientv1.Client, description string, public bool, files map[string]string) (*gogithub.Gist, error) {
    // Implementation, e.g. using client.Raw() until clientv1.Client wraps gist operations
}

func Get(ctx context.Context, client clientv1.Client, id string) (*gogithub.Gist, error) {
    // Implementation
}
```

## Deprecated Functions

A few legacy functions that return go-github types directly are deprecated in favor of `clientv1`:

| Deprecated | Use instead |
|------------|-------------|
| `auth.NewGitHubClient()` | `clientv1.NewClient()` |
| `auth.GetAuthenticatedUser()` | `client.GetAuthenticatedUser()` |
| `auth.GetUser()` | `client.GetUser()` |
| `config.Config.NewClient()` / `MustNewClient()` | `config.Config.NewClientV1()` / `MustNewClientV1()` |

They still work — go-github upgrades are the only thing that can break them — but new code should
use `clientv1` so it never needs to change when go-github does.

## Dependencies

- [google/go-github](https://github.com/google/go-github) v89 - GitHub API client (wrapped internally by `clientv1`; avoid importing it directly, see [Version-Isolated Client](#version-isolated-client-recommended))
- [golang.org/x/oauth2](https://golang.org/x/oauth2) - OAuth2 authentication

## License

MIT License
