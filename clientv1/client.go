package clientv1

import (
	"context"
	"time"

	"github.com/grokify/gogithub"
)

// Client provides a version-isolated interface to GitHub operations.
// All methods return stable types defined in the gogithub package, not go-github types.
type Client interface {
	// Authentication

	// GetAuthenticatedUser returns the currently authenticated user.
	GetAuthenticatedUser(ctx context.Context) (*gogithub.User, error)

	// GetUser returns information about a specific user.
	GetUser(ctx context.Context, username string) (*gogithub.User, error)

	// Repositories

	// GetRepository retrieves a repository by owner and name.
	GetRepository(ctx context.Context, owner, repo string) (*gogithub.Repository, error)

	// ListUserRepos lists all repositories for a user.
	ListUserRepos(ctx context.Context, user string) ([]*gogithub.Repository, error)

	// ListOrgRepos lists all repositories for an organization.
	ListOrgRepos(ctx context.Context, org string) ([]*gogithub.Repository, error)

	// Content

	// GetFileContent fetches a file's content from a repository.
	GetFileContent(ctx context.Context, owner, repo, path string, opts *gogithub.ContentOptions) ([]byte, error)

	// GetFileContentString fetches a file's content as a string.
	GetFileContentString(ctx context.Context, owner, repo, path string, opts *gogithub.ContentOptions) (string, error)

	// ListDirectory lists files in a directory.
	ListDirectory(ctx context.Context, owner, repo, path string, opts *gogithub.ContentOptions) ([]*gogithub.FileContent, error)

	// FileExists checks if a file exists in a repository.
	FileExists(ctx context.Context, owner, repo, path string, opts *gogithub.ContentOptions) (bool, error)

	// Git References

	// GetRef retrieves a git reference by its full name (e.g., "refs/heads/main").
	GetRef(ctx context.Context, owner, repo, ref string) (*gogithub.Reference, error)

	// GetBranchSHA returns the commit SHA for a branch.
	GetBranchSHA(ctx context.Context, owner, repo, branch string) (string, error)

	// GetTagSHA returns the commit SHA for a tag.
	GetTagSHA(ctx context.Context, owner, repo, tag string) (string, error)

	// ListBranches lists all branches in a repository.
	ListBranches(ctx context.Context, owner, repo string) ([]*gogithub.Branch, error)

	// Commits

	// GetCommit retrieves a commit by SHA.
	GetCommit(ctx context.Context, owner, repo, sha string) (*gogithub.Commit, error)

	// ListCommits lists commits in a repository.
	ListCommits(ctx context.Context, owner, repo string, opts *ListCommitsOptions) ([]*gogithub.Commit, error)

	// Pull Requests

	// GetPullRequest retrieves a pull request by number.
	GetPullRequest(ctx context.Context, owner, repo string, number int) (*gogithub.PullRequest, error)

	// ListPullRequests lists pull requests in a repository.
	ListPullRequests(ctx context.Context, owner, repo string, opts *ListPullRequestsOptions) ([]*gogithub.PullRequest, error)

	// CreatePullRequest creates a new pull request.
	CreatePullRequest(ctx context.Context, owner, repo string, input *CreatePullRequestInput) (*gogithub.PullRequest, error)

	// Checks

	// ListCheckRuns lists check runs for a git reference.
	ListCheckRuns(ctx context.Context, owner, repo, ref string) ([]*gogithub.CheckRun, error)

	// Releases

	// GetLatestRelease retrieves the latest release.
	GetLatestRelease(ctx context.Context, owner, repo string) (*gogithub.Release, error)

	// GetReleaseByTag retrieves a release by its tag name.
	GetReleaseByTag(ctx context.Context, owner, repo, tag string) (*gogithub.Release, error)

	// ListReleases lists all releases in a repository.
	ListReleases(ctx context.Context, owner, repo string) ([]*gogithub.Release, error)

	// Raw returns the underlying go-github client for advanced use cases.
	// WARNING: Using this couples your code to a specific go-github version.
	// The returned value is *github.Client from the go-github package.
	Raw() any
}

// ListCommitsOptions specifies options for listing commits.
type ListCommitsOptions struct {
	// SHA is the branch name or commit SHA to start from.
	SHA string
	// Path filters to commits containing this file path.
	Path string
	// Since filters to commits after this time.
	Since *time.Time
	// Until filters to commits before this time.
	Until *time.Time
	// Author filters to commits by this author (GitHub login or email).
	Author string
}

// ListPullRequestsOptions specifies options for listing pull requests.
type ListPullRequestsOptions struct {
	// State filters by state: "open", "closed", or "all". Default: "open".
	State string
	// Head filters by head branch (format: "user:branch" or "org:branch").
	Head string
	// Base filters by base branch name.
	Base string
	// Sort specifies the sort order: "created", "updated", "popularity", "long-running".
	Sort string
	// Direction specifies sort direction: "asc" or "desc".
	Direction string
}

// CreatePullRequestInput specifies the input for creating a pull request.
type CreatePullRequestInput struct {
	// Title is the PR title (required).
	Title string
	// Head is the branch containing changes (required).
	Head string
	// Base is the branch to merge into (required).
	Base string
	// Body is the PR description.
	Body string
	// Draft creates the PR as a draft.
	Draft bool
	// MaintainerCanModify allows maintainers to push to the head branch.
	MaintainerCanModify bool
}
