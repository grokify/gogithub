# Repository Operations

The `repo` package provides high-level functions for repository operations including forking, branching, committing, and batch file operations.

## Listing Repositories

### List User Repositories

```go
import "github.com/grokify/gogithub/repo"

repos, err := repo.ListUserRepos(ctx, gh, "octocat", nil)
for _, r := range repos {
    fmt.Printf("%s: %s\n", r.GetFullName(), r.GetDescription())
}
```

### List Organization Repositories

```go
repos, err := repo.ListOrgRepos(ctx, gh, "github", nil)
```

### Get Single Repository

```go
repository, err := repo.GetRepo(ctx, gh, "octocat", "hello-world")
```

## Branch Operations

### Get Branch SHA

```go
sha, err := repo.GetBranchSHA(ctx, gh, "owner", "repo", "main")
fmt.Printf("Branch SHA: %s\n", sha)
```

### Create Branch

```go
// First get the SHA of the base branch
sha, err := repo.GetBranchSHA(ctx, gh, "owner", "repo", "main")
if err != nil {
    return err
}

// Create new branch from that SHA
err = repo.CreateBranch(ctx, gh, "owner", "repo", "feature-branch", sha)
```

### Delete Branch

```go
err := repo.DeleteBranch(ctx, gh, "owner", "repo", "feature-branch")
```

### Get Default Branch

```go
defaultBranch, err := repo.GetDefaultBranch(ctx, gh, "owner", "repo")
fmt.Printf("Default branch: %s\n", defaultBranch)
```

## Fork Operations

### Ensure Fork Exists

`EnsureFork` creates a fork if it doesn't exist, or returns the existing fork:

```go
fork, err := repo.EnsureFork(ctx, gh, "upstream-owner", "upstream-repo")
fmt.Printf("Fork: %s/%s\n", fork.GetOwner().GetLogin(), fork.GetName())
```

## Commit Operations

### Create Single Commit

Create a commit with multiple files using the Git Data API (tree-based):

```go
files := []repo.FileContent{
    {Path: "README.md", Content: []byte("# Hello World")},
    {Path: "src/main.go", Content: []byte("package main\n\nfunc main() {}")},
}

commit, err := repo.CreateCommit(ctx, gh, "owner", "repo", "feature-branch", "Add initial files", files)
fmt.Printf("Commit SHA: %s\n", commit.GetSHA())
```

### Read Local Files

Helper to read files from the local filesystem:

```go
files, err := repo.ReadLocalFiles("path/to/dir", []string{"file1.txt", "file2.txt"})
if err != nil {
    return err
}

commit, err := repo.CreateCommit(ctx, gh, "owner", "repo", "branch", "Add files", files)
```

## Batch Operations

The `Batch` type allows atomic multi-file commits with queued operations:

### Create a Batch

```go
batch, err := repo.NewBatch(ctx, gh, "owner", "repo", "feature-branch")
if err != nil {
    return err
}
```

### Queue Operations

```go
// Queue file writes
err = batch.Write("README.md", []byte("# Updated README"))
err = batch.Write("docs/guide.md", []byte("# Guide"))

// Queue file deletions
err = batch.Delete("old-file.txt")
```

### Commit All Changes

```go
commit, err := batch.Commit(ctx, "Update documentation")
if err != nil {
    return err
}
fmt.Printf("Committed: %s\n", commit.GetSHA())
```

### With Custom Author

```go
batch, err := repo.NewBatch(ctx, gh, "owner", "repo", "branch",
    repo.WithCommitAuthor("Bot", "bot@example.com"),
)
```

### Full Example

```go
batch, err := repo.NewBatch(ctx, gh, "owner", "repo", "main")
if err != nil {
    return err
}

// Queue multiple operations
batch.Write("config.json", []byte(`{"version": 2}`))
batch.Write("src/app.go", []byte("package main"))
batch.Delete("deprecated.txt")

// Commit atomically
commit, err := batch.Commit(ctx, "Refactor configuration")
if err != nil {
    return err
}

fmt.Printf("All changes committed in: %s\n", commit.GetSHA())
```

!!! warning "Batch Commits Are Single-Use"
    A `Batch` can only be committed once. After calling `Commit()`, create a new `Batch` for additional changes.

## Path Validation

The `pathutil` package (used internally) validates and normalizes file paths:

```go
import "github.com/grokify/gogithub/pathutil"

// Validate path (rejects traversal attempts)
err := pathutil.Validate("../etc/passwd") // Returns error

// Normalize path
normalized := pathutil.Normalize("/foo//bar/./baz")  // "foo/bar/baz"

// Join paths safely
path := pathutil.Join("dir", "subdir", "file.txt")  // "dir/subdir/file.txt"
```

## Error Types

### CommitError

```go
commit, err := repo.CreateCommit(ctx, gh, owner, repo, branch, msg, files)
if err != nil {
    var commitErr *repo.CommitError
    if errors.As(err, &commitErr) {
        fmt.Printf("Commit failed for %s/%s: %v\n",
            commitErr.Owner, commitErr.Repo, commitErr.Err)
    }
}
```

### BatchError

```go
commit, err := batch.Commit(ctx, message)
if err != nil {
    var batchErr *repo.BatchError
    if errors.As(err, &batchErr) {
        fmt.Printf("Batch commit failed: %s\n", batchErr.Message)
    }
}
```

## API Reference

See [pkg.go.dev/github.com/grokify/gogithub/repo](https://pkg.go.dev/github.com/grokify/gogithub/repo) for complete API documentation.
