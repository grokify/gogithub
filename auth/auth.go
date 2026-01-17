// Package auth provides GitHub authentication utilities.
package auth

import (
	"context"
	"net/http"

	"github.com/google/go-github/v81/github"
	"golang.org/x/oauth2"
)

// AuthError indicates an authentication failure.
type AuthError struct {
	Message string
}

func (e *AuthError) Error() string {
	return "authentication failed: " + e.Message
}

// NewTokenClient creates an HTTP client authenticated with the given token.
func NewTokenClient(ctx context.Context, token string) *http.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	return oauth2.NewClient(ctx, ts)
}

// NewGitHubClient creates a GitHub client authenticated with the given token.
func NewGitHubClient(ctx context.Context, token string) *github.Client {
	return github.NewClient(NewTokenClient(ctx, token))
}

// GetAuthenticatedUser returns the authenticated user's login.
func GetAuthenticatedUser(ctx context.Context, gh *github.Client) (string, error) {
	user, _, err := gh.Users.Get(ctx, "")
	if err != nil {
		return "", &AuthError{Message: err.Error()}
	}
	return user.GetLogin(), nil
}

// GetUser returns information about a specific user.
func GetUser(ctx context.Context, gh *github.Client, username string) (*github.User, error) {
	user, _, err := gh.Users.Get(ctx, username)
	return user, err
}
