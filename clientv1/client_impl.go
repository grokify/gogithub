package clientv1

import (
	"context"
	"fmt"

	"github.com/google/go-github/v89/github"
	"github.com/grokify/gogithub"
	"github.com/grokify/gogithub/auth"
)

// client implements the Client interface.
type client struct {
	gh *github.Client
}

// NewClient creates a new GitHub client with token authentication.
// This is the primary way to create a version-isolated GitHub client.
func NewClient(ctx context.Context, token string) (Client, error) {
	gh, err := auth.NewGitHubClient(ctx, token)
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
