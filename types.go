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
	Merged    bool
	Mergeable *bool
	Draft     bool
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
	ID        int64
	Number    int
	State     string
	Title     string
	Body      string
	HTMLURL   string
	User      *User
	Labels    []Label
	Assignees []*User
	CreatedAt time.Time
	UpdatedAt time.Time
	ClosedAt  *time.Time
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
