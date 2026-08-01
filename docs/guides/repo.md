# Repository Operations

The `repo` package provides high-level functions for repository operations including forking,
branching, committing, and batch file operations. All functions take a
[`clientv1.Client`](clientv1.md) and return stable `gogithub.*` types.

## Listing Repositories

### List User Repositories

```go
import "github.com/grokify/gogithub/repo"

repos, err := repo.ListUserRepos(ctx, client, "octocat")
for _, r := range repos {
    fmt.Printf("%s: %s\n", r.FullName, r.Description)
}
```

### List Organization Repositories

```go
repos, err := repo.ListOrgRepos(ctx, client, "github")
```

### Get Single Repository

```go
repository, err := repo.GetRepo(ctx, client, "octocat", "hello-world")
```

## Branch Operations

### Get Branch SHA

```go
sha, err := repo.GetBranchSHA(ctx, client, "owner", "repo", "main")
fmt.Printf("Branch SHA: %s\n", sha)
```

### Create Branch

```go
// First get the SHA of the base branch
sha, err := repo.GetBranchSHA(ctx, client, "owner", "repo", "main")
if err != nil {
    return err
}

// Create new branch from that SHA
err = repo.CreateBranch(ctx, client, "owner", "repo", "feature-branch", sha)
```

### Delete Branch

```go
err := repo.DeleteBranch(ctx, client, "owner", "repo", "feature-branch")
```

### Get Default Branch

```go
defaultBranch, err := repo.GetDefaultBranch(ctx, client, "owner", "repo")
fmt.Printf("Default branch: %s\n", defaultBranch)
```

## Fork Operations

### Ensure Fork Exists

`EnsureFork` creates a fork if it doesn't exist, or returns the existing fork's owner/repo names:

```go
forkOwner, forkRepo, err := repo.EnsureFork(ctx, client, "upstream-owner", "upstream-repo", "your-username")
fmt.Printf("Fork: %s/%s\n", forkOwner, forkRepo)
```

## Commit Operations

### Create Single Commit

Create a commit with multiple files using the Git Data API (tree-based). Returns the new commit's SHA:

```go
files := []repo.FileContent{
    {Path: "README.md", Content: []byte("# Hello World")},
    {Path: "src/main.go", Content: []byte("package main\n\nfunc main() {}")},
}

commitSHA, err := repo.CreateCommit(ctx, client, "owner", "repo", "feature-branch", "Add initial files", files)
fmt.Printf("Commit SHA: %s\n", commitSHA)
```

### Read Local Files

Helper to read files from the local filesystem, prepending `prefix` to each destination path:

```go
files, err := repo.ReadLocalFiles("path/to/dir", "")
if err != nil {
    return err
}

commitSHA, err := repo.CreateCommit(ctx, client, "owner", "repo", "branch", "Add files", files)
```

## Batch Operations

The `Batch` type allows atomic multi-file commits with queued operations. The commit message is
set when the batch is created:

### Create a Batch

```go
batch, err := repo.NewBatch(ctx, client, "owner", "repo", "feature-branch", "Update documentation")
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
commitSHA, err := batch.Commit(ctx)
if err != nil {
    return err
}
fmt.Printf("Committed: %s\n", commitSHA)
```

### With Custom Author

```go
batch, err := repo.NewBatch(ctx, client, "owner", "repo", "branch", "Update as bot",
    repo.WithCommitAuthor("Bot", "bot@example.com"),
)
```

### Full Example

```go
batch, err := repo.NewBatch(ctx, client, "owner", "repo", "main", "Refactor configuration")
if err != nil {
    return err
}

// Queue multiple operations
batch.Write("config.json", []byte(`{"version": 2}`))
batch.Write("src/app.go", []byte("package main"))
batch.Delete("deprecated.txt")

// Commit atomically
commitSHA, err := batch.Commit(ctx)
if err != nil {
    return err
}

fmt.Printf("All changes committed in: %s\n", commitSHA)
```

!!! warning "Batch Commits Are Single-Use"
    A `Batch` can only be committed once (check with `batch.Committed()`). After calling `Commit()`, create a new `Batch` for additional changes.

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
commitSHA, err := repo.CreateCommit(ctx, client, owner, repo, branch, msg, files)
if err != nil {
    var commitErr *repo.CommitError
    if errors.As(err, &commitErr) {
        fmt.Printf("Commit failed (%s): %v\n", commitErr.Message, commitErr.Err)
    }
}
```

### BatchError

```go
commitSHA, err := batch.Commit(ctx)
if err != nil {
    var batchErr *repo.BatchError
    if errors.As(err, &batchErr) {
        fmt.Printf("Batch %s failed: %v\n", batchErr.Op, batchErr.Err)
    }
}
```

## API Reference

See [pkg.go.dev/github.com/grokify/gogithub/repo](https://pkg.go.dev/github.com/grokify/gogithub/repo) for complete API documentation.
