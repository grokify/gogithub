package gogithub

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/go-github/v56/github"
)

type Client struct {
	*github.Client
}

func NewClient(httpClient *http.Client) *Client {
	c := github.NewClient(httpClient)
	return &Client{Client: c}
}

func (c *Client) SearchOpenPullRequests(ctx context.Context, username string, opts *github.SearchOptions) (*github.IssuesSearchResult, *github.Response, error) {
	qry := Query{
		"user":  username,
		"state": "open",
		"is":    "pr",
	}
	return c.SearchIssues(ctx, qry, opts)
}

func (c *Client) SearchIssues(ctx context.Context, qry Query, opts *github.SearchOptions) (*github.IssuesSearchResult, *github.Response, error) {
	return c.Client.Search.Issues(ctx, qry.Encode(), opts)
}

type Query map[string]string

func (q Query) Encode() string {
	parts := []string{}
	for k, v := range q {
		parts = append(parts, k+":"+v)
	}
	return strings.Join(parts, " ")
}

// client := github.NewClient(nil)
// func (s *SearchService) Issues(ctx context.Context, query string, opts *SearchOptions) (*IssuesSearchResult, *Response, error)
