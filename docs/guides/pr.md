# Pull Requests

The `pr` package provides functions for creating and managing pull requests.

## Creating Pull Requests

### Basic PR Creation

```go
import "github.com/grokify/gogithub/pr"

pullRequest, err := pr.CreatePR(ctx, gh,
    "upstream-owner", "upstream-repo",  // Base repository
    "fork-owner", "feature-branch",     // Head (your fork and branch)
    "main",                             // Base branch to merge into
    "Add new feature",                  // PR title
    "This PR adds a new feature...",    // PR body
)
if err != nil {
    return err
}

fmt.Printf("PR created: %s\n", pullRequest.GetHTMLURL())
```

### Cross-Fork PRs

When creating a PR from a fork to an upstream repository:

```go
// First, ensure you have a fork
fork, err := repo.EnsureFork(ctx, gh, "upstream-owner", "upstream-repo")
if err != nil {
    return err
}

// Create a branch in your fork
sha, _ := repo.GetBranchSHA(ctx, gh, fork.GetOwner().GetLogin(), fork.GetName(), "main")
repo.CreateBranch(ctx, gh, fork.GetOwner().GetLogin(), fork.GetName(), "my-feature", sha)

// Make commits to your fork's branch
// ...

// Create PR from fork to upstream
pullRequest, err := pr.CreatePR(ctx, gh,
    "upstream-owner", "upstream-repo",           // base
    fork.GetOwner().GetLogin(), "my-feature",    // head
    "main",
    "My contribution",
    "Description of changes",
)
```

## Listing Pull Requests

### List PRs for a Repository

```go
prs, err := pr.ListPRs(ctx, gh, "owner", "repo", &github.PullRequestListOptions{
    State: "open",
    Sort:  "created",
    Direction: "desc",
})

for _, p := range prs {
    fmt.Printf("#%d: %s\n", p.GetNumber(), p.GetTitle())
}
```

### Get Single PR

```go
pullRequest, err := pr.GetPR(ctx, gh, "owner", "repo", 123)
fmt.Printf("State: %s\n", pullRequest.GetState())
fmt.Printf("Mergeable: %v\n", pullRequest.GetMergeable())
```

## Managing Pull Requests

### Merge a PR

```go
result, err := pr.MergePR(ctx, gh, "owner", "repo", 123, &github.PullRequestOptions{
    CommitTitle: "Merge PR #123",
    MergeMethod: "squash",  // "merge", "squash", or "rebase"
})

if result.GetMerged() {
    fmt.Println("PR merged successfully")
}
```

### Close a PR

```go
pullRequest, err := pr.ClosePR(ctx, gh, "owner", "repo", 123)
fmt.Printf("PR state: %s\n", pullRequest.GetState())  // "closed"
```

## Complete Workflow Example

Here's a complete example of creating a contribution via PR:

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
    gh := auth.NewGitHubClient(ctx, "your-token")

    upstreamOwner := "upstream-owner"
    upstreamRepo := "upstream-repo"

    // 1. Fork the repository
    fork, err := repo.EnsureFork(ctx, gh, upstreamOwner, upstreamRepo)
    if err != nil {
        panic(err)
    }
    forkOwner := fork.GetOwner().GetLogin()

    // 2. Create a feature branch
    sha, err := repo.GetBranchSHA(ctx, gh, forkOwner, upstreamRepo, "main")
    if err != nil {
        panic(err)
    }

    branchName := "add-documentation"
    err = repo.CreateBranch(ctx, gh, forkOwner, upstreamRepo, branchName, sha)
    if err != nil {
        panic(err)
    }

    // 3. Make changes
    files := []repo.FileContent{
        {Path: "CONTRIBUTING.md", Content: []byte("# Contributing\n\nWelcome!")},
    }
    _, err = repo.CreateCommit(ctx, gh, forkOwner, upstreamRepo, branchName, "Add contributing guide", files)
    if err != nil {
        panic(err)
    }

    // 4. Create pull request
    pullRequest, err := pr.CreatePR(ctx, gh,
        upstreamOwner, upstreamRepo,
        forkOwner, branchName,
        "main",
        "Add contributing guide",
        "This PR adds a CONTRIBUTING.md file to help new contributors.",
    )
    if err != nil {
        panic(err)
    }

    fmt.Printf("PR created: %s\n", pullRequest.GetHTMLURL())
}
```

## Error Handling

### PRError

```go
pullRequest, err := pr.CreatePR(ctx, gh, baseOwner, baseRepo, headOwner, headBranch, baseBranch, title, body)
if err != nil {
    var prErr *pr.PRError
    if errors.As(err, &prErr) {
        fmt.Printf("PR operation failed for %s/%s: %v\n",
            prErr.Owner, prErr.Repo, prErr.Err)
    }
}
```

## API Reference

See [pkg.go.dev/github.com/grokify/gogithub/pr](https://pkg.go.dev/github.com/grokify/gogithub/pr) for complete API documentation.
