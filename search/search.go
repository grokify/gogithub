// Package search provides GitHub search API functionality.
package search

import (
	"context"
	"errors"

	"github.com/grokify/gogithub"
	"github.com/grokify/gogithub/clientv1"
	"github.com/grokify/mogo/encoding/jsonutil"
	"github.com/grokify/mogo/errors/errorsutil"
)

// Client wraps the GitHub client for search operations.
type Client struct {
	client clientv1.Client
}

// NewClient creates a new search client.
func NewClient(client clientv1.Client) *Client {
	return &Client{client: client}
}

// SearchOpenPullRequests searches for open pull requests by username.
func (c *Client) SearchOpenPullRequests(ctx context.Context, username string, opts *clientv1.SearchOptions) (*gogithub.IssueSearchResult, error) {
	qry := NewQuery().User(username).StateOpen().IsPR().Build()
	return c.SearchIssues(ctx, qry, opts)
}

// SearchIssues is a wrapper for SearchService.Issues().
func (c *Client) SearchIssues(ctx context.Context, qry Query, opts *clientv1.SearchOptions) (*gogithub.IssueSearchResult, error) {
	return c.client.SearchIssues(ctx, qry.Encode(), opts)
}

// SearchIssuesAll retrieves all issues matching the query with pagination.
func (c *Client) SearchIssuesAll(ctx context.Context, qry Query, opts *clientv1.SearchOptions) (Issues, error) {
	if opts == nil {
		opts = &clientv1.SearchOptions{
			PerPage: ParamPerPageValueMax,
			Page:    1,
		}
	}
	var iss Issues
	page := opts.Page
	if page == 0 {
		page = 1
	}
	for {
		pageOpts := &clientv1.SearchOptions{
			Sort:    opts.Sort,
			Order:   opts.Order,
			PerPage: opts.PerPage,
			Page:    page,
		}
		result, err := c.SearchIssues(ctx, qry, pageOpts)
		if err != nil {
			return iss, errorsutil.Wrapf(err, "error on SearchIssues with params (%s)", string(jsonutil.MustMarshalSimple(pageOpts, "", "")))
		}
		if result == nil {
			return iss, errors.New("nil response for IssueSearchResult")
		}
		if len(result.Items) == 0 {
			break
		}
		iss = append(iss, result.Items...)
		if len(iss) >= result.Total {
			break
		}
		page++
	}
	return iss, nil
}

// Raw returns the underlying clientv1.Client for advanced use cases.
func (c *Client) Raw() clientv1.Client {
	return c.client
}
