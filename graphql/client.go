// Package graphql provides GitHub GraphQL API utilities.
package graphql

import (
	"context"

	"github.com/grokify/gogithub/auth"
	"github.com/shurcooL/githubv4"
)

// NewClient creates a GitHub GraphQL client authenticated with the given token.
func NewClient(ctx context.Context, token string) *githubv4.Client {
	return githubv4.NewClient(auth.NewTokenClient(ctx, token))
}

// NewEnterpriseClient creates a GitHub GraphQL client for GitHub Enterprise.
func NewEnterpriseClient(ctx context.Context, token, baseURL string) *githubv4.Client {
	return githubv4.NewEnterpriseClient(baseURL, auth.NewTokenClient(ctx, token))
}
