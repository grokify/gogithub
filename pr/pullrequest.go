// Package pr provides GitHub pull request operations.
package pr

import (
	"context"
	"fmt"

	"github.com/google/go-github/v81/github"
)

// PRError indicates a failure to create a pull request.
type PRError struct {
	Title string
	Err   error
}

func (e *PRError) Error() string {
	return "failed to create PR '" + e.Title + "': " + e.Err.Error()
}

func (e *PRError) Unwrap() error {
	return e.Err
}

// CreatePR creates a pull request.
func CreatePR(ctx context.Context, gh *github.Client, upstreamOwner, upstreamRepo, forkOwner, branch, baseBranch, title, body string) (*github.PullRequest, error) {
	head := fmt.Sprintf("%s:%s", forkOwner, branch)

	pr, _, err := gh.PullRequests.Create(ctx, upstreamOwner, upstreamRepo, &github.NewPullRequest{
		Title: github.Ptr(title),
		Body:  github.Ptr(body),
		Head:  github.Ptr(head),
		Base:  github.Ptr(baseBranch),
	})
	if err != nil {
		return nil, &PRError{Title: title, Err: err}
	}

	return pr, nil
}

// GetPR retrieves a pull request by number.
func GetPR(ctx context.Context, gh *github.Client, owner, repo string, number int) (*github.PullRequest, error) {
	pr, _, err := gh.PullRequests.Get(ctx, owner, repo, number)
	return pr, err
}

// ListPRs lists pull requests for a repository.
func ListPRs(ctx context.Context, gh *github.Client, owner, repo string, opts *github.PullRequestListOptions) ([]*github.PullRequest, error) {
	prs, _, err := gh.PullRequests.List(ctx, owner, repo, opts)
	return prs, err
}

// MergePR merges a pull request.
func MergePR(ctx context.Context, gh *github.Client, owner, repo string, number int, commitMessage string, opts *github.PullRequestOptions) (*github.PullRequestMergeResult, error) {
	result, _, err := gh.PullRequests.Merge(ctx, owner, repo, number, commitMessage, opts)
	return result, err
}

// ClosePR closes a pull request without merging.
func ClosePR(ctx context.Context, gh *github.Client, owner, repo string, number int) (*github.PullRequest, error) {
	pr, _, err := gh.PullRequests.Edit(ctx, owner, repo, number, &github.PullRequest{
		State: github.Ptr("closed"),
	})
	return pr, err
}

// AddPRReviewers adds reviewers to a pull request.
func AddPRReviewers(ctx context.Context, gh *github.Client, owner, repo string, number int, reviewers, teamReviewers []string) (*github.PullRequest, error) {
	pr, _, err := gh.PullRequests.RequestReviewers(ctx, owner, repo, number, github.ReviewersRequest{
		Reviewers:     reviewers,
		TeamReviewers: teamReviewers,
	})
	return pr, err
}

// ListPRFiles lists files changed in a pull request.
func ListPRFiles(ctx context.Context, gh *github.Client, owner, repo string, number int) ([]*github.CommitFile, error) {
	files, _, err := gh.PullRequests.ListFiles(ctx, owner, repo, number, nil)
	return files, err
}

// ListPRComments lists comments on a pull request.
func ListPRComments(ctx context.Context, gh *github.Client, owner, repo string, number int) ([]*github.PullRequestComment, error) {
	comments, _, err := gh.PullRequests.ListComments(ctx, owner, repo, number, nil)
	return comments, err
}
