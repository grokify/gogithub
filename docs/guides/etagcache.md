# Conditional Requests

The `etagcache` package provides an HTTP transport that makes conditional
requests against the GitHub API. GitHub returns an `ETag` with most `GET`
responses and answers a repeat request carrying `If-None-Match` with
`304 Not Modified` when nothing changed. **A 304 does not count against the
rate limit.** For anything that polls — a dashboard refreshing repository
health every few minutes, a tool that re-reads the same repositories each
run — most responses are unchanged, so most requests become free.

```go
import (
    "github.com/grokify/gogithub/clientv1"
    "github.com/grokify/gogithub/etagcache"
)

cache := etagcache.NewTransport(nil)
client, err := clientv1.NewClientWithOptions(ctx, clientv1.ClientOptions{
    Token:     token,
    Transport: cache,
})

// ... use the client as usual, then:
fmt.Println(cache.Stats()) // "412 requests, 397 served from cache, 15 fetched, 0 uncacheable"
```

## How it works

The transport sits below the token: `clientv1` attaches `Authorization`, then
the transport sees the request.

1. On a `GET` or `HEAD`, it looks up the URL in its store. With an entry, it
   adds `If-None-Match` (and `If-Modified-Since` when known) to the request.
2. On `304 Not Modified`, it returns the cached body and headers as a
   `200 OK`, so go-github and every gogithub package see an ordinary success.
   Rate limit headers, `Date`, and the request ID come from the 304, so
   `GetRateLimit` and error messages stay accurate.
3. On `200 OK` with an `ETag`, it stores the body and headers (including the
   `Link` pagination header) and returns the response.

Requests with other methods, non-200 responses, and responses without an
`ETag` pass through untouched. A request that already carries its own
`If-None-Match` is sent as is.

Entries are keyed by method, URL, and a hash of the `Authorization` header,
so a store shared by clients with different tokens never serves one token's
response to another. The token itself is never stored.

## Stores

| Store | Use |
|-------|-----|
| `NewMemoryStore(maxEntries)` | Long-running processes. Least-recently-used eviction when `maxEntries` is exceeded; `0` means unlimited. `NewTransport` uses an unlimited one. |
| `NewFileStore(dir)` | Command-line tools, so the cache survives between runs. One JSON file per entry, directory `0700`, files `0600`, written atomically. No eviction; remove the directory to clear it. |

```go
store, err := etagcache.NewFileStore(filepath.Join(os.UserCacheDir(), "myapp", "github"))
cache := etagcache.NewTransportWithStore(nil, store)
```

Implement `Store` for anything else (Redis, a database). Store failures never
fail a request: the transport proceeds without the cache and calls
`Transport.OnStoreError`, if set.

Cached bodies contain whatever the API returned, including private
repository data. Treat the store accordingly.

## What to expect

For [repository health](health.md), the four requests per repository all
return `ETag`s, so a refresh in which nothing changed costs zero rate limit.
The requests are still made — a 304 is a round trip — so wall-clock time is
similar; what the cache buys is rate limit headroom and, for large bodies,
bandwidth.

GitHub's documentation on
[conditional requests](https://docs.github.com/en/rest/using-the-rest-api/best-practices-for-using-the-rest-api#use-conditional-requests-if-appropriate)
describes the behavior the transport relies on.

## Command line

`gogithub health --cache-dir <dir>` uses a `FileStore` in that directory and
prints the hit statistics to stderr after the tables. See the
[CLI guide](cli.md#health).
