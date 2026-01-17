// Package gogithub provides a Go client for the GitHub API.
//
// This package is organized into subpackages by operation type:
//   - auth: Authentication utilities
//   - search: Search API (issues, PRs, code, etc.)
//   - repo: Repository operations (fork, branch, commit)
//   - pr: Pull request operations
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

import (
	"context"
	"net/http"

	"github.com/google/go-github/v81/github"
	"github.com/grokify/gogithub/auth"
	"github.com/grokify/gogithub/search"
)

// Client wraps the GitHub client with convenience methods.
// For new code, prefer using the subpackages directly.
type Client struct {
	*github.Client
	Search *search.Client
}

// NewClient creates a new client from an HTTP client.
// Deprecated: Use auth.NewGitHubClient and search.NewClient directly.
func NewClient(httpClient *http.Client) *Client {
	gh := github.NewClient(httpClient)
	return &Client{
		Client: gh,
		Search: search.NewClient(gh),
	}
}

// NewClientWithToken creates a new client authenticated with a token.
func NewClientWithToken(ctx context.Context, token string) *Client {
	gh := auth.NewGitHubClient(ctx, token)
	return &Client{
		Client: gh,
		Search: search.NewClient(gh),
	}
}

// Re-export types for backward compatibility.
// Deprecated: Import from subpackages directly.
type (
	Query  = search.Query
	Issues = search.Issues
	Issue  = search.Issue
)

// Re-export constants for backward compatibility.
// Deprecated: Import from subpackages directly.
const (
	ParamUser            = search.ParamUser
	ParamState           = search.ParamState
	ParamStateValueOpen  = search.ParamStateValueOpen
	ParamIs              = search.ParamIs
	ParamIsValuePR       = search.ParamIsValuePR
	ParamPerPageValueMax = search.ParamPerPageValueMax

	UsernameDependabot = search.UsernameDependabot
	UserIDDependabot   = search.UserIDDependabot

	BaseURLRepoAPI  = search.BaseURLRepoAPI
	BaseURLRepoHTML = search.BaseURLRepoHTML
)

// SearchOpenPullRequests searches for open pull requests by username.
// Deprecated: Use search.Client.SearchOpenPullRequests directly.
func (c *Client) SearchOpenPullRequests(ctx context.Context, username string, opts *github.SearchOptions) (*github.IssuesSearchResult, *github.Response, error) {
	return c.Search.SearchOpenPullRequests(ctx, username, opts)
}

// SearchIssues is a wrapper for SearchService.Issues().
// Deprecated: Use search.Client.SearchIssues directly.
func (c *Client) SearchIssues(ctx context.Context, qry Query, opts *github.SearchOptions) (*github.IssuesSearchResult, *github.Response, error) {
	return c.Search.SearchIssues(ctx, qry, opts)
}

// SearchIssuesAll retrieves all issues matching the query with pagination.
// Deprecated: Use search.Client.SearchIssuesAll directly.
func (c *Client) SearchIssuesAll(ctx context.Context, qry Query, opts *github.SearchOptions) (Issues, error) {
	return c.Search.SearchIssuesAll(ctx, qry, opts)
}
