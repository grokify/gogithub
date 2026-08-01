package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/grokify/gogithub"
	"github.com/grokify/gogithub/clientv1"
)

// ListOrgRepos lists all repositories for an organization.
func ListOrgRepos(ctx context.Context, client clientv1.Client, org string) ([]*gogithub.Repository, error) {
	return client.ListOrgRepos(ctx, org)
}

// ListUserRepos lists all repositories for a user.
func ListUserRepos(ctx context.Context, client clientv1.Client, user string) ([]*gogithub.Repository, error) {
	return client.ListUserRepos(ctx, user)
}

// GetRepo retrieves a repository by owner and name.
func GetRepo(ctx context.Context, client clientv1.Client, owner, repo string) (*gogithub.Repository, error) {
	return client.GetRepository(ctx, owner, repo)
}

// ParseRepoName splits a full repo name (owner/repo) into owner and repo.
func ParseRepoName(fullName string) (owner, repo string, err error) {
	parts := strings.SplitN(fullName, "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repo name %q: expected owner/repo", fullName)
	}
	return parts[0], parts[1], nil
}
