# Releases

The `release` package provides functions for listing releases and downloading release assets.

## Listing Releases

### List All Releases

```go
import "github.com/grokify/gogithub/release"

releases, err := release.ListReleases(ctx, gh, "owner", "repo", nil)
for _, r := range releases {
    fmt.Printf("%s: %s\n", r.GetTagName(), r.GetName())
}
```

### With Pagination Options

```go
opts := &github.ListOptions{
    PerPage: 10,
    Page:    1,
}

releases, err := release.ListReleases(ctx, gh, "owner", "repo", opts)
```

### Get Latest Release

```go
latest, err := release.GetLatestRelease(ctx, gh, "owner", "repo")
fmt.Printf("Latest version: %s\n", latest.GetTagName())
fmt.Printf("Published: %s\n", latest.GetPublishedAt())
```

## Release Assets

### List Assets for a Release

```go
// First get the release
latest, _ := release.GetLatestRelease(ctx, gh, "owner", "repo")

// List its assets
assets, err := release.ListReleaseAssets(ctx, gh, "owner", "repo", latest.GetID(), nil)
for _, asset := range assets {
    fmt.Printf("  %s (%d bytes)\n", asset.GetName(), asset.GetSize())
}
```

### Download an Asset

Assets can be downloaded via their browser download URL:

```go
for _, asset := range assets {
    if asset.GetName() == "myapp-linux-amd64.tar.gz" {
        fmt.Printf("Download URL: %s\n", asset.GetBrowserDownloadURL())
    }
}
```

## Complete Example

List all releases with their assets:

```go
package main

import (
    "context"
    "fmt"

    "github.com/grokify/gogithub/auth"
    "github.com/grokify/gogithub/release"
)

func main() {
    ctx := context.Background()
    gh := auth.NewGitHubClient(ctx, "your-token")

    releases, err := release.ListReleases(ctx, gh, "cli", "cli", nil)
    if err != nil {
        panic(err)
    }

    for _, r := range releases {
        fmt.Printf("\n%s - %s\n", r.GetTagName(), r.GetName())
        fmt.Printf("  Published: %s\n", r.GetPublishedAt().Format("2006-01-02"))

        if r.GetPrerelease() {
            fmt.Println("  (prerelease)")
        }

        assets, _ := release.ListReleaseAssets(ctx, gh, "cli", "cli", r.GetID(), nil)
        for _, asset := range assets {
            fmt.Printf("  - %s (%d KB)\n", asset.GetName(), asset.GetSize()/1024)
        }
    }
}
```

## API Reference

See [pkg.go.dev/github.com/grokify/gogithub/release](https://pkg.go.dev/github.com/grokify/gogithub/release) for complete API documentation.
