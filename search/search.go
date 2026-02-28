// Package search provides GitHub search API functionality.
package search

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/go-github/v84/github"
	"github.com/grokify/mogo/encoding/jsonutil"
	"github.com/grokify/mogo/errors/errorsutil"
	"github.com/grokify/mogo/io/ioutil"
	"github.com/grokify/mogo/pointer"
)

// Client wraps the GitHub client for search operations.
type Client struct {
	gh *github.Client
}

// NewClient creates a new search client.
func NewClient(ghClient *github.Client) *Client {
	return &Client{gh: ghClient}
}

// SearchOpenPullRequests searches for open pull requests by username.
func (c *Client) SearchOpenPullRequests(ctx context.Context, username string, opts *github.SearchOptions) (*github.IssuesSearchResult, *github.Response, error) {
	qry := NewQuery().User(username).StateOpen().IsPR().Build()
	return c.SearchIssues(ctx, qry, opts)
}

// SearchIssues is a wrapper for SearchService.Issues().
func (c *Client) SearchIssues(ctx context.Context, qry Query, opts *github.SearchOptions) (*github.IssuesSearchResult, *github.Response, error) {
	return c.gh.Search.Issues(ctx, qry.Encode(), opts)
}

// SearchIssuesAll retrieves all issues matching the query with pagination.
func (c *Client) SearchIssuesAll(ctx context.Context, qry Query, opts *github.SearchOptions) (Issues, error) {
	if opts == nil {
		opts = &github.SearchOptions{
			ListOptions: github.ListOptions{
				Page:    1,
				PerPage: ParamPerPageValueMax,
			},
		}
	}
	var iss Issues
	for {
		apiRespIssueSearch, httpRespGH, err := c.SearchIssues(ctx, qry, opts)
		if err != nil {
			return iss, errorsutil.Wrapf(err, "error on SearchIssues with params (%s)", string(jsonutil.MustMarshalSimple(opts, "", "")))
		}
		if httpRespGH.StatusCode >= 300 {
			return iss, fmt.Errorf("error on httpresponse status (%d) body (%s)", httpRespGH.StatusCode, string(ioutil.ReadAllOrError(httpRespGH.Body)))
		}
		if apiRespIssueSearch == nil {
			return iss, errors.New("nil response for github.IssuesSearchResult")
		}
		if len(apiRespIssueSearch.Issues) == 0 {
			break
		}
		iss = append(iss, apiRespIssueSearch.Issues...)
		if apiRespIssueSearch.Total != nil {
			if len(iss) >= pointer.Dereference(apiRespIssueSearch.Total) {
				break
			}
		}
		opts.ListOptions.Page++
	}
	return iss, nil
}
