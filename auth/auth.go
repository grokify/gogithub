// Package auth provides GitHub authentication utilities.
//
// This package provides low-level authentication primitives for GitHub.
// For version-isolated client creation, use clientv1.NewClient directly.
package auth

import (
	"context"
	"net/http"

	"github.com/google/go-github/v89/github"
	"golang.org/x/oauth2"
)

// Known bot users.
const (
	// UsernameDependabot is the username for GitHub's Dependabot.
	UsernameDependabot = "dependabot[bot]"
	// UserIDDependabot is the user ID for GitHub's Dependabot.
	UserIDDependabot = 49699333
)

// AuthError indicates an authentication failure.
type AuthError struct {
	Message string
	Err     error // Wrapped error for Go 1.13+ error chain compatibility
}

func (e *AuthError) Error() string {
	return "authentication failed: " + e.Message
}

// Unwrap returns the wrapped error for Go 1.13+ error chain compatibility.
func (e *AuthError) Unwrap() error {
	return e.Err
}

// NewTokenClient creates an HTTP client authenticated with the given token.
func NewTokenClient(ctx context.Context, token string) *http.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	return oauth2.NewClient(ctx, ts)
}

// NewGitHubClient creates a GitHub client authenticated with the given token.
// Deprecated: Use clientv1.NewClient for version-isolated clients.
func NewGitHubClient(ctx context.Context, token string) (*github.Client, error) {
	return github.NewClient(github.WithHTTPClient(NewTokenClient(ctx, token)))
}

// GetAuthenticatedUser returns the authenticated user's login.
// Deprecated: Use clientv1.Client.GetAuthenticatedUser for version-isolated access.
func GetAuthenticatedUser(ctx context.Context, gh *github.Client) (string, error) {
	user, _, err := gh.Users.Get(ctx, "")
	if err != nil {
		return "", &AuthError{Message: err.Error(), Err: err}
	}
	return user.GetLogin(), nil
}

// GetUser returns information about a specific user.
// Deprecated: Use clientv1.Client.GetUser for version-isolated access.
func GetUser(ctx context.Context, gh *github.Client, username string) (*github.User, error) {
	user, _, err := gh.Users.Get(ctx, username)
	return user, err
}
