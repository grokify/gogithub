// Package release provides GitHub release operations.
package release

import (
	"context"

	"github.com/grokify/gogithub"
	"github.com/grokify/gogithub/clientv1"
)

// ListReleases lists all releases for a repository.
func ListReleases(ctx context.Context, client clientv1.Client, owner, repo string) ([]*gogithub.Release, error) {
	return client.ListReleases(ctx, owner, repo)
}

// ListReleasesSince lists releases published after a specific release ID.
// Useful for incremental syncs.
func ListReleasesSince(ctx context.Context, client clientv1.Client, owner, repo string, sinceID int64) ([]*gogithub.Release, error) {
	allReleases, err := client.ListReleases(ctx, owner, repo)
	if err != nil {
		return nil, err
	}

	var newReleases []*gogithub.Release
	for _, r := range allReleases {
		if r.ID > sinceID {
			newReleases = append(newReleases, r)
		}
	}

	return newReleases, nil
}

// GetRelease retrieves a specific release by ID.
func GetRelease(ctx context.Context, client clientv1.Client, owner, repo string, id int64) (*gogithub.Release, error) {
	return client.GetRelease(ctx, owner, repo, id)
}

// GetLatestRelease retrieves the latest published release.
func GetLatestRelease(ctx context.Context, client clientv1.Client, owner, repo string) (*gogithub.Release, error) {
	return client.GetLatestRelease(ctx, owner, repo)
}

// GetReleaseByTag retrieves a release by its tag name.
func GetReleaseByTag(ctx context.Context, client clientv1.Client, owner, repo, tag string) (*gogithub.Release, error) {
	return client.GetReleaseByTag(ctx, owner, repo, tag)
}

// ListReleaseAssets lists assets for a release.
func ListReleaseAssets(ctx context.Context, client clientv1.Client, owner, repo string, releaseID int64) ([]*gogithub.ReleaseAsset, error) {
	return client.ListReleaseAssets(ctx, owner, repo, releaseID)
}

// CreateRelease creates a new release for a repository.
func CreateRelease(ctx context.Context, client clientv1.Client, owner, repo string, input *clientv1.CreateReleaseInput) (*gogithub.Release, error) {
	return client.CreateRelease(ctx, owner, repo, input)
}

// CreateReleaseSimple creates a release with common options.
func CreateReleaseSimple(ctx context.Context, client clientv1.Client, owner, repo, tagName, name, body string, draft, prerelease, generateNotes bool) (*gogithub.Release, error) {
	input := &clientv1.CreateReleaseInput{
		TagName:              tagName,
		Name:                 name,
		Body:                 body,
		Draft:                draft,
		Prerelease:           prerelease,
		GenerateReleaseNotes: generateNotes,
	}
	return client.CreateRelease(ctx, owner, repo, input)
}

// DeleteRelease deletes a release by ID.
func DeleteRelease(ctx context.Context, client clientv1.Client, owner, repo string, releaseID int64) error {
	return client.DeleteRelease(ctx, owner, repo, releaseID)
}

// UpdateRelease updates a release.
func UpdateRelease(ctx context.Context, client clientv1.Client, owner, repo string, releaseID int64, input *clientv1.UpdateReleaseInput) (*gogithub.Release, error) {
	return client.UpdateRelease(ctx, owner, repo, releaseID, input)
}
