package gogithub

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-github/v56/github"
	"github.com/grokify/mogo/encoding/jsonutil"
	"github.com/grokify/mogo/errors/errorsutil"
	"github.com/grokify/mogo/io/ioutil"
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
		ParamUser:  username,
		ParamState: ParamStateValueOpen,
		ParamIs:    ParamIsValuePR,
	}
	return c.SearchIssues(ctx, qry, opts)
}

// SearchIssues is a wrapper for `SearchService.Issues()`.
func (c *Client) SearchIssues(ctx context.Context, qry Query, opts *github.SearchOptions) (*github.IssuesSearchResult, *github.Response, error) {
	// client := github.NewClient(nil)
	// func (s *SearchService) Issues(ctx context.Context, query string, opts *SearchOptions) (*IssuesSearchResult, *Response, error)
	return c.Client.Search.Issues(ctx, qry.Encode(), opts)
}

func (c *Client) SearchIssuesAll(ctx context.Context, qry Query, opts *github.SearchOptions) (Issues, error) {
	if opts == nil {
		opts = &github.SearchOptions{
			ListOptions: github.ListOptions{
				Page:    1,
				PerPage: ParamPerPageValueMax}}
	}
	var iss Issues
	for {
		if apiRespIssueSearch, httpRespGH, err := c.SearchIssues(ctx, qry, opts); err != nil {
			return iss, errorsutil.Wrapf(err, "error on `c.SearchIssues()` with params (%s)", string(jsonutil.MustMarshalSimple(opts, "", "")))
		} else if httpRespGH.StatusCode >= 300 {
			return iss, fmt.Errorf("error on httpresponse status (%d) body (%s)", httpRespGH.StatusCode, string(ioutil.ReadAllOrError(httpRespGH.Body)))
		} else if apiRespIssueSearch == nil {
			return iss, errors.New("nil response for `github.IssuesSearchResult`")
		} else if len(apiRespIssueSearch.Issues) == 0 {
			break
		} else {
			iss = append(iss, apiRespIssueSearch.Issues...)
			fmt.Printf("ISS_COUNT (%d)\n", len(apiRespIssueSearch.Issues))
			if apiRespIssueSearch.Total != nil {
				if len(iss) >= pointer.Dereference(apiRespIssueSearch.Total) {
					break
				}
			}
			opts.ListOptions.Page++
		}
	}
	return iss, nil
}

type Query map[string]string

// Encode implements a version of GitHub API encoding.
func (q Query) Encode() string {
	var parts []string
	for k, v := range q {
		parts = append(parts, k+":"+v)
	}
	return strings.Join(parts, " ")
}
