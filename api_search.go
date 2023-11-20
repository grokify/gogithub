package gogithub

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/go-github/v56/github"
	"github.com/grokify/mogo/pointer"
)

type Client struct {
	*github.Client
}

func NewClient(httpClient *http.Client) *Client {
	return &Client{Client: github.NewClient(httpClient)}
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

func (c *Client) SearchIssuesAll(ctx context.Context, qry Query, opts *github.SearchOptions) (Issues, error) {
	if opts == nil {
		opts = &github.SearchOptions{
			ListOptions: github.ListOptions{
				Page:    1,
				PerPage: ParamPerPageMax,
			},
		}
	}
	iss := Issues{}
	for {
		isRes, _, err := c.SearchIssues(ctx, qry, opts)
		if err != nil {
			return iss, err
		}
		if len(isRes.Issues) > 0 {
			iss = append(iss, isRes.Issues...)
		}
		tc := ParamPerPageMax
		if isRes != nil && isRes.Total != nil {
			tc = pointer.Dereference(isRes.Total)
		}
		if tc < ParamPerPageMax {
			break
		}
		opts.ListOptions.Page++
	}
	return iss, nil
	//return c.Client.Search.Issues(ctx, qry.Encode(), opts)
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
