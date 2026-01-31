package repo

import (
	"context"
	"strings"

	"github.com/google/go-github/v82/github"
)

// BranchError indicates a failure to create or update a branch.
type BranchError struct {
	Branch string
	Err    error
}

func (e *BranchError) Error() string {
	return "failed to create branch " + e.Branch + ": " + e.Err.Error()
}

func (e *BranchError) Unwrap() error {
	return e.Err
}

// GetBranchSHA returns the SHA of the given branch.
func GetBranchSHA(ctx context.Context, gh *github.Client, owner, repo, branch string) (string, error) {
	ref, _, err := gh.Git.GetRef(ctx, owner, repo, "refs/heads/"+branch)
	if err != nil {
		return "", err
	}
	return ref.GetObject().GetSHA(), nil
}

// CreateBranch creates a new branch from the given base SHA.
func CreateBranch(ctx context.Context, gh *github.Client, owner, repo, branch, baseSHA string) error {
	createRef := github.CreateRef{
		Ref: "refs/heads/" + branch,
		SHA: baseSHA,
	}

	_, _, err := gh.Git.CreateRef(ctx, owner, repo, createRef)
	if err != nil {
		// Check if branch already exists
		if strings.Contains(err.Error(), "already exists") {
			return nil
		}
		return &BranchError{Branch: branch, Err: err}
	}

	return nil
}

// DeleteBranch deletes a branch.
func DeleteBranch(ctx context.Context, gh *github.Client, owner, repo, branch string) error {
	_, err := gh.Git.DeleteRef(ctx, owner, repo, "refs/heads/"+branch)
	return err
}

// BranchExists checks if a branch exists.
func BranchExists(ctx context.Context, gh *github.Client, owner, repo, branch string) (bool, error) {
	_, resp, err := gh.Git.GetRef(ctx, owner, repo, "refs/heads/"+branch)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
