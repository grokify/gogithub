package clientv1

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/go-github/v89/github"
	"github.com/grokify/gogithub"
	"golang.org/x/oauth2"
)

// client implements the Client interface.
type client struct {
	gh *github.Client
}

// NewClient creates a new GitHub client with token authentication.
// This is the primary way to create a version-isolated GitHub client.
func NewClient(ctx context.Context, token string) (Client, error) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	gh, err := github.NewClient(github.WithHTTPClient(tc))
	if err != nil {
		return nil, err
	}
	return &client{gh: gh}, nil
}

// NewClientWithHTTP creates a new GitHub client with a custom HTTP client.
func NewClientWithHTTP(httpClient *http.Client) (Client, error) {
	gh, err := github.NewClient(github.WithHTTPClient(httpClient))
	if err != nil {
		return nil, err
	}
	return &client{gh: gh}, nil
}

// ClientOptions configures the GitHub client.
type ClientOptions struct {
	// Token is the GitHub personal access token.
	Token string
	// BaseURL is the GitHub API base URL (for GitHub Enterprise).
	// Leave empty for github.com.
	BaseURL string
	// UploadURL is the GitHub upload URL (for GitHub Enterprise).
	// Leave empty for github.com.
	UploadURL string
}

// NewClientWithOptions creates a new GitHub client with the given options.
// This supports GitHub Enterprise by specifying custom BaseURL and UploadURL.
func NewClientWithOptions(ctx context.Context, opts ClientOptions) (Client, error) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: opts.Token},
	)
	tc := oauth2.NewClient(ctx, ts)

	ghOpts := []github.ClientOptionsFunc{github.WithHTTPClient(tc)}
	if opts.BaseURL != "" {
		ghOpts = append(ghOpts, github.WithEnterpriseURLs(opts.BaseURL, opts.UploadURL))
	}

	gh, err := github.NewClient(ghOpts...)
	if err != nil {
		return nil, err
	}
	return &client{gh: gh}, nil
}

// MustNewClient creates a new GitHub client, panicking on error.
func MustNewClient(ctx context.Context, token string) Client {
	c, err := NewClient(ctx, token)
	if err != nil {
		panic(err)
	}
	return c
}

// NewClientFromRaw wraps an existing go-github client.
// Use this when you already have a *github.Client from another source.
func NewClientFromRaw(gh *github.Client) Client {
	return &client{gh: gh}
}

// Raw returns the underlying go-github client.
func (c *client) Raw() any {
	return c.gh
}

// GetAuthenticatedUser returns the currently authenticated user.
func (c *client) GetAuthenticatedUser(ctx context.Context) (*gogithub.User, error) {
	u, _, err := c.gh.Users.Get(ctx, "")
	if err != nil {
		return nil, err
	}
	return userFromGitHub(u), nil
}

// GetUser returns information about a specific user.
func (c *client) GetUser(ctx context.Context, username string) (*gogithub.User, error) {
	u, _, err := c.gh.Users.Get(ctx, username)
	if err != nil {
		return nil, err
	}
	return userFromGitHub(u), nil
}

// GetRepository retrieves a repository by owner and name.
func (c *client) GetRepository(ctx context.Context, owner, repo string) (*gogithub.Repository, error) {
	r, _, err := c.gh.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	return repositoryFromGitHub(r), nil
}

// ListUserRepos lists all repositories for a user.
func (c *client) ListUserRepos(ctx context.Context, user string) ([]*gogithub.Repository, error) {
	var allRepos []*github.Repository
	opts := &github.RepositoryListByUserOptions{
		Type:        "all",
		ListOptions: github.ListOptions{PerPage: 100},
	}
	for {
		repos, resp, err := c.gh.Repositories.ListByUser(ctx, user, opts)
		if err != nil {
			return nil, fmt.Errorf("list user repos: %w", err)
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return repositoriesFromGitHub(allRepos), nil
}

// ListOrgRepos lists all repositories for an organization.
func (c *client) ListOrgRepos(ctx context.Context, org string) ([]*gogithub.Repository, error) {
	var allRepos []*github.Repository
	opts := &github.RepositoryListByOrgOptions{
		Type:        "all",
		ListOptions: github.ListOptions{PerPage: 100},
	}
	for {
		repos, resp, err := c.gh.Repositories.ListByOrg(ctx, org, opts)
		if err != nil {
			return nil, fmt.Errorf("list org repos: %w", err)
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return repositoriesFromGitHub(allRepos), nil
}

// GetFileContent fetches a file's content from a repository.
func (c *client) GetFileContent(ctx context.Context, owner, repo, path string, opts *gogithub.ContentOptions) ([]byte, error) {
	getOpts := &github.RepositoryContentGetOptions{}
	if opts != nil && opts.Ref != "" {
		getOpts.Ref = opts.Ref
	}
	content, _, resp, err := c.gh.Repositories.GetContents(ctx, owner, repo, path, getOpts)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			return nil, fmt.Errorf("file not found: %s", path)
		}
		return nil, fmt.Errorf("get file content %s: %w", path, err)
	}
	if content == nil {
		return nil, fmt.Errorf("path is a directory, not a file: %s", path)
	}
	if content.GetType() != "file" {
		return nil, fmt.Errorf("path is not a file: %s (type: %s)", path, content.GetType())
	}
	decoded, err := content.GetContent()
	if err != nil {
		return nil, fmt.Errorf("decode file content %s: %w", path, err)
	}
	return []byte(decoded), nil
}

// GetFileContentString fetches a file's content as a string.
func (c *client) GetFileContentString(ctx context.Context, owner, repo, path string, opts *gogithub.ContentOptions) (string, error) {
	content, err := c.GetFileContent(ctx, owner, repo, path, opts)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// ListDirectory lists files in a directory.
func (c *client) ListDirectory(ctx context.Context, owner, repo, path string, opts *gogithub.ContentOptions) ([]*gogithub.FileContent, error) {
	getOpts := &github.RepositoryContentGetOptions{}
	if opts != nil && opts.Ref != "" {
		getOpts.Ref = opts.Ref
	}
	_, dirContents, resp, err := c.gh.Repositories.GetContents(ctx, owner, repo, path, getOpts)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			return nil, fmt.Errorf("directory not found: %s", path)
		}
		return nil, fmt.Errorf("list directory %s: %w", path, err)
	}
	if dirContents == nil {
		return nil, fmt.Errorf("path is a file, not a directory: %s", path)
	}
	return fileContentsFromGitHub(dirContents), nil
}

// FileExists checks if a file exists in a repository.
func (c *client) FileExists(ctx context.Context, owner, repo, path string, opts *gogithub.ContentOptions) (bool, error) {
	getOpts := &github.RepositoryContentGetOptions{}
	if opts != nil && opts.Ref != "" {
		getOpts.Ref = opts.Ref
	}
	content, _, resp, err := c.gh.Repositories.GetContents(ctx, owner, repo, path, getOpts)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			return false, nil
		}
		return false, fmt.Errorf("check file exists %s: %w", path, err)
	}
	return content != nil && content.GetType() == "file", nil
}

// GetRef retrieves a git reference by its full name.
func (c *client) GetRef(ctx context.Context, owner, repo, ref string) (*gogithub.Reference, error) {
	r, _, err := c.gh.Git.GetRef(ctx, owner, repo, ref)
	if err != nil {
		return nil, err
	}
	return referenceFromGitHub(r), nil
}

// GetBranchSHA returns the commit SHA for a branch.
func (c *client) GetBranchSHA(ctx context.Context, owner, repo, branch string) (string, error) {
	ref, err := c.GetRef(ctx, owner, repo, "refs/heads/"+branch)
	if err != nil {
		return "", fmt.Errorf("get branch %s: %w", branch, err)
	}
	if ref.Object == nil {
		return "", fmt.Errorf("branch %s has no object", branch)
	}
	return ref.Object.SHA, nil
}

// GetTagSHA returns the commit SHA for a tag.
func (c *client) GetTagSHA(ctx context.Context, owner, repo, tag string) (string, error) {
	ref, err := c.GetRef(ctx, owner, repo, "refs/tags/"+tag)
	if err != nil {
		return "", fmt.Errorf("get tag %s: %w", tag, err)
	}
	if ref.Object == nil {
		return "", fmt.Errorf("tag %s has no object", tag)
	}
	return ref.Object.SHA, nil
}

// ListBranches lists all branches in a repository.
func (c *client) ListBranches(ctx context.Context, owner, repo string) ([]*gogithub.Branch, error) {
	var allBranches []*github.Branch
	opts := &github.BranchListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	for {
		branches, resp, err := c.gh.Repositories.ListBranches(ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("list branches: %w", err)
		}
		allBranches = append(allBranches, branches...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return branchesFromGitHub(allBranches), nil
}

// GetCommit retrieves a commit by SHA.
func (c *client) GetCommit(ctx context.Context, owner, repo, sha string) (*gogithub.Commit, error) {
	commit, _, err := c.gh.Repositories.GetCommit(ctx, owner, repo, sha, nil)
	if err != nil {
		return nil, err
	}
	return commitFromGitHub(commit), nil
}

// ListCommits lists commits in a repository.
func (c *client) ListCommits(ctx context.Context, owner, repo string, opts *ListCommitsOptions) ([]*gogithub.Commit, error) {
	listOpts := &github.CommitsListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	if opts != nil {
		listOpts.SHA = opts.SHA
		listOpts.Path = opts.Path
		listOpts.Author = opts.Author
		if opts.Since != nil {
			listOpts.Since = *opts.Since
		}
		if opts.Until != nil {
			listOpts.Until = *opts.Until
		}
	}
	var allCommits []*github.RepositoryCommit
	for {
		commits, resp, err := c.gh.Repositories.ListCommits(ctx, owner, repo, listOpts)
		if err != nil {
			return nil, fmt.Errorf("list commits: %w", err)
		}
		allCommits = append(allCommits, commits...)
		if resp.NextPage == 0 {
			break
		}
		listOpts.Page = resp.NextPage
	}
	return commitsFromGitHub(allCommits), nil
}

// GetPullRequest retrieves a pull request by number.
func (c *client) GetPullRequest(ctx context.Context, owner, repo string, number int) (*gogithub.PullRequest, error) {
	pr, _, err := c.gh.PullRequests.Get(ctx, owner, repo, number)
	if err != nil {
		return nil, err
	}
	return pullRequestFromGitHub(pr), nil
}

// ListPullRequests lists pull requests in a repository.
func (c *client) ListPullRequests(ctx context.Context, owner, repo string, opts *ListPullRequestsOptions) ([]*gogithub.PullRequest, error) {
	listOpts := &github.PullRequestListOptions{
		State:       "open",
		ListOptions: github.ListOptions{PerPage: 100},
	}
	if opts != nil {
		if opts.State != "" {
			listOpts.State = opts.State
		}
		listOpts.Head = opts.Head
		listOpts.Base = opts.Base
		listOpts.Sort = opts.Sort
		listOpts.Direction = opts.Direction
	}
	var allPRs []*github.PullRequest
	for {
		prs, resp, err := c.gh.PullRequests.List(ctx, owner, repo, listOpts)
		if err != nil {
			return nil, fmt.Errorf("list pull requests: %w", err)
		}
		allPRs = append(allPRs, prs...)
		if resp.NextPage == 0 {
			break
		}
		listOpts.Page = resp.NextPage
	}
	return pullRequestsFromGitHub(allPRs), nil
}

// CreatePullRequest creates a new pull request.
func (c *client) CreatePullRequest(ctx context.Context, owner, repo string, input *CreatePullRequestInput) (*gogithub.PullRequest, error) {
	newPR := &github.NewPullRequest{
		Title:               github.Ptr(input.Title),
		Head:                github.Ptr(input.Head),
		Base:                github.Ptr(input.Base),
		Body:                github.Ptr(input.Body),
		Draft:               github.Ptr(input.Draft),
		MaintainerCanModify: github.Ptr(input.MaintainerCanModify),
	}
	pr, _, err := c.gh.PullRequests.Create(ctx, owner, repo, newPR)
	if err != nil {
		return nil, err
	}
	return pullRequestFromGitHub(pr), nil
}

// ListCheckRuns lists check runs for a git reference.
func (c *client) ListCheckRuns(ctx context.Context, owner, repo, ref string) ([]*gogithub.CheckRun, error) {
	var allRuns []*github.CheckRun
	opts := &github.ListCheckRunsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	for {
		result, resp, err := c.gh.Checks.ListCheckRunsForRef(ctx, owner, repo, ref, opts)
		if err != nil {
			return nil, fmt.Errorf("list check runs: %w", err)
		}
		allRuns = append(allRuns, result.CheckRuns...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return checkRunsFromGitHub(allRuns), nil
}

// GetLatestRelease retrieves the latest release.
func (c *client) GetLatestRelease(ctx context.Context, owner, repo string) (*gogithub.Release, error) {
	release, _, err := c.gh.Repositories.GetLatestRelease(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	return releaseFromGitHub(release), nil
}

// GetReleaseByTag retrieves a release by its tag name.
func (c *client) GetReleaseByTag(ctx context.Context, owner, repo, tag string) (*gogithub.Release, error) {
	release, _, err := c.gh.Repositories.GetReleaseByTag(ctx, owner, repo, tag)
	if err != nil {
		return nil, err
	}
	return releaseFromGitHub(release), nil
}

// ListReleases lists all releases in a repository.
func (c *client) ListReleases(ctx context.Context, owner, repo string) ([]*gogithub.Release, error) {
	var allReleases []*github.RepositoryRelease
	opts := &github.ListOptions{PerPage: 100}
	for {
		releases, resp, err := c.gh.Repositories.ListReleases(ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("list releases: %w", err)
		}
		allReleases = append(allReleases, releases...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return releasesFromGitHub(allReleases), nil
}

// GetDefaultBranch returns the default branch name for a repository.
func (c *client) GetDefaultBranch(ctx context.Context, owner, repo string) (string, error) {
	r, _, err := c.gh.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return "", err
	}
	return r.GetDefaultBranch(), nil
}

// CreateFork creates a fork of a repository.
func (c *client) CreateFork(ctx context.Context, owner, repo string, opts *CreateForkOptions) (*gogithub.Repository, error) {
	forkOpts := &github.RepositoryCreateForkOptions{}
	if opts != nil {
		forkOpts.Organization = opts.Organization
		forkOpts.Name = opts.Name
		forkOpts.DefaultBranchOnly = opts.DefaultBranch
	}
	fork, _, err := c.gh.Repositories.CreateFork(ctx, owner, repo, forkOpts)
	if err != nil {
		// AcceptedError means fork is being created asynchronously
		if _, ok := err.(*github.AcceptedError); ok {
			return repositoryFromGitHub(fork), nil
		}
		return nil, fmt.Errorf("create fork: %w", err)
	}
	return repositoryFromGitHub(fork), nil
}

// CreateRef creates a git reference.
func (c *client) CreateRef(ctx context.Context, owner, repo, ref, sha string) (*gogithub.Reference, error) {
	newRef := github.CreateRef{
		Ref: ref,
		SHA: sha,
	}
	r, _, err := c.gh.Git.CreateRef(ctx, owner, repo, newRef)
	if err != nil {
		return nil, fmt.Errorf("create ref: %w", err)
	}
	return referenceFromGitHub(r), nil
}

// UpdateRef updates a git reference to point to a new SHA.
func (c *client) UpdateRef(ctx context.Context, owner, repo, ref, sha string, force bool) (*gogithub.Reference, error) {
	updateRef := github.UpdateRef{
		SHA:   sha,
		Force: github.Ptr(force),
	}
	r, _, err := c.gh.Git.UpdateRef(ctx, owner, repo, ref, updateRef)
	if err != nil {
		return nil, fmt.Errorf("update ref: %w", err)
	}
	return referenceFromGitHub(r), nil
}

// DeleteRef deletes a git reference.
func (c *client) DeleteRef(ctx context.Context, owner, repo, ref string) error {
	_, err := c.gh.Git.DeleteRef(ctx, owner, repo, ref)
	return err
}

// ListTags lists all tags in a repository.
func (c *client) ListTags(ctx context.Context, owner, repo string) ([]*gogithub.Tag, error) {
	var allTags []*github.RepositoryTag
	opts := &github.ListOptions{PerPage: 100}
	for {
		tags, resp, err := c.gh.Repositories.ListTags(ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("list tags: %w", err)
		}
		allTags = append(allTags, tags...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return tagsFromGitHub(allTags), nil
}

// CreateTag creates an annotated tag.
func (c *client) CreateTag(ctx context.Context, owner, repo, tag, sha, message string) error {
	// Create annotated tag object
	tagObj := github.CreateTag{
		Tag:     tag,
		Message: message,
		Object:  sha,
		Type:    "commit",
	}
	createdTag, _, err := c.gh.Git.CreateTag(ctx, owner, repo, tagObj)
	if err != nil {
		return fmt.Errorf("create tag object: %w", err)
	}

	// Create reference to tag
	ref := github.CreateRef{
		Ref: "refs/tags/" + tag,
		SHA: createdTag.GetSHA(),
	}
	_, _, err = c.gh.Git.CreateRef(ctx, owner, repo, ref)
	if err != nil {
		return fmt.Errorf("create tag reference: %w", err)
	}
	return nil
}

// CreateCommit creates a commit with the given tree and parent.
func (c *client) CreateCommit(ctx context.Context, owner, repo string, opts *CreateCommitOptions) (*gogithub.Commit, error) {
	// Build parent commits
	parents := make([]*github.Commit, len(opts.Parents))
	for i, sha := range opts.Parents {
		parents[i] = &github.Commit{SHA: github.Ptr(sha)}
	}

	commit := &github.Commit{
		Message: github.Ptr(opts.Message),
		Tree:    &github.Tree{SHA: github.Ptr(opts.Tree)},
		Parents: parents,
	}
	if opts.Author != nil {
		commit.Author = &github.CommitAuthor{
			Name:  github.Ptr(opts.Author.Name),
			Email: github.Ptr(opts.Author.Email),
		}
		if opts.Author.Date != nil {
			commit.Author.Date = &github.Timestamp{Time: *opts.Author.Date}
		}
	}
	created, _, err := c.gh.Git.CreateCommit(ctx, owner, repo, *commit, nil)
	if err != nil {
		return nil, fmt.Errorf("create commit: %w", err)
	}
	return gitCommitToCommit(created), nil
}

// CreateTree creates a git tree from file entries.
func (c *client) CreateTree(ctx context.Context, owner, repo, baseTree string, entries []TreeEntry) (string, error) {
	ghEntries := make([]*github.TreeEntry, len(entries))
	for i, e := range entries {
		ghEntries[i] = &github.TreeEntry{
			Path: github.Ptr(e.Path),
			Mode: github.Ptr(e.Mode),
			Type: github.Ptr(e.Type),
		}
		if e.SHA != "" {
			ghEntries[i].SHA = github.Ptr(e.SHA)
		}
		if e.Content != "" {
			ghEntries[i].Content = github.Ptr(e.Content)
		}
	}
	tree, _, err := c.gh.Git.CreateTree(ctx, owner, repo, baseTree, ghEntries)
	if err != nil {
		return "", fmt.Errorf("create tree: %w", err)
	}
	return tree.GetSHA(), nil
}

// UpdatePullRequest updates a pull request.
func (c *client) UpdatePullRequest(ctx context.Context, owner, repo string, number int, input *UpdatePullRequestInput) (*gogithub.PullRequest, error) {
	pr := &github.PullRequest{}
	if input.Title != nil {
		pr.Title = input.Title
	}
	if input.Body != nil {
		pr.Body = input.Body
	}
	if input.State != nil {
		pr.State = input.State
	}
	if input.Base != nil {
		pr.Base = &github.PullRequestBranch{Ref: input.Base}
	}
	if input.MaintainerCanModify != nil {
		pr.MaintainerCanModify = input.MaintainerCanModify
	}
	updated, _, err := c.gh.PullRequests.Edit(ctx, owner, repo, number, pr)
	if err != nil {
		return nil, err
	}
	return pullRequestFromGitHub(updated), nil
}

// MergePullRequest merges a pull request.
func (c *client) MergePullRequest(ctx context.Context, owner, repo string, number int, opts *MergePullRequestOptions) (*gogithub.MergeResult, error) {
	var commitMsg string
	var ghOpts *github.PullRequestOptions
	if opts != nil {
		commitMsg = opts.CommitMessage
		ghOpts = &github.PullRequestOptions{
			CommitTitle: opts.CommitTitle,
			MergeMethod: opts.MergeMethod,
			SHA:         opts.SHA,
		}
	}
	result, _, err := c.gh.PullRequests.Merge(ctx, owner, repo, number, commitMsg, ghOpts)
	if err != nil {
		return nil, err
	}
	return &gogithub.MergeResult{
		SHA:     result.GetSHA(),
		Merged:  result.GetMerged(),
		Message: result.GetMessage(),
	}, nil
}

// ListPullRequestFiles lists files changed in a pull request.
func (c *client) ListPullRequestFiles(ctx context.Context, owner, repo string, number int) ([]*gogithub.CommitFile, error) {
	var allFiles []*github.CommitFile
	opts := &github.ListOptions{PerPage: 100}
	for {
		files, resp, err := c.gh.PullRequests.ListFiles(ctx, owner, repo, number, opts)
		if err != nil {
			return nil, fmt.Errorf("list PR files: %w", err)
		}
		allFiles = append(allFiles, files...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return commitFilesFromGitHub(allFiles), nil
}

// GetPullRequestDiff gets the diff for a pull request.
func (c *client) GetPullRequestDiff(ctx context.Context, owner, repo string, number int) (string, error) {
	diff, _, err := c.gh.PullRequests.GetRaw(ctx, owner, repo, number, github.RawOptions{Type: github.Diff})
	if err != nil {
		return "", fmt.Errorf("get PR diff: %w", err)
	}
	return diff, nil
}

// GetPullRequestPatch gets the patch for a pull request.
func (c *client) GetPullRequestPatch(ctx context.Context, owner, repo string, number int) (string, error) {
	patch, _, err := c.gh.PullRequests.GetRaw(ctx, owner, repo, number, github.RawOptions{Type: github.Patch})
	if err != nil {
		return "", fmt.Errorf("get PR patch: %w", err)
	}
	return patch, nil
}

// CreatePullRequestReview creates a review on a pull request.
func (c *client) CreatePullRequestReview(ctx context.Context, owner, repo string, number int, input *CreateReviewInput) (*gogithub.PullRequestReview, error) {
	review := &github.PullRequestReviewRequest{
		Event: github.Ptr(input.Event),
		Body:  github.Ptr(input.Body),
	}
	result, _, err := c.gh.PullRequests.CreateReview(ctx, owner, repo, number, review)
	if err != nil {
		return nil, err
	}
	return pullRequestReviewFromGitHub(result), nil
}

// ListPullRequestReviews lists reviews on a pull request.
func (c *client) ListPullRequestReviews(ctx context.Context, owner, repo string, number int) ([]*gogithub.PullRequestReview, error) {
	var allReviews []*github.PullRequestReview
	opts := &github.ListOptions{PerPage: 100}
	for {
		reviews, resp, err := c.gh.PullRequests.ListReviews(ctx, owner, repo, number, opts)
		if err != nil {
			return nil, fmt.Errorf("list PR reviews: %w", err)
		}
		allReviews = append(allReviews, reviews...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return pullRequestReviewsFromGitHub(allReviews), nil
}

// RequestReviewers requests reviewers for a pull request.
func (c *client) RequestReviewers(ctx context.Context, owner, repo string, number int, reviewers, teamReviewers []string) (*gogithub.PullRequest, error) {
	pr, _, err := c.gh.PullRequests.RequestReviewers(ctx, owner, repo, number, github.ReviewersRequest{
		Reviewers:     reviewers,
		TeamReviewers: teamReviewers,
	})
	if err != nil {
		return nil, err
	}
	return pullRequestFromGitHub(pr), nil
}

// CreatePullRequestComment creates a comment on a pull request diff.
func (c *client) CreatePullRequestComment(ctx context.Context, owner, repo string, number int, input *CreatePRCommentInput) (*gogithub.PullRequestComment, error) {
	comment := &github.PullRequestComment{
		Body:     github.Ptr(input.Body),
		CommitID: github.Ptr(input.CommitID),
		Path:     github.Ptr(input.Path),
		Line:     github.Ptr(input.Line),
	}
	if input.Side != "" {
		comment.Side = github.Ptr(input.Side)
	}
	result, _, err := c.gh.PullRequests.CreateComment(ctx, owner, repo, number, comment)
	if err != nil {
		return nil, err
	}
	return pullRequestCommentFromGitHub(result), nil
}

// ListPullRequestComments lists comments on a pull request.
func (c *client) ListPullRequestComments(ctx context.Context, owner, repo string, number int) ([]*gogithub.PullRequestComment, error) {
	var allComments []*github.PullRequestComment
	opts := &github.PullRequestListCommentsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	for {
		comments, resp, err := c.gh.PullRequests.ListComments(ctx, owner, repo, number, opts)
		if err != nil {
			return nil, fmt.Errorf("list PR comments: %w", err)
		}
		allComments = append(allComments, comments...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return pullRequestCommentsFromGitHub(allComments), nil
}

// CreateIssueComment creates a comment on an issue or pull request.
func (c *client) CreateIssueComment(ctx context.Context, owner, repo string, number int, body string) (*gogithub.IssueComment, error) {
	comment := &github.IssueComment{
		Body: github.Ptr(body),
	}
	result, _, err := c.gh.Issues.CreateComment(ctx, owner, repo, number, comment)
	if err != nil {
		return nil, err
	}
	return issueCommentFromGitHub(result), nil
}

// GetCheckRun retrieves a check run by ID.
func (c *client) GetCheckRun(ctx context.Context, owner, repo string, checkRunID int64) (*gogithub.CheckRun, error) {
	check, _, err := c.gh.Checks.GetCheckRun(ctx, owner, repo, checkRunID)
	if err != nil {
		return nil, err
	}
	return checkRunFromGitHub(check), nil
}

// ListCheckSuites lists check suites for a git reference.
func (c *client) ListCheckSuites(ctx context.Context, owner, repo, ref string) ([]*gogithub.CheckSuite, error) {
	var allSuites []*github.CheckSuite
	opts := &github.ListCheckSuiteOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	for {
		result, resp, err := c.gh.Checks.ListCheckSuitesForRef(ctx, owner, repo, ref, opts)
		if err != nil {
			return nil, fmt.Errorf("list check suites: %w", err)
		}
		allSuites = append(allSuites, result.CheckSuites...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return checkSuitesFromGitHub(allSuites), nil
}

// GetRelease retrieves a release by ID.
func (c *client) GetRelease(ctx context.Context, owner, repo string, id int64) (*gogithub.Release, error) {
	release, _, err := c.gh.Repositories.GetRelease(ctx, owner, repo, id)
	if err != nil {
		return nil, err
	}
	return releaseFromGitHub(release), nil
}

// CreateRelease creates a new release.
func (c *client) CreateRelease(ctx context.Context, owner, repo string, input *CreateReleaseInput) (*gogithub.Release, error) {
	req := github.CreateReleaseRequest{
		TagName:              input.TagName,
		TargetCommitish:      github.Ptr(input.TargetCommitish),
		Name:                 github.Ptr(input.Name),
		Body:                 github.Ptr(input.Body),
		Draft:                github.Ptr(input.Draft),
		Prerelease:           github.Ptr(input.Prerelease),
		GenerateReleaseNotes: github.Ptr(input.GenerateReleaseNotes),
	}
	release, _, err := c.gh.Repositories.CreateRelease(ctx, owner, repo, req)
	if err != nil {
		return nil, fmt.Errorf("create release: %w", err)
	}
	return releaseFromGitHub(release), nil
}

// UpdateRelease updates a release.
func (c *client) UpdateRelease(ctx context.Context, owner, repo string, id int64, input *UpdateReleaseInput) (*gogithub.Release, error) {
	req := github.UpdateReleaseRequest{
		TagName:         input.TagName,
		TargetCommitish: input.TargetCommitish,
		Name:            input.Name,
		Body:            input.Body,
		Draft:           input.Draft,
		Prerelease:      input.Prerelease,
	}
	release, _, err := c.gh.Repositories.UpdateRelease(ctx, owner, repo, id, req)
	if err != nil {
		return nil, fmt.Errorf("update release: %w", err)
	}
	return releaseFromGitHub(release), nil
}

// DeleteRelease deletes a release.
func (c *client) DeleteRelease(ctx context.Context, owner, repo string, id int64) error {
	_, err := c.gh.Repositories.DeleteRelease(ctx, owner, repo, id)
	return err
}

// ListReleaseAssets lists assets for a release.
func (c *client) ListReleaseAssets(ctx context.Context, owner, repo string, releaseID int64) ([]*gogithub.ReleaseAsset, error) {
	var allAssets []*github.ReleaseAsset
	opts := &github.ListOptions{PerPage: 100}
	for {
		assets, resp, err := c.gh.Repositories.ListReleaseAssets(ctx, owner, repo, releaseID, opts)
		if err != nil {
			return nil, fmt.Errorf("list release assets: %w", err)
		}
		allAssets = append(allAssets, assets...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return releaseAssetsFromGitHub(allAssets), nil
}

// SearchIssues searches for issues and pull requests.
func (c *client) SearchIssues(ctx context.Context, query string, opts *SearchOptions) (*gogithub.IssueSearchResult, error) {
	searchOpts := &github.SearchOptions{}
	if opts != nil {
		searchOpts.Sort = opts.Sort
		searchOpts.Order = opts.Order
		if opts.PerPage > 0 {
			searchOpts.ListOptions.PerPage = opts.PerPage
		}
		if opts.Page > 0 {
			searchOpts.ListOptions.Page = opts.Page
		}
	}
	result, _, err := c.gh.Search.Issues(ctx, query, searchOpts)
	if err != nil {
		return nil, fmt.Errorf("search issues: %w", err)
	}
	return issueSearchResultFromGitHub(result), nil
}

// GetContributorStats gets contribution statistics for a repository.
func (c *client) GetContributorStats(ctx context.Context, owner, repo string) ([]*gogithub.ContributorStats, error) {
	stats, _, err := c.gh.Repositories.ListContributorsStats(ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("get contributor stats: %w", err)
	}
	return contributorStatsFromGitHub(stats), nil
}
