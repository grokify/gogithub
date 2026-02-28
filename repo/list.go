package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v84/github"
)

// ListOrgRepos lists all repositories for an organization with pagination.
func ListOrgRepos(ctx context.Context, gh *github.Client, org string) ([]*github.Repository, error) {
	var allRepos []*github.Repository

	opts := &github.RepositoryListByOrgOptions{
		Type: "all",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for {
		repos, resp, err := gh.Repositories.ListByOrg(ctx, org, opts)
		if err != nil {
			return nil, fmt.Errorf("list org repos: %w", err)
		}

		allRepos = append(allRepos, repos...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allRepos, nil
}

// ListUserRepos lists all repositories for a user with pagination.
func ListUserRepos(ctx context.Context, gh *github.Client, user string) ([]*github.Repository, error) {
	var allRepos []*github.Repository

	opts := &github.RepositoryListByUserOptions{
		Type: "all",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for {
		repos, resp, err := gh.Repositories.ListByUser(ctx, user, opts)
		if err != nil {
			return nil, fmt.Errorf("list user repos: %w", err)
		}

		allRepos = append(allRepos, repos...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allRepos, nil
}

// GetRepo retrieves a repository by owner and name.
func GetRepo(ctx context.Context, gh *github.Client, owner, repo string) (*github.Repository, error) {
	repository, _, err := gh.Repositories.Get(ctx, owner, repo)
	return repository, err
}

// ParseRepoName splits a full repo name (owner/repo) into owner and repo.
func ParseRepoName(fullName string) (owner, repo string, err error) {
	parts := strings.SplitN(fullName, "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repo name %q: expected owner/repo", fullName)
	}
	return parts[0], parts[1], nil
}
