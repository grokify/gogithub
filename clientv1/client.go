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

	// GetDefaultBranch returns the default branch name for a repository.
	GetDefaultBranch(ctx context.Context, owner, repo string) (string, error)

	// Forks

	// CreateFork creates a fork of a repository.
	CreateFork(ctx context.Context, owner, repo string, opts *CreateForkOptions) (*gogithub.Repository, error)

	// Content

	// GetFileContent fetches a file's content from a repository.
	GetFileContent(ctx context.Context, owner, repo, path string, opts *gogithub.ContentOptions) ([]byte, error)

	// GetFileContentString fetches a file's content as a string.
	GetFileContentString(ctx context.Context, owner, repo, path string, opts *gogithub.ContentOptions) (string, error)

	// GetFileContentWithSHA fetches a file's content and returns its SHA.
	// Returns the content bytes, the file's SHA, and any error.
	GetFileContentWithSHA(ctx context.Context, owner, repo, path string, opts *gogithub.ContentOptions) ([]byte, string, error)

	// ListDirectory lists files in a directory.
	ListDirectory(ctx context.Context, owner, repo, path string, opts *gogithub.ContentOptions) ([]*gogithub.FileContent, error)

	// FileExists checks if a file exists in a repository.
	FileExists(ctx context.Context, owner, repo, path string, opts *gogithub.ContentOptions) (bool, error)

	// CreateFile creates a new file in a repository.
	CreateFile(ctx context.Context, owner, repo, path string, opts *CreateFileOptions) (*gogithub.CreateFileResult, error)

	// UpdateFile updates an existing file in a repository.
	// Requires the current SHA of the file for optimistic locking.
	UpdateFile(ctx context.Context, owner, repo, path string, opts *UpdateFileOptions) (*gogithub.CreateFileResult, error)

	// DeleteFile deletes a file from a repository.
	// Requires the current SHA of the file.
	DeleteFile(ctx context.Context, owner, repo, path, sha, message string, opts *DeleteFileOptions) (*gogithub.DeleteFileResult, error)

	// Git References

	// GetRef retrieves a git reference by its full name (e.g., "refs/heads/main").
	GetRef(ctx context.Context, owner, repo, ref string) (*gogithub.Reference, error)

	// CreateRef creates a git reference.
	CreateRef(ctx context.Context, owner, repo, ref, sha string) (*gogithub.Reference, error)

	// UpdateRef updates a git reference to point to a new SHA.
	UpdateRef(ctx context.Context, owner, repo, ref, sha string, force bool) (*gogithub.Reference, error)

	// DeleteRef deletes a git reference.
	DeleteRef(ctx context.Context, owner, repo, ref string) error

	// GetBranchSHA returns the commit SHA for a branch.
	GetBranchSHA(ctx context.Context, owner, repo, branch string) (string, error)

	// GetTagSHA returns the commit SHA for a tag.
	GetTagSHA(ctx context.Context, owner, repo, tag string) (string, error)

	// ListBranches lists all branches in a repository.
	ListBranches(ctx context.Context, owner, repo string) ([]*gogithub.Branch, error)

	// Tags

	// ListTags lists all tags in a repository.
	ListTags(ctx context.Context, owner, repo string) ([]*gogithub.Tag, error)

	// CreateTag creates an annotated tag.
	CreateTag(ctx context.Context, owner, repo, tag, sha, message string) error

	// Commits

	// GetCommit retrieves a commit by SHA.
	GetCommit(ctx context.Context, owner, repo, sha string) (*gogithub.Commit, error)

	// ListCommits lists commits in a repository.
	ListCommits(ctx context.Context, owner, repo string, opts *ListCommitsOptions) ([]*gogithub.Commit, error)

	// CreateCommit creates a commit with the given tree and parent.
	CreateCommit(ctx context.Context, owner, repo string, opts *CreateCommitOptions) (*gogithub.Commit, error)

	// Git Trees

	// GetTree retrieves a git tree by SHA.
	GetTree(ctx context.Context, owner, repo, sha string, recursive bool) ([]*gogithub.TreeNode, error)

	// CreateTree creates a git tree from file entries.
	CreateTree(ctx context.Context, owner, repo, baseTree string, entries []TreeEntry) (string, error)

	// Git Blobs

	// CreateBlob creates a git blob with the given content.
	CreateBlob(ctx context.Context, owner, repo string, content []byte, encoding string) (string, error)

	// Pull Requests

	// GetPullRequest retrieves a pull request by number.
	GetPullRequest(ctx context.Context, owner, repo string, number int) (*gogithub.PullRequest, error)

	// ListPullRequests lists pull requests in a repository.
	ListPullRequests(ctx context.Context, owner, repo string, opts *ListPullRequestsOptions) ([]*gogithub.PullRequest, error)

	// CreatePullRequest creates a new pull request.
	CreatePullRequest(ctx context.Context, owner, repo string, input *CreatePullRequestInput) (*gogithub.PullRequest, error)

	// UpdatePullRequest updates a pull request.
	UpdatePullRequest(ctx context.Context, owner, repo string, number int, input *UpdatePullRequestInput) (*gogithub.PullRequest, error)

	// MergePullRequest merges a pull request.
	MergePullRequest(ctx context.Context, owner, repo string, number int, opts *MergePullRequestOptions) (*gogithub.MergeResult, error)

	// ListPullRequestFiles lists files changed in a pull request.
	ListPullRequestFiles(ctx context.Context, owner, repo string, number int) ([]*gogithub.CommitFile, error)

	// GetPullRequestDiff gets the diff for a pull request.
	GetPullRequestDiff(ctx context.Context, owner, repo string, number int) (string, error)

	// GetPullRequestPatch gets the patch for a pull request.
	GetPullRequestPatch(ctx context.Context, owner, repo string, number int) (string, error)

	// Pull Request Reviews

	// CreatePullRequestReview creates a review on a pull request.
	CreatePullRequestReview(ctx context.Context, owner, repo string, number int, input *CreateReviewInput) (*gogithub.PullRequestReview, error)

	// ListPullRequestReviews lists reviews on a pull request.
	ListPullRequestReviews(ctx context.Context, owner, repo string, number int) ([]*gogithub.PullRequestReview, error)

	// RequestReviewers requests reviewers for a pull request.
	RequestReviewers(ctx context.Context, owner, repo string, number int, reviewers, teamReviewers []string) (*gogithub.PullRequest, error)

	// Pull Request Comments

	// CreatePullRequestComment creates a comment on a pull request diff.
	CreatePullRequestComment(ctx context.Context, owner, repo string, number int, input *CreatePRCommentInput) (*gogithub.PullRequestComment, error)

	// ListPullRequestComments lists comments on a pull request.
	ListPullRequestComments(ctx context.Context, owner, repo string, number int) ([]*gogithub.PullRequestComment, error)

	// Issues

	// CreateIssueComment creates a comment on an issue or pull request.
	CreateIssueComment(ctx context.Context, owner, repo string, number int, body string) (*gogithub.IssueComment, error)

	// Checks

	// GetCheckRun retrieves a check run by ID.
	GetCheckRun(ctx context.Context, owner, repo string, checkRunID int64) (*gogithub.CheckRun, error)

	// ListCheckRuns lists check runs for a git reference.
	ListCheckRuns(ctx context.Context, owner, repo, ref string) ([]*gogithub.CheckRun, error)

	// ListCheckSuites lists check suites for a git reference.
	ListCheckSuites(ctx context.Context, owner, repo, ref string) ([]*gogithub.CheckSuite, error)

	// Releases

	// GetRelease retrieves a release by ID.
	GetRelease(ctx context.Context, owner, repo string, id int64) (*gogithub.Release, error)

	// GetLatestRelease retrieves the latest release.
	GetLatestRelease(ctx context.Context, owner, repo string) (*gogithub.Release, error)

	// GetReleaseByTag retrieves a release by its tag name.
	GetReleaseByTag(ctx context.Context, owner, repo, tag string) (*gogithub.Release, error)

	// ListReleases lists all releases in a repository.
	ListReleases(ctx context.Context, owner, repo string) ([]*gogithub.Release, error)

	// CreateRelease creates a new release.
	CreateRelease(ctx context.Context, owner, repo string, input *CreateReleaseInput) (*gogithub.Release, error)

	// UpdateRelease updates a release.
	UpdateRelease(ctx context.Context, owner, repo string, id int64, input *UpdateReleaseInput) (*gogithub.Release, error)

	// DeleteRelease deletes a release.
	DeleteRelease(ctx context.Context, owner, repo string, id int64) error

	// ListReleaseAssets lists assets for a release.
	ListReleaseAssets(ctx context.Context, owner, repo string, releaseID int64) ([]*gogithub.ReleaseAsset, error)

	// Search

	// SearchIssues searches for issues and pull requests.
	SearchIssues(ctx context.Context, query string, opts *SearchOptions) (*gogithub.IssueSearchResult, error)

	// Contributors

	// GetContributorStats gets contribution statistics for a repository.
	GetContributorStats(ctx context.Context, owner, repo string) ([]*gogithub.ContributorStats, error)

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

// UpdatePullRequestInput specifies the input for updating a pull request.
type UpdatePullRequestInput struct {
	Title               *string
	Body                *string
	State               *string // "open" or "closed"
	Base                *string
	MaintainerCanModify *bool
}

// MergePullRequestOptions specifies options for merging a pull request.
type MergePullRequestOptions struct {
	CommitTitle   string
	CommitMessage string
	MergeMethod   string // "merge", "squash", or "rebase"
	SHA           string // Expected head SHA for optimistic locking
}

// CreateReviewInput specifies input for creating a PR review.
type CreateReviewInput struct {
	Event string // "APPROVE", "REQUEST_CHANGES", or "COMMENT"
	Body  string
}

// CreatePRCommentInput specifies input for creating a PR diff comment.
type CreatePRCommentInput struct {
	Body     string
	CommitID string
	Path     string
	Line     int
	Side     string // "LEFT" or "RIGHT"
}

// CreateForkOptions specifies options for creating a fork.
type CreateForkOptions struct {
	Organization  string // Fork to this org instead of user's account
	Name          string // Custom name for the fork
	DefaultBranch bool   // Only fork the default branch
}

// CreateCommitOptions specifies options for creating a commit.
type CreateCommitOptions struct {
	Message string
	Tree    string   // Tree SHA
	Parents []string // Parent commit SHAs
	Author  *CommitAuthor
}

// CommitAuthor specifies author information for a commit.
type CommitAuthor struct {
	Name  string
	Email string
	Date  *time.Time
}

// TreeEntry represents a file entry for creating a git tree.
type TreeEntry struct {
	Path    string
	Mode    string // "100644" (file), "100755" (executable), "040000" (directory), "160000" (submodule), "120000" (symlink)
	Type    string // "blob", "tree", or "commit"
	SHA     string // SHA of existing blob, or empty to use Content
	Content string // File content (creates new blob)
}

// CreateReleaseInput specifies input for creating a release.
type CreateReleaseInput struct {
	TagName              string
	TargetCommitish      string
	Name                 string
	Body                 string
	Draft                bool
	Prerelease           bool
	GenerateReleaseNotes bool
}

// UpdateReleaseInput specifies input for updating a release.
type UpdateReleaseInput struct {
	TagName         *string
	TargetCommitish *string
	Name            *string
	Body            *string
	Draft           *bool
	Prerelease      *bool
}

// SearchOptions specifies options for search queries.
type SearchOptions struct {
	Sort    string // "comments", "reactions", "created", "updated", etc.
	Order   string // "asc" or "desc"
	PerPage int
	Page    int
}

// CreateFileOptions specifies options for creating a file.
type CreateFileOptions struct {
	Content []byte
	Message string
	Branch  string // Optional: defaults to repository's default branch
	Author  *CommitAuthor
}

// UpdateFileOptions specifies options for updating a file.
type UpdateFileOptions struct {
	Content []byte
	SHA     string // Current file SHA (required for optimistic locking)
	Message string
	Branch  string
	Author  *CommitAuthor
}

// DeleteFileOptions specifies options for deleting a file.
type DeleteFileOptions struct {
	Branch string
	Author *CommitAuthor
}
