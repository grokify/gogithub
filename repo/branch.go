package repo

import (
	"context"
	"strings"

	"github.com/grokify/gogithub/clientv1"
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
func GetBranchSHA(ctx context.Context, client clientv1.Client, owner, repo, branch string) (string, error) {
	return client.GetBranchSHA(ctx, owner, repo, branch)
}

// CreateBranch creates a new branch from the given base SHA.
func CreateBranch(ctx context.Context, client clientv1.Client, owner, repo, branch, baseSHA string) error {
	_, err := client.CreateRef(ctx, owner, repo, RefHeadsPrefix+branch, baseSHA)
	if err != nil {
		// Check if branch already exists
		if strings.Contains(err.Error(), ErrAlreadyExists) {
			return nil
		}
		return &BranchError{Branch: branch, Err: err}
	}

	return nil
}

// DeleteBranch deletes a branch.
func DeleteBranch(ctx context.Context, client clientv1.Client, owner, repo, branch string) error {
	return client.DeleteRef(ctx, owner, repo, RefHeadsPrefix+branch)
}

// BranchExists checks if a branch exists.
func BranchExists(ctx context.Context, client clientv1.Client, owner, repo, branch string) (bool, error) {
	_, err := client.GetRef(ctx, owner, repo, RefHeadsPrefix+branch)
	if err != nil {
		// Check if it's a 404 error
		if isNotFoundError(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// isNotFoundError checks if an error is a 404 not found error.
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "404") || strings.Contains(errStr, "not found")
}
