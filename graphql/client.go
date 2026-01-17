// Package graphql provides GitHub GraphQL API utilities.
package graphql

import (
	"context"
	"net/http"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

// NewTokenClient creates an HTTP client authenticated with the given token.
func NewTokenClient(ctx context.Context, token string) *http.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	return oauth2.NewClient(ctx, ts)
}

// NewClient creates a GitHub GraphQL client authenticated with the given token.
func NewClient(ctx context.Context, token string) *githubv4.Client {
	return githubv4.NewClient(NewTokenClient(ctx, token))
}

// NewEnterpriseClient creates a GitHub GraphQL client for GitHub Enterprise.
func NewEnterpriseClient(ctx context.Context, token, baseURL string) *githubv4.Client {
	return githubv4.NewEnterpriseClient(baseURL, NewTokenClient(ctx, token))
}
