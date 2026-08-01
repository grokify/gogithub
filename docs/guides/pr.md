# Pull Requests

The `pr` package provides functions for creating and managing pull requests. All functions take a
[`clientv1.Client`](clientv1.md) and return stable `gogithub.*` types.

## Creating Pull Requests

### Basic PR Creation

```go
import "github.com/grokify/gogithub/pr"

pullRequest, err := pr.CreatePR(ctx, client,
    "upstream-owner", "upstream-repo",  // Base repository
    "fork-owner", "feature-branch",     // Head (your fork and branch)
    "main",                             // Base branch to merge into
    "Add new feature",                  // PR title
    "This PR adds a new feature...",    // PR body
)
if err != nil {
    return err
}

fmt.Printf("PR created: %s\n", pullRequest.HTMLURL)
```

### Cross-Fork PRs

When creating a PR from a fork to an upstream repository, `repo.EnsureFork` creates the fork if it
doesn't already exist and returns its owner/repo names:

```go
// First, ensure you have a fork
forkOwner, forkRepo, err := repo.EnsureFork(ctx, client, "upstream-owner", "upstream-repo", "your-username")
if err != nil {
    return err
}

// Create a branch in your fork
sha, _ := repo.GetBranchSHA(ctx, client, forkOwner, forkRepo, "main")
repo.CreateBranch(ctx, client, forkOwner, forkRepo, "my-feature", sha)

// Make commits to your fork's branch
// ...

// Create PR from fork to upstream
pullRequest, err := pr.CreatePR(ctx, client,
    "upstream-owner", "upstream-repo",  // base
    forkOwner, "my-feature",            // head
    "main",
    "My contribution",
    "Description of changes",
)
```

## Listing Pull Requests

### List PRs for a Repository

```go
prs, err := pr.ListPRs(ctx, client, "owner", "repo", &clientv1.ListPullRequestsOptions{
    State:     "open",
    Sort:      "created",
    Direction: "desc",
})

for _, p := range prs {
    fmt.Printf("#%d: %s\n", p.Number, p.Title)
}
```

### Get Single PR

```go
pullRequest, err := pr.GetPR(ctx, client, "owner", "repo", 123)
fmt.Printf("State: %s\n", pullRequest.State)
if pullRequest.Mergeable != nil {
    fmt.Printf("Mergeable: %v\n", *pullRequest.Mergeable)
}
```

## Managing Pull Requests

### Merge a PR

```go
result, err := pr.MergePR(ctx, client, "owner", "repo", 123, "Merge PR #123", &clientv1.MergePullRequestOptions{
    MergeMethod: "squash",  // "merge", "squash", or "rebase"
})

if result.Merged {
    fmt.Println("PR merged successfully")
}
```

### Close a PR

```go
pullRequest, err := pr.ClosePR(ctx, client, "owner", "repo", 123)
fmt.Printf("PR state: %s\n", pullRequest.State)  // "closed"
```

## Reviews and Comments

### Get PR Diff

Retrieve the diff content for a pull request:

```go
diff, err := pr.GetPRDiff(ctx, client, "owner", "repo", 123)
if err != nil {
    return err
}
fmt.Println(diff)  // Raw diff output
```

### List Reviews

```go
reviews, err := pr.ListPRReviews(ctx, client, "owner", "repo", 123)
for _, review := range reviews {
    fmt.Printf("%s: %s - %s\n",
        review.User.Login,
        review.State,
        review.Body)
}
```

### Submit a Review

Use `CreateReview` to submit a formal review:

```go
// Approve the PR
review, err := pr.CreateReview(ctx, client, "owner", "repo", 123,
    pr.ReviewEventApprove,
    "LGTM! Great work.",
)

// Request changes
review, err := pr.CreateReview(ctx, client, "owner", "repo", 123,
    pr.ReviewEventRequestChanges,
    "Please address the comments below.",
)

// Add a comment review (neither approve nor request changes)
review, err := pr.CreateReview(ctx, client, "owner", "repo", 123,
    pr.ReviewEventComment,
    "Some observations about the implementation...",
)
```

Review events:

| Event | Description |
|-------|-------------|
| `pr.ReviewEventApprove` | Approve the PR |
| `pr.ReviewEventRequestChanges` | Request changes before merging |
| `pr.ReviewEventComment` | General comment without approval status |

`pr.ApprovePR`, `pr.RequestChangesPR`, and `pr.CommentPR` wrap `CreateReview` with the matching
event for convenience.

### Add Comments

**General PR comment** (appears in the conversation):

```go
comment, err := pr.CreateIssueComment(ctx, client, "owner", "repo", 123,
    "Thanks for the contribution! I have a few suggestions.",
)
```

**Line-level comment** (appears on specific code):

```go
comment, err := pr.CreateLineComment(ctx, client, "owner", "repo", 123,
    "abc123def",           // Commit SHA
    "src/main.go",         // File path
    "Consider using a constant here for better readability.",
    42,                    // Line number
)
```

### List PR Comments

```go
comments, err := pr.ListPRComments(ctx, client, "owner", "repo", 123)
for _, c := range comments {
    fmt.Printf("%s at %s:%d: %s\n",
        c.User.Login,
        c.Path,
        c.Line,
        c.Body)
}
```

## Complete Workflow Example

Here's a complete example of creating a contribution via PR:

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
    client, err := clientv1.NewClient(ctx, "your-token")
    if err != nil {
        panic(err)
    }

    upstreamOwner := "upstream-owner"
    upstreamRepo := "upstream-repo"

    // 1. Ensure a fork exists
    forkOwner, forkRepo, err := repo.EnsureFork(ctx, client, upstreamOwner, upstreamRepo, "your-username")
    if err != nil {
        panic(err)
    }

    // 2. Create a feature branch
    sha, err := repo.GetBranchSHA(ctx, client, forkOwner, forkRepo, "main")
    if err != nil {
        panic(err)
    }

    branchName := "add-documentation"
    err = repo.CreateBranch(ctx, client, forkOwner, forkRepo, branchName, sha)
    if err != nil {
        panic(err)
    }

    // 3. Make changes
    files := []repo.FileContent{
        {Path: "CONTRIBUTING.md", Content: []byte("# Contributing\n\nWelcome!")},
    }
    _, err = repo.CreateCommit(ctx, client, forkOwner, forkRepo, branchName, "Add contributing guide", files)
    if err != nil {
        panic(err)
    }

    // 4. Create pull request
    pullRequest, err := pr.CreatePR(ctx, client,
        upstreamOwner, upstreamRepo,
        forkOwner, branchName,
        "main",
        "Add contributing guide",
        "This PR adds a CONTRIBUTING.md file to help new contributors.",
    )
    if err != nil {
        panic(err)
    }

    fmt.Printf("PR created: %s\n", pullRequest.HTMLURL)
}
```

## Error Handling

### PRError

```go
pullRequest, err := pr.CreatePR(ctx, client, baseOwner, baseRepo, headOwner, headBranch, baseBranch, title, body)
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
