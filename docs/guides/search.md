# Search API

The `search` package provides a high-level interface for GitHub's Search API with automatic pagination and a fluent query builder.

## Creating a Search Client

```go
import (
    "github.com/grokify/gogithub/auth"
    "github.com/grokify/gogithub/search"
)

ctx := context.Background()
gh := auth.NewGitHubClient(ctx, "your-token")
client := search.NewClient(gh)
```

## Building Queries

### Using Query Map

The `Query` type is a map that accepts search parameters:

```go
query := search.Query{
    search.ParamUser:  "octocat",
    search.ParamState: search.ParamStateValueOpen,
    search.ParamIs:    search.ParamIsValuePR,
}
```

### Using QueryBuilder (Recommended)

The `QueryBuilder` provides a type-safe, fluent interface:

```go
qb := search.NewQueryBuilder().
    User("octocat").
    State(search.ParamStateValueOpen).
    Is(search.ParamIsValuePR)

query := qb.Build()
```

## Searching Issues and PRs

### Search All (with Pagination)

`SearchIssuesAll` automatically handles pagination to retrieve all results:

```go
issues, err := client.SearchIssuesAll(ctx, query, nil)
if err != nil {
    return err
}

fmt.Printf("Found %d results\n", len(issues))
```

### With Search Options

```go
opts := &github.SearchOptions{
    Sort:  "created",
    Order: "desc",
}

issues, err := client.SearchIssuesAll(ctx, query, opts)
```

## Query Parameters

### Common Parameters

| Parameter | Values | Description |
|-----------|--------|-------------|
| `ParamUser` | username | Filter by user |
| `ParamOrg` | org name | Filter by organization |
| `ParamRepo` | owner/repo | Filter by repository |
| `ParamState` | `open`, `closed` | Issue/PR state |
| `ParamIs` | `issue`, `pr`, `open`, `closed`, `merged` | Type filters |
| `ParamAuthor` | username | Filter by author |
| `ParamAssignee` | username | Filter by assignee |
| `ParamLabel` | label name | Filter by label |

### QueryBuilder Methods

```go
qb := search.NewQueryBuilder().
    User("octocat").           // user:octocat
    Org("github").             // org:github
    Repo("octocat/hello").     // repo:octocat/hello
    State("open").             // state:open
    Is("pr").                  // is:pr
    Author("defunkt").         // author:defunkt
    Assignee("jlord").         // assignee:jlord
    Mentions("tpope").         // mentions:tpope
    Involves("jessepollak").   // involves:jessepollak
    Label("bug")               // label:bug
```

## Working with Results

### Issues Type

The `Issues` type (`[]*github.Issue`) provides convenience methods:

```go
issues, _ := client.SearchIssuesAll(ctx, query, nil)

// Get issue counts by repository
repoCounts := issues.RepositoryIssueCounts(true) // true = HTML URLs

// Generate a table for export
tbl, err := issues.Table("My Issues")

// Generate a table set (issues + repo summary)
ts, err := issues.TableSet()
ts.WriteXLSX("issues.xlsx")
```

### Issue Wrapper

Individual issues can be wrapped for additional methods:

```go
for _, is := range issues {
    issue := search.Issue{Issue: is}

    username, _ := issue.AuthorUsername()
    userID, _ := issue.AuthorUserID()
    created, _ := issue.CreatedTime()
    age, _ := issue.CreatedAge()

    fmt.Printf("%s by %s (%s old)\n",
        is.GetTitle(),
        username,
        age.Round(time.Hour*24),
    )
}
```

## Examples

### Find Open PRs by User

```go
qb := search.NewQueryBuilder().
    User("octocat").
    Is(search.ParamIsValuePR).
    State(search.ParamStateValueOpen)

prs, err := client.SearchIssuesAll(ctx, qb.Build(), nil)
```

### Find Issues with Label in Org

```go
qb := search.NewQueryBuilder().
    Org("github").
    Is(search.ParamIsValueIssue).
    Label("good first issue").
    State(search.ParamStateValueOpen)

issues, err := client.SearchIssuesAll(ctx, qb.Build(), nil)
```

### Find Merged PRs by Author

```go
qb := search.NewQueryBuilder().
    Author("octocat").
    Is(search.ParamIsValuePR).
    Is("merged")

mergedPRs, err := client.SearchIssuesAll(ctx, qb.Build(), nil)
```

### Export to Excel

```go
issues, _ := client.SearchIssuesAll(ctx, query, nil)

ts, err := search.Issues(issues).TableSet()
if err != nil {
    return err
}

err = ts.WriteXLSX("github-issues.xlsx")
```

## API Reference

See [pkg.go.dev/github.com/grokify/gogithub/search](https://pkg.go.dev/github.com/grokify/gogithub/search) for complete API documentation.
