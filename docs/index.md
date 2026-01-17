# GoGitHub

[![Build Status](https://github.com/grokify/gogithub/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/grokify/gogithub/actions/workflows/ci.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/grokify/gogithub)](https://goreportcard.com/report/github.com/grokify/gogithub)
[![Docs](https://pkg.go.dev/badge/github.com/grokify/gogithub)](https://pkg.go.dev/github.com/grokify/gogithub)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/grokify/gogithub/blob/master/LICENSE)

**GoGitHub** is a high-level Go module for interacting with the GitHub API. It wraps [google/go-github](https://github.com/google/go-github) with convenience functions organized by operation type.

## Features

- **Search API** - Query issues, pull requests, code, and commits with a fluent query builder
- **Repository Operations** - Fork, branch, commit, and batch file operations
- **Pull Requests** - Create, list, merge, and manage PRs
- **Releases** - List releases and download assets
- **GraphQL API** - User contribution statistics and detailed commit stats
- **Error Handling** - Typed errors with helper functions for common cases
- **GitHub Enterprise** - Full support for GitHub Enterprise Server

## Quick Example

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
    gh := auth.NewGitHubClient(ctx, "your-github-token")

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

## Package Overview

| Package | Description |
|---------|-------------|
| [`auth`](guides/auth.md) | Authentication and client creation |
| [`config`](guides/auth.md#configuration) | Configuration from environment variables |
| [`search`](guides/search.md) | Search API with query builder |
| [`repo`](guides/repo.md) | Repository operations (fork, branch, commit, batch) |
| [`pr`](guides/pr.md) | Pull request operations |
| [`release`](guides/release.md) | Release and asset operations |
| [`graphql`](guides/graphql.md) | GraphQL API for contribution statistics |
| [`errors`](guides/errors.md) | Error types and translation |

## Next Steps

- [Getting Started](getting-started.md) - Installation and first steps
- [API Reference](api-reference.md) - Links to pkg.go.dev documentation
- [Changelog](changelog.md) - Version history and release notes
