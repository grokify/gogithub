# Release Notes v0.5.0

This release adds a `QueryBuilder` for type-safe, fluent query construction while preserving the flexible `Query` map type for forward compatibility with new GitHub search qualifiers.

## New Features

### QueryBuilder for Type-Safe Query Construction

The new `QueryBuilder` provides a fluent API for building search queries with IDE autocomplete and compile-time validation:

```go
// Fluent builder pattern
qry := search.NewQuery().
    User("grokify").
    StateOpen().
    IsPR().
    Build()

// Use with search client
results, err := client.SearchIssuesAll(ctx, qry, nil)
```

### Available Builder Methods

| Method | Description |
|--------|-------------|
| `User(username)` | Filter by repository owner |
| `Org(org)` | Filter by organization |
| `Repo(owner/repo)` | Filter by specific repository |
| `State(state)` | Filter by state (open, closed) |
| `StateOpen()` | Shorthand for open state |
| `StateClosed()` | Shorthand for closed state |
| `Is(value)` | Filter by type/state (pr, issue, merged, etc.) |
| `IsPR()` | Shorthand for pull requests |
| `IsIssue()` | Shorthand for issues |
| `Type(type)` | Filter by type |
| `Author(username)` | Filter by creator |
| `Assignee(username)` | Filter by assignee |
| `Label(label)` | Filter by label |
| `Mentions(username)` | Filter by mentioned user |
| `Involves(username)` | Filter by user involvement |
| `Set(key, value)` | Escape hatch for any qualifier |
| `Build()` | Return the constructed `Query` map |

### Escape Hatch for New Qualifiers

The `Set()` method allows using any GitHub search qualifier, including new ones added after this SDK release:

```go
// Use new/uncommon qualifiers via Set()
qry := search.NewQuery().
    User("grokify").
    Set("reason", "completed").
    Set("review", "approved").
    Build()
```

### Flexible Query Map Still Available

The original `Query` map type remains unchanged for maximum flexibility:

```go
// Direct map construction still works
qry := search.Query{
    "user":   "grokify",
    "state":  "open",
    "is":     "pr",
    "reason": "completed",  // New qualifier - no SDK update needed
}
```

## Breaking Changes

### Removed `NewClientHTTP`

`search.NewClientHTTP()` has been removed. Use `search.NewClient()` with an explicit GitHub client instead.

| Before | After |
|--------|-------|
| `search.NewClientHTTP(httpClient)` | `search.NewClient(github.NewClient(httpClient))` |
| `search.NewClientHTTP(nil)` | `search.NewClient(github.NewClient(nil))` |

**Migration example:**

```go
// Before (v0.4.x)
client := search.NewClientHTTP(nil)

// After (v0.5.0)
import "github.com/google/go-github/v81/github"

client := search.NewClient(github.NewClient(nil))
```

## Migration Guide

### Updating Query Construction (Optional)

Existing code using `Query` map literals continues to work unchanged:

```go
// This still works - no migration required
qry := search.Query{
    search.ParamUser:  "grokify",
    search.ParamState: search.ParamStateValueOpen,
    search.ParamIs:    search.ParamIsValuePR,
}
```

You can optionally migrate to `QueryBuilder` for better IDE support:

```go
// Optional: migrate to QueryBuilder
qry := search.NewQuery().User("grokify").StateOpen().IsPR().Build()
```

### Updating Client Creation (Required)

Replace `NewClientHTTP` with explicit client creation:

```go
// Before
client := search.NewClientHTTP(myHTTPClient)

// After
client := search.NewClient(github.NewClient(myHTTPClient))
```

For authenticated clients, use the `auth` package:

```go
import "github.com/grokify/gogithub/auth"

gh := auth.NewGitHubClient(ctx, token)
client := search.NewClient(gh)
```
