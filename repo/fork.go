// Package repo provides GitHub repository operations.
package repo

import (
	"context"
	"strings"
	"time"

	"github.com/google/go-github/v82/github"
)

// ForkError indicates a failure to fork a repository.
type ForkError struct {
	Owner string
	Repo  string
	Err   error
}

func (e *ForkError) Error() string {
	return "failed to fork " + e.Owner + "/" + e.Repo + ": " + e.Err.Error()
}

func (e *ForkError) Unwrap() error {
	return e.Err
}

// EnsureFork ensures a fork exists for the given repository.
// Returns the fork owner and repo name.
func EnsureFork(ctx context.Context, gh *github.Client, upstreamOwner, upstreamRepo, forkOwner string) (string, string, error) {
	// Check if fork already exists
	fork, resp, err := gh.Repositories.Get(ctx, forkOwner, upstreamRepo)
	if err == nil && fork != nil {
		return fork.GetOwner().GetLogin(), fork.GetName(), nil
	}

	// If not 404, return the error
	if resp != nil && resp.StatusCode != 404 {
		return "", "", &ForkError{Owner: upstreamOwner, Repo: upstreamRepo, Err: err}
	}

	// Create fork
	fork, _, err = gh.Repositories.CreateFork(ctx, upstreamOwner, upstreamRepo, &github.RepositoryCreateForkOptions{})
	if err != nil {
		// Check if it's an "already exists" error (async fork creation)
		if strings.Contains(err.Error(), "already exists") {
			// Wait a bit and try to get it again
			time.Sleep(2 * time.Second)
			fork, _, err = gh.Repositories.Get(ctx, forkOwner, upstreamRepo)
			if err != nil {
				return "", "", &ForkError{Owner: upstreamOwner, Repo: upstreamRepo, Err: err}
			}
			return fork.GetOwner().GetLogin(), fork.GetName(), nil
		}
		return "", "", &ForkError{Owner: upstreamOwner, Repo: upstreamRepo, Err: err}
	}

	// Wait for fork to be ready
	time.Sleep(3 * time.Second)

	return fork.GetOwner().GetLogin(), fork.GetName(), nil
}

// GetDefaultBranch returns the default branch of a repository.
func GetDefaultBranch(ctx context.Context, gh *github.Client, owner, repo string) (string, error) {
	repository, _, err := gh.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return "", err
	}
	return repository.GetDefaultBranch(), nil
}
