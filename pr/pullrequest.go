// Package pr provides GitHub pull request operations.
package pr

import (
	"context"

	"github.com/grokify/gogithub"
	"github.com/grokify/gogithub/clientv1"
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
func CreatePR(ctx context.Context, client clientv1.Client, upstreamOwner, upstreamRepo, forkOwner, branch, baseBranch, title, body string) (*gogithub.PullRequest, error) {
	head := forkOwner + ":" + branch

	input := &clientv1.CreatePullRequestInput{
		Title: title,
		Head:  head,
		Base:  baseBranch,
		Body:  body,
	}

	pr, err := client.CreatePullRequest(ctx, upstreamOwner, upstreamRepo, input)
	if err != nil {
		return nil, &PRError{Title: title, Err: err}
	}

	return pr, nil
}

// GetPR retrieves a pull request by number.
func GetPR(ctx context.Context, client clientv1.Client, owner, repo string, number int) (*gogithub.PullRequest, error) {
	return client.GetPullRequest(ctx, owner, repo, number)
}

// ListPRs lists pull requests for a repository.
func ListPRs(ctx context.Context, client clientv1.Client, owner, repo string, opts *clientv1.ListPullRequestsOptions) ([]*gogithub.PullRequest, error) {
	return client.ListPullRequests(ctx, owner, repo, opts)
}

// MergePR merges a pull request.
func MergePR(ctx context.Context, client clientv1.Client, owner, repo string, number int, commitMessage string, opts *clientv1.MergePullRequestOptions) (*gogithub.MergeResult, error) {
	if opts == nil {
		opts = &clientv1.MergePullRequestOptions{}
	}
	if commitMessage != "" {
		opts.CommitMessage = commitMessage
	}
	return client.MergePullRequest(ctx, owner, repo, number, opts)
}

// ClosePR closes a pull request without merging.
func ClosePR(ctx context.Context, client clientv1.Client, owner, repo string, number int) (*gogithub.PullRequest, error) {
	state := "closed"
	input := &clientv1.UpdatePullRequestInput{
		State: &state,
	}
	return client.UpdatePullRequest(ctx, owner, repo, number, input)
}

// AddPRReviewers adds reviewers to a pull request.
func AddPRReviewers(ctx context.Context, client clientv1.Client, owner, repo string, number int, reviewers, teamReviewers []string) (*gogithub.PullRequest, error) {
	return client.RequestReviewers(ctx, owner, repo, number, reviewers, teamReviewers)
}

// ListPRFiles lists files changed in a pull request.
func ListPRFiles(ctx context.Context, client clientv1.Client, owner, repo string, number int) ([]*gogithub.CommitFile, error) {
	return client.ListPullRequestFiles(ctx, owner, repo, number)
}

// ListPRComments lists comments on a pull request.
func ListPRComments(ctx context.Context, client clientv1.Client, owner, repo string, number int) ([]*gogithub.PullRequestComment, error) {
	return client.ListPullRequestComments(ctx, owner, repo, number)
}

// ApprovePR adds an approval review to a pull request.
func ApprovePR(ctx context.Context, client clientv1.Client, owner, repo string, number int, body string) (*gogithub.PullRequestReview, error) {
	input := &clientv1.CreateReviewInput{
		Event: "APPROVE",
		Body:  body,
	}
	return client.CreatePullRequestReview(ctx, owner, repo, number, input)
}

// RequestChangesPR requests changes on a pull request.
func RequestChangesPR(ctx context.Context, client clientv1.Client, owner, repo string, number int, body string) (*gogithub.PullRequestReview, error) {
	input := &clientv1.CreateReviewInput{
		Event: "REQUEST_CHANGES",
		Body:  body,
	}
	return client.CreatePullRequestReview(ctx, owner, repo, number, input)
}

// CommentPR adds a comment review to a pull request.
func CommentPR(ctx context.Context, client clientv1.Client, owner, repo string, number int, body string) (*gogithub.PullRequestReview, error) {
	input := &clientv1.CreateReviewInput{
		Event: "COMMENT",
		Body:  body,
	}
	return client.CreatePullRequestReview(ctx, owner, repo, number, input)
}

// MergeableState represents the mergeable state of a PR.
type MergeableState struct {
	Mergeable bool
	State     string // clean, dirty, blocked, behind, unstable, unknown
	Message   string
}

// IsMergeable checks if a PR can be merged and returns detailed status.
func IsMergeable(ctx context.Context, client clientv1.Client, owner, repo string, number int) (*MergeableState, error) {
	pr, err := client.GetPullRequest(ctx, owner, repo, number)
	if err != nil {
		return nil, err
	}

	state := &MergeableState{
		State: "unknown",
	}
	if pr.Mergeable != nil {
		state.Mergeable = *pr.Mergeable
	}

	if pr.Draft {
		state.Mergeable = false
		state.Message = "PR is a draft"
		return state, nil
	}

	if pr.State != "open" {
		state.Mergeable = false
		state.Message = "PR is not open"
		return state, nil
	}

	// Note: MergeableState is not directly exposed in our stable types
	// For now, we provide basic status based on available fields
	if state.Mergeable {
		state.State = "clean"
		state.Message = "PR appears ready to merge"
	} else {
		state.State = "blocked"
		state.Message = "PR cannot be merged"
	}

	return state, nil
}

// ListPRReviews lists reviews on a pull request.
func ListPRReviews(ctx context.Context, client clientv1.Client, owner, repo string, number int) ([]*gogithub.PullRequestReview, error) {
	return client.ListPullRequestReviews(ctx, owner, repo, number)
}

// GetPRDiff fetches the diff for a pull request.
func GetPRDiff(ctx context.Context, client clientv1.Client, owner, repo string, number int) (string, error) {
	return client.GetPullRequestDiff(ctx, owner, repo, number)
}

// GetPRPatch fetches the patch for a pull request.
func GetPRPatch(ctx context.Context, client clientv1.Client, owner, repo string, number int) (string, error) {
	return client.GetPullRequestPatch(ctx, owner, repo, number)
}

// CreateLineComment adds a comment on a specific line in a PR diff.
func CreateLineComment(ctx context.Context, client clientv1.Client, owner, repo string, number int, commitID, path, body string, line int) (*gogithub.PullRequestComment, error) {
	input := &clientv1.CreatePRCommentInput{
		Body:     body,
		CommitID: commitID,
		Path:     path,
		Line:     line,
		Side:     "RIGHT",
	}
	return client.CreatePullRequestComment(ctx, owner, repo, number, input)
}

// CreateIssueComment adds a general comment to a pull request (as an issue comment).
func CreateIssueComment(ctx context.Context, client clientv1.Client, owner, repo string, number int, body string) (*gogithub.IssueComment, error) {
	return client.CreateIssueComment(ctx, owner, repo, number, body)
}

// ReviewEvent represents the type of review action.
type ReviewEvent string

const (
	ReviewEventApprove        ReviewEvent = "APPROVE"
	ReviewEventRequestChanges ReviewEvent = "REQUEST_CHANGES"
	ReviewEventComment        ReviewEvent = "COMMENT"
)

// CreateReview creates a pull request review with the specified event type.
func CreateReview(ctx context.Context, client clientv1.Client, owner, repo string, number int, event ReviewEvent, body string) (*gogithub.PullRequestReview, error) {
	input := &clientv1.CreateReviewInput{
		Event: string(event),
		Body:  body,
	}
	return client.CreatePullRequestReview(ctx, owner, repo, number, input)
}
