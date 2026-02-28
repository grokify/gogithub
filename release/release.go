// Package release provides GitHub release operations.
package release

import (
	"context"
	"fmt"

	"github.com/google/go-github/v84/github"
)

// ListReleases lists all releases for a repository with pagination.
func ListReleases(ctx context.Context, gh *github.Client, owner, repo string) ([]*github.RepositoryRelease, error) {
	var allReleases []*github.RepositoryRelease

	opts := &github.ListOptions{
		PerPage: 100,
	}

	for {
		releases, resp, err := gh.Repositories.ListReleases(ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("list releases: %w", err)
		}

		allReleases = append(allReleases, releases...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allReleases, nil
}

// ListReleasesSince lists releases published after a specific release ID.
// Useful for incremental syncs.
func ListReleasesSince(ctx context.Context, gh *github.Client, owner, repo string, sinceID int64) ([]*github.RepositoryRelease, error) {
	allReleases, err := ListReleases(ctx, gh, owner, repo)
	if err != nil {
		return nil, err
	}

	var newReleases []*github.RepositoryRelease
	for _, r := range allReleases {
		if r.GetID() > sinceID {
			newReleases = append(newReleases, r)
		}
	}

	return newReleases, nil
}

// GetRelease retrieves a specific release by ID.
func GetRelease(ctx context.Context, gh *github.Client, owner, repo string, id int64) (*github.RepositoryRelease, error) {
	release, _, err := gh.Repositories.GetRelease(ctx, owner, repo, id)
	return release, err
}

// GetLatestRelease retrieves the latest published release.
func GetLatestRelease(ctx context.Context, gh *github.Client, owner, repo string) (*github.RepositoryRelease, error) {
	release, _, err := gh.Repositories.GetLatestRelease(ctx, owner, repo)
	return release, err
}

// GetReleaseByTag retrieves a release by its tag name.
func GetReleaseByTag(ctx context.Context, gh *github.Client, owner, repo, tag string) (*github.RepositoryRelease, error) {
	release, _, err := gh.Repositories.GetReleaseByTag(ctx, owner, repo, tag)
	return release, err
}

// ListReleaseAssets lists assets for a release.
func ListReleaseAssets(ctx context.Context, gh *github.Client, owner, repo string, releaseID int64) ([]*github.ReleaseAsset, error) {
	var allAssets []*github.ReleaseAsset

	opts := &github.ListOptions{
		PerPage: 100,
	}

	for {
		assets, resp, err := gh.Repositories.ListReleaseAssets(ctx, owner, repo, releaseID, opts)
		if err != nil {
			return nil, fmt.Errorf("list release assets: %w", err)
		}

		allAssets = append(allAssets, assets...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allAssets, nil
}

// CreateRelease creates a new release for a repository.
func CreateRelease(ctx context.Context, gh *github.Client, owner, repo string, release *github.RepositoryRelease) (*github.RepositoryRelease, error) {
	created, _, err := gh.Repositories.CreateRelease(ctx, owner, repo, release)
	if err != nil {
		return nil, fmt.Errorf("create release: %w", err)
	}
	return created, nil
}

// CreateReleaseSimple creates a release with common options.
func CreateReleaseSimple(ctx context.Context, gh *github.Client, owner, repo, tagName, name, body string, draft, prerelease, generateNotes bool) (*github.RepositoryRelease, error) {
	release := &github.RepositoryRelease{
		TagName:              github.Ptr(tagName),
		Name:                 github.Ptr(name),
		Body:                 github.Ptr(body),
		Draft:                github.Ptr(draft),
		Prerelease:           github.Ptr(prerelease),
		GenerateReleaseNotes: github.Ptr(generateNotes),
	}
	return CreateRelease(ctx, gh, owner, repo, release)
}

// DeleteRelease deletes a release by ID.
func DeleteRelease(ctx context.Context, gh *github.Client, owner, repo string, releaseID int64) error {
	_, err := gh.Repositories.DeleteRelease(ctx, owner, repo, releaseID)
	return err
}

// EditRelease updates a release.
func EditRelease(ctx context.Context, gh *github.Client, owner, repo string, releaseID int64, release *github.RepositoryRelease) (*github.RepositoryRelease, error) {
	updated, _, err := gh.Repositories.EditRelease(ctx, owner, repo, releaseID, release)
	return updated, err
}
