# Releases

The `release` package provides functions for listing releases and downloading release assets. All
functions take a [`clientv1.Client`](clientv1.md) and return stable `gogithub.*` types.

## Listing Releases

### List All Releases

```go
import "github.com/grokify/gogithub/release"

releases, err := release.ListReleases(ctx, client, "owner", "repo")
for _, r := range releases {
    fmt.Printf("%s: %s\n", r.TagName, r.Name)
}
```

### List Releases Since an ID

```go
releases, err := release.ListReleasesSince(ctx, client, "owner", "repo", sinceReleaseID)
```

### Get Latest Release

```go
latest, err := release.GetLatestRelease(ctx, client, "owner", "repo")
fmt.Printf("Latest version: %s\n", latest.TagName)
if latest.PublishedAt != nil {
    fmt.Printf("Published: %s\n", latest.PublishedAt.Format("2006-01-02"))
}
```

## Release Assets

### List Assets for a Release

```go
// First get the release
latest, _ := release.GetLatestRelease(ctx, client, "owner", "repo")

// List its assets
assets, err := release.ListReleaseAssets(ctx, client, "owner", "repo", latest.ID)
for _, asset := range assets {
    fmt.Printf("  %s (%d bytes)\n", asset.Name, asset.Size)
}
```

### Download an Asset

Assets can be downloaded via their browser download URL:

```go
for _, asset := range assets {
    if asset.Name == "myapp-linux-amd64.tar.gz" {
        fmt.Printf("Download URL: %s\n", asset.BrowserDownloadURL)
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

    "github.com/grokify/gogithub/clientv1"
    "github.com/grokify/gogithub/release"
)

func main() {
    ctx := context.Background()
    client, err := clientv1.NewClient(ctx, "your-token")
    if err != nil {
        panic(err)
    }

    releases, err := release.ListReleases(ctx, client, "cli", "cli")
    if err != nil {
        panic(err)
    }

    for _, r := range releases {
        fmt.Printf("\n%s - %s\n", r.TagName, r.Name)
        if r.PublishedAt != nil {
            fmt.Printf("  Published: %s\n", r.PublishedAt.Format("2006-01-02"))
        }

        if r.Prerelease {
            fmt.Println("  (prerelease)")
        }

        assets, _ := release.ListReleaseAssets(ctx, client, "cli", "cli", r.ID)
        for _, asset := range assets {
            fmt.Printf("  - %s (%d KB)\n", asset.Name, asset.Size/1024)
        }
    }
}
```

## API Reference

See [pkg.go.dev/github.com/grokify/gogithub/release](https://pkg.go.dev/github.com/grokify/gogithub/release) for complete API documentation.
