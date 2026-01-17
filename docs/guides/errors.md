# Error Handling

The `errors` package provides typed errors and helper functions for handling GitHub API errors.

## APIError Type

The `APIError` type wraps GitHub API errors with additional context:

```go
import "github.com/grokify/gogithub/errors"

type APIError struct {
    StatusCode int
    Message    string
    Err        error  // Original error
}
```

### Checking Error Types

Use the helper functions to check for specific error conditions:

```go
_, err := repo.GetRepo(ctx, gh, "owner", "nonexistent-repo")
if err != nil {
    switch {
    case errors.IsNotFound(err):
        fmt.Println("Repository not found")
    case errors.IsPermissionDenied(err):
        fmt.Println("Access denied")
    case errors.IsRateLimited(err):
        fmt.Println("Rate limit exceeded, try again later")
    default:
        fmt.Printf("Error: %v\n", err)
    }
}
```

## Helper Functions

| Function | Status Codes | Description |
|----------|--------------|-------------|
| `IsNotFound(err)` | 404 | Resource doesn't exist |
| `IsPermissionDenied(err)` | 401, 403 | Authentication or authorization failed |
| `IsRateLimited(err)` | 403 (with rate limit) | API rate limit exceeded |
| `IsConflict(err)` | 409 | Resource conflict (e.g., branch already exists) |
| `IsValidation(err)` | 422 | Validation error (invalid input) |
| `IsServerError(err)` | 500, 502, 503 | GitHub server error |

## Translating Errors

The `Translate` function converts GitHub API errors to `APIError`:

```go
_, resp, err := gh.Repositories.Get(ctx, "owner", "repo")
if err != nil {
    apiErr := errors.Translate(err, resp)
    fmt.Printf("Status: %d, Message: %s\n", apiErr.StatusCode, apiErr.Message)
}
```

## Getting Status Code

Extract the status code from any error:

```go
code := errors.StatusCode(err)
if code == 404 {
    // Handle not found
}
```

Returns 0 if the error is not an `APIError`.

## Error Unwrapping

All error types support Go 1.13+ error unwrapping:

```go
var apiErr *errors.APIError
if errors.As(err, &apiErr) {
    fmt.Printf("Status: %d\n", apiErr.StatusCode)

    // Access the original error
    originalErr := apiErr.Unwrap()
}
```

## Package-Specific Errors

Each package defines its own error types for specific operations:

### auth.AuthError

```go
username, err := auth.GetAuthenticatedUser(ctx, gh)
if err != nil {
    var authErr *auth.AuthError
    if errors.As(err, &authErr) {
        fmt.Printf("Authentication failed: %s\n", authErr.Message)
    }
}
```

### repo.CommitError

```go
commit, err := repo.CreateCommit(ctx, gh, owner, repo, branch, msg, files)
if err != nil {
    var commitErr *repo.CommitError
    if errors.As(err, &commitErr) {
        fmt.Printf("Commit to %s/%s failed: %v\n",
            commitErr.Owner, commitErr.Repo, commitErr.Err)
    }
}
```

### repo.BatchError

```go
commit, err := batch.Commit(ctx, message)
if err != nil {
    var batchErr *repo.BatchError
    if errors.As(err, &batchErr) {
        fmt.Printf("Batch operation failed: %s\n", batchErr.Message)
    }
}
```

## Best Practices

### 1. Check Specific Errors First

```go
if errors.IsNotFound(err) {
    // Handle 404 specifically
    return createResource()
}
if errors.IsRateLimited(err) {
    // Wait and retry
    time.Sleep(time.Minute)
    return retry()
}
// Handle other errors
return err
```

### 2. Use Error Wrapping

```go
if err != nil {
    return fmt.Errorf("failed to create PR for %s/%s: %w", owner, repo, err)
}
```

### 3. Handle Rate Limits Gracefully

```go
for retries := 0; retries < 3; retries++ {
    result, err := doOperation()
    if err == nil {
        return result, nil
    }
    if errors.IsRateLimited(err) {
        time.Sleep(time.Duration(retries+1) * time.Minute)
        continue
    }
    return nil, err
}
```

## API Reference

See [pkg.go.dev/github.com/grokify/gogithub/errors](https://pkg.go.dev/github.com/grokify/gogithub/errors) for complete API documentation.
