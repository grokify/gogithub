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

## Creating a Client

### Standard (github.com)

```go
client, err := clientv1.NewClient(ctx, "your-github-token")
```

### GitHub Enterprise

```go
client, err := clientv1.NewClientWithOptions(ctx, clientv1.ClientOptions{
    Token:     "your-token",
    BaseURL:   "https://github.mycompany.com/api/v3/",
    UploadURL: "https://github.mycompany.com/api/uploads/",
})
```

### From Config

```go
import "github.com/grokify/gogithub/config"

cfg := config.FromEnv()
client, err := cfg.NewClientV1(ctx)
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
| `GetDefaultBranch(ctx, owner, repo)` | `string` | Get default branch name |
| `CreateFork(ctx, owner, repo, opts)` | `*gogithub.Repository` | Fork a repository |

### Content

| Method | Returns | Description |
|--------|---------|-------------|
| `GetFileContent(ctx, owner, repo, path, opts)` | `[]byte` | Get file content |
| `GetFileContentString(ctx, owner, repo, path, opts)` | `string` | Get file content as string |
| `GetFileContentWithSHA(ctx, owner, repo, path, opts)` | `[]byte, string` | Get file content and its SHA |
| `ListDirectory(ctx, owner, repo, path, opts)` | `[]*gogithub.FileContent` | List directory contents |
| `FileExists(ctx, owner, repo, path, opts)` | `bool` | Check if file exists |
| `CreateFile(ctx, owner, repo, path, opts)` | `*gogithub.CreateFileResult` | Create a new file |
| `UpdateFile(ctx, owner, repo, path, opts)` | `*gogithub.CreateFileResult` | Update an existing file (needs current SHA) |
| `DeleteFile(ctx, owner, repo, path, sha, msg, opts)` | `*gogithub.DeleteFileResult` | Delete a file (needs current SHA) |

### Git References

| Method | Returns | Description |
|--------|---------|-------------|
| `GetRef(ctx, owner, repo, ref)` | `*gogithub.Reference` | Get a git reference |
| `CreateRef(ctx, owner, repo, ref, sha)` | `*gogithub.Reference` | Create a git reference |
| `UpdateRef(ctx, owner, repo, ref, sha, force)` | `*gogithub.Reference` | Update a git reference |
| `DeleteRef(ctx, owner, repo, ref)` | `error` | Delete a git reference |
| `GetBranchSHA(ctx, owner, repo, branch)` | `string` | Get commit SHA for branch |
| `GetTagSHA(ctx, owner, repo, tag)` | `string` | Get commit SHA for tag |
| `ListBranches(ctx, owner, repo)` | `[]*gogithub.Branch` | List all branches |

### Tags

| Method | Returns | Description |
|--------|---------|-------------|
| `ListTags(ctx, owner, repo)` | `[]*gogithub.Tag` | List all tags |
| `CreateTag(ctx, owner, repo, tag, sha, msg)` | `error` | Create an annotated tag |

### Commits

| Method | Returns | Description |
|--------|---------|-------------|
| `GetCommit(ctx, owner, repo, sha)` | `*gogithub.Commit` | Get commit details |
| `ListCommits(ctx, owner, repo, opts)` | `[]*gogithub.Commit` | List commits |
| `CreateCommit(ctx, owner, repo, opts)` | `*gogithub.Commit` | Create a commit |

### Git Trees and Blobs

| Method | Returns | Description |
|--------|---------|-------------|
| `GetTree(ctx, owner, repo, sha, recursive)` | `[]*gogithub.TreeNode` | Get a git tree by SHA |
| `CreateTree(ctx, owner, repo, base, entries)` | `string` | Create a git tree |
| `CreateBlob(ctx, owner, repo, content, encoding)` | `string` | Create a git blob |

Together with `GetRef`, `CreateCommit`, and `UpdateRef`, these support building an atomic multi-file
commit: create a blob per file, assemble a tree from those blobs, create a commit pointing at the
new tree, then update the branch ref to the new commit.

### Pull Requests

| Method | Returns | Description |
|--------|---------|-------------|
| `GetPullRequest(ctx, owner, repo, number)` | `*gogithub.PullRequest` | Get PR details |
| `ListPullRequests(ctx, owner, repo, opts)` | `[]*gogithub.PullRequest` | List pull requests |
| `CreatePullRequest(ctx, owner, repo, input)` | `*gogithub.PullRequest` | Create a new PR |
| `UpdatePullRequest(ctx, owner, repo, num, input)` | `*gogithub.PullRequest` | Update a PR |
| `MergePullRequest(ctx, owner, repo, num, opts)` | `*gogithub.MergeResult` | Merge a PR |
| `ListPullRequestFiles(ctx, owner, repo, num)` | `[]*gogithub.CommitFile` | List changed files |
| `GetPullRequestDiff(ctx, owner, repo, num)` | `string` | Get PR diff |
| `GetPullRequestPatch(ctx, owner, repo, num)` | `string` | Get PR patch |

### Pull Request Reviews

| Method | Returns | Description |
|--------|---------|-------------|
| `CreatePullRequestReview(ctx, owner, repo, num, input)` | `*gogithub.PullRequestReview` | Create a review |
| `ListPullRequestReviews(ctx, owner, repo, num)` | `[]*gogithub.PullRequestReview` | List reviews |
| `RequestReviewers(ctx, owner, repo, num, users, teams)` | `*gogithub.PullRequest` | Request reviewers |

### Pull Request Comments

| Method | Returns | Description |
|--------|---------|-------------|
| `CreatePullRequestComment(ctx, owner, repo, num, input)` | `*gogithub.PullRequestComment` | Create a diff comment |
| `ListPullRequestComments(ctx, owner, repo, num)` | `[]*gogithub.PullRequestComment` | List diff comments |

### Issues

| Method | Returns | Description |
|--------|---------|-------------|
| `GetIssue(ctx, owner, repo, number)` | `*gogithub.Issue` | Get issue details |
| `ListIssues(ctx, owner, repo, opts)` | `[]*gogithub.Issue` | List issues |
| `CreateIssue(ctx, owner, repo, input)` | `*gogithub.Issue` | Create a new issue |
| `UpdateIssue(ctx, owner, repo, number, input)` | `*gogithub.Issue` | Update an issue |
| `CreateIssueComment(ctx, owner, repo, num, body)` | `*gogithub.IssueComment` | Create an issue/PR comment |

`gogithub.Issue` includes an `IsPullRequest` field, since GitHub's issue-listing endpoints also
return pull requests — check it to filter PRs out of issue results.

### Checks

| Method | Returns | Description |
|--------|---------|-------------|
| `GetCheckRun(ctx, owner, repo, id)` | `*gogithub.CheckRun` | Get a check run by ID |
| `ListCheckRuns(ctx, owner, repo, ref)` | `[]*gogithub.CheckRun` | List check runs for a ref |
| `ListCheckSuites(ctx, owner, repo, ref)` | `[]*gogithub.CheckSuite` | List check suites for a ref |

### Releases

| Method | Returns | Description |
|--------|---------|-------------|
| `GetRelease(ctx, owner, repo, id)` | `*gogithub.Release` | Get release by ID |
| `GetLatestRelease(ctx, owner, repo)` | `*gogithub.Release` | Get latest release |
| `GetReleaseByTag(ctx, owner, repo, tag)` | `*gogithub.Release` | Get release by tag |
| `ListReleases(ctx, owner, repo)` | `[]*gogithub.Release` | List all releases |
| `CreateRelease(ctx, owner, repo, input)` | `*gogithub.Release` | Create a release |
| `UpdateRelease(ctx, owner, repo, id, input)` | `*gogithub.Release` | Update a release |
| `DeleteRelease(ctx, owner, repo, id)` | `error` | Delete a release |
| `ListReleaseAssets(ctx, owner, repo, id)` | `[]*gogithub.ReleaseAsset` | List release assets |

### Search

| Method | Returns | Description |
|--------|---------|-------------|
| `SearchIssues(ctx, query, opts)` | `*gogithub.IssueSearchResult` | Search issues/PRs |
| `SearchCode(ctx, query, opts)` | `*gogithub.CodeSearchResult` | Search code |

### Contributors

| Method | Returns | Description |
|--------|---------|-------------|
| `GetContributorStats(ctx, owner, repo)` | `[]*gogithub.ContributorStats` | Get contributor statistics |

### Activity

| Method | Returns | Description |
|--------|---------|-------------|
| `ListUserEvents(ctx, username, opts)` | `[]*gogithub.Event` | List activity events performed by a user (their public timeline) |

GitHub's Events API only returns the most recent ~300 events regardless of pagination. Set
`opts.PublicOnly` to `false` (the default) to include private events when authenticated as
`username`, or `true` to restrict to public events.

## Stable Types

All types are defined in the root `gogithub` package:

```go
import "github.com/grokify/gogithub"

// Core types
var user *gogithub.User
var repo *gogithub.Repository
var ref *gogithub.Reference
var branch *gogithub.Branch
var tag *gogithub.Tag
var commit *gogithub.Commit

// Pull requests
var pr *gogithub.PullRequest
var review *gogithub.PullRequestReview
var prComment *gogithub.PullRequestComment
var issueComment *gogithub.IssueComment
var commitFile *gogithub.CommitFile
var mergeResult *gogithub.MergeResult

// CI/CD
var checkRun *gogithub.CheckRun
var checkSuite *gogithub.CheckSuite

// Releases
var release *gogithub.Release
var asset *gogithub.ReleaseAsset

// Content
var fileContent *gogithub.FileContent
var contentOpts *gogithub.ContentOptions
var createFileResult *gogithub.CreateFileResult
var deleteFileResult *gogithub.DeleteFileResult

// Git data
var treeNode *gogithub.TreeNode

// Issues
var issue *gogithub.Issue

// Search
var searchResult *gogithub.IssueSearchResult
var codeSearchResult *gogithub.CodeSearchResult
var codeResult *gogithub.CodeResult

// Stats
var stats *gogithub.ContributorStats

// Activity
var event *gogithub.Event
var eventRepo *gogithub.EventRepo
```

## Escape Hatch

For advanced use cases not yet wrapped, use `Raw()` to access the underlying go-github client:

```go
import "github.com/google/go-github/v89/github"

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
    "github.com/google/go-github/v89/github"
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
