package gogithub

import "time"

// User represents a GitHub user. This is a stable type that won't change
// when go-github updates its major version.
type User struct {
	ID        int64
	Login     string
	Name      string
	Email     string
	AvatarURL string
	HTMLURL   string
	Type      string // "User" or "Organization"
	Bio       string
	Company   string
	Location  string
	Blog      string
	Followers int
	Following int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Repository represents a GitHub repository.
type Repository struct {
	ID              int64
	Owner           *User
	Name            string
	FullName        string
	Description     string
	HTMLURL         string
	CloneURL        string
	SSHURL          string
	DefaultBranch   string
	Private         bool
	Fork            bool
	Archived        bool
	Disabled        bool
	Language        string
	ForksCount      int
	StargazersCount int
	WatchersCount   int
	OpenIssuesCount int
	Size            int
	CreatedAt       time.Time
	UpdatedAt       time.Time
	PushedAt        time.Time
}

// Reference represents a git reference (branch, tag).
type Reference struct {
	Ref    string // e.g., "refs/heads/main"
	SHA    string // Convenience field: same as Object.SHA
	URL    string
	Object *GitObject
}

// GitObject represents the object a reference points to.
type GitObject struct {
	Type string // "commit", "tag", etc.
	SHA  string
	URL  string
}

// Commit represents a git commit.
type Commit struct {
	SHA       string
	Message   string
	Author    *CommitAuthor
	Committer *CommitAuthor
	HTMLURL   string
	Tree      *GitObject // Root tree of this commit
	Parents   []CommitParent
}

// CommitAuthor represents the author/committer of a commit.
type CommitAuthor struct {
	Name  string
	Email string
	Date  time.Time
}

// CommitParent represents a parent commit reference.
type CommitParent struct {
	SHA string
	URL string
}

// PullRequest represents a GitHub pull request.
type PullRequest struct {
	ID        int64
	Number    int
	State     string // "open", "closed"
	Title     string
	Body      string
	HTMLURL   string
	User      *User
	Head      *PullRequestBranch
	Base      *PullRequestBranch
	Labels    []Label
	Assignees []*User
	Merged    bool
	Mergeable *bool
	Draft     bool
	Additions int
	Deletions int
	Commits   int
	CreatedAt time.Time
	UpdatedAt time.Time
	ClosedAt  *time.Time
	MergedAt  *time.Time
}

// PullRequestBranch represents the head or base branch of a PR.
type PullRequestBranch struct {
	Label string
	Ref   string
	SHA   string
	User  *User
	Repo  *Repository
}

// CheckRun represents a GitHub Actions check run.
type CheckRun struct {
	ID          int64
	HeadSHA     string
	Status      string // "queued", "in_progress", "completed"
	Conclusion  string // "success", "failure", "neutral", "cancelled", "skipped", "timed_out", "action_required"
	Name        string
	HTMLURL     string
	StartedAt   *time.Time
	CompletedAt *time.Time
}

// Release represents a GitHub release.
type Release struct {
	ID              int64
	TagName         string
	TargetCommitish string
	Name            string
	Body            string
	Draft           bool
	Prerelease      bool
	HTMLURL         string
	TarballURL      string
	ZipballURL      string
	CreatedAt       time.Time
	PublishedAt     *time.Time
	Author          *User
	Assets          []ReleaseAsset
}

// ReleaseAsset represents an asset attached to a release.
type ReleaseAsset struct {
	ID                 int64
	Name               string
	Label              string
	State              string
	ContentType        string
	Size               int
	DownloadCount      int
	BrowserDownloadURL string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// Branch represents a repository branch.
type Branch struct {
	Name      string
	Protected bool
	Commit    *Commit
}

// Issue represents a GitHub issue.
type Issue struct {
	ID            int64
	Number        int
	State         string
	Title         string
	Body          string
	HTMLURL       string
	RepositoryURL string // API URL of the repository
	User          *User
	Labels        []Label
	Assignees     []*User
	Comments      int
	IsPullRequest bool // true if this issue is actually a pull request
	CreatedAt     time.Time
	UpdatedAt     time.Time
	ClosedAt      *time.Time
}

// Label represents a GitHub label.
type Label struct {
	ID          int64
	Name        string
	Description string
	Color       string
}

// FileContent represents file content from a repository.
type FileContent struct {
	Path        string
	Name        string
	SHA         string
	Size        int
	Type        string // "file", "dir", "symlink", "submodule"
	Content     []byte // Decoded content (for files)
	DownloadURL string
}

// ContentOptions specifies options for fetching repository content.
type ContentOptions struct {
	Ref string // Branch, tag, or commit SHA. Empty uses default branch.
}

// Tag represents a git tag.
type Tag struct {
	Name   string
	Commit *Commit
	SHA    string // SHA of the tag object (for annotated) or commit (for lightweight)
}

// CheckSuite represents a GitHub Actions check suite.
type CheckSuite struct {
	ID         int64
	HeadBranch string
	HeadSHA    string
	Status     string // "queued", "in_progress", "completed"
	Conclusion string // "success", "failure", "neutral", "cancelled", "skipped", "timed_out", "action_required"
	URL        string
	App        *App
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// App represents a GitHub App.
type App struct {
	ID          int64
	Slug        string
	Name        string
	Description string
	HTMLURL     string
}

// PullRequestReview represents a review on a pull request.
type PullRequestReview struct {
	ID          int64
	User        *User
	Body        string
	State       string // "APPROVED", "CHANGES_REQUESTED", "COMMENTED", "DISMISSED", "PENDING"
	HTMLURL     string
	CommitID    string
	SubmittedAt *time.Time
}

// PullRequestComment represents a comment on a pull request diff.
type PullRequestComment struct {
	ID        int64
	User      *User
	Body      string
	Path      string
	Line      int
	Side      string // "LEFT" or "RIGHT"
	CommitID  string
	HTMLURL   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// IssueComment represents a comment on an issue or pull request.
type IssueComment struct {
	ID        int64
	User      *User
	Body      string
	HTMLURL   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CommitFile represents a file changed in a commit.
type CommitFile struct {
	SHA         string
	Filename    string
	Status      string // "added", "removed", "modified", "renamed", "copied", "changed", "unchanged"
	Additions   int
	Deletions   int
	Changes     int
	Patch       string
	BlobURL     string
	RawURL      string
	ContentsURL string
	Previous    string // Previous filename for renamed files
}

// MergeResult represents the result of merging a pull request.
type MergeResult struct {
	SHA     string
	Merged  bool
	Message string
}

// ContributorStats represents contribution statistics for a user.
type ContributorStats struct {
	Author *User
	Total  int
	Weeks  []WeeklyStats
}

// WeeklyStats represents contribution stats for a single week.
type WeeklyStats struct {
	Week      time.Time
	Additions int
	Deletions int
	Commits   int
}

// SearchResult represents search results from the GitHub API.
type SearchResult[T any] struct {
	Total             int
	IncompleteResults bool
	Items             []T
}

// IssueSearchResult is a search result containing issues.
type IssueSearchResult = SearchResult[*Issue]

// ReleaseAssetUpload represents an asset being uploaded to a release.
type ReleaseAssetUpload struct {
	Name        string
	Label       string
	ContentType string
	Content     []byte
}

// TreeNode represents a node in a git tree.
type TreeNode struct {
	Path string
	Mode string // "100644" (file), "100755" (executable), "040000" (dir), "160000" (submodule), "120000" (symlink)
	Type string // "blob", "tree", "commit"
	SHA  string
	Size int
	URL  string
}

// CreateFileResult represents the result of creating or updating a file.
type CreateFileResult struct {
	Content *FileContent
	Commit  *Commit
}

// DeleteFileResult represents the result of deleting a file.
type DeleteFileResult struct {
	Commit *Commit
}

// CodeSearchResult represents search results for code.
type CodeSearchResult struct {
	Total             int
	IncompleteResults bool
	Items             []*CodeResult
}

// CodeResult represents a single code search result.
type CodeResult struct {
	Name       string
	Path       string
	SHA        string
	HTMLURL    string
	Repository *Repository
}

// Event represents a GitHub activity event, such as those returned by a
// user's public timeline (e.g., "PushEvent", "PullRequestEvent", "IssuesEvent").
type Event struct {
	ID        string
	Type      string
	Public    bool
	Actor     *User
	Repo      *EventRepo
	CreatedAt time.Time
}

// EventRepo identifies the repository an Event occurred in. It carries only
// the fields the GitHub Events API populates, not the full Repository.
type EventRepo struct {
	ID   int64
	Name string // "owner/repo"
	URL  string
}

// BranchProtection represents a repository branch's protection settings.
type BranchProtection struct {
	RequiredStatusChecks       *RequiredStatusChecks
	RequiredPullRequestReviews *PullRequestReviewsEnforcement
	EnforceAdmins              bool
	RequireSignedCommits       bool
	AllowForcePushes           bool
	AllowDeletions             bool
	URL                        string
}

// RequiredStatusChecks represents the required status checks of a protected branch.
type RequiredStatusChecks struct {
	Strict   bool
	Contexts []string
}

// PullRequestReviewsEnforcement represents the pull request review
// requirements of a protected branch.
type PullRequestReviewsEnforcement struct {
	DismissStaleReviews          bool
	RequireCodeOwnerReviews      bool
	RequiredApprovingReviewCount int
}

// Workflow represents a GitHub Actions workflow.
type Workflow struct {
	ID        int64
	Name      string
	Path      string
	State     string
	URL       string
	HTMLURL   string
	BadgeURL  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// WorkflowRun represents a run of a GitHub Actions workflow.
type WorkflowRun struct {
	ID         int64
	Name       string
	WorkflowID int64
	RunNumber  int
	Event      string
	Status     string
	Conclusion string
	HeadBranch string
	HeadSHA    string
	URL        string
	HTMLURL    string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
