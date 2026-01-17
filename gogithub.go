// Package gogithub provides a Go client for the GitHub API.
//
// This package is organized into subpackages by operation type:
//   - auth: Authentication utilities and bot user constants
//   - config: Configuration and environment variable loading
//   - errors: Error types and translation utilities
//   - pathutil: Path validation and normalization
//   - search: Search API (issues, PRs, code, etc.)
//   - repo: Repository operations (fork, branch, commit, batch)
//   - pr: Pull request operations
//   - release: Release and asset operations
//
// Example usage:
//
//	package main
//
//	import (
//	    "context"
//	    "fmt"
//
//	    "github.com/grokify/gogithub/auth"
//	    "github.com/grokify/gogithub/search"
//	)
//
//	func main() {
//	    ctx := context.Background()
//	    gh := auth.NewGitHubClient(ctx, "your-token")
//
//	    client := search.NewClient(gh)
//	    issues, err := client.SearchIssuesAll(ctx, search.Query{
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
