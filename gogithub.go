// Package gogithub provides a Go client for the GitHub API.
//
// # Recommended: Version-Isolated Client
//
// For new code, use the clientv1 package which provides stable types that won't
// change when go-github updates its major version:
//
//	import "github.com/grokify/gogithub/clientv1"
//
//	client, err := clientv1.NewClient(ctx, "your-token")
//	user, err := client.GetAuthenticatedUser(ctx)      // Returns *clientv1.User
//	repos, err := client.ListUserRepos(ctx, "user")    // Returns []*clientv1.Repository
//	sha, err := client.GetBranchSHA(ctx, "owner", "repo", "main")
//
// # Operation Packages
//
// The following packages also accept clientv1.Client and return stable gogithub.*
// types, so they stay version-isolated like the client itself:
//
//   - search: Search API (issues, PRs, code, etc.)
//   - repo: Repository operations (fork, branch, commit, batch)
//   - pr: Pull request operations
//   - release: Release and asset operations
//   - checks: Check run polling and status
//   - tag: Git tag operations
//   - sarif: SARIF upload for code scanning
//
// A few legacy functions in auth and config still return go-github types directly
// and are deprecated (auth.NewGitHubClient, config.Config.NewClient); prefer
// clientv1.NewClient and config.Config.NewClientV1 instead.
//
// Example:
//
//	package main
//
//	import (
//	    "context"
//	    "fmt"
//
//	    "github.com/grokify/gogithub/clientv1"
//	    "github.com/grokify/gogithub/search"
//	)
//
//	func main() {
//	    ctx := context.Background()
//	    client, err := clientv1.NewClient(ctx, "your-token")
//	    if err != nil {
//	        panic(err)
//	    }
//
//	    c := search.NewClient(client)
//	    issues, err := c.SearchIssuesAll(ctx, search.Query{
//	        search.ParamUser:  "grokify",
//	        search.ParamState: search.ParamStateValueOpen,
//	    }, nil)
//	    if err != nil {
//	        panic(err)
//	    }
//	    fmt.Printf("Found %d issues\n", len(issues))
//	}
package gogithub

// GitHub API base URLs.
const (
	// BaseURLRepoAPI is the base URL for the GitHub API repository endpoints.
	BaseURLRepoAPI = "https://api.github.com/repos"
	// BaseURLRepoHTML is the base URL for GitHub repository web pages.
	BaseURLRepoHTML = "https://github.com"
)
